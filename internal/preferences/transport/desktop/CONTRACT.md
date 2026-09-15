# Preferences desktop wire contract

Revision: 1 · State: specified · Task: desktop-wire-contract
Owner: coordinator · Parent: [../../CONTRACT.md](../../CONTRACT.md) revision 1 ·
Application: [../../app/CONTRACT.md](../../app/CONTRACT.md) revision 1 ·
Errors: shared errors revision 2

## Purpose and consumer

The desktop transport is the plain-data boundary between the frontend and
the Preferences application. It owns the JSON shapes of every request,
response, event and failure, decodes requests into domain values, encodes
outcomes, projects failures through the shared public projection, and states
for every failure whether a write was applied. It contains no framework type:
the native facade (Tauri commands, Wails bound methods) is composed in host
and root by the native slice and only forwards to the handler declared here.
The frontend kernel and platform codecs consume these shapes verbatim; this
document is the agreed vocabulary they validate against.

## Owned files and dependencies

Go: `internal/preferences/transport/desktop/`. Rust:
`domains/preferences/transport/desktop.rs` and its `desktop/` directory.
Allowed imports: the standard library, the preferences domain, the
application services, shared errors and shared IDs, plus serde in Rust.
Forbidden: infrastructure, host, root, logger, config, secret, clock, peer
contexts and every native framework crate or package. Tests may construct the
memory adapter and the store fault wrappers.

## Wire shapes

All shapes are JSON objects. Every 64-bit integer travels as a decimal string
so JavaScript never rounds it: an int value is `-?[0-9]+`, a revision is
`[0-9]+`. Absent optional keys are omitted on requests and encoded as `null`
on responses where the shape lists them. Unknown request keys are ignored.

```text
Scope     {"kind":"global"} | {"kind":"workspace","workspace_id":"<uuid>"}
Value     {"kind":"text","text":"…"} | {"kind":"bool","bool":true} | {"kind":"int","int":"-42"}
Expected  {"kind":"absent"} | {"kind":"revision","revision":"3"}
Entry     {"scope":Scope,"key":"editor.theme","value":Value,"revision":"3"}
Event     {"name":"preference.changed","scope":Scope,"key":"…","revision":"3","value":Value|null}

read      {"scope":Scope,"key":"…"}                    → {"found":true,"entry":Entry} | {"found":false,"entry":null}
list      {"scope":Scope}                              → {"entries":[Entry…]}
replace   {"scope":Scope,"key":"…","value":Value,"expected":Expected}
                                                       → {"status":"changed"|"unchanged","entry":Entry,"event":Event|null}
remove    {"scope":Scope,"key":"…","expected":Expected}
                                                       → {"removed":true,"revision":"4","event":Event} | {"removed":false,"revision":null,"event":null}
resolve   {"scope":Scope,"key":"…","fallback":Value}   → {"value":Value,"stored":true,"revision":"3"} | {"value":Value,"stored":false,"revision":null}

Outcome   {"ok":true,"value":<response>} | {"ok":false,"failure":Failure}
Failure   {"kind":"conflict","message":"…","type":"preferences.conflict"|null,"fields":{"<path>":"<problem>"},"commit":"none"|"not_applied"|"unknown"}
```

Logical operation names are `preferences.read`, `preferences.list`,
`preferences.replace`, `preferences.remove` and `preferences.resolve`. Tauri
binds them as commands `preferences_read` … `preferences_resolve` taking one
`request` argument; Wails binds them as methods `Read` … `Resolve` on a bound
`Preferences` struct. Both return the outcome envelope as the call's success
value.

## Decoding and validation

The handler decodes each request before any application call. Structural
problems are refused with `invalid / desktop.invalid_request`, message
`invalid request`, and `fields` mapping a JSON path to exactly one problem
word: `missing` (absent or empty), `unknown` (unrecognized kind) or `invalid`
(malformed content such as a non-decimal number or a malformed workspace ID).
Paths use the request key names: `scope.kind`, `scope.workspace_id`, `key`,
`value.kind`, `value.text`, `value.bool`, `value.int`, `fallback.*`,
`expected.kind`, `expected.revision`. Fields never carry the offending value.

Content that is structurally sound but refused by a domain constructor keeps
the domain's own category and message: `preferences.invalid_key`,
`preferences.invalid_value` (text over 4096 bytes), `preferences.invalid_revision`
(revision `0`). A JSON document that does not decode into the request type
at all is reported as `desktop.invalid_request` with `{"request":"invalid"}`
by the handler's document entry points (`ReadDocument` … `ResolveDocument`;
Rust `read_document` … `resolve_document`), which take the raw document and
dispatch to the typed operation. Native facades pass the document through as
one string argument so this failure never depends on framework decoding.

## Failure projection and commit

Every failure leaving the handler is the shared public projection: kind name,
public message, condition type or `null`, public fields. Internal and
unclassified failures project exactly `internal error` with no type and no
fields, so no diagnostic text, cause or stack ever reaches the wire.

`commit` answers whether the request's write was applied:

