<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ScrollArea as Primitive } from 'bits-ui';
	import './collections.css';

	interface Props { children: Snippet; horizontal?: boolean; type?: 'hover' | 'scroll' | 'auto' | 'always'; class?: string; }
	let { children, horizontal = false, type = 'auto', class: className }: Props = $props();
</script>

<Primitive.Root {type} class="atelier-scroll-area {className ?? ''}">
	<Primitive.Viewport class="atelier-scroll-area__viewport">{@render children()}</Primitive.Viewport>
	<Primitive.Scrollbar orientation="vertical" class="atelier-scroll-area__scrollbar"><Primitive.Thumb class="atelier-scroll-area__thumb" /></Primitive.Scrollbar>
	{#if horizontal}<Primitive.Scrollbar orientation="horizontal" class="atelier-scroll-area__scrollbar"><Primitive.Thumb class="atelier-scroll-area__thumb" /></Primitive.Scrollbar><Primitive.Corner />{/if}
</Primitive.Root>
