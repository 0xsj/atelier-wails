# workspace-runtime-open: activate an active Workspace in the Workbench

Dependency: `workspace-persistence`. This is a frontend workflow slice over
the existing native/preview Workspace transport; it does not add a registry
mutation, filesystem inspection or native window command.

The service re-reads the selected Workspace ID before activation. An active
record yields a session containing the authoritative snapshot and a stable
`workspace:<id>` scope. An absent record returns `workspace.not_found`; an
archived record returns `workspace.archived` with `not_applied`. The picker
surfaces failures as values and only hands a successful session to the
Workbench. The Workbench displays the active session context while leaving
window orchestration behind the injected runtime seam.

The sibling Tauri repository carries the same workflow and scenarios. Exact
commands and results belong in `work/handoffs/workspace-runtime-open.md`.