| Operation | Failure kind | commit |
| --- | --- | --- |
| read, list, resolve | any | `none` |
| replace, remove | `invalid`, `conflict` | `not_applied` (refused before or without any effect, by domain and application guarantee) |
| replace, remove | any other kind | `unknown` (the store was reached and did not report a refusal; the write may have committed) |

A client that receives `unknown` must not retry blindly. It reconciles by
reading the entry and retrying with the observed revision as `expected`; the
domain's compare-and-replace then either applies once or refuses with
`conflict`. A rejected native promise, which the framework raises for faults
outside the handler, is mapped by the frontend adapter to kind `internal`
with `commit` `unknown` for writes and `none` for reads.

## Acceptance scenarios

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| W01 | Decode scopes: global; workspace with a valid UUID; missing kind; unknown kind; workspace without ID; workspace with a malformed ID | First two round-trip to the same JSON. The rest refuse `desktop.invalid_request` with fields `scope.kind: missing`, `scope.kind: unknown`, `scope.workspace_id: missing`, `scope.workspace_id: invalid` |
| W02 | Decode values: text, bool, int at int64 min and max; int `abc`, `1.5`, `+1`, `` ; missing text; unknown kind; text of 4097 bytes | Sound values round-trip byte-identically. Malformed ints refuse `value.int: invalid`; missing text `value.text: missing`; unknown kind `value.kind: unknown`; the long text keeps `preferences.invalid_value` |
| W03 | Decode expected: absent; revision `3`; revision `0`; revision `x`; missing revision; unknown kind | Absent and `3` round-trip. `0` keeps `preferences.invalid_revision`; `x` refuses `expected.revision: invalid`; missing `expected.revision: missing`; unknown `expected.kind: unknown` |
| W04 | Decode keys: `editor.theme`; `Editor`; missing | Valid key passes; the others keep `preferences.invalid_key` |
| W05 | Read over a memory store: present key; absent key | `{"ok":true,"value":{"found":true,"entry":…}}` with revision as a string; `{"found":false,"entry":null}` |
| W06 | List a scope holding `z.last` and `a.first`; list an empty scope | Entries ordered `a.first`, `z.last`; `{"entries":[]}` |
| W07 | Replace to create, then replace the equal value at revision `1` | `status` `changed` with an event carrying the value and revision `1`; then `unchanged` with `event` `null` |
| W08 | Remove at revision `1`; remove again with expected absent | `removed` true with revision `2` and a `preference.removed` event whose value is `null`; then `removed` false with `revision` and `event` `null` |
| W09 | Resolve a present key; resolve an absent key with fallback `true` | Stored value with `stored` true and its revision; fallback with `stored` false and `revision` `null` |
| W10 | Replace with expected absent on a present key; replace with a `value.int` of `abc` | Both `{"ok":false}`: the first `conflict / preferences.conflict`, the second `invalid / desktop.invalid_request` with `value.int: invalid`; both `commit` `not_applied`; no message or field contains the key or value text |
| W11 | Replace through `FailBefore`; replace through `LoseAcknowledgement`, then read, then replace with expected absent, then with the observed revision | `unavailable` with `commit` `unknown`; `timeout` with `commit` `unknown` while the read shows revision `1`; expected absent refuses `conflict` `not_applied`; the observed revision applies. A read through `FailBefore` reports `commit` `none` |
| W12 | Encode the W07 create outcome and the W10 conflict outcome to JSON | Byte-identical to the fixtures in this document's Fixtures section in both languages |
| W13 | Document entry points: the W07 create request as a raw document; then `not json`, an int given as a JSON number, and a JSON array, through every operation | The create document yields the create fixture; each malformed document yields `invalid / desktop.invalid_request` with `request: invalid`, commit `not_applied` for replace and remove and `none` otherwise (added for preferences-native-slice) |

## Fixtures

Create outcome for global `editor.theme` = text `dark`:

```json
{"ok":true,"value":{"status":"changed","entry":{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"revision":"1"},"event":{"name":"preference.changed","scope":{"kind":"global"},"key":"editor.theme","revision":"1","value":{"kind":"text","text":"dark"}}}}
```

Conflict outcome:

```json
{"ok":false,"failure":{"kind":"conflict","message":"preference revision conflict","type":"preferences.conflict","fields":{},"commit":"not_applied"}}
```

## Verification

Go:

```sh
go test -count=1 -v ./internal/preferences/transport/...
go test -race -count=1 ./internal/preferences/...
go vet ./...
gofmt -l internal/preferences
```

Rust:

```sh
cargo fmt --manifest-path src-tauri/Cargo.toml --check
cargo test --manifest-path src-tauri/Cargo.toml --locked --offline --lib domains::preferences::
cargo check --manifest-path src-tauri/Cargo.toml --locked --offline
```

This proves the handler and the JSON shapes. It does not prove a native
window, a Tauri or Wails binding, or a frontend codec; that is the native
slice's evidence.

## Exclusions and open decisions

No streaming, subscriptions, event push to the frontend, batching,
versioned envelopes, localization of messages, or retry policy is selected.
Whether the first UI exposes workspace scopes is still a UI decision; the
wire accepts them. Event push over the native event channel is deferred to
the native slice or a later contract; events are currently returned inline.

## Completion evidence

See work/handoffs/desktop-wire-contract.md.
