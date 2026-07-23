# Serve MCP over the diff HTTP server

The diff service exposes a Streamable HTTP MCP endpoint at `/mcp` on its existing HTTP server. The MCP and browser-facing adapters share the same in-process Comparison Workspace; a separate stdio process or an adapter that reconstructs results through the browser HTTP API would add lifecycle and synchronization complexity without adding a useful boundary.
