# view-host

`ViewHost` renders explicitly registered views through the shared tab primitive
and keeps active view state controlled by its host. `EditorTabs` is the document
tab strip: closable tabs with dirty, pinned and preview states, horizontal
overflow, arrow-key and Delete handling, middle-click close, drag reorder and a
context menu for close others, close to the right, close all and pin. It reports
every intent through callbacks and renders no document content. View
registration and application routing remain outside these UI components.
