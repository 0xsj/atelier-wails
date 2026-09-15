<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Tooltip as Primitive } from 'bits-ui';
	import './overlays.css';

	type Side = 'top' | 'right' | 'bottom' | 'left';

	interface Props {
		content: string;
		delayDuration?: number;
		side?: Side;
		sideOffset?: number;
		children: Snippet;
		class?: string;
	}

	let { content, delayDuration = 500, side = 'top', sideOffset = 6, children, class: className }: Props = $props();
</script>

<Primitive.Provider {delayDuration}>
	<Primitive.Root>
		<Primitive.Trigger class="atelier-tooltip__trigger {className ?? ''}">
			{@render children()}
		</Primitive.Trigger>
		<Primitive.Content {side} {sideOffset} class="atelier-tooltip__content">
			{content}
			<Primitive.Arrow class="atelier-tooltip__arrow" />
		</Primitive.Content>
	</Primitive.Root>
</Primitive.Provider>
