# patterns

Reusable compositions built only from components and layout:

- `ChoiceDialog` is a controlled alert dialog with N actions that resolves to
  one action id; Escape and outside dismissal resolve to the cancel id, dismissal
  is ignored while `busy`, the primary action takes focus and loads while busy,
  and an action can declare a `mod+key` shortcut.
- `ConfirmDialog` is the two-outcome case, `confirm` or `cancel`.
- `SaveChangesDialog` is the three-outcome case for closing dirty documents:
  `save`, `discard` or `cancel`, with Don't Save on the left and ⌘D or Ctrl+D as
  its shortcut. It titles itself from the supplied file names and lists them
  when there is more than one.
- `SettingsRow` pairs a label and description with a right-aligned control and
  hands the consumer the ids to reference from `aria-labelledby`.
- `EditableLabel` toggles between a focusable label and an inline text input;
  Enter commits, Escape reverts, and focus returns to the label.
- `MasterDetail` composes a resizable list region with a detail region and an
  explicit empty state when nothing is selected.

Search entry lives in forms as `SearchField`. These patterns own no feature
state, persistence, or service calls; consumers supply values and callbacks.
