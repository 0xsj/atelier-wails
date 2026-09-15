# workspace-domain handoff

Contract revision: 1 (`internal/workspace/domain/CONTRACT.md`). Implemented
and coordinator-integrated on 2026-09-12.

## Origin

The v0.1 commit carried an unrecorded workspace draft: domain, application
services and memory registry with no task, contract or handoff. The domain
exposed mutable exported fields, an untyped store failure hid conflicts, and
absence on rename, archive and restore returned a zero result with a nil
error. This task reconciled the domain against a written contract and
adapted the drafts so absence is reported distinctly.

## Files

`internal/workspace/domain/domain.go` (rewritten), `spec_test.go` (new,
replaces `domain_test.go`), `README.md`, `CONTRACT.md` (new). Coordinator
integration outside the task: `app/port/port.go`, `app/command/command.go`,
`app/query/query.go` and `infra/memory/memory.go` adapted to accessor-based
workspaces and a found flag for every single-workspace operation; the memory
draft now classifies ID and location collisions as `conflict /
workspace.id_taken` and `workspace.location_taken`; `memory_test.go`
reduced to one draft lifecycle check.

## Scenario coverage

WD01 `TestWD01Names` … WD11 `TestWD11InvalidCurrent`; WD12 is asserted by
`requireFailure` in every refusal check (kind, type, no name or location in
the message or fields). WD03 covers the Go normalization of a sub-millisecond
instant in a non-UTC zone.

## Verification

`go test -count=1 -v ./internal/workspace/domain`: 11 tests passed.
`go test -race -count=1 ./internal/workspace/...`: domain and memory passed.
`go test -race -count=1 ./...`: all packages passed. `go vet ./...`:
passed. `gofmt -l internal/workspace`: empty.
`python3 tools/work/verify_manifest.py`: valid.

Coordinator source review: production imports are the standard library,
shared errors and shared IDs. Decisions validate the new input before
comparing the current workspace and never write through it; the current
workspace is checked for validity first because Go zero values are
constructible. Implementation-visible verification only.

## Limitations

The application services and memory registry remain drafts with no
contract; workspace-app and workspace-memory must follow the preferences
pattern before any consumer. No transport, host or root wiring exists for
Workspace. Frontend files were not changed.
