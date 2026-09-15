<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import './collections.css';

	interface Props extends Omit<HTMLButtonAttributes, 'children' | 'class' | 'value'> { id: string; label: string; description?: string; meta?: string; selected?: boolean; disabled?: boolean; onSelect?: () => void; class?: string; }
	let { label, description, meta, selected = false, disabled = false, onSelect, class: className, ...rest }: Props = $props();
</script>

<button {...rest} type="button" class="atelier-list-item {className ?? ''}" disabled={disabled} role="option" aria-selected={selected} data-state={selected ? 'selected' : 'idle'} onclick={onSelect}>
	<span class="atelier-list-item__copy"><strong>{label}</strong>{#if description}<small>{description}</small>{/if}</span>
	{#if meta}<span class="atelier-list-item__meta">{meta}</span>{/if}
</button>
