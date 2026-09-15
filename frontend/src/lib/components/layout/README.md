# layout

Layout components own bounds, spacing, and overflow; they carry no feature
state. `Stack`, `Inline`, and `Grid` place children on the spacing scale with
token-driven gaps and alignment. `Resizable` is a controlled two-pane split.
`SplitGroup` generalizes it to N panes in one direction with pointer and
arrow-key resizing, minimum sizes, and optional collapse to zero with restore
on Enter or double-click; nest one inside a pane for mixed splits. Scroll
containment lives with `ScrollArea` in collections.
