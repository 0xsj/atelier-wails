<script lang="ts">
	// Visible round trip for the Preferences slice: list, set and clear entries
	// through the service and show every outcome, including refusals, commit
	// state and reconcile attempts. Values and callbacks only; root supplies
	// the transport.
	import { onMount } from 'svelte';
	import { Button, Text } from '../../components/primitives';
	import { Inline, Stack } from '../../components/layout';
	import { Field, Input, Select, Switch } from '../../components/forms';
	import { Alert, Badge, EmptyState } from '../../components/feedback';
	import { Card, Table, type TableColumn, type TableRow } from '../../components/data-display';
	import { fieldProblems, type Failure } from '../../kernel';
	import { clearPreference, GLOBAL_SCOPE, isValidKey, listPreferences, setPreference, type PreferenceEntry, type PreferenceValue, type PreferencesTransport } from '../../services/preferences';
	import type { PreviewPreferencesTransport } from '../../platform/preview/preferences';

	interface Props {
		transport: PreferencesTransport;
		storeLabel: string;
		preview?: PreviewPreferencesTransport | null;
		onChanged?: () => void;
	}

	let { transport, storeLabel, preview = null, onChanged }: Props = $props();

	let entries = $state<readonly PreferenceEntry[]>([]);
	let loading = $state(true);
	let busy = $state(false);
	let key = $state('editor.font_size');
	let kind = $state<'text' | 'bool' | 'int'>('int');
	let textValue = $state('13');
	let boolValue = $state(true);
	let note = $state('Nothing written yet');
	let lastFailure = $state<Failure | null>(null);

	const kinds = [{ value: 'text', label: 'Text' }, { value: 'bool', label: 'Boolean' }, { value: 'int', label: 'Integer' }] as const;
	const columns: readonly TableColumn[] = [
		{ key: 'key', label: 'Key' },
		{ key: 'kind', label: 'Kind' },
		{ key: 'value', label: 'Value' },
		{ key: 'revision', label: 'Revision', align: 'end' }
	];
	const rows = $derived<readonly TableRow[]>(entries.map((entry) => ({ id: entry.key, key: entry.key, kind: entry.value.kind, value: display(entry.value), revision: entry.revision })));
	const removable = $derived(entries.map((entry) => ({ value: entry.key, label: entry.key })));
	let removeKey = $state('');
	const keyValid = $derived(isValidKey(key));
	const intValid = $derived(kind !== 'int' || /^-?[0-9]+$/.test(textValue.trim()));

	function display(value: PreferenceValue): string {
		switch (value.kind) {
			case 'text':
				return value.text === '' ? '(empty text)' : value.text;
			case 'bool':
				return value.value ? 'true' : 'false';
			case 'int':
				return value.value.toString(10);
		}
	}

	function describe(failure: Failure): string {
		const fields = fieldProblems(failure).map(([path, problem]) => `${path}: ${problem}`).join(', ');
		return `${failure.kind}${failure.type ? ` · ${failure.type}` : ''} · commit ${failure.commit}${fields ? ` · ${fields}` : ''} · ${failure.message}`;
	}

	async function refresh(): Promise<void> {
		loading = true;
		const result = await listPreferences(transport, GLOBAL_SCOPE);
		loading = false;
		if (result.ok) {
			entries = result.value;
			lastFailure = null;
		} else {
			lastFailure = result.error;
		}
	}

	function composeValue(): PreferenceValue | null {
		if (kind === 'text') return { kind: 'text', text: textValue };
		if (kind === 'bool') return { kind: 'bool', value: boolValue };
		return intValid ? { kind: 'int', value: BigInt(textValue.trim()) } : null;
	}

	async function save(): Promise<void> {
		const value = composeValue();
		if (!value) return;
		busy = true;
		const outcome = await setPreference(transport, GLOBAL_SCOPE, key, value);
		busy = false;
		if (outcome.ok) {
			lastFailure = null;
			note = `${outcome.value.status} · ${outcome.value.entry.key} at revision ${outcome.value.entry.revision} after ${outcome.value.attempts} attempt${outcome.value.attempts === 1 ? '' : 's'}`;
			onChanged?.();
		} else {
			lastFailure = outcome.error;
			note = `Save refused: ${describe(outcome.error)}`;
		}
		await refresh();
	}

	async function remove(target: string): Promise<void> {
		busy = true;
		const outcome = await clearPreference(transport, GLOBAL_SCOPE, target);
		busy = false;
		if (outcome.ok) {
			lastFailure = null;
			note = outcome.value.removed ? `removed ${target}${outcome.value.revision ? ` at revision ${outcome.value.revision}` : ''}${outcome.value.reconciled ? ' (reconciled after an uncertain commit)' : ''}` : `${target} was already absent`;
			onChanged?.();
		} else {
			lastFailure = outcome.error;
			note = `Remove refused: ${describe(outcome.error)}`;
		}
		await refresh();
	}

	function queueFault(mode: 'fail-before' | 'lose-acknowledgement'): void {
		preview?.failNext({ kind: mode === 'fail-before' ? 'unavailable' : 'timeout', mode, operation: 'replace' });
		note = mode === 'fail-before' ? 'Next replace will fail before any effect; the service re-reads and retries once' : 'Next replace will apply but lose its acknowledgement; the service re-reads and reports reconciled';
	}

	onMount(() => {
		void refresh();
	});
