import type { StatsFs } from 'node:fs'
import { Buffer } from 'node:buffer'
import { chmod, mkdir, mkdtemp, readdir, readFile, rm, writeFile } from 'node:fs/promises'
import os from 'node:os'
import path from 'node:path'
import process from 'node:process'
import { afterAll, beforeAll, describe, expect, test } from 'bun:test'
import { strToU8, zipSync } from 'fflate'
import { prepareSevenZipRuntime } from '../../../scripts/prepare-seven-zip'
import { createSevenZipArchiveExtractor } from './archive-extractor'
import { createServeWorkspace } from './workspace'

const RAR_FIXTURE = 'UmFyIRoHAM+QcwAADQAAAAAAAACEUnQgkDIAFAAAABQAAAADQqLIvrd22j4UMAgApIEAAHRlc3QudHh0gAi3dto+t3baPnRlc3QgdGV4dCBkb2N1bWVudA0KnS90IJAyAAgAAAAIAAAAA3tEybbRTNg+FDAIAP+hAAB0ZXN0bGlua8AI0UzYPlBf2j50ZXN0LnR4dM3gdCCQOgAUAAAAFAAAAANCosi+Y3faPhQwEACkgQAAdGVzdGRpclx0ZXN0LnR4dMDMY3faPmN32j50ZXN0IHRleHQgZG9jdW1lbnQNCqHIdOCQMQAAAAAAAAAAAAMAAAAAY3faPhQwBwDtQQAAdGVzdGRpcsDMY3faPmR32j7m53TgkDYAAAAAAAAAAAADAAAAAJ2r1T4UMAwA7UEAAHRlc3RlbXB0eWRpcoDMnavVPsVd2j7EPXsAQAcA'
const temporaryDirectories: string[] = []
let sevenZip = ''

beforeAll(async () => {
  const files = await prepareSevenZipRuntime()
  sevenZip = files.find(file => /ycy-7zz\.bin$|ycy-7z\.exe$/.test(file))!
})

afterAll(async () => {
  await Promise.all(temporaryDirectories.splice(0).map(directory => rm(directory, { recursive: true, force: true })))
})

async function temporaryRoot(): Promise<string> {
  const root = await mkdtemp(path.join(os.tmpdir(), 'ycy-real-archive-'))
  temporaryDirectories.push(root)
  return root
}

async function runSevenZip(arguments_: string[], cwd: string): Promise<void> {
  const child = Bun.spawn([sevenZip, ...arguments_], { cwd, stdin: 'ignore', stdout: 'ignore', stderr: 'pipe' })
  const errorOutput = new Response(child.stderr).text()
  const [exitCode, error] = await Promise.all([child.exited, errorOutput])
  if (exitCode !== 0)
    throw new Error(error || `7-Zip exited with ${exitCode}`)
}

async function createPayload(root: string): Promise<void> {
  await mkdir(path.join(root, 'payload', 'nested'), { recursive: true })
  await mkdir(path.join(root, 'payload', 'empty'))
  await Promise.all([
    writeFile(path.join(root, 'payload', 'Unicode 文件.txt'), 'unicode'),
    writeFile(path.join(root, 'payload', 'nested', 'space name.txt'), 'nested'),
    writeFile(path.join(root, 'payload', '-leading.txt'), 'leading'),
  ])
}

function filesystem(update: Partial<StatsFs> = {}): StatsFs {
  return {
    type: 0,
    bsize: 4096,
    blocks: 1_000_000,
    bfree: 1_000_000,
    bavail: 1_000_000,
    files: 1_000_000,
    ffree: 1_000_000,
    ...update,
  }
}

