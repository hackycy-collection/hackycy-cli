import { readdir, readFile, stat } from 'node:fs/promises'
import { dirname, relative, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(fileURLToPath(new URL('.', import.meta.url)), 'dist')
const shells = ['diff/index.html', 'fs/index.html', 'tunnel-server/index.html']

async function exists(path) {
  try {
    await stat(path)
    return true
  }
  catch {
    return false
  }
}

async function filesBelow(directory) {
  const entries = await readdir(directory, { withFileTypes: true })
  const files = []
  for (const entry of entries) {
    const path = resolve(directory, entry.name)
    if (entry.isDirectory())
      files.push(...await filesBelow(path))
    else if (entry.isFile())
      files.push(path)
  }
  return files
}

function outputPath(reference, sourcePath = root) {
  const pathReference = reference.split(/[?#]/, 1)[0]
  let output
  if (pathReference.startsWith('/assets/'))
    output = resolve(root, `.${pathReference}`)
  else if (pathReference.startsWith('./') || pathReference.startsWith('../'))
    output = resolve(dirname(sourcePath), pathReference)
  else
    return null

  const outputRelative = relative(root, output)
  if (outputRelative === '' || outputRelative.startsWith(`..${sep}`) || outputRelative === '..')
    throw new Error(`generated reference escapes dist: ${reference}`)
  return output
}

function manifestEntry(manifest, shell) {
  return Object.entries(manifest).find(([key, value]) => key === shell || value.src === shell)?.[1]
}

function collectEntryFiles(manifest, entry, seen, files) {
  if (seen.has(entry))
    return
  seen.add(entry)
  const item = manifest[entry]
  if (!item)
    throw new Error(`manifest references missing entry: ${entry}`)
  files.add(item.file)
  for (const file of item.css ?? [])
    files.add(file)
  for (const file of item.assets ?? [])
    files.add(file)
  for (const imported of [...(item.imports ?? []), ...(item.dynamicImports ?? [])])
    collectEntryFiles(manifest, imported, seen, files)
}

async function collectWorkerFiles(entryPath, seen, files) {
  if (seen.has(entryPath) || !entryPath.endsWith('.js'))
    return
  seen.add(entryPath)

  const source = await readFile(entryPath, 'utf8')
  const workerReferences = /new\s+Worker\s*\(\s*(?:new\s+URL\s*\(\s*)?(["'`])([^"'`]+)\1/g
  for (const match of source.matchAll(workerReferences)) {
    const workerPath = outputPath(match[2], entryPath)
    if (!workerPath)
      continue
    if (!await exists(workerPath))
      throw new Error(`generated worker reference is missing: ${match[2]}`)
    const workerRelative = relative(root, workerPath).split(sep).join('/')
    files.add(workerRelative)
    await collectWorkerFiles(workerPath, seen, files)
  }
}

if (!await exists(root))
  throw new Error('missing Vite output: web/dist')

const manifestPath = resolve(root, '.vite/manifest.json')
const manifest = JSON.parse(await readFile(manifestPath, 'utf8'))
const referenced = new Set()
const seen = new Set()

for (const shell of shells) {
  const shellPath = resolve(root, shell)
  if (!await exists(shellPath))
    throw new Error(`missing required shell: dist/${shell}`)

  const html = await readFile(shellPath, 'utf8')
  for (const match of html.matchAll(/\b(?:src|href)=["']([^"']+)["']/g)) {
    const assetPath = outputPath(match[1])
    if (assetPath && !await exists(assetPath))
      throw new Error(`shell ${shell} references missing asset: ${match[1]}`)
  }

  const entry = manifestEntry(manifest, shell)
  if (!entry)
    throw new Error(`manifest is missing shell entry: ${shell}`)
  collectEntryFiles(manifest, Object.keys(manifest).find(key => manifest[key] === entry), seen, referenced)
}

const workerScanSeen = new Set()
for (const asset of [...referenced])
  await collectWorkerFiles(resolve(root, asset), workerScanSeen, referenced)

const assets = await filesBelow(resolve(root, 'assets'))
for (const asset of assets) {
  const relative = asset.slice(root.length + 1).split(sep).join('/')
  if (relative.endsWith('.map'))
    throw new Error(`source map was emitted: ${relative}`)
  if (!referenced.has(relative))
    throw new Error(`unreferenced generated asset: ${relative}`)
}

console.log(`verified ${shells.length} Vite shells and ${assets.length} reachable generated assets`)
