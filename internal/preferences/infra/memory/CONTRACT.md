# Preferences memory adapter contract

Revision: 1 · State: specified · Task: preferences-memory
Owner: coordinator · Parent: [../../CONTRACT.md](../../CONTRACT.md) revision 1 ·
Application: [../../app/CONTRACT.md](../../app/CONTRACT.md) revision 1 ·
Domain: [../../domain/CONTRACT.md](../../domain/CONTRACT.md) revision 1

## Purpose and consumer

The memory adapter is the first real implementation of the store ports the
application declares: the command store (conditional replace and remove) and
the query store (read and list). One instance satisfies both. It is executable
storage behavior for tests, demos and an explicitly selected ephemeral mode,
not a stand-in that answers with canned values. It also ships the reusable
logical scenario suite and the fault wrappers that a later persistent adapter
must rerun unchanged. Its first consumers are the application scenario tests
and the native slice's root, which will select it explicitly.

## Owned files and dependencies

Go: `internal/preferences/infra/memory/` (adapter and its test) and
`internal/preferences/infra/storetest/` (the reusable suite and fault
wrappers, imported only by tests). Rust: `infra/memory.rs`,
`infra/memory/CONTRACT.md` and the test-only `infra/contract.rs`. Allowed
imports: the standard library, the preferences domain, the application port
declarations and shared errors. Forbidden: transport, host, root, logger,
config, secret, clock, IDs, peer contexts, and any database or framework
crate. Nothing in production imports the suite.

## Public API

`memory.New()` / `Store::new()` returns an empty, independent instance that
implements both ports. There is no shared default instance and no seed.

The suite: Go `storetest.Run(t, open)` runs every scenario below as a subtest
against a fresh store from `open`; `storetest.Store` is the union of the two
ports. Rust `contract::m01_…` through `contract::m12_…` are functions generic
over `command::Store + query::Store + Sync` taking a fresh-store factory, so an
adapter's test module lists them one per `#[test]`.

Fault wrappers, both languages: `FailBefore(inner, failure)` refuses every
operation before touching the inner store; `LoseAcknowledgement(inner,
failure)` performs each write on the inner store and then reports the failure
instead of the result when the write changed state, while reads pass through.
They exist so application and transport tests can exercise dependency failure
and write uncertainty against real state.

## Observable contract

| Concern | Promise |
| --- | --- |
| Lifetime | A fresh instance per construction; state is shared by every port call on that instance and lasts only until it is dropped. No restart recovery is claimed |
| Ownership | Returned entries and lists are independent values; a caller cannot change stored state except through the ports. Go: reordering or overwriting a returned list leaves later lists unchanged. Rust: returned values are owned clones |
| Absence | Read reports absent distinctly from failure; list of an unknown scope is a successful empty list |
| Invariants | One entry per scope and key; revisions follow the domain decisions exactly; keys are unique per scope; scopes never interfere |
| Atomicity | Each conditional replace and remove checks the current entry and applies the change under one critical section; concurrent contenders see either the state before or the state after, never both succeeding |
| Failure | The adapter itself has no dependency to fail. Rust reports a poisoned lock as `internal / preferences.store_poisoned`. Injected failures come from the wrappers: `unavailable / preferences.store_unavailable` before any effect and `timeout / preferences.store_unacknowledged` after a committed write |
| Concurrency | Safe for concurrent use from any number of goroutines or threads; the Go race detector must stay silent |
| Durability | None beyond the instance. Root must select memory explicitly; it is never a silent fallback for a failed persistent adapter |
| Determinism | List order is bytewise key order within a scope, stable across calls; no clock, ID or randomness is consulted |

Go: read and list refuse a zero scope or key with the domain's `invalid_scope`
or `invalid_key` category; replace and remove reach the domain decision, which
validates before any state is inspected. Rust inputs are valid by construction.

## Acceptance scenarios

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| M01 | Fresh store; read `editor.theme` in the global and a workspace scope; list both | Absent without failure; both lists empty without failure |
| M02 | Replace `editor.theme` with `dark`, expected absent | Changed at revision 1 with a changed event; read finds an entry equal to the result; list holds exactly that entry |
| M03 | The same key again with expected absent, then with revision 1 and an equal value | First refused with `conflict / preferences.conflict`; second Unchanged with no event; read still shows revision 1 and `dark` |
| M04 | Replace at revision 2 and remove at revision 2 while current is 1 | Both refused with `conflict / preferences.conflict`; state unchanged |
| M05 | Replace with `light` at revision 1, then with integer 1 at revision 2 | Revision 2 then 3; read reflects each new value and revision; events carry the new value |
| M06 | Remove at the current revision; remove again with expected absent; create the key again | Removed at revision plus one with a removed event; read absent and list empty; second remove is the absent no-op; the new create starts at revision 1 |
| M07 | Create keys `z.last`, `m.mid`, `a-b`, `a.b`, `a_b`, `a0`, `ab` in that order; list twice | Both lists are `a-b`, `a.b`, `a0`, `a_b`, `ab`, `m.mid`, `z.last` |
| M08 | The same key in the global scope and two workspace scopes with different values; remove it from one workspace | Each read returns its own value; each list holds only its own entries; the removal leaves the other two scopes intact |
| M09 | Two instances; write into the first | The second reads absent and lists empty; a third fresh instance is likewise empty |
| M10 | Read the same key twice; take a list, overwrite its first element and reverse it, then list again | Both reads are equal; the second list equals the first as originally returned |
| M11 | Eight concurrent creates of one key with expected absent; then eight concurrent replaces at revision 1 with distinct values; then eight concurrent creates of distinct keys | Exactly one create succeeds and seven conflict; exactly one replace succeeds and the final entry is at revision 2 holding that winner's value; all eight distinct creates succeed and list holds eight entries |
| M12 | `FailBefore` around a fresh store: replace; `LoseAcknowledgement` around a fresh store: replace with expected absent, read, then replace again at revision 1 | FailBefore returns `unavailable / preferences.store_unavailable` and the inner store stays empty. LoseAcknowledgement returns `timeout / preferences.store_unacknowledged`, yet the read shows revision 1 committed, and the retry at revision 1 succeeds through the wrapper's inner store while again reporting the timeout |

## Verification

Go:

```sh
go test -count=1 -v ./internal/preferences/infra/...
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

This is real adapter evidence for memory only. Restart, corruption, migration
and IO failure evidence belongs to the persistent adapter, which must rerun
M01 to M12 and add its own.

## Exclusions and open decisions

No persistence, serialization, seed data, snapshot export, change
notification, transaction across keys, TTL or size limit is selected. The
Rust poisoned-lock path is documented, not exercised, because the adapter
offers no hook to panic inside its critical section.

## Completion evidence

See work/handoffs/preferences-memory.md.
