<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ContextMenu as Primitive } from 'bits-ui';
	import './navigation.css';

	interface Props { trigger: Snippet; children: Snippet; open?: boolean; class?: string; }
	let { trigger, children, open = $bindable(false), class: className }: Props = $props();
</script>
<Primitive.Root bind:open><Primitive.Trigger class="atelier-context-menu__trigger {className ?? ''}">{@render trigger()}</Primitive.Trigger><Primitive.Portal><Primitive.Content class="atelier-menu__content">{@render children()}</Primitive.Content></Primitive.Portal></Primitive.Root>
