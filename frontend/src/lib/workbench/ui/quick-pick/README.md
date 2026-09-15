# quick-pick

`QuickPick` is the generic picker behind the command palette, file switcher and
similar prompts. It filters supplied items by every query token, groups them
with optional ordered headers and a recents group, supports prefix modes such as
`>` for commands and `@` for symbols (typed or chosen from chips, Backspace on
an empty query returns to the default mode), keeps a selected row with arrow,
Home and End keys, chooses on Enter or click, and exposes per-row secondary
actions. It reports the chosen item and mode through callbacks; the host owns
the item source, recents and what selection does.
