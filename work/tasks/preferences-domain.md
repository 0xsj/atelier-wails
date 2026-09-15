# preferences-domain: pure preference values and transitions

Contract: preferences domain CONTRACT.md, revision 1 (beside the domain
module). Dependencies: preferences-contract revision 1, native-leaves-integrate,
native-errors revision 2, native-id revision 1. State authority:
work/manifest.json.

Implement P01–P15. Own the domain directory source, tests, README and
work/handoffs/preferences-domain.md. Coordinator owns the contract, module
registration, manifests, lockfiles and the adjacent application/memory drafts,
which are adapted to this API as integration work rather than specified here.
No storage, clock, ID generation, event dispatch, transport or host code in
this slice.

Checks are listed in the contract's Verification section. Record command
results and actual test counts. Compilation alone is not completion.
