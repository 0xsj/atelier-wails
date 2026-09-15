# keybindings

Scope and precedence for shortcuts. A key event targeted at a text input,
textarea, select or contenteditable is in the `text-input` scope: only chords
with a command, control or alt modifier and the escape and function keys reach
the workbench; plain keys and shift combinations stay with the field.
Composition input is never a shortcut. Resolution returns match, none,
composing or reserved-for-text so the host can decide whether to prevent the
default. Conflict policy is first match by category and title.
