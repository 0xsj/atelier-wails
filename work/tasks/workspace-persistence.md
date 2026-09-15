# workspace-persistence: durable Workspace registry

Contract: `internal/workspace/infra/persistence/CONTRACT.md`, revision 1.
Dependency: `workspace-native-transport`.

Implement the Workspace persistence adapter and switch native root composition
to it. The adapter must rerun WM01–WM10, preserve complete Workspace metadata
across restart, use atomic file replacement, emit deterministic bytes, refuse
corrupt or unsupported documents, and leave state unchanged when a write fails.

Root owns the configured path and directory creation. Do not add runtime-open,
window behavior, migrations or event subscribers. The sibling Tauri repository
carries the same task and WP01–WP07 scenario IDs.

State authority: `work/manifest.json`. Record exact verification commands and
counts in `work/handoffs/workspace-persistence.md`.
