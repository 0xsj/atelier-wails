# desktop-wire-contract: wire shapes, decoding and commit uncertainty

Contract: preferences desktop CONTRACT.md, revision 1 (beside the desktop
transport). Dependencies: preferences-app revision 1, preferences-memory
revision 1, native-errors revision 2. State authority: work/manifest.json.

Specify the JSON shapes, failure projection and commit rule, and implement
W01–W12 as the framework-free transport handler. Own the desktop transport
directory, its tests, README and work/handoffs/desktop-wire-contract.md.
Coordinator owns the contract, module registration, the Rust dependency
manifest and lockfile (serde), and manifests. No host facade, root wiring,
native window or frontend codec in this slice; those belong to
preferences-native-slice, which consumes this contract. The frontend kernel
and platform codecs are coordinated through the contract document only.

Checks are listed in the contract's Verification section. Record command
results and actual test counts. Compilation alone is not completion.
