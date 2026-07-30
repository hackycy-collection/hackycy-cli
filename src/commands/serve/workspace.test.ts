import type { ServeWorkspace } from './types'
import { mkdir, mkdtemp, rm, symlink, writeFile } from 'node:fs/promises'
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
})
