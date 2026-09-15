# panels

`WorkbenchPanel` provides a titled, optionally scrollable tool region.
`PaneStack` stacks several collapsible sections in one sidebar column, like
Explorer, Outline and Timeline: each header toggles its section, expanded
sections share the remaining height by weight, and the divider between two
expanded sections resizes them by pointer or arrow keys. Collapsed ids and
weights are controlled by the host so layout can be restored later.
