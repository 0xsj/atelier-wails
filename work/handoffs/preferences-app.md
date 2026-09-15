# preferences-app handoff

Contract revision: 1 (`internal/preferences/app/CONTRACT.md`). Implemented and
coordinator-integrated on 2026-09-12.

## Origin

The v0.1 draft had a shared `app/port` package, services that returned plain
`errors.New` values, no publisher, no context, no default merging and no
tests. This task replaced it: ports now live beside the command and query
services that consume them, failures keep their classification under an
operation annotation, events are published after the store returns, and
`Resolve` merges a fallback without writing.

## Files

`internal/preferences/app/command/command.go`, `spec_test.go`, `README.md`;
`internal/preferences/app/query/query.go`, `spec_test.go`, `README.md`;
`internal/preferences/app/CONTRACT.md` (new). Removed:
`internal/preferences/app/port/`. Coordinator integration outside the task:
`internal/preferences/infra/memory/memory.go` and `memory_test.go` gained a
leading `context.Context` on every method so the draft satisfies both ports;
the adapter ignores it and its contract still waits for preferences-memory.

## Scenario coverage

Command `spec_test.go`: A01 `TestA01Construction`, A02
`TestA02CreatePublishesOnce`, A03 `TestA03UnchangedPublishesNothing`, A04
`TestA04ConflictLeavesStateAndPublishesNothing`, A05
`TestA05InvalidInputsRefusedBeforeStore`, A06 `TestA06RemovePresentThenAbsent`,
A07 `TestA07StoreFailurePassesThrough`, A08 `TestA08PublisherSeesCommittedState`,
A11 `TestA11CanceledContextPassesThrough`. Query `spec_test.go`: A01, A05, A07,
A11 for the read side, A09 `TestA09ReadAndList`, A10 `TestA10ResolveNeverWrites`.
A12 is asserted by `requireFailure` in every refusal check of both files:
operation prefix, kind, type, and no echo of key or value text; A07
additionally checks `errors.Is` identity and an unclassified passthrough.

## Verification

`go test -count=1 -v ./internal/preferences/app/...`: 9 command and 6 query
tests passed. `go test -race -count=1 ./internal/preferences/...`: four
packages passed. `go test -race -count=1 ./...`: all packages with tests
passed. `go vet ./...`: passed. `gofmt -l internal/preferences`: empty.
`python3 tools/work/verify_manifest.py`: valid.

Coordinator source review: production imports are the standard library, the
preferences domain and shared errors; the memory adapter is imported only by
tests. Annotation uses a wrapping error so inspection reaches the inner
Failure and `errors.Is` still matches the original condition. Publication
happens only after the store call returns Changed or Removed. Implementation-
visible verification only.

## Limitations

The publisher is an in-process seam; delivery is not durable and no
subscriber exists yet. The memory adapter remains a draft. No transport,
host or root wiring exists for Preferences; nothing constructs these services
in production yet. Frontend files were not changed.
