import { Box, Text, useApp, useInput } from 'ink'
import React, { useState } from 'react'

interface TaskConfirmProps {
  title: string
  message: string
  detail?: string
  confirmLabel: string
  onDone: (confirmed: boolean) => void
}

const CONFIRM_WIDTH = 70

function Header({ title }: { title: string }) {
  return (
    <Box width={CONFIRM_WIDTH} justifyContent="space-between">
      <Box>
        <Text bold color="cyan">HACKYCY CLI</Text>
        <Text color="gray"> / </Text>
        <Text bold>{title}</Text>
      </Box>
      <Text color="gray">Task Studio</Text>
    </Box>
  )
}

export function TaskConfirm({ title, message, detail, confirmLabel, onDone }: TaskConfirmProps) {
  const { exit } = useApp()
  const [confirmed, setConfirmed] = useState(false)

  function finish(value: boolean): void {
    onDone(value)
    exit()
  }

  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      finish(false)
      return
    }
    if (key.escape || input === 'q' || input === 'n') {
      finish(false)
      return
    }
    if (input === 'y') {
      finish(true)
      return
    }
    if (key.leftArrow || key.rightArrow || key.tab) {
      setConfirmed(value => !value)
      return
    }
    if (key.return) {
      finish(confirmed)
    }
  })

  return (
    <Box flexDirection="column" width={CONFIRM_WIDTH}>
      <Header title={title} />
      <Text> </Text>
      <Box borderStyle="round" borderColor="red" flexDirection="column" paddingX={1} width={CONFIRM_WIDTH}>
        <Text bold color="red">{message}</Text>
        {detail ? <Text color="gray">{detail}</Text> : null}
        <Text> </Text>
        <Box>
          <Text inverse={confirmed} color={confirmed ? undefined : 'red'}>
            {' '}
            {confirmLabel}
            {' '}
          </Text>
          <Text>  </Text>
          <Text inverse={!confirmed} color={!confirmed ? undefined : 'gray'}> Cancel </Text>
        </Box>
      </Box>
      <Box marginTop={1}>
        <Text bold color="cyan">←/→</Text>
        <Text color="gray"> choose   </Text>
        <Text bold color="cyan">Enter</Text>
        <Text color="gray"> confirm   </Text>
        <Text bold color="cyan">q</Text>
        <Text color="gray"> cancel</Text>
      </Box>
    </Box>
  )
}
