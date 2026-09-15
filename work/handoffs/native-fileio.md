# native-fileio handoff

Contract revision: 1 (`pkg/fileio/CONTRACT.md`). Implemented and
coordinator-integrated on 2026-09-12.

Files: `pkg/fileio/fileio.go`, `spec_test.go`, `README.md`, `CONTRACT.md`.
Production imports are the standard library and shared errors.

Scenario coverage: F01 `TestF01ReadAbsentAndDirectory`, F02
`TestF02ReplaceThenRead`, F03 `TestF03ReplaceExisting`, F04
`TestF04FailedReplaceLeavesTarget` (missing directory and a read-only
directory holding the target; skipped as root), F05 `TestF05EnsureDir`. F06
is asserted by `requireFailure` in every failure check: kind, type, path
detail, operating system cause, and a public message without the path.

Verification: `go test -race -count=1 -v ./pkg/fileio`: 5 tests passed.
`go vet ./...`: passed. `gofmt -l pkg/fileio`: empty.

Mechanism: unique temporary name beside the target, exclusive create with
mode 0600, write, fsync, rename, then directory fsync on non-Windows. The
temporary file is removed on any failure. Crash consistency under power loss
is a property of the mechanism and the filesystem, not proven here.
