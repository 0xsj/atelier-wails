<script lang="ts">
	import type { Snippet } from 'svelte';
	import './panels.css';

	interface Props {
		title?: string;
		actions?: Snippet;
		children: Snippet;
		scrollable?: boolean;
		class?: string;
	}

	let { title, actions, children, scrollable = true, class: className }: Props = $props();
</script>

<section class="atelier-workbench-panel {className ?? ''}" data-scrollable={scrollable || undefined}>
	{#if title || actions}<header class="atelier-workbench-panel__header">{#if title}<h3>{title}</h3>{/if}{#if actions}<div>{@render actions()}</div>{/if}</header>{/if}
	<div class="atelier-workbench-panel__body">{@render children()}</div>
</section>
