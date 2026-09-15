# Architecture import rules

Revision: 1 · State: specified · Task: architecture-checks
Owner: coordinator

## Purpose and consumer

`check_imports.py` enforces the layer graph that DOMAIN_GUIDE.md and the
frontend library README describe in prose. Until now every handoff claimed
its import boundaries by review; this check makes the claim executable for
the native tree. The frontend tree is out of scope for this revision.

## Layers

Every native source file belongs to exactly one layer, derived from its
path. Go paths are relative to the module root; Rust paths are relative to
`src-tauri/src`.

| Layer | Go | Rust |
| --- | --- | --- |
| leaf `<name>` | `pkg/<name>/` | `shared/<name>/` |
| context `<ctx>` layer `<layer>` | `internal/<ctx>/<layer>/` for `domain`, `app`, `infra`, `transport` | `domains/<ctx>/<layer>` (`mod.rs`, `<layer>.rs` or `<layer>/…`) |
| context root | `internal/<ctx>/` itself | `domains/<ctx>/mod.rs`, `domains/<ctx>/tests.rs` |
| host | `internal/host/` | `host/` |
| root | `root/`, `main.go` | `root/`, `lib.rs`, `main.rs` |

`internal/host` is host, never a context. Files outside these paths (tools,
generated code) are ignored. A file is a test file when its name ends in
`_test.go`, or in Rust is `tests.rs`, ends in `_test.rs` or `_tests.rs`, or
is an inline `#[cfg(test)] mod … { … }` block, which the checker skips.

## Rules

Only imports inside the module are checked; standard library and third-party
imports are ignored.

| Source | May import | Must not import |
| --- | --- | --- |
| leaf | other leaves | any context, host, root |
| domain | leaves `errors`, `id`; its own domain | any other leaf, any other layer, peers, host, root |
| app | leaves `errors`, `id`; its own domain and app | infra, transport, other leaves, peers, host, root |
| infra | leaves `errors`, `id`, `clock`, `fileio`; its own domain, app and infra | transport, other leaves, peers, host, root; and `persistence` must not import `memory` |
| transport | leaves `errors`, `id`; its own domain, app and transport | infra, other leaves, peers, host, root |
| context root | anything in its own context; leaves `errors`, `id` | peers, host, root |
| host | any leaf; any context's app and transport; host | any domain or infra, root |
| root | anything | nothing is forbidden |

Test files of a context layer may additionally import their own context's
infra, so application and transport tests can run over the memory adapter
and the fault wrappers. Nothing else changes for tests.

## Acceptance scenarios

| ID | Given / when | Required observable outcome |
| --- | --- | --- |
| C01 | The real repository tree | Zero violations; the summary names the file and import counts |
| C02 | Fixtures: a domain importing infra; an app importing a peer context; a leaf importing a context; a transport importing infra; a host importing a domain; anything importing root | One violation each, naming the file, the import and the rule |
| C03 | Fixtures: a test file importing its own infra; an infra importing its own app; a host importing app and transport; root importing everything | No violations |
| C04 | Rust fixtures: `super::super::infra` from an app file resolves and is refused; an inline `#[cfg(test)] mod` block importing infra in an app file is skipped; `persistence` importing `memory` is refused; a bare `crate::root::…` path outside a `use` is refused | Exactly those outcomes |
| C05 | Both fixture flavors | The script detects the flavor from `go.mod` or `src-tauri/Cargo.toml` and applies the same rule table |

## Verification

```sh
python3 tools/architecture/check_imports.py
python3 -m unittest discover -s tools/architecture -p 'test_*.py'
```

The check is a source-text analysis. It does not compile, does not follow
re-exports, and does not see conditional compilation beyond the inline
`#[cfg(test)]` skip. It complements the manifest verifier, which checks task
metadata, and does not replace coordinator review of behavior.

## Exclusions and open decisions

No frontend rules, no cyclic-dependency detection inside a layer, no
enforcement of the "no adapter constructed inside a command" rule, and no
CI wiring. Those are separate decisions.
