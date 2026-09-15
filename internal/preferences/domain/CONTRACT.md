# Preferences domain contract

Revision: 1 · State: specified · Task: preferences-domain
Owner: coordinator · Parent: [../CONTRACT.md](../CONTRACT.md) revision 1

## Purpose and consumer

The domain module owns the Preferences vocabulary (scope, key, value, expected
revision, entry, event) and the two pure compare-and-replace decisions. Its first
consumers are the preferences application services and memory adapter, which are
scheduled after this task and currently exist only as unreviewed drafts. The
domain performs no storage, clock, ID generation, logging or host call: a caller
supplies the current entry, and the domain returns the decided values.

## Owned files and dependencies

Production and tests live in this directory. Allowed imports: the standard
library, shared errors and shared IDs. Forbidden: application, infrastructure,
transport, host, root, logger, config, secret, clock and every peer context.
Prerequisites: errors revision 2, IDs revision 1.

## Public API

Concrete signatures are language-local; promises and scenario IDs are shared.

Scope is `global` or `workspace(id)`. Go: `Global()`, `ForWorkspace(id.ID) (Scope,
error)`; a zero ID and the zero `Scope{}` are invalid. Rust: an enum over the
shared `Id`, which has no nil value, so every constructible scope is valid.
Display is `global` or `workspace:<canonical uuid>`. Scopes are comparable.

Key matches `[a-z][a-z0-9_.-]{0,127}`, byte-wise, case-sensitive, never trimmed.
Construction is the only way to obtain a key: Go `NewKey(string) (Key, error)`
with a private field so a cast cannot bypass validation; Rust `Key::new(&str)`
owning a `String`. Keys are comparable, hashable and ordered by bytes.

Value is text (valid UTF-8, at most 4096 bytes, empty allowed), boolean or
signed 64-bit integer. Constructors: Go `Text(string) (Value, error)`,
`Bool(bool)`, `Int(int64)`; Rust `Value::text(impl Into<String>) -> Result`,
`Value::boolean`, `Value::integer`. A kind accessor and per-kind accessors
return the payload only for the matching kind. Equality compares kind and
payload; text `"1"` and integer `1` are different values.

Expected is absent-for-create or an exact existing revision (at least 1). Go
`ExpectAbsent()`, `ExpectRevision(uint64) (Expected, error)`; the zero value is
absent. Rust `Expected::absent()`, `Expected::revision(u64) -> Result`.

Entry carries scope, key, value and a positive revision behind private fields.
Adapters rebuild entries only through validated construction: Go
`Restore(scope, key, value, revision) (Entry, error)`; Rust `Entry::restore`.
Accessors expose each part. Go additionally exposes `Valid()` on scope, key,
value, expected and entry because zero values are constructible there.

Events are returned facts named `preference.changed` (with the new value and
revision) and `preference.removed` (revision only). The domain never calls a
dispatcher.

Decisions: `DecideReplace(current, scope, key, value, expected)` returns
Changed(entry, event) or Unchanged(entry); `DecideRemove(current, scope, key,
expected)` returns Removed(revision, event) or Absent. Rust spells these
`decide_replace` / `decide_remove` with outcome enums. `current` is the caller's
view of the stored entry for exactly that scope and key, or none.

## Observable contract

All inputs are validated before the current entry is inspected. A refusal
returns a shared failure with kind and diagnostic type and no other effect; the
domain never mutates the supplied current entry or returns aliases of it.

| Kind / type | Meaning |
| --- | --- |
| `invalid / preferences.invalid_scope` | Go zero or zero-workspace scope |
| `invalid / preferences.invalid_key` | Key text outside the grammar |
| `invalid / preferences.invalid_value` | Text over 4096 bytes or not UTF-8, or a Go zero value |
| `invalid / preferences.invalid_revision` | Expected or restored revision of 0 |
| `invalid / preferences.invalid_entry` | Current entry does not belong to the requested scope and key, or is invalid |
| `conflict / preferences.conflict` | Expected presence or revision does not match the current entry |
| `conflict / preferences.revision_exhausted` | Current revision is the maximum; a further change or removal cannot be numbered. Re-reading and retrying does not clear it |

