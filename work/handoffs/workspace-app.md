# workspace-app handoff

Contract revision: 1 (`internal/workspace/app/CONTRACT.md`). Implemented and
coordinator-integrated on 2026-09-15.

## Delivered

The shared draft `app/port` package was removed. `command` now owns the
context-aware conditional write port, publisher seam, `Discard` publisher and
operation service; `query` owns the read port and query service. Store
failures are wrapped with operation names while preserving their typed
classification, cause and context cancellation. Rename, archive and restore
map absence to `not_found / workspace.not_found`; Forget treats absence as a
successful no-op. Events publish synchronously only after committed changed
results, including successful Forget.

The memory registry implements both application ports and checks canceled
contexts before beginning an operation. No transport, host, root,
persistence or runtime-open behavior was added.

## Scenario evidence

The command package passes WA01, WA02, WA03, WA05, WA06, WA07, WA08 and WA09;
the query package passes WA01, WA04, WA07, WA09 and WA10. Together they cover
all shared WA01–WA10 promises, with Go-specific nil dependency, context and
pre-port validation checks.

## Verification

- `go test -count=1 -v ./internal/workspace/app/... ./internal/workspace/infra/memory`: command 8 tests passed, query 5 tests passed, and the memory contract's 10 named subtests passed.
- `go test -race -count=1 ./...`: all packages passed, including Workspace command, query, domain and memory packages.
- `go vet ./...`: clean.
- `gofmt -l internal/workspace`: no output.
- `python3 tools/work/verify_manifest.py`: passed after the paired handoffs were added.
- `python3 tools/architecture/check_imports.py`: passed with 0 violations.

## Limits

Publication is an in-process, non-durable seam wired to `Discard` until a
subscriber exists. Workspace native transport, root composition, persistence
and runtime opening remain future slices.
