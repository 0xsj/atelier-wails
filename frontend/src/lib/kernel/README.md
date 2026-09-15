# kernel

Pure TypeScript leaves with no Svelte, platform, service or IO imports.

- `result.ts`: Ok and Err as values with map, mapError, andThen, match and all.
- `failure.ts`: the closed failure vocabulary from the native errors contract,
  the ten exact lowercase kinds, the public shape (kind, message, type,
  fields) and the wire contract's commit state; the internal fallback
  projection is a fixed value.
- `presence.ts`: found and absent as a value so a lookup is
  `Result<Presence<T>, Failure>` and never conflates absence with failure.
- `identity.ts`: a branded workspace id parsed from canonical UUID text.
