# Tunnel command compatibility inventory

> **Roadmap transition-scope amendment (2026-08-22):** this inventory remains the factual source for normal Tunnel behavior, but first-release acceptance covers Go client to Go server on protocol v3 only. Remembered connections inside `config.json` retain Bun-written compatibility; mixed Bun/Go peers and Bun-written sessions, SQLite, client cache, and generated runtime state have no carryover gate. Normal startup diagnostics and internal operator recovery apply.

This inventory records the observable contract of `ycy tunnel server` and `ycy tunnel connect` before the Bun CLI, TypeScript control plane, native-agent implementation, and tests are frozen under `legacy/bun/`. The retained React application moves to `web/tunnel-server`; it remains an active client of the Go HTTP service. This document is a migration-specification input, not approval to reproduce confirmed security, resource, persistence, or platform defects.

## First-release scope

The first Go release is parity-first: actual Bun control-plane, client, fresh-state persistence, FRP, and protocol-v3 behavior remains the test baseline even where this inventory labels it defective. Future protocol, trust, resource, schema, permission, and process hardening is deferred. Only a demonstrated Go-to-Go v3, `config.json`, platform, dependency, or FRP mismatch may create a narrow compatibility exception.

## Contract classification

The migration uses three compatibility levels:

1. **Exact, data, or protocol compatibility:** command and option names; environment names and precedence; URL normalization and remembered-connection selection; Bun-written fields inside `config.json`; fresh-Go session, client-state, and SQLite behavior; JSON fields and omission rules consumed by the retained frontend; Go-to-Go WebSocket protocol v3; FRP version and artifact identity; routing reservations; revision semantics; and authentication-first client startup.
2. **Intent compatibility:** terminal color and exact log wording; non-machine-consumed UI copy and layout; generated TOML whitespace and key order; timing that is only an implementation grace period; and visual presentation may change while retaining the same operational workflow and state meaning.
3. **Post-parity defect:** accepting and overwriting an unknown SQLite schema version, insufficient protection of secret-bearing directories and database files, unbounded login/body/collection/snapshot/SSE work, an agent connection with no hello deadline, non-cancellable reconnect sleeps, unlocked and unbounded FRP installation, and unverified Windows child-tree/ACL behavior are later hardening findings, not first-release redesign requirements.

The first-release rule is to preserve currently accepted behavior for Go client to Go server on protocol v3. Only remembered connections inside Bun-written `config.json` are read directly; a proven mismatch there may invoke the map's narrow compatibility process. Every other Tunnel store starts with Go-owned state and receives no Bun fixture, detection, migration, or mixed-runtime gate.

## Verification baseline

The local baseline was run on 2026-08-22 with Bun 1.3.14:

```text
bun test src/commands/tunnel

105 pass
0 fail
577 expect() calls
18 files
11.08s
```

`acceptance.ts` is intentionally not named `*.test.ts`, is not selected by that directory command, and has no package script. It was therefore run explicitly:

```text
bun test ./src/commands/tunnel/acceptance.ts

1 pass
0 fail
3 expect() calls
4.86s
```

The real-FRP acceptance used pinned FRP 0.70.1 and proved two simultaneous trusted clients plus HTTP Host routing, TCP, UDP, route changes, and the managed custom FRP 404 page. Neither command built a standalone ycy artifact or exercised another native OS/architecture.

## Domain and ownership model

Use these names consistently in the Go specification and tests:

- **Tunnel Control Plane:** the browser/API/agent HTTP listener owned by `tunnel server`.
- **Deployment Administrator:** the stable `environment-admin` account whose username and password come from the server environment. Its password is never stored in SQLite.
- **Local Account:** a SQLite-backed `admin` or `user` identity with an Argon2id PHC password hash.
- **Trusted Tunnel Client:** the durable server record identified externally by its current recoverable Client Token and internally by a UUID.
- **Client Connection Instance:** one local `(normalized control-plane origin, raw Client Token)` pair, with one opaque state directory and at most one foreground supervisor.
- **Tunnel Definition:** one typed HTTP, TCP, or UDP proxy in a client's complete desired snapshot.
- **Desired Revision:** the durable monotonically increasing revision of a client's complete Tunnel Definition set.
- **Applied Revision:** the greatest durable revision that the agent reported as successfully activated.
- **Internal FRP Token:** the one server-wide FRP data-plane token, distinct from every Client Token.
- **Resource Owner:** the immutable account that created a Trusted Tunnel Client; every tunnel inherits that ownership.
- **Resource Reservation:** an HTTP `(normalized hostname, exact location)` pair or a `(protocol, server port)` pair. Disabled tunnels continue to reserve resources.

The runtime topology is:

```text
ycy tunnel server
  -> state-directory and session locks
  -> SQLite control-plane truth
  -> HTTP JSON + SSE + React shell
  -> WebSocket agent gateway
  -> generated frps.toml
  -> supervised pinned frps

ycy tunnel connect
  -> remembered-connection catalog and interactive resolution
  -> opaque instance lock and cached rollback state
  -> authenticated WebSocket agent
  -> generated frpc.toml
  -> supervised pinned frpc
```

The future Go package graph is not selected here. SQLite transactions remain the sole desired-state truth; the agent gateway must not become a second persistence layer, and the reconciler must remain the only owner of frpc activation and rollback.

## Shared logging contract

The root option is `ycy --log-level <debug|info|warn|error> tunnel ...`; neither tunnel subcommand defines a local log-level flag. Precedence is root CLI, `YCY_LOG_LEVEL`, then `info`.

- Logs go to `stderr`, one record per line. Redirected output is plain; interactive output colors only the timestamp, level, and scope prefix.
- Records contain an ISO timestamp, uppercase padded level, optional dotted scope, one-line message, and optional JSON context/error. Exact wording and color are intent-compatible; stream, level filtering, and secret redaction are retained behavior.
- Context keys matching authorization, cookie, password, secret, or token are redacted. Similar assignments and Bearer values are redacted in messages and serialized errors.
- frpc/frps stdout is logged at `info`; stderr is logged at `warn`, line by line. ycy retains no log history.
- The resolved ycy log level is also written into generated FRP configuration. Complete configurations, request bodies, tokens, and passwords must never be logged.

## `tunnel server` CLI contract

### Command surface and precedence

The exact leaf is:

```text
ycy tunnel server [options]
```

