# runtime

Portable interaction state plus explicit Svelte bindings. Theme and density
are applied as root data attributes; `applyTheme` and `applyDensity` are the
only functions the preferences feature needs. The localStorage-backed
`hydrate`/`set` helpers remain for a frontend-only fallback but the gallery
now persists appearance through the Preferences store. Keep future
subscriptions scoped and disposable.
