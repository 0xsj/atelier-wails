# preferences-memory: memory adapter and reusable store scenarios

Contract: preferences memory CONTRACT.md, revision 1 (beside the memory
adapter). Dependencies: preferences-app revision 1, preferences-domain
revision 1. State authority: work/manifest.json.

Implement M01–M12: the adapter, the reusable scenario suite that a persistent
adapter reruns, and the fault wrappers. Own the memory adapter, the test-only
suite, their READMEs and work/handoffs/preferences-memory.md. Coordinator owns
the contract, module registration and manifests. No persistence, transport,
host or root selection in this slice.

Checks are listed in the contract's Verification section. Record command
results and actual test counts, including the concurrency scenario under the
Go race detector. Compilation alone is not completion.
