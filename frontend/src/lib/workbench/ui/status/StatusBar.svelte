<script lang="ts">
	import type { Snippet } from 'svelte';
	import './status.css';

	interface Props {
		start?: Snippet;
		end?: Snippet;
		class?: string;
	}

	let { start, end, class: className }: Props = $props();
</script>

<div class="atelier-status-bar {className ?? ''}">
	<div class="atelier-status-bar__group">{@render start?.()}</div>
	<div class="atelier-status-bar__group" data-side="end">{@render end?.()}</div>
</div>
