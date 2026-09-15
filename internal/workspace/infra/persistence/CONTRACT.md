# Workspace persistence adapter contract

Revision: 1 · State: implemented · Task: workspace-persistence

The persistence adapter stores the complete Workspace registry in one JSON
document and uses the shared atomic file replacement primitive. A missing file
is an empty registry. A corrupt or unsupported document refuses startup; it is
never treated as empty and is never rewritten during open.

## Stored shape

The document is `{format: 1, workspaces: [...]}`. Each record carries the
canonical ID, name, location, `active`/`archived` status, decimal-string
revision and decimal millisecond timestamps. Records are emitted in ID order
for deterministic bytes. Unknown fields are rejected.

## Scenarios

- WP01 the persistent adapter reruns WM01–WM10 over fresh isolated files;
- WP02 register, rename, archive and restore state survives reopening;
- WP03 a missing file starts empty and the first committed mutation creates it;
- WP04 corrupt, unsupported, duplicate and structurally invalid documents
  refuse open without changing the original bytes;
- WP05 a failed file write does not change in-memory or on-disk state;
- WP06 equivalent registry state produces deterministic document bytes;
- WP07 root composes the Workspace facade over the persistent adapter.

The adapter does not open a runtime window, subscribe to events or infer a
filesystem location. Root supplies the path and continues to own directory
creation and process composition.
