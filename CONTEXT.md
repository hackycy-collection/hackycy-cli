# Directory Comparison

The language used when comparing two directory trees and presenting their differences.

## Language

**Baseline Directory**:
The first directory in a comparison, representing the state changes are measured from.
_Avoid_: Left directory, old directory, source directory

**Target Directory**:
The second directory in a comparison, representing the state changes are measured to.
_Avoid_: Right directory, new directory, destination directory

**Comparison Status**:
The relationship of a Comparison Entry at the same Comparison Path between the Baseline Directory and Target Directory: Added, Deleted, Modified, or Unchanged.
_Avoid_: Left-only, right-only

**Added**:
A path that exists in the Target Directory but not in the Baseline Directory.

**Deleted**:
A path that exists in the Baseline Directory but not in the Target Directory.

**Modified**:
A Comparison Path that exists in both directories but whose entry kinds or comparable contents differ.

**Unchanged**:
A Comparison Path that exists in both directories with the same entry kind and equal comparable content.

**Comparison Snapshot**:
A stable view of the relationship between a Baseline Directory and Target Directory at one point in time. It remains unchanged until the user explicitly refreshes the comparison.
_Avoid_: Live comparison, real-time comparison

**Comparison Path**:
The relative path from a comparison directory root, used as the identity of an entry across both directories. Entries at different Comparison Paths are treated as Deleted and Added even when their content is identical.
_Avoid_: Absolute path, renamed file, moved file

**Comparison Entry**:
A regular file or Symbolic Link Entry inside the Comparison Scope. Directories organize Comparison Paths but do not receive their own Comparison Status.
_Avoid_: Directory diff, filesystem node

**Content Equality**:
Two files are equal only when their bytes are identical. File timestamps and other filesystem metadata do not affect equality.
_Avoid_: Timestamp equality, metadata equality

**Comparison Scope**:
The set of entries eligible for a Comparison Snapshot after ignore policies and hard exclusions are applied. The same Comparison Scope applies to both directories so filtering cannot create false one-sided statuses.
_Avoid_: Per-directory scope

**Symbolic Link Entry**:
A symbolic link included in the Comparison Scope and compared by its stored link target. It is never followed to the entry it references.
_Avoid_: Linked file, traversed link

**Binary Entry**:
A non-text file whose bytes participate in Content Equality but whose internal structure is not rendered as a textual diff. Browser-supported images may instead receive a visual comparison.
_Avoid_: Uncomparable file, ignored binary

**Image Entry**:
A browser-supported image rendered as a visual comparison while its Comparison Status remains byte-exact. Image presentation takes precedence over textual presentation for formats such as SVG.
_Avoid_: Pixel-equal image, image text file

**Text Entry**:
A file that can be deterministically decoded as UTF-8 or as BOM-marked UTF-16 for textual diff rendering. Its Comparison Status is still determined from the original bytes.
_Avoid_: Extension-based text file, guessed-encoding file

**Diff View Mode**:
A presentation rule for a textual diff, such as Exact or Ignore Whitespace. It never changes Content Equality, Comparison Status, or snapshot statistics.
_Avoid_: Comparison mode, status filter

**Change Set**:
All Added, Deleted, and Modified Comparison Paths in a Comparison Snapshot. Unchanged paths belong to the snapshot but not to its Change Set.
_Avoid_: All files, changed folder

**Comparison Issue**:
A Comparison Path that could not be assigned a reliable Comparison Status because it was unreadable, unsupported, or changed repeatedly while a snapshot was being built. Issues are reported separately from the Change Set.
_Avoid_: Modified file, ignored file

**Stale Entry**:
A Comparison Entry whose underlying filesystem state changed after its Comparison Snapshot was published. Its content is unavailable until the user refreshes the comparison.
_Avoid_: Live entry, automatically updated entry
