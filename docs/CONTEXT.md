# Directory Comparison

Directory Comparison describes a directional comparison between two directory trees while keeping all observable results inside an explicitly authorized filesystem scope.

## Language

**Baseline Directory**:
The directory tree representing the older side of a comparison. Entries found only here are Deleted.
_Avoid_: Left directory, source directory, old directory

**Target Directory**:
The directory tree representing the newer side of a comparison. Entries found only here are Added.
_Avoid_: Right directory, destination directory, new directory

**Comparison Workspace**:
The authorized pairing of one Baseline Directory and one Target Directory. The pair is fixed for the lifetime of the comparison and defines its complete filesystem scope.
_Avoid_: Comparison session, diff session

**Public Comparison**:
A Comparison Workspace that its operator has explicitly made available to clients on the local network. Public access covers every read-only interface to that workspace, including browser and agent access.
_Avoid_: Public UI, browser sharing

**Comparison Snapshot**:
An immutable, published view of every included result and Comparison Issue in a Comparison Workspace at one point in time. It remains authoritative until a complete replacement is published.
_Avoid_: Live diff, current files

**Refresh**:
An explicit request to build and publish a replacement Comparison Snapshot from the same Comparison Workspace. It does not imply file watching, change subscriptions, or continuous synchronization.
_Avoid_: Rescan, reload, live update

**Comparison Entry**:
The result for one exact Comparison Path in a Comparison Snapshot, describing the relationship between its Baseline and Target Entry States as Added, Deleted, Modified, or Unchanged.
_Avoid_: File result, diff item

**Entry State**:
The recorded presence, kind, and relevant metadata of a Comparison Entry on either the Baseline or Target side. An absent Baseline State means Added; an absent Target State means Deleted.
_Avoid_: Entry kind, side details

**Comparison Path**:
An exact, case-sensitive, POSIX-style path relative to both roots of a Comparison Workspace. It identifies comparison scope but never grants filesystem read access by itself.
_Avoid_: Absolute path, local path, file locator

**Entry ID**:
An opaque reference to one Comparison Entry within one Comparison Snapshot. It has no meaning outside that snapshot and cannot be interpreted as a filesystem path.
_Avoid_: File ID, path ID

**Comparison Issue**:
A scoped statement that the service could not determine a reliable comparison result for a path or comparison rule. It is reported alongside, but is not itself, a Comparison Entry or comparison status.
_Avoid_: Error entry, failed file, issue status

**Text Difference**:
A line-oriented explanation of the textual change between safely decoded Baseline and Target Entry States. A Modified Comparison Entry may have no Text Difference when only its encoding or byte-order mark changed.
_Avoid_: Byte diff, rendered diff, file contents
