<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { SVGAttributes } from 'svelte/elements';
	import './icon.css';

	type IconSize = 'xs' | 'sm' | 'md' | 'lg';

	interface Props extends Omit<SVGAttributes<SVGSVGElement>, 'children' | 'class' | 'viewBox'> {
		size?: IconSize;
		label?: string;
		children: Snippet;
		class?: string;
	}

	let { size = 'md', label, children, class: className, ...rest }: Props = $props();
</script>

<svg
	{...rest}
	class="atelier-icon {className ?? ''}"
	data-size={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width="1.75"
	stroke-linecap="round"
	stroke-linejoin="round"
	role={label ? 'img' : undefined}
	aria-label={label}
	aria-hidden={label ? undefined : true}
	focusable="false"
>
	{@render children()}
</svg>
