# codecs

Total readers for unknown wire payloads. `preferences.ts` encodes the five
Preferences requests exactly as the desktop wire contract (revision 1) spells
them and reads every response, event and failure shape without trusting a
cast: ints stay exact as bigint, revisions stay decimal strings, a malformed
failure projects to the internal fallback with the operation's commit state,
and a malformed envelope becomes an `internal` failure typed
`codec.invalid_response`. Tests pin the contract's fixtures.
