# Choose persistent-data compatibility mechanisms

Type: grilling
Status: resolved
Blocked by: 01, 04, 07, 08, 09, 22, 23

## Question

For encrypted configuration, fs sessions, tunnel sessions, remembered tunnel connections, `tunnel.sqlite`, password hashes, updater state, and extracted runtime state, which formats can Go read unchanged and which require an automatic migration? Decide the adapter or migration mechanism, backup/rollback and failure detection, concurrency and locking behavior, cross-version direction, and tests for every exception to direct readability. No option may require manual user reconfiguration.

## Comments

- 2026-08-22, roadmap scope amendment: guaranteed Bun-written compatibility is narrowed to `~/.ycy-cli/config.json` only. Other legacy state is operator-managed and receives no migration mechanism or release gate.

## Answer

Closed as out of scope in its all-formats form. The first release directly reads and preserves Bun-written `~/.ycy-cli/config.json`, including Fork, CM, and remembered Tunnel connection fields. Only a reproducible failure of that document may stop integration for a narrow Wayfinder compatibility decision. Bun-written FS sessions, Tunnel sessions/SQLite/client cache/generated runtime state, Legacy Update State, and every other non-config format receive no compatibility adapter, migration, backup, fixture, or release gate; normal Go startup and diagnostic logging apply, and internal operators manage any residue.
