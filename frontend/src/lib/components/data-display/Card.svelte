<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './data-display.css';

	type CardTone = 'panel' | 'raised' | 'sunk';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> {
		title?: string;
		description?: string;
		tone?: CardTone;
		header?: Snippet;
		actions?: Snippet;
		children?: Snippet;
		footer?: Snippet;
		class?: string;
	}

	let { title, description, tone = 'panel', header, actions, children, footer, class: className, ...rest }: Props = $props();
</script>

<article {...rest} class="atelier-card {className ?? ''}" data-tone={tone}>
	{#if title || description || header || actions}
		<header class="atelier-card__header">
			<div class="atelier-card__heading">
				{#if title}<h3>{title}</h3>{/if}
				{#if description}<p>{description}</p>{/if}
			</div>
			{#if header}<div class="atelier-card__header-content">{@render header()}</div>{/if}
			{#if actions}<div class="atelier-card__actions">{@render actions()}</div>{/if}
		</header>
	{/if}
	{#if children}<div class="atelier-card__body">{@render children()}</div>{/if}
	{#if footer}<footer class="atelier-card__footer">{@render footer()}</footer>{/if}
</article>
