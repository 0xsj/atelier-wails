# Backend status journal

One dated entry per working session. Newest first. Written so that a fresh
agent can continue without the previous conversation. The manifest is the
task authority; this journal records why things are the way they are and
what to do next. Frontend status is not tracked here.

## How to resume

1. Read AGENTS.md, WORKERS.md, DOMAIN_GUIDE.md and this file.
2. Run the two gates before and after any change:

   ```sh
   python3 tools/work/verify_manifest.py
   python3 tools/architecture/check_imports.py
   ```

3. Run the full native suite:

   ```sh
   go test -race -count=1 ./...
   go vet ./...
   gofmt -l internal pkg root
   go build -o /tmp/atelier-wails-check .
   ```

4. Every task follows one pattern: a `CONTRACT.md` beside the module with a
   `Revision: N · State: specified · Task: <id>` header and numbered
   scenarios; a task file in `work/tasks/<id>.md`; implementation with one
   test per scenario; a handoff in `work/handoffs/<id>.md` with the exact
   commands and counts; the manifest entry moved to `complete`; a note in
   `work/BUILD_ORDER.md` and FOUNDATION.md. The sibling Tauri repository
   carries the same task IDs and scenario IDs so the two stay in step.
5. Do not touch the root README.md, AGENTS.md or
   `notes/svelte-foundation-verification.md`. Frontend work is in progress in
   the working tree by a separate effort; this slice's Workspace desktop
   adapter and root selection are the explicitly scoped exception. Do not
   commit or revert unrelated frontend files.

## 2026-09-15

### Completed today

| Task | What it delivered | Evidence |
| --- | --- | --- |
| workspace-app | Context-aware command/query ports, explicit absence policy, post-commit publisher and operation-preserving failures | WA01–WA10 across `internal/workspace/app/*/spec_test.go`; [handoff](handoffs/workspace-app.md) |
| workspace-memory | Isolated synchronized registry and reusable adapter suite with deterministic ordering, uniqueness/restore collision and concurrency guarantees | WM01–WM10 in `internal/workspace/infra/storetest`; [handoff](handoffs/workspace-memory.md) |
| workspace-native-transport | Framework-free Workspace wire handler, Wails facade/root composition over process-local memory, and native frontend adapter selection | WT01–WT07 in `internal/workspace/transport/desktop/spec_test.go`; [handoff](handoffs/workspace-native-transport.md) |
| workspace-persistence | Atomic versioned Workspace document, restart preservation, corruption refusal, failed-write isolation, and durable Wails root composition | WP01–WP07 in `internal/workspace/infra/persistence`; [handoff](handoffs/workspace-persistence.md) |
| workspace-runtime-open | Re-read/validate selected Workspace metadata and hand the active session into the Workbench in preview and desktop roots | RO01–RO03 in `frontend/src/lib/services/workspace/runtime.test.ts`; [handoff](handoffs/workspace-runtime-open.md) |

The paired Wails/Tauri batch uses the same WA and WM scenario promises. The
Wails persistence has the shared ten named memory subtests plus restart,
corruption, failed-write and deterministic-byte coverage; root adds the WP07
restart proof. The full race suite, vet, manifest and architecture gates are
clean. Manifest: 25 tasks complete, 1 implemented, 0 planned.

### Decisions

- Rename, archive and restore map absent records to application
  `not_found / workspace.not_found`; Forget remains an idempotent absence
  no-op.
- Publishers receive only committed transition events; unchanged, refused,
  absent and failed operations do not publish.
- The memory contract is deliberately adapter-shaped so persistence can rerun
  WM01–WM10 before adding storage-specific behavior.

### Next task: workspace-runtime-open

Workspace state now survives restart through the versioned native document. The
runtime-open workflow now re-reads the authoritative record and activates the
Workbench context without coupling persistence to window orchestration. The
next bounded slice is workspace-scoped Workbench restoration. Jobs remains
contract-only.

## 2026-09-12

### Starting point

