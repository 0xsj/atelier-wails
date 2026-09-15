# Foundation build order

Current backend status: errors, clocks, secrets, config, IDs, provenance, logger and file IO are implemented, and the layer import rules are enforced by tools/architecture. The first
combined gate is complete. [Integration evidence](work/handoffs/native-leaves-integrate.md).
Console logging is wired in native root; Preferences, Workspace and Jobs
contracts are now specified.

The directory skeleton exists upfront; implementations arrive from leaves into
complete slices. There are no placeholder success responses or empty exported APIs.
Svelte 5 + TypeScript + Vite is selected for both independent projects. SvelteKit,
query caching and a UI primitive vendor are not installed by this scaffold.

| Order | Library/module | First contract to implement |
| --- | --- | --- |
| 1 | kernel/result.ts | Ok/Err, exact constructors, map/andThen/match; expected failure stays a value |
| 1 | kernel/failure.ts | Closed failure variants, operation-specific narrowing, exhaustive handling; desktop vocabulary agreed first |
| 1 | kernel/presence.ts | Found, absent and unsuccessful lookup stay distinct |
| 1 | kernel/identity.ts | Identity values only when a consuming operation needs them |
| 2 | platform/codecs | Validate unknown successes and every failure payload; native wire shapes are plain data |
| 2 | services/preferences | A real read/change use case with injected capability and explicit errors |
| 2 | platform/desktop + preview | Equivalent observable contracts; native persistence and honest preview behavior |
| 3 | runtime/models + bindings | Scoped state, disposal, latest-read ordering; Svelte is confined to bindings |
| 3 | styles + components | Tokens → primitives → forms/feedback → compositions; accessible behavior with rendered checks |
| 4 | workbench/model → app → ui | Two views, activation, a registered command and shortcut; then restoration |
| 5 | workspace | Define workspace semantics, then its pure native rules and complete UI/native slice |
| 6 | jobs | Only when a real operation needs tracked progress/cancellation/shutdown |

Flover supplies reference concepts, not a runtime dependency. Adapt selected code
and tests locally. Pure domain rules remain framework-free. Result class instances
stay inside TypeScript; native adapters decode plain data into frontend Results.
Cancellation does not establish whether a write committed; retries need operation
safety, not merely a retryable failure kind. Diagnostic causes are not blindly
serialized or displayed to users.

Native leaves mirror the same ordering: errors/identity/time values before the
application that consumes them; config/logger/file IO only at effectful boundaries.
Preferences, Workspace and Jobs are candidate bounded contexts. Preferences
has a composed native slice; Workspace now has a composed native transport and
persistent root slice; Jobs remains contract-only. Workspace runtime-open
ownership remains a separate future slice.

Current working code: the Svelte component gallery and workbench UI run through
portable models and app services; root selects preview or native Preferences and
workbench restoration adapters. Native startup composes the file-backed
Preferences store and atomic `workbench.json` snapshot facade. Workspace
startup composes its Wails facade over one isolated memory registry, and the
frontend selects the matching native adapter in desktop mode, and Workspace
metadata persists through a versioned `workspace.json` document. Jobs remains
reserved until a concrete operation needs it.

See [DOMAIN_GUIDE.md](DOMAIN_GUIDE.md) for backend domain build order,
consumer-owned ports, memory adapters, fixtures and progression to persistence.

See [WORKERS.md](WORKERS.md) for bounded worker tasks, templates,
[build order](work/BUILD_ORDER.md) and [task status](work/manifest.json).

Native errors is implemented and integrated. See
[errors evidence](work/handoffs/native-errors.md) and the completed combined gate above.

Errors revision 2 now adds condition preservation, owned metadata, classification
inspection and safe public projection in both languages. Bridge encoding remains
a future contract. See the updated errors handoff for 19-test coverage.

Process config is implemented and integrated: captured environment/map lookups,
strict readers and redacted manifests. [Config evidence](work/handoffs/native-config.md).
Root settings and persisted user preferences remain separate future work.

IDs now provide validated UUID values, UUIDv7 generation with explicit rollback
and exhaustion behavior, and deterministic fixture sequences.
[ID evidence](work/handoffs/native-id.md). Provenance now consumes IDs through an explicit generation capability.

Execution provenance is implemented: work contexts, scopes, attribution,
causation, preparation, execution and retries.
[Provenance evidence](work/handoffs/native-provenance.md). Logger is now implemented;
transport propagation and durable history remain separate future slices.

Console logging is implemented with swappable sinks, safe error/provenance
projection and memory capture. Native root owns the stderr default, a generated
bootstrap provenance scope and lifecycle records. [Logger evidence](work/handoffs/native-logger.md).
[Bootstrap evidence](work/handoffs/native-root-bootstrap.md). The Preferences pure
domain is implemented against its module contract; see
[domain evidence](work/handoffs/preferences-domain.md). Its application
services are implemented with consumer-owned ports and an in-process event
seam; see [application evidence](work/handoffs/preferences-app.md). The
memory adapter passes the reusable store scenario suite; see
[memory evidence](work/handoffs/preferences-memory.md). The desktop wire
contract is specified and its framework-free handler implemented; see
[wire evidence](work/handoffs/desktop-wire-contract.md). Root now composes the
slice over the memory adapter and binds the native facade; see
[native slice evidence](work/handoffs/preferences-native-slice.md). Preferences persist
as one atomically replaced document selected explicitly by root; see
[persistence evidence](work/handoffs/preferences-persistence.md) and the
[storage decision](decisions/2026-09-12-preferences-file-persistence.md). The
frontend half of the round trip is still open. The Workspace pure domain is
implemented against its module contract; see
[workspace evidence](work/handoffs/workspace-domain.md). Its application
boundary and memory adapter now pass the shared WA01–WA10 and WM01–WM10
scenarios; see [application evidence](work/handoffs/workspace-app.md) and
[memory evidence](work/handoffs/workspace-memory.md). Its native transport,
Wails facade/root composition and desktop adapter are now implemented; see
[transport evidence](work/handoffs/workspace-native-transport.md). Workspace
state now survives restart through the versioned persistence adapter; see
[persistence evidence](work/handoffs/workspace-persistence.md). Runtime-open
now re-reads the authoritative record and activates Workbench context through
[its workflow handoff](work/handoffs/workspace-runtime-open.md). Workspace-
scoped restoration remains separate.

The three revision 1 domain boundaries are collected in
[DOMAIN_CONTRACTS.md](DOMAIN_CONTRACTS.md), with mirrored contracts in the Tauri
project.