</script>

<Stack gap="md" class="preferences-panel">
	<Inline gap="sm" wrap justify="between">
		<Inline gap="xs"><Badge tone={preview ? 'warning' : 'success'}>{preview ? 'preview' : 'native'}</Badge><Text size="sm" tone="muted">{storeLabel}</Text></Inline>
		<Inline gap="xs"><Button size="sm" variant="quiet" onclick={refresh} disabled={loading}>Refresh</Button>{#if preview}<Button size="sm" variant="quiet" onclick={() => queueFault('fail-before')}>Fault: fail before</Button><Button size="sm" variant="quiet" onclick={() => queueFault('lose-acknowledgement')}>Fault: lose ack</Button>{/if}</Inline>
	</Inline>
	<Card title="Global scope">
		{#if loading && entries.length === 0}
			<Text size="sm" tone="quiet">Loading…</Text>
		{:else if entries.length === 0}
			<EmptyState title="No preferences stored" description="Save one below. Theme and density choices in the header land here too." />
		{:else}
			<Stack gap="md">
				<Table {columns} {rows} ariaLabel="Stored preferences" compact />
				<Inline gap="sm" wrap align="end"><Field label="Remove a key">{#snippet children({ id })}<Select {id} options={removable} bind:value={removeKey} placeholder="Choose a key" size="sm" aria-label="Key to remove" />{/snippet}</Field><Button size="sm" variant="danger" disabled={busy || !removeKey} onclick={() => remove(removeKey)}>Remove</Button></Inline>
			</Stack>
		{/if}
	</Card>
	<Card title="Set a preference">
		<Stack gap="md">
			<Inline gap="md" wrap align="start">
				<Field label="Key" description="Lowercase, up to 128 characters" message={keyValid ? undefined : 'Key must match [a-z][a-z0-9_.-]{0,127}'} tone={keyValid ? 'hint' : 'error'}>{#snippet children({ id, describedBy, invalid })}<Input {id} aria-describedby={describedBy} {invalid} bind:value={key} placeholder="editor.font_size" />{/snippet}</Field>
				<Field label="Kind">{#snippet children({ id })}<Select {id} options={kinds} bind:value={kind} aria-label="Value kind" />{/snippet}</Field>
				{#if kind === 'bool'}
					<Field label="Value">{#snippet children({ id })}<Switch {id} bind:checked={boolValue} aria-label="Boolean value" />{/snippet}</Field>
				{:else}
					<Field label="Value" message={intValid ? undefined : 'Integers are decimal digits with an optional leading minus'} tone={intValid ? 'hint' : 'error'}>{#snippet children({ id, describedBy, invalid })}<Input {id} aria-describedby={describedBy} {invalid} bind:value={textValue} placeholder={kind === 'int' ? '13' : 'text'} />{/snippet}</Field>
				{/if}
			</Inline>
			<Inline gap="sm" wrap><Button variant="primary" loading={busy} disabled={!keyValid || !intValid} onclick={save}>Save</Button><Text size="sm" tone="quiet">{note}</Text></Inline>
			{#if lastFailure}<Alert tone="danger" title={`Refused as ${lastFailure.kind}`}>{describe(lastFailure)}</Alert>{/if}
		</Stack>
	</Card>
</Stack>
