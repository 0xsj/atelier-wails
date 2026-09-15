<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './feedback.css';

	type Tone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger';

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'class' | 'role'> {
		tone?: Tone;
		title?: string;
		children?: Snippet;
		actions?: Snippet;
		dismissible?: boolean;
		onDismiss?: () => void;
		class?: string;
	}

	let { tone = 'neutral', title, children, actions, dismissible = false, onDismiss, class: className, ...rest }: Props = $props();
</script>

<div {...rest} class="atelier-banner {className ?? ''}" data-tone={tone} role="status">
	<div class="atelier-banner__content"><span class="atelier-banner__marker" aria-hidden="true"></span><div class="atelier-banner__copy">{#if title}<strong>{title}</strong>{/if}{#if children}<span>{@render children()}</span>{/if}</div></div>
	{#if actions}<div class="atelier-banner__actions">{@render actions()}</div>{/if}
	{#if dismissible}<button type="button" class="atelier-banner__close" aria-label="Dismiss banner" onclick={onDismiss}>×</button>{/if}
</div>
