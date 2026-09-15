<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { LayoutAlign, LayoutGap, LayoutTag } from './layout';
	import './layout.css';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> {
		as?: LayoutTag;
		gap?: LayoutGap;
		align?: LayoutAlign;
		children?: Snippet;
		class?: string;
	}

	let { as: tag = 'div', gap = 'md', align = 'stretch', children, class: className, ...rest }: Props = $props();
</script>

<svelte:element this={tag} {...rest} class="atelier-stack {className ?? ''}" data-gap={gap} data-align={align}>
	{@render children?.()}
</svelte:element>
