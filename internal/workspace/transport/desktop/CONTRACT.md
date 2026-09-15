# Workspace desktop transport contract

Revision: 1 · State: implemented · Task: workspace-native-transport

This framework-free wire boundary mirrors the frontend Workspace codec. Each
native method accepts one JSON object document and returns an `{ok,value}` or
`{ok:false,failure}` envelope. Revisions are decimal strings and timestamps
are UTC RFC3339 millisecond strings.

## Operations

| Operation | Request | Success value |
| --- | --- | --- |
| `Read` | `{id}` | `{found,workspace?}` |
| `List` | `{filter: "active"|"all"}` | `{workspaces}` |
| `Register` | `{id,name,location,at}` | `{workspace,event}` |
| `Rename` | `{id,name,expected,at}` | `{status,workspace,event?}` |
| `Archive`, `Restore` | `{id,expected,at}` | `{status,workspace,event?}` |
| `Forget` | `{id,expected}` | `{found,workspace?}` |

Invalid documents and fields are public `desktop.invalid_request` failures.
Write failures carry `not_applied` for invalid/conflict decisions and
`unknown` for other classified failures; reads carry `none`.

## Verification scenarios

- WT01 request documents decode only from JSON objects;
- WT02 read/list preserve found absence and deterministic metadata;
- WT03 register and lifecycle responses mirror domain events;
- WT04 decimal revisions and UTC millisecond timestamps round-trip;
- WT05 invalid fields expose paths without echoing input values;
- WT06 application failures preserve their public classification and commit;
- WT07 the composed native facade serves the frontend transport shape.

The handler owns no filesystem or runtime-open behavior. Root selects the
process-local Workspace memory adapter until persistence receives its own
contract.
