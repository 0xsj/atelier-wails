# Svelte foundation verification · 2026-09-11

- Svelte 5.57.0, vite-plugin-svelte 7.3.0 and svelte-check 4.7.6 installed with exact versions and matching npm lockfiles.
- `npm run build`: Svelte/TypeScript check reports 0 errors and 0 warnings; Vite production build passes.
- Library roots and manifest/lockfile consistency checked.
- No domain behavior is implemented, and no behavioral test suite is claimed.
- Native windows were inspected through macOS accessibility and displayed Hello world plus their host identity.

- `go vet ./...`: passes.
- `gofmt`: applied to main, root and host startup files.
- `make dev`: native compile, packaging and launch pass; Svelte content verified in Wails.
- Existing linker warning remains: macOS 13.0 object linked for 11.0. Development build also reports private API usage. Release/platform policy remains undecided.

# Component library verification · 2026-09-12

State: the gallery is the working screen and exercises every live component.
Live component groups: primitives (Button, IconButton, Icon, Surface, Separator,
Text, Heading, Label, Kbd), layout (Stack, Inline, Grid, Resizable), forms
(Input, Textarea, SearchField, InputGroup, Field, FieldMessage, Select, Combobox,
Slider, Checkbox, Switch, RadioGroup), feedback, navigation (including Toolbar),
overlays, collections, data display, content utilities, patterns (SettingsRow,
ConfirmDialog, EditableLabel, MasterDetail) and workbench UI (WorkbenchShell,
WorkbenchPanel, ViewHost, CommandPalette, StatusBar, StatusItem).

- Fixed three `svelte-check` errors present at the start of the day: Combobox
  passed `class` to a Bits UI root that renders no element; SearchField's control
  `size` collided with the native numeric `size` attribute; Slider used a
  `Slider.Track` part that Bits UI 2.19.2 does not export, which was also a
  runtime crash in the gallery. The Slider root is now the track container.
- `npm run check`: 638 files, 0 errors, 0 warnings.
- `npm run build`: check passes and the Vite production build succeeds.
- Server-side render smoke: a temporary `vite build --ssr` entry called
  `render()` from `svelte/server` on the gallery. The layout, forms, navigation,
  patterns and workbench sections rendered and every new component class
  appeared in the output. The entry was removed afterwards; it is not a test.

Limitations: no browser or native window was opened for this batch, so keyboard,
focus, dialog and toolbar behavior is unverified at runtime. No frontend unit or
rendered tests exist. The type check does not cover Bits UI runtime API drift, as
the Slider case showed.

# Workbench chrome batch · 2026-09-12

Added `SplitGroup` (layout), `EditorTabs` (view host), `PaneStack` (panels),
`ActivityRail`, `ActivityItem` and `TitleBar` (shell). `WorkbenchShell` gained a
`titlebar` snippet and `windowActive`. The gallery's workbench demo now uses all
of them plus an "Editor tabs" section covering pinned, dirty, preview, locked
and overflowing tabs. The title bar takes drag-region attributes from the root
so the component never names a native host; the root composition supplies them.

- `npm run check`: 645 files, 0 errors, 0 warnings.
- `npm run build`: passes.
- Server-side render smoke as above: layout, editor-tabs and workbench sections
  rendered; split handles, tab strip, pane stack, activity badges, title bar and
  the injected drag attributes all appear in the output.

Limitations unchanged: pointer drag, keyboard resizing, tab reorder, context
menus and window-drag behavior are unverified in a browser or native window.
`ViewHost` is no longer exercised by the gallery since the shell demo moved to
editor tabs.

# Dialogs and quick pick · 2026-09-12

Added `ChoiceDialog` (N actions, one resolved id, cancel on dismissal, busy
guard, `mod+key` shortcuts), rebuilt `ConfirmDialog` on it, and added
`SaveChangesDialog` with `save`, `discard` and `cancel` outcomes. Added
`QuickPick` (token filtering, ordered groups, recents, prefix modes, row
actions, combobox and listbox semantics) and rebuilt `CommandPalette` on it.
The gallery shows both dialogs in the patterns section, a three-mode quick pick
in the commands section, and closing a dirty editor tab in the workbench demo
now asks through `SaveChangesDialog`; quick-open adds tabs to either group.

