# Atelier domain contracts

Revision 1 contracts are now specified for three independent bounded contexts:
[Preferences](internal/preferences/CONTRACT.md),
[Workspace](internal/workspace/CONTRACT.md) and [Jobs](internal/jobs/CONTRACT.md).

The root composes shared capabilities and may coordinate application APIs. A
domain owns its values, transitions and event facts; its application owns narrow
ports; infrastructure owns memory and later SQLite adapters; desktop transport
owns plain DTO mapping. Peer domains do not import one another. Workspace-scoped
preferences and job workspace ownership use opaque IDs and explicit application
validation.

The contracts deliberately use events as returned facts, not as calls to a global
bus. An application publishes a domain event after its state transaction commits.
An in-process subscriber or UI refresh is enough for the first slice. A durable
SQLite outbox is introduced only for a workflow that needs crash recovery or
cross-process delivery. Provenance identifies the work and execution that caused
an event; it is not the event record or proof of a committed side effect.

## Build order

1. Preferences domain → application port → memory adapter → native boundary.
2. Workspace domain → registry/application port → memory adapter; add filesystem
   inspection only to the concrete open workflow.
3. Jobs domain → application/worker port → memory adapter; use provenance scopes
   for executions and add persistence only when restart behavior is specified.

The first complete vertical slice remains Preferences. Workspace and Jobs are
specified now so their boundaries are visible, but their implementations should
wait for a concrete UI or workflow consumer. No generic repository, event bus,
database package or cross-context transaction manager is implied by these docs.

The Preferences desktop wire shapes, decode vocabulary and commit rule are
specified in [internal/preferences/transport/desktop/CONTRACT.md](internal/preferences/transport/desktop/CONTRACT.md);
the frontend kernel and platform codecs validate against those shapes.

## Shared storage direction

SQLite is the default candidate for local metadata, indexes, jobs and history.
Preferences persist as one atomically replaced JSON document instead; see
[the storage decision](decisions/2026-09-12-preferences-file-persistence.md).
User files stay on the filesystem. Postgres remains a future server adapter for
shared workspaces, remote workers or synchronization; it is not a dependency of
these domain contracts.

See [DOMAIN_GUIDE.md](DOMAIN_GUIDE.md) for adapter scenarios, memory ownership,
persistence evidence and the required order from pure rules to a desktop round
trip.
