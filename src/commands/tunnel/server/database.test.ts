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

describe('TunnelDatabase schema', () => {
  test('creates the fresh account and ownership schema', () => {
    const database = new TunnelDatabase(':memory:')
    const tables = database.sqlite.query<{ name: string }, []>('SELECT name FROM sqlite_master WHERE type = \'table\' ORDER BY name').all().map(row => row.name)
    const clientColumns = database.sqlite.query<{ name: string }, []>('PRAGMA table_info(clients)').all().map(row => row.name)
    const tunnelColumns = database.sqlite.query<{ name: string }, []>('PRAGMA table_info(tunnels)').all().map(row => row.name)
    expect(tables).toContain('accounts')
    expect(tables).toContain('tunnel_http_routes')
    expect(clientColumns).toContain('owner_account_id')
    expect(tunnelColumns).toContain('location')
    expect(tunnelColumns).toContain('options_json')
    expect(database.sqlite.query<{ value: string }, []>('SELECT value FROM meta WHERE key = \'schema_version\'').get()).toEqual({ value: '5' })
    database.close()
  })

  test('rejects an old development schema instead of migrating it', async () => {
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

    expect(() => new TunnelDatabase(filePath)).toThrow('incompatible')
  })

  test('migrates schema 3 HTTP tunnels to catch-all route reservations', async () => {
    const directory = await mkdtemp(path.join(os.tmpdir(), 'ycy-tunnel-database-v3-'))
    temporaryDirectories.push(directory)
    const filePath = path.join(directory, 'tunnel.sqlite')
    const legacy = new Database(filePath, { create: true, strict: true })
    legacy.run(`
      CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
      CREATE TABLE clients (internal_id TEXT PRIMARY KEY, desired_revision INTEGER NOT NULL DEFAULT 0);
      CREATE TABLE tunnels (
        id TEXT PRIMARY KEY,
        client_internal_id TEXT NOT NULL REFERENCES clients(internal_id) ON DELETE CASCADE,
        protocol TEXT NOT NULL,
        hostname TEXT,
        server_port INTEGER,
        local_host TEXT NOT NULL,
        local_port INTEGER NOT NULL,
        enabled INTEGER NOT NULL,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL
      );
      CREATE UNIQUE INDEX tunnels_unique_http_hostname ON tunnels(lower(hostname)) WHERE protocol = 'http';
      CREATE UNIQUE INDEX tunnels_unique_transport_port ON tunnels(protocol, server_port) WHERE protocol IN ('tcp', 'udp');
      CREATE INDEX tunnels_by_client ON tunnels(client_internal_id, created_at, id);
      INSERT INTO meta(key, value) VALUES('schema_version', '3');
      INSERT INTO clients VALUES('client-id', 4);
      INSERT INTO tunnels VALUES('http-id', 'client-id', 'http', 'App.Example.com', NULL, '127.0.0.1', 3000, 1, 'created', 'updated');
    `)
    legacy.close(false)

    const database = new TunnelDatabase(filePath)
    expect(database.sqlite.query<{ value: string }, []>('SELECT value FROM meta WHERE key = \'schema_version\'').get()).toEqual({ value: '5' })
    expect(database.sqlite.query<{ custom_domains: string, location: string | null }, []>('SELECT custom_domains, location FROM tunnels WHERE id = \'http-id\'').get()).toEqual({ custom_domains: '["app.example.com"]', location: null })
    expect(database.sqlite.query<{ hostname: string, location: string }, []>('SELECT hostname, location FROM tunnel_http_routes').all()).toEqual([{ hostname: 'app.example.com', location: '' }])
    expect(database.sqlite.query<{ desired_revision: number }, []>('SELECT desired_revision FROM clients WHERE internal_id = \'client-id\'').get()).toEqual({ desired_revision: 5 })
    database.close()
  })

  test('splits schema 4 HTTP locations into independently managed Tunnel Definitions', async () => {
    const directory = await mkdtemp(path.join(os.tmpdir(), 'ycy-tunnel-database-v4-'))
    temporaryDirectories.push(directory)
    const filePath = path.join(directory, 'tunnel.sqlite')
    const legacy = new Database(filePath, { create: true, strict: true })
    legacy.run(`
      CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
      CREATE TABLE clients (internal_id TEXT PRIMARY KEY, desired_revision INTEGER NOT NULL DEFAULT 0);
      CREATE TABLE tunnels (
        id TEXT PRIMARY KEY,
        client_internal_id TEXT NOT NULL REFERENCES clients(internal_id) ON DELETE CASCADE,
        protocol TEXT NOT NULL,
        custom_domains TEXT,
        locations TEXT NOT NULL DEFAULT '[]',
        server_port INTEGER,
        local_host TEXT NOT NULL,
        local_port INTEGER NOT NULL,
        enabled INTEGER NOT NULL,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL
      );
      CREATE TABLE tunnel_http_routes (tunnel_id TEXT NOT NULL, hostname TEXT NOT NULL, location TEXT NOT NULL, PRIMARY KEY(hostname, location));
      CREATE UNIQUE INDEX tunnels_unique_transport_port ON tunnels(protocol, server_port) WHERE protocol IN ('tcp', 'udp');
      CREATE INDEX tunnels_by_client ON tunnels(client_internal_id, created_at, id);
      INSERT INTO meta(key, value) VALUES('schema_version', '4');
      INSERT INTO clients VALUES('client-id', 7);
      INSERT INTO tunnels VALUES('http-id', 'client-id', 'http', '["app.example.com","alias.example.com"]', '["/one","/two"]', NULL, '127.0.0.1', 3000, 0, 'created', 'updated');
      INSERT INTO tunnel_http_routes VALUES('http-id', 'app.example.com', '/one');
      INSERT INTO tunnel_http_routes VALUES('http-id', 'app.example.com', '/two');
      INSERT INTO tunnel_http_routes VALUES('http-id', 'alias.example.com', '/one');
      INSERT INTO tunnel_http_routes VALUES('http-id', 'alias.example.com', '/two');
    `)
    legacy.close(false)

    const database = new TunnelDatabase(filePath)
    const rows = database.sqlite.query<{ id: string, custom_domains: string, location: string, enabled: number }, []>('SELECT id, custom_domains, location, enabled FROM tunnels ORDER BY location').all()
    expect(rows).toHaveLength(2)
    expect(rows.map(row => row.location)).toEqual(['/one', '/two'])
    expect(rows.every(row => row.custom_domains === '["app.example.com","alias.example.com"]' && row.enabled === 0)).toBe(true)
    expect(rows.some(row => row.id === 'http-id')).toBe(true)
    expect(database.sqlite.query<{ hostname: string, location: string }, []>('SELECT hostname, location FROM tunnel_http_routes ORDER BY hostname, location').all()).toEqual([
      { hostname: 'alias.example.com', location: '/one' },
      { hostname: 'alias.example.com', location: '/two' },
      { hostname: 'app.example.com', location: '/one' },
      { hostname: 'app.example.com', location: '/two' },
    ])
    expect(database.sqlite.query<{ desired_revision: number }, []>('SELECT desired_revision FROM clients WHERE internal_id = \'client-id\'').get()).toEqual({ desired_revision: 8 })
    database.close()
  })
})