- `npm run check`: 649 files, 0 errors, 0 warnings.
- `npm run build`: passes.
- Server-side render smoke: patterns, commands and workbench sections rendered
  with their trigger buttons present. Closed Bits UI dialogs emit no markup, so
  the dialog and picker bodies themselves are not covered by this smoke.

Limitations unchanged: dialog focus, shortcuts and picker keyboard behavior are
unverified in a browser or native window.

# Workbench pure models · 2026-09-12

First behavioral frontend code and first frontend test suite. Under
`workbench/model`: `views` (groups, activation, MRU close fallback, preview,
pinned zones, reorder, move, add and remove groups), `layout` (sidebar, panes,
panel, editor sizes, clamping to bounds), `snapshot` (versioned snapshot with a
total reader over unknown input that prunes and reports dangling references),
`commands` (registry with outcome values, availability from context, palette
entries) and `keybinding` (parse, per-platform match, format). No Svelte, DOM or
native imports; grep-checked.

- vitest 5.0.0 added as an exact devDependency; `npm test` runs `vitest run`
  over `src/**/*.test.ts` in a node environment.
- `npm test`: 5 files, 31 tests pass.
- `npm run check`: 683 files, 0 errors, 0 warnings.

Not yet covered: the `app` layer (dispatch, keybinding scopes, restoration
port) and wiring the models into the gallery shell.

# Workbench app layer and store-driven gallery · 2026-09-12

Under `workbench/app`: `CommandDispatcher` (handlers beside definitions,
outcome values, change notification), keybinding scopes (text inputs keep plain
keys, modifier chords and escape/function keys reach the workbench, composition
never matches), navigation history (back/forward with pruning on close),
restoration (`SnapshotPort`, in-memory port, debounced saver with injected
scheduler) and `WorkbenchStore` (store contract without Svelte, dirty-close
policy returning needs-confirmation, context for command availability, snapshot
and restore). `registerWorkbenchCommands` adds toggle sidebar, toggle panel,
close editor, close others, pin, split right, back and forward with default
chords. The gallery's workbench section is now `WorkbenchDemo`, which seeds the
store, registers demo commands (save, edit, settings, quick open, palette,
snapshot save/restore, reset), routes key events through `resolveShortcut`, and
closes dirty views through `SaveChangesDialog`. `PaneStack` gained
`onWeightsChange`.

- `npm test`: 10 files, 50 tests pass.
- `npm run check`: 696 files, 0 errors, 0 warnings.
- `npm run build`: passes.
- Server-side render smoke: the workbench section rendered through the store
  with two editor groups, tabs, sidebar panes and the status bar present.

Limitations: shortcut handling in the demo is scoped to the shell element, so
focus must be inside it; pointer drag, dialogs and the palette remain
unverified in a browser or native window. No native snapshot port exists yet.

# Preferences round trip, frontend half · 2026-09-12

Implements the frontend half of preferences-native-slice against the desktop
wire contract revision 1 and errors revision 2.

- `kernel`: `result.ts`, `failure.ts` (ten exact kinds, public shape, commit
  state, internal fallback), `presence.ts`, `identity.ts`.
- `platform/codecs/preferences.ts`: request encoding and total response
  readers; ints as bigint, revisions as decimal strings; the contract's create
  and conflict fixtures are pinned by tests.
- `services/preferences`: the five operations over the `PreferencesTransport`
  port, rejected promises mapped to internal with commit by operation,
  `setPreference` and `clearPreference` with the reconcile-after-unknown rule.
- `platform/preview/preferences.ts`: in-memory transport honoring W01–W13
  decoding, compare-and-replace, envelope, projection and commit rules, with
  injectable W11 faults; tests pin fixtures byte for byte.
