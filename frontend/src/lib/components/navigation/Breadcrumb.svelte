<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import './navigation.css';

	export type BreadcrumbItem = { label: string; href?: string; current?: boolean };

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> {
		items: readonly BreadcrumbItem[];
		ariaLabel?: string;
		separator?: string;
		class?: string;
	}

	let { items, ariaLabel = 'Breadcrumb', separator = '/', class: className, ...rest }: Props = $props();
</script>

<nav {...rest} class="atelier-breadcrumb {className ?? ''}" aria-label={ariaLabel}>
	{#each items as item, index (item.label)}
		{@const current = item.current ?? index === items.length - 1}
		{#if index > 0}<span class="atelier-breadcrumb__separator" aria-hidden="true">{separator}</span>{/if}
		{#if item.href && !current}<a href={item.href}>{item.label}</a>{:else}<span aria-current={current ? 'page' : undefined}>{item.label}</span>{/if}
	{/each}
</nav>
