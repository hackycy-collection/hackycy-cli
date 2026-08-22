# Choose safe and bounded contracts for the FS service

Type: grilling
Status: resolved
Blocked by: 07

## Question

Before the Go FS server is approved, which corrected contracts replace the legacy service's unsafe, unbounded, or defective behavior? Decide the default and compatibility story for active HTML/XHTML; Host, Origin, bind-address, and unauthenticated read policy; opened-handle Browse Root containment across symlinks, reparse points, ancestor replacement, and non-UTF-8 names; remote-download URL, redirect, DNS, dial-address, and proxy validation; body, directory, task, byte, disk, duration, session, login, SSE, and shutdown bounds; and the exact nonzero startup failures and working chunk-size option semantics. Coordinate persistent session locking and permissions with `Choose persistent-data compatibility mechanisms`, and verified 7-Zip state with `Inventory upgrade and release-artifact compatibility contracts`, without duplicating those decisions. Record exact CLI/HTTP behavior, error codes, limits, platform rules, observability, and adversarial tests while preserving ordinary loopback/LAN file-browser workflows and the retained React client.

## Comments

## Answer

Closed as out of scope for the parity-first Go release. FS retains the CLI, HTTP/SSE, authentication, filesystem, upload, download, archive, session, and React-facing behavior recorded in its inventory. New active-content, trust, DNS, path-race, resource-limit, startup-status, and chunk-size policies are deferred hardening. Proven capability gaps such as a CGO-free thumbnail implementation may still receive focused research or a narrow compatibility decision when the FS command is migrated.
