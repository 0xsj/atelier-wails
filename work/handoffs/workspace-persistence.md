# workspace-persistence handoff

Contract revision: 1 (`internal/workspace/infra/persistence/CONTRACT.md`).
Implemented and coordinator-integrated on 2026-09-15.

## Delivered

The Workspace registry now has a file-backed adapter using one versioned JSON
document and the shared atomic replacement primitive. It stores IDs, names,
locations, lifecycle status, decimal revisions and millisecond timestamps in
deterministic ID order. Missing files open empty; corrupt, unsupported,
duplicate or structurally invalid documents refuse composition without rewriting
the original bytes. Failed writes leave both memory and disk unchanged.

The Wails root now selects this adapter at `workspace.json` beside the existing
Preferences and Workbench documents. The bound Workspace facade therefore
survives recomposition/restart without changing the frontend wire contract.

## Scenario evidence

WP01 reruns WM01–WM10, WP02 verifies restart and lifecycle preservation, WP03
covers absent-file creation, WP04 covers corruption refusal, WP05 covers failed
writes, WP06 covers deterministic bytes, and the root composition tests cover
WP07.

## Verification

- `gofmt -l internal/workspace internal/host/wails root`: no output.
- `go test -count=1 ./internal/workspace/infra/persistence ./internal/workspace/infra/storetest ./root`: all packages passed; persistence and root restart/corruption tests passed.
- `go test -race -count=1 ./...`: all packages passed.
- `go vet ./...`: clean.
- `python3 tools/work/verify_manifest.py`: passed after the paired manifest entries and handoffs were added.
- `python3 tools/architecture/check_imports.py`: passed with 0 violations.

## Limits

The schema is revision 1 with no migration path yet. Runtime-open behavior,
window orchestration and event subscriptions remain separate workflow slices.
