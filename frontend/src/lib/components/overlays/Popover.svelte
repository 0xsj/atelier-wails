<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Popover as Primitive } from 'bits-ui';
	import './overlays.css';

	type Side = 'top' | 'right' | 'bottom' | 'left';
	type Align = 'start' | 'center' | 'end';

	interface Props {
		trigger: Snippet;
		children: Snippet;
		open?: boolean;
		side?: Side;
		align?: Align;
		sideOffset?: number;
		class?: string;
	}

	let { trigger, children, open = $bindable(false), side = 'bottom', align = 'start', sideOffset = 6, class: className }: Props = $props();
</script>

<Primitive.Root bind:open>
	<Primitive.Trigger class="atelier-popover__trigger {className ?? ''}">
		{@render trigger()}
	</Primitive.Trigger>
	<Primitive.Portal>
		<Primitive.Content {side} {align} {sideOffset} class="atelier-popover__content">
			{@render children()}
			<Primitive.Arrow class="atelier-popover__arrow" />
		</Primitive.Content>
	</Primitive.Portal>
</Primitive.Root>
