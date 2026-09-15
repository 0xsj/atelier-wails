# Frontend composition root

App.svelte owns adapter selection and injection: it detects the host, builds
the native Preferences transport or the browser preview transport, and hands
them to the gallery together with host-specific title bar drag attributes.
This layer owns app lifetime, provider wiring and feature registration as those
capabilities are introduced. Nothing below root imports it.
