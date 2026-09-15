# native-fileio: read with absence, atomic durable replace, ensure directory

Contract: fileio CONTRACT.md, revision 1 (beside the leaf). Dependencies:
native-errors revision 2. State authority: work/manifest.json.

Implement F01–F06. Own the fileio leaf, its tests and README, and
work/handoffs/native-fileio.md. Coordinator owns the contract, module
registration and manifests. No locking, watching, listing, streaming or
application save rules in this slice; the first consumer is the preferences
persistent adapter.

Checks are listed in the contract's Verification section. Record command
results and actual test counts. Compilation alone is not completion.
