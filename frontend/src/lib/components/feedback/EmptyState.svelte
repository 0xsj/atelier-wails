<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './feedback.css';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> {
		title: string;
		description?: string;
		icon?: Snippet;
		action?: Snippet;
		children?: Snippet;
		class?: string;
	}

	let { title, description, icon, action, children, class: className, ...rest }: Props = $props();
</script>

<section {...rest} class="atelier-empty-state {className ?? ''}" aria-label={title}>
	<div class="atelier-empty-state__icon" aria-hidden="true">{#if icon}{@render icon()}{:else}∅{/if}</div>
	<div class="atelier-empty-state__copy"><h3>{title}</h3>{#if description}<p>{description}</p>{/if}{#if children}<div>{@render children()}</div>{/if}</div>
	{#if action}<div class="atelier-empty-state__action">{@render action()}</div>{/if}
</section>
