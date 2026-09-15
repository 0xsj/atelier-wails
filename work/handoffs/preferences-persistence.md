# preferences-persistence handoff

Contract revision: 1 (`internal/preferences/infra/persistence/CONTRACT.md`).
Implemented and coordinator-integrated on 2026-09-12. Storage decision:
`decisions/2026-09-12-preferences-file-persistence.md`.

## Files

`internal/preferences/infra/persistence/persistence.go`, `record.go`,
`spec_test.go`, `README.md`, `CONTRACT.md` (new). Root now selects this
adapter: `root/preferences.go` resolves the document under the platform
config directory, ensures the directory and opens the store;
`root/preferences_test.go` covers composition, restart and corrupt refusal;
`root/root.go` logs a composition failure before returning it. The Wails
facade and transport are unchanged.

## Scenario coverage

S01 `TestS01Contract` runs the shared suite's twelve subtests over fresh
temporary paths. S02 `TestS02Restart`, S03 `TestS03AbsentFile`, S04
`TestS04CorruptDocuments` (eight documents), S05 `TestS05WriteFailures`
(read-only directory, then restored; missing directory; skipped as root),
S06 `TestS06DeterministicBytes` with the exact document bytes. S07 is
asserted by `requireFailure` in every failure check. Root:
`TestComposedPreferencesSurviveRestart` and
`TestCorruptDocumentRefusesComposition` prove the default selection.

## Verification

`go test -race -count=1 -v ./internal/preferences/infra/persistence`: 6
tests and 12 subtests passed, race detector silent through M11 over real
files. `go test -race -count=1 ./...`: all 17 packages passed. `go vet
./...`: passed. `gofmt -l internal root pkg`: empty.
`go build -o /tmp/atelier-wails-persistence-check .`: built.
`python3 tools/work/verify_manifest.py`: valid.

Coordinator review: production imports are the standard library, the
domain, the application ports, shared errors, shared IDs and shared fileio.
Every change clones the state, encodes the whole document, replaces the
file, and only then commits memory; a failed replacement leaves both memory
and file at the previous committed state. Decoding is strict (unknown keys
and trailing content are corruption) and goes through the domain's validated
construction.

## Limitations

One process owns the file; no cross-process coordination. No migration,
backup or recovery flow: a corrupt document refuses startup with a logged
failure. Real power-loss durability is a property of the fileio mechanism.
The native window and frontend round trip remain unverified, as recorded in
the native slice handoff. Frontend files were not changed.
