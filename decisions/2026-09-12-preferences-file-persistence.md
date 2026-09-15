# Preferences persist as one atomically replaced JSON document

Accepted by the coordinator, 2026-09-12, while completing the preferences
slice. Preferences are a small set of key/value entries per installation;
they are stored as one JSON document per store, written through the shared
fileio leaf's temp-write, fsync, rename and directory-sync sequence, and
loaded once at startup. Every change rewrites the whole document; the file is
always a previous or a new complete document.

SQLite remains the default candidate for multi-record local metadata, jobs
and history, as DOMAIN_CONTRACTS.md states. Choosing a file here avoids a
database dependency, schema and migration machinery for a store that never
needs indexes or partial updates, and it lets the planned fileio leaf earn its
first consumer. The persistent adapter passes the same M01–M12 suite as the
memory adapter, so a later SQLite adapter can replace it behind the same
ports with the same evidence.

Tradeoffs accepted: one process owns the file; concurrent processes are not
coordinated. A corrupt document refuses startup with a logged failure rather
than starting empty, because silently losing user preferences is worse than
a visible failure; a recovery flow is future work. Format 1 is the first;
migration policy is decided when a second format exists.
