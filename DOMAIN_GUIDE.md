# Backend domain guide

Working guide · 2026-09-11. This describes how to build a native domain in
Atelier. The current app is still Hello world; domain implementations are
reserved.
Memory adapters, persistent adapters and the tests described here are not yet
implemented. Each project owns its implementation and does not import siblings.

## What belongs to a domain

A bounded context owns a vocabulary, decisions and invariants. Its `domain/`
contains pure values and transitions; its `app/` owns use cases and required
effects. Its adapters translate those contracts into storage or delivery APIs.
A screen, table, native capability or directory does not by itself establish a
bounded context.

Start with one named behavior. Define its success, expected refusals, absence,
consistency and lifecycle before adding interfaces or choosing storage. A short
`CONTRACT.md` beside the context is enough to begin. Include concrete examples
and unresolved decisions; do not let a fake adapter accidentally choose them.

Use Preferences as the first complete native slice; its revision 1 boundary is
in [internal/preferences/CONTRACT.md](internal/preferences/CONTRACT.md).
Workspace and Jobs boundaries are also specified in
[DOMAIN_CONTRACTS.md](DOMAIN_CONTRACTS.md), but their implementations wait for
concrete workflows that consume registry or tracked-operation behavior.

## Build order within a slice

| Step | Deliverable | Evidence to obtain before advancing |
| --- | --- | --- |
| 1. Contract | Vocabulary, one use case, inputs/outcomes, invariants and unresolved questions | Concrete success, refusal, absence and failure scenarios |
| 2. Required leaves | Only the pure errors, IDs or value helpers this use case needs | Deterministic value/validation checks |
| 3. Pure domain | Valid construction and explicit state transitions | Invariants hold across valid and invalid transitions |
| 4. Application + ports | Operation and narrow capabilities it consumes | Required effects and failure paths are explicit |
| 5. Memory adapter | A stateful implementation of the real storage port | Adapter contract scenarios pass; state isolation is proven |
| 6. Application scenarios | Real operation over memory, controlled time/IDs | Successful changes, refusals without mutation, conflicts and dependency failures |
| 7. Desktop boundary | Plain DTO mapping and native facade, composed in root | Real frontend → native → application → memory → frontend round trip |
| 8. Persistent adapter | Context-owned records, codecs, migrations and recovery policy | Same logical adapter scenarios plus real storage/restart/concurrency checks |
| 9. Persistent default | Explicit root selection, matching UI save semantics | A completed write survives restart; corruption/failure are reported honestly |

Application and port design can iterate with the first memory adapter. The
sequence establishes dependencies, not a requirement to finish every abstraction
before learning from a consumer. Keep the memory implementation after persistence
arrives: it remains useful for focused tests and explicit ephemeral demos.

A memory-backed desktop slice is an intermediate milestone. It does not satisfy
a promise of durable save or the foundation's preference-restart milestone.

## Layout for one context

```text
internal/<context>/
├── CONTRACT.md                  # behavior and ownership, before implementation
├── domain/                      # values.go, transition.go; tests beside code
├── app/
│   ├── command/                 # operation + consumer-owned write/effect ports
│   └── query/                   # operation + consumer-owned read ports
├── infra/
│   ├── memory/                  # stateful implementation + contract tests
│   └── persistence/             # selected driver, records, migrations + tests
└── transport/desktop/           # plain DTOs and outcome mapping

internal/host/wails/              # native bound facades and Wails runtime APIs
root/                            # adapter selection and operation construction
root/workflows/                  # cross-context coordination if required
pkg/                            # reusable pure leaves and owned effectful packages
```

`memory/` is a proposed adapter directory alongside `persistence/`; add it when
implementing the first port. Keep provider-specific packages beneath persistence
once selected. These names do not select a database, driver or migration system.

## Dependencies and seams

- Domain code depends on the standard library and deliberately selected pure
  leaves. Supply time and generated IDs as values to pure transitions.
- Application code depends on its own domain and consumer-owned ports. Describe
  the capability needed by the use case rather than mirror an entire database API.
- Infrastructure implements ports. Transport translates inputs/outcomes; native
  facades own framework calls. Neither bypasses application operations for storage.
- Root constructs adapters and operations, injects capabilities, and owns their
  lifetime. Nothing below root imports root.
- Peer contexts do not import one another. Root provides a translation adapter
  for a required capability or coordinates an explicit workflow through APIs.
- Shared leaves do not import contexts. Effectful logger/config/file IO packages
  are for adapters and root, not for pure rules.

Do not start with a universal Repository<T>, generic command bus, service locator
or event bus. Name an operation such as “load preference override” or “replace
at expected revision” when that is the actual contract. Commands and queries can
use one store; they distinguish responsibilities, not deployment topology.

### Go details

Use ordinary values and typed errors with `errors.Is`/`errors.As` where appropriate.
Declare interfaces with their consuming command/query package; one concrete memory
store may satisfy several small interfaces. Do not export a native framework
context or database type through those interfaces. `context.Context` can carry
cancellation through effectful application calls; it is not a dependency of pure
value validation.

A memory map requires synchronization when shared across concurrent calls.
Protect an entire compare-and-write operation, not just individual map accesses.
Copy slices/maps and nested mutable values at storage boundaries. Avoid package
singletons and adapters constructed inside commands. A contract test package can
construct both memory and persistent adapters without making either import the
other. Keep test scaffolding out of production dependency paths.

