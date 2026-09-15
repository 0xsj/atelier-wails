# domain

Workspace bounded context. Pure values, invariants and lifecycle transitions
(register, rename, archive, restore, forget). No host, IO, application or
peer-context imports; only shared errors and IDs.

Implemented against [CONTRACT.md](CONTRACT.md) revision 1 (task
workspace-domain); the context boundary is in [../CONTRACT.md](../CONTRACT.md).
Scenario tests live in spec_test.go. The application services and memory
adapter beside this module remain drafts adapted to compile.
