import type { TaskGroupConfig } from '../../../config/types'
import { Box, Text, useApp, useInput } from 'ink'
import React, { useState } from 'react'
import {
  getTaskCommandError,
  getTaskNameError,
  normalizeTaskName,
  validateTaskCommands,
} from '../../../config/task'

export type TaskEditorResult
  = | { saved: false }
    | { saved: true, name: string, commands: string[], previousName?: string }

interface TaskEditorProps {
  title: string
  initialName: string
  initialCommands: string[]
  existingTasks: Record<string, TaskGroupConfig>
  allowLoadExisting: boolean
  onDone: (result: TaskEditorResult) => void
}

interface TextInputState {
  value: string
  cursor: number
  error?: string
}

interface NameInputState extends TextInputState {
  kind: 'initial' | 'rename'
}

interface CommandInputState extends TextInputState {
  index: number | null
}

interface InputKey {
  leftArrow: boolean
  rightArrow: boolean
  home: boolean
  end: boolean
  backspace: boolean
  delete: boolean
  ctrl: boolean
  meta: boolean
  tab: boolean
}

type EditorMode = 'list' | 'name' | 'command'

const STUDIO_WIDTH = 96
const PANEL_WIDTH = 96
const PREVIEW_BOX_WIDTH = 50
const PREVIEW_TEXT_WIDTH = 42
const NAME_INPUT_WIDTH = 82
const COMMAND_INPUT_WIDTH = 78

function updateTextInput<T extends TextInputState>(state: T, input: string, key: InputKey): T {
  if (key.leftArrow)
    return { ...state, cursor: Math.max(0, state.cursor - 1), error: undefined }
  if (key.rightArrow)
    return { ...state, cursor: Math.min(state.value.length, state.cursor + 1), error: undefined }
  if (key.home)
    return { ...state, cursor: 0, error: undefined }
  if (key.end)
    return { ...state, cursor: state.value.length, error: undefined }
  if (key.backspace) {
    if (state.cursor === 0)
      return { ...state, error: undefined }
    return {
      ...state,
      value: `${state.value.slice(0, state.cursor - 1)}${state.value.slice(state.cursor)}`,
      cursor: state.cursor - 1,
      error: undefined,
    }
  }
  if (key.delete) {
    if (state.cursor >= state.value.length)
      return { ...state, error: undefined }
    return {
      ...state,
      value: `${state.value.slice(0, state.cursor)}${state.value.slice(state.cursor + 1)}`,
      cursor: state.cursor,
      error: undefined,
    }
  }
  if (!input || key.ctrl || key.meta || key.tab)
    return state

  const text = input.replace(/[\r\n]/g, '')
  if (!text)
    return state

  return {
    ...state,
    value: `${state.value.slice(0, state.cursor)}${text}${state.value.slice(state.cursor)}`,
    cursor: state.cursor + text.length,
    error: undefined,
  }
}

function clipText(value: string, maxWidth: number): string {
  if (value.length <= maxWidth)
    return value
  if (maxWidth <= 1)
    return value.slice(0, maxWidth)
  return `${value.slice(0, maxWidth - 1)}…`
}

function visibleInput(value: string, cursor: number, maxWidth: number): { text: string, cursor: number } {
  if (value.length <= maxWidth)
    return { text: value, cursor }

  const half = Math.floor(maxWidth / 2)
  const start = Math.max(0, Math.min(cursor - half, value.length - maxWidth))
  return {
    text: value.slice(start, start + maxWidth),
    cursor: cursor - start,
  }
}

function commandCountLabel(count: number): string {
  return `${count} command${count === 1 ? '' : 's'}`
}

function Header({ title }: { title: string }) {
  return (
    <Box width={STUDIO_WIDTH} justifyContent="space-between">
      <Box>
        <Text bold color="cyan">HACKYCY CLI</Text>
        <Text color="gray"> / </Text>
        <Text bold>{title}</Text>
      </Box>
      <Text color="gray">Task Studio</Text>
    </Box>
  )
}

