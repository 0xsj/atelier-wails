<script lang="ts">
	import { Toolbar as Primitive } from 'bits-ui';
	import type { ToggleOption } from './ToggleGroup.svelte';
	import './navigation.css';

	type ToolbarGroupType = 'single' | 'multiple';

	interface Props {
		items: readonly ToggleOption[];
		type?: ToolbarGroupType;
		value?: string | string[];
		ariaLabel?: string;
		class?: string;
	}

	let { items, type = 'single', value = $bindable(''), ariaLabel, class: className }: Props = $props();
</script>

{#if type === 'multiple'}
	<Primitive.Group type="multiple" value={Array.isArray(value) ? value : []} aria-label={ariaLabel} class="atelier-toolbar__group {className ?? ''}" onValueChange={(next: string[]) => (value = next)}>
		{#each items as item (item.value)}<Primitive.GroupItem value={item.value} disabled={item.disabled} class="atelier-toolbar__item">{item.label}</Primitive.GroupItem>{/each}
	</Primitive.Group>
{:else}
	<Primitive.Group type="single" value={typeof value === 'string' ? value : ''} aria-label={ariaLabel} class="atelier-toolbar__group {className ?? ''}" onValueChange={(next: string) => (value = next)}>
		{#each items as item (item.value)}<Primitive.GroupItem value={item.value} disabled={item.disabled} class="atelier-toolbar__item">{item.label}</Primitive.GroupItem>{/each}
	</Primitive.Group>
{/if}
