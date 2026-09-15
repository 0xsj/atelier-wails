# Worker build order

Source of task status: [manifest.json](manifest.json). Session-by-session
narrative and resume instructions: [STATUS.md](STATUS.md). The leaves, their integration gate, and the preferences domain, application,
memory adapter and desktop wire transport are complete; the Workspace native
transport/root slice and file-backed persistence are complete; file IO and
file-backed persistence are complete and root selects the persistent stores.
Run `python3 tools/work/verify_manifest.py` before dispatch.

```text
native-errors ─┐
native-clock ──┼→ native-leaves-integrate → preferences-domain → preferences-app
native-secret ┘                              ↑                    ├→ preferences-memory ───┐
preferences-contract ────────────────────────┘                    └→ desktop-wire-contract ─┤
                                                                                           ↓
                                                                             preferences-native-slice
                                                                                           ↓
                                                                             preferences-persistence
```

The leaf integration task can review implemented handoffs together; mark the
leaves complete only after that gate passes. Other dependent tasks require their
prerequisites complete before dispatch. Planned tasks require contracts and task
files even when all dependencies are complete.

The graph schedules the calibration batch before domain implementation; it does
not force preferences to import clocks or secrets. Config, IDs, logger, file IO and the
source architecture check are complete.
Preferences, Workspace and Jobs contracts are specified independently of the
leaf batch. Preferences remains the first implementation slice.

Do not parallelize changes to Cargo module registration, root wiring, dependency
manifests or lockfiles. Keep one coordinator responsible for those files.

## Completed first assignments

- [native-errors](tasks/native-errors.md)
- [native-clock](tasks/native-clock.md)
- [native-secret](tasks/native-secret.md)

Then the coordinator runs [native-leaves-integrate](tasks/native-leaves-integrate.md).
Setup evidence: [worker setup verification](../notes/worker-setup-verification.md).

Config is also complete after errors and secrets. See [native-config](tasks/native-config.md)
and its [handoff](handoffs/native-config.md). No new task becomes ready merely
because these dependencies are complete; planned work still needs its contract.

IDs are complete after errors: [native-id](tasks/native-id.md),
[verification](handoffs/native-id.md). Provenance is now complete; logger is now complete.

Completed: [native-provenance](tasks/native-provenance.md),
[verification](handoffs/native-provenance.md), after errors/IDs/clocks.
Logger uses provenance for its explicit projection contract.

Completed: [native-logger](tasks/native-logger.md),
[verification](handoffs/native-logger.md). Console is composed at native startup.
Preferences, Workspace and Jobs remain planned for implementation; their
revision 1 contracts are complete.

Completed contract tasks: [preferences-contract](tasks/preferences-contract.md),
[workspace-contract](tasks/workspace-contract.md) and
[jobs-contract](tasks/jobs-contract.md), with specification evidence in their
matching handoffs.

Completed: [native-root-bootstrap](tasks/native-root-bootstrap.md),
[verification](handoffs/native-root-bootstrap.md). The root now creates the
bootstrap ID/provenance scope and binds it to lifecycle logging.

Completed: [preferences-domain](tasks/preferences-domain.md),
[verification](handoffs/preferences-domain.md). The v0.1 commit carried
unrecorded draft code for preferences and workspace; the domain was reconciled
against its own contract, and the application, memory and workspace drafts
stay unreviewed until preferences-app, preferences-memory and a workspace
reconcile task are specified.

Completed: [preferences-app](tasks/preferences-app.md),
[verification](handoffs/preferences-app.md). Command and query services own
their store ports and an in-process publisher; the memory adapter draft
satisfies those ports and is next to specify under preferences-memory.

Completed: [preferences-memory](tasks/preferences-memory.md),
[verification](handoffs/preferences-memory.md). The adapter passes the
reusable M01–M12 suite, which a persistent adapter reruns. Next is
desktop-wire-contract, then preferences-native-slice composes root, memory
and the frontend round trip.

Completed: [desktop-wire-contract](tasks/desktop-wire-contract.md),
[verification](handoffs/desktop-wire-contract.md). The framework-free
handler passes W01–W12 with byte-identical JSON fixtures in both languages.
preferences-native-slice now composes host facade, root, memory and the
frontend codecs against that contract.

Implemented (backend half): [preferences-native-slice](tasks/preferences-native-slice.md),
[evidence](handoffs/preferences-native-slice.md). Root composes memory,
services, handler and the bound facade. The frontend adapter, codecs and
visible round trip are the frontend session's half; the task stays
`implemented` until both halves are reviewed together.

Completed: [native-fileio](tasks/native-fileio.md),
[verification](handoffs/native-fileio.md), and
[preferences-persistence](tasks/preferences-persistence.md),
[verification](handoffs/preferences-persistence.md). Preferences now survive
restart through one atomically replaced document under the platform config
directory; see the storage decision record. Remaining planned backend work:
workspace-app and workspace-memory.

Completed: [workspace-domain](tasks/workspace-domain.md),
[verification](handoffs/workspace-domain.md). The workspace draft's domain
is reconciled against its contract; the Rust registry no longer panics on an
absent workspace. Its application and memory drafts wait for workspace-app
and workspace-memory, which follow the preferences pattern.

Completed: [architecture-checks](tasks/architecture-checks.md),
[verification](handoffs/architecture-checks.md). The import rules that every
handoff claimed by review are now executable; run the check with the
manifest verifier before dispatch and completion.

Completed: [workspace-app](tasks/workspace-app.md),
[verification](handoffs/workspace-app.md). Workspace command and query
services now own their context-aware ports, absence policy, operation context
and post-commit publisher seam in step with the Tauri implementation.

Completed: [workspace-memory](tasks/workspace-memory.md),
[verification](handoffs/workspace-memory.md). The synchronized in-memory
registry passes reusable WM01–WM10 scenarios, including location uniqueness,
restore collision and concurrent conditional writes. Native Workspace
transport, host/root composition and the matching desktop adapter are now
complete: [workspace-native-transport](tasks/workspace-native-transport.md),
[verification](handoffs/workspace-native-transport.md). Workspace persistence
is complete through [workspace-persistence](tasks/workspace-persistence.md)
and [verification](handoffs/workspace-persistence.md). Runtime-open is now
complete through [workspace-runtime-open](tasks/workspace-runtime-open.md) and
[verification](handoffs/workspace-runtime-open.md). Next is workspace-scoped
Workbench restoration.
