import { mkdirSync } from 'node:fs'
import path from 'node:path'
import { Database } from 'bun:sqlite'
import { TunnelError } from '../types'

const SCHEMA_VERSION = 5

const TUNNEL_SCHEMA = `
CREATE TABLE IF NOT EXISTS tunnels (
  id TEXT PRIMARY KEY,
  client_internal_id TEXT NOT NULL REFERENCES clients(internal_id) ON DELETE CASCADE,
  label TEXT NOT NULL DEFAULT '' CHECK (length(label) <= 100),
  protocol TEXT NOT NULL CHECK (protocol IN ('http', 'tcp', 'udp')),
  custom_domains TEXT,
  location TEXT,
  server_port INTEGER,
  local_host TEXT NOT NULL,
  local_port INTEGER NOT NULL CHECK (local_port BETWEEN 1 AND 65535),
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  options_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(options_json)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  CHECK (
    (protocol = 'http' AND custom_domains IS NOT NULL AND server_port IS NULL)
    OR (protocol IN ('tcp', 'udp') AND custom_domains IS NULL AND location IS NULL AND server_port IS NOT NULL)
  )
);

CREATE TABLE IF NOT EXISTS tunnel_http_routes (
  tunnel_id TEXT NOT NULL REFERENCES tunnels(id) ON DELETE CASCADE,
  hostname TEXT NOT NULL COLLATE NOCASE,
  location TEXT NOT NULL,
  PRIMARY KEY(hostname, location)
);

CREATE UNIQUE INDEX IF NOT EXISTS tunnels_unique_transport_port
ON tunnels(protocol, server_port) WHERE protocol IN ('tcp', 'udp');

CREATE INDEX IF NOT EXISTS tunnels_by_client
ON tunnels(client_internal_id, created_at, id);
`

const SCHEMA = `
CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE accounts (
  internal_id TEXT PRIMARY KEY,
  kind TEXT NOT NULL CHECK (kind IN ('environment', 'local')),
  username TEXT NOT NULL,
  username_key TEXT NOT NULL UNIQUE,
  role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
  password_hash TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  CHECK (
    (kind = 'environment' AND role = 'admin' AND password_hash IS NULL)
    OR
    (kind = 'local' AND password_hash IS NOT NULL)
  )
);

CREATE UNIQUE INDEX accounts_single_environment
ON accounts(kind) WHERE kind = 'environment';

CREATE TABLE IF NOT EXISTS clients (
  internal_id TEXT PRIMARY KEY,
  owner_account_id TEXT NOT NULL REFERENCES accounts(internal_id) ON DELETE RESTRICT,
  remark TEXT NOT NULL DEFAULT '' CHECK (length(remark) <= 100),
  token TEXT NOT NULL UNIQUE,
  desired_revision INTEGER NOT NULL DEFAULT 0 CHECK (desired_revision >= 0),
  last_applied_revision INTEGER NOT NULL DEFAULT 0 CHECK (last_applied_revision >= 0),
  revocation_pending INTEGER NOT NULL DEFAULT 0 CHECK (revocation_pending IN (0, 1)),
  created_at TEXT NOT NULL,
  rotated_at TEXT
);

${TUNNEL_SCHEMA}

CREATE INDEX clients_by_owner
ON clients(owner_account_id, created_at, internal_id);
`

function bumpMigratedClients(database: Database): void {
  database.run(`
    UPDATE clients SET desired_revision = desired_revision + 1
    WHERE EXISTS (SELECT 1 FROM tunnels WHERE tunnels.client_internal_id = clients.internal_id)
  `)
}

function rebuildHttpRouteReservations(database: Database): void {
  database.run(`
    INSERT INTO tunnel_http_routes(tunnel_id, hostname, location)
    SELECT tunnels.id, domains.value, COALESCE(tunnels.location, '')
    FROM tunnels, json_each(tunnels.custom_domains) AS domains
    WHERE tunnels.protocol = 'http'
  `)
}

