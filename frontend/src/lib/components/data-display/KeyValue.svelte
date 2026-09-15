<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import './data-display.css';

	export type KeyValueItem = { label: string; value: string | number; description?: string };

	interface Props extends Omit<HTMLAttributes<HTMLDListElement>, 'children' | 'class'> {
		items: readonly KeyValueItem[];
		columns?: 1 | 2;
		class?: string;
	}

	let { items, columns = 1, class: className, ...rest }: Props = $props();
</script>

<dl {...rest} class="atelier-key-value {className ?? ''}" data-columns={columns}>
	{#each items as item (item.label)}
		<div class="atelier-key-value__item">
			<dt>{item.label}</dt>
			<dd>{item.value}</dd>
			{#if item.description}<small>{item.description}</small>{/if}
		</div>
	{/each}
</dl>
