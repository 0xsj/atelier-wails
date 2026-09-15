<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Resizable } from '../layout';
	import './patterns.css';

	interface Props {
		master: Snippet;
		detail: Snippet;
		empty?: Snippet;
		hasSelection?: boolean;
		masterLabel?: string;
		detailLabel?: string;
		value?: number;
		min?: number;
		max?: number;
		class?: string;
	}

	let { master, detail, empty, hasSelection = true, masterLabel = 'Items', detailLabel = 'Detail', value = $bindable(32), min = 20, max = 60, class: className }: Props = $props();
</script>

<Resizable bind:value {min} {max} class="atelier-master-detail {className ?? ''}">
	{#snippet first()}<aside class="atelier-master-detail__master" aria-label={masterLabel}>{@render master()}</aside>{/snippet}
	{#snippet second()}
		<section class="atelier-master-detail__detail" aria-label={detailLabel} aria-live="polite">
			{#if hasSelection}{@render detail()}{:else if empty}<div class="atelier-master-detail__empty">{@render empty()}</div>{/if}
		</section>
	{/snippet}
</Resizable>
