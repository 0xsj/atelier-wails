# workspace-domain: pure workspace values and lifecycle transitions

Contract: workspace domain CONTRACT.md, revision 1 (beside the domain
module). Dependencies: workspace-contract revision 1, native-leaves-integrate,
native-errors revision 2, native-id revision 1. State authority:
work/manifest.json.

Implement WD01–WD12. Own the domain directory source, tests, README and
work/handoffs/workspace-domain.md. Coordinator owns the contract, module
registration, manifests and the adaptation of the application and memory
drafts to this API, which is integration work rather than a specified task.
No storage, clock, ID generation, filesystem inspection, uniqueness rule,
transport or host code in this slice.

Checks are listed in the contract's Verification section. Record command
results and actual test counts. Compilation alone is not completion.
