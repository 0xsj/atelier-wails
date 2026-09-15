# Workspace application contract

Revision: 1 · State: implemented · Task: workspace-app

The application layer owns consumer-facing operation names,
absence policy, and event publication. It does not own workspace lifecycle
rules or storage atomicity.

## Boundary

- `command.Store` is a context-aware conditional write port.
- `query.Store` is a context-aware read port.
- Store failures are returned with an operation prefix while preserving their
  error classification and cause through `errors.Is`/diagnostic helpers.
- `Publisher` observes returned events synchronously after the store commits;
  publication is not durable and cannot fail.
- `Discard` is the explicit no-subscriber publisher.
- Invalid query IDs/filters are refused before the store is called.

## Absence policy

| Operation | Absent workspace |
| --- | --- |
| Read | successful `found=false` |
| List | successful empty result |
| Rename, Archive, Restore | application `workspace.not_found` error |
| Forget | successful `found=false` no-op |

## Event policy

Register publishes `workspace.registered`. Changed Rename, Archive, Restore,
and successful Forget publish the event returned by the domain. Unchanged
results, absent no-ops, refusals, invalid input, and store failures publish
nothing.

## Shared application scenarios

| ID | Guarantee |
| --- | --- |
| WA01 | construction refuses a missing store or publisher |
| WA02 | register commits, publishes once, and is readable |
| WA03 | changed versus unchanged mutations control publication |
| WA04 | query reads/list use the shared store and preserve absence |
| WA05 | mutation absence follows the policy above |
| WA06 | domain refusals preserve state and publication history |
| WA07 | store failures preserve operation and classification |
| WA08 | publisher observes committed state |
| WA09 | canceled context reaches the store and publishes nothing |
| WA10 | invalid query inputs are refused before the store |

The Rust and Go suites use the same IDs and promises with language-native
signatures. Native transport, persistence, and runtime opening remain outside
this slice.

## Verification

```text
go test -count=1 ./internal/workspace/app/...
go test -race -count=1 ./internal/workspace/...
```
