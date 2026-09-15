# views

Pure view identity, grouping, activation and open/close state. A view is a
descriptor (id, kind, title) plus instance flags: pinned, preview, dirty. Groups
order views, hold the active view and a most-recently-used history that decides
what becomes active on close. Transitions are total functions returning new
state; impossible ones return the input. A view does not imply a document
domain; the host decides what a `kind` means. Tests live beside the code.
