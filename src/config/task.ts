import type { TaskConfig, TaskGroupConfig } from './types'
import { readConfig, writeConfig } from './store'

function emptyTaskConfig(): TaskConfig {
  return { groups: {} }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}

function normalizeTaskConfig(task: TaskConfig | undefined): TaskConfig {
  if (!task || !isRecord(task.groups))
    return emptyTaskConfig()

  const groups: Record<string, TaskGroupConfig> = {}
  for (const [name, group] of Object.entries(task.groups)) {
    const taskName = name.trim()
    if (!taskName || /\s/.test(taskName))
      continue
    if (!isRecord(group) || !Array.isArray(group.commands))
      continue

    const commands = group.commands
      .filter(command => typeof command === 'string')
      .map(command => command.trim())
      .filter(command => command && !/[\r\n]/.test(command))
    if (commands.length > 0)
      groups[taskName] = { commands }
  }

  return { groups }
}

export function normalizeTaskName(name: string): string {
  return name.trim()
}

export function getTaskNameError(name: string): string | undefined {
  const normalized = normalizeTaskName(name)
  if (!normalized)
    return 'Task name is required'
  if (/\s/.test(normalized))
    return 'Task name cannot contain whitespace'
}

export function validateTaskName(name: string): string {
  const error = getTaskNameError(name)
  if (error)
    throw new Error(error)
  return normalizeTaskName(name)
}

export function getTaskCommandError(command: string): string | undefined {
  const normalized = command.trim()
  if (!normalized)
    return 'Command is required'
  if (/[\r\n]/.test(command))
    return 'Command must be a single line'
}

export function validateTaskCommand(command: string): string {
  const error = getTaskCommandError(command)
  if (error)
    throw new Error(error)
  return command.trim()
}

export function validateTaskCommands(commands: string[]): string[] {
  const normalized = commands.map(command => validateTaskCommand(command))
  if (normalized.length === 0)
    throw new Error('Task must contain at least one command')
  return normalized
}

export async function readTaskConfig(): Promise<TaskConfig> {
  const config = await readConfig()
  return normalizeTaskConfig(config.task)
}

export async function listTaskGroups(): Promise<Record<string, TaskGroupConfig>> {
  const task = await readTaskConfig()
  return task.groups
}

export async function getTaskGroup(name: string): Promise<TaskGroupConfig | null> {
  const taskName = validateTaskName(name)
  const task = await readTaskConfig()
  return task.groups[taskName] ?? null
}

export async function saveTaskGroup(
  name: string,
  commands: string[],
  options: { previousName?: string } = {},
): Promise<void> {
  const taskName = validateTaskName(name)
  const previousName = options.previousName ? validateTaskName(options.previousName) : undefined
  const taskCommands = validateTaskCommands(commands)

  const config = await readConfig()
  const task = normalizeTaskConfig(config.task)

  if (previousName && previousName !== taskName) {
    if (!task.groups[previousName])
      throw new Error(`Task not found: ${previousName}`)
    if (task.groups[taskName])
      throw new Error(`Task already exists: ${taskName}`)
    delete task.groups[previousName]
  }
  else if (!previousName && task.groups[taskName]) {
    throw new Error(`Task already exists: ${taskName}`)
  }

  task.groups[taskName] = { commands: taskCommands }
  config.task = task
  await writeConfig(config)
}

export async function removeTaskGroup(name: string): Promise<boolean> {
  const taskName = validateTaskName(name)
  const config = await readConfig()
  const task = normalizeTaskConfig(config.task)

  if (!task.groups[taskName])
    return false

  delete task.groups[taskName]
  config.task = task
  await writeConfig(config)
  return true
}
