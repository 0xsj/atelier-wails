<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import ListItem from './ListItem.svelte';
	import './collections.css';

	export type ListEntry = { id: string; label: string; description?: string; meta?: string; disabled?: boolean };
	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'class' | 'role'> { items: readonly ListEntry[]; value?: string; ariaLabel?: string; onSelect?: (item: ListEntry) => void; class?: string; }
	let { items, value = $bindable(''), ariaLabel = 'List', onSelect, class: className, ...rest }: Props = $props();

	function select(item: ListEntry): void { if (item.disabled) return; value = item.id; onSelect?.(item); }
</script>

<div {...rest} class="atelier-list {className ?? ''}" role="listbox" aria-label={ariaLabel}>
	{#each items as item (item.id)}<ListItem {...item} selected={value === item.id} onSelect={() => select(item)} />{/each}
</div>
