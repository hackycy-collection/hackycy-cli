import type { TaskEditorResult } from './components/TaskEditor'
import process from 'node:process'
import ansis from 'ansis'
import { render } from 'ink'
import React from 'react'
import {
  getTaskGroup,
  listTaskGroups,
  removeTaskGroup,
  saveTaskGroup,
  validateTaskName,
} from '../../config/task'
import { clearScreen } from '../../shared/utils'
import { TaskConfirm } from '../task/components/TaskConfirm'
import { TaskPicker } from '../task/components/TaskPicker'
import { TaskEditor } from './components/TaskEditor'

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

function printRule(label?: string): void {
  const width = 72
  if (!label) {
    console.log(ansis.gray('─'.repeat(width)))
    return
  }

  const title = ` ${label} `
  console.log(ansis.gray(`─${title}${'─'.repeat(Math.max(0, width - title.length - 1))}`))
}

function printTaskListRow(name: string, commandCount: number): void {
  const taskName = clipText(name, 52)
  const count = commandCountLabel(commandCount)
  console.log(`${ansis.cyan(taskName.padEnd(52, ' '))} ${ansis.gray(count)}`)
}

function printCommandRow(index: number, command: string): void {
  const number = String(index + 1).padStart(2, '0')
  console.log(`${ansis.gray(number)}  ${clipText(command, 68)}`)
}

async function openTaskEditor(options: {
  title: string
  initialName: string
  initialCommands: string[]
  allowLoadExisting: boolean
}): Promise<TaskEditorResult> {
  let result: TaskEditorResult = { saved: false }
  const existingTasks = await listTaskGroups()

  const inst = render(
    React.createElement(TaskEditor, {
      ...options,
      existingTasks,
      onDone: (nextResult: TaskEditorResult) => {
        result = nextResult
      },
    }),
  )

  await inst.waitUntilExit()
  return result
}

async function openTaskPicker(title: string): Promise<string | null> {
  let result: string | null = null
  const tasks = await listTaskGroups()
  const names = Object.keys(tasks)

  if (names.length === 0) {
    console.log(ansis.dim('No tasks configured. Run "ycy config task add" to add one.'))
    return null
  }

  const inst = render(
    React.createElement(TaskPicker, {
      title,
      tasks,
      onDone: (name: string | null) => {
        result = name
      },
    }),
  )

  await inst.waitUntilExit()
  return result
}

async function openTaskConfirm(options: {
  title: string
  message: string
  detail?: string
  confirmLabel: string
}): Promise<boolean> {
  let result = false
  const inst = render(
    React.createElement(TaskConfirm, {
      ...options,
      onDone: (confirmed: boolean) => {
        result = confirmed
      },
    }),
  )

  await inst.waitUntilExit()
  return result
}

function printSaveResult(result: Extract<TaskEditorResult, { saved: true }>): void {
  if (result.previousName && result.previousName !== result.name) {
    console.log(`Task ${ansis.cyan(result.previousName)} renamed to ${ansis.cyan(result.name)}`)
    return
  }

  console.log(`Task ${ansis.cyan(result.name)} saved`)
}

export async function runTaskConfigAdd(): Promise<void> {
  clearScreen()

  const result = await openTaskEditor({
    title: 'Task Config - Add Task',
    initialName: '',
    initialCommands: [],
    allowLoadExisting: true,
  })

  if (!result.saved) {
    console.log(ansis.dim('Cancelled'))
    return
  }

  try {
    await saveTaskGroup(result.name, result.commands, { previousName: result.previousName })
    printSaveResult(result)
  }
  catch (err) {
    console.error(ansis.red((err as Error).message))
    process.exit(1)
  }
}

export async function runTaskConfigEdit(name?: string): Promise<void> {
  clearScreen()

  const selectedName = name ? validateTaskName(name) : await openTaskPicker('Task Config - Edit Task')
  if (!selectedName)
    return

  const task = await getTaskGroup(selectedName)
  if (!task) {
    console.error(ansis.red(`Task not found: ${selectedName}`))
    process.exit(1)
  }

  if (!name)
    clearScreen()

  const result = await openTaskEditor({
    title: 'Task Config - Edit Task',
    initialName: selectedName,
    initialCommands: task.commands,
    allowLoadExisting: false,
  })

  if (!result.saved) {
    console.log(ansis.dim('Cancelled'))
    return
  }

  try {
    await saveTaskGroup(result.name, result.commands, { previousName: result.previousName })
    printSaveResult(result)
  }
  catch (err) {
    console.error(ansis.red((err as Error).message))
    process.exit(1)
  }
}

export async function runTaskConfigRemove(name?: string): Promise<void> {
  clearScreen()

  const selectedName = name ? validateTaskName(name) : await openTaskPicker('Task Config - Remove Task')
  if (!selectedName)
    return

  const task = await getTaskGroup(selectedName)
  if (!task) {
    console.error(ansis.red(`Task not found: ${selectedName}`))
    process.exit(1)
  }

  if (!name)
    clearScreen()

  const confirmed = await openTaskConfirm({
    title: 'Task Config - Remove Task',
    message: `Remove task "${selectedName}"?`,
    detail: commandCountLabel(task.commands.length),
    confirmLabel: 'Remove',
  })

  if (!confirmed) {
    console.log(ansis.dim('Cancelled'))
    return
  }

  await removeTaskGroup(selectedName)

  console.log(`Task ${ansis.cyan(selectedName)} removed`)
}

export async function runTaskConfigList(name?: string): Promise<void> {
  clearScreen()
  console.log(ansis.bold.cyan('HACKYCY CLI') + ansis.gray(' / Task Config'))
  console.log()

  const groups = await listTaskGroups()

  if (name) {
    let taskName: string
    try {
      taskName = validateTaskName(name)
    }
    catch (err) {
      console.error(ansis.red((err as Error).message))
      process.exit(1)
    }

    const task = groups[taskName]
    if (!task) {
      console.error(ansis.red(`Task not found: ${taskName}`))
      process.exit(1)
    }

    console.log(ansis.bold.cyan(taskName))
    console.log(ansis.gray(commandCountLabel(task.commands.length)))
    console.log()
    printRule('Commands')
    for (const [index, command] of task.commands.entries())
      printCommandRow(index, command)
    printRule()
    console.log()
    return
  }

  const names = Object.keys(groups).sort((a, b) => a.localeCompare(b))
  if (names.length === 0) {
    console.log(ansis.dim('No tasks configured. Run "ycy config task add" to add one.'))
    return
  }

  console.log(ansis.bold('Tasks'))
  printRule()
  console.log(`${ansis.gray('Name'.padEnd(52, ' '))} ${ansis.gray('Commands')}`)
  for (const taskName of names)
    printTaskListRow(taskName, groups[taskName]!.commands.length)
  printRule()
  console.log(ansis.gray(`${names.length} task${names.length === 1 ? '' : 's'} configured`))
  console.log()
}
