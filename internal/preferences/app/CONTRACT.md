# Preferences application contract

Revision: 1 · State: specified · Task: preferences-app
Owner: coordinator · Parent: [../CONTRACT.md](../CONTRACT.md) revision 1 ·
Domain: [../domain/CONTRACT.md](../domain/CONTRACT.md) revision 1

## Purpose and consumer

The application owns the Preferences use cases: replace and remove a
preference at an expected revision, read one entry, list a scope, and resolve
a value against a caller-supplied fallback. It declares the narrow capabilities
those use cases consume and publishes returned event facts after the store
commits. Its first consumers are the memory adapter (preferences-memory, which
must satisfy the ports declared here) and the desktop transport
(desktop-wire-contract). The drafts beside this module are replaced by this
revision.

## Owned files and dependencies

Go: `internal/preferences/app/command/` and `internal/preferences/app/query/`
own their services and the ports they consume; the former `app/port` package is
removed. Rust: `app/command.rs`, `app/query.rs`, `app/tests.rs`; `app/port.rs`
is removed. Allowed imports: the standard library, the preferences domain and
shared errors. Forbidden: infrastructure, transport, host, root, logger,
config, secret, clock, IDs and every peer context. Tests may construct the
memory adapter; production code never does.

## Public API

Ports are consumer-owned and declared beside the service that uses them. One
concrete adapter may satisfy several.

Command store: conditional `Replace(scope, key, value, expected)` returning the
domain replace result, and conditional `Remove(scope, key, expected)` returning
the domain remove result. The adapter applies the domain decision and its
state change as one atomic operation and returns domain refusals unchanged.

Publisher: `Publish(event)` receives each returned event once, synchronously,
after the store call has returned success. Publication cannot fail and is not
durable delivery; a subscriber that needs durability is a later outbox
concern. `Discard` is provided for roots without subscribers.

Query store: `Read(scope, key)` returning found-or-absent, and `List(scope)`
returning entries sorted by key.

Command service: constructed from a store and a publisher; `Replace` and
`Remove` with the domain's argument types. Query service: constructed from a
store; `Read`, `List` and `Resolve(scope, key, fallback)`. Resolve returns the
stored value with its revision when present, otherwise the fallback marked as
not stored. It never writes.

Go signatures take a leading `context.Context` on every port and service
method and return `(result, error)`; construction returns an error for a nil
dependency. Rust services are generic over their ports, take no context, and
return `Result<_, Context<Failure>>`; a missing dependency is unrepresentable.

## Observable contract

Go services validate scope, key, value, expected and fallback before any port
call, refusing with the domain's `invalid_*` categories. Rust inputs are valid
by construction.

Every failure leaving a service is annotated with the operation name
(`preferences.replace`, `preferences.remove`, `preferences.read`,
`preferences.list`, `preferences.resolve`) and preserves the underlying
classification: kind, diagnostic type, condition identity (Go `errors.Is`) and
source. A domain refusal, a classified store failure and an unclassified store
error each pass through with their meaning intact; none becomes absence,
success or a different kind. Go additionally preserves `context.Canceled` and
`context.DeadlineExceeded` from the store.

Events are published only after a store call returns Changed or Removed, and
exactly once per such call. Unchanged, absent, refused and failed calls
publish nothing. At publication time the store already holds the committed
state.

Workspace scopes are accepted by every operation; whether the first UI offers
them is a transport decision. Nothing here checks that a workspace exists.

| Kind / type | Meaning |
| --- | --- |
| `internal / preferences.missing_dependency` | Go construction with a nil store or publisher |
| domain `invalid_*` | Go zero or invalid inputs, refused before the port |
| passthrough | Domain refusals and store failures keep their own kind and type |

## Acceptance scenarios

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| A01 | Go: construct a command service with a nil store or nil publisher; a query service with a nil store | Refused with `internal / preferences.missing_dependency` |
| A02 | Fresh memory store, recording publisher; replace `editor.theme` with `dark`, expected absent | Changed at revision 1; the publisher received exactly one `preference.changed` event with revision 1; a query read finds the entry at revision 1 |
| A03 | Replace again with the equal value at expected revision 1 | Unchanged; nothing published; revision still 1 |
| A04 | Replace with `light` at expected absent, then at revision 9 | Each refused with `conflict / preferences.conflict`; nothing published; the read still shows `dark` at revision 1 |
| A05 | Go: replace or remove with a zero scope, zero key, zero value or invalid expected; read, list or resolve with zero inputs | Refused with the matching `invalid_*` type; the store records no call |
| A06 | Remove at expected revision 1; remove the same key again with expected absent | First: Removed at revision 2, one `preference.removed` event published, the read is absent. Second: absent outcome, nothing published, no failure |
| A07 | A fault store that fails every call with `unavailable / preferences.store_unavailable` before any effect | Replace, remove, read, list and resolve return kind unavailable with that type; nothing published; no absence or success is reported. Go: a store returning an unclassified error yields a non-nil unclassified error |
| A08 | A publisher that reads the store from inside `Publish` during a replace | The publisher observes the new entry already committed |
| A09 | Read a present key, an absent key; list a scope holding `z.last` and `a.first`; list an empty scope | Found and absent are distinct; the list is ordered `a.first`, `z.last`; the empty scope lists zero entries without failure |
| A10 | Resolve a present key with a fallback; resolve an absent key with fallback `true` | Present: stored value, marked stored, with its revision. Absent: fallback, marked not stored; a following read is still absent and the scope list is unchanged |
| A11 | Go: call every operation with an already-canceled context against a store that honors cancellation | Each returns an error classified `canceled`; nothing published |
| A12 | Every failure from A04, A05 and A07 | Reports the operation name; keeps the original kind and type; Go `errors.Is` matches the original store failure; Rust `Context::inner` returns it; the message never contains the key or value text |

## Verification

Go:

```sh
go test -count=1 -v ./internal/preferences/app/...
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

Application scenarios run over the memory adapter draft with a recording
publisher and a fault wrapper. That exercises real operations, not the
adapter's own contract, which preferences-memory specifies separately.

## Exclusions and open decisions

No retry, batching, transaction across keys, default registry, schema, event
bus, durable outbox, provenance scope, logging or transport mapping is
selected. The publisher is an in-process notification seam only. Whether the
root wires a real subscriber or `Discard` is decided in the native slice.

## Completion evidence

See work/handoffs/preferences-app.md.
