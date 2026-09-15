# preferences-domain handoff

Contract revision: 1 (`internal/preferences/domain/CONTRACT.md`). Implemented
and coordinator-integrated on 2026-09-12.

## Origin

The v0.1 commit carried unrecorded draft code for the preferences domain,
application services and memory adapter with no task, contract or handoff.
This task reconciled the domain against a written contract instead of building
on the draft as-is. Defects removed from the draft: a cast could bypass key
validation, entries and events exposed mutable exported fields, cross-kind value
equality compared every payload field, and refusal messages were not
distinguished from application errors.

## Files

`internal/preferences/domain/domain.go` (rewritten), `spec_test.go` (new,
replaces `domain_test.go`), `README.md`, `CONTRACT.md` (new). Coordinator
integration outside the task: `internal/preferences/infra/memory/memory.go`
and `memory_test.go` adapted to accessor-based entries and the shared failure
type for invalid input; no behavior change intended, their scenarios remain
unspecified until preferences-memory. The application command/query drafts
compile unchanged.

## Scenario coverage

P01–P14 map to `TestP01Scopes` … `TestP14ValidationPrecedesState` in
spec_test.go. P15 (shared kind and type, no echo of key or value text) is
asserted by `requireFailure` inside every refusal check rather than a separate
test. Go-only zero-value refusals are covered in P01, P02, P03, P04, P05, P13
and P14.

## Verification

`go test -count=1 -v ./internal/preferences/domain`: 14 tests passed.
`go test -race -count=1 ./internal/preferences/...`: domain and memory passed.
`go test -race -count=1 ./...`: all 12 packages with tests passed.
`go vet ./...`: passed. `gofmt -l internal pkg root main.go`: empty.
`python3 tools/work/verify_manifest.py`: valid.

Coordinator source review: production imports are the standard library, shared
errors and shared IDs only. Decisions validate inputs before inspecting the
current entry and never write through the current pointer. Every refusal is a
shared Failure with a fixed message; the revision-exhausted refusal is a
conflict with its own type. This is implementation-visible verification, not an
independent oracle.

## Limitations

No adapter, native bridge or rendered UI evidence exists for Preferences. The
memory adapter and services beside this module remain unreviewed drafts; their
tasks must adapt to this API. Workspace draft code is untouched and still
unrecorded. Frontend files were not changed by this task.
