<script lang="ts">
	import type { Snippet } from 'svelte';
	import './shell.css';

	interface Props {
		label: string;
		active?: boolean;
		selectable?: boolean;
		focusable?: boolean;
		badge?: number | string;
		disabled?: boolean;
		onclick?: (event: MouseEvent) => void;
		children: Snippet;
		class?: string;
	}

	let { label, active = false, selectable = true, focusable = true, badge, disabled = false, onclick, children, class: className }: Props = $props();
	const badgeText = $derived(typeof badge === 'number' && badge > 99 ? '99+' : badge);
</script>

<button
	type="button"
	class="atelier-activity-item {className ?? ''}"
	role={selectable ? 'tab' : undefined}
	aria-selected={selectable ? active : undefined}
	aria-label={badge !== undefined && badge !== '' ? `${label}, ${badge}` : label}
	title={label}
	tabindex={focusable ? 0 : -1}
	data-active={active || undefined}
	{disabled}
	{onclick}
>
	<span class="atelier-activity-item__icon" aria-hidden="true">{@render children()}</span>
	{#if badge !== undefined && badge !== ''}<span class="atelier-activity-item__badge" aria-hidden="true">{badgeText}</span>{/if}
</button>
