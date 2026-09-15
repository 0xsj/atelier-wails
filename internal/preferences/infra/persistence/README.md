# persistence

Preferences bounded context. One JSON document per store, replaced atomically
through the shared fileio leaf; persistence-owned records and format version.

Implemented against [CONTRACT.md](CONTRACT.md) revision 1 (task
preferences-persistence). Passes the shared store suite M01–M12 and adds
restart, atomic-commit and corruption scenarios S02–S06. Root selects this
adapter by default; the memory adapter remains for tests and ephemeral use.
