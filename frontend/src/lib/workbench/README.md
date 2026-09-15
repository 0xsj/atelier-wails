# workbench

Reusable frontend workbench composition. Depends on components and runtime state;
never imports features. `model/` holds the pure workbench models: views and
groups, layout and the versioned snapshot, commands and keybindings. `app/`
binds them: `WorkbenchStore` (store contract, no Svelte import) applies model
transitions, owns the dirty-close policy, dispatches commands and shortcuts and
produces snapshots; `CommandDispatcher`, keybinding scopes, navigation history
and restoration through a `SnapshotPort` live beside it. `ui/` provides
`WorkbenchShell`, `TitleBar`, `ActivityRail`, `PaneStack`, `EditorTabs`,
`ViewHost`, `QuickPick`, `CommandPalette` and `StatusBar`. The gallery's
`WorkbenchDemo` shows the UI running on the store. Root-level registration of
features remain a root concern; root injects memory restoration for previews and
file-backed restoration for desktop hosts.
