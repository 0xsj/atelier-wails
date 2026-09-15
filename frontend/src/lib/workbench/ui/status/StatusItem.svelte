<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Spinner } from '../../../components/feedback';
	import './status.css';

	type StatusTone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger';

	interface Props {
		tone?: StatusTone;
		busy?: boolean;
		label?: string;
		icon?: Snippet;
		onclick?: (event: MouseEvent) => void;
		children?: Snippet;
		class?: string;
	}

	let { tone = 'neutral', busy = false, label, icon, onclick, children, class: className }: Props = $props();
</script>

{#if onclick}
	<button type="button" class="atelier-status-item {className ?? ''}" data-tone={tone} data-interactive aria-label={label} title={label} {onclick}>
		{#if busy}<Spinner size="sm" label={label ?? 'Working'} />{:else if icon}<span class="atelier-status-item__icon" aria-hidden="true">{@render icon()}</span>{/if}
		{#if children}<span class="atelier-status-item__text">{@render children()}</span>{/if}
	</button>
{:else}
	<span class="atelier-status-item {className ?? ''}" data-tone={tone} role={busy ? 'status' : undefined} aria-label={label} title={label}>
		{#if busy}<Spinner size="sm" label={label ?? 'Working'} />{:else if icon}<span class="atelier-status-item__icon" aria-hidden="true">{@render icon()}</span>{/if}
		{#if children}<span class="atelier-status-item__text">{@render children()}</span>{/if}
	</span>
{/if}
