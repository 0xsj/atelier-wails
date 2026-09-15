# desktop-wire-contract handoff

Contract revision: 1 (`internal/preferences/transport/desktop/CONTRACT.md`).
Specified, implemented and coordinator-integrated on 2026-09-12.

## Scope

The task specified the JSON wire shapes, the decode-failure vocabulary, the
public failure projection and the commit rule, and implemented them as the
framework-free `desktop.Handler`. It did not add a Wails bound struct, root
wiring, a native window or a frontend codec; those consume this contract in
preferences-native-slice. The frontend kernel is coordinated through the
contract document and its fixtures, not through frontend edits.

## Files

`internal/preferences/transport/desktop/dto.go`, `codec.go`, `handler.go`,
`spec_test.go`, `README.md`, `CONTRACT.md` (all new). `DOMAIN_CONTRACTS.md`
gained a pointer to the wire contract. No other production file changed.

## Scenario coverage

W01 `TestW01Scopes` … W12 `TestW12Fixtures`, one test per scenario. W12
compares the create and conflict outcomes byte for byte with the fixtures in
the contract and proves a JSON number in an int field does not decode into
the request type, which is the case the facade must report as
`desktop.invalid_request` with `request: invalid`. W11 runs the handler over
the store fault wrappers and walks the reconciliation path: lost
acknowledgement, read, refused blind retry, applied retry at the observed
revision.

## Verification

`go test -count=1 -v ./internal/preferences/transport/...`: 12 tests passed.
`go test -race -count=1 ./internal/preferences/...`: five packages passed.
`go test -race -count=1 ./...`: all packages passed. `go vet ./...`: passed.
`gofmt -l internal/preferences`: empty.
`python3 tools/work/verify_manifest.py`: valid.

Coordinator source review: production imports are the standard library, the
domain, the application services, shared errors and shared IDs; the memory
adapter and fault wrappers are reached only from tests. Decode failures name
a path and a fixed problem word and never carry the offending value. Outcome
types are concrete per operation so the Wails binding layer sees no generic
type. Implementation-visible verification only.

## Limitations

No native binding, window or frontend evidence exists. Events are returned
inline; push over the native event channel is deferred. The workspace draft
is untouched. Frontend files were not changed.