function Panel({ title, width, children, borderColor = 'gray' }: { title: string, width: number, children: React.ReactNode, borderColor?: string }) {
  return (
    <Box
      borderStyle="round"
      borderColor={borderColor}
      flexDirection="column"
      minHeight={16}
      paddingX={1}
      width={width}
    >
      <Text bold color={borderColor === 'gray' ? 'white' : borderColor}>{title}</Text>
      <Text> </Text>
      {children}
    </Box>
  )
}

function KeyBar({ mode }: { mode: EditorMode }) {
  const keys = mode === 'list'
    ? [
        ['↑/↓', 'select'],
        ['a', 'add'],
        ['e', 'edit'],
        ['x', 'delete'],
        ['u/d', 'move'],
        ['r', 'rename'],
        ['s', 'save'],
        ['q', 'cancel'],
      ]
    : [
        ['Enter', 'save'],
        ['Esc', 'discard'],
      ]

  return (
    <Box marginTop={1}>
      {keys.map(([key, label], index) => (
        <React.Fragment key={key}>
          {index > 0 ? <Text color="gray">   </Text> : null}
          <Text bold color="cyan">{key}</Text>
          <Text color="gray">
            {' '}
            {label}
          </Text>
        </React.Fragment>
      ))}
    </Box>
  )
}

function InputLine({ value, cursor, placeholder, width }: { value: string, cursor: number, placeholder: string, width: number }) {
  if (!value) {
    return (
      <Box width={width}>
        <Text inverse> </Text>
        <Text color="gray">
          {' '}
          {clipText(placeholder, width - 2)}
        </Text>
      </Box>
    )
  }

  const visible = visibleInput(value, cursor, width)
  const before = visible.text.slice(0, visible.cursor)
  const current = visible.text[visible.cursor] ?? ' '
  const after = visible.cursor < visible.text.length ? visible.text.slice(visible.cursor + 1) : ''

  return (
    <Box width={width}>
      <Text>
        {before}
        <Text inverse>{current}</Text>
        {after}
      </Text>
    </Box>
  )
}

function CommandPanel({
  name,
  commands,
  selectedIndex,
  error,
  mode,
  nameInput,
  commandInput,
}: {
  name: string
  commands: string[]
  selectedIndex: number
  error?: string
  mode: EditorMode
  nameInput: NameInputState
  commandInput: CommandInputState
}) {
  const selectedCommand = commands[selectedIndex]
  const inputError = mode === 'name' ? nameInput.error : mode === 'command' ? commandInput.error : undefined
  const previewCommand = mode === 'command'
    ? commandInput.value.trim() || (commandInput.index === null ? 'New command' : selectedCommand)
    : selectedCommand
  let status = error || inputError
  if (!status) {
    status = mode === 'name'
      ? 'Editing task name'
      : mode === 'command'
        ? commandInput.index === null ? 'Adding command' : `Editing command ${commandInput.index + 1}`
        : commands.length === 0 ? 'Add the first command before saving' : `${commandCountLabel(commands.length)} ready`
  }

  return (
    <Panel title="Command Group" width={PANEL_WIDTH} borderColor="magenta">
      <Box flexDirection="column">
        <Text color="gray">Name</Text>
        {mode === 'name'
          ? <InputLine value={nameInput.value} cursor={nameInput.cursor} placeholder="release" width={NAME_INPUT_WIDTH} />
          : <Text bold>{name || '(new task)'}</Text>}
      </Box>
      <Text> </Text>
      <Text color="gray">Commands</Text>
      {commands.length === 0 && !(mode === 'command' && commandInput.index === null)
        ? (
            <Box flexDirection="column" marginTop={1} marginBottom={1}>
              <Text color="gray">No commands yet</Text>
              <Text color="gray">Press a to add the first command</Text>
            </Box>
          )
        : commands.map((command, index) => {
            const selected = index === selectedIndex
            const number = String(index + 1).padStart(2, '0')
            const isEditing = mode === 'command' && commandInput.index === index
            return (
              <Box key={`${index}:${command}`}>
                <Text color={selected ? 'cyan' : 'gray'}>{selected ? '▸ ' : '  '}</Text>
                <Text color="gray">{number}</Text>
                <Text color="gray">  </Text>
                {isEditing
                  ? <InputLine value={commandInput.value} cursor={commandInput.cursor} placeholder="pnpm build" width={COMMAND_INPUT_WIDTH} />
                  : <Text inverse={selected}>{clipText(command, COMMAND_INPUT_WIDTH)}</Text>}
              </Box>
            )
          })}
      {mode === 'command' && commandInput.index === null
        ? (
            <Box>
              <Text color="cyan">▸ </Text>
              <Text color="gray">++  </Text>
              <InputLine value={commandInput.value} cursor={commandInput.cursor} placeholder="pnpm build" width={COMMAND_INPUT_WIDTH} />
            </Box>
          )
        : null}
      <Text> </Text>
      <Box borderStyle="single" borderColor="gray" flexDirection="column" paddingX={1} width={PREVIEW_BOX_WIDTH}>
        <Text bold color="gray">Selected</Text>
        <Text wrap="wrap">{clipText(previewCommand || 'No command selected', PREVIEW_TEXT_WIDTH)}</Text>
      </Box>
      <Box marginTop={1}>
        <Text color={error || inputError ? 'red' : commands.length === 0 ? 'yellow' : 'green'}>{status}</Text>
      </Box>
    </Panel>
  )
}

