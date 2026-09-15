# desktop

Preferences bounded context. Plain JSON request/response shapes, request
decoding, outcome encoding and public failure projection with commit status.
No native framework import; the host facade forwards to `Handler`.

Implemented against [CONTRACT.md](CONTRACT.md) revision 1 (task
desktop-wire-contract). Scenario tests live in spec_test.go and run over the
memory adapter and the store fault wrappers.
