<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { LayoutAlign, LayoutGap, LayoutTag } from './layout';
	import './layout.css';

	type GridColumns = 1 | 2 | 3 | 4 | 5 | 6 | 'auto';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> {
		as?: LayoutTag;
		columns?: GridColumns;
		minColumn?: string;
		gap?: LayoutGap;
		align?: LayoutAlign;
		children?: Snippet;
		class?: string;
	}

	let { as: tag = 'div', columns = 2, minColumn = '220px', gap = 'md', align = 'stretch', children, class: className, ...rest }: Props = $props();
</script>

<svelte:element this={tag} {...rest} class="atelier-grid {className ?? ''}" data-columns={columns} data-gap={gap} data-align={align} style={columns === 'auto' ? `--grid-min-column: ${minColumn}` : undefined}>
	{@render children?.()}
</svelte:element>
