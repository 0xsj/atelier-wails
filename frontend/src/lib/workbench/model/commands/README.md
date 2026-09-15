# commands

Pure command registry and keybinding values. A command is an id, title,
optional category, optional keybinding and optional `when` predicate over a
plain context object. Registration returns an outcome (registered, duplicate,
invalid keybinding) rather than throwing. Resolution reports available,
unavailable or unknown. `keybinding.ts` parses chords such as `mod+shift+p`,
matches them against a key-event shape per platform (mod is ⌘ on mac and Ctrl
elsewhere) and formats them for display. Dispatch, scopes and precedence belong
to the app layer; UI actions here are distinct from native application commands.
