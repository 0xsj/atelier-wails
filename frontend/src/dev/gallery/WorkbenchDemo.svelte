<script lang="ts">
	// The gallery's workbench runs on the real store: pure models underneath,
	// commands and shortcuts through the dispatcher, restoration through the
	// injected preview or native snapshot port. Content inside a view is still
	// placeholder text.
	import { onDestroy } from 'svelte';
	import { Button, Icon, IconButton, Text } from '../../lib/components/primitives';
	import { Inline, SplitGroup } from '../../lib/components/layout';
	import { EmptyState } from '../../lib/components/feedback';
	import { Switch } from '../../lib/components/forms';
	import { List, Tree } from '../../lib/components/collections';
	import { SaveChangesDialog, type SaveChangesOutcome } from '../../lib/components/patterns';
	import { PaneStack, type PaneStackItem } from '../../lib/workbench/ui/panels';
	import { ActivityItem, ActivityRail, TitleBar, WorkbenchShell, type ActivityEntry } from '../../lib/workbench/ui/shell';
	import { EditorTabs, type EditorTabItem } from '../../lib/workbench/ui/view-host';
	import { StatusBar, StatusItem } from '../../lib/workbench/ui/status';
	import { CommandPalette, type CommandItem } from '../../lib/workbench/ui/command-palette';
	import { QuickPick, type QuickPickAction, type QuickPickItem } from '../../lib/workbench/ui/quick-pick';
	import { createWorkbenchStore, loadSnapshot, registerWorkbenchCommands, resolveShortcut, WORKBENCH_COMMANDS, type LoadOutcome, type SnapshotPort } from '../../lib/workbench/app';
	import { paletteEntries, viewsIn, formatKeybinding, parseKeybinding, type Platform, type ViewDescriptor } from '../../lib/workbench/model';
	import type { WorkspaceSession } from '../../lib/services/workspace';

	let { host, dragRegion = {}, dragExclude = {}, snapshotPort, session = null }: { host: string; dragRegion?: Record<string, string>; dragExclude?: Record<string, string>; snapshotPort: SnapshotPort; session?: WorkspaceSession | null } = $props();

	const platform: Platform = typeof navigator !== 'undefined' && /mac/i.test(navigator.platform) ? 'mac' : 'linux';
	const store = createWorkbenchStore({ platform });
	const files: readonly (ViewDescriptor & { body: string })[] = [
		{ id: 'readme', kind: 'file', title: 'README.md', description: 'atelier-core/README.md', body: 'Pinned views keep their place and hide the close button.' },
		{ id: 'tokens', kind: 'file', title: 'semantic.css', description: 'src/lib/styles/tokens/semantic.css', body: 'A dirty view shows a dot; closing it asks through SaveChangesDialog.' },
		{ id: 'shell', kind: 'file', title: 'WorkbenchShell.svelte', description: 'src/lib/workbench/ui/shell/WorkbenchShell.svelte', body: 'Middle-click a tab, press Delete on it, or use the Close Editor command.' },
		{ id: 'store', kind: 'file', title: 'workbench-store.ts', description: 'src/lib/workbench/app/workbench-store.ts', body: 'Every change here goes through the store and its pure models.' },
		{ id: 'contract', kind: 'file', title: 'CONTRACT.md', description: 'domains/preferences/CONTRACT.md', body: 'Editor groups come from the views model; sizes from the layout model.' },
		{ id: 'notes', kind: 'file', title: 'notes.md', description: 'work/notes.md', body: 'Back and forward walk the activation history.' },
		{ id: 'quick', kind: 'file', title: 'QuickPick.svelte', description: 'src/lib/workbench/ui/quick-pick/QuickPick.svelte', body: 'Opened from the quick pick as a preview view.' },
		{ id: 'tree', kind: 'file', title: 'Tree.svelte', description: 'src/lib/components/collections/Tree.svelte', body: 'Opened from the quick pick as a preview view.' }
	];
	const settingsView: ViewDescriptor & { body: string } = { id: 'settings', kind: 'settings', title: 'Settings', body: 'A non-file view kind. The store does not care what a kind means.', closable: true };
	const bodies = new Map(files.map((file) => [file.id, file.body]));
	bodies.set(settingsView.id, settingsView.body);

	const activityItems: readonly ActivityEntry[] = [
		{ id: 'files', label: 'Files', badge: 3 },
		{ id: 'search', label: 'Search' },
		{ id: 'changes', label: 'Source control', badge: 12 },
		{ id: 'jobs', label: 'Jobs', badge: '!' }
	];
	const sidebarPanes: readonly PaneStackItem[] = [
		{ id: 'explorer', title: 'Explorer' },
		{ id: 'outline', title: 'Outline' },
		{ id: 'timeline', title: 'Timeline' }
	];
	const treeNodes = [{ id: 'atelier', label: 'Atelier', children: [{ id: 'foundations', label: 'Foundations', children: [{ id: 'tokens-node', label: 'Tokens' }, { id: 'components-node', label: 'Components' }] }, { id: 'workbench-node', label: 'Workbench' }, { id: 'preferences-node', label: 'Preferences' }] }] as const;
	const outlineItems = [{ id: 'overview', label: 'Overview', description: 'Workspace summary' }, { id: 'tokens', label: 'Tokens', description: 'Semantic values' }, { id: 'preferences', label: 'Preferences', description: 'App settings' }] as const;
	const quickActions: readonly QuickPickAction[] = [{ id: 'side', label: 'Open to the side', glyph: '⫿' }];

	let windowActive = $state(true);
	let paletteOpen = $state(false);
	let quickOpen = $state(false);
	let quickQuery = $state('');
	let registryVersion = $state(0);
	let lastOutcome = $state('Focus the workbench and press a shortcut, or run a command from the palette');
	let saveOpen = $state(false);
	let saveBusy = $state(false);
	let pendingClose = $state<string[]>([]);
	let snapshotNote = $state('No snapshot saved yet');
	let treeSelection = $state('atelier');
	let expandedTreeIds = $state<string[]>(['atelier']);
	let outlineSelection = $state('overview');

	const stopRegistry = store.dispatcher.onChange(() => (registryVersion += 1));
	onDestroy(stopRegistry);

	function seed(): void {
		store.restore({ kind: 'empty' });
		const fresh = createWorkbenchStore({ platform });
		fresh.openView(files[0]);
		fresh.openView(files[1]);
		fresh.openView(files[2]);
		fresh.openView(files[3]);
		fresh.pinView('readme', true);
		fresh.setDirty('tokens', true);
		fresh.activateView('tokens');
		fresh.splitGroup();
		fresh.openView(files[4]);
		fresh.openView(files[5]);
		fresh.setDirty('notes', true);
		fresh.activateView('contract');
		fresh.selectActivity('files');
		fresh.togglePane('timeline');
		fresh.setPaneWeights({ explorer: 2, outline: 1, timeline: 1 });
		fresh.setEditorSizes([55, 45]);
		store.restore({ kind: 'restored', layout: fresh.state.layout, views: fresh.state.views, dropped: [] });
	}

	function askClose(viewIds: readonly string[]): void {
		if (!viewIds.length) return;
		pendingClose = [...viewIds];
		saveOpen = true;
	}

	function resolveSave(outcome: SaveChangesOutcome): void {
		const ids = pendingClose;
		if (outcome === 'cancel') {
			pendingClose = [];
			lastOutcome = `Kept ${ids.length} view${ids.length === 1 ? '' : 's'} open`;
			return;
		}
		if (outcome === 'discard') {
			for (const id of ids) store.closeView(id, { force: true });
			pendingClose = [];
			saveOpen = false;
			lastOutcome = `Discarded changes in ${ids.length} view${ids.length === 1 ? '' : 's'}`;
			return;
		}
		saveBusy = true;
		setTimeout(() => {
			for (const id of ids) {
				store.setDirty(id, false);
				store.closeView(id, { force: true });
			}
			saveBusy = false;
			pendingClose = [];
			saveOpen = false;
			lastOutcome = `Saved and closed ${ids.length} view${ids.length === 1 ? '' : 's'}`;
		}, 700);
	}

	registerWorkbenchCommands(store, { onCloseNeedsConfirmation: (id) => askClose([id]) });
	store.registerCommand({ id: 'workbench.showCommands', title: 'Show All Commands', category: 'View', keybinding: 'mod+shift+p' }, () => { paletteOpen = true; });
	store.registerCommand({ id: 'workbench.quickOpen', title: 'Go to File…', category: 'Navigate', keybinding: 'mod+p' }, () => { quickQuery = ''; quickOpen = true; });
	store.registerCommand({ id: 'file.save', title: 'Save', category: 'File', keybinding: 'mod+s', when: (context) => context.activeViewDirty === true }, () => {
		const id = activeViewId();
		if (id) store.setDirty(id, false);
	});
	store.registerCommand({ id: 'file.edit', title: 'Simulate an Edit', category: 'File', keybinding: 'mod+e', when: (context) => context.hasActiveView === true && context.activeViewDirty !== true }, () => {
		const id = activeViewId();
		if (id) store.setDirty(id, true);
	});
	store.registerCommand({ id: 'workbench.openSettings', title: 'Open Settings', category: 'View', keybinding: 'mod+,' }, () => store.openView(settingsView));
	store.registerCommand({ id: 'workbench.saveSnapshot', title: 'Save Layout Snapshot', category: 'Workbench' }, async () => {
		await snapshotPort.save(store.snapshot());
		snapshotNote = `Snapshot saved at ${new Date().toLocaleTimeString()}`;
	});
	store.registerCommand({ id: 'workbench.restoreSnapshot', title: 'Restore Layout Snapshot', category: 'Workbench' }, async () => {
		const outcome: LoadOutcome = await loadSnapshot(snapshotPort);
		snapshotNote = describeLoad(outcome);
		store.restore(outcome);
	});
	store.registerCommand({ id: 'workbench.resetDemo', title: 'Reset Demo', category: 'Workbench' }, () => seed());
	seed();

	function describeLoad(outcome: LoadOutcome): string {
		switch (outcome.kind) {
			case 'restored':
				return outcome.dropped.length ? `Restored; dropped ${outcome.dropped.join(', ')}` : 'Restored from snapshot';
			case 'empty':
				return 'No snapshot to restore';
			case 'invalid':
				return `Snapshot invalid: ${outcome.reason}`;
			case 'unsupported':
				return `Snapshot version ${String(outcome.version)} is unsupported`;
			case 'unavailable':
				return 'Snapshot port unavailable';
		}
	}

	function activeViewId(): string | null {
		const views = store.state.views;
		return views.groups.find((group) => group.id === views.activeGroupId)?.activeViewId ?? null;
	}

	async function run(id: string): Promise<void> {
		const outcome = await store.dispatch(id);
		lastOutcome = outcome.kind === 'executed' ? `Ran ${id}` : `${id}: ${outcome.kind}`;
	}

	function handleKeydown(event: KeyboardEvent): void {
		const resolution = resolveShortcut(store.dispatcher.registry, event, store.platform, store.context());
		if (resolution.kind !== 'match') return;
		event.preventDefault();
		event.stopPropagation();
		void run(resolution.command.id);
	}

	function tabsFor(groupId: string): EditorTabItem[] {
		return viewsIn($store.views, groupId).map((view) => ({ id: view.id, label: view.title, description: view.description, dirty: view.dirty, pinned: view.pinned, preview: view.preview, closable: view.closable }));
	}

	function shortcutFor(commandId: string): string {
		const binding = store.dispatcher.registry[commandId]?.binding;
		return binding ? formatKeybinding(binding, platform) : '';
	}

	const paletteCommands = $derived.by<readonly CommandItem[]>(() => {
		void registryVersion;
		void $store;
		return paletteEntries(store.dispatcher.registry, store.context(), platform).map((entry) => ({ id: entry.id, label: entry.label, description: entry.group, shortcut: entry.shortcut, disabled: entry.disabled }));
	});
	const quickItems = $derived.by<readonly QuickPickItem[]>(() => {
		const open = new Set(Object.keys($store.views.views));
		return files.map((file) => ({ id: file.id, label: file.title, description: file.description, recent: open.has(file.id), actions: quickActions }));
	});
	const context = $derived.by(() => {
		void $store;
		return store.context();
	});
	const activeGroup = $derived($store.views.groups.find((group) => group.id === $store.views.activeGroupId) ?? null);
	const activeView = $derived(activeGroup?.activeViewId ? ($store.views.views[activeGroup.activeViewId] ?? null) : null);
	const modKey = $derived(formatKeybinding(parseKeybinding('mod+b')!, platform).replace(/B$/, ''));