Go's `internal/` restriction does not prohibit imports between sibling contexts;
`tools/architecture/check_imports.py` refuses them.


## Memory adapter contract

A memory adapter is executable storage behavior, not a list of canned answers.
It implements the same logical port the persistent adapter will implement.

| Concern | Required behavior |
| --- | --- |
| Lifetime | Construct a fresh instance per isolated test or demo root. Share one instance across operations belonging to that running application. |
| Ownership | A caller cannot mutate stored state by changing an input after writing or a returned value after reading. Copy mutable data or expose immutable values. |
| Absence | Return the declared absent outcome. Missing is not an empty record, a backend failure or an unknown operation. |
| Invariants | Respect port guarantees for uniqueness, ordering, pagination and revision checks when the contract promises them. Domain rules still belong to domain/application. |
| Atomicity | A conditional write checks and changes state as one operation. A multi-write promise needs an explicit transaction capability. |
| Failure | Support deterministic dependency failures through a wrapper or dedicated test adapter. Do not smuggle success through a no-op implementation. |
| Concurrency | Synchronize shared mutable storage if concurrent access is admitted. A lock around individual calls does not make a read-then-write workflow atomic. |
| Durability | State lasts only for the declared in-process lifetime. Do not claim restart recovery. |
| Determinism | Stable ordering and controllable generated values; no random delays or wall-clock sleeps in behavior tests. |

Native memory adapters run behind the real native facade and service path.
Frontend preview adapters run without the native host and are a separate boundary.
A successful browser preview does not prove native decoding or application rules.
Prefer exercising the native memory path early in the desktop window.

Root chooses memory explicitly. A persistent adapter that fails to initialize
must not silently switch to empty memory and make existing user data appear lost.
Expose an initialization failure or a deliberately selected ephemeral mode.

## Fixtures, seeds and failure scenarios

Keep fixture builders close to the context's tests. Build domain-valid values
through public constructors. Use deliberately malformed wire/storage records
only in adapter decoding tests and label them as such.

Separate three things:

- **Test fixtures:** small named scenarios with fixed IDs/time, fresh per test.
- **Demo seeds:** optional data installed by an explicit demo root or startup
  option. Never seed production implicitly when a store looks empty.
- **Fault adapters:** deterministic wrappers returning an intended error before
  an effect or simulating a lost acknowledgement after a committed effect.

A seed procedure must define whether rerunning is refused, idempotent or replaces
an explicitly disposable store. Tests must not depend on another test's seed.
A “lost response” scenario should really change the underlying state before
reporting uncertainty; otherwise it cannot exercise reconciliation behavior.

## Adapter scenarios and persistence second

Write a reusable logical contract suite that accepts an adapter factory and
fresh-store cleanup. Run it against memory first, then the real persistent
adapter. Share promises and scenario names across Go and Rust, not source code
or exact API signatures.

Select scenarios relevant to the port: absent read; successful write/read;
replacement; duplicate key; revision conflict; deterministic ordering; independent
instances; input/output mutation isolation; and atomic failure where promised.
Do not invent operations solely to fill this list.

The persistent adapter also needs evidence memory cannot supply: reopen after
process restart, real concurrent access, atomic commit/rollback, serialization
round trips, migration from supported formats, corrupt records, and relevant IO
failures. State which guarantees the storage engine actually provides. A shared
suite alone does not prove crash durability or compatibility.

Persistence owns its schema and mapping. Restore domain values through validated
construction; invalid records are corruption, not normal absence. Decide schema
versioning and migration behavior before existing data depends on it.

Avoid using a global lock or transaction manager across contexts by default.
An operation that promises both state and an event needs an atomic contract for
both; an event callback after saving is not durable delivery. Cross-context
workflows must state partial failure and recovery explicitly.

## Errors, cancellation and desktop data

Pure domain refusals carry meaning, not HTTP status or UI copy. Application code
adds use-case context without flattening conflicts, absence and IO failures into
one string. Desktop mapping chooses a stable public failure shape and retains
appropriate diagnostics privately. Do not serialize arbitrary native errors,
stack traces or unrestricted cause chains.

Distinguish an operation's declared refusals from infrastructure failures it
cannot rule out. The frontend Result/Failure vocabulary maps those meanings;
it does not replace native domain error types. Every native payload is plain
data, and frontend codecs validate it before constructing Results.

Cancellation is cooperative. A canceled caller or lost acknowledgement does not
prove that a write failed. Retrying a write requires a defined idempotency or
reconciliation contract. Dialog dismissal can be a normal outcome. Retryable
failure classification alone does not authorize automatic retries.

## Completing a slice

Record the actual implementation in README/status notes and consequential
commitments in decisions. A completed memory slice has tested pure rules,
application behavior, memory semantics and a working native boundary. Mark its
volatile lifetime explicitly. A completed persistent slice additionally has
restart and real adapter evidence and an explicit default selected in root.

Keep tests focused on observable contracts, including negative cases that could
otherwise look like success. Keep pure tests close to owners and real adapter
checks close to adapters. `tools/architecture/check_imports.py` refuses forbidden imports; folder
layout alone enforces neither purity nor bounded contexts, and the checker
does not verify behavior.

This guide was added as documentation only. No adapters or domain behavior were
implemented or runtime-tested as part of writing it.
