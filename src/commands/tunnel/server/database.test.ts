import { mkdtemp, rm } from 'node:fs/promises'
import os from 'node:os'
import path from 'node:path'
import { Database } from 'bun:sqlite'
import { afterEach, describe, expect, test } from 'bun:test'
import { TunnelDatabase } from './database'

const temporaryDirectories: string[] = []

afterEach(async () => {
  await Promise.all(temporaryDirectories.splice(0).map(directory => rm(directory, { force: true, recursive: true })))
})

describe('TunnelDatabase migrations', () => {
  test('adds an empty Client Remark to schema version 1 records', async () => {
    const directory = await mkdtemp(path.join(os.tmpdir(), 'ycy-tunnel-database-'))
    temporaryDirectories.push(directory)
    const filePath = path.join(directory, 'tunnel.sqlite')
    const legacy = new Database(filePath, { create: true, strict: true })
    legacy.run(`
      CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
      CREATE TABLE clients (
        internal_id TEXT PRIMARY KEY,
        token TEXT NOT NULL UNIQUE,
        desired_revision INTEGER NOT NULL DEFAULT 0,
        last_applied_revision INTEGER NOT NULL DEFAULT 0,
        revocation_pending INTEGER NOT NULL DEFAULT 0,
        created_at TEXT NOT NULL,
        rotated_at TEXT
      );
      INSERT INTO meta(key, value) VALUES('schema_version', '1');
      INSERT INTO clients(internal_id, token, created_at) VALUES('legacy-client', 'legacy-token', '2026-01-01T00:00:00.000Z');
    `)
    legacy.close(false)

    const migrated = new TunnelDatabase(filePath)
    expect(migrated.sqlite.query<{ remark: string }, []>('SELECT remark FROM clients WHERE internal_id = \'legacy-client\'').get()).toEqual({ remark: '' })
    expect(migrated.sqlite.query<{ value: string }, []>('SELECT value FROM meta WHERE key = \'schema_version\'').get()).toEqual({ value: '2' })
    migrated.close()
  })
})
