# Atelier foundation

Read README.md and FOUNDATION.md. Svelte 5 + TypeScript + Vite is selected.
The directory scaffold is authorized; README-only modules reserve ownership.
The working UI is the component kitchen sink under frontend/src/dev/gallery, which
exercises the live component library and workbench UI slice; new components
belong in the gallery too. The frontend kernel, services, platform adapters and
workbench model/app layers are still reserved. Revision 1 contracts for
Preferences, Workspace and Jobs are specified in DOMAIN_CONTRACTS.md; their
runtime implementations remain reserved. Do not mistake reserved folders for
implementations.

Keep each project independently usable; do not import siblings. Follow the
consumer-owned ports and pure-domain boundaries documented in src/lib/README.md
(frontend/src/lib/README.md for Wails). Native host APIs stay in host/root and
frontend platform/desktop. Preserve lockfiles and record checks actually run.

For backend domain work, follow [DOMAIN_GUIDE.md](DOMAIN_GUIDE.md).

For assigned implementation tasks, follow WORKERS.md and the task
contract revision. The coordinator owns shared wiring and manifest state.
