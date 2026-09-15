# File IO contract

Revision: 1 · State: specified · Task: native-fileio
Owner: coordinator · Errors: shared errors revision 2

## Purpose and consumer

The first filesystem operation the foundation needs is "replace one file's
whole content atomically and durably", paired with "read one file, telling
absence apart from failure" and "make sure a directory exists". The first
consumer is the preferences persistent adapter, which stores one document per
store. Consuming applications own save and conflict rules; this leaf owns the
mechanics only.

## Owned files and dependencies

Go: `pkg/fileio/`. Rust: `src-tauri/src/shared/fileio/`. Allowed imports: the
standard library and shared errors. Forbidden: every context, host, root,
logger, config and secret.

## Public API

Go:

```go
func Read(path string) (data []byte, found bool, err error)
func Replace(path string, data []byte) error
func EnsureDir(path string) error
```

Rust:

```rust
pub(crate) fn read(path: &Path) -> Result<Option<Vec<u8>>, Failure>
pub(crate) fn replace(path: &Path, data: &[u8]) -> Result<(), Failure>
pub(crate) fn ensure_dir(path: &Path) -> Result<(), Failure>
```

`Replace` writes the content to a uniquely named temporary file in the same
directory, flushes it to stable storage, renames it over the target, and then
flushes the directory on platforms that support directory sync. The target is
therefore either the previous complete content or the new complete content;
a reader never observes a partial file under the target name. The parent
directory must exist; `Replace` does not create it. Files are created with
owner-only permissions (`0600`), directories with `0700`; existing
permissions are not changed.

## Observable contract

| Kind / type | Meaning |
| --- | --- |
| `unavailable / fileio.read_failed` | The path exists but cannot be read (a directory, permission denied, IO error) |
| `unavailable / fileio.write_failed` | The temporary file could not be created, written, flushed or renamed; the target is unchanged and the temporary file is removed on a best-effort basis |
| `unavailable / fileio.dir_failed` | The directory could not be created, or the path exists and is not a directory |

Absence is not a failure: `Read` of a missing path returns not found. Every
failure keeps the operating system error as a private cause and the path as
a private detail; the public message is fixed and never contains the path.
Nothing here retries, locks across processes or watches files.

## Acceptance scenarios

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| F01 | Read a missing path; read a directory path | Missing: not found, no failure. Directory: `fileio.read_failed` |
| F02 | Replace a new path with bytes, then read; replace with empty content | Read returns exactly the written bytes, including the empty case; the file mode is owner-only |
| F03 | Replace an existing file twice with different content | Each read returns the latest content; no temporary file remains in the directory |
| F04 | Replace into a directory that does not exist; replace into a directory without write permission that already holds the target | Both refuse with `fileio.write_failed`; the existing target still holds its previous content; the directory holds no temporary file |
| F05 | EnsureDir on a nested missing path, twice; EnsureDir on a path that is a file | Directory exists after the first call; the second call is a no-op success; the file path refuses with `fileio.dir_failed` |
| F06 | Every failure above | Kind unavailable, the stated type, a private cause and detail; the public message excludes the path |

## Verification

Go:

```sh
go test -count=1 -v ./pkg/fileio
go test -race -count=1 ./pkg/fileio
go vet ./...
gofmt -l pkg/fileio
```

Rust:

```sh
cargo fmt --manifest-path src-tauri/Cargo.toml --check
cargo test --manifest-path src-tauri/Cargo.toml --locked --offline --lib shared::fileio::
cargo check --manifest-path src-tauri/Cargo.toml --locked --offline
```

Tests use fresh temporary directories and run on the local platform only.
Crash-consistency across power loss is a property of the mechanism and the
filesystem, not something these tests can prove.

## Exclusions and open decisions

No streaming, partial writes, append, locking, watching, listing, copying,
removal, symlink policy or cross-process coordination. Windows directory sync
is skipped; rename-over-existing relies on the platform's replace semantics.

## Completion evidence

See work/handoffs/native-fileio.md.