Non-secret values use CLI, then environment, then default precedence:

| CLI option | Environment | Default and validation |
| --- | --- | --- |
| `--address <address>` | `YCY_TUNNEL_ADDRESS` | `0.0.0.0`; trimmed and non-empty, otherwise passed through as the bind address |
| `--control-port <port>` | `YCY_TUNNEL_CONTROL_PORT` | `7500`; strict `Number`, safe integer 1-65535 |
| `--frp-port <port>` | `YCY_TUNNEL_FRP_PORT` | `7000`; strict safe integer 1-65535 |
| `--http-port <port>` | `YCY_TUNNEL_HTTP_PORT` | `8080`; strict safe integer 1-65535 |
| `--port-range <start-end>` | `YCY_TUNNEL_PORT_RANGE` | `20000-20100`; trimmed decimal `start-end`, each 1-65535, start <= end |
| `--advertise-frp-addr <host:port>` | `YCY_TUNNEL_ADVERTISE_FRP_ADDR` | absent; hostname/IPv4 uses one colon, IPv6 must be `[IPv6]:port`; credentials, URL delimiters, query, and fragment are rejected |
| `--data-dir <path>` | `YCY_TUNNEL_DATA_DIR` | platform state root plus `ycy/tunnel/server`; made absolute with `path.resolve` |
| `--session-idle-days <days>` | `YCY_TUNNEL_SESSION_IDLE_DAYS` | `7`; positive safe integer days |

The control, FRP bind, and FRP HTTP ports must be distinct. None may fall inside the Server Port Pool. Port zero is not accepted.

Environment-only values are:

| Environment | Contract |
| --- | --- |
| `YCY_TUNNEL_ADMIN_USER` | Defaults to `admin`; 1-64 ASCII letters, digits, dot, underscore, or hyphen; not trimmed |
| `YCY_TUNNEL_ADMIN_PASSWORD` | Required; 5-256 JavaScript UTF-16 code units |
| `YCY_TUNNEL_FRP_TOKEN` | Optional fixed Internal FRP Token; trimmed and required to remain non-empty when present |

If `YCY_TUNNEL_FRP_TOKEN` is absent, the first server run generates 32 random bytes as unpadded base64url and persists the result in SQLite `meta`. A later run reads it unchanged. This token configures ycy's managed frps and every trusted frpc; it is not an external-frps connection option and is never returned by the browser API.

The legacy day parser validates the input before multiplying by milliseconds but does not validate the product or JavaScript timer/date range. Very large positive safe integers can therefore produce unsafe timing behavior. Preserve that parser behavior for first-release parity; a replacement limit is post-parity hardening.

### Listener and startup behavior

One `--address` applies to all listeners:

| Listener | Default | Purpose |
| --- | ---: | --- |
| Control HTTP | `0.0.0.0:7500/tcp` | React shell, management JSON/SSE, agent WebSocket, liveness |
| FRP bind | `0.0.0.0:7000/tcp` | frpc data-plane sessions |
| FRP HTTP vhost | `0.0.0.0:8080/tcp` | Host-header HTTP tunnel routing |
| Server Port Pool | `0.0.0.0:20000-20100/tcp+udp` | public TCP and UDP tunnels |

Startup acquires the server data-directory `.lock` first, opens SQLite, obtains the Internal FRP Token, opens the persistent session manager, synchronizes the Deployment Administrator, and starts the control HTTP listener. frps preparation and activation then run in the background. A missing/corrupt binary, invalid generated configuration, or listener conflict leaves the control plane and lock available with frps in `stopped` plus a structured error; an administrator may retry Start or Restart.

State-directory, database, session, bind, and other composition failures are logged and propagate to the root CLI as nonzero failures. There is no browser auto-open and no decorative terminal UI.

`SIGINT` and `SIGTERM` use the same idempotent path: abort bootstrap, close agent sockets and HTTP, close sessions, stop frps, close SQLite, release the directory lock, and log the stop. The code has no process-wide shutdown deadline. Go must preserve ordered ownership while adding the bounded/native behavior selected by the follow-up decision.

## `tunnel connect` resolution contract

### Command surface

The exact leaf is:

```text
ycy tunnel connect [--server <control-plane>] [--token <client-token>]
```

Resolution is field-aware rather than one flat configuration object:

| Value | Sources in precedence order |
| --- | --- |
| Server | `--server`, `YCY_TUNNEL_SERVER`, remembered-pair selection, compile-time `DEFAULT_TUNNEL_SERVER` |
| Token | `--token`, `YCY_TUNNEL_TOKEN`, UTF-8 contents of `YCY_TUNNEL_TOKEN_FILE`, remembered-pair selection |

`DEFAULT_TUNNEL_SERVER` is currently the empty compile-time string. A token file path is made absolute, read completely as UTF-8, and trimmed. A provided but empty server or token is an error. Client Tokens are otherwise treated as opaque non-empty strings; `connect` does not require the server-generated `ycy_` shape.

A scheme-less server becomes `https://<value>`. Explicit `http://` remains supported. The normalized value must be an HTTP(S) origin with a hostname and no user information, path other than `/`, query, or fragment. WHATWG URL origin normalization removes the root path and default port from instance identity.

### Remembered-pair selection

The catalog is sorted by newest successful authentication and then stable ID. Resolution behaves as follows:

- Complete explicit server and token values bypass selection.
- A server without a token filters remembered pairs to that normalized origin.
- A token without a server first selects remembered pairs containing the same decrypted token. If it is a new token, every distinct remembered origin becomes a token-rotation candidate.
- With neither field, all valid remembered pairs are candidates.
- One candidate is selected automatically. Several candidates require an interactive selector and show the origin plus a token mask containing the first eight and last four characters. A non-TTY process fails instead of guessing and instructs the caller to provide both values.
- Cancelling the selector prints the Clack cancellation, creates no instance state, changes no configuration, and returns normally.
- If no candidate supplies the missing field, the empty compile-time default does not help and the command fails with a precise missing-server or missing-token configuration error.

Only a successful protocol welcome triggers recency persistence. An explicit CLI token opts the pair into insertion. Reuse of an already remembered pair updates its recency regardless of how its fields were resolved. A new secret sourced only from `YCY_TUNNEL_TOKEN` or `YCY_TUNNEL_TOKEN_FILE` is not newly remembered. A configuration write failure is logged as a warning and does not stop an authenticated client.

