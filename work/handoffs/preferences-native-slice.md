# preferences-native-slice handoff (backend half)

Coordinator integration on 2026-09-12. State: implemented, not complete. The
Wails root now composes the preferences slice over the memory adapter and
binds its facade; the frontend adapter, codecs, service and visible round
trip remain with the frontend session.

## Files

`internal/host/wails/preferences.go` (new facade), `internal/host/wails/host.go`
(`Run` takes `Facades`, binds the facade and hands it the startup context),
`root/preferences.go` (composition), `root/preferences_test.go`, `root/root.go`.
Transport additions: `ReadDocument` … `ResolveDocument` and the object-only
document decoder in `internal/preferences/transport/desktop/handler.go`, W13
in `spec_test.go`, and the matching wording and scenario appended to the wire
contract without changing its promises.

## Scenario coverage

N01, N02 `TestComposedPreferencesRoundTrip`; N03
`TestComposedPreferencesInvalidDocument`; N05 `TestComposedPreferencesAreIndependent`;
N04 by `go build` of the application with the facade in the bind list. W13
`TestW13Documents` covers the document entry points, including `[]`, `42`,
`null` and a JSON string, which both hosts now refuse identically.

## Verification

`go test -race -count=1 ./...`: all 15 packages passed. `go vet ./...`:
passed. `gofmt -l internal root`: empty.
`go build -o /tmp/atelier-wails-native-slice-check .`: built.
`python3 tools/work/verify_manifest.py`: valid.

Coordinator review: the facade holds no Wails type; only `host.go` imports
the framework. Root selects `memory.New()` explicitly and documents the
volatile lifetime. The startup context reaches the facade through an
unexported method so Wails does not bind it.

## Not verified

No native window was launched and no frontend code invokes the facade yet,
so the JavaScript-to-Go round trip, Wails' TypeScript binding generation for
the outcome types, and the visible UI are unproven. Wails generates its
`frontend/wailsjs/go` bindings on the next `make dev` or `make build`; that
touches the frontend tree and was deliberately not run here. The workspace
draft is untouched. Frontend files were not changed.

## Frontend half (frontend session, 2026-09-12)

Files: `kernel/`, `platform/codecs/preferences.ts`, `services/preferences/`,
`platform/preview/preferences.ts`, `platform/desktop/` (Wails adapter and
host detection), `features/preferences/`, `root/App.svelte`, the gallery's
Preferences section, and tests beside each. The service maps a rejected native
promise to `internal` with commit `unknown` for writes and `none` for reads,
and reconciles an uncertain commit by re-reading before retrying with the
observed revision, as the wire contract requires.

Evidence: `npm test` 14 files, 73 tests; `npm run check` clean; `npm run
build` passes; codec and preview tests pin the contract's create and conflict
fixtures byte for byte. Native launch evidence:

Native round trip, 2026-09-12: screen capture and assistive access were both
denied to the session, so the round trip was proven on disk instead. A probe
build (`VITE_ATELIER_PROBE=preferences make build`) makes the gallery write one preference at startup through
the real adapter and read it back (`src/dev/probe.ts`, compiled away without
the flag). Launching the built app produced `~/Library/Application Support/atelier-wails/preferences.json`:

```json
{"format":1,"entries":[{"scope":{"kind":"global"},"key":"probe.launch","value":{"kind":"text","text":"Wails / Go 2026-09-12T13:24:41.111Z"},"revision":"1"}]}
```

That document is the file-backed store's format 1, written by the native
handler after decoding the request the frontend codec encoded. The probe key
was deleted afterwards and the apps rebuilt without the flag. The visible UI
was not inspected; keyboard and dialog behavior remain unverified in a window.

