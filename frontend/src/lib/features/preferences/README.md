# preferences

The first product-facing slice on the Preferences round trip. `appearance.ts`
loads theme and density from the store with fallbacks, applies them through
the runtime and writes changes back with compare-and-replace; it is the
consumer that makes the header's theme buttons persist. `PreferencesPanel`
lists, sets and clears global entries and shows every outcome verbatim: kind,
condition type, field problems, commit state and reconcile attempts. Root
supplies the transport; the feature never constructs adapters.