### Foreground lifecycle

After resolution, the command acquires its Client Connection Instance lock and reads cached state only to report `lastAppliedRevision` during authentication. It then probes the Bearer token, opens the agent WebSocket, validates the welcome and pinned FRP artifact, prepares frpc, applies the complete snapshot, and remains in the foreground.

- The same normalized server/token pair maps to the same lock and a second process fails `INSTANCE_ACTIVE`.
- Changing either field creates an independent instance directory and allows a separate frpc child.
- Ordinary control-link failure leaves an already active frpc running and reconnects after `1, 2, 4, 8, 15, 30` seconds, capped at 30 seconds.
- Authentication rejection, explicit revoke, or protocol incompatibility stops frpc and terminates with an error where applicable.
- `SIGINT` and `SIGTERM` close the socket, stop frpc, release the instance lock, and log the stop.

The reconnect sleep is not abort-aware, so a signal during backoff can delay shutdown by up to 30 seconds. Preserve the observable delay for first-release parity unless Go cancellation mechanics prove it cannot be reproduced; making it abort-aware is post-parity work.

## Remembered connection and instance identity formats

### Shared configuration file

Remembered connections live inside `~/.ycy-cli/config.json` at:

```text
tunnel.connections.<instance-id> = {
  server,
  token,
  lastAuthenticatedAt
}
```

`server` is the normalized origin, `token` is encrypted text, and `lastAuthenticatedAt` is an ISO timestamp. At most 32 valid pairs remain after a successful write; newest authentication wins and older entries are evicted. Invalid IDs, origins, timestamps, ciphertext, decrypted empty tokens, and ID/token mismatches are ignored independently, so one corrupt entry does not hide the rest. The older single remembered-connection schema is deliberately ignored by the current product and does not need to be revived.

Configuration mutation uses the shared `.config.lock` directory, a PID/UUID owner record, a ten-second retry window, and atomic candidate rename. A Go writer must coordinate with that existing lock during cutover and preserve unrelated `fork` and `cm` configuration rather than treating Tunnel as the whole file.

### Encryption and stable identity

The existing key contract is:

```text
PBKDF2-HMAC-SHA256(
  passphrase = <machine-id> + ":" + <OS username>,
  salt = base64-decoded config.salt,
  iterations = 100000,
  length = 32 bytes
)
```

Machine identity is the macOS `IOPlatformUUID` from `ioreg`, Linux `/etc/machine-id`, or Windows `MachineGuid` from `reg`; failure falls back to `<hostname>-<username>`. `config.salt` is generated as 32 random bytes in standard base64. Encryption is AES-256-GCM with a 16-byte IV and text encoding:

```text
base64(iv):base64(authentication-tag):base64(ciphertext)
```

The stable instance ID is:

```text
v1_ + base64url_unpadded(
  HMAC-SHA256(
    configuration-key,
    "ycy:tunnel-client-instance:v1\0" + normalized-origin + "\0" + raw-token
  )
)
```

Its pattern is `v1_` plus 43 base64url characters. Go can reproduce every primitive directly. Direct-read tests cover machine-ID fallback, username lookup, permissive legacy base64 decoding, missing/corrupt salt behavior, default-port origin normalization, and existing ciphertext only where these values are stored in `config.json`. Bun-written instance directories are not read or migrated; Go-created instance identity must remain stable across later Go runs.

## Platform paths, locks, and sensitive files

The platform state root is `%LOCALAPPDATA%` or `%USERPROFILE%/AppData/Local` on Windows, `$HOME/Library/Application Support` on macOS, and `$XDG_STATE_HOME` or `$HOME/.local/state` elsewhere.

| State | Default location and important entries |
| --- | --- |
| Server | `<state-root>/ycy/tunnel/server`; `.lock/owner.json`, `tunnel.sqlite` plus WAL/SHM, `sessions/`, `frps.toml`, optional `404.html` |
| Client root | `<state-root>/ycy/tunnel/client`; `.instances.lock/`, versioned instance directories |
| Client instance | `<client-root>/v1_<digest>`; `.lock/owner.json`, `last-applied.json`, `frpc.toml` |
| Managed FRP | `<state-root>/ycy/frp/0.70.1/{frpc,frps}` or `.exe`; `/opt/ycy/frp/0.70.1` only when `YCY_TUNNEL_DOCKER=1` |

State and registry locks are directories containing a JSON owner with UUID, PID, timestamp, and state path. A lock with a live PID is active; a missing/malformed owner receives a one-second publication grace and is then renamed away as stale. The per-instance lock is held for the process lifetime; the registry lock waits up to ten seconds and is held only while acquiring an instance and cleaning siblings.

Recognized `v1_` instance directories older than 90 days by directory mtime are removed only when not current and not actively locked. Current, active, recent, unrecognized, and legacy entries are untouched. Cleanup failure only warns.

`frps.toml`, `frpc.toml`, and `last-applied.json` contain the Internal FRP Token and may contain HTTP Basic Auth passwords. SQLite stores Client Tokens, the Internal FRP Token, and recoverable HTTP Basic Auth values in plaintext. The generic atomic writer requests mode 0600 for files, and the shared session manager explicitly uses directory 0700/files 0600 on POSIX. The server/client state directories and SQLite database/WAL files are not explicitly hardened, and there is no Windows ACL policy. Reproduce the observed publication modes for first-release parity; broader directory and Windows ACL protection is post-parity hardening.

## Server SQLite contract

The database is the standard SQLite file `<data-dir>/tunnel.sqlite`. Every open enables foreign keys, WAL journal mode, and a 5000 ms busy timeout. IDs are UUID strings; timestamps are JavaScript ISO UTC strings with millisecond precision. The durable v1 shape is:

| Table | Durable fields and constraints |
| --- | --- |
| `meta` | `key` primary key, `value`; currently `schema_version` and `internal_frp_token` |
| `accounts` | internal ID; `environment|local`; display username; unique lowercased key; `admin|user`; nullable local Argon2 hash; created/updated timestamps; one environment row |
| `clients` | internal ID; owner FK with restrict delete; trimmed remark; unique recoverable token; desired/applied revisions; revocation flag; created/rotated timestamps |
| `tunnels` | ID; client FK with cascade; label; `http|tcp|udp`; JSON domains; nullable location; nullable port; local endpoint; enabled; typed `options_json`; timestamps and type checks |
| `tunnel_http_routes` | tunnel FK with cascade; case-insensitive hostname; non-null location where empty text represents catch-all; primary key `(hostname, location)` |

