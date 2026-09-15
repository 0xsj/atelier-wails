<script lang="ts">
	import ChoiceDialog, { type ChoiceAction } from './ChoiceDialog.svelte';
	import './patterns.css';

	export type SaveChangesOutcome = 'save' | 'discard' | 'cancel';

	interface Props {
		open?: boolean;
		items?: readonly string[];
		title?: string;
		description?: string;
		saveLabel?: string;
		discardLabel?: string;
		cancelLabel?: string;
		busy?: boolean;
		onResolve?: (outcome: SaveChangesOutcome) => void;
		class?: string;
	}

	let { open = $bindable(false), items = [], title, description, saveLabel, discardLabel = "Don't Save", cancelLabel = 'Cancel', busy = false, onResolve, class: className }: Props = $props();

	const count = $derived(items.length);
	const resolvedTitle = $derived(title ?? (count === 1 ? `Save changes to ${items[0]}?` : count > 1 ? `Save changes to ${count} files?` : 'Save changes?'));
	const resolvedDescription = $derived(description ?? "Your changes will be lost if you don't save them.");
	const resolvedSaveLabel = $derived(saveLabel ?? (count > 1 ? 'Save All' : 'Save'));
	const actions = $derived<readonly ChoiceAction[]>([
		{ id: 'discard', label: discardLabel, variant: 'secondary', side: 'start', key: 'mod+d' },
		{ id: 'cancel', label: cancelLabel, variant: 'secondary' },
		{ id: 'save', label: resolvedSaveLabel, variant: 'primary', primary: true }
	]);

	function resolve(outcome: string): void {
		onResolve?.(outcome === 'save' || outcome === 'discard' ? outcome : 'cancel');
	}
</script>

<ChoiceDialog bind:open title={resolvedTitle} description={resolvedDescription} {actions} {busy} onResolve={resolve} class={className}>
	{#if count > 1}
		<ul class="atelier-save-changes__list" aria-label="Files with unsaved changes">
			{#each items as item (item)}<li>{item}</li>{/each}
		</ul>
	{/if}
</ChoiceDialog>