function migrateSchema3(database: Database): void {
  database.run(`
    DROP INDEX IF EXISTS tunnels_unique_http_hostname;
    DROP INDEX IF EXISTS tunnels_unique_transport_port;
    DROP INDEX IF EXISTS tunnels_by_client;
    ALTER TABLE tunnels RENAME TO tunnels_schema_3;
  `)
  database.run(TUNNEL_SCHEMA)
  database.run(`
    INSERT INTO tunnels(id, client_internal_id, label, protocol, custom_domains, location, server_port, local_host, local_port, enabled, options_json, created_at, updated_at)
    SELECT id, client_internal_id, '', protocol,
      CASE WHEN protocol = 'http' THEN json_array(lower(hostname)) END,
      NULL, server_port, local_host, local_port, enabled, '{}', created_at, updated_at
    FROM tunnels_schema_3;
    DROP TABLE tunnels_schema_3;
  `)
  rebuildHttpRouteReservations(database)
  bumpMigratedClients(database)
}

function migrateSchema4(database: Database): void {
  database.run(`
    DROP TABLE tunnel_http_routes;
    DROP INDEX IF EXISTS tunnels_unique_transport_port;
    DROP INDEX IF EXISTS tunnels_by_client;
    ALTER TABLE tunnels RENAME TO tunnels_schema_4;
  `)
  database.run(TUNNEL_SCHEMA)
  database.run(`
    INSERT INTO tunnels(id, client_internal_id, label, protocol, custom_domains, location, server_port, local_host, local_port, enabled, options_json, created_at, updated_at)
    SELECT
      CASE WHEN route.key = 0 THEN legacy.id ELSE lower(hex(randomblob(16))) END,
      legacy.client_internal_id, '', 'http', legacy.custom_domains,
      CAST(route.value AS TEXT), NULL, legacy.local_host, legacy.local_port,
      legacy.enabled, '{}', legacy.created_at, legacy.updated_at
    FROM tunnels_schema_4 AS legacy,
      json_each(CASE WHEN json_array_length(legacy.locations) = 0 THEN '[null]' ELSE legacy.locations END) AS route
    WHERE legacy.protocol = 'http';

    INSERT INTO tunnels(id, client_internal_id, label, protocol, custom_domains, location, server_port, local_host, local_port, enabled, options_json, created_at, updated_at)
    SELECT id, client_internal_id, '', protocol, NULL, NULL, server_port, local_host, local_port, enabled, '{}', created_at, updated_at
    FROM tunnels_schema_4 WHERE protocol IN ('tcp', 'udp');

    DROP TABLE tunnels_schema_4;
  `)
  rebuildHttpRouteReservations(database)
  bumpMigratedClients(database)
}

export class TunnelDatabase {
  readonly sqlite: Database

  constructor(readonly filePath: string) {
    if (filePath !== ':memory:')
      mkdirSync(path.dirname(filePath), { recursive: true })
    this.sqlite = new Database(filePath, { create: true, strict: true })
    this.sqlite.run('PRAGMA foreign_keys = ON')
    this.sqlite.run('PRAGMA journal_mode = WAL')
    this.sqlite.run('PRAGMA busy_timeout = 5000')
    try {
      this.initialize()
    }
    catch (cause) {
      this.sqlite.close(false)
      throw cause
    }
  }

  private initialize(): void {
    const initialize = this.sqlite.transaction(() => {
      const hasMeta = this.sqlite.query<{ name: string }, []>('SELECT name FROM sqlite_master WHERE type = \'table\' AND name = \'meta\'').get()
      if (hasMeta) {
        const current = this.sqlite.query<{ value: string }, []>('SELECT value FROM meta WHERE key = \'schema_version\'').get()
        const version = Number(current?.value)
        if (version > SCHEMA_VERSION)
          throw new TunnelError('DATABASE_TOO_NEW', `Tunnel database schema ${current!.value} requires a newer ycy release`)
        if (version === 3 || version === 4) {
          if (version === 3)
            migrateSchema3(this.sqlite)
          else
            migrateSchema4(this.sqlite)
          this.sqlite.query('UPDATE meta SET value = ? WHERE key = \'schema_version\'').run(String(SCHEMA_VERSION))
          return
        }
        if (version !== SCHEMA_VERSION)
          throw new TunnelError('DATABASE_INCOMPATIBLE', `Tunnel database schema ${current?.value ?? 'unknown'} is incompatible; remove the development database and restart`)
        return
      }
      this.sqlite.run(SCHEMA)
      this.sqlite.query('INSERT INTO meta(key, value) VALUES(\'schema_version\', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value').run(String(SCHEMA_VERSION))
    })
    initialize.immediate()
  }

  close(): void {
    this.sqlite.close(false)
  }
}

export function openTunnelDatabase(dataDirectory: string): TunnelDatabase {
  return new TunnelDatabase(path.join(dataDirectory, 'tunnel.sqlite'))
}