Partial indexes enforce one environment account, one TCP reservation per port, one UDP reservation per port, and client/owner lookup order. The same numeric port may be reserved once for TCP and once for UDP. The browser and agent operate on the same stored Client Token; hashing it in the Go schema would be an incompatible data and UI change unless a separately approved migration retained recoverability.

### Critical schema-version defect

Fresh initialization creates schema v1, then writes `meta.schema_version=1`. Existing startup does not read or validate that value. If `meta` exists, it skips all schema creation and unconditionally overwrites any declared version with `1`, even when the database is newer, unknown, or incomplete. Consequences include silent downgrade marking, later query failures, and possible opening of a database whose schema does not match the code.

The first Go release emulates this observed schema handling because it is part of the Bun behavior baseline. A fail-closed schema policy and backed migration are post-parity work unless a focused SQLite compatibility test proves the legacy behavior cannot be reproduced.

## Accounts, authentication, and persistent sessions

### Account invariants

- The Deployment Administrator always has internal ID `environment-admin`, kind `environment`, role `admin`, and no SQLite password hash. Restart updates its display/lowercase username and revokes old sessions through credential-revision mismatch. A collision with a local username fails startup.
- Local usernames are immutable, 1-64 ASCII word/dot/hyphen characters, and unique case-insensitively. Roles are exactly `admin` or `user`.
- Password validation uses 5-256 JavaScript UTF-16 code units. Local hashes are Bun Argon2id PHC strings with version 19, memory 65,536 KiB, time cost 3, and parallelism 1, for example `$argon2id$v=19$m=65536,t=3,p=1$...`.
- The environment password is hashed with the same parameters in process memory. Unknown/bad usernames verify against that hash before returning the same generic 401, avoiding a cheap unknown-user path.
- Password change/reset, role change, and account deletion immediately revoke every session for the affected account. The Deployment Administrator cannot be edited through the API. An account owning clients cannot be deleted.
- Users see and mutate only owned clients/tunnels; cross-owner IDs return the same not-found behavior as missing IDs. Administrators can operate all resources, manage accounts, and control frps without changing resource ownership.

Go's Argon2 primitive can reproduce the algorithm but does not by itself parse PHC strings. The Go authentication adapter must verify existing Bun hashes directly and emit the selected compatible PHC form for new hashes. Existing local accounts may not be reset or manually recreated.

### Session v1 format

Sessions live at `<data-dir>/sessions` and reuse the shared file-session v1 format:

- `.session-key`: 32 raw random bytes.
- `.session.lock`: one JSON line with an owner UUID and positive PID.
- `<token-hash>.json`: compact JSON plus newline; filename and `tokenHash` are lowercase SHA-256 of the raw 32-byte base64url cookie token.
- Session fields are `version:1`, `tokenHash`, account-ID `subject`, credential `revision`, `createdAt`, `lastAccessAt`, and `expiresAt`.

The Tunnel credential revision is unpadded base64url HMAC-SHA256 under `.session-key` over:

```text
account-id + NUL + lowercase-username + NUL + role + NUL + credential
```

The credential is the environment plaintext password or the persisted local PHC hash. Go can read and refresh this format directly.

The management cookie is `ycy_tunnel_session` with `HttpOnly; SameSite=Strict; Path=/; Max-Age=<remaining idle seconds>`. It has no `Secure` attribute. Idle lifetime defaults to seven days with no absolute lifetime. Every successful authenticated API resume synchronously rewrites and fsyncs the record, extending expiry. Limits are eight sessions per subject and 128 total, evicting least recently accessed entries. Startup prunes malformed, expired, interrupted, and misnamed records and refuses another live session-directory owner.

There is no login rate/concurrency limit. Cookie TLS/proxy policy, session bounds, Windows ACL semantics, and storage-failure behavior must be made explicit by the safe-contract and persistent-data decisions.

## Trusted clients, tunnels, and transactions

### Client and revision invariants

- Client IDs are UUIDs. New Client Tokens are `ycy_` plus 32 random base64url bytes and remain recoverable in authorized browser responses.
- Remarks are trimmed, preserve internal newlines, and contain at most 100 JavaScript code units. Updating a remark changes neither token nor Desired Revision.
- Rotation atomically replaces the token, sets `revocation_pending`, retains tunnels and revisions, sends a cooperative `revoke` to a connected old client, and clears the flag only when the replacement token opens a socket.
- Deletion sends cooperative revoke, cascade-deletes every tunnel, and releases reservations. An unreachable old agent may continue its existing FRP session until it reconnects and is rejected.
- Every create, update, or delete transaction increments Desired Revision once and emits a complete-snapshot invalidation only after commit. Applied Revision never exceeds Desired Revision and advances monotonically only on a successful agent acknowledgement.
- Runtime connection/process/error state is process-local. Only Desired Revision and greatest Applied Revision persist; no history, metrics, traffic, or logs are stored.

There are no server-side maxima for accounts, clients, tunnels per client, total routes, or snapshot bytes. This is not permission for the Go implementation to remain unbounded.

### HTTP Tunnel Definition

- `customDomains` contains 1-32 exact DNS hostnames. Values are trimmed, one trailing dot is removed, IDNs become lowercase ASCII, duplicates collapse, and schemes, paths, ports, IP literals, wildcard characters, whitespace, and invalid labels are rejected.
- `location` is nullable catch-all or an exact 1-2048-character URL path prefix starting with `/`, with no controls, whitespace, backslash, query, or fragment. Case and trailing slash are preserved. FRP matches literal prefixes and does not strip them.
- Every normalized domain times the one location is reserved globally. SQLite stores catch-all as empty text in the reservation table. Disabled tunnels reserve identically to enabled tunnels; overlapping but non-identical prefixes are allowed.
- The local host defaults to `127.0.0.1` and accepts the existing word/dot/colon/hyphen grammar up to 253 code units. Local port is 1-65535. ycy does not probe local endpoints.

### TCP and UDP Tunnel Definition

- Protocol is exactly `tcp` or `udp`; each selected server port must fall inside the configured pool.
- An omitted port is allocated transactionally as the lowest free number for that protocol. Import requires an explicit remote port.
- `(tcp, port)` and `(udp, port)` are independent unique keys. Disabled definitions retain reservations.

