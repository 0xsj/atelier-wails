# preferences-memory handoff

Contract revision: 1 (`internal/preferences/infra/memory/CONTRACT.md`).
Implemented and coordinator-integrated on 2026-09-12.

## Origin

The v0.1 draft adapter already held a synchronized map with the domain
decision inside the lock, but had two ad hoc tests, an untyped private error,
and no reusable suite. This task kept the shape, keyed the map by scope
value, moved input refusals to the domain categories, and delivered the
scenario suite and fault wrappers a persistent adapter reruns.

## Files

`internal/preferences/infra/memory/memory.go`, `memory_test.go`,
`CONTRACT.md` (new); `internal/preferences/infra/storetest/storetest.go`
and `README.md` (new, test-only package). No production file imports
`storetest`.

## Scenario coverage

`storetest.Run` executes M01–M12 as subtests named after each scenario
(`M01FreshIsEmpty` … `M12FaultWrappers`); `memory_test.go` calls it through
`TestContract` and adds the Go-only `TestInvalidReadInputs` for zero-value
inputs. M11 runs eight goroutines released by a shared start signal for each
of the three races. M12 exercises `FailBefore` and `LoseAcknowledgement`,
including the retry at revision 1 that commits behind a second lost
acknowledgement and the equal write that must not lose its acknowledgement.

## Verification

`go test -count=1 -v ./internal/preferences/infra/...`: 12 subtests and 2
tests passed in the memory package; storetest has no tests of its own.
`go test -race -count=1 ./internal/preferences/...`: four packages passed,
race detector silent through M11.
`go test -race -count=1 ./...`: all 14 packages passed. `go vet ./...`:
passed. `gofmt -l internal/preferences`: empty.
`python3 tools/work/verify_manifest.py`: valid.

Coordinator source review: production imports are the standard library, the
domain and shared errors. Each conditional write copies the current entry
out of the map, runs the domain decision and applies the change under one
write lock. Reads take the read lock. Lists are freshly allocated and sorted
bytewise. Implementation-visible verification only.

## Limitations

Memory evidence only: no restart, corruption, migration or IO failure has
been exercised, and none is claimed. Root does not yet construct this store;
the native slice must select it explicitly. The workspace draft is untouched.
Frontend files were not changed.
