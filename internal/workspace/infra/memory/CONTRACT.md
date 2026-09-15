# Workspace memory adapter contract

Revision: 1 · State: implemented · Task: workspace-memory

The in-memory adapter is the reference implementation of the
Workspace store ports. Every future persistent adapter reruns the same logical
scenarios before adding storage-specific evidence.

## Guarantees

- each `New` store is empty and isolated;
- registration is atomic and rejects duplicate IDs and active locations;
- archived locations may be reused, but restore refuses an active collision;
- conditional lifecycle mutations are atomic and preserve stale revisions;
- read absence and mutation absence follow the application port shape;
- lists are deterministic by Workspace ID and honor `All`/`ActiveOnly`;
- returned values cannot mutate stored state;
- concurrent duplicate registration and conditional rename yield one winner;
- canceled contexts do not begin a store operation.

## Shared scenarios

| ID | Guarantee |
| --- | --- |
| WM01 | fresh store is empty |
| WM02 | register then read/list |
| WM03 | location uniqueness and restore collision |
| WM04 | lifecycle transitions and revision conflicts |
| WM05 | absence and invalid input |
| WM06 | deterministic ordering and filters |
| WM07 | independent instances |
| WM08 | output isolation |
| WM09 | concurrent registrations |
| WM10 | concurrent conditional renames |

## Verification

```text
go test -count=1 ./internal/workspace/infra/memory/...
go test -race -count=1 ./internal/workspace/...
```
