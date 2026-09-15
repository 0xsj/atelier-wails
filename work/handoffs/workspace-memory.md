# workspace-memory handoff

Contract revision: 1 (`internal/workspace/infra/memory/CONTRACT.md`).
Implemented and coordinator-integrated on 2026-09-15.

## Delivered

The in-memory Workspace registry now implements both context-aware store
ports. Each conditional decision and commit runs under one write lock. Active
location uniqueness is atomic, archived locations may be reused, restore
detects active collisions, lists are deterministic by Workspace ID, each
store instance is isolated, and returned values are safe to reorder without
changing stored state.

`infra/storetest` is a reusable test-only contract suite. Its factory shape
lets a future persistent adapter rerun WM01–WM10 before adding storage-specific
evidence.

## Scenario evidence

The memory contract runs WM01 fresh state, WM02 register/read/list, WM03
location uniqueness and restore collision, WM04 lifecycle and stale conflict,
WM05 absence and invalid input, WM06 deterministic ordering and filters, WM07
independent instances, WM08 output isolation, WM09 concurrent registration
and WM10 concurrent conditional rename.

## Verification

- `go test -count=1 -v ./internal/workspace/app/... ./internal/workspace/infra/memory`: 10 named WM subtests passed under `TestWorkspaceMemoryContract`.
- `go test -race -count=1 ./...`: all packages passed, including the concurrent memory scenarios.
- `go vet ./...`: clean.
- `gofmt -l internal/workspace`: no output.
- `python3 tools/work/verify_manifest.py`: passed after the paired handoffs were added.
- `python3 tools/architecture/check_imports.py`: passed with 0 violations.

## Limits

The adapter is process-local and non-persistent. No Workspace persistence,
transport or root wiring was inferred from this contract.
