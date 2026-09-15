# Preferences persistent adapter contract

Revision: 1 · State: specified · Task: preferences-persistence
Owner: coordinator · Memory: [../memory/CONTRACT.md](../memory/CONTRACT.md) revision 1 ·
File IO: shared fileio revision 1 · Decision:
[decisions/2026-09-12-preferences-file-persistence.md](../../../../decisions/2026-09-12-preferences-file-persistence.md)

## Purpose and consumer

The persistent adapter makes a completed preference write survive a process
restart. It implements the same command and query store ports as the memory
adapter and must pass the same M01–M12 suite, then adds what memory cannot
prove: reopen after restart, atomic commit with rollback on failure,
serialization round trips, corruption reported as corruption, and IO
failures reported honestly. Root selects it explicitly as the default store;
the memory adapter remains for tests and a deliberate ephemeral mode.

Storage is one JSON document per store, replaced atomically through the
shared fileio leaf. SQLite stays the candidate for multi-record metadata and
jobs; the decision record explains the choice.

## Owned files and dependencies

Go: `internal/preferences/infra/persistence/`. Rust: `infra/persistence.rs`
and `infra/persistence/`. Allowed imports: the standard library, the domain,
the application ports, shared errors, shared IDs and shared fileio, plus
serde in Rust. Forbidden: transport, host, root, logger, config, secret,
clock, peer contexts, the memory adapter and any database package.

## Public API

Go `Open(path string) (*Store, error)`; Rust `Store::open(path: &Path) ->
Result<Store, Failure>`. The store then offers the same read, list,
conditional replace and conditional remove operations as memory. `Open`
reads the whole document once; reads are served from memory; every change
rewrites the whole document atomically and only then updates memory. The
parent directory must exist; root ensures it.

## Document format

```json
{"format":1,"entries":[{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"revision":"1"}]}
```

Records are persistence-owned; they mirror the wire shapes today but are a
separate schema, so a wire change never silently changes storage. Entries are
sorted by scope then key, so identical state produces identical bytes.
64-bit numbers are decimal strings. Unknown keys, an unsupported `format`, a
malformed record, a record the domain refuses, or a duplicate scope and key
make the document corrupt. Migration from an older format is not selected;
`format` 1 is the first.

## Observable contract

| Kind / type | Meaning |
| --- | --- |
| `unavailable / preferences.storage_unavailable` | The document could not be read or written; the fileio failure is the private cause |
| `unavailable / preferences.storage_corrupt` | The document exists but is not a valid format 1 document; the private detail names the reason |
| `unavailable / preferences.storage_unsupported` | The document declares a `format` other than 1 |

Absence of the file at open is an empty store, not a failure; the file is
created by the first change. A refused or unchanged write creates nothing.
On a write failure the in-memory state and the file both keep the previous
committed state; the operation reports the failure and nothing else changed.
Open never returns an empty store for a corrupt or unreadable file. One
process owns a store's file; cross-process sharing is outside this revision.

## Acceptance scenarios

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| S01 | The shared suite over fresh temporary paths | M01–M12 pass unchanged |
| S02 | Write text, bool, int at int64 min and max, in the global scope and two workspace scopes; replace one at revision 1; reopen the same path | Every entry, value and revision reads back identically; a replace at the preserved revision succeeds after reopen |
| S03 | Open a missing path; read, list, refuse a conflicting write, make an unchanged write | Store is empty; no file exists until the first change; the refused and unchanged writes create no file |
| S04 | Open documents that are: not JSON; format 2; format 1 with an unknown key; a record with key `Bad`; revision `0`; a duplicate scope and key; an int of `x` | Not JSON, unknown key, bad key, bad revision, duplicate and bad int refuse `storage_corrupt`; format 2 refuses `storage_unsupported`; the file bytes are unchanged and no store is returned |
| S05 | Open in a directory, write once, remove write permission from the directory, write again; restore permission, write again; open a path whose directory does not exist and write | The second write refuses `storage_unavailable`; a read still shows the first write; reopening shows the first write; the third write succeeds; the missing-directory write refuses `storage_unavailable` |
| S06 | Two stores in different directories receive the same entries in different orders | The two files are byte-identical; each reopens to the same list |
| S07 | The failures of S04 and S05 | Kind unavailable with the stated type; the public message contains neither the path nor file content; the private detail or cause carries the reason |

## Verification

Go:

```sh
go test -count=1 -v ./internal/preferences/infra/...
go test -race -count=1 ./internal/preferences/... ./root
go vet ./...
gofmt -l internal root pkg
```

Rust:

```sh
cargo fmt --manifest-path src-tauri/Cargo.toml --check
cargo test --manifest-path src-tauri/Cargo.toml --locked --offline --lib
cargo check --manifest-path src-tauri/Cargo.toml --locked --offline
```

Restart is proven by reopening the same path in one test process; real
power loss is a property of the fileio mechanism and the filesystem.

## Exclusions and open decisions

No migration, backup, history, compaction, file watching, cross-process
locking, encryption or size limit. A user-facing recovery flow for a corrupt
document is not selected: root refuses to start with a logged failure.
Whether an explicit ephemeral mode is exposed through configuration is open.

## Completion evidence

See work/handoffs/preferences-persistence.md.
