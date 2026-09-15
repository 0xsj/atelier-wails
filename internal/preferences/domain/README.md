# domain

Preferences bounded context. Pure values, invariants and compare-and-replace
transitions. No host, IO, application or peer-context imports; only shared
errors and IDs.

Implemented against [CONTRACT.md](CONTRACT.md) revision 1 (task
preferences-domain); the context boundary is in [../CONTRACT.md](../CONTRACT.md).
Scenario tests live in spec_test.go. The application services and memory
adapter beside this module remain drafts until their own tasks are specified.
