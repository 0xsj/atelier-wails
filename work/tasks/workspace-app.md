# workspace-app: application services and consumer-owned ports

Contract: `internal/workspace/app/CONTRACT.md`, revision 1.
Dependencies: `workspace-domain`, `native-errors`, `native-id`.

Implement the Workspace command/query boundary over context-aware,
consumer-owned ports. Decide absence per operation, add the publisher seam,
preserve classified failures with operation context, and verify the services
over the memory adapter. Keep transport, host wiring, persistence and runtime
opening out of this slice. The sibling Tauri repository carries the same task
and scenario IDs with Rust-native signatures.

State authority: `work/manifest.json`. Coordinator owns the manifest and this
task's handoff.

Checks are listed in the contract's Verification section. Record exact
commands and counts in `work/handoffs/workspace-app.md`.
