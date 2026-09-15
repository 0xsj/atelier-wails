<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Tabs as Primitive } from 'bits-ui';
	import './navigation.css';

	export type TabItem = {
		value: string;
		label: string;
		disabled?: boolean;
	};

	type TabsSize = 'sm' | 'md';

	interface Props {
		items: readonly TabItem[];
		value?: string;
		size?: TabsSize;
		activationMode?: 'automatic' | 'manual';
		children?: Snippet;
		class?: string;
	}

	let { items, value = $bindable(''), size = 'md', activationMode = 'automatic', children, class: className }: Props = $props();
</script>

<Primitive.Root bind:value {activationMode} class="atelier-tabs {className ?? ''}" data-size={size}>
	<Primitive.List class="atelier-tabs__list">
		{#each items as item (item.value)}
			<Primitive.Trigger value={item.value} disabled={item.disabled} class="atelier-tabs__trigger">{item.label}</Primitive.Trigger>
		{/each}
	</Primitive.List>
	{@render children?.()}
</Primitive.Root>
