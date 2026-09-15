# commands

`CommandDispatcher` pairs command definitions with handlers and dispatches by id
from the palette, menus and shortcuts through one contract. Dispatch resolves
availability against a context first and returns an outcome value: executed,
unavailable, unknown, no-handler or failed with the captured error. Registry
changes notify subscribers so palettes can re-list. `defaults.ts` registers the
workbench's own commands against a store.
