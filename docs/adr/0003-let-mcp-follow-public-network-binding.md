# Let MCP follow the public network binding

The `/mcp` endpoint follows the diff service's existing network binding: it is loopback-only by default and available without authentication on the local network when the operator starts `ycy diff` with `--public`. The flag is treated as explicit authorization to expose the complete read-only Comparison Workspace, rather than browser access alone; remote authenticated MCP access remains outside this service's scope.
