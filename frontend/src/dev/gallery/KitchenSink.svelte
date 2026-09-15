<script lang="ts">
	import { onMount } from 'svelte';
	import { List, ScrollArea, Tree } from '../../lib/components/collections';
	import { Avatar, CodeBlock, CopyButton, InlineCode, Tag } from '../../lib/components/content';
	import { Card, KeyValue, Table, type TableColumn, type TableRow } from '../../lib/components/data-display';
	import { Checkbox, Combobox, Field, FieldMessage, Input, InputGroup, RadioGroup, SearchField, Select, Slider, Switch, Textarea } from '../../lib/components/forms';
	import { Alert, Badge, Banner, EmptyState, Progress, Skeleton, Spinner, Status, Toaster, type ToastItem } from '../../lib/components/feedback';
	import { Accordion, Breadcrumb, Collapsible, ContextMenu, DropdownMenu, MenuItem, MenuSeparator, TabPanel, Tabs, Toggle, ToggleGroup, Toolbar, ToolbarButton, ToolbarGroup, type AccordionItem, type BreadcrumbItem } from '../../lib/components/navigation';
	import { Dialog, Popover, Tooltip } from '../../lib/components/overlays';
	import { Button, Heading, Icon, IconButton, Kbd, Label, Separator, Surface, Text } from '../../lib/components/primitives';
	import { Grid, Inline, Resizable, SplitGroup, Stack } from '../../lib/components/layout';
	import { ConfirmDialog, EditableLabel, MasterDetail, SaveChangesDialog, SettingsRow, type ConfirmOutcome, type SaveChangesOutcome } from '../../lib/components/patterns';
	import { CommandPalette, type CommandItem } from '../../lib/workbench/ui/command-palette';
	import { QuickPick, type QuickPickAction, type QuickPickItem, type QuickPickMode } from '../../lib/workbench/ui/quick-pick';
	import { EditorTabs, type EditorTabItem } from '../../lib/workbench/ui/view-host';
	import WorkbenchDemo from './WorkbenchDemo.svelte';
	import { createAppearanceSync, PreferencesPanel } from '../../lib/features/preferences';
	import { WorkspacePicker } from '../../lib/features/workspace';
	import { runLaunchProbe, runSnapshotProbe } from '../probe';
	import type { PreferencesTransport } from '../../lib/services/preferences';
	import type { WorkspaceOpenPort, WorkspaceSession, WorkspaceTransport } from '../../lib/services/workspace';
	import type { PreviewPreferencesTransport } from '../../lib/platform/preview/preferences';
	import type { SnapshotPort } from '../../lib/workbench/app';
	import { DENSITIES, THEMES, type Density, type Theme } from '../../lib/runtime';
	import './kitchen-sink.css';

	interface PreferencesWiring {
		transport: PreferencesTransport;
		preview: PreviewPreferencesTransport | null;
		storeLabel: string;
	}
	interface WorkbenchWiring {
		snapshotPort: SnapshotPort;
	}
	interface WorkspaceWiring {
		transport: WorkspaceTransport;
		openPort: WorkspaceOpenPort;
		storeLabel: string;
	}
	let { host, dragRegion = {}, dragExclude = {}, preferences, workbench, workspace }: { host: string; dragRegion?: Record<string, string>; dragExclude?: Record<string, string>; preferences: PreferencesWiring; workbench: WorkbenchWiring; workspace: WorkspaceWiring } = $props();
	// svelte-ignore state_referenced_locally
	const appearance = createAppearanceSync(preferences.transport);
	let appearanceNote = $state('Loading appearance from the store…');
	let workspaceNote = $state('No workspace opened yet');
	let activeSession = $state<WorkspaceSession | null>(null);

	function handleWorkspaceOpen(session: WorkspaceSession): void {
		activeSession = session;
		workspaceNote = `Opened ${session.workspace.name} · ${session.scope}`;
	}
	let theme = $state<Theme>('system');
	let density = $state<Density>('comfortable');
	let formChecked = $state(true);
	let formPinned = $state(false);
	let formChoice = $state('canvas');
	let demoTab = $state('overview');
	let paletteOpen = $state(false);
	let lastCommand = $state('No command selected');
	let menuRadio = $state('system');
	let toggleValue = $state('grid');
	let compactLabels = $state(false);
	let lastMenuAction = $state('No menu action selected');
	let listSelection = $state('tokens');
	let treeSelection = $state('atelier');
	let expandedTreeIds = $state<string[]>(['atelier']);
	let comboboxValue = $state('tokens');
	let searchValue = $state('');
	let sliderValue = $state(68);
	let inputGroupValue = $state('atelier-core');
	let alertVisible = $state(true);
	let bannerVisible = $state(true);
	let feedbackToasts = $state<ToastItem[]>([
		{ id: 'indexing', title: 'Workspace indexed', description: '12 files are ready to inspect.', tone: 'success' }
	]);
	let toastSequence = $state(1);
	const inspectorColumns: readonly TableColumn[] = [
		{ key: 'name', label: 'View' },
		{ key: 'owner', label: 'Owner' },
		{ key: 'updated', label: 'Updated', align: 'end' },
		{ key: 'status', label: 'Status' }
	];
	const inspectorRows: readonly TableRow[] = [
		{ id: 'overview', name: 'Overview', owner: 'You', updated: '2m ago', status: 'Ready' },
		{ id: 'tokens', name: 'Token inspector', owner: 'You', updated: '18m ago', status: 'Draft' },
		{ id: 'preview', name: 'Preview', owner: 'Team', updated: 'Yesterday', status: 'Needs review' }
	];
	const workspaceProperties = [
		{ label: 'Workspace', value: 'atelier-core', description: 'Local project root' },
		{ label: 'Branch', value: 'main', description: 'No pending changes' },
		{ label: 'Views', value: 12, description: '3 pinned views' },
		{ label: 'Last indexed', value: '08:42:19', description: 'All files up to date' }
	] as const;
	const tokenSnippet = `:root {
	--accent: oklch(72% 0.17 68);
	--surface-panel: oklch(21% 0.02 260);
}`;
	let copyFeedback = $state('Nothing copied yet');
	let fieldName = $state('');
	let toolbarAlignment = $state('left');
	let toolbarMarks = $state<string[]>(['grid']);
	let lastToolbarAction = $state('No toolbar action yet');
	let confirmOpen = $state(false);
	let confirmBusy = $state(false);
	let confirmOutcome = $state('No decision yet');
	let workspaceTitle = $state('Atelier core');
	let masterSelection = $state('overview');
	let settingsAutosave = $state(true);
	let settingsTelemetry = $state(false);
	let splitSizes = $state([25, 50, 25]);
	let nestedSizes = $state([60, 40]);
	let stripTabs = $state<EditorTabItem[]>([
		{ id: 'a', label: 'App.svelte', pinned: true },
		{ id: 'b', label: 'Very-long-file-name-that-overflows-the-tab.svelte', description: 'A long label truncates and shows the full path on hover' },
		{ id: 'c', label: 'index.ts', dirty: true },
		{ id: 'd', label: 'layout.css', preview: true },
		{ id: 'e', label: 'settings.json', closable: false },
		{ id: 'f', label: 'Tree.svelte' },
		{ id: 'g', label: 'List.svelte' },
		{ id: 'h', label: 'Table.svelte' },
		{ id: 'i', label: 'Card.svelte' },
		{ id: 'j', label: 'Toast.svelte' }
	]);
	let stripTab = $state('c');
	let lastTabEvent = $state('No tab event yet');
	let saveOpen = $state(false);
	let saveBusy = $state(false);
	let saveItems = $state<string[]>([]);
	let saveOutcome = $state('No decision yet');
	let quickOpen = $state(false);
	let quickMode = $state('');
	let quickQuery = $state('');
	let quickLast = $state('Nothing picked yet');
	let quickRecents = $state<string[]>(['file-shell', 'file-tokens']);
	const quickModes: readonly QuickPickMode[] = [
		{ prefix: '', label: 'Files', placeholder: 'Search files by name' },
		{ prefix: '>', label: 'Commands', placeholder: 'Run a command' },
		{ prefix: '@', label: 'Symbols', placeholder: 'Go to a symbol in the active file' }
	];
	const quickFileActions: readonly QuickPickAction[] = [{ id: 'side', label: 'Open to the side', glyph: '⫿' }, { id: 'forget', label: 'Remove from recently used', glyph: '×' }];
	const quickFiles: readonly { id: string; label: string; path: string }[] = [
		{ id: 'file-shell', label: 'WorkbenchShell.svelte', path: 'src/lib/workbench/ui/shell' },
		{ id: 'file-tokens', label: 'semantic.css', path: 'src/lib/styles/tokens' },
		{ id: 'file-tree', label: 'Tree.svelte', path: 'src/lib/components/collections' },
		{ id: 'file-list', label: 'List.svelte', path: 'src/lib/components/collections' },
		{ id: 'file-quick', label: 'QuickPick.svelte', path: 'src/lib/workbench/ui/quick-pick' },
		{ id: 'file-readme', label: 'README.md', path: '.' },
		{ id: 'file-foundation', label: 'FOUNDATION.md', path: '.' }
	];
	const quickSymbols: readonly QuickPickItem[] = [
		{ id: 'sym-props', label: 'Props', description: 'interface', group: 'Types' },
		{ id: 'sym-item', label: 'QuickPickItem', description: 'type', group: 'Types' },
		{ id: 'sym-select', label: 'select', description: '(position: number) => void', group: 'Functions' },
		{ id: 'sym-keydown', label: 'handleKeydown', description: '(event: KeyboardEvent) => void', group: 'Functions' },
		{ id: 'sym-groups', label: 'groups', description: 'derived', group: 'State' }
	];
	const quickItems = $derived.by<readonly QuickPickItem[]>(() => {
		if (quickMode === '>') return paletteCommands.map((command) => ({ ...command, group: command.id === 'toggle-sidebar' ? 'View' : 'Navigation' }));
		if (quickMode === '@') return quickSymbols;
		return quickFiles.map((file) => ({ id: file.id, label: file.label, description: file.path, recent: quickRecents.includes(file.id), actions: quickFileActions }));
	});
	const toolbarAlignments = [{ value: 'left', label: 'Left' }, { value: 'center', label: 'Center' }, { value: 'right', label: 'Right' }] as const;
	const toolbarMarkOptions = [{ value: 'grid', label: 'Grid' }, { value: 'rulers', label: 'Rulers' }, { value: 'guides', label: 'Guides' }] as const;
	const masterDetails: Record<string, { title: string; body: string }> = {
		overview: { title: 'Overview', body: 'A summary of the workspace: recent views, pinned items and indexing state.' },
		tokens: { title: 'Tokens', body: 'Semantic values grouped by surface, ink and status meaning.' },
		preferences: { title: 'Preferences', body: 'Effective values for this workspace with their scope and source.' }
	};
	let splitSize = $state(38);
	let inspectorOpen = $state(true);
	let accordionValue = $state('tokens');
	const breadcrumbItems: readonly BreadcrumbItem[] = [
		{ label: 'Atelier', href: '#overview' },
		{ label: 'Foundations', href: '#tokens' },
		{ label: 'Tokens' }
	];
	const accordionItems: readonly AccordionItem[] = [
		{ value: 'tokens', title: 'Token metadata', content: 'Semantic values are grouped by the surface, ink, and state meaning they carry.' },
		{ value: 'usage', title: 'Usage guidance', content: 'Prefer semantic tokens at the component boundary and keep feature state outside this layer.' },
		{ value: 'notes', title: 'Implementation notes', content: 'This section is intentionally controlled so a host can persist the open item.' }
	];
	const demoTabs = [
		{ value: 'overview', label: 'Overview' },
		{ value: 'tokens', label: 'Tokens' },
		{ value: 'preview', label: 'Preview', disabled: true }
	] as const;
	const paletteCommands: readonly CommandItem[] = [
		{ id: 'open-overview', label: 'Open overview', description: 'Show the overview view', shortcut: '⌘ 1' },
		{ id: 'inspect-tokens', label: 'Inspect tokens', description: 'Jump to the token inspector', shortcut: '⌘ 2' },
		{ id: 'toggle-sidebar', label: 'Toggle sidebar', description: 'Show or hide the Explorer panel', shortcut: '⌘ B' }
	];
	const formOptions = [
		{ value: 'canvas', label: 'Canvas view' },
		{ value: 'tokens', label: 'Token inspector' },
		{ value: 'preview', label: 'Preview', disabled: true }
	] as const;
	const radioOptions = [{ value: 'system', label: 'Use system theme' }, { value: 'light', label: 'Light theme' }, { value: 'dark', label: 'Dark theme' }] as const;
	const toggleOptions = [{ value: 'grid', label: 'Grid' }, { value: 'list', label: 'List' }, { value: 'split', label: 'Split' }] as const;
	const collectionItems = [{ id: 'overview', label: 'Overview', description: 'Workspace summary', meta: '⌘1' }, { id: 'tokens', label: 'Tokens', description: 'Semantic values', meta: '⌘2' }, { id: 'preferences', label: 'Preferences', description: 'App settings', meta: '⌘,' }, { id: 'archived', label: 'Archived', description: 'Read-only history', disabled: true }] as const;
	const treeNodes = [{ id: 'atelier', label: 'Atelier', children: [{ id: 'foundations', label: 'Foundations', children: [{ id: 'tokens-node', label: 'Tokens' }, { id: 'components-node', label: 'Components' }] }, { id: 'workbench-node', label: 'Workbench' }, { id: 'preferences-node', label: 'Preferences' }] }] as const;

	const surfaces = [
		'--surface-ground',
		'--surface-rail',
		'--surface-sunk',
		'--surface-panel',
		'--surface-panel-2',
		'--surface-raised'
	] as const;
	const ink = ['--ink', '--ink-muted', '--ink-subtle', '--ink-disabled'] as const;
	const spacing = [
		'--space-1',
		'--space-2',
		'--space-3',
		'--space-4',
		'--space-5',
		'--space-6',
		'--space-7',
		'--space-8',
		'--space-9',
		'--space-10'
	] as const;
	const status = [
		{ name: 'success', color: '--success', tint: '--success-tint' },
		{ name: 'warning', color: '--warning', tint: '--warning-tint' },
		{ name: 'danger', color: '--danger', tint: '--danger-tint' }
	] as const;

	async function loadAppearance(): Promise<void> {
		const state = await appearance.load();
		theme = state.theme;
		density = state.density;
		appearanceNote = state.failure ? `Appearance load failed: ${state.failure.kind}${state.failure.type ? ` · ${state.failure.type}` : ''}` : `Theme ${state.themeSource}, density ${state.densitySource}`;
	}

	onMount(() => {
		void loadAppearance();
		if (import.meta.env.VITE_ATELIER_PROBE === 'preferences') void runLaunchProbe(preferences.transport, host).then((result) => (appearanceNote = result));
		if (import.meta.env.VITE_ATELIER_PROBE === 'workbench') void runSnapshotProbe(workbench.snapshotPort, host).then((result) => (appearanceNote = result));
	});

	async function chooseTheme(value: Theme): Promise<void> {
		theme = value;
		const outcome = await appearance.setTheme(value);
		appearanceNote = outcome.ok ? 'Theme stored' : `Theme not stored: ${outcome.error.kind} · commit ${outcome.error.commit}`;
	}

	async function chooseDensity(value: Density): Promise<void> {
		density = value;
		const outcome = await appearance.setDensity(value);
		appearanceNote = outcome.ok ? 'Density stored' : `Density not stored: ${outcome.error.kind} · commit ${outcome.error.commit}`;
	}

	function runCommand(command: CommandItem): void {
		lastCommand = command.label;
	}

	function chooseMenuAction(action: string): void {
		lastMenuAction = action;
	}

	function showToast(): void {
		const id = `toast-${toastSequence}`;
		toastSequence += 1;
		feedbackToasts = [
			...feedbackToasts,
			{ id, title: 'Preview queued', description: 'The preview will update when the view is ready.', tone: 'accent' }
		];
	}

	function dismissToast(id: string): void {
		feedbackToasts = feedbackToasts.filter((toast) => toast.id !== id);
	}

	function showCopyResult(value: string): void {
		copyFeedback = `Copied ${value.length} characters`;
	}

	function resolveConfirm(outcome: ConfirmOutcome): void {
		if (outcome === 'cancel') {
			confirmOutcome = 'Cancelled; nothing changed';
			return;
		}
		confirmBusy = true;
		setTimeout(() => {
			confirmBusy = false;
			confirmOpen = false;
			confirmOutcome = 'Confirmed; workspace layout reset';
		}, 900);
	}

	function closeTabIn(group: { tabs: EditorTabItem[]; active: string }, tab: EditorTabItem): { tabs: EditorTabItem[]; active: string } {
		const index = group.tabs.findIndex((entry) => entry.id === tab.id);
		const tabs = group.tabs.filter((entry) => entry.id !== tab.id);
		const active = group.active === tab.id ? (tabs[Math.min(index, tabs.length - 1)]?.id ?? '') : group.active;
		return { tabs, active };
	}
	function reorderTabs(tabs: EditorTabItem[], from: number, to: number): EditorTabItem[] {
		const next = [...tabs];
		const [moved] = next.splice(from, 1);
		next.splice(to, 0, moved);
		return next;
	}
	function resolveSave(outcome: SaveChangesOutcome): void {
		if (outcome === 'cancel') {
			saveOutcome = `Cancelled closing ${saveItems.join(', ')}`;
			return;
		}
		if (outcome === 'discard') {
			saveOutcome = `Discarded changes in ${saveItems.join(', ')}`;
			saveOpen = false;
			return;
		}
		saveBusy = true;
		setTimeout(() => {
			saveBusy = false;
			saveOutcome = `Saved ${saveItems.join(', ')}`;
			saveOpen = false;
		}, 900);
	}
	function askSaveMany(): void {
		saveItems = ['semantic.css', 'notes.md', 'CONTRACT.md'];
		saveOpen = true;
	}
	function openQuick(mode: string): void {
		quickMode = mode;
		quickQuery = '';
		quickOpen = true;
	}
	function pickQuick(item: QuickPickItem, mode: string): void {
		if (mode === '>') {
			quickLast = `Ran command: ${item.label}`;
			lastCommand = item.label;
		} else if (mode === '@') {
			quickLast = `Jumped to symbol: ${item.label}`;
		} else {
			quickRecents = [item.id, ...quickRecents.filter((entry) => entry !== item.id)].slice(0, 4);
			quickLast = `Opened file: ${item.label}`;
		}
	}
	function quickAction(item: QuickPickItem, action: QuickPickAction): void {
		const file = quickFiles.find((entry) => entry.id === item.id);
		if (!file) return;
		if (action.id === 'side') {
			quickLast = `Opened ${file.label} to the side`;
			quickOpen = false;
		} else {
			quickRecents = quickRecents.filter((entry) => entry !== file.id);
			quickLast = `Removed ${file.label} from recents`;
		}
	}
	function closeStrip(tab: EditorTabItem): void {
		({ tabs: stripTabs, active: stripTab } = closeTabIn({ tabs: stripTabs, active: stripTab }, tab));
		lastTabEvent = `Closed ${tab.label}`;
	}
	function closeStripOthers(tab: EditorTabItem): void {
		stripTabs = stripTabs.filter((entry) => entry.id === tab.id || entry.pinned || entry.closable === false);
		stripTab = tab.id;
		lastTabEvent = `Closed others around ${tab.label}`;
	}
	function closeStripRight(tab: EditorTabItem): void {
		const index = stripTabs.findIndex((entry) => entry.id === tab.id);
		stripTabs = stripTabs.filter((entry, position) => position <= index || entry.pinned || entry.closable === false);
		if (!stripTabs.some((entry) => entry.id === stripTab)) stripTab = tab.id;
		lastTabEvent = `Closed tabs to the right of ${tab.label}`;
	}
	function closeStripAll(): void {
		stripTabs = stripTabs.filter((entry) => entry.pinned || entry.closable === false);
		stripTab = stripTabs[0]?.id ?? '';
		lastTabEvent = 'Closed all closable tabs';
	}
	function pinStrip(tab: EditorTabItem, pinned: boolean): void {
		const rest = stripTabs.filter((entry) => entry.id !== tab.id);
		const updated = { ...tab, pinned, preview: pinned ? false : tab.preview };
		stripTabs = pinned ? [...rest.filter((entry) => entry.pinned), updated, ...rest.filter((entry) => !entry.pinned)] : [...rest, updated];
		lastTabEvent = `${pinned ? 'Pinned' : 'Unpinned'} ${tab.label}`;
	}
	function toggleDirty(): void {
		stripTabs = stripTabs.map((entry) => (entry.id === stripTab ? { ...entry, dirty: !entry.dirty } : entry));
	}
