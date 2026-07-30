import type { ServeWorkspace } from './types'
import { lstat, mkdir, mkdtemp, readlink, rm, symlink, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, test } from 'bun:test'
import { createServeWorkspace } from './workspace'

const temporaryDirectories: string[] = []

afterEach(async () => {
  await Promise.all(temporaryDirectories.splice(0).map(directory => rm(directory, { recursive: true, force: true })))
})

async function createFixture(): Promise<{ root: string, workspace: ServeWorkspace }> {
  const fixture = await mkdtemp(path.join(tmpdir(), 'ycy-serve-workspace-'))
  temporaryDirectories.push(fixture)
  const root = path.join(fixture, 'shared files')
  await mkdir(path.join(root, 'docs'), { recursive: true })
  await Promise.all([
    writeFile(path.join(root, 'zeta.txt'), 'zeta'),
    writeFile(path.join(root, 'Alpha.txt'), 'alpha'),
  ])
  return { root, workspace: await createServeWorkspace(root) }
}

describe('ServeWorkspace', () => {
  test('lists one directory with stable browser paths and directories first', async () => {
    const { workspace } = await createFixture()

    const listing = await workspace.listDirectory('')

    expect(listing).toEqual({
      rootName: 'shared files',
      path: '',
      parentPath: undefined,
      entries: [
        expect.objectContaining({
          name: 'docs',
          path: 'docs',
          kind: 'directory',
          browseUrl: '/browse/docs',
        }),
        expect.objectContaining({
          name: 'Alpha.txt',
          path: 'Alpha.txt',
          kind: 'file',
          fileUrl: '/files/Alpha.txt',
          downloadUrl: '/files/Alpha.txt?download=1',
        }),
        expect.objectContaining({
          name: 'zeta.txt',
          path: 'zeta.txt',
          kind: 'file',
          fileUrl: '/files/zeta.txt',
          downloadUrl: '/files/zeta.txt?download=1',
        }),
      ],
    })
  })

  test('accepts only root-confined POSIX paths and marks escaping symlinks unavailable', async () => {
    const { root, workspace } = await createFixture()
    const outside = path.join(path.dirname(root), 'outside')
    await mkdir(outside)
    await writeFile(path.join(outside, 'secret.txt'), 'secret')
    await Promise.all([
      symlink(path.join(root, 'docs'), path.join(root, 'internal-docs')),
      symlink(outside, path.join(root, 'external-docs')),
    ])

    const listing = await workspace.listDirectory('')

    expect(listing.entries.find(entry => entry.name === 'internal-docs')).toEqual(expect.objectContaining({
      kind: 'directory',
      isSymlink: true,
      browseUrl: '/browse/internal-docs',
    }))
    expect(listing.entries.find(entry => entry.name === 'external-docs')).toEqual(expect.objectContaining({
      kind: 'unavailable',
      isSymlink: true,
    }))
    await expect(workspace.listDirectory('../outside')).rejects.toMatchObject({ code: 'PATH_FORBIDDEN' })
    await expect(workspace.listDirectory('docs\\..\\outside')).rejects.toMatchObject({ code: 'INVALID_PATH' })
    await expect(workspace.listDirectory('external-docs')).rejects.toMatchObject({ code: 'PATH_FORBIDDEN' })
  })

  test('rejects NUL bytes before passing paths or filenames to the filesystem', async () => {
    const { workspace } = await createFixture()

    await expect(workspace.listDirectory('\0')).rejects.toMatchObject({ code: 'INVALID_PATH' })
    await expect(workspace.uploadFile('docs', new File(['bad'], 'bad\0name.txt'))).rejects.toMatchObject({ code: 'INVALID_UPLOAD' })
  })

  test('opens a root-confined file without exposing its absolute path', async () => {
    const { workspace } = await createFixture()

    const file = await workspace.openFile('Alpha.txt')

    expect(file).toEqual(expect.objectContaining({
      name: 'Alpha.txt',
      size: 5,
      mimeType: 'text/plain;charset=utf-8',
    }))
    expect(await file.body.text()).toBe('alpha')
    await expect(workspace.openFile('docs')).rejects.toMatchObject({ code: 'NOT_FILE' })
  })

  test('classifies structured text MIME types with parameters as text previews', async () => {
    const { root, workspace } = await createFixture()
    await writeFile(path.join(root, 'data.json'), '{"ready":true}')

    const listing = await workspace.listDirectory('')

    expect(listing.entries.find(entry => entry.name === 'data.json')).toEqual(expect.objectContaining({
      mimeType: 'application/json;charset=utf-8',
      previewKind: 'text',
    }))
  })

  test('returns bounded decoded text without loading oversized or binary files as text', async () => {
    const { root, workspace } = await createFixture()
    await Promise.all([
      writeFile(path.join(root, 'utf16.txt'), Uint8Array.from([0xFF, 0xFE, 0x68, 0x00, 0x69, 0x00])),
      writeFile(path.join(root, 'binary.txt'), Uint8Array.from([0xC3, 0x28])),
      writeFile(path.join(root, 'nul.txt'), Uint8Array.from([0x00, 0x01, 0x02])),
      writeFile(path.join(root, 'large.txt'), new Uint8Array(2 * 1024 * 1024 + 1)),
    ])

    expect(await workspace.readTextPreview('Alpha.txt')).toEqual({
      status: 'ready',
      text: 'alpha',
      encoding: 'utf-8',
      size: 5,
    })
    expect(await workspace.readTextPreview('utf16.txt')).toEqual({
      status: 'ready',
      text: 'hi',
      encoding: 'utf-16le',
      size: 6,
    })
    expect(await workspace.readTextPreview('binary.txt')).toEqual({ status: 'binary', size: 2 })
    expect(await workspace.readTextPreview('nul.txt')).toEqual({ status: 'binary', size: 3 })
    expect(await workspace.readTextPreview('large.txt')).toEqual({
      status: 'too_large',
      size: 2 * 1024 * 1024 + 1,
      maxBytes: 2 * 1024 * 1024,
    })
  })

  test('uploads atomically and chooses a new name instead of overwriting', async () => {
    const { workspace } = await createFixture()

    const first = await workspace.uploadFile('docs', new File(['first'], 'report.txt'))
    const second = await workspace.uploadFile('docs', new File(['second'], 'report.txt'))

    expect(first).toEqual({ filename: 'report.txt', path: 'docs/report.txt', size: 5 })
    expect(second).toEqual({ filename: 'report (1).txt', path: 'docs/report (1).txt', size: 6 })
    expect(await (await workspace.openFile(first.path)).body.text()).toBe('first')
    expect(await (await workspace.openFile(second.path)).body.text()).toBe('second')
    expect((await workspace.listDirectory('docs')).entries.map(entry => entry.name)).toEqual([
      'report (1).txt',
      'report.txt',
    ])
    await expect(workspace.uploadFile('docs', new File(['bad'], '../secret.txt'))).rejects.toMatchObject({ code: 'INVALID_UPLOAD' })
  })

  test('creates a directory without overwriting an existing entry', async () => {
    const { workspace } = await createFixture()

    expect(await workspace.applyOperation({ action: 'create-directory', parentPath: '', name: 'projects' })).toEqual({
      action: 'create-directory',
      items: [{ status: 'ok', destinationPath: 'projects' }],
    })
    expect((await workspace.listDirectory('')).entries.map(entry => entry.name)).toContain('projects')

    expect(await workspace.applyOperation({ action: 'create-directory', parentPath: '', name: 'projects' })).toEqual({
      action: 'create-directory',
      items: [{
        status: 'error',
        destinationPath: 'projects',
        error: { code: 'ALREADY_EXISTS', message: 'An entry with that name already exists' },
      }],
    })
  })

  test('rejects invalid names for create and rename operations', async () => {
    const { workspace } = await createFixture()

    for (const name of ['', '   ', '.', '..', 'nested/name', 'nested\\name', 'nul\0name']) {
      expect(await workspace.applyOperation({ action: 'create-directory', parentPath: '', name })).toEqual({
        action: 'create-directory',
        items: [{
          status: 'error',
          error: { code: 'INVALID_NAME', message: 'Entry name is invalid' },
        }],
      })
    }

    expect(await workspace.applyOperation({ action: 'rename', path: 'Alpha.txt', newName: '../renamed.txt' })).toEqual({
      action: 'rename',
      items: [{
        status: 'error',
        sourcePath: 'Alpha.txt',
        error: { code: 'INVALID_NAME', message: 'Entry name is invalid' },
      }],
    })
  })

  test('renames an entry without overwriting another entry', async () => {
    const { workspace } = await createFixture()

    expect(await workspace.applyOperation({ action: 'rename', path: 'Alpha.txt', newName: 'beta.txt' })).toEqual({
      action: 'rename',
      items: [{ status: 'ok', sourcePath: 'Alpha.txt', destinationPath: 'beta.txt' }],
    })
    expect(await (await workspace.openFile('beta.txt')).body.text()).toBe('alpha')

    expect(await workspace.applyOperation({ action: 'rename', path: 'zeta.txt', newName: 'beta.txt' })).toEqual({
      action: 'rename',
      items: [{
        status: 'error',
        sourcePath: 'zeta.txt',
        destinationPath: 'beta.txt',
        error: { code: 'ALREADY_EXISTS', message: 'An entry with that name already exists' },
      }],
    })
  })

  test('copies files and directories recursively with collision-safe names', async () => {
    const { root, workspace } = await createFixture()
    await writeFile(path.join(root, 'docs', 'guide.txt'), 'guide')

    expect(await workspace.applyOperation({ action: 'copy', paths: ['Alpha.txt', 'docs'], destinationPath: 'docs' })).toEqual({
      action: 'copy',
      items: [
        { status: 'ok', sourcePath: 'Alpha.txt', destinationPath: 'docs/Alpha.txt' },
        { status: 'error', sourcePath: 'docs', error: { code: 'INVALID_OPERATION', message: 'A directory cannot be copied into itself' } },
      ],
    })
    expect(await (await workspace.openFile('docs/Alpha.txt')).body.text()).toBe('alpha')

    expect(await workspace.applyOperation({ action: 'copy', paths: ['docs'], destinationPath: '' })).toEqual({
      action: 'copy',
      items: [{ status: 'ok', sourcePath: 'docs', destinationPath: 'docs (1)' }],
    })
    expect(await (await workspace.openFile('docs (1)/guide.txt')).body.text()).toBe('guide')
  })

  test('copies symlinks without dereferencing them', async () => {
    const { root, workspace } = await createFixture()
    const outside = path.join(path.dirname(root), 'outside-copy-target')
    await mkdir(outside)
    await writeFile(path.join(outside, 'keep.txt'), 'keep')
    await Promise.all([
      symlink('../Alpha.txt', path.join(root, 'docs', 'alpha-link')),
      symlink(outside, path.join(root, 'docs', 'outside-link')),
    ])

    expect(await workspace.applyOperation({ action: 'copy', paths: ['docs'], destinationPath: '' })).toEqual({
      action: 'copy',
      items: [{ status: 'ok', sourcePath: 'docs', destinationPath: 'docs (1)' }],
    })
    expect((await lstat(path.join(root, 'docs (1)', 'alpha-link'))).isSymbolicLink()).toBe(true)
    expect((await lstat(path.join(root, 'docs (1)', 'outside-link'))).isSymbolicLink()).toBe(true)
    expect(await readlink(path.join(root, 'docs (1)', 'alpha-link'))).toBe('../Alpha.txt')
    expect(await readlink(path.join(root, 'docs (1)', 'outside-link'))).toBe(outside)
    expect(await Bun.file(path.join(outside, 'keep.txt')).text()).toBe('keep')
  })

  test('rejects mutations through a symlink ancestor that escapes the root', async () => {
    const { root, workspace } = await createFixture()
    const outside = path.join(path.dirname(root), 'outside-operation-target')
    await mkdir(outside)
    await writeFile(path.join(outside, 'secret.txt'), 'secret')
    await symlink(outside, path.join(root, 'escape'))

    expect(await workspace.applyOperation({ action: 'create-directory', parentPath: 'escape', name: 'created' })).toEqual({
      action: 'create-directory',
      items: [{ status: 'error', error: { code: 'PATH_FORBIDDEN', message: 'Path escapes the served directory' } }],
    })
    expect(await workspace.applyOperation({ action: 'rename', path: 'escape/secret.txt', newName: 'renamed.txt' })).toEqual({
      action: 'rename',
      items: [{ status: 'error', sourcePath: 'escape/secret.txt', error: { code: 'PATH_FORBIDDEN', message: 'Path escapes the served directory' } }],
    })
    expect(await workspace.applyOperation({ action: 'copy', paths: ['Alpha.txt'], destinationPath: 'escape' })).toEqual({
      action: 'copy',
      items: [{ status: 'error', sourcePath: 'Alpha.txt', error: { code: 'PATH_FORBIDDEN', message: 'Path escapes the served directory' } }],
    })
    expect(await workspace.applyOperation({ action: 'move', paths: ['Alpha.txt'], destinationPath: 'escape' })).toEqual({
      action: 'move',
      items: [{ status: 'error', sourcePath: 'Alpha.txt', error: { code: 'PATH_FORBIDDEN', message: 'Path escapes the served directory' } }],
    })
    expect(await workspace.applyOperation({ action: 'delete', paths: ['escape/secret.txt'] })).toEqual({
      action: 'delete',
      items: [{ status: 'error', sourcePath: 'escape/secret.txt', error: { code: 'PATH_FORBIDDEN', message: 'Path escapes the served directory' } }],
    })
    expect(await Bun.file(path.join(outside, 'secret.txt')).text()).toBe('secret')
  })

  test('moves entries without overwriting or moving a directory into itself', async () => {
    const { root, workspace } = await createFixture()
    await writeFile(path.join(root, 'docs', 'zeta.txt'), 'occupied')

    expect(await workspace.applyOperation({ action: 'move', paths: ['Alpha.txt', 'zeta.txt', 'docs'], destinationPath: 'docs' })).toEqual({
      action: 'move',
      items: [
        { status: 'ok', sourcePath: 'Alpha.txt', destinationPath: 'docs/Alpha.txt' },
        {
          status: 'error',
          sourcePath: 'zeta.txt',
          destinationPath: 'docs/zeta.txt',
          error: { code: 'ALREADY_EXISTS', message: 'An entry with that name already exists' },
        },
        {
          status: 'error',
          sourcePath: 'docs',
          error: { code: 'INVALID_OPERATION', message: 'A directory cannot be moved into itself' },
        },
      ],
    })
    expect(await (await workspace.openFile('docs/Alpha.txt')).body.text()).toBe('alpha')
    await expect(workspace.openFile('Alpha.txt')).rejects.toMatchObject({ code: 'NOT_FOUND' })
  })

  test('permanently deletes entries without following final symlinks and reports partial failures', async () => {
    const { root, workspace } = await createFixture()
    const outside = path.join(path.dirname(root), 'outside-delete-target')
    await mkdir(outside)
    await writeFile(path.join(outside, 'keep.txt'), 'keep')
    await symlink(outside, path.join(root, 'outside-link'))

    expect(await workspace.applyOperation({ action: 'delete', paths: ['Alpha.txt', 'docs', 'outside-link', 'missing.txt', ''] })).toEqual({
      action: 'delete',
      items: [
        { status: 'ok', sourcePath: 'Alpha.txt' },
        { status: 'ok', sourcePath: 'docs' },
        { status: 'ok', sourcePath: 'outside-link' },
        { status: 'error', sourcePath: 'missing.txt', error: { code: 'NOT_FOUND', message: 'Path does not exist' } },
        { status: 'error', sourcePath: '', error: { code: 'ROOT_IMMUTABLE', message: 'The served root cannot be changed' } },
      ],
    })
    expect((await workspace.listDirectory('')).entries.map(entry => entry.name)).toEqual(['zeta.txt'])
    expect(await Bun.file(path.join(outside, 'keep.txt')).text()).toBe('keep')
  })
})
