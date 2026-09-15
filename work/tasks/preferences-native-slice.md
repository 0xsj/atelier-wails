# preferences-native-slice: compose the real native round trip

Coordination task. Dependencies: preferences-memory revision 1,
desktop-wire-contract revision 1. State authority: work/manifest.json.

Backend half (this repository's host and root): bind the desktop transport
handler as the native facade, select the memory adapter explicitly in root,
and prove the composed facade serves the wire documents. Frontend half: the
platform/desktop adapter, codecs and a service that invoke the facade and
validate the envelope, then a visible round trip in the native window. The
frontend half is owned by the frontend session; the task reaches `complete`
only when the coordinator has both halves' evidence.

Scenarios, backend half:

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| N01 | Root composes the slice | One memory store serves the command and query ports; the publisher is Discard; no persistence is claimed; the host binds the facade |
| N02 | The composed facade receives the create, read, list, resolve and remove documents | The create envelope equals the wire contract fixture; the others succeed with the expected values |
| N03 | The facade receives a document that is not a JSON object | `invalid / desktop.invalid_request` with `request: invalid`, commit by operation; never a framework error |
| N04 | Host registration | The five operations are bound (Wails methods `Read` … `Resolve`; Tauri commands `preferences_read` … `preferences_resolve`) and the app compiles with them |
| N05 | Two compositions | Independent stores; state does not leak between them |

Checks: the full native test suite, vet or check, formatting, and a build
(`go build` or `cargo check` including the generated Tauri context).
