import { mkdirSync } from 'node:fs'
import path from 'node:path'
import { Database } from 'bun:sqlite'
import { TunnelError } from '../types'

const SCHEMA_VERSION = 2

const SCHEMA = `
CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS clients (
  internal_id TEXT PRIMARY KEY,
  remark TEXT NOT NULL DEFAULT '' CHECK (length(remark) <= 100),
  token TEXT NOT NULL UNIQUE,
  desired_revision INTEGER NOT NULL DEFAULT 0 CHECK (desired_revision >= 0),
  last_applied_revision INTEGER NOT NULL DEFAULT 0 CHECK (last_applied_revision >= 0),
  revocation_pending INTEGER NOT NULL DEFAULT 0 CHECK (revocation_pending IN (0, 1)),
  created_at TEXT NOT NULL,
  rotated_at TEXT
);

CREATE TABLE IF NOT EXISTS tunnels (
  id TEXT PRIMARY KEY,
  client_internal_id TEXT NOT NULL REFERENCES clients(internal_id) ON DELETE CASCADE,
  protocol TEXT NOT NULL CHECK (protocol IN ('http', 'tcp', 'udp')),
  hostname TEXT,
  server_port INTEGER,
  local_host TEXT NOT NULL,
  local_port INTEGER NOT NULL CHECK (local_port BETWEEN 1 AND 65535),
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  CHECK (
    (protocol = 'http' AND hostname IS NOT NULL AND server_port IS NULL)
    OR
    (protocol IN ('tcp', 'udp') AND hostname IS NULL AND server_port IS NOT NULL)
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS tunnels_unique_http_hostname
ON tunnels(lower(hostname)) WHERE protocol = 'http';

CREATE UNIQUE INDEX IF NOT EXISTS tunnels_unique_transport_port
ON tunnels(protocol, server_port) WHERE protocol IN ('tcp', 'udp');

CREATE INDEX IF NOT EXISTS tunnels_by_client
ON tunnels(client_internal_id, created_at, id);
`

export class TunnelDatabase {
  readonly sqlite: Database

  constructor(readonly filePath: string) {
    if (filePath !== ':memory:')
      mkdirSync(path.dirname(filePath), { recursive: true })
    this.sqlite = new Database(filePath, { create: true, strict: true })
    this.sqlite.run('PRAGMA foreign_keys = ON')
    this.sqlite.run('PRAGMA journal_mode = WAL')
    this.sqlite.run('PRAGMA busy_timeout = 5000')
    this.migrate()
  }

  private migrate(): void {
    const migrate = this.sqlite.transaction(() => {
      this.sqlite.run(SCHEMA)
      const current = this.sqlite.query<{ value: string }, []>('SELECT value FROM meta WHERE key = \'schema_version\'').get()
      if (current && Number(current.value) > SCHEMA_VERSION)
        throw new TunnelError('DATABASE_TOO_NEW', `Tunnel database schema ${current.value} requires a newer ycy release`)
      if (current && Number(current.value) < 2)
        this.sqlite.run('ALTER TABLE clients ADD COLUMN remark TEXT NOT NULL DEFAULT \'\' CHECK (length(remark) <= 100)')
      this.sqlite.query('INSERT INTO meta(key, value) VALUES(\'schema_version\', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value').run(String(SCHEMA_VERSION))
    })
    migrate.immediate()
  }

  close(): void {
    this.sqlite.close(false)
  }
}

export function openTunnelDatabase(dataDirectory: string): TunnelDatabase {
  return new TunnelDatabase(path.join(dataDirectory, 'tunnel.sqlite'))
}