export function TaskEditor({
  title,
  initialName,
  initialCommands,
  existingTasks,
  allowLoadExisting,
  onDone,
}: TaskEditorProps) {
  const { exit } = useApp()
  const [mode, setMode] = useState<EditorMode>(initialName ? 'list' : 'name')
  const [sourceName, setSourceName] = useState<string | undefined>(initialName || undefined)
  const [name, setName] = useState(initialName)
  const [commands, setCommands] = useState(initialCommands)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [error, setError] = useState<string | undefined>()
  const [confirmExistingName, setConfirmExistingName] = useState<string | undefined>()
  const [nameInput, setNameInput] = useState<NameInputState>({
    kind: initialName ? 'rename' : 'initial',
    value: initialName,
    cursor: initialName.length,
  })
  const [commandInput, setCommandInput] = useState<CommandInputState>({
    index: null,
    value: '',
    cursor: 0,
  })

  function finish(result: TaskEditorResult): void {
    onDone(result)
    exit()
  }

  function startRename(): void {
    setNameInput({
      kind: 'rename',
      value: name,
      cursor: name.length,
    })
    setConfirmExistingName(undefined)
    setError(undefined)
    setMode('name')
  }

  function startAddCommand(): void {
    setCommandInput({
      index: null,
      value: '',
      cursor: 0,
    })
    setError(undefined)
    setMode('command')
  }

  function startEditCommand(): void {
    const current = commands[selectedIndex]
    if (!current)
      return

    setCommandInput({
      index: selectedIndex,
      value: current,
      cursor: current.length,
    })
    setError(undefined)
    setMode('command')
  }

  function deleteSelectedCommand(): void {
    if (commands.length === 0)
      return

    setCommands(current => current.filter((_, index) => index !== selectedIndex))
    setSelectedIndex(index => Math.max(0, Math.min(index, commands.length - 2)))
    setError(undefined)
  }

  function moveSelectedCommand(offset: -1 | 1): void {
    const nextIndex = selectedIndex + offset
    if (nextIndex < 0 || nextIndex >= commands.length)
      return

    const selectedCommand = commands[selectedIndex]
    if (!selectedCommand)
      return

    const nextCommands = [...commands]
    nextCommands.splice(selectedIndex, 1)
    nextCommands.splice(nextIndex, 0, selectedCommand)
    setCommands(nextCommands)
    setSelectedIndex(nextIndex)
    setError(undefined)
  }

  function saveTask(): void {
    const nameError = getTaskNameError(name)
    if (nameError) {
      setNameInput({
        kind: 'rename',
        value: name,
        cursor: name.length,
        error: nameError,
      })
      setConfirmExistingName(undefined)
      setMode('name')
      return
    }

    try {
      const normalizedCommands = validateTaskCommands(commands)
      finish({
        saved: true,
        name: normalizeTaskName(name),
        commands: normalizedCommands,
        previousName: sourceName,
      })
    }
    catch (err) {
      setError((err as Error).message)
    }
  }

  function confirmNameInput(): void {
    const nextName = normalizeTaskName(nameInput.value)
    const nameError = getTaskNameError(nextName)
    if (nameError) {
      setNameInput(current => ({ ...current, error: nameError }))
      return
    }

    const existingTask = existingTasks[nextName]
    const isSameTask = sourceName === nextName
    if (existingTask && !isSameTask) {
      if (allowLoadExisting) {
        if (confirmExistingName === nextName) {
          setSourceName(nextName)
          setName(nextName)
          setCommands(existingTask.commands)
          setSelectedIndex(0)
          setConfirmExistingName(undefined)
          setError(undefined)
          setMode('list')
          return
        }

        setConfirmExistingName(nextName)
        setNameInput(current => ({
          ...current,
          error: `Task "${nextName}" already exists. Press Enter to edit it, or Esc to cancel.`,
        }))
        return
      }

      setNameInput(current => ({ ...current, error: `Task already exists: ${nextName}` }))
      return
    }

    setName(nextName)
    setConfirmExistingName(undefined)
    setError(undefined)
    setMode('list')
  }

  function confirmCommandInput(): void {
    const commandError = getTaskCommandError(commandInput.value)
    if (commandError) {
      setCommandInput(current => ({ ...current, error: commandError }))
      return
    }

    const command = commandInput.value.trim()
    const nextCommands = [...commands]
    if (commandInput.index === null) {
      nextCommands.push(command)
      setSelectedIndex(nextCommands.length - 1)
    }
    else {
      nextCommands[commandInput.index] = command
      setSelectedIndex(commandInput.index)
    }

    setCommands(nextCommands)
    setError(undefined)
    setMode('list')
  }

  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      finish({ saved: false })
      return
    }

    if (mode === 'name') {
      if (key.return) {
        confirmNameInput()
        return
      }
      if (key.escape) {
        if (nameInput.kind === 'initial' && !name) {
          finish({ saved: false })
          return
        }

        setConfirmExistingName(undefined)
        setNameInput({
          kind: 'rename',
          value: name,
          cursor: name.length,
        })
        setMode('list')
        return
      }

      setConfirmExistingName(undefined)
      setNameInput(current => updateTextInput(current, input, key))
      return
    }

    if (mode === 'command') {
      if (key.return) {
        confirmCommandInput()
        return
      }
      if (key.escape) {
        setMode('list')
        setCommandInput({
          index: null,
          value: '',
          cursor: 0,
        })
        return
      }

      setCommandInput(current => updateTextInput(current, input, key))
      return
    }

    if (key.upArrow || input === 'k') {
      setSelectedIndex(index => Math.max(0, index - 1))
      return
    }
    if (key.downArrow || input === 'j') {
      if (commands.length === 0)
        return
      setSelectedIndex(index => Math.min(commands.length - 1, index + 1))
      return
    }
    if (key.escape || input === 'q') {
      finish({ saved: false })
      return
    }
    if (input === 'a') {
      startAddCommand()
      return
    }
    if (input === 'e') {
      startEditCommand()
      return
    }
    if (input === 'x' || key.delete) {
      deleteSelectedCommand()
      return
    }
    if (input === 'u') {
      moveSelectedCommand(-1)
      return
    }
    if (input === 'd') {
      moveSelectedCommand(1)
      return
    }
    if (input === 'r') {
      startRename()
      return
    }
    if (input === 's') {
      saveTask()
    }
  })

  return (
    <Box flexDirection="column" width={STUDIO_WIDTH}>
      <Header title={title} />
      <Text> </Text>
      <CommandPanel
        commandInput={commandInput}
        commands={commands}
        error={error}
        mode={mode}
        name={name}
        nameInput={nameInput}
        selectedIndex={selectedIndex}
      />
      <KeyBar mode={mode} />
    </Box>
  )
}
