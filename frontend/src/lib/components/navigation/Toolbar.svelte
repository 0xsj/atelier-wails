<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Toolbar as Primitive } from 'bits-ui';
	import './navigation.css';

	type ToolbarOrientation = 'horizontal' | 'vertical';

	interface Props {
		ariaLabel: string;
		orientation?: ToolbarOrientation;
		loop?: boolean;
		children?: Snippet;
		class?: string;
	}

	let { ariaLabel, orientation = 'horizontal', loop = true, children, class: className }: Props = $props();
</script>

<Primitive.Root {orientation} {loop} aria-label={ariaLabel} class="atelier-toolbar {className ?? ''}">
	{@render children?.()}
</Primitive.Root>
