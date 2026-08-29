// Package gitprocess owns the process-level Git capability used by the command Factory.
package gitprocess

// Runner identifies a lazy Git process capability. Command-specific execution
// and result adaptation remain outside this type until their owning migration slice.
type Runner struct{}
