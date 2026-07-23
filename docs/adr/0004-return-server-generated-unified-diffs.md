# Return server-generated Unified Diffs

The MCP text-difference tool returns a server-generated Unified Diff plus structured metadata, rather than exposing renderer-specific hunks or invoking the browser's `@pierre/diffs` workers. Unified Diff is compact and broadly understood by AI clients; Added and Deleted text entries use `/dev/null` for the absent side, while Modified entries compare Baseline with Target.
