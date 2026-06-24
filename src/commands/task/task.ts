import process from 'node:process'
import ansis from 'ansis'
import { render } from 'ink'
import React from 'react'
import { getTaskGroup, listTaskGroups, validateTaskName } from '../../config/task'
import { clearScreen } from '../../shared/utils'
import { TaskPicker } from './components/TaskPicker'

interface RunTaskOptions {
  delay?: string
}

function parseDelayMs(value: string | undefined): number {
  const raw = value ?? '1'
  const seconds = Number(raw)
  if (!Number.isFinite(seconds))
    throw new Error('Delay must be a number')
  if (seconds < 0)
    throw new Error('Delay cannot be negative')
  return seconds * 1000
}

async function selectTaskName(): Promise<string> {
  const groups = await listTaskGroups()
  const names = Object.keys(groups).sort((a, b) => a.localeCompare(b))

  if (names.length === 0) {
    console.log(ansis.dim('No tasks configured. Run "ycy config task add" to add one.'))
    process.exit(1)
  }

  let selected: string | null = null
  const inst = render(
    React.createElement(TaskPicker, {
      title: 'Task Run / Select Task',
      tasks: groups,
      onDone: (name: string | null) => {
        selected = name
      },
    }),
  )

  await inst.waitUntilExit()

  if (!selected) {
    console.log(ansis.dim('Cancelled'))
    process.exit(0)
  }

  return selected
}

async function runShellCommand(command: string): Promise<number> {
  const result = await Bun.$`${{ raw: command }}`.nothrow()
  return result.exitCode
}

async function wait(ms: number): Promise<void> {
  if (ms <= 0)
    return
  await new Promise(resolve => setTimeout(resolve, ms))
}

function commandCountLabel(count: number): string {
  return `${count} command${count === 1 ? '' : 's'}`
}

function clipText(value: string, maxWidth: number): string {
  if (value.length <= maxWidth)
    return value
  if (maxWidth <= 1)
    return value.slice(0, maxWidth)
  return `${value.slice(0, maxWidth - 1)}…`
}

function printRunHeader(taskName: string, commandCount: number): void {
  console.log(ansis.bold.cyan('HACKYCY CLI'))
  console.log()
  console.log(ansis.bold(taskName))
  console.log(ansis.gray(commandCountLabel(commandCount)))
}

function printCommandBlock(index: number, total: number, command: string): void {
  const label = `${String(index + 1).padStart(2, '0')} / ${String(total).padStart(2, '0')}`
  const width = 68
  const title = ` ${label} `
  const titleFill = Math.max(0, width - title.length - 1)
  const content = ` ${clipText(command, width - 4)}`

  console.log()
  console.log(ansis.gray(`┌${title}${'─'.repeat(titleFill)}┐`))
  console.log(ansis.gray('│') + ansis.bold.cyan(content.padEnd(width - 2, ' ')) + ansis.gray('│'))
  console.log(ansis.gray(`└${'─'.repeat(width - 2)}┘`))
  console.log()
}

function formatDelay(ms: number): string {
  return `${ms / 1000}s`
}

function printDelay(ms: number): void {
  if (ms <= 0)
    return
  console.log(ansis.gray(`Waiting ${formatDelay(ms)} before next command...`))
}

function printTaskFailure(taskName: string, index: number, exitCode: number, command: string): void {
  console.log()
  console.error(ansis.bold.red('Task failed'))
  console.error(ansis.red(taskName))
  console.error(ansis.gray(`Command ${String(index + 1).padStart(2, '0')} exited with code ${exitCode}`))
  console.error(command)
}

function printTaskSuccess(taskName: string, commandCount: number): void {
  console.log()
  console.log(ansis.bold.green('Task completed'))
  console.log(ansis.green(taskName))
  console.log(ansis.gray(`${commandCountLabel(commandCount)} finished`))
}

export async function runTask(name?: string, options: RunTaskOptions = {}): Promise<void> {
  clearScreen()

  let taskName: string
  let delayMs: number
  try {
    delayMs = parseDelayMs(options.delay)
    taskName = name ? validateTaskName(name) : await selectTaskName()
  }
  catch (err) {
    console.error(ansis.red((err as Error).message))
    process.exit(1)
  }

  if (!name)
    clearScreen()

  const group = await getTaskGroup(taskName)
  if (!group) {
    console.error(ansis.red(`Task not found: ${taskName}`))
    process.exit(1)
  }

  printRunHeader(taskName, group.commands.length)
  for (const [index, command] of group.commands.entries()) {
    printCommandBlock(index, group.commands.length, command)

    const exitCode = await runShellCommand(command)
    if (exitCode !== 0) {
      printTaskFailure(taskName, index, exitCode, command)
      process.exit(exitCode)
    }

    if (index < group.commands.length - 1) {
      printDelay(delayMs)
      await wait(delayMs)
    }
  }

  printTaskSuccess(taskName, group.commands.length)
}