Messages are fixed strings. No failure message, field or detail contains the
supplied key or value text. Revision exhaustion is classified as a conflict
because it is a refusal by state, not by input or infrastructure; the type
distinguishes it so callers do not loop on re-read and retry.

Ordering, atomicity, concurrency, cancellation and durability are not domain
concerns; the store port and adapters own them.

## Acceptance scenarios

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| P01 | Construct global and workspace scopes; compare; display | Global displays `global`; workspace displays `workspace:<uuid>`; equal inputs are equal, different workspaces are not. Go: zero ID refused with `invalid_scope`; `Scope{}` reports invalid and displays `invalid` |
| P02 | Key texts `a`, `editor.theme`, `a-b_c.d9`, a 128-character key | Accepted and round-trip unchanged. Empty, 129 characters, `A`, `1a`, leading or trailing space, `a/b`, `é` and `Editor.Theme` refused with `invalid_key`; the failure message excludes the input |
| P03 | Values: empty text, 4096-byte text, 4097-byte text, invalid UTF-8 (Go), true, min and max int64 | Sizes within limits accepted; 4097 bytes and invalid UTF-8 refused with `invalid_value`. Accessors return the payload only for the matching kind. Text `"1"`, integer 1 and boolean true are pairwise unequal |
| P04 | Expected absent; revision 1; revision 0 | Absent reports absent; 1 reports revision 1; 0 refused with `invalid_revision` |
| P05 | Restore an entry with revision 5; with revision 0 | Accessors return each part at revision 5; revision 0 refused with `invalid_revision`. Go: zero scope, key or value refused with their own categories |
| P06 | Replace with no current entry and expected absent; then with expected revision 1 | Changed entry at revision 1 carrying the value; `preference.changed` event with scope, key, value and revision 1. Expected revision on an absent entry refused with `conflict`, no result |
| P07 | Current at revision 3; replace with expected absent, revision 2 and revision 4 | Each refused with `conflict`; no entry or event returned |
| P08 | Current at revision 3 holding `dark`; replace with `dark` at expected revision 3 | Unchanged with the current entry, revision still 3, no event |
| P09 | Current at revision 3 holding text `dark`; replace with `light`, then with integer 1, each at the matching revision | Changed at revision 4 (and 5) carrying the new value; `preference.changed` carries the new value and new revision; the caller's current entry is untouched |
| P10 | Current at the maximum revision; replace with a different value, replace with an equal value, remove, each at that revision | Different value and remove refused with `revision_exhausted`; equal value still returns Unchanged |
| P11 | Remove with no current entry, with expected absent and with expected revision 7 | Absent outcome, no event, no failure |
| P12 | Current at revision 3; remove with expected absent, revision 2, then revision 3 | Absent and revision 2 refused with `conflict`; revision 3 returns Removed with revision 4 and a `preference.removed` event carrying scope, key and revision 4 and no value |
| P13 | Current entry for another key, or for another scope, passed as current | Replace and remove refused with `invalid_entry`. Go: an invalid current entry is refused the same way |
| P14 | Invalid key together with a current entry that would conflict | Refused with `invalid_key`: input validation precedes state comparison. Rust: the same holds for a zero expected revision constructed before the call |
| P15 | Every refusal above | Carries a kind and diagnostic type from the shared errors module; the message and fields never contain the key or value text |

## Verification

From the repository root, Go:

```sh
go test -count=1 -v ./internal/preferences/domain
go test -race -count=1 ./internal/preferences/...
go vet ./...
gofmt -l internal/preferences
```

Rust:

```sh
cargo fmt --manifest-path src-tauri/Cargo.toml --check
cargo test --manifest-path src-tauri/Cargo.toml --locked --offline --lib domains::preferences::
cargo check --manifest-path src-tauri/Cargo.toml --locked --offline
```

These are pure tests. No adapter, native bridge or rendered UI evidence exists
for Preferences yet; a passing domain suite does not claim any.

## Exclusions and open decisions

No defaults, schema, key registry, list ordering, workspace existence check,
event dispatch, serialization, history or persistence codec belongs here. The
memory adapter and application services drafted beside this module are not
covered by this revision; their contracts follow in preferences-app and
preferences-memory, which must adapt to this API rather than the reverse.

## Completion evidence

See work/handoffs/preferences-domain.md for files, scenario-to-test mapping,
commands and coordinator review.
