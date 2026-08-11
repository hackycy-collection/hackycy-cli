import type {
  EvidenceCoverage,
  EvidenceFact,
  FileRole,
  GitChangeSnapshot,
  SnapshotFile,
} from './types'
import path from 'node:path'

export const TARGET_INPUT_TOKENS = 2_000
export const MAX_INPUT_TOKENS = 3_000

interface ChangeCluster {
  key: string
  files: SnapshotFile[]
}

type CommitTypeHint = 'build' | 'chore' | 'ci' | 'docs' | 'style' | 'test'

export interface CompiledEvidence {
  text: string
  coverage: EvidenceCoverage
  facts: EvidenceFact[]
}

function estimateTextTokens(text: string): number {
  let ascii = 0
  let nonAscii = 0
  for (const codePoint of text) {
    if ((codePoint.codePointAt(0) ?? 0) <= 0x7F)
      ascii += 1
    else
      nonAscii += 1
  }
  return Math.ceil(ascii / 2) + nonAscii
}

export function estimateInputTokens(system: string, evidence: string): number {
  return estimateTextTokens(JSON.stringify([
    { role: 'system', content: system },
    { role: 'user', content: evidence },
  ])) + 32
}

function directory(filePath: string): string {
  const normalized = filePath.replaceAll('\\', '/')
  const separator = normalized.lastIndexOf('/')
  return separator === -1 ? '.' : normalized.slice(0, separator)
}

function moduleRoot(filePath: string): string | undefined {
  const normalized = filePath.replaceAll('\\', '/')
  const match = /^(?:packages|apps|services|crates)\/[^/]+|^src\/commands\/[^/]+/.exec(normalized)
  return match?.[0]
}

function commonDirectoryDepth(left: string, right: string): number {
  const leftParts = left === '.' ? [] : left.split('/')
  const rightParts = right === '.' ? [] : right.split('/')
  let depth = 0
  while (leftParts[depth] && leftParts[depth] === rightParts[depth])
    depth += 1
  return depth
}

function initialClusterKey(file: SnapshotFile): string {
  return moduleRoot(file.path) ?? directory(file.path)
}

function attachSupportingFile(file: SnapshotFile, existingKeys: string[]): string {
  const fileDirectory = directory(file.path)
  const candidates = existingKeys
    .map(key => ({ key, depth: commonDirectoryDepth(fileDirectory, key) }))
    .filter(candidate => candidate.depth > 0)
    .toSorted((left, right) => right.depth - left.depth || left.key.localeCompare(right.key))
  return candidates[0]?.key ?? initialClusterKey(file)
}

export function buildChangeClusters(snapshot: GitChangeSnapshot): ChangeCluster[] {
  const primary = snapshot.files.filter(file => !['config', 'dependency', 'docs'].includes(file.role))
  const keys = new Map<string, SnapshotFile[]>()
  for (const file of primary) {
    const key = initialClusterKey(file)
    const files = keys.get(key) ?? []
    files.push(file)
    keys.set(key, files)
  }
  const existingKeys = [...keys.keys()].sort()
  for (const file of snapshot.files.filter(file => !primary.includes(file))) {
    const key = attachSupportingFile(file, existingKeys)
    const files = keys.get(key) ?? []
    files.push(file)
    keys.set(key, files)
    if (!existingKeys.includes(key))
      existingKeys.push(key)
  }
  return [...keys]
    .map(([key, files]) => ({ key, files: files.toSorted((left, right) => left.path.localeCompare(right.path)) }))
    .toSorted((left, right) => left.key.localeCompare(right.key))
}

function roleRank(role: FileRole): number {
  if (role === 'dependency' || role === 'config')
    return 2
  if (role === 'test')
    return 3
  if (role === 'generated' || role === 'binary' || role === 'sensitive')
    return 6
  if (role === 'source')
    return 4
  return 5
}

