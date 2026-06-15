import type { ChangeSummary, CmOptions, CommitLanguage, ResolvedAiProfile } from './types'
import process from 'node:process'
import * as p from '@clack/prompts'
import ansis from 'ansis'
import { printTitle } from '../../../shared/utils'
import { createChatCompletion } from './client'
import { resolveAiProfile } from './config'
import {
  collectChangeSummary,
  commitWithMessage,
  getRecentCommitSubjects,
  hasStagedChanges,
  stageAllChanges,
} from './diff'

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

function buildMessages(
  summary: ChangeSummary,
  lang: CommitLanguage,
  history: string[],
  includeBody: boolean,
): Parameters<typeof createChatCompletion>[1] {
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

async function generateCommitMessage(
  profile: ResolvedAiProfile,
  summary: ChangeSummary,
  options: CmOptions,
): Promise<string> {
  const history = options.history ? await getRecentCommitSubjects(summary.repoRoot) : []
  const messages = buildMessages(summary, normalizeLanguage(options.lang), history, Boolean(options.body))
  const content = await createChatCompletion(profile, messages)
  return cleanCommitMessage(content)
}

function printGeneratedMessage(message: string, summary: ChangeSummary, profile: ResolvedAiProfile): void {
  console.log()
  console.log(ansis.bold('Generated commit message:'))
  console.log()
  console.log(ansis.green(message))
  console.log()
  console.log(ansis.bold('Changed files:'))
  for (const file of summary.files)
    console.log(ansis.dim(file.status))
  console.log()
  console.log(ansis.dim(`AI profile: ${profile.name} (${profile.model})`))
  if (summary.truncated)
    console.log(ansis.yellow('Some diffs were omitted or truncated to save tokens.'))
}

export async function runGitAiCm(options: CmOptions): Promise<void> {
  printTitle()
  p.intro(ansis.cyan('Git AI Commit Message'))

  const shouldStageAll = Boolean(options.stageAll && options.commit && !options.dryRun)
  const stagedOnly = Boolean(options.staged || (options.commit && !options.stageAll))

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
      ? 'No staged changes. Use --stage-all with --commit to stage all changes before committing.'
      : 'No uncommitted changes.')
    return
  }

  let profile: ResolvedAiProfile
  try {
    profile = await resolveAiProfile(options.profile)
  }
  catch (err) {
    spin.stop('AI profile not configured')
    p.log.error((err as Error).message)
    process.exit(1)
  }

  spin.message(`Generating commit message with ${profile.name}...`)

  let message: string
  try {
    message = await generateCommitMessage(profile, summary, options)
  }
  catch (err) {
    spin.stop('AI request failed')
    p.log.error((err as Error).message)
    p.log.info(`Provider: ${profile.name}`)
    p.log.info(`Base URL: ${profile.baseURL}`)
    p.log.info(`Model: ${profile.model}`)
    process.exit(1)
  }

  spin.stop('Commit message generated')
  printGeneratedMessage(message, summary, profile)

  if (!options.commit || options.dryRun) {
    p.outro(options.dryRun ? 'Dry run complete' : 'Done')
    return
  }

  if (!(await hasStagedChanges(summary.repoRoot))) {
    p.log.warn('No staged changes. Use --stage-all with --commit to stage all changes before committing.')
    p.outro('No commit created')
    return
  }

  const confirmed = await p.confirm({
    message: 'Commit staged changes with this message?',
  })

  if (p.isCancel(confirmed) || !confirmed) {
    p.outro('Cancelled')
    return
  }

  const commitSpin = p.spinner()
  commitSpin.start('Creating commit...')
  try {
    await commitWithMessage(summary.repoRoot, message)
    commitSpin.stop('Commit created')
    p.outro(ansis.green('Done'))
  }
  catch (err) {
    commitSpin.stop('git commit failed')
    p.log.error((err as Error).message)
    process.exit(1)
  }
}
