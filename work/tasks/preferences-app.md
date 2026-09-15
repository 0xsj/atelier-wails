# preferences-app: application services and consumer-owned ports

Contract: preferences application CONTRACT.md, revision 1 (beside the app
module). Dependencies: preferences-domain revision 1, native-errors revision 2.
State authority: work/manifest.json.

Implement A01–A12. Own the command and query modules, their tests, READMEs and
work/handoffs/preferences-app.md. Coordinator owns the contract, module
registration, manifests and the adaptation of the memory adapter draft to the
ports declared here. No adapter contract, transport, host, root wiring or
event bus in this slice.

Checks are listed in the contract's Verification section. Record command
results and actual test counts. Compilation alone is not completion.
