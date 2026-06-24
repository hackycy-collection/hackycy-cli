import type { TaskGroupConfig } from '../../../config/types'
import { Box, Text, useApp, useInput } from 'ink'
import React, { useState } from 'react'

interface TaskPickerProps {
  title: string
  tasks: Record<string, TaskGroupConfig>
  onDone: (name: string | null) => void
}

const PICKER_WIDTH = 70
const LIST_WIDTH = 58

function clipText(value: string, maxWidth: number): string {
  if (value.length <= maxWidth)
    return value
  if (maxWidth <= 1)
    return value.slice(0, maxWidth)
  return `${value.slice(0, maxWidth - 1)}…`
}

function commandCountLabel(count: number): string {
  return `${count} command${count === 1 ? '' : 's'}`
}

function Header({ title }: { title: string }) {
  return (
    <Box width={PICKER_WIDTH} justifyContent="space-between">
      <Box>
        <Text bold color="cyan">HACKYCY CLI</Text>
        <Text color="gray"> / </Text>
        <Text bold>{title}</Text>
      </Box>
      <Text color="gray">Task Studio</Text>
    </Box>
  )
}

export function TaskPicker({ title, tasks, onDone }: TaskPickerProps) {
  const { exit } = useApp()
  const names = Object.keys(tasks).sort((a, b) => a.localeCompare(b))
  const [selectedIndex, setSelectedIndex] = useState(0)

  function finish(name: string | null): void {
    onDone(name)
    exit()
  }

  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      finish(null)
      return
    }
    if (key.escape || input === 'q') {
      finish(null)
      return
    }
    if (key.return) {
      finish(names[selectedIndex] ?? null)
      return
    }
    if (key.upArrow || input === 'k') {
      setSelectedIndex(index => Math.max(0, index - 1))
      return
    }
    if (key.downArrow || input === 'j') {
      setSelectedIndex(index => Math.min(names.length - 1, index + 1))
    }
  })

  return (
    <Box flexDirection="column" width={PICKER_WIDTH}>
      <Header title={title} />
      <Text> </Text>
      <Box borderStyle="round" borderColor="cyan" flexDirection="column" paddingX={1} width={PICKER_WIDTH}>
        <Text bold>Select Task</Text>
        <Text> </Text>
        {names.map((name, index) => {
          const selected = index === selectedIndex
          const count = tasks[name]!.commands.length
          return (
            <Box key={name} justifyContent="space-between" width={LIST_WIDTH}>
              <Box>
                <Text color={selected ? 'cyan' : 'gray'}>{selected ? '▸ ' : '  '}</Text>
                <Text inverse={selected}>{clipText(name, 34)}</Text>
              </Box>
              <Text color="gray">{commandCountLabel(count)}</Text>
            </Box>
          )
        })}
      </Box>
      <Box marginTop={1}>
        <Text bold color="cyan">↑/↓</Text>
        <Text color="gray"> select   </Text>
        <Text bold color="cyan">Enter</Text>
        <Text color="gray"> open   </Text>
        <Text bold color="cyan">q</Text>
        <Text color="gray"> cancel</Text>
      </Box>
    </Box>
  )
}
