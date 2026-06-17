import type { HeatReport, PathHeat } from '../types'
import path from 'node:path'
import process from 'node:process'
import { Box, Text } from 'ink'
import React, { useEffect } from 'react'

interface HeatReportViewProps {
  report: HeatReport
  onDone: () => void
}

const COL = { rank: 4, total: 7, stat: 4 }
const FIXED_WIDTH = COL.rank + COL.total + COL.stat * 5
const MAX_ROWS = 12

export function HeatReportView({ report, onDone }: HeatReportViewProps) {
  useEffect(() => {
    const timer = setTimeout(onDone, 100)
    return () => clearTimeout(timer)
  }, [onDone])

  return (
    <Box flexDirection="column">
      <Summary report={report} />
      <Text> </Text>
      <HeatTable
        title={report.target === 'files' ? 'Top changed files' : 'Top changed directories'}
        rows={(report.target === 'files' ? report.files : report.directories).slice(0, MAX_ROWS)}
        pathLabel={report.target === 'files' ? 'File' : 'Directory'}
      />
      <Box marginTop={1}>
        <Text color="gray">
          Legend
          {'  '}
          <Text color="yellow">M</Text>
          {' modified  '}
          <Text color="green">A</Text>
          {' added  '}
          <Text color="red">D</Text>
          {' deleted  '}
          <Text color="magenta">R</Text>
          {' renamed  '}
          <Text color="blue">C</Text>
          {' copied'}
        </Text>
      </Box>
    </Box>
  )
}

function Summary({ report }: { report: HeatReport }) {
  const count = report.target === 'files' ? report.files.length : report.directories.length
  const label = report.target === 'files'
    ? `file${count === 1 ? '' : 's'}`
    : `director${count === 1 ? 'y' : 'ies'}`
  const shouldShowCommitCount = !report.rangeLabel.includes('commit')

  return (
    <Text color="gray" dimColor>
      {report.repoName}
      {' · '}
      {report.rangeLabel}
      {shouldShowCommitCount
        ? (
            <>
              {' · '}
              {report.commitCount}
              {' '}
              commit
              {report.commitCount === 1 ? '' : 's'}
            </>
          )
        : null}
      {' · '}
      {count}
      {' '}
      {label}
    </Text>
  )
}

function HeatTable({ title, rows, pathLabel }: { title: string, rows: PathHeat[], pathLabel: string }) {
  const pathWidth = getPathWidth()

  return (
    <Box flexDirection="column">
      <Text bold color="magenta">{title}</Text>
      <Box flexDirection="row">
        <HeaderCell width={COL.rank}>#</HeaderCell>
        <HeaderCell width={COL.total}>Total</HeaderCell>
        <HeaderCell width={COL.stat}>M</HeaderCell>
        <HeaderCell width={COL.stat}>A</HeaderCell>
        <HeaderCell width={COL.stat}>D</HeaderCell>
        <HeaderCell width={COL.stat}>R</HeaderCell>
        <HeaderCell width={COL.stat}>C</HeaderCell>
        <Text bold color="cyan">{pathLabel}</Text>
      </Box>
      <Text color="gray">{'─'.repeat(FIXED_WIDTH + pathWidth)}</Text>
      {rows.length === 0
        ? <Text color="gray">No changes found.</Text>
        : rows.map((row, index) => (
            <HeatRow key={row.path} index={index} row={row} pathWidth={pathWidth} />
          ))}
    </Box>
  )
}

function HeatRow({ index, row, pathWidth }: { index: number, row: PathHeat, pathWidth: number }) {
  const parsed = splitPath(row.path)

  return (
    <Box flexDirection="row">
      <Cell width={COL.rank}><Text color="gray">{String(index + 1)}</Text></Cell>
      <Cell width={COL.total}><Text color="cyan">{String(row.total)}</Text></Cell>
      <CountCell color="yellow" width={COL.stat} value={row.modified} />
      <CountCell color="green" width={COL.stat} value={row.added} />
      <CountCell color="red" width={COL.stat} value={row.deleted} />
      <CountCell color="magenta" width={COL.stat} value={row.renamed} />
      <CountCell color="blue" width={COL.stat} value={row.copied} />
      {parsed.dir
        ? (
            <Text>
              <Text color="gray" dimColor>{truncatePath(parsed.dir, pathWidth - parsed.base.length - 1)}</Text>
              <Text color="gray" dimColor>/</Text>
              <Text>{truncateEnd(parsed.base, pathWidth)}</Text>
            </Text>
          )
        : <Text>{truncateEnd(row.path, pathWidth)}</Text>}
    </Box>
  )
}

function HeaderCell({ width, children }: { width: number, children: React.ReactNode }) {
  return (
    <Box minWidth={width}>
      <Text bold color="cyan">{children}</Text>
    </Box>
  )
}

function Cell({ width, children }: { width: number, children: React.ReactNode }) {
  return <Box minWidth={width}>{children}</Box>
}

function CountCell({ color, width, value }: { color: string, width: number, value: number }) {
  return (
    <Cell width={width}>
      <Text color={value > 0 ? color : 'gray'}>{String(value)}</Text>
    </Cell>
  )
}

function getPathWidth(): number {
  const columns = Math.min(process.stdout.columns || 80, 80)
  return Math.max(28, columns - FIXED_WIDTH - 2)
}

function splitPath(filePath: string): { dir: string, base: string } {
  if (filePath === '.')
    return { dir: '', base: '.' }

  const dir = path.dirname(filePath)
  return {
    dir: dir === '.' ? '' : dir,
    base: path.basename(filePath),
  }
}

function truncatePath(value: string, width: number): string {
  if (width <= 0)
    return ''
  if (value.length <= width)
    return value
  if (width <= 3)
    return '.'.repeat(width)
  return `...${value.slice(value.length - width + 3)}`
}

function truncateEnd(value: string, width: number): string {
  if (width <= 0)
    return ''
  if (value.length <= width)
    return value
  if (width <= 3)
    return '.'.repeat(width)
  return `${value.slice(0, width - 3)}...`
}
