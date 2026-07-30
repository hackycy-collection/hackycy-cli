import type { DirectoryEntry } from './api'
import { describe, expect, test } from 'bun:test'
import { clipboardOperation, createNavigationHistory, entryNameError, moveNavigation, operationActivities, parentDirectoryPath, pushNavigation, renameSelectionEnd, selectEntry, settleClipboard, visibleEntries } from './explorer-state'

const entries: DirectoryEntry[] = [
  { name: 'zeta.txt', path: 'zeta.txt', kind: 'file', isSymlink: false, previewKind: 'text', size: 8 },
  { name: 'Alpha.txt', path: 'Alpha.txt', kind: 'file', isSymlink: false, previewKind: 'text', size: 4 },
  { name: 'Alpha docs', path: 'Alpha docs', kind: 'directory', isSymlink: false, previewKind: 'none' },
]

describe('explorer state', () => {
  test('validates entry names and selects file names without their final extension', () => {
    expect(entryNameError('')).toBe('Name cannot be empty')
    expect(entryNameError('..')).toBe('Name cannot contain path separators or reserved segments')
    expect(entryNameError('nested/name')).toBe('Name cannot contain path separators or reserved segments')
    expect(entryNameError('report.txt')).toBeUndefined()

    expect(renameSelectionEnd({ kind: 'file', name: 'report.tar.gz' })).toBe('report.tar'.length)
    expect(renameSelectionEnd({ kind: 'file', name: '.env' })).toBe('.env'.length)
    expect(renameSelectionEnd({ kind: 'directory', name: 'reports.2026' })).toBe('reports.2026'.length)
    expect(parentDirectoryPath('docs/reports/file.txt')).toBe('docs/reports')
    expect(parentDirectoryPath('root-file.txt')).toBe('')
  })

  test('filters the current directory by name while preserving directory-first sorting', () => {
    expect(visibleEntries(entries, 'ALPHA', 'name', 'asc').map(entry => entry.path)).toEqual([
      'Alpha docs',
      'Alpha.txt',
    ])
  })

  test('supports plain, additive, and anchored range selection', () => {
    const order = ['Alpha docs', 'Alpha.txt', 'zeta.txt']
    const first = selectEntry(order, { paths: [] }, 'Alpha.txt', {})
    const additive = selectEntry(order, first, 'zeta.txt', { toggle: true })
    const range = selectEntry(order, additive, 'Alpha docs', { range: true })

    expect(first).toEqual({ paths: ['Alpha.txt'], anchorPath: 'Alpha.txt' })
    expect(additive).toEqual({ paths: ['Alpha.txt', 'zeta.txt'], anchorPath: 'zeta.txt' })
    expect(range).toEqual({ paths: ['Alpha docs', 'Alpha.txt', 'zeta.txt'], anchorPath: 'zeta.txt' })
  })

  test('creates paste operations and only removes successfully moved clipboard items', () => {
    const copy = { mode: 'copy' as const, paths: ['Alpha.txt'] }
    const move = { mode: 'move' as const, paths: ['Alpha.txt', 'zeta.txt'] }

    expect(clipboardOperation(copy, 'docs')).toEqual({ action: 'copy', paths: ['Alpha.txt'], destinationPath: 'docs' })
    expect(settleClipboard(copy, [{ status: 'ok', sourcePath: 'Alpha.txt' }])).toEqual(copy)
    expect(settleClipboard(move, [
      { status: 'ok', sourcePath: 'Alpha.txt' },
      { status: 'error', sourcePath: 'zeta.txt' },
    ])).toEqual({ mode: 'move', paths: ['zeta.txt'] })
  })

  test('turns batch operation results into per-item activity entries', () => {
    expect(operationActivities('activity', 'Copy items', {
      version: 1,
      action: 'copy',
      items: [
        { status: 'ok', sourcePath: 'Alpha.txt', destinationPath: 'docs/Alpha.txt' },
        { status: 'error', sourcePath: 'archive', error: { code: 'INVALID_OPERATION', message: 'A directory cannot be copied into itself' } },
      ],
    })).toEqual([
      { id: 'activity:0', label: 'Copy Alpha.txt', status: 'done', detail: 'Copied to /docs/Alpha.txt' },
      { id: 'activity:1', label: 'Copy archive', status: 'error', detail: 'A directory cannot be copied into itself' },
    ])

    expect(operationActivities('delete', 'Delete 1 item', {
      version: 1,
      action: 'delete',
      items: [{ status: 'ok', sourcePath: 'obsolete.txt' }],
    })).toEqual([
      { id: 'delete:0', label: 'Delete 1 item', status: 'done', detail: 'Deleted permanently' },
    ])
  })

  test('moves backward and forward while new navigation truncates forward history', () => {
    const root = createNavigationHistory('')
    const nested = pushNavigation(pushNavigation(root, 'docs'), 'docs/api')
    const back = moveNavigation(nested, -1)
    const replacedForward = pushNavigation(back, 'examples')

    expect(back).toEqual({ paths: ['', 'docs', 'docs/api'], cursor: 1 })
    expect(moveNavigation(back, 1).cursor).toBe(2)
    expect(replacedForward).toEqual({ paths: ['', 'docs', 'examples'], cursor: 2 })
  })
})