- `platform/desktop`: the Wails adapter and host detection for this
  repository only; the two frontends differ in `platform/desktop` and `root`.
- `features/preferences`: appearance sync (theme and density stored as
  `ui.theme` and `ui.density`) and `PreferencesPanel`; the gallery gained a
  Preferences section and a header note reporting where appearance came from.

- `npm test`: 14 files, 73 tests pass.
- `npm run check`: 0 errors, 0 warnings.
- `npm run build`: passes.
- Server-side render smoke: the preferences section renders with the preview
  transport injected.

Native round trip, 2026-09-12: screen capture and assistive access were both
denied to the session, so the round trip was proven on disk instead. A probe
build (`VITE_ATELIER_PROBE=preferences make build`) makes the gallery write one preference at startup through
the real adapter and read it back (`src/dev/probe.ts`, compiled away without
the flag). Launching the built app produced `~/Library/Application Support/atelier-wails/preferences.json`:

```json
{"format":1,"entries":[{"scope":{"kind":"global"},"key":"probe.launch","value":{"kind":"text","text":"Wails / Go 2026-09-12T13:24:41.111Z"},"revision":"1"}]}
```

That document is the file-backed store's format 1, written by the native
handler after decoding the request the frontend codec encoded. The probe key
was deleted afterwards and the apps rebuilt without the flag. The visible UI
was not inspected; keyboard and dialog behavior remain unverified in a window.

# Workbench file-backed snapshot port · 2026-09-14

The workbench restoration boundary now has a JSON document adapter. Browser
previews keep an in-memory port; Tauri and Wails roots inject native adapters
whose hosts read and atomically replace `workbench.json` beside the Preferences
document. The shared adapter preserves malformed JSON as text so the existing
total snapshot reader reports `invalid`, while native read/write failures remain
`unavailable` port outcomes.

- Wails `make check`: frontend check/test/build and `go vet ./...` passed;
  `go test -count=1 ./...` also passed.
- The Go host facade has focused missing-file, round-trip and replacement tests.
- The shared frontend suite: 14 files, 74 tests passed; type check: 0 errors,
  0 warnings; production build: passed.

The Tauri implementation and its native evidence are recorded in the sibling
repository's matching handoff. The Wails probe bundle passed: the launched app
restored `probe.snapshot` through the real generated binding and wrote
`workbench.json`, which was read from the config directory and moved to
`/private/tmp/atelier-workbench-probe-wails.json`. Visual interaction was not
inspected; screen capture and assistive access remain unavailable to the
terminal session.

# Rendered component contracts · 2026-09-14

Vitest browser mode is now configured with the pinned Playwright provider and
Chromium runtime. `npm --prefix frontend run test:browser` passed in both
independent builds: one browser file with 3 tests in each. The tests mount
real Svelte components and verify Button DOM attributes plus click dispatch,
WorkbenchShell regions and landmarks, and ActivityRail selection, focus and
ArrowUp keyboard behavior.

The browser suite is intentionally representative rather than a claim that
every component state has visual coverage. Screen capture and native-window
interaction remain unavailable to this session.

# Workspace registry picker · 2026-09-14

Added the shared Workspace registry frontend slice. The kernel rejects zero
Workspace IDs; the JSON codecs validate canonical locations, lifecycle status,
UTC timestamps and decimal revisions; the service preserves Found versus
Absent reads; and the preview adapter covers register, rename, archive,
restore, forget, active-location uniqueness, conditional revisions and
revision exhaustion. `WorkspacePicker` now appears in the gallery with
active/all filtering, metadata inspection and an explicit Open callback. The
native Workspace application and transport remain reserved pending backend
review.

- `make check`: frontend check/test/build and `go vet ./...` passed.
- `npm --prefix frontend run test:browser`: 1 file, 4 tests passed in
  Chromium.
- `make build`: passed; the app was packaged and self-signed, with the known
  macOS deployment-target linker warning.
- `frontend/package.json.md5` matches the current package manifest.
- Screen capture and native-window interaction remain unavailable.