### Typed FRP options

Every definition owns normalized transport options: encryption and compression booleans; optional positive finite bandwidth plus `KB|MB` and `client|server`; and optional Proxy Protocol `v1|v2`. Health checks are TCP or HTTP with positive safe-integer interval, timeout, and failure threshold; HTTP additionally requires a slash-prefixed no-space path and up to 32 headers.

HTTP definitions additionally support recoverable Basic Auth, Host Header Rewrite, and up to 32 request and 32 response headers. Header names use HTTP token grammar, compare case-insensitively for duplicates, and are at most 128 code units; values are trimmed, at most 4096 code units, and contain no CR/LF. Basic Auth username/password are each non-empty and at most 256 code units. Browser projections redact the password to `{username,passwordConfigured:true}`, while agent snapshots retain it.

Switching protocol normalizes away HTTP-only fields; patching an existing HTTP Basic Auth object without a password preserves the current password. Go JSON decoding must preserve the API's omitted-versus-null patch semantics.

### TOML import transaction

The browser may preview a strict JSON body containing non-empty TOML source limited by schema to 1 MiB, then submit that same source with one or more candidate IDs. The parser accepts FRP v1 `proxies`, maps typed HTTP/TCP/UDP fields, expands each HTTP location into a separate candidate, reports unsupported fields/settings as ignored notices, and redacts Basic Auth passwords from preview.

All selected candidates are reparsed, validated against current ownership/reservations, created disabled in one immediate SQLite transaction, and increment Desired Revision exactly once. Duplicate or stale candidate IDs, a collision, or any invalid selected entry rolls back the complete batch. Client-level FRP connection settings and unsupported proxy types are never executed as raw configuration.

## Browser HTTP/JSON contract

### Common response shape

Management JSON is version 1. Errors use:

```json
{ "version": 1, "error": { "code": "UPPER_SNAKE_CASE", "message": "..." } }
```

JSON carries `Cache-Control: no-store`, `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`, `Referrer-Policy: no-referrer`, and `X-Content-Type-Options: nosniff`. Request JSON requires media type `application/json` (parameters allowed) and strict object schemas; malformed JSON is `400 INVALID_REQUEST`, wrong media type is `415 UNSUPPORTED_MEDIA_TYPE`, authentication is 401, authorization is 403, reservation/state conflicts are 409, frps unavailability is 503, and unhandled errors return a generic 500 after server-side logging.

Agent authorization errors also use version-1 JSON but currently omit these common API security/cache headers. That inconsistency is not a new Go contract.

### Route matrix

| Method and route | Successful contract |
| --- | --- |
| `GET /healthz` | public `{status:"ok"}` liveness; does not assert frps health |
| `POST /api/session` | public account login and session cookie |
| `GET /api/session` | authenticated account plus refreshed persistent session |
| `DELETE /api/session` | revoke session, clear cookie, 204 |
| `PUT /api/session/password` | local self-password change, revoke all sessions, clear cookie, 204 |
| `GET /api/events` | scoped session-authenticated invalidation SSE |
| `GET /api/state` | scoped overview; admin response also contains frps state and redacted deployment settings |
| `GET/POST /api/accounts` | admin list or local-account create (201) |
| `PATCH/DELETE /api/accounts/:id` | admin role change or empty local-account deletion |
| `PUT /api/accounts/:id/password` | admin local-account reset, 204 |
| `GET/POST /api/clients` | scoped list or create (201) |
| `GET/PATCH/DELETE /api/clients/:id` | detail, remark update, or cascade delete |
| `POST /api/clients/:id/rotate` | replace and return recoverable Client Token |
| `POST /api/clients/:id/restart` | send `restart_frpc`, 202; offline is 409 |
| `GET/POST /api/clients/:id/tunnels` | list or create (201) |
| `POST /api/clients/:id/tunnels/import/preview` | redacted typed TOML preview |
| `POST /api/clients/:id/tunnels/import` | atomic disabled batch, 201 |
| `PATCH/DELETE /api/tunnels/:id` | typed patch or delete |
| `POST /api/server/frp/start|stop|restart` | admin supervised frps action after activation result |
| `GET/PUT /api/server/frps/config/custom-404-page` | admin read/atomic update; UTF-8 content <=512 KiB |
| WebSocket `GET /api/agent` | Bearer-authenticated native agent channel |

Unknown management paths are 404; known paths reject unsupported methods explicitly. All `/api/*` except login and the separately Bearer-authenticated agent route require the session cookie. A stale cookie returns 401 rather than an unauthenticated session projection.

### Static React application

The shell is served only for `GET /`, `/clients`, `/clients/*`, `/accounts`, and `/server`. HTML is `no-store`; bundled non-HTML assets are `public, max-age=31536000, immutable`. Both receive `no-referrer` and `nosniff`; HTML uses a self-only CSP with the current inline-style allowance and same-origin connect policy. API/agent routes take precedence over the shell.

The active source must move to `web/tunnel-server/index.html` inside the single pnpm/Vite workspace. The retained routes, API field names, cache behavior, CSP capability, and deep-link fallback are exact inputs to `Prove the Vite MPA to Go embed path`; Bun HTML bundle internals and emitted filenames are not.

### Origin, Host, proxy, and request-size behavior

Mutations compare only the `Origin` URL's `host` (hostname plus port) with the request URL host. The scheme is deliberately ignored so an external HTTPS origin can reach an HTTP backend through a TLS reverse proxy. A missing Origin is allowed. There is no Host allowlist, trusted-proxy configuration, forwarded-header contract, CSRF token, or CORS grant.

The agent's default advertised FRP host comes from the authenticated request URL hostname, not `Forwarded` or `X-Forwarded-Host`; deployments with a distinct public FRP endpoint must set `YCY_TUNNEL_ADVERTISE_FRP_ADDR`.

No per-route JSON reader limit is enforced before buffering. Bun's inherited server request ceiling is 128 MiB; the 1 MiB import-source and 512 KiB custom-page checks happen only after the JSON body has been read. Login and ordinary mutations therefore admit far more buffered work than their schemas need. Because `net/http` has different defaults, the Go adapter must explicitly reproduce the observed ceilings and Host/Origin/proxy/cookie/body behavior or raise a narrow, evidence-backed compatibility exception.

## Browser SSE and retained UI behavior

