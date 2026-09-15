# layout

Pure split, region and sizing rules: sidebar visibility, width and active
activity, collapsed sidebar panes and their weights, the bottom panel, editor
group sizes and status bar visibility. Widths and heights are clamped to limits
and, when window bounds are supplied, to a fraction of the window. No DOM
measurement or native windows.

`snapshot.ts` defines the versioned, serializable workbench snapshot and a total
reader over unknown input that prunes dangling references and reports what it
dropped instead of trusting a cast. Dirty flags and activation history are
runtime facts and are not persisted.
