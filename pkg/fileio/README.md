# fileio

Effectful reusable file primitives: read with absence, atomic durable replace,
ensure directory. Consuming applications own save and conflict rules.

Implemented against [CONTRACT.md](CONTRACT.md) revision 1 (task native-fileio).
The first consumer is the preferences persistent adapter.
