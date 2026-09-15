# restoration

Restore a workbench snapshot through an injected `SnapshotPort` and save it
back, debounced. Loading returns an outcome value: restored, empty, invalid,
unsupported or unavailable with the port's error; the model's total reader does
the validation. `createMemorySnapshotPort` backs previews and tests; a native
or file-backed port is a platform adapter chosen by the root. The native JSON
document adapter leaves malformed text for the model's invalid outcome.
