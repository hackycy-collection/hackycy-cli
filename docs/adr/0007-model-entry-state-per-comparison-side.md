# Model Entry State per comparison side

`ComparisonEntry` records optional Baseline and Target Entry States instead of one ambiguous kind and two detached sizes. This preserves file-to-symlink changes and one-sided presence in the comparison truth; the browser HTTP adapter may derive its legacy display-oriented `kind`, `baselineSize`, and `targetSize` fields, while MCP exposes the canonical per-side model directly.
