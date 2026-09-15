<script lang="ts">
	import type { Snippet } from 'svelte';
	import { DropdownMenu as Primitive } from 'bits-ui';
	import './navigation.css';

	type Align = 'start' | 'center' | 'end';
	interface Props { trigger: Snippet; children: Snippet; open?: boolean; align?: Align; sideOffset?: number; class?: string; }
	let { trigger, children, open = $bindable(false), align = 'start', sideOffset = 6, class: className }: Props = $props();
</script>
<Primitive.Root bind:open><Primitive.Trigger class="atelier-menu__trigger {className ?? ''}">{@render trigger()}</Primitive.Trigger><Primitive.Portal><Primitive.Content {align} {sideOffset} class="atelier-menu__content">{@render children()}</Primitive.Content></Primitive.Portal></Primitive.Root>
