# gallery

The working development kitchen sink. It exposes the token layer, control
states, theme/density behavior and every live component. `WorkbenchDemo` runs
the workbench UI on the real `WorkbenchStore` with the snapshot port injected
by root: browser previews use memory, while desktop hosts use the file-backed
document. The Workspace section runs the registry picker against its browser
preview transport until the native application boundary lands. The shell demo
therefore exercises models, commands, shortcuts and restoration rather than
local state. Keep this view frontend-only and use it to exercise real
components as they arrive.
