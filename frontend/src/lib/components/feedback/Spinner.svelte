<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import './feedback.css';

	type SpinnerSize = 'sm' | 'md' | 'lg';

	interface Props extends Omit<HTMLAttributes<HTMLSpanElement>, 'class'> {
		size?: SpinnerSize;
		label?: string;
		class?: string;
	}

	let { size = 'md', label = 'Loading', class: className, ...rest }: Props = $props();
</script>

<span {...rest} class="atelier-spinner {className ?? ''}" data-size={size} role="status" aria-label={label}></span>
