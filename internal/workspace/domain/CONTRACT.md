# Workspace domain contract

Revision: 1 · State: specified · Task: workspace-domain
Owner: coordinator · Parent: [../CONTRACT.md](../CONTRACT.md) revision 1

## Purpose and consumer

The domain module owns the Workspace vocabulary (identity, name, location,
status, revision, timestamps, expected revision, event) and the pure
lifecycle decisions: register, rename, archive, restore and forget. Its
consumers are the workspace application services and memory adapter, which
exist as unreviewed drafts and are adapted to this API as integration work.
The domain reads no clock, no filesystem and generates no ID: the application
supplies the ID and the UTC timestamp; the platform path adapter supplies a
canonical location string.

## Owned files and dependencies

Production and tests live in this directory. Allowed imports: the standard
library, shared errors and shared IDs. Forbidden: application,
infrastructure, transport, host, root, logger, config, secret, clock,
fileio and every peer context, including Preferences.

## Public API

Concrete signatures are language-local; promises and scenario IDs are shared.

A workspace carries a nonzero ID (Go: the zero ID is refused; Rust: the
shared `Id` has no nil value), a name, a location, a status (`active` or
`archived`), a positive revision and `created_at` and `updated_at`
timestamps. Fields are private; adapters rebuild a workspace only through
validated construction (Go `Rebuild`, Rust `Workspace::rebuild`), and
accessors expose each part.

Names are 1 to 120 Unicode scalar values, contain no control characters and
have no leading or trailing whitespace; interior spaces and non-ASCII are
fine. Locations are nonempty UTF-8 with no NUL that is absolute: a leading
`/`, a drive letter with `:` and a separator, or a leading `\\`. The domain
stores the string as supplied and never reinterprets platform syntax.

Timestamps are UTC millisecond instants. Go accepts any `time.Time` and
normalizes it; Rust accepts a `SystemTime` at or after the Unix epoch and
refuses earlier instants with `invalid_time`. `Expected` is absent-for-create
or an exact existing revision, as in Preferences, but is this context's own
type.

Decisions take the current workspace by value or reference and never mutate
it: `Register(id, name, location, at)`; `Rename(current, name, expected, at)`;
`Archive(current, expected, at)`; `Restore(current, expected, at)`;
`Forget(current, expected)`. Rename, archive and restore return Changed with
a new workspace and an event, or Unchanged with the current workspace and no
event. Forget returns the removal revision and an event. Absence is not a
domain concern: the store or application decides what an absent workspace
means for each operation.

Events are returned facts named `workspace.registered`, `workspace.renamed`,
`workspace.archived`, `workspace.restored` and `workspace.forgotten`, each
carrying the workspace ID, the new revision, the name, the location and the
status after the transition.

## Observable contract

Inputs are validated before the current workspace is inspected. Every
refusal is a shared failure with kind and diagnostic type; messages are fixed
and never contain the name or location.

| Kind / type | Meaning |
| --- | --- |
| `invalid / workspace.invalid_id` | Go zero ID |
| `invalid / workspace.invalid_name` | Name outside the rules |
| `invalid / workspace.invalid_location` | Location outside the rules |
| `invalid / workspace.invalid_time` | Rust instant before the Unix epoch |
| `invalid / workspace.invalid_revision` | Expected or restored revision of 0 |
| `invalid / workspace.invalid_status` | Restored status is neither active nor archived |
| `invalid / workspace.invalid_record` | Go current workspace is invalid |
| `conflict / workspace.conflict` | Expected presence or revision does not match the current workspace |
| `conflict / workspace.active_forget` | Forget on an active workspace; archive it first |
| `conflict / workspace.revision_exhausted` | Current revision is the maximum; no further change can be numbered |

## Acceptance scenarios

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| WD01 | Names `Atelier`, `My project`, `Atelier ✨`, 120 scalar values; empty, 121 scalar values, leading space, trailing space, a tab inside, a control character | Accepted names round-trip; the rest refuse `invalid_name`; no message contains the name |
| WD02 | Locations `/tmp/atelier`, `C:\Users\me`, `D:/work`, `\\server\share`; empty, `relative/path`, `./here`, `tmp`, a NUL inside | Accepted locations round-trip; the rest refuse `invalid_location` |
| WD03 | Register at an instant with sub-millisecond precision and a non-UTC zone (Go); at an instant before the epoch (Rust) | Timestamps read back as UTC millisecond instants; Rust refuses the pre-epoch instant with `invalid_time` |
| WD04 | Register with valid inputs; Go with the zero ID | Active workspace at revision 1, `created_at` equal to `updated_at`, `workspace.registered` carrying id, revision 1, name, location and `active`; the zero ID refuses `invalid_id` |
| WD05 | Rename to the same name at the matching revision; to a new name; to an invalid name; at a stale revision; with expected absent | Same name: Unchanged, same workspace, no event. New name: revision plus one, `updated_at` set to `at`, `created_at` unchanged, `workspace.renamed` with the new name. Invalid name refuses `invalid_name`; stale and absent refuse `conflict`; the current workspace is untouched throughout |
| WD06 | Archive an active workspace; archive it again; archive at a stale revision | Archived at revision plus one with `workspace.archived`; again: Unchanged, no event; stale: `conflict` |
| WD07 | Restore an archived workspace; restore an active one | Active at revision plus one with `workspace.restored`; active: Unchanged |
| WD08 | Forget an archived workspace at its revision; forget an active one; forget at a stale revision | Archived: revision plus one and `workspace.forgotten` with status `archived`; active: `active_forget`; stale: `conflict` |
| WD09 | Current at the maximum revision: rename to a new name, archive, restore from archived, forget from archived; rename to the same name | Changes refuse `revision_exhausted`; the same-name rename is still Unchanged |
| WD10 | Rebuild records: valid active and archived; revision 0; unknown status; invalid name; invalid location; Go zero ID | Valid records read back through accessors; each invalid one refuses its category |
| WD11 | Go: pass an invalid current workspace to rename, archive, restore and forget | Refused with `invalid_record` before any comparison |
| WD12 | Every refusal above | Carries a kind and type from the shared errors module; the message and fields never contain the name or location |

## Verification

Go:

```sh
go test -count=1 -v ./internal/workspace/domain
go test -race -count=1 ./internal/workspace/...
go vet ./...
gofmt -l internal/workspace
```

Rust:

```sh
cargo fmt --manifest-path src-tauri/Cargo.toml --check
cargo test --manifest-path src-tauri/Cargo.toml --locked --offline --lib domains::workspace::
cargo check --manifest-path src-tauri/Cargo.toml --locked --offline
```

Pure tests only. The application services and memory adapter beside this
module remain drafts adapted to compile; their own contracts follow.

## Exclusions and open decisions

No location uniqueness, existence check, path canonicalization, ordering,
open or close runtime action, filesystem inspection, persistence or event
dispatch belongs here. Whether absence on rename, archive and restore is a
refusal or a distinct outcome at the application boundary is decided by the
workspace application contract.

## Completion evidence

See work/handoffs/workspace-domain.md.
