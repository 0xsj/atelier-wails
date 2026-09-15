# workspace-runtime-open handoff

Implemented and coordinator-integrated on 2026-09-15.

## Delivered

`createWorkspaceOpenService` is the explicit frontend workflow port for
activation. It re-reads the selected Workspace through the existing native or
preview transport, returns an authoritative `WorkspaceSession` for active
records, and maps absent or archived records to explicit failures. The picker
now awaits that port before calling its success callback. The composition root
injects it in both browser preview and desktop modes, and the Workbench renders
the active workspace name/location as its session context.

No registry mutation, filesystem inspection or window orchestration was added.
The existing native Workspace persistence remains the source of truth.

## Scenario evidence

- RO01 re-reads and refreshes a selected active record.
- RO02 refuses archived records with `workspace.archived` and `not_applied`.
- RO03 maps a record that disappeared before activation to `workspace.not_found`.
- The rendered WorkspacePicker contract proves a successful open reaches the
  Workbench-facing callback.

## Verification

- `npm --prefix frontend run check`: 0 Svelte errors and 0 warnings.
- `npm --prefix frontend test`: 17 test files and 85 tests passed.
- `go test -race -count=1 ./...`: passed.
- `go vet ./...`: passed.
- `python3 tools/work/verify_manifest.py`: passed after the paired manifest
  entries and handoffs were added.
- `python3 tools/architecture/check_imports.py`: passed with 0 violations.

## Limits

The session currently scopes the Workbench context but does not yet select a
workspace-specific snapshot document. Native window orchestration and event
subscriptions remain separate slices.
