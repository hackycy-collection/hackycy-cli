# Treat Unified Diffs as analysis output

MCP Unified Diffs explain changes but are not guaranteed to apply through Git or `patch`. The Comparison Path is returned as structured metadata, while patch headers use fixed `baseline`, `target`, or `/dev/null` labels and omit timestamps, modes, indexes, and untrusted filenames; this avoids header injection and keeps the knowledge service separate from file mutation workflows.
