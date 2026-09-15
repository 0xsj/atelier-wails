<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './feedback.css';

	type Tone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger';
	type BadgeSize = 'sm' | 'md';
	type BadgeVariant = 'soft' | 'solid' | 'outline';

	interface Props extends Omit<HTMLAttributes<HTMLSpanElement>, 'children' | 'class'> {
		tone?: Tone;
		size?: BadgeSize;
		variant?: BadgeVariant;
		children?: Snippet;
		class?: string;
	}

	let { tone = 'neutral', size = 'sm', variant = 'soft', children, class: className, ...rest }: Props = $props();
</script>

<span {...rest} class="atelier-badge {className ?? ''}" data-tone={tone} data-size={size} data-variant={variant}>
	{@render children?.()}
</span>