`GET /api/events` disables the request timeout and returns `text/event-stream; charset=utf-8`, `X-Accel-Buffering: no`, and the common no-store/security headers. Each frame is only:

```text
data: {"version":1,"event":"changed|session_revoked"}

```

It sends `changed` immediately. Resource changes notify the owner and all administrators; server/frps/account changes notify administrators. Session revocation sends `session_revoked` and closes the stream. There is no event ID, retry field, heartbeat, subscriber limit, explicit write/backpressure policy, or slow-client eviction.

The React app treats SSE as invalidation, then reloads JSON. Browser state is otherwise ephemeral except `localStorage['ycy-tunnel-admin-theme']`. It refreshes the session every 24 hours, exposes disconnected-event state, redirects non-admins away from `/accounts` and `/server`, and implements operational account/client/tunnel/frps/custom-404 workflows. Token values remain intentionally revealable/copyable to authorized owners/admins; HTTP Basic Auth passwords remain write-only in the browser. There are no charts, retained logs, traffic polling, or endpoint health probes.

## Agent WebSocket protocol v3

### Upgrade and session ownership

The endpoint is `/api/agent`, using `Authorization: Bearer <Client Token>` on both the token probe and WebSocket upgrade. Only one pending or open socket is allowed per client. frps must be `running`; otherwise authorization is 503. Invalid tokens are 401, duplicate tokens are 409, and a non-upgrade request with an otherwise valid token is 426 after releasing its pending reservation.

Inbound WebSocket payload is capped at 1 MiB. The server pings every 30 seconds and closes a socket that did not pong by the next interval. There is no deadline for the required first `hello`: an authenticated socket that answers pings can remain in `awaiting_hello`, consume the Client Token's sole slot, and appear connected indefinitely. Sends ignore backpressure, and neither total connections nor outbound snapshot bytes are bounded.

### Exact messages

Every message is JSON with a tagged `type` and `tunnelProtocolVersion`. The current version is exactly integer `3`.

Client to server:

- `hello`: `ycyVersion`, Node-style `platform`, Node-style `architecture`, and non-negative `lastAppliedRevision`.
- `apply_result`: non-negative revision, success boolean, and optional structured error.
- `process_state`: `stopped|running|recovering|configuration_failed` plus optional structured error.

Server to client:

- `welcome`: required FRP version; archive name/URL/archive SHA/frpc SHA; advertised FRP host/port; Internal FRP Token; and complete snapshot.
- `desired_state`: complete replacement snapshot.
- `restart_frpc`: restart from the Applied Revision.
- `revoke`: reason `rotated|deleted`.
- `incompatible`: human-readable upgrade/unsupported-platform reason.

The server accepts hello only when protocol is 3, the reported platform/architecture resolves to the pinned manifest, and client Applied Revision does not exceed server Desired Revision. The client accepts welcome only when protocol, FRP version, archive, official URL, archive SHA, and frpc SHA exactly match its own compiled manifest. Extra JSON fields are currently tolerated; a Go decoder must not reject them during a v3 rolling rollout.

Close codes in active behavior include 4400 invalid/unexpected messages, 4401 revoked authentication, 4406 incompatibility, 4408 liveness timeout, 4409 duplicate token, and 4503 frps unavailable. A Go implementation must reproduce old-peer-visible close/authentication meaning even if internal libraries expose different APIs.

### Rolling migration constraint

The protocol is not capability-negotiated: either side rejects any version other than 3, and the client additionally pins every FRP artifact field. Therefore the first Go Tunnel release must implement v3 unchanged and prove both directions:

```text
legacy Bun tunnel connect -> Go tunnel server
Go tunnel connect         -> legacy Bun tunnel server
Go tunnel connect         -> Go tunnel server
```

Go runtime names must be mapped to legacy wire values (`darwin|linux|win32` and `x64|arm64`), not sent as raw `GOOS/GOARCH` (`windows`, `amd64`). Protocol v4 or artifact changes are outside the first-release port; protocol v3 remains unchanged.

## Client reconciliation and cached state

`last-applied.json` is pretty JSON plus newline containing `revision` and the complete last successful `FrpcDesiredConfiguration`: advertised endpoint, Internal FRP Token, and snapshot. `frpc.toml` is the active generated configuration. Both are rollback material and sensitive.

Reading cached state catches malformed/missing content and treats it as absent. Cached state never authorizes a cold child start. Every new process authenticates, validates welcome, receives current desired state, and creates the reconciler before frpc can start.

Apply is serialized and transactional at the process/file level:

1. Ignore a snapshot older than the current Applied Revision; ignore the same revision only when this process has already activated it.
2. Render a revision candidate and atomically write it mode 0600.
3. If any tunnel is enabled, run pinned `frpc verify -c <candidate>` with a ten-second timeout.
4. On verification failure, retain the previous child/file/revision and report `CONFIGURATION_FAILED`.
5. Stop the current child, publish `frpc.toml`, and start the candidate when needed.
6. On activation failure, stop the failed child, restore the previous file and previous enabled child where possible, retain Applied Revision, and report `ACTIVATION_FAILED`.
7. On success, persist the complete applied JSON, acknowledge the revision, and report current process state. An empty enabled set deliberately leaves no frpc child.

A valid change briefly interrupts every enabled tunnel for that client. Ordinary WebSocket reconnect with the same already-activated revision does not restart frpc. A process restart with the same cached revision still authenticates and reactivates it.

## Pinned FRP artifacts and installation

The required FRP version is exactly `0.70.1`. The manifest covers the six release targets and pins archive plus both extracted binary digests:

