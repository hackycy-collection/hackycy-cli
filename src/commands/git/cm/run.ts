import type { ChatCompletionTokenUsage } from '../../../config/client'
import type { ResolvedCmProfile } from '../../../config/types'
import type { ChangeSummary, CmOptions, CommitLanguage } from './types'
import process from 'node:process'
import * as p from '@clack/prompts'
import ansis from 'ansis'
import { createChatCompletionWithUsage } from '../../../config/client'
import { resolveCmProfile } from '../../../config/cm'
import { printTitle } from '../../../shared/utils'
import {
  collectChangeSummary,
  commitWithMessage,
  getRecentCommitSubjects,
  hasStagedChanges,
  pushChanges,
  stageAllChanges,
  stageFiles,
  unstageFiles,
} from './changes'

function normalizeLanguage(lang: string | undefined): CommitLanguage {
  if (!lang)
    return 'en'
  if (lang !== 'en' && lang !== 'zh')
    throw new Error('Unsupported language. Use "en" or "zh".')
  return lang
}

function cleanCommitMessage(content: string): string {
  return content
    .trim()
    .replace(/^```(?:text)?/, '')
    .replace(/```$/, '')
    .trim()
    .replace(/^["']|["']$/g, '')
}

function formatTokenCount(value: number | undefined): string {
  return value === undefined ? 'unknown' : value.toLocaleString('en-US')
}

function formatTokenUsage(usage: ChatCompletionTokenUsage | undefined): string {
  if (!usage)
    return 'Token usage: unavailable from provider'

  return [
    'Token usage:',
    `prompt ${formatTokenCount(usage.promptTokens)}`,
    `completion ${formatTokenCount(usage.completionTokens)}`,
    `total ${formatTokenCount(usage.totalTokens)}`,
  ].join(' ')
}

function buildMessages(
  summary: ChangeSummary,
  lang: CommitLanguage,
  history: string[],
  includeBody: boolean,
): Parameters<typeof createChatCompletionWithUsage>[1] {
  const language = lang === 'zh' ? 'Chinese' : 'English'
  const historyText = history.length > 0
    ? `\nRecent commit subjects for style reference:\n${history.map(item => `- ${item}`).join('\n')}\n`
    : ''
  const bodyRule = includeBody
    ? 'You may include a short body only when it adds useful context.'
    : 'Return one line only.'

  return [
    {
      role: 'system',
      content: [
        'You generate concise Angular-style git commit messages.',
        'Return only the commit message. No markdown. No explanations.',
        'Use format: type(scope): subject.',
        'Allowed types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.',
        bodyRule,
      ].join(' '),
    },
    {
      role: 'user',
      content: [
        `Language: ${language}`,
        'Infer a short scope from changed paths.',
        'Prefer the most important user-facing or code behavior change.',
        historyText,
        summary.promptText,
      ].join('\n'),
    },
  ]
}

interface GeneratedCommitMessage {
  message: string
  tokenUsage?: ChatCompletionTokenUsage
}

async function generateCommitMessage(
  profile: ResolvedCmProfile,
  summary: ChangeSummary,
  options: CmOptions,
): Promise<GeneratedCommitMessage> {
  const history = options.history ? await getRecentCommitSubjects(summary.repoRoot) : []
  const messages = buildMessages(summary, normalizeLanguage(options.lang), history, Boolean(options.body))
  const result = await createChatCompletionWithUsage(profile, messages)
  return {
    message: cleanCommitMessage(result.content),
    tokenUsage: result.usage,
  }
}

function printGeneratedMessage(
  message: string,
  summary: ChangeSummary,
  profile: ResolvedCmProfile,
  tokenUsage: ChatCompletionTokenUsage | undefined,
): void {
  const lines = [
    ansis.green(message),
    '',
    ansis.bold('Changed files:'),
    ...summary.files.map(file => ansis.dim(file.status)),
    '',
    ansis.dim(`CM profile: ${profile.name} (${profile.model})`),
    ansis.dim(formatTokenUsage(tokenUsage)),
  ]

  if (summary.truncated)
    lines.push(ansis.yellow('Some diffs were omitted or truncated to save tokens.'))

  p.note(lines.join('\n'), 'Generated commit message')
}

async function promptForStageFiles(summary: ChangeSummary): Promise<void> {
  if (summary.files.length === 0) {
    p.log.info('No uncommitted changes.')
    return
  }

  const selected = await p.multiselect({
    message: 'Select files to stage',
    options: summary.files.map(file => ({
      value: file.path,
      label: file.status,
    })),
    initialValues: summary.files.map(file => file.path),
  })

  if (p.isCancel(selected)) {
    p.cancel('Cancelled')
    process.exit(0)
  }

  const selectedPaths = selected as string[]
  if (selectedPaths.length === 0) {
    p.cancel('Nothing selected.')
    process.exit(0)
  }

  const selectedSet = new Set(selectedPaths)
  const unselectedPaths = summary.files
    .filter(file => file.indexStatus !== '?')
    .map(file => file.path)
    .filter(filePath => !selectedSet.has(filePath))

  const stageSpin = p.spinner()
  stageSpin.start('Staging selected changes...')
  try {
    if (unselectedPaths.length > 0)
      await unstageFiles(summary.repoRoot, unselectedPaths)
    await stageFiles(summary.repoRoot, selectedPaths)
    stageSpin.stop('Staged selected changes')
  }
  catch (err) {
    stageSpin.stop('Failed to stage selected changes')
    p.log.error((err as Error).message)
    process.exit(1)
  }
}

export async function runGitCm(options: CmOptions): Promise<void> {
  printTitle()
  p.intro(ansis.cyan('Git Commit Message'))

  if (options.stage && options.stageAll) {
    p.log.error('Use either --stage or --stage-all, not both.')
    process.exit(1)
  }

  if (options.stage && options.dryRun) {
    p.log.error('Use either --stage or --dry-run, not both.')
    process.exit(1)
  }

  if (options.push && options.dryRun) {
    p.log.error('Use either --push or --dry-run, not both.')
    process.exit(1)
  }

  if (options.push && !options.stage && !options.staged && !options.stageAll) {
    p.log.error('Use --push with --stage, --staged, or --stage-all.')
    process.exit(1)
  }

  const shouldPromptStage = Boolean(options.stage)
  const shouldStageAll = Boolean(options.stageAll && !options.dryRun)
  const stagedOnly = Boolean(options.staged || shouldPromptStage || shouldStageAll)
  const shouldCreateCommit = stagedOnly && !options.dryRun
  const shouldPush = Boolean(options.push && shouldCreateCommit)
  const pushRemote = typeof options.push === 'string' ? options.push : undefined

  if (shouldPromptStage) {
    const spin = p.spinner()
    spin.start('Collecting git changes...')
    try {
      const summary = await collectChangeSummary()
      spin.stop('Git changes collected')
      await promptForStageFiles(summary)
    }
    catch (err) {
      spin.stop('Failed to collect git changes')
      p.log.error((err as Error).message)
      process.exit(1)
    }
  }

  if (shouldStageAll) {
    const stageSpin = p.spinner()
    stageSpin.start('Staging all changes...')
    try {
      const repoSummary = await collectChangeSummary()
      await stageAllChanges(repoSummary.repoRoot)
      stageSpin.stop('Staged all changes')
    }
    catch (err) {
      stageSpin.stop('Failed to stage changes')
      p.log.error((err as Error).message)
      process.exit(1)
    }
  }

  const spin = p.spinner()
  spin.start('Collecting git changes...')

  let summary: ChangeSummary
  try {
    summary = await collectChangeSummary({ stagedOnly })
  }
  catch (err) {
    spin.stop('Failed to collect git changes')
    p.log.error((err as Error).message)
    process.exit(1)
  }

  if (summary.files.length === 0) {
    spin.stop(stagedOnly
      ? 'No staged changes.'
      : 'No uncommitted changes.')
    return
  }

  let profile: ResolvedCmProfile
  try {
    profile = await resolveCmProfile(options.profile)
  }
  catch (err) {
    spin.stop('CM profile not configured')
    p.log.error((err as Error).message)
    process.exit(1)
  }

  spin.message(`Generating commit message with ${profile.name}...`)

  let generated: GeneratedCommitMessage
  try {
    generated = await generateCommitMessage(profile, summary, options)
  }
  catch (err) {
    spin.stop('Provider request failed')
    p.log.error((err as Error).message)
    p.log.info(`Provider: ${profile.name}`)
    p.log.info(`Base URL: ${profile.baseURL}`)
    p.log.info(`Model: ${profile.model}`)
    p.log.info('If the response was empty, the provider likely returned no assistant content.')
    process.exit(1)
  }

  spin.stop('Commit message generated')
  printGeneratedMessage(generated.message, summary, profile, generated.tokenUsage)

  if (!shouldCreateCommit) {
    p.outro(options.dryRun ? 'Dry run complete' : 'Done')
    return
  }

  if (!(await hasStagedChanges(summary.repoRoot))) {
    p.log.warn('No staged changes.')
    p.outro('No commit created')
    return
  }

  const confirmed = await p.confirm({
    message: 'Create commit with this message?',
  })

  if (p.isCancel(confirmed) || !confirmed) {
    p.outro('Cancelled')
    return
  }

  const commitSpin = p.spinner()
  commitSpin.start('Creating commit...')
  try {
    await commitWithMessage(summary.repoRoot, generated.message)
    commitSpin.stop('Commit created')
  }
  catch (err) {
    commitSpin.stop('git commit failed')
    p.log.error((err as Error).message)
    process.exit(1)
  }

  if (!shouldPush) {
    p.outro(ansis.green('Done'))
    return
  }

  const pushSpin = p.spinner()
  pushSpin.start('Pushing to remote...')
  try {
    await pushChanges(summary.repoRoot, pushRemote)
    pushSpin.stop('Pushed to remote')
    p.outro(ansis.green('Done'))
  }
  catch (err) {
    pushSpin.stop('git push failed')
    p.log.error((err as Error).message)
    process.exit(1)
  }
}
