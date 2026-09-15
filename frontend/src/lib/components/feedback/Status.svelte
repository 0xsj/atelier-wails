<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './feedback.css';

	type Tone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger';
	type StatusSize = 'sm' | 'md';

	interface Props extends Omit<HTMLAttributes<HTMLSpanElement>, 'children' | 'class'> {
		tone?: Tone;
		size?: StatusSize;
		pulse?: boolean;
		children?: Snippet;
		class?: string;
	}

	let { tone = 'neutral', size = 'sm', pulse = false, children, class: className, ...rest }: Props = $props();
</script>

<span {...rest} class="atelier-status {className ?? ''}" data-tone={tone} data-size={size} data-pulse={pulse || undefined}>
	<span class="atelier-status__dot" aria-hidden="true"></span>
	<span class="atelier-status__label">{@render children?.()}</span>
</span>
