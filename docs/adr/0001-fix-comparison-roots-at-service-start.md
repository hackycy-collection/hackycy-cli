# Fix comparison roots at service start

The diff service fixes its Baseline Directory and Target Directory when it starts, and MCP clients may only inspect or refresh that Comparison Workspace. Allowing MCP calls to select arbitrary roots would make the service an arbitrary local-file comparison and reading capability, while also requiring a multi-workspace lifecycle; comparing another pair therefore requires another diff service instance.
