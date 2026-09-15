# platform

Native and preview adapters. Only `desktop/` may import native bindings; it
implements the consumer-owned Preferences and workbench restoration ports for
this host. `preview/` provides the browser doubles that honor the Preferences
and Workspace wire contracts plus workbench restoration.
`codecs/` holds the total readers and writers for the wire shapes. Root
selects between desktop and preview; nothing else constructs adapters.