</script>

<svelte:head><title>Kitchen sink · Atelier</title></svelte:head>

<div class="sink-shell">
	<a class="skip-link" href="#gallery-main">Skip to gallery</a>
	<header class="sink-header">
		<div class="brand-lockup">
			<div class="brand-mark" aria-hidden="true">A</div>
			<div>
				<strong>Atelier</strong>
				<span>Design system kitchen sink</span>
			</div>
		</div>
		<div class="header-meta">
			<span>{host}</span>
			<span class="appearance-note" aria-live="polite">{appearanceNote}</span>
			<div class="segmented" role="group" aria-label="Theme">
				{#each THEMES as option}
					<button
						type="button"
						class:active={theme === option}
						aria-pressed={theme === option}
						onclick={() => chooseTheme(option)}>{option}</button
					>
				{/each}
			</div>
			<div class="segmented" role="group" aria-label="Density">
				{#each DENSITIES as option}
					<button
						type="button"
						class:active={density === option}
						aria-pressed={density === option}
						onclick={() => chooseDensity(option)}>{option}</button
					>
				{/each}
			</div>
		</div>
	</header>

	<div class="sink-body">
		<nav class="sink-nav" aria-label="Kitchen sink sections">
			<p class="nav-label">Sections</p>
			<a href="#overview">Overview</a>
			<a href="#tokens">Tokens</a>
			<a href="#preferences">Preferences</a>
			<a href="#workspace">Workspace</a>
			<a href="#structure">Structure</a>
			<a href="#layout">Layout</a>
			<a href="#typography">Typography</a>
			<a href="#forms">Forms</a>
			<a href="#advanced-inputs">Advanced inputs</a>
			<a href="#feedback">Feedback</a>
			<a href="#feedback-states">Feedback states</a>
			<a href="#data-display">Data display</a>
			<a href="#content-utilities">Content utilities</a>
			<a href="#workbench-interactions">Workbench interactions</a>
			<a href="#patterns">Patterns</a>
			<a href="#navigation">Navigation</a>
			<a href="#overlays">Overlays</a>
			<a href="#commands">Commands</a>
			<a href="#menus">Menus &amp; choices</a>
			<a href="#collections">Collections</a>
			<a href="#controls">Controls</a>
			<a href="#editor-tabs">Editor tabs</a>
			<a href="#workbench">Workbench</a>
		</nav>

		<main id="gallery-main" class="sink-main" tabindex="-1">
			<section id="overview" class="intro" aria-labelledby="overview-title">
				<p class="eyebrow">Atelier / frontend foundation</p>
				<h1 id="overview-title">A place for every piece.</h1>
				<p class="intro-copy">
					A small, live view of the values the workbench will read. Theme and density are real token
					changes, not presentation-only examples.
				</p>
			</section>

			<section id="preferences" class="gallery-section" aria-labelledby="preferences-title">
				<div class="section-heading"><div><p class="eyebrow">Native round trip</p><h2 id="preferences-title">Preferences</h2></div><p>The first slice that crosses the bridge: kernel Results, validated wire codecs, compare-and-replace with reconcile, and the header's theme and density persisted through it.</p></div>
				<PreferencesPanel transport={preferences.transport} preview={preferences.preview} storeLabel={preferences.storeLabel} onChanged={loadAppearance} />
			</section>

			<section id="workspace" class="gallery-section" aria-labelledby="workspace-title">
				<div class="section-heading"><div><p class="eyebrow">Registry metadata</p><h2 id="workspace-title">Workspace</h2></div><p>Active and archived registry records are listed through the presence-aware service. Open re-reads the record and hands an active session to the Workbench; it does not inspect project files.</p></div>
				<WorkspacePicker transport={workspace.transport} openPort={workspace.openPort} storeLabel={workspace.storeLabel} onOpen={handleWorkspaceOpen} />
				<Text size="sm" tone="quiet">{workspaceNote}</Text>
			</section>

			<section id="tokens" class="gallery-section" aria-labelledby="tokens-title">
				<div class="section-heading">
					<div><p class="eyebrow">Foundations</p><h2 id="tokens-title">Semantic tokens</h2></div>
					<p>Components read these meanings instead of raw palette values.</p>
				</div>
				<div class="token-grid">
					<div class="token-card token-card-wide">
						<h3>Surfaces</h3>
						<div class="surface-swatches">
							{#each surfaces as token}
								<div class="surface-swatch" style={`--swatch: var(${token})`}>
									<span></span><code>{token}</code>
								</div>
							{/each}
						</div>
					</div>
					<div class="token-card">
						<h3>Ink</h3>
						<div class="ink-list">
							{#each ink as token}
								<p style={`color: var(${token})`}><code>{token}</code> Readable interface text</p>
							{/each}
						</div>
					</div>
					<div class="token-card">
						<h3>Status</h3>
						<div class="status-list">
							{#each status as item}
								<div class="status-chip" style={`--status-color: var(${item.color}); --status-tint: var(${item.tint})`}>
									<span></span>{item.name}<code>{item.color}</code>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</section>

			<section id="structure" class="gallery-section" aria-labelledby="structure-title">
				<div class="section-heading">
					<div><p class="eyebrow">Primitives</p><h2 id="structure-title">Structural primitives</h2></div>
					<p>Surfaces establish depth; separators establish relationships without adding noise.</p>
				</div>
				<div class="surface-examples">
					<Surface tone="sunk" padding="md" class="surface-example"><strong>Sunk</strong><span>Recessed workspace or input background.</span></Surface>
					<Surface tone="panel" padding="md" class="surface-example"><strong>Panel</strong><span>Default surface for a workbench region.</span></Surface>
					<Surface tone="raised" padding="md" class="surface-example"><strong>Raised</strong><span>Nearest surface for active controls.</span></Surface>
				</div>
				<div class="separator-demo"><span>Sidebar</span><Separator orientation="vertical" /><span>Main content</span></div>
			</section>

			<section id="layout" class="gallery-section" aria-labelledby="layout-title">
				<div class="section-heading"><div><p class="eyebrow">Primitives</p><h2 id="layout-title">Layout and icons</h2></div><p>Stack, inline and grid place children on the spacing scale; icons frame caller-supplied paths.</p></div>
				<Surface tone="panel" padding="lg" class="layout-demo">
					<Grid columns={2} gap="lg">
						<Stack gap="sm"><p class="field-label">Stack · gap sm</p><Stack gap="sm"><div class="layout-box">one</div><div class="layout-box">two</div><div class="layout-box">three</div></Stack></Stack>
						<Stack gap="sm"><p class="field-label">Inline · wrap, justify between</p><Inline gap="sm" wrap justify="between"><div class="layout-box">start</div><div class="layout-box">middle</div><div class="layout-box">end</div></Inline><p class="field-label">Inline · align baseline</p><Inline gap="md" align="baseline"><Heading level={3} size="lg">Title</Heading><Text size="sm" tone="quiet">caption</Text></Inline></Stack>
					</Grid>
					<Stack gap="sm"><p class="field-label">Grid · auto-fit, min 140px</p><Grid columns="auto" minColumn="140px" gap="sm">{#each spacing.slice(0, 6) as token}<div class="layout-box">{token}</div>{/each}</Grid></Stack>
					<Stack gap="sm"><div class="interaction-label-row"><p class="field-label">SplitGroup · three collapsible panes, nested vertical split</p><Text size="sm" tone="quiet">{splitSizes.map((size) => Math.round(size)).join(' / ')} · Enter or double-click a divider to collapse</Text></div><div class="split-group-demo"><SplitGroup bind:sizes={splitSizes} minSize={12} collapsible ariaLabel="Editor columns">{#snippet pane(index)}{#if index === 1}<SplitGroup direction="vertical" bind:sizes={nestedSizes} minSize={20} ariaLabel="Center rows">{#snippet pane(row)}<div class="split-pane"><strong>{row === 0 ? 'Editor' : 'Terminal'}</strong><span>{row === 0 ? 'Nested split inside the center column.' : 'Vertical resize with Up and Down.'}</span></div>{/snippet}</SplitGroup>{:else}<div class="split-pane"><strong>{index === 0 ? 'Explorer' : 'Inspector'}</strong><span>Drag past the minimum to collapse this pane.</span><div class="split-placeholder"></div></div>{/if}{/snippet}</SplitGroup></div></Stack>
					<Separator />
					<Inline gap="lg" wrap>
						<Inline gap="xs"><Icon size="xs"><path d="M5 12h14" /><path d="M12 5v14" /></Icon><Icon size="sm"><path d="M5 12h14" /><path d="M12 5v14" /></Icon><Icon size="md"><path d="M5 12h14" /><path d="M12 5v14" /></Icon><Icon size="lg"><path d="M5 12h14" /><path d="M12 5v14" /></Icon><Text size="sm" tone="quiet">sizes</Text></Inline>
						<Inline gap="xs"><Icon label="Search"><circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" /></Icon><Icon label="Folder"><path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" /></Icon><Icon label="Warning" class="icon-warning"><path d="M12 3 2 20h20z" /><path d="M12 9v5" /><path d="M12 17h.01" /></Icon><Text size="sm" tone="quiet">labelled, inherit color</Text></Inline>
						<Button variant="secondary" size="sm"><Icon size="sm"><path d="M12 5v14" /><path d="M5 12h14" /></Icon>New view</Button>
					</Inline>
				</Surface>
			</section>

			<section id="typography" class="gallery-section" aria-labelledby="typography-title">
				<div class="section-heading">
					<div><p class="eyebrow">Primitives</p><h2 id="typography-title">Typography</h2></div>
					<p>Document structure and visual scale stay separate.</p>
				</div>
				<Surface tone="panel" padding="lg" class="typography-demo">
					<p class="eyebrow">Section label</p>
					<Heading level={3} size="lg">A heading with a deliberate level</Heading>
					<Text size="lg" tone="muted" measure>Readable interface copy has a measure and a tone, while the heading still owns the document outline.</Text>
					<Separator />
					<div class="typography-field"><Label for="typography-sample">Field label</Label><input id="typography-sample" class="field" value="Typography pairs with controls" /></div>
				</Surface>
			</section>

			<section id="forms" class="gallery-section" aria-labelledby="forms-title">
				<div class="section-heading">
					<div><p class="eyebrow">Forms</p><h2 id="forms-title">Fields and choices</h2></div>
					<p>Native text entry stays familiar; Bits UI owns the interaction-heavy choices.</p>
				</div>
				<Surface tone="panel" padding="lg" class="forms-demo">
					<div class="form-grid">
						<Field label="Name" description="Shown in the title bar and the recent list." message={fieldName.trim() ? 'This name is available.' : 'A name helps you find this workspace later.'} tone={fieldName.trim() ? 'success' : 'hint'} required>{#snippet children({ id, describedBy, invalid })}<Input {id} aria-describedby={describedBy} {invalid} bind:value={fieldName} placeholder="Untitled workspace" />{/snippet}</Field>
						<div class="form-field"><Label for="form-view">Default view</Label><Select id="form-view" bind:value={formChoice} options={formOptions} aria-label="Default view" /></div>
						<div class="form-field form-field-wide"><Label for="form-notes">Notes</Label><Textarea id="form-notes" rows={3} placeholder="A quiet place for context..." /></div>
					</div>
					<Separator />
					<div class="choice-list">
						<div class="choice-row"><Checkbox bind:checked={formChecked} aria-label="Remember this workspace" /><div class="choice-copy"><strong>Remember this workspace</strong><span>Restore the active view when the app opens.</span></div></div>
						<div class="choice-row"><Switch bind:checked={formPinned} aria-label="Pin workspace" /><div class="choice-copy"><strong>Pin workspace</strong><span>Keep this workspace visible in the activity rail.</span></div></div>
					</div>
				</Surface>
			</section>

			<section id="advanced-inputs" class="gallery-section" aria-labelledby="advanced-inputs-title">
				<div class="section-heading"><div><p class="eyebrow">Forms</p><h2 id="advanced-inputs-title">Advanced inputs</h2></div><p>Search, filtering, ranges, and field messaging compose without taking ownership of feature state.</p></div>
				<Surface tone="panel" padding="lg" class="advanced-inputs-demo">
					<div class="advanced-input-grid">
						<div class="advanced-input-field"><Label>View filter</Label><Combobox options={formOptions} bind:value={comboboxValue} placeholder="Filter views" ariaLabel="Filter views" /></div>
						<div class="advanced-input-field"><Label for="advanced-search">Search workspace</Label><SearchField id="advanced-search" bind:value={searchValue} placeholder="Search files and views" ariaLabel="Search files and views" /></div>
						<div class="advanced-input-field advanced-input-field-wide"><div class="slider-value-row"><Label>Preview opacity</Label><Text size="sm" tone="quiet">{sliderValue}%</Text></div><Slider bind:value={sliderValue} step={5} ariaLabel="Preview opacity" /></div>
						<div class="advanced-input-field advanced-input-field-wide"><Label for="input-group-name">Workspace name</Label><InputGroup><Input id="input-group-name" bind:value={inputGroupValue} placeholder="Workspace name" />{#snippet leading()}<span aria-hidden="true">⌁</span>{/snippet}{#snippet trailing()}<span>local</span>{/snippet}</InputGroup><FieldMessage tone={inputGroupValue ? 'success' : 'error'}>{inputGroupValue ? 'This name is available.' : 'Enter a workspace name.'}</FieldMessage></div>
					</div>
					<Text size="sm" tone="quiet">Selected view: {comboboxValue || 'None'} · Search query: {searchValue || 'None'}</Text>
				</Surface>
			</section>

			<section id="feedback" class="gallery-section" aria-labelledby="feedback-title">
				<div class="section-heading">
					<div><p class="eyebrow">Feedback</p><h2 id="feedback-title">State and progress</h2></div>
					<p>Small signals carry status without taking over the workbench.</p>
				</div>
				<Surface tone="panel" padding="lg" class="feedback-demo">
					<div class="feedback-row"><Status tone="success" pulse>Connected</Status><Status tone="warning">Needs attention</Status><Status tone="danger">Blocked</Status><Status tone="accent">Syncing</Status></div>
					<div class="feedback-row"><Badge tone="neutral">Draft</Badge><Badge tone="accent">Active</Badge><Badge tone="success" variant="solid">Ready</Badge><Badge tone="warning" variant="outline">Review</Badge><Badge tone="danger" size="md">Error</Badge></div>
					<Separator />
					<div class="progress-list"><div class="progress-item"><div class="progress-meta"><span>Workspace indexing</span><span>68%</span></div><Progress value={68} aria-label="Workspace indexing 68 percent" /></div><div class="progress-item"><div class="progress-meta"><span>Waiting for preview</span><Spinner size="sm" label="Waiting for preview" /></div><Progress aria-label="Waiting for preview" /></div></div>
				</Surface>
			</section>

			<section id="feedback-states" class="gallery-section" aria-labelledby="feedback-states-title">
				<div class="section-heading">
					<div><p class="eyebrow">Feedback</p><h2 id="feedback-states-title">Loading, empty and alert states</h2></div>
					<p>Longer-lived states explain what is happening and give the user a clear next step.</p>
				</div>
				<div class="feedback-state-stack">
					{#if alertVisible}
						<Alert tone="warning" title="Preview is stale" dismissible onDismiss={() => (alertVisible = false)}>
							Re-run the preview to see the latest workspace changes.
						</Alert>
					{/if}
					{#if bannerVisible}
						<Banner tone="accent" title="New workbench update" dismissible onDismiss={() => (bannerVisible = false)}>
							Panels now remember their last size.
							{#snippet actions()}<Button variant="quiet" size="sm">Review</Button>{/snippet}
						</Banner>
					{/if}
					<Surface tone="panel" padding="lg" class="feedback-state-grid">
						<div class="feedback-state-column">
							<p class="field-label">Loading</p>
							<div class="skeleton-stack"><Skeleton width="78%" /><Skeleton width="48%" /><Skeleton variant="rect" width="100%" height="72px" /></div>
						</div>
						<EmptyState title="Nothing here yet" description="Create a view to start filling this workspace.">
							{#snippet action()}<Button variant="secondary" size="sm">Create view</Button>{/snippet}
						</EmptyState>
					</Surface>
					<div class="toast-demo"><Button variant="secondary" onclick={showToast}>Show toast</Button><Text size="sm" tone="muted">The host owns lifecycle; Toaster only renders the current queue.</Text></div>
					<Toaster toasts={feedbackToasts} onDismiss={dismissToast} />
				</div>
			</section>

			<section id="navigation" class="gallery-section" aria-labelledby="navigation-title">
				<div class="section-heading"><div><p class="eyebrow">Navigation</p><h2 id="navigation-title">Tabs and toolbar</h2></div><p>Keyboard-oriented navigation for view-local state; a toolbar is one tab stop with arrow-key movement.</p></div>
				<Surface tone="panel" padding="lg" class="navigation-demo"><Tabs items={demoTabs} bind:value={demoTab}><TabPanel value="overview"><div class="tab-panel-copy"><strong>Overview</strong><span>The active view can hold a workbench canvas without changing the surrounding shell.</span></div></TabPanel><TabPanel value="tokens"><div class="tab-panel-copy"><strong>Tokens</strong><span>Local navigation stays independent from application-level routing.</span></div></TabPanel></Tabs><Separator /><div class="toolbar-demo"><Toolbar ariaLabel="Canvas tools"><ToolbarButton label="Undo" onclick={() => (lastToolbarAction = 'Undo')}><Icon size="sm"><path d="M9 14 4 9l5-5" /><path d="M4 9h10a6 6 0 0 1 0 12h-3" /></Icon></ToolbarButton><ToolbarButton label="Redo" onclick={() => (lastToolbarAction = 'Redo')}><Icon size="sm"><path d="m15 14 5-5-5-5" /><path d="M20 9H10a6 6 0 0 0 0 12h3" /></Icon></ToolbarButton><Separator orientation="vertical" class="toolbar-separator" /><ToolbarGroup items={toolbarAlignments} bind:value={toolbarAlignment} ariaLabel="Alignment" /><ToolbarGroup type="multiple" items={toolbarMarkOptions} bind:value={toolbarMarks} ariaLabel="Canvas marks" /><Separator orientation="vertical" class="toolbar-separator" /><ToolbarButton label="Export" disabled>Export</ToolbarButton></Toolbar><Text size="sm" tone="quiet">Last action: {lastToolbarAction} · Alignment: {toolbarAlignment} · Marks: {toolbarMarks.length ? toolbarMarks.join(', ') : 'none'}</Text></div></Surface>
			</section>

			<section id="overlays" class="gallery-section" aria-labelledby="overlays-title">
				<div class="section-heading"><div><p class="eyebrow">Overlays</p><h2 id="overlays-title">Tooltip</h2></div><p>Short help appears on hover or keyboard focus without competing with the canvas.</p></div>
				<Surface tone="panel" padding="lg" class="overlay-demo">
					<div class="overlay-actions"><Tooltip content="This is a keyboard-accessible hint" class="tooltip-demo-trigger"><span>Hover or focus me</span></Tooltip><Popover>
						{#snippet trigger()}<span class="popover-demo-trigger">Open popover</span>{/snippet}
						<div class="popover-demo-card"><strong>Quick view</strong><span>Contextual actions can sit close to the thing they affect.</span></div>
					</Popover><Dialog title="Rename view" description="Give this view a short name for the workbench." size="sm">
						{#snippet trigger()}<span class="dialog-demo-trigger">Open dialog</span>{/snippet}
						<div class="dialog-demo-body"><Label for="dialog-name">View name</Label><Input id="dialog-name" value="Overview" /><div class="dialog-actions"><Button variant="quiet">Cancel</Button><Button variant="primary">Save view</Button></div></div>
					</Dialog></div>
					<Text size="sm" tone="muted">Tooltip, popover and dialog all own their keyboard and dismissal behavior through Bits UI.</Text>
				</Surface>
			</section>

			<section id="commands" class="gallery-section" aria-labelledby="commands-title">
				<div class="section-heading"><div><p class="eyebrow">Workbench</p><h2 id="commands-title">Command palette and quick pick</h2></div><p>Search, keyboard navigation and execution stay separate from the command registry; the quick pick adds modes, groups, recents and row actions.</p></div>
				<Surface tone="panel" padding="lg" class="command-demo"><div class="command-demo-copy"><strong>Quick actions</strong><Text size="sm" tone="muted">The palette is controlled by the host and reports the selected command through a callback.</Text></div><div class="command-demo-actions"><Button variant="secondary" onclick={() => (paletteOpen = true)}>Open command palette</Button><span class="kbd">⌘ K</span></div><Text size="sm" tone="quiet">Last command: {lastCommand}</Text></Surface>
				<CommandPalette bind:open={paletteOpen} commands={paletteCommands} onExecute={runCommand} />
				<Surface tone="panel" padding="lg" class="command-demo"><div class="command-demo-copy"><strong>Quick pick</strong><Text size="sm" tone="muted">One picker for files, commands and symbols. Type <InlineCode>&gt;</InlineCode> or <InlineCode>@</InlineCode> to switch modes, Backspace on an empty query to return, and hover a file for its row actions.</Text></div><Inline gap="sm" wrap><Button variant="secondary" onclick={() => openQuick('')}>Quick open</Button><Button variant="quiet" onclick={() => openQuick('>')}>Commands</Button><Button variant="quiet" onclick={() => openQuick('@')}>Symbols</Button><Text size="sm" tone="quiet">{quickLast}</Text></Inline></Surface>
				<QuickPick bind:open={quickOpen} bind:mode={quickMode} bind:query={quickQuery} items={quickItems} modes={quickModes} groupOrder={['Navigation', 'View', 'Types', 'Functions', 'State']} onSelect={pickQuick} onAction={quickAction}>{#snippet icon(item)}<Icon size="sm">{#if quickMode === '>'}<path d="m4 17 6-5-6-5" /><path d="M12 19h8" />{:else if quickMode === '@'}<circle cx="12" cy="12" r="4" /><path d="M16 12v1.5a2.5 2.5 0 0 0 5 0V12a9 9 0 1 0-3.5 7.1" />{:else if item.label.endsWith('.css')}<path d="M4 4h16v16H4z" /><path d="M8 12h8" />{:else if item.label.endsWith('.md')}<path d="M4 6h16v12H4z" /><path d="m8 15 0-6 2 3 2-3v6" /><path d="M16 9v6" />{:else}<path d="M6 3h8l4 4v14H6z" /><path d="M14 3v4h4" />{/if}</Icon>{/snippet}</QuickPick>
			</section>

			<section id="menus" class="gallery-section" aria-labelledby="menus-title">
				<div class="section-heading"><div><p class="eyebrow">Navigation + forms</p><h2 id="menus-title">Menus and choices</h2></div><p>Desktop actions, right-click context and preference choices share the same keyboard-first language.</p></div>
				<Surface tone="panel" padding="lg" class="menus-demo">
					<div class="menu-row"><DropdownMenu>
						{#snippet trigger()}<span class="menu-demo-trigger">Actions</span>{/snippet}
						<MenuItem label="Open overview" shortcut="⌘ O" onSelect={() => chooseMenuAction('Open overview')} /><MenuItem label="Rename view" shortcut="F2" onSelect={() => chooseMenuAction('Rename view')} /><MenuSeparator /><MenuItem label="Delete view" disabled />
					</DropdownMenu><ContextMenu>
						{#snippet trigger()}<div class="context-demo-target">Right-click this canvas</div>{/snippet}
						<MenuItem label="Add panel" onSelect={() => chooseMenuAction('Add panel')} /><MenuItem label="Duplicate view" onSelect={() => chooseMenuAction('Duplicate view')} /><MenuSeparator /><MenuItem label="Paste" disabled />
					</ContextMenu><Text size="sm" tone="muted">Last action: {lastMenuAction}</Text></div>
					<Separator />
					<div class="choices-grid"><div class="choice-demo"><Label>Display mode</Label><RadioGroup items={radioOptions} bind:value={menuRadio} ariaLabel="Display mode" /></div><div class="choice-demo"><Label>Layout</Label><ToggleGroup items={toggleOptions} bind:value={toggleValue} ariaLabel="Layout" /></div><div class="choice-demo"><Label>Compact labels</Label><Toggle bind:pressed={compactLabels}> {compactLabels ? 'On' : 'Off'} </Toggle></div></div>
					<div class="shortcut-row"><Text size="sm" tone="muted">Shortcuts</Text><Kbd>⌘ K</Kbd><Kbd>⌘ B</Kbd><Kbd>F2</Kbd></div>
				</Surface>
			</section>

			<section id="collections" class="gallery-section" aria-labelledby="collections-title">
				<div class="section-heading"><div><p class="eyebrow">Collections</p><h2 id="collections-title">List and tree</h2></div><p>Selection is explicit, expansion is controlled, and scroll behavior belongs to the collection surface.</p></div>
				<Surface tone="panel" padding="lg" class="collections-demo"><div class="collection-columns"><div class="collection-column"><Label>List</Label><List items={collectionItems} bind:value={listSelection} ariaLabel="Workspace views" /></div><div class="collection-column"><Label>Tree</Label><ScrollArea class="collection-scroll"><Tree items={treeNodes} bind:selectedId={treeSelection} bind:expandedIds={expandedTreeIds} ariaLabel="Workspace tree" /></ScrollArea></div></div><Text size="sm" tone="muted">Selected list item: {listSelection} · Selected tree item: {treeSelection}</Text></Surface>
			</section>

			<section id="data-display" class="gallery-section" aria-labelledby="data-display-title">
				<div class="section-heading"><div><p class="eyebrow">Data display</p><h2 id="data-display-title">Inspection surfaces</h2></div><p>Cards frame related information while tables and property lists keep dense workbench data readable.</p></div>
				<div class="data-display-grid">
					<Card title="Workspace metadata" description="A compact property surface for inspectors."><KeyValue items={workspaceProperties} columns={2} /></Card>
					<Card title="Recent views" description="Rows stay presentational; selection and sorting belong to the feature." actions={undefined}><Table columns={inspectorColumns} rows={inspectorRows} caption="Views in this workspace" compact /></Card>
				</div>
			</section>

			<section id="content-utilities" class="gallery-section" aria-labelledby="content-utilities-title">
				<div class="section-heading"><div><p class="eyebrow">Content utilities</p><h2 id="content-utilities-title">Identity and code</h2></div><p>Compact metadata, identity, and source snippets use the same quiet visual language as the workbench.</p></div>
				<div class="content-demo-grid">
					<Card title="Workspace identity" description="Tags and avatars keep context close to the thing it describes.">
						<div class="content-tag-row"><Avatar name="Atelier Core" status="online" size="lg" /><div class="content-identity-copy"><strong>Atelier Core</strong><span>Maintained by the platform team</span></div></div>
						<div class="content-tag-row"><Tag tone="accent" dot>Design system</Tag><Tag tone="success" variant="outline">Synced</Tag><Tag tone="warning" dismissible>Review</Tag></div>
						<p class="content-note">Tokens resolve through <InlineCode>var(--accent)</InlineCode><CopyButton value="var(--accent)" onCopy={showCopyResult} /></p>
						<Text size="sm" tone="quiet">{copyFeedback}</Text>
					</Card>
					<Card title="Token source" description="Code stays readable and copyable without owning syntax highlighting."><CodeBlock code={tokenSnippet} language="tokens.css" onCopy={showCopyResult} /></Card>
				</div>
			</section>

			<section id="workbench-interactions" class="gallery-section" aria-labelledby="workbench-interactions-title">
				<div class="section-heading"><div><p class="eyebrow">Workbench interactions</p><h2 id="workbench-interactions-title">Adjustable structure</h2></div><p>Panel geometry and disclosure state stay keyboard-accessible and controlled by the consuming surface.</p></div>
				<Surface tone="panel" padding="lg" class="interaction-demo">
					<Breadcrumb items={breadcrumbItems} />
					<div class="interaction-grid">
						<div class="interaction-column"><p class="field-label">Disclosure</p><Collapsible title="Inspector details" bind:open={inspectorOpen}><p>Keep secondary context available without adding another route or modal.</p></Collapsible><p class="field-label">Accordion</p><Accordion items={accordionItems} bind:value={accordionValue} ariaLabel="Token guidance" /></div>
						<div class="interaction-column"><div class="interaction-label-row"><p class="field-label">Resizable split</p><Text size="sm" tone="quiet">{Math.round(splitSize)} / {Math.round(100 - splitSize)}</Text></div><div class="split-demo"><Resizable bind:value={splitSize}>{#snippet first()}<div class="split-pane"><strong>Explorer</strong><span>Drag the divider or focus it and use the arrow keys.</span><div class="split-placeholder"></div></div>{/snippet}{#snippet second()}<div class="split-pane"><strong>Canvas</strong><span>The second pane fills the remaining workbench space.</span><div class="split-placeholder split-placeholder-wide"></div></div>{/snippet}</Resizable></div></div>
					</div>
				</Surface>
			</section>

			<section id="patterns" class="gallery-section" aria-labelledby="patterns-title">
				<div class="section-heading"><div><p class="eyebrow">Patterns</p><h2 id="patterns-title">Settings, confirmation and inline editing</h2></div><p>Compositions built only from components; the consumer owns values, outcomes and persistence.</p></div>
				<Surface tone="panel" padding="lg" class="patterns-demo">
					<div class="patterns-grid">
						<div class="patterns-column">
							<p class="field-label">Settings rows</p>
							<div class="settings-list">
								<SettingsRow label="Autosave" description="Write changes as soon as the editor is idle.">{#snippet control({ labelId, descriptionId })}<Switch bind:checked={settingsAutosave} aria-labelledby={labelId} aria-describedby={descriptionId} />{/snippet}</SettingsRow>
								<SettingsRow label="Share diagnostics" description="Send anonymous crash reports.">{#snippet control({ labelId, descriptionId })}<Checkbox bind:checked={settingsTelemetry} aria-labelledby={labelId} aria-describedby={descriptionId} />{/snippet}</SettingsRow>
								<SettingsRow label="Default view" description="Opened when the workspace starts.">{#snippet control({ labelId })}<Select options={formOptions} bind:value={formChoice} size="sm" aria-labelledby={labelId} class="settings-select" />{/snippet}</SettingsRow>
								<SettingsRow label="Reset layout" description="Restore the default panel arrangement.">{#snippet control()}<Button variant="danger" size="sm" onclick={() => (confirmOpen = true)}>Reset…</Button>{/snippet}</SettingsRow>
							</div>
							<Text size="sm" tone="quiet">Confirmation: {confirmOutcome}</Text>
							<Inline gap="sm" wrap><Button size="sm" variant="secondary" onclick={() => { saveItems = ['semantic.css']; saveOpen = true; }}>Close one dirty file…</Button><Button size="sm" variant="secondary" onclick={askSaveMany}>Close three dirty files…</Button></Inline>
							<Text size="sm" tone="quiet">Save changes: {saveOutcome} · closing a dirty view in the workbench demo asks the same question</Text>
							<ConfirmDialog bind:open={confirmOpen} title="Reset the workspace layout?" description="Panel sizes and open views return to their defaults. Files are not affected." confirmLabel="Reset layout" tone="danger" busy={confirmBusy} onResolve={resolveConfirm} />
							<SaveChangesDialog bind:open={saveOpen} items={saveItems} busy={saveBusy} onResolve={resolveSave} />
						</div>
						<div class="patterns-column">
							<p class="field-label">Editable label</p>
							<Inline gap="sm" wrap><EditableLabel bind:value={workspaceTitle} ariaLabel="Workspace title" size="lg" /><EditableLabel value="" placeholder="Add a subtitle" ariaLabel="Workspace subtitle" size="sm" /><EditableLabel value="Read only" ariaLabel="Locked title" disabled /></Inline>
							<Text size="sm" tone="quiet">Click or press Enter to edit; Escape reverts. Title: {workspaceTitle}</Text>
							<p class="field-label">Master–detail</p>
							<div class="master-detail-demo"><MasterDetail hasSelection={masterSelection in masterDetails} masterLabel="Views">{#snippet master()}<List items={collectionItems} bind:value={masterSelection} ariaLabel="Views" />{/snippet}{#snippet detail()}<div class="detail-copy"><strong>{masterDetails[masterSelection]?.title}</strong><p>{masterDetails[masterSelection]?.body}</p></div>{/snippet}{#snippet empty()}<span>Select a view to inspect it.</span>{/snippet}</MasterDetail></div>
						</div>
					</div>
				</Surface>
			</section>

			<section id="controls" class="gallery-section" aria-labelledby="controls-title">
				<div class="section-heading">
					<div><p class="eyebrow">Primitives</p><h2 id="controls-title">Control states</h2></div>
					<p>Compact by default, with visible focus and honest disabled states.</p>
				</div>
				<div class="token-card controls-card">
					<div class="control-row">
						<Button variant="primary">Primary action</Button>
						<Button variant="secondary">Secondary</Button>
						<Button variant="quiet">Quiet action</Button>
						<Button variant="danger">Danger</Button>
						<Button variant="primary" loading>Loading</Button>
						<Button variant="secondary" disabled>Disabled</Button>
					</div>
					<div class="control-row">
						<label class="field-label" for="sample-input">Input</label>
						<input id="sample-input" class="field" value="Workbench value" />
						<select class="field" aria-label="Example select">
							<option>Example select</option>
							<option>Another option</option>
						</select>
						<IconButton label="Open command palette" size="sm">⌘</IconButton>
						<span class="kbd">⌘ K</span>
					</div>
				</div>
			</section>

			<section id="editor-tabs" class="gallery-section" aria-labelledby="editor-tabs-title">
				<div class="section-heading"><div><p class="eyebrow">Workbench</p><h2 id="editor-tabs-title">Editor tabs</h2></div><p>Pinned, dirty, preview and locked tabs; overflow scrolls, arrows move, Delete closes, right-click for group actions, drag to reorder.</p></div>
				<Surface tone="panel" padding="none" class="editor-tabs-demo">
					<EditorTabs tabs={stripTabs} bind:activeId={stripTab} onClose={closeStrip} onCloseOthers={closeStripOthers} onCloseRight={closeStripRight} onCloseAll={closeStripAll} onPin={pinStrip} onReorder={(from, to) => (stripTabs = reorderTabs(stripTabs, from, to))} onActivate={(tab) => (lastTabEvent = `Activated ${tab.label}`)}>
						{#snippet icon(tab)}<Icon size="xs">{#if tab.label.endsWith('.css')}<path d="M4 4h16v16H4z" /><path d="M8 12h8" />{:else if tab.label.endsWith('.json')}<path d="M8 4c-3 0-3 3-3 4s0 4-3 4c3 0 3 3 3 4s0 4 3 4" /><path d="M16 4c3 0 3 3 3 4s0 4 3 4c-3 0-3 3-3 4s0 4-3 4" />{:else}<path d="M6 3h8l4 4v14H6z" /><path d="M14 3v4h4" />{/if}</Icon>{/snippet}
						{#snippet trailing()}<IconButton label="Split editor" size="sm" onclick={() => (lastTabEvent = 'Split requested')}>⫿</IconButton><IconButton label="More actions" size="sm">…</IconButton>{/snippet}
					</EditorTabs>
					<div class="editor-tabs-demo__body"><Inline gap="sm" wrap><Button size="sm" variant="secondary" onclick={toggleDirty} disabled={!stripTab}>Toggle dirty on active</Button><Button size="sm" variant="quiet" onclick={() => { stripTabs = [{ id: 'a', label: 'App.svelte', pinned: true }, { id: 'b', label: 'Very-long-file-name-that-overflows-the-tab.svelte', description: 'A long label truncates and shows the full path on hover' }, { id: 'c', label: 'index.ts', dirty: true }, { id: 'd', label: 'layout.css', preview: true }, { id: 'e', label: 'settings.json', closable: false }, { id: 'f', label: 'Tree.svelte' }, { id: 'g', label: 'List.svelte' }, { id: 'h', label: 'Table.svelte' }, { id: 'i', label: 'Card.svelte' }, { id: 'j', label: 'Toast.svelte' }]; stripTab = 'c'; }}>Reset tabs</Button></Inline><Text size="sm" tone="quiet">Active: {stripTabs.find((tab) => tab.id === stripTab)?.label ?? 'none'} · {stripTabs.length} open · {lastTabEvent}</Text></div>
				</Surface>
			</section>

			<section id="workbench" class="gallery-section" aria-labelledby="workbench-title">
				<div class="section-heading">
					<div><p class="eyebrow">Composition</p><h2 id="workbench-title">Workbench shell</h2></div>
					<p>Runs on the workbench store: views, layout and history come from the pure models, every action is a registered command reachable from the palette and its shortcut, and the snapshot buttons go through the injected restoration port.</p>
				</div>
				<WorkbenchDemo {host} {dragRegion} {dragExclude} snapshotPort={workbench.snapshotPort} session={activeSession} />
			</section>

			<section class="scale-section" aria-labelledby="scale-title">
				<div class="section-heading"><div><p class="eyebrow">Scale</p><h2 id="scale-title">Spacing rhythm</h2></div><p>Values stay small and predictable for dense desktop layouts.</p></div>
				<div class="scale-list">
					{#each spacing as token}
						<div class="scale-row"><code>{token}</code><span class="scale-bar" style={`inline-size: var(${token})`}></span></div>
					{/each}
				</div>
			</section>
		</main>
	</div>
</div>
