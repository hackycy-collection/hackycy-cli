# Put text differences behind Comparison Snapshots

`ComparisonSnapshot` owns on-demand Text Difference generation together with comparison truth and safe reads. The MCP adapter supplies only protocol validation and result mapping; computing patches in that adapter or rebuilding them through browser HTTP endpoints would split security, stale-read, encoding, and limit semantics across callers instead of keeping one deep workspace module.
