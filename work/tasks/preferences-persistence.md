# preferences-persistence: file-backed store and restart evidence

Contract: preferences persistence CONTRACT.md, revision 1 (beside the
adapter). Dependencies: preferences-memory revision 1, native-fileio
revision 1, preferences-native-slice (backend half). Storage decision:
decisions/2026-09-12-preferences-file-persistence.md. State authority:
work/manifest.json.

Implement S01–S07: the adapter over the shared fileio leaf, the reuse of the
M01–M12 suite, and the restart, corruption, write-failure and deterministic
bytes scenarios. Then select the adapter explicitly in root under the
platform config directory and prove composition, restart and corrupt-refusal
through the facade. Own the persistence directory, its tests and README,
root composition and work/handoffs/preferences-persistence.md. Coordinator
owns the contract, module registration and manifests. No migration, backup,
recovery UI or cross-process locking in this slice.

Checks are listed in the contract's Verification section. Record command
results and actual test counts. Compilation alone is not completion.