| Target | Official archive | Archive SHA-256 | frpc SHA-256 | frps SHA-256 |
| --- | --- | --- | --- | --- |
| macOS arm64 | `frp_0.70.1_darwin_arm64.tar.gz` | `cfa733b5a261c1647edee3c1fc4133d2542989b28f5602e81d47fc821d25c55f` | `dced7d6e9c35ecfd5a4625ddf3073660dd28e700387e7d838c64ef3cc1e4c1a9` | `5ec9a8d3a25c117b737c9318c3d52805f829a61d8942411bda2f5f11d990416f` |
| macOS x64 | `frp_0.70.1_darwin_amd64.tar.gz` | `cbf69cf26e5553e914e97d37f5d4367fa30f5f531d073a889465af4719281e25` | `32808dfdf91c4729f3c450d5a46afaa2fc293c8f6ee891743e3ca58685ba7a05` | `1bc014d4f52b687c7bb27344273b1ae504ca7a992021feed1e8445b67d981ef6` |
| Linux arm64 | `frp_0.70.1_linux_arm64.tar.gz` | `3990f396a9a490ee7f0e5f355287750ed41520064ed999eab443b5e9a78d773d` | `312be2787dc17c79b68ebf6cc9b536039b2fba595431782c68e3c056c1d491f8` | `1930b2cf4ccf7b37834f2c88279d89c2aff5a177ecc307f77c483dbfe1adeb4a` |
| Linux x64 | `frp_0.70.1_linux_amd64.tar.gz` | `333da23d1b9009d7c01638e9ba38cf4600f7d37d393f854e96ee1396adefa9a6` | `7d0270753bd171566a5389d2709fea29d2151f8fb4066ac20947e548e1da193a` | `ed1dfde60fd9f6b10237b5ab5953a6d791072c9a378ff9d77d6dfb5f370be482` |
| Windows arm64 | `frp_0.70.1_windows_arm64.zip` | `74d3acaf0f03ee190dd0462f9b49861dca50b0559c5488af4b36572fc951fcca` | `66c6f031d36bed993d0b54ee2f6f834b85d01d8f502c42f62308a4368f5e8936` | `29c7b664a6b2b12f0168c72bcca4c9ab19733ca58659cd944cd3b2555c4668df` |
| Windows x64 | `frp_0.70.1_windows_amd64.zip` | `531f3cd3cc41c0b4f077b54fe6b7dd83c0ff727e7f0bf412a4c78fa279165de5` | `1320325b3fd46d83ef7b2161d5e19f2b5dd9341b3391084a58d75ad82ef374d3` | `9df8a65fe693de28a8fa4baf7189c44a354a34b94c31f4254e18cc26dea3c57f` |

URLs are `https://github.com/fatedier/frp/releases/download/v0.70.1/<archive>`. Managed startup never searches `PATH` and accepts no custom binary override.

If the fixed target already has the expected binary digest, ycy runs `<binary> --version` with a five-second kill timer and requires a 0.70.1 report. Otherwise it downloads the complete official archive, follows redirects, verifies archive SHA, decompresses, finds both frpc/frps, verifies each extracted digest, and atomically publishes both fixed paths before version verification. Failure reports version, official URL, archive SHA, fixed manual-placement path, expected binary SHA, and cause.

The implementation has no cross-process installation lock, no response byte ceiling or independent download deadline, reads the full archive and target binary into memory, decompresses synchronously in memory, and publishes frpc/frps as separate file transactions. Concurrent supervisors can race. Go must retain official-source and digest guarantees while adopting selected locking, resource, timeout, cleanup, and permission behavior.

## Generated FRP configuration and import

Generated TOML text layout is not a compatibility contract; effective fields are. The server configuration contains bind address/port, HTTP vhost port, custom 404 path, token auth, one allowed port range, console logging, and no dashboard, metrics, admin API, plugins, or raw configuration escape hatch.

The client configuration contains the advertised endpoint, Internal FRP Token, `loginFailExit=false`, console logging, stable hidden FRP user `ycy_<client UUID>`, and one proxy named `t_<tunnel UUID>` per enabled definition. HTTP emits aliases plus an optional one-element locations array; TCP/UDP emit `remotePort`. Typed transport, health, Basic Auth, Host rewrite, and header-set options map to FRP v1 fields.

Managed server files are `<data-dir>/frps.toml` and optional `<data-dir>/404.html`. Non-empty custom 404 content is atomically published without restarting frps; empty content removes the file and restores FRP's built-in behavior. It affects FRP-generated no-route/backend failures, not a tunneled application's own 404.

The active Go project must replace the Bun manifest generator/build dependency. The later module/cutover roadmap must select a deterministic Go-owned or language-neutral generation/check path while retaining these pinned inputs and keeping Bun confined to `legacy/bun/`.

## frps/frpc supervision and recovery

One serialized supervisor owns zero or one child. It spawns `<binary> -c <config>`, captures stdout/stderr, and reports `stopped`, `running`, `recovering`, or `configuration_failed` plus at most one latest structured error.

- Client activation grace is 250 ms; server activation grace is 3 seconds. Exiting inside the grace fails Start/Restart instead of claiming `running`.
- An unexpected exit schedules `1, 2, 4, 8, 15, 30` second recovery. A 60-second stable run resets the failure count.
- Manual Stop clears retry/stability timers and suppresses recovery. A deterministic configuration failure remains stopped until new desired state or explicit action.
- Stop sends SIGTERM, waits five seconds, then sends SIGKILL and awaits exit. Start/restart/stop operations are serialized to prevent duplicate children.
- frps Start and Restart repeat binary resolution, TOML publication, verification, and activation. A failure remains visible while control HTTP continues.

The legacy tests use fake children for supervision. There is no native Windows proof that graceful signals, force-kill, descendants, console control events, lock PID checks, atomic replacement, or shutdown semantics match Unix. A Go `os.Process.Signal` call alone is insufficient on Windows and does not guarantee process-tree cleanup. Native platform ownership and ACL acceptance are hard release gates.

## Compatibility and risk matrix

| Area | Direct Go path | Incompatibility if changed |
| --- | --- | --- |
| CLI/config | Cobra compatibility adapter plus explicit strict Tunnel parsing | Different precedence, URL origins, TTY selection, or exit meaning can choose the wrong credential/instance |
| Remembered catalog | Standard PBKDF2, HMAC, AES-GCM, base64, JSON | Key or machine-ID drift makes secrets unreadable and changes instance directories |
| Session files | Standard SHA/HMAC/base64url/JSON and a Go-owned lock layout | Credential-domain or refresh drift logs out fresh-Go browser sessions |
| SQLite | Selected CGO-free SQLite driver against a fresh Go-owned schema v1/WAL | Schema, pragma, collation, transaction, or JSON drift can corrupt Go server truth |
| Local passwords | PHC parser plus `x/crypto/argon2` | New verifier/encoder drift makes every local account unusable |
| Browser protocol | `net/http` adapter with explicit routing/body/cache/origin behavior | Standard mux redirects/HEAD/body semantics or JSON omission drift breaks retained React |
| SSE | Explicit invalidation stream | Heartbeat/backpressure corrections must remain EventSource-compatible and account-scoped |
| Agent protocol | coder/websocket behind exact v3 adapter | Version/platform/artifact/header/close-code drift disconnects Go client/server peers |
| TOML | go-toml/v2 behind typed renderer/import adapter | Raw passthrough or semantic field drift bypasses reservations and FRP verification |
| FRP install | standard HTTP/archive/hash plus fixed manifest | PATH fallback, digest/version drift, or partial publication runs an unverified data plane |
| Processes | OS-specific supervisor implementation | Unix-only signals or child-only kill can leak frpc/frps on Windows |
| React | move current source to `web/tunnel-server` and Vite MPA | Route/cache/CSP/API drift produces a blank shell or stale management UI |