function statusToken(file: SnapshotFile): string {
  if (file.indexStatus === '?' && file.worktreeStatus === '?')
    return 'A'
  if (file.indexStatus === 'R' || file.indexStatus === 'C')
    return file.indexStatus
  return `${file.indexStatus}${file.worktreeStatus}`.trim() || 'M'
}

function fact(
  priority: EvidenceFact['priority'],
  clusterKey: string,
  filePath: string,
  suffix: string,
  text: string,
  hunkId?: string,
): EvidenceFact {
  return { id: `${priority}:${clusterKey}:${filePath}:${suffix}`, priority, clusterKey, filePath, hunkId, text }
}

function declarationName(line: string): string | undefined {
  const match = /\bexport\s+(?:default\s+)?(?:async\s+)?(?:class|interface|type|enum|function|def|func|struct|const|let|var)\s+([A-Za-z_$][\w$]*)/.exec(line)
  return match?.[1]
}

function testName(line: string): string | undefined {
  const match = /\b(?:describe|test|it)\s*\(\s*['"`]([^'"`]+)['"`]/.exec(line)
  return match?.[1]
}

function documentationHeading(line: string): string | undefined {
  const match = /^#{1,6} (.+)$/.exec(line.trim())
  return match?.[1]?.trim()
}

function configurationKey(line: string): string | undefined {
  const match = /^\s*["']?([a-z_$][\w$.-]*)["']?\s*[:=]/i.exec(line)
  return match?.[1]
}

function meaningfulLine(line: string): boolean {
  return Boolean(line.trim()) && !/^[{});,]+$/.test(line.trim())
}

function isBehaviorLine(line: string): boolean {
  return /throw\s+new|\bError\(|--[\w-]+|\b(?:error|message|title|description|help)\b/i.test(line)
}

function hunkContext(line: string): string | undefined {
  const match = /\b(?:async\s+)?(class|interface|type|function|def|func|struct)\s+([A-Za-z_$][\w$]*)/.exec(line)
  return match ? `context ${match[1]} ${match[2]}` : undefined
}

function isLowInformationLine(line: string): boolean {
  return /^(?:import|export\s*\{).*|^\/\/|^\*|^[{});,]+$/.test(line.trim())
}

function changeKind(file: SnapshotFile, line: string): 'added' | 'removed' {
  return file.hunks.some(hunk => hunk.addedLines.includes(line)) ? 'added' : 'removed'
}

function packageFacts(file: SnapshotFile, clusterKey: string): EvidenceFact[] {
  const manifest = file.manifest
  if (!manifest)
    return []
  if (!manifest.before)
    return [fact(1, clusterKey, file.path, 'package-added', `package manifest added ${file.path}`)]
  if (!manifest.after)
    return [fact(1, clusterKey, file.path, 'package-removed', `package manifest removed ${file.path}`)]
  try {
    const before = JSON.parse(manifest.before) as Record<string, unknown>
    const after = JSON.parse(manifest.after) as Record<string, unknown>
    const facts: EvidenceFact[] = []
    for (const field of ['dependencies', 'devDependencies'] as const) {
      const beforeDependencies = (before[field] ?? {}) as Record<string, string>
      const afterDependencies = (after[field] ?? {}) as Record<string, string>
      const names = new Set([...Object.keys(beforeDependencies), ...Object.keys(afterDependencies)])
      const added: string[] = []
      const removed: string[] = []
      const changed: string[] = []
      for (const name of [...names].sort()) {
        if (beforeDependencies[name] === afterDependencies[name])
          continue
        if (!beforeDependencies[name])
          added.push(`${name}@${afterDependencies[name]}`)
        else if (!afterDependencies[name])
          removed.push(`${name}@${beforeDependencies[name]}`)
        else
          changed.push(`${name} ${beforeDependencies[name]} -> ${afterDependencies[name]}`)
      }
      if (added.length && removed.length) {
        facts.push(fact(1, clusterKey, file.path, `dependency:${field}:replacement`, `dependency replacement add ${added.join(', ')}; remove ${removed.join(', ')}`))
      }
      else {
        if (added.length)
          facts.push(fact(1, clusterKey, file.path, `dependency:${field}:added`, `dependency added ${added.join(', ')}`))
        if (removed.length)
          facts.push(fact(1, clusterKey, file.path, `dependency:${field}:removed`, `dependency removed ${removed.join(', ')}`))
      }
      if (changed.length)
        facts.push(fact(1, clusterKey, file.path, `dependency:${field}:changed`, `dependency updated ${changed.join(', ')}`))
    }
    for (const field of ['scripts', 'name', 'version', 'type', 'exports', 'bin', 'engines'] as const) {
      if (JSON.stringify(before[field]) !== JSON.stringify(after[field])) {
        facts.push(fact(
          1,
          clusterKey,
          file.path,
          `manifest:${field}`,
          field === 'version' ? `release chore ${String(before.version)}->${String(after.version)}` : `package ${field} changed`,
        ))
      }
    }
    return facts.length > 0 ? facts : [fact(1, clusterKey, file.path, 'package-generic', `package manifest changed ${file.path}`)]
  }
  catch {
    return [fact(1, clusterKey, file.path, 'package-unparsed', `package manifest changed ${file.path}`)]
  }
}

export function extractEvidenceFacts(clusters: ChangeCluster[]): EvidenceFact[] {
  const facts: EvidenceFact[] = []
  for (const cluster of clusters) {
    for (const file of cluster.files) {
      if (file.contentPolicy !== 'inspect')
        continue
      const factsBeforeFile = facts.length
      if (file.role === 'dependency') {
        facts.push(...packageFacts(file, cluster.key))
        continue
      }
      if (file.originalPath && (file.indexStatus === 'R' || file.indexStatus === 'C')) {
        facts.push(fact(1, cluster.key, file.path, 'rename', `${file.indexStatus === 'R' ? 'rename' : 'copy'} ${file.originalPath} -> ${file.path}`))
      }
      let includedHeading = false
      let includedDeclaration = false
      let includedTest = false
      let includedDocumentationHeading = false
      for (const hunk of file.hunks) {
        const context = hunk.heading && file.role !== 'test' ? hunkContext(hunk.heading) : undefined
        if (context && !includedHeading) {
          facts.push(fact(1, cluster.key, file.path, `heading:${hunk.id}`, context, hunk.id))
          includedHeading = true
        }
        for (const line of [...hunk.addedLines, ...hunk.deletedLines]) {
          const kind = changeKind(file, line)
          const declaration = declarationName(line)
          if (declaration && file.role !== 'test' && !includedDeclaration) {
            facts.push(fact(1, cluster.key, file.path, `declaration:${hunk.id}:${declaration}`, `symbol ${kind} ${declaration}`, hunk.id))
            includedDeclaration = true
          }
          const test = testName(line)
          if (test && !includedTest) {
            facts.push(fact(1, cluster.key, file.path, `test:${hunk.id}:${test}`, `test ${kind} ${JSON.stringify(test)}`, hunk.id))
            includedTest = true
          }
          const heading = file.role === 'docs' ? documentationHeading(line) : undefined
          if (heading && !includedDocumentationHeading) {
            facts.push(fact(1, cluster.key, file.path, `docs:${hunk.id}:${heading}`, `docs ${kind} ${JSON.stringify(heading)}`, hunk.id))
            includedDocumentationHeading = true
          }
          const key = file.role === 'config' ? configurationKey(line) : undefined
          if (key)
            facts.push(fact(1, cluster.key, file.path, `config:${hunk.id}:${key}`, `config ${kind} ${key}`, hunk.id))
        }
        if (file.role === 'test' || file.role === 'docs')
          continue
        const candidateLines = [
          ...hunk.addedLines.map(line => ({ kind: 'added', line })),
          ...hunk.deletedLines.map(line => ({ kind: 'removed', line })),
        ].filter(candidate => meaningfulLine(candidate.line) && !isLowInformationLine(candidate.line))
        let includedBehavior = false
        for (const [index, candidate] of candidateLines.entries()) {
          const trimmed = candidate.line.trim()
          if (declarationName(trimmed) || testName(trimmed))
            continue
          const priority: EvidenceFact['priority'] = isBehaviorLine(trimmed) ? 2 : 3
          if (priority === 2 && includedBehavior)
            continue
          includedBehavior ||= priority === 2
          facts.push(fact(priority, cluster.key, file.path, `line:${hunk.id}:${index}`, `${candidate.kind} ${trimmed}`, hunk.id))
        }
      }
      if (!facts.slice(factsBeforeFile).some(item => item.priority < 3))
        facts.push(fact(1, cluster.key, file.path, 'file', `file changed ${file.path}`))
    }
  }
  const unique = new Map<string, EvidenceFact>()
  for (const item of facts)
    unique.set(item.id, item)
  return [...unique.values()]
    .toSorted((left, right) => left.priority - right.priority
      || left.clusterKey.localeCompare(right.clusterKey)
      || left.filePath.localeCompare(right.filePath)
      || (left.hunkId ?? '').localeCompare(right.hunkId ?? '')
      || left.id.localeCompare(right.id))
}

function renderCluster(cluster: ChangeCluster): string {
  const stats = cluster.files.reduce((total, file) => ({
    additions: total.additions + file.stats.additions,
    deletions: total.deletions + file.stats.deletions,
  }), { additions: 0, deletions: 0 })
  return `cluster=${cluster.key} files=${cluster.files.length} +${stats.additions} -${stats.deletions}`
}

function protectionCounts(snapshot: GitChangeSnapshot): { metadataOnly: number, redacted: number } {
  return {
    metadataOnly: snapshot.files.filter(file => file.contentPolicy === 'metadata-only').length,
    redacted: snapshot.files.filter(file => file.contentPolicy === 'redacted').length,
  }
}

function packageVersionChanged(file: SnapshotFile): boolean {
  if (path.basename(file.path) !== 'package.json' || !file.manifest?.before || !file.manifest.after)
    return false
  try {
    const before = JSON.parse(file.manifest.before) as { version?: unknown }
    const after = JSON.parse(file.manifest.after) as { version?: unknown }
    return before.version !== after.version
  }
  catch {
    return false
  }
}

function commitTypeHint(snapshot: GitChangeSnapshot): CommitTypeHint | undefined {
  const files = snapshot.files
  const paths = files.map(file => file.path)
  if (paths.some(filePath => filePath.startsWith('.github/workflows/')))
    return 'ci'
  if (paths.some(filePath => path.basename(filePath).toLowerCase() === 'dockerfile'))
    return 'build'
  if (files.length > 0 && files.every(file => file.role === 'config' || /(?:^|\/)tsconfig(?:\.|$)|\.d\.ts$/i.test(file.path)))
    return 'build'
  if (files.every(file => file.role === 'docs'))
    return 'docs'
  if (paths.every(filePath => /\.(?:css|scss|less)$/i.test(filePath)))
    return 'style'
  if (files.every(file => file.role === 'test'))
    return 'test'
  if (files.some(packageVersionChanged) && files.every(file => file.role === 'dependency' || file.role === 'generated'))
    return 'chore'
  if (files.every(file => file.role === 'dependency' || file.role === 'generated'))
    return 'chore'
  return undefined
}

function renderProtection(snapshot: GitChangeSnapshot): string {
  const { metadataOnly, redacted } = protectionCounts(snapshot)
  return `protected metadata-only=${metadataOnly} redacted=${redacted}`
}

function renderScope(snapshot: GitChangeSnapshot, clusters: ChangeCluster[], concise = false): string {
  const states = new Map<string, number>()
  for (const file of snapshot.files) {
    const status = statusToken(file)
    states.set(status, (states.get(status) ?? 0) + 1)
  }
  const stateSummary = [...states]
    .toSorted(([left], [right]) => left.localeCompare(right))
    .map(([status, count]) => `${status}=${count}`)
    .join(' ')
  const capture = snapshot.scope === 'staged' ? 'index' : 'all-uncommitted'
  const typeHint = commitTypeHint(snapshot)
  if (concise) {
    const { metadataOnly, redacted } = protectionCounts(snapshot)
    const conciseStates = snapshot.files.length === 1 ? [...states.keys()][0] : stateSummary
    const protection = metadataOnly === 0 && redacted === 0 ? 'p=0' : `p=${metadataOnly}/${redacted}`
    const summary = `s=${capture === 'all-uncommitted' ? 'all' : capture} f=${snapshot.files.length} +${snapshot.totals.additions} -${snapshot.totals.deletions} c=${clusters.length} ${conciseStates} ${protection}`
    return typeHint ? `${summary} type=${typeHint}` : summary
  }
  const summary = `git-change-set=${capture} files=${snapshot.files.length} +${snapshot.totals.additions} -${snapshot.totals.deletions} clusters=${clusters.length} status=${stateSummary} ${renderProtection(snapshot)}`
  return typeHint ? `${summary} type=${typeHint}` : summary
}

function renderP0(snapshot: GitChangeSnapshot, clusters: ChangeCluster[], compact = false, concise = false): string[] {
  if (concise) {
    return [
      renderScope(snapshot, clusters, true),
      `c=${clusters.map(cluster => cluster.key).join(' ')}`,
    ]
  }
  const lines = [
    renderScope(snapshot, clusters),
  ]
  if (compact) {
    const prefixes = new Map<string, number>()
    for (const cluster of clusters) {
      const prefix = cluster.key.split('/', 1)[0] ?? '.'
      prefixes.set(prefix, (prefixes.get(prefix) ?? 0) + cluster.files.length)
    }
    lines.push(`c-summary ${[...prefixes].toSorted(([left], [right]) => left.localeCompare(right)).map(([key, count]) => `${key}=${count}`).join(' ')}`)
  }
  else {
    lines.push(...clusters.map(renderCluster))
  }
  return lines
}

function sortClustersForFacts(clusters: ChangeCluster[], facts: EvidenceFact[]): ChangeCluster[] {
  const priorityByCluster = new Map<string, number>()
  for (const cluster of clusters) {
    const factPriority = facts.filter(fact => fact.clusterKey === cluster.key).reduce((lowest, fact) => Math.min(lowest, fact.priority), 10)
    const rolePriority = Math.min(...cluster.files.map(file => roleRank(file.role)))
    priorityByCluster.set(cluster.key, Math.min(factPriority, rolePriority))
  }
  return clusters.toSorted((left, right) => (priorityByCluster.get(left.key) ?? 10) - (priorityByCluster.get(right.key) ?? 10)
    || left.key.localeCompare(right.key))
}

function factRank(item: EvidenceFact): number {
  if (item.text.startsWith('rename ') || item.text.startsWith('copy ') || item.text.startsWith('dependency ') || item.text.startsWith('package ') || item.text.startsWith('config '))
    return 1
  if (item.text.startsWith('symbol '))
    return 2
  if (item.text.startsWith('test '))
    return 3
  if (item.text.startsWith('docs '))
    return 4
  if (item.text.startsWith('context '))
    return 5
  if (/^(?:added|removed)\s+(?:async\s+)?(?:function|class|interface|type|enum|const)\b/.test(item.text))
    return 6
  if (/^(?:added|removed)\s+(?:return|throw)\b/.test(item.text))
    return 8
  return 7
}

function orderFactsForCluster(facts: EvidenceFact[]): EvidenceFact[] {
  const byFile = new Map<string, EvidenceFact[]>()
  for (const item of facts) {
    const fileFacts = byFile.get(item.filePath) ?? []
    fileFacts.push(item)
    byFile.set(item.filePath, fileFacts)
  }
  const queues = [...byFile.entries()]
    .map(([filePath, fileFacts]) => ({
      filePath,
      facts: fileFacts.toSorted((left, right) => factRank(left) - factRank(right)
        || (left.hunkId ?? '').localeCompare(right.hunkId ?? '')
        || left.id.localeCompare(right.id)),
    }))
    .toSorted((left, right) => left.filePath.localeCompare(right.filePath))
  const ordered: EvidenceFact[] = []
  while (queues.some(queue => queue.facts.length > 0)) {
    for (const queue of queues) {
      const item = queue.facts.shift()
      if (item)
        ordered.push(item)
    }
  }
  return ordered
}

function renderFact(item: EvidenceFact, concise = false): string {
  return concise ? `${item.filePath}: ${item.text}` : `  ${item.filePath}: ${item.text}`
}

export function compileEvidence(snapshot: GitChangeSnapshot, system: string): CompiledEvidence {
  const clusters = buildChangeClusters(snapshot)
  const facts = extractEvidenceFacts(clusters)
  const conciseP0 = snapshot.files.length <= 2
  const evidenceHeader = conciseP0 ? [] : ['COMMIT_EVIDENCE v1']
  const renderSelected = (item: EvidenceFact): string => renderFact(item, conciseP0)
  let p0 = [...evidenceHeader, ...renderP0(snapshot, clusters, false, conciseP0)]
  let p0Compacted = false
  if (estimateInputTokens(system, p0.join('\n')) > MAX_INPUT_TOKENS) {
    p0 = [...evidenceHeader, ...renderP0(snapshot, clusters, true, conciseP0)]
    p0Compacted = true
  }
  while (estimateInputTokens(system, p0.join('\n')) > MAX_INPUT_TOKENS && p0.length > 4) {
    p0.splice(3, 1)
    p0Compacted = true
  }

  const selected: EvidenceFact[] = []
  const orderedClusters = sortClustersForFacts(clusters, facts)
  const fits = (next: EvidenceFact[]): boolean => estimateInputTokens(system, [...p0, ...next.map(renderSelected)].join('\n')) <= MAX_INPUT_TOKENS
  const priorities = [1, 2, 3] as const
  for (const priority of priorities) {
    const pending = new Map<string, EvidenceFact[]>()
    for (const cluster of orderedClusters) {
      const priorityFacts = facts.filter(fact => fact.clusterKey === cluster.key && fact.priority === priority)
      const limit = priority === 3 ? new Set(priorityFacts.map(fact => fact.filePath)).size : undefined
      pending.set(cluster.key, orderFactsForCluster(priorityFacts).slice(0, limit))
    }
    while ([...pending.values()].some(queue => queue.length > 0)) {
      for (const cluster of orderedClusters) {
        const queue = pending.get(cluster.key)!
        const item = queue.shift()
        if (!item)
          continue
        const aboveTarget = estimateInputTokens(
          system,
          [...p0, ...selected.map(renderSelected)].join('\n'),
        ) >= TARGET_INPUT_TOKENS
        if (aboveTarget && priority > 1)
          continue
        if (fits([...selected, item]))
          selected.push(item)
      }
    }
  }
  const text = [...p0, ...selected.map(renderSelected)].join('\n')
  const coverage: EvidenceCoverage = {
    estimatedInputTokens: estimateInputTokens(system, text),
    representedClusters: clusters.length,
    totalClusters: clusters.length,
    includedFacts: selected.length,
    omittedFacts: facts.length - selected.length,
    contentCompacted: p0Compacted || selected.length < facts.length,
  }
  return { text, coverage, facts }
}
