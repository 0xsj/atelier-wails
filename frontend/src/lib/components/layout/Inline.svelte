<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { LayoutAlign, LayoutGap, LayoutJustify, LayoutTag } from './layout';
	import './layout.css';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> {
		as?: LayoutTag;
		gap?: LayoutGap;
		align?: LayoutAlign;
		justify?: LayoutJustify;
		wrap?: boolean;
		children?: Snippet;
		class?: string;
	}

	let { as: tag = 'div', gap = 'sm', align = 'center', justify = 'start', wrap = false, children, class: className, ...rest }: Props = $props();
</script>

<svelte:element this={tag} {...rest} class="atelier-inline {className ?? ''}" data-gap={gap} data-align={align} data-justify={justify} data-wrap={wrap || undefined}>
	{@render children?.()}
</svelte:element>