## Existing coverage and required Go tests

### Legacy evidence to consult

The 105-test suite covers:

- strict server configuration, client precedence, remembered encryption/selection/rotation, 32-entry concurrent retention, platform paths, and legacy-schema ignoring;
- domain normalization, reservations, lowest-port allocation, transactional import/revisions, token rotation, cascade deletion, and the fresh SQLite schema;
- ownership matrices, account lifecycle, Argon2 authentication, bounded persistent sessions, session revocation, projections, and scoped events;
- API auth/error flows, Basic Auth redaction, TOML import, agent snapshots/acknowledgements/revoke, frps actions/failures, proxy-origin handling, and managed 404 content;
- protocol authentication, artifact rejection, desired-state races, control-link outage, rotated-token rejection, reconciliation verify/rollback/empty-set/cold-start behavior;
- artifact table/rendering, archive rejection, manual diagnostics, download cancellation, TOML codec behavior, state/registry locks, cleanup, and supervisors;
- idle no-polling behavior and bounded latest runtime state; plus React tunnel-form mapping and validation.

The separate acceptance proves one host's real pinned frps/frpc forwarding. Tests are implementation evidence, not an active dual-runtime oracle: after archival, Go tests must be written command-by-command from `legacy/bun/` and fixtures shaped in each test. Active tests must not execute Bun or dispatch into legacy.

### Missing or weak coverage to add

Before either Tunnel leaf is complete, Go coverage must include:

1. A legacy-shaped `config.json` with real deterministic ciphertext/instance IDs, including machine-ID fallback and malformed salt/base64/ciphertext cases, while preserving unrelated config under concurrent writes.
2. A fresh-Go session directory proving login resume/refresh/revocation/eviction, environment and local credential revisions, restrictive permissions, and same-version lock exclusion. Do not add a Bun-written session fixture.
3. Go-created SQLite v1 database/WAL fixtures, including every table/check/index, plaintext secret fields, Argon2 PHC login, busy transactions, rollback, and normal restart behavior. Do not recognize or migrate a Bun-written database.
4. Byte/field-level browser API tests for every route, method, status, headers, cookie, null/omission distinction, reserved static route, deep link, reverse-proxy origin case, and the observed legacy Host/body/session/SSE behavior.
5. Protocol-frame tests between Go client and Go server, including platform name mapping, unknown fields, close codes, 1 MiB inbound boundary, large outbound snapshot policy, hello timeout, duplicate token, rotation, and control outage. No mixed Bun/Go direction is a gate.
6. Reconciliation crash points around candidate write, verify, stop, publish, start, rollback, state write, and acknowledgement; signals during apply and 30-second reconnect backoff; no cold cache activation.
7. FRP installation concurrency, response/time/decompression limits, truncated/duplicate/pathological archives, digest/version mismatch, pair publication recovery, offline manual placement, and fixed-path permissions.
8. Native macOS, Linux, and Windows amd64/arm64 artifact tests for paths, locks, SQLite/WAL, PBKDF2 machine identity, executable permissions/ACLs, process trees, graceful/forced shutdown, and real FRP HTTP/TCP/UDP forwarding. Cross-compilation alone is insufficient.
9. Vite production-graph tests proving the `tunnel-server` shell, shared hashed assets, CSP, cache headers, API/WebSocket/SSE namespace priority, and history-route fallback without committed `web/dist`.
10. Compatibility tests for the observed login concurrency, accounts/clients/tunnels/routes, import candidates, request bytes, snapshot bytes, WebSocket/SSE subscribers, backpressure, idle sockets, and shutdown behavior; new resource bounds are post-parity.

No CI, Docker, deployment, or release workflow changes belong to this planning/refactor phase. Final six-artifact verification remains a roadmap gate, not work performed by this ticket.

## Migration boundary and ordering implications

`tunnel server` and `tunnel connect` are separate migration leaves but share five foundations: `config.json` remembered-connection compatibility, fresh-Go state ownership, protocol v3, the pinned FRP manifest/installer, typed FRP rendering, and native supervision. The Go implementation order must therefore be:

1. Lock direct-read remembered configuration plus fresh-Go session/SQLite contracts and the observed operational behavior; defer corrected safety policies.
2. Implement and test shared v3 wire types, platform mapping, manifest, locks, file publication, and OS-specific child ownership.
3. Implement server persistence/domain transactions without the HTTP adapter; prove Go-created DB/account/token behavior.
4. Implement the Go server agent gateway and prove it with scripted Go protocol clients before registering `tunnel server`.
5. Implement the Go client/reconciler and prove real Go client to Go server forwarding before registering `tunnel connect`.
6. Move the retained React source to `web/tunnel-server`, serve it through the selected Vite/embed boundary, and prove the browser API/SSE workflows.
7. Run real pinned-FRP acceptance and native platform/artifact gates before either leaf is marked release-compatible.

The final command roadmap may refine prerequisites but may not add mixed-runtime rollout or Bun-state conversion. Legacy remains reference-only; no active Go code may import from it or invoke Bun.

## Compatibility checks surfaced

The first-release implementation must prove unchanged protocol-v3 operation from Go client to Go server, direct readability of remembered connections inside `config.json`, fresh-Go session/SQLite restart behavior, exact wire platform mapping, pinned FRP behavior, and native process behavior on required targets. A failed in-scope focused probe may open one narrow compatibility-exception ticket. Mixed runtimes and Bun-written non-config state are excluded; protocol evolution, HTTP/proxy/cookie hardening, new resource limits, fail-closed schema behavior, stronger secret storage, FRP install locking, and Windows ACL hardening are post-parity work.