</script>

{#snippet sidebarSnippet()}
	<PaneStack panes={sidebarPanes} collapsedIds={[...$store.layout.sidebar.collapsedPanes]} weights={$store.layout.sidebar.paneWeights} ariaLabel="Sidebar panels" onToggle={(pane) => store.togglePane(pane.id)} onWeightsChange={(weights) => store.setPaneWeights(weights)}>
		{#snippet actions(item)}<IconButton label={`${item.title} options`} size="sm">…</IconButton>{/snippet}
		{#snippet pane(item)}
			{#if item.id === 'explorer'}<Tree items={treeNodes} bind:selectedId={treeSelection} bind:expandedIds={expandedTreeIds} ariaLabel="Explorer tree" />
			{:else if item.id === 'outline'}<List items={outlineItems} bind:value={outlineSelection} ariaLabel="Outline" />
			{:else}<EmptyState title="No history yet" description="Saved versions of the active file appear here." />{/if}
		{/snippet}
	</PaneStack>
{/snippet}

<div class="workbench-demo-frame">
	<Inline gap="md" wrap class="workbench-demo-controls">
		<Inline gap="xs"><Switch bind:checked={windowActive} aria-label="Window focused" /><Text size="sm" tone="muted">Window focused</Text></Inline>
		<Button size="sm" variant="secondary" onclick={() => run('workbench.quickOpen')}>Go to file… <kbd class="kbd">{shortcutFor('workbench.quickOpen')}</kbd></Button>
		<Button size="sm" variant="secondary" onclick={() => run('workbench.showCommands')}>Commands <kbd class="kbd">{shortcutFor('workbench.showCommands')}</kbd></Button>
		<Button size="sm" variant="quiet" onclick={() => run('workbench.saveSnapshot')}>Save snapshot</Button>
		<Button size="sm" variant="quiet" onclick={() => run('workbench.restoreSnapshot')}>Restore snapshot</Button>
		<Button size="sm" variant="quiet" onclick={() => run('workbench.resetDemo')}>Reset</Button>
	</Inline>
	<Text size="sm" tone="quiet" class="workbench-demo-note">{session ? `Active workspace: ${session.workspace.name} · ${session.scope}` : 'No active workspace — choose Open workspace to activate this Workbench'}</Text>
	<Text size="sm" tone="quiet" class="workbench-demo-note">{lastOutcome} · {snapshotNote} · shortcuts use {modKey} on this platform</Text>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="workbench-demo-focus" onkeydown={handleKeydown}>
		<WorkbenchShell class="workbench-demo" {windowActive} sidebar={$store.layout.sidebar.visible ? sidebarSnippet : undefined}>
			{#snippet titlebar()}
				<TitleBar title={session?.workspace.name ?? 'Atelier workspace'} subtitle={session ? `${host} · ${session.workspace.location}` : `${host} · no workspace open`} inset="macos" active={windowActive} {dragRegion} {dragExclude}>
					{#snippet leading()}<IconButton label="Go back" size="sm" disabled={context.canGoBack !== true} onclick={() => run(WORKBENCH_COMMANDS.back)}>‹</IconButton><IconButton label="Go forward" size="sm" disabled={context.canGoForward !== true} onclick={() => run(WORKBENCH_COMMANDS.forward)}>›</IconButton>{/snippet}
					{#snippet trailing()}<IconButton label="Toggle sidebar" size="sm" onclick={() => run(WORKBENCH_COMMANDS.toggleSidebar)}>▤</IconButton><IconButton label="Split editor right" size="sm" onclick={() => run(WORKBENCH_COMMANDS.splitEditor)}>⫿</IconButton><IconButton label="Open command palette" size="sm" onclick={() => run('workbench.showCommands')}>⌘</IconButton>{/snippet}
				</TitleBar>
			{/snippet}
			{#snippet activity()}
				<ActivityRail items={activityItems} activeId={$store.layout.sidebar.visible ? ($store.layout.sidebar.activityId ?? '') : ''} ariaLabel="Workbench activities" onSelect={(item) => store.selectActivity(item.id)}>
					{#snippet icon(item)}<Icon size="lg">{#if item.id === 'files'}<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />{:else if item.id === 'search'}<circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" />{:else if item.id === 'changes'}<path d="M6 3v12" /><circle cx="6" cy="18" r="3" /><circle cx="18" cy="6" r="3" /><path d="M18 9a9 9 0 0 1-9 9" />{:else}<circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 3" />{/if}</Icon>{/snippet}
					{#snippet end()}<ActivityItem label="Settings" selectable={false} onclick={() => run('workbench.openSettings')}><Icon size="lg"><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" /></Icon></ActivityItem>{/snippet}
				</ActivityRail>
			{/snippet}
			{#snippet main()}
				<SplitGroup sizes={[...$store.layout.editorSizes]} minSize={20} collapsible ariaLabel="Editor groups" class="workbench-editor-groups" onChange={(sizes) => store.setEditorSizes(sizes)}>
					{#snippet pane(index)}
						{@const group = $store.views.groups[index]}
						{#if group}
							{@const active = group.activeViewId ? $store.views.views[group.activeViewId] : null}
							<!-- svelte-ignore a11y_no_static_element_interactions -->
							<div class="workbench-editor-group" data-active-group={group.id === $store.views.activeGroupId || undefined} onfocusin={() => store.setActiveGroup(group.id)} onpointerdown={() => store.setActiveGroup(group.id)}>
								<EditorTabs tabs={tabsFor(group.id)} activeId={group.activeViewId ?? ''} ariaLabel={`Editor group ${index + 1}`} onActivate={(tab) => store.activateView(tab.id)} onClose={(tab) => { const request = store.requestClose(tab.id); if (request.kind === 'needs-confirmation') askClose([tab.id]); }} onCloseOthers={(tab) => askClose(store.closeOthers(tab.id).map((view) => view.id))} onCloseRight={(tab) => askClose(store.closeToRight(tab.id).map((view) => view.id))} onCloseAll={() => askClose(store.closeGroupViews(group.id).map((view) => view.id))} onPin={(tab, pinned) => store.pinView(tab.id, pinned)} onReorder={(from, to) => store.reorderView(group.id, from, to)}>
									{#snippet icon(tab)}<Icon size="xs">{#if tab.id === 'settings'}<circle cx="12" cy="12" r="3" /><circle cx="12" cy="12" r="9" />{:else if tab.label.endsWith('.css')}<path d="M4 4h16v16H4z" /><path d="M8 12h8" />{:else}<path d="M6 3h8l4 4v14H6z" /><path d="M14 3v4h4" />{/if}</Icon>{/snippet}
									{#snippet trailing()}<IconButton label="Split editor right" size="sm" onclick={() => store.splitGroup(group.id)}>⫿</IconButton>{#if $store.views.groups.length > 1}<IconButton label="Close group" size="sm" onclick={() => store.removeGroup(group.id)}>×</IconButton>{/if}{/snippet}
								</EditorTabs>
								<div class="workbench-editor-body" role="tabpanel" aria-labelledby={active ? `editor-tab-${active.id}` : undefined} tabindex="-1">
									{#if active}
										<span class="preview-kicker">{active.description ?? active.kind}{#if active.preview} · preview{/if}{#if active.dirty} · unsaved{/if}</span>
										<strong>{active.title}</strong>
										<p>{bodies.get(active.id) ?? 'Content is owned by the feature that opened this view.'}</p>
										<Inline gap="sm" wrap><Button size="sm" variant="secondary" disabled={active.dirty} onclick={() => run('file.edit')}>Edit <kbd class="kbd">{shortcutFor('file.edit')}</kbd></Button><Button size="sm" variant="primary" disabled={!active.dirty} onclick={() => run('file.save')}>Save <kbd class="kbd">{shortcutFor('file.save')}</kbd></Button><Button size="sm" variant="quiet" onclick={() => run(WORKBENCH_COMMANDS.closeEditor)}>Close <kbd class="kbd">{shortcutFor(WORKBENCH_COMMANDS.closeEditor)}</kbd></Button></Inline>
									{:else}
										<EmptyState title="No editor open" description="Use Go to file, or drag a tab here." />
									{/if}
								</div>
							</div>
						{/if}
					{/snippet}
				</SplitGroup>
			{/snippet}
			{#snippet status()}
				<StatusBar>
					{#snippet start()}<StatusItem tone={activeView?.dirty ? 'warning' : 'success'} label="Save state">{activeView?.dirty ? 'Unsaved changes' : 'All saved'}</StatusItem><StatusItem label="Branch">{#snippet icon()}<Icon size="xs"><path d="M6 3v12" /><circle cx="6" cy="18" r="3" /><circle cx="18" cy="6" r="3" /><path d="M18 9a9 9 0 0 1-9 9" /></Icon>{/snippet}main</StatusItem>{/snippet}
					{#snippet end()}<StatusItem label="Open views">{Object.keys($store.views.views).length} open · {$store.views.groups.length} group{$store.views.groups.length === 1 ? '' : 's'}</StatusItem><StatusItem label="Panel" onclick={() => run(WORKBENCH_COMMANDS.togglePanel)}>Panel {$store.layout.panel.visible ? 'on' : 'off'}</StatusItem><StatusItem label="Active view kind">{activeView?.kind ?? 'none'}</StatusItem>{/snippet}
				</StatusBar>
			{/snippet}
		</WorkbenchShell>
	</div>
	<CommandPalette bind:open={paletteOpen} commands={paletteCommands} onExecute={(command) => void run(command.id)} />
	<QuickPick bind:open={quickOpen} bind:query={quickQuery} items={quickItems} title="Go to file" placeholder="Search files by name" recentLabel="Open now" onSelect={(item) => { const file = files.find((entry) => entry.id === item.id); if (file) store.openView(file, { preview: true }); lastOutcome = `Opened ${item.label}`; }} onAction={(item) => { const file = files.find((entry) => entry.id === item.id); if (!file) return; const groupId = store.splitGroup(); store.openView(file, { groupId }); quickOpen = false; lastOutcome = `Opened ${item.label} to the side`; }}>
		{#snippet icon(item)}<Icon size="sm">{#if item.label.endsWith('.css')}<path d="M4 4h16v16H4z" /><path d="M8 12h8" />{:else if item.label.endsWith('.md')}<path d="M4 6h16v12H4z" /><path d="m8 15 0-6 2 3 2-3v6" /><path d="M16 9v6" />{:else}<path d="M6 3h8l4 4v14H6z" /><path d="M14 3v4h4" />{/if}</Icon>{/snippet}
	</QuickPick>
	<SaveChangesDialog bind:open={saveOpen} items={pendingClose.map((id) => $store.views.views[id]?.title ?? id)} busy={saveBusy} onResolve={resolveSave} />
</div>
