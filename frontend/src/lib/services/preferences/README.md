# preferences

Read, list, replace, remove and resolve operations over the `PreferencesTransport`
port, which this service owns because it is the only consumer. Requests are
encoded and responses validated by `platform/codecs`; a rejected transport
promise becomes an `internal` failure with commit `unknown` for writes and
`none` for reads, as the wire contract requires. `setPreference` and
`clearPreference` implement compare-and-replace with the reconcile rule: after
an uncertain commit they re-read to see whether the write landed rather than
retrying blindly, and a conflict re-reads and retries while attempts remain.
`model.ts` holds the plain frontend values: scopes, values with exact bigint
ints, decimal-string revisions, entries and events.
