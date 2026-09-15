<script lang="ts">
	// Workspace selection is a runtime action over registry metadata. The
	// picker never inspects files; root supplies the transport and receives the
	// selected record when the user chooses Open.
	import { onMount } from 'svelte';
	import { Alert, Badge, EmptyState } from '../../components/feedback';
	import { List, type ListEntry } from '../../components/collections';
	import { Card, KeyValue } from '../../components/data-display';
	import { Button, Text } from '../../components/primitives';
	import { Inline, Stack } from '../../components/layout';
	import { parseWorkspaceId, type Failure } from '../../kernel';
	import { listWorkspaces, readWorkspace, type WorkspaceFilter, type WorkspaceOpenPort, type WorkspaceSession, type WorkspaceSnapshot, type WorkspaceTransport } from '../../services/workspace';

	interface Props {
		transport: WorkspaceTransport;
		openPort: WorkspaceOpenPort;
		storeLabel: string;
		onOpen?: (session: WorkspaceSession) => void | Promise<void>;
	}

	let { transport, openPort, storeLabel, onOpen }: Props = $props();
	let filter = $state<WorkspaceFilter>('active');
	let workspaces = $state<readonly WorkspaceSnapshot[]>([]);
	let selectedId = $state('');
	let selected = $state<WorkspaceSnapshot | null>(null);
	let loading = $state(true);
	let opening = $state(false);
	let failure = $state<Failure | null>(null);
	let note = $state('Select a workspace to inspect its registry metadata');

	const filterOptions = [
		{ value: 'active', label: 'Active only' },
		{ value: 'all', label: 'Active and archived' }
	] as const;
	const listItems = $derived<readonly ListEntry[]>(workspaces.map((workspace) => ({
		id: workspace.id,
		label: workspace.name,
		description: workspace.location,
		meta: workspace.status === 'archived' ? 'archived' : `r${workspace.revision}`
	})));

	async function refresh(): Promise<void> {
		loading = true;
		const result = await listWorkspaces(transport, filter);
		loading = false;
		if (!result.ok) {
			failure = result.error;
			return;
		}
		failure = null;
		workspaces = result.value;
		if (selectedId && !workspaces.some((workspace) => workspace.id === selectedId)) {
			selectedId = '';
			selected = null;
			note = 'Select a workspace to inspect its registry metadata';
		}
	}

	async function choose(item: ListEntry): Promise<void> {
		const id = parseWorkspaceId(item.id);
		if (!id) return;
		selectedId = id;
		const result = await readWorkspace(transport, id);
		if (!result.ok) {
			selected = null;
			failure = result.error;
			return;
		}
		failure = null;
		if (!result.value.present) {
			selected = null;
			note = 'This workspace is no longer registered';
			return;
		}
		selected = result.value.value;
		note = `${selected.name} selected · ${selected.status}`;
	}

	function changeFilter(event: Event): void {
		const value = (event.currentTarget as HTMLSelectElement).value;
		filter = value === 'all' ? 'all' : 'active';
		void refresh();
	}

	async function open(): Promise<void> {
		if (!selected || opening) return;
		opening = true;
		const result = await openPort.open(selected);
		opening = false;
		if (!result.ok) {
			failure = result.error;
			note = `Could not open ${selected.name}`;
			return;
		}
		failure = null;
		await onOpen?.(result.value);
		note = `Opened ${result.value.workspace.name}`;
	}

	onMount(() => {
		void refresh();
	});
</script>

<Stack gap="md" class="workspace-picker">
	<Inline gap="sm" wrap justify="between">
		<Inline gap="xs"><Badge tone="warning">preview</Badge><Text size="sm" tone="muted">{storeLabel}</Text></Inline>
		<Inline gap="xs" align="end">
			<label class="workspace-picker__filter"><span class="field-label">Show</span><select class="field" value={filter} onchange={changeFilter} aria-label="Workspace filter">{#each filterOptions as option}<option value={option.value}>{option.label}</option>{/each}</select></label>
			<Button size="sm" variant="quiet" onclick={refresh} disabled={loading}>Refresh</Button>
		</Inline>
	</Inline>
	{#if loading && workspaces.length === 0}
		<Text size="sm" tone="quiet">Loading workspaces…</Text>
	{:else if failure}
		<Alert tone="danger" title="Workspace registry unavailable">{failure.message}</Alert>
	{:else if workspaces.length === 0}
		<EmptyState title="No workspaces in this view" description={filter === 'active' ? 'Archived workspaces are hidden from the default list.' : 'Register a workspace to make it available here.'} />
	{:else}
		<List items={listItems} value={selectedId} onSelect={(item) => void choose(item)} ariaLabel="Registered workspaces" />
	{/if}
	{#if selected}
		<Card title={selected.name} description={selected.status === 'active' ? 'Active registry record' : 'Archived registry record'}>
			{#snippet actions()}<Button variant="primary" size="sm" onclick={() => void open()} loading={opening} disabled={opening}>Open workspace</Button>{/snippet}
			<KeyValue columns={2} items={[{ label: 'Location', value: selected.location }, { label: 'Status', value: selected.status }, { label: 'Revision', value: selected.revision }, { label: 'Updated', value: selected.updatedAt }]} />
		</Card>
	{/if}
	<Text size="sm" tone="quiet">{note}</Text>
</Stack>
