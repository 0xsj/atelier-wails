# workspace-native-transport: native Workspace wire and root slice

Contract: `internal/workspace/transport/desktop/CONTRACT.md`, revision 1.
Dependency: `workspace-memory`.

Implement the concrete Workspace desktop consumer slice in the Wails host:

- define a framework-free plain-data handler for read, list, register, rename,
  archive, restore and forget;
- preserve the frontend wire vocabulary, decimal revisions, UTC millisecond
  timestamps, public failure projection and write commit classification;
- expose one JSON document per bound method through the Wails facade;
- compose one process-local memory store into both Workspace application ports;
- select the native Workspace adapter from the desktop frontend root while
  retaining the preview transport in browser mode.

Do not add filesystem persistence, runtime-open behavior, event subscribers or
window orchestration. Those are separate slices with their own contracts.

State authority: `work/manifest.json`. The sibling Tauri repository carries the
same task and WT01–WT07 scenario IDs.

Verification is listed in the manifest and the transport contract. Record exact
commands and counts in `work/handoffs/workspace-native-transport.md`.
