# Workbench snapshot port handoff

Date: 2026-09-14 · Scope: frontend restoration and native host document port

The existing consumer-owned `SnapshotPort` now has a JSON document bridge.
Browser previews continue to use `createMemorySnapshotPort`; the Wails root
injects a desktop adapter backed by `workbench.json` in the user config
directory. The Go host facade uses the shared atomic `fileio` leaf for reads
and whole-document replacement. Missing files return empty; malformed UTF-8 or
filesystem failures reject as unavailable; malformed JSON remains text so the
workbench model reports its existing invalid outcome.

Files:

- `frontend/src/lib/workbench/app/restoration/` — document port, adapter and
  tests, synchronized from the Tauri frontend.
- `frontend/src/lib/platform/desktop/workbench.ts` — Wails binding adapter.
- `internal/host/wails/workbench.go` and its tests — file facade.
- `internal/host/wails/host.go` — binding registration.
- `root/root.go` and `root/workbench.go` — config-path composition.

Verification:

- `make check`: frontend check/test/build passed; `go vet ./...` passed.
- `go test -count=1 ./...`: all packages passed.
- `VITE_ATELIER_PROBE=workbench make build`: passed; generated the Wails
  bindings, launched the desktop app, and restored `probe.snapshot` through
  the real bound-struct boundary. The resulting `workbench.json` was read
  from the Wails config directory and moved to
  `/private/tmp/atelier-workbench-probe-wails.json` afterwards.

Not verified: native window interaction and visual UI behavior. Screen capture
and assistive access remain unavailable to the terminal session. The generated
Wails bindings are exercised by `make build`, not by the frontend type check.
