# workspace-memory: reusable in-memory adapter contract

Contract: `internal/workspace/infra/memory/CONTRACT.md`, revision 1.
Dependency: `workspace-app`.

Implement the isolated, synchronized Workspace registry and the reusable
WM01–WM10 adapter scenarios. Cover active-location uniqueness, archived
location reuse, restore collision, stale conditional writes, deterministic
lists, output isolation and concurrency. The contract suite must be reusable
by a future persistent adapter. Do not add persistence, transport or root
wiring here.

State authority: `work/manifest.json`. Coordinator owns the manifest and this
task's handoff. The sibling Tauri repository carries the same task and
scenario IDs.

Checks are listed in the contract's Verification section. Record exact
commands and counts in `work/handoffs/workspace-memory.md`.
