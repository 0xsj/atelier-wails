# workspace-native-transport handoff

Contract revision: 1 (`internal/workspace/transport/desktop/CONTRACT.md`).
Implemented and coordinator-integrated on 2026-09-15.

## Delivered

The Wails Workspace transport now accepts one JSON document per operation and
returns a plain `{ok,value}` or `{ok:false,failure}` envelope. The handler
decodes the shared frontend vocabulary, validates IDs, filters, expected
revisions and UTC RFC3339 timestamps, and projects failures without echoing
untrusted input. Bound methods live on a Workspace facade and root composes one
synchronized process-local memory store into its command and query services.

The frontend desktop adapter calls the seven bound Workspace methods, while the
browser root keeps the preview transport. This makes `WorkspacePicker` use the
same service and codec in both modes without making generated Wails bindings a
frontend type-check dependency.

## Scenario evidence

WT01 object-only documents, WT02 read/list, WT03 register and lifecycle event
responses, WT04 canonical revisions/timestamps, WT05 safe invalid-field
projection, WT06 failure classification/commit and WT07 shared-store state all
pass in `transport/desktop/spec_test.go`.

## Verification

- `gofmt -l internal/workspace internal/host/wails root`: no output.
- `go test -count=1 ./internal/workspace/transport/desktop`: 7 transport tests passed.
- `go test -race -count=1 ./...`: all packages passed.
- `go vet ./...`: clean.
- `npm run check`: Svelte diagnostics found 0 errors and 0 warnings.
- `npm test`: 16 test files and 82 tests passed.
- `python3 tools/work/verify_manifest.py`: passed after the paired manifest entries and handoffs were added.
- `python3 tools/architecture/check_imports.py`: passed with 0 violations.

## Limits

Workspace state is process-local and is lost on restart. Runtime-open,
filesystem persistence, event subscriptions and native window behavior remain
unimplemented and should be introduced only through their own contracts.