The v0.1 commit contained unrecorded draft code for the preferences and
workspace contexts (domain, application services, memory adapters) while
the manifest still called those tasks planned and the directory READMEs
said reserved. The user chose to reconcile the drafts task by task against
written contracts rather than delete them. Everything below was done in
that spirit and nothing has been committed yet: the backend changes sit in
the working tree alongside the separate frontend work.

### Completed today, in order

| Task | What it delivered | Evidence |
| --- | --- | --- |
| preferences-domain | Pure values with private fields and validated `Restore`, compare-and-replace decisions, returned events | `internal/preferences/domain/spec_test.go`, P01–P15 |
| preferences-app | Command and query services owning their ports, publisher seam, operation-annotated failures, `Resolve` with fallback | A01–A12 across command and query `spec_test.go` |
| preferences-memory | Memory adapter plus the reusable `storetest` suite and fault wrappers | M01–M12 in `internal/preferences/infra/storetest` |
| desktop-wire-contract | JSON wire shapes, decode vocabulary, public failure projection, commit field, framework-free handler | W01–W13 in `internal/preferences/transport/desktop` |
| preferences-native-slice (backend half, state `implemented`) | Wails-bound `Preferences` facade taking one JSON document per call, root composition, host binding | N01–N05 in `root/preferences_test.go` |
| native-fileio | Read with absence, atomic durable replace, ensure directory | F01–F06 in `pkg/fileio` |
| preferences-persistence | One JSON document per store replaced atomically; root selects it under the platform config directory | S01–S07 in `internal/preferences/infra/persistence`; decision record `decisions/2026-09-12-preferences-file-persistence.md` |
| workspace-domain | Workspace vocabulary and lifecycle decisions; drafts adapted so absence is a found flag | WD01–WD12 in `internal/workspace/domain/spec_test.go` |
| architecture-checks | Executable layer import rules with fixture tests | `tools/architecture/RULES.md`, C01–C05 |

Manifest: 20 tasks complete, 1 implemented, 0 planned. Native suite at end
of day: 17 packages green with the race detector, 177 test runs including
subtests, vet and gofmt clean, application binary builds. Architecture
check: 74 files, 142 internal imports, 0 violations.

### Decisions worth knowing

- Revision exhaustion, forgetting an active workspace and similar
  state-based refusals are `conflict` kind with their own type, so callers
  do not loop on re-read and retry.
- Wire and storage carry every 64-bit integer as a decimal string.
- Native facades take the request as one JSON string argument, so the
  contract's decode failure is produced by the transport, never by the
  framework.
- The publisher is an in-process seam; root wires `Discard` until a
  subscriber exists.
- Preferences persist as a file, not SQLite; SQLite remains the candidate
  for multi-record metadata and jobs. A corrupt document refuses startup
  with a logged failure rather than starting empty.

### Known limits and open items

- preferences-native-slice stays `implemented`: no native window was
  launched and no frontend code calls the facade. Wails will generate its
  TypeScript bindings under `frontend/wailsjs/go` on the next `make dev` or
  `make build`; that was deliberately not run. The frontend half must land
  its adapter, codecs and a visible round trip before the task is complete.
- Preferences persistence: one process owns the file, no migration or
  recovery flow, real power-loss durability rests on the fileio mechanism.
- Workspace application services and memory registry are drafts adapted
  to compile, not specified. Location uniqueness lives in the registry
  draft and is untested beyond one lifecycle check.
- Jobs is a contract only.
- The architecture check does not cover the frontend and is not wired into
  CI.
- The root README still says no behavioral test suite exists; left alone
  because the frontend effort is editing it.

### Next task: workspace-app

Follow the preferences-app pattern exactly. Write
`internal/workspace/app/CONTRACT.md` with scenarios, declare the store port
beside the command service and the read port beside the query service
(remove the shared `app/port` package), decide what absence means on
rename, archive, restore and forget at the application boundary (the
domain contract leaves it open; the registry draft reports a found flag),
add a publisher seam, annotate failures with the operation name, and test
over the memory registry draft with a recording publisher and a fault
store. Then workspace-memory: a reusable suite like `storetest` including
the location-uniqueness and restore-collision scenarios from the parent
contract. Do not add a workspace transport or root wiring until a UI or
workflow consumer exists.
