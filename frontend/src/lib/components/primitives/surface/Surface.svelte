<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './surface.css';

	type SurfaceTone = 'ground' | 'sunk' | 'panel' | 'raised' | 'selected';
	type SurfacePadding = 'none' | 'sm' | 'md' | 'lg';
	type SurfaceBorder = 'none' | 'default' | 'strong';
	type SurfaceRadius = 'none' | 'sm' | 'md' | 'lg';

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
		tone?: SurfaceTone;
		padding?: SurfacePadding;
		border?: SurfaceBorder;
		radius?: SurfaceRadius;
		children?: Snippet;
		class?: string;
	}

	let {
		tone = 'panel',
		padding = 'md',
		border = 'default',
		radius = 'md',
		children,
		class: className,
		...rest
	}: Props = $props();
</script>

<div
	{...rest}
	class="atelier-surface {className ?? ''}"
	data-tone={tone}
	data-padding={padding}
	data-border={border}
	data-radius={radius}
>
	{@render children?.()}
</div>
