<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './content.css';

	type TagTone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger';
	type TagSize = 'sm' | 'md';
	type TagVariant = 'soft' | 'solid' | 'outline';

	interface Props extends Omit<HTMLAttributes<HTMLSpanElement>, 'children' | 'class'> {
		tone?: TagTone;
		size?: TagSize;
		variant?: TagVariant;
		dot?: boolean;
		dismissible?: boolean;
		removeLabel?: string;
		children?: Snippet;
		onRemove?: () => void;
		class?: string;
	}

	let { tone = 'neutral', size = 'sm', variant = 'soft', dot = false, dismissible = false, removeLabel = 'Remove tag', children, onRemove, class: className, ...rest }: Props = $props();
</script>

<span {...rest} class="atelier-tag {className ?? ''}" data-tone={tone} data-size={size} data-variant={variant}>
	{#if dot}<span class="atelier-tag__dot" aria-hidden="true"></span>{/if}
	<span class="atelier-tag__label">{@render children?.()}</span>
	{#if dismissible}<button type="button" class="atelier-tag__remove" aria-label={removeLabel} onclick={onRemove}>×</button>{/if}
</span>