describe('SevenZipArchiveExtractor', () => {
  test('extracts 7z, ZIP, RAR, TAR, and every supported compressed TAR layer', async () => {
    const root = await temporaryRoot()
    await createPayload(root)
    await Promise.all([
      runSevenZip(['a', '-t7z', 'plain.7z', '--', 'payload'], root),
      runSevenZip(['a', '-ttar', 'payload.tar', '--', 'payload'], root),
      writeFile(path.join(root, '-leading.zip'), zipSync({
        'Unicode 文件.txt': strToU8('unicode'),
        'nested/space name.txt': strToU8('nested'),
        '-leading.txt': strToU8('leading'),
        'empty/': new Uint8Array(),
      })),
      writeFile(path.join(root, 'empty.zip'), zipSync({})),
      writeFile(path.join(root, 'fixture.rar'), Buffer.from(RAR_FIXTURE, 'base64')),
    ])
    const tar = new Uint8Array(await readFile(path.join(root, 'payload.tar')))
    await Promise.all([
      writeFile(path.join(root, 'gzip.tar.gz'), Bun.gzipSync(tar)),
      writeFile(path.join(root, 'zstd.tar.zst'), Bun.zstdCompressSync(tar)),
      runSevenZip(['a', '-tbzip2', 'bzip.tar.bz2', '--', 'payload.tar'], root),
      runSevenZip(['a', '-txz', 'xz.tar.xz', '--', 'payload.tar'], root),
    ])

    const workspace = await createServeWorkspace(root, {
      archiveExtractor: createSevenZipArchiveExtractor({ executable: async () => sevenZip }),
    })
    const archivePaths = ['plain.7z', 'payload.tar', 'gzip.tar.gz', 'bzip.tar.bz2', 'xz.tar.xz', 'zstd.tar.zst']
    for (const archivePath of archivePaths) {
      const result = await workspace.extractArchive(archivePath)
      expect(await readFile(path.join(root, result.destinationPath, 'payload', 'Unicode 文件.txt'), 'utf8')).toBe('unicode')
      expect(await readFile(path.join(root, result.destinationPath, 'payload', 'nested', 'space name.txt'), 'utf8')).toBe('nested')
      expect(await readFile(path.join(root, result.destinationPath, 'payload', '-leading.txt'), 'utf8')).toBe('leading')
    }

    const leading = await workspace.extractArchive('-leading.zip')
    expect(leading.destinationPath).toBe('-leading')
    expect(await readFile(path.join(root, '-leading', 'Unicode 文件.txt'), 'utf8')).toBe('unicode')
    expect((await workspace.extractArchive('empty.zip')).entryCount).toBe(0)
    expect(await readdirNames(path.join(root, 'empty'))).toEqual([])

    const rar = await workspace.extractArchive('fixture.rar')
    expect(await readFile(path.join(root, rar.destinationPath, 'test.txt'), 'utf8')).toBe('test text document\r\n')
    expect(await readFile(path.join(root, rar.destinationPath, 'testdir', 'test.txt'), 'utf8')).toBe('test text document\r\n')
  }, 30_000)

  test('rejects damaged, encrypted, traversal, absolute, and drive-letter entries before publication', async () => {
    const root = await temporaryRoot()
    await writeFile(path.join(root, 'secret.txt'), 'secret')
    await runSevenZip(['a', '-tzip', '-psecret', '-mem=AES256', 'encrypted.zip', '--', 'secret.txt'], root)
    await Promise.all([
      writeFile(path.join(root, 'damaged.zip'), 'not an archive'),
      writeFile(path.join(root, 'traversal.zip'), zipSync({ '../escape.txt': strToU8('escape') })),
      writeFile(path.join(root, 'absolute.zip'), zipSync({ '/absolute.txt': strToU8('absolute') })),
      writeFile(path.join(root, 'drive.zip'), zipSync({ 'C:/escape.txt': strToU8('drive') })),
    ])
    const workspace = await createServeWorkspace(root, {
      archiveExtractor: createSevenZipArchiveExtractor({ executable: async () => sevenZip }),
    })

    await expect(workspace.extractArchive('damaged.zip')).rejects.toMatchObject({ code: 'INVALID_ARCHIVE' })
    await expect(workspace.extractArchive('encrypted.zip')).rejects.toMatchObject({ code: 'ENCRYPTED_ARCHIVE' })
    for (const archive of ['traversal.zip', 'absolute.zip', 'drive.zip'])
      await expect(workspace.extractArchive(archive)).rejects.toMatchObject({ code: 'INVALID_ARCHIVE' })
    expect((await readdirNames(root)).some(name => name.startsWith('.extract-'))).toBe(false)
  }, 20_000)

  test('rejects multi-volume metadata before invoking extraction', async () => {
    if (process.platform === 'win32')
      return
    const root = await temporaryRoot()
    const executable = path.join(root, 'fake-7z')
    await writeFile(executable, [
      '#!/bin/sh',
      'printf "Path = archive.rar\\nType = Rar\\nVolumes = 2\\n\\n----------\\nPath = safe.txt\\nSize = 1\\nEncrypted = -\\n\\n"',
    ].join('\n'))
    await chmod(executable, 0o755)
    const archive = path.join(root, 'archive.rar')
    const destination = path.join(root, 'destination')
    await writeFile(archive, 'fixture')
    await mkdir(destination)
    const extractor = createSevenZipArchiveExtractor({ executable: async () => executable })

    await expect(extractor.extract(archive, destination)).rejects.toMatchObject({
      code: 'INVALID_ARCHIVE',
      message: 'Multi-volume archives are not supported',
    })
    expect(await readdirNames(destination)).toEqual([])
  })

  test('checks available bytes and inodes using the reported archive totals', async () => {
    const root = await temporaryRoot()
    const archive = path.join(root, 'capacity.zip')
    await writeFile(archive, zipSync({ 'one.txt': strToU8('1234567890'), 'two.txt': strToU8('1234567890') }))

    const byteDestination = path.join(root, 'byte-destination')
    await mkdir(byteDestination)
    const byteExtractor = createSevenZipArchiveExtractor({
      executable: async () => sevenZip,
      statfs: async () => filesystem({ bsize: 1, bavail: 20, ffree: 100 }),
    })
    await expect(byteExtractor.extract(archive, byteDestination)).rejects.toMatchObject({ code: 'INSUFFICIENT_SPACE' })

    const inodeDestination = path.join(root, 'inode-destination')
    await mkdir(inodeDestination)
    const inodeExtractor = createSevenZipArchiveExtractor({
      executable: async () => sevenZip,
      statfs: async () => filesystem({ bavail: 1_000_000, ffree: 1 }),
    })
    await expect(inodeExtractor.extract(archive, inodeDestination)).rejects.toMatchObject({ code: 'INSUFFICIENT_SPACE' })
  })

  test('does not start a child process for an already cancelled extraction', async () => {
    const controller = new AbortController()
    controller.abort(new Error('already cancelled'))
    const extractor = createSevenZipArchiveExtractor({ executable: async () => '/missing/seven-zip' })

    await expect(extractor.extract('/missing/archive.zip', '/missing/destination', { signal: controller.signal })).rejects.toThrow('already cancelled')
  })
})

async function readdirNames(directory: string): Promise<string[]> {
  return readdir(directory)
}
