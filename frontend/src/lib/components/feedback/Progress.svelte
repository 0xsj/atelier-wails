<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import './feedback.css';

	type ProgressSize = 'sm' | 'md' | 'lg';

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'class' | 'role'> {
		value?: number;
		max?: number;
		size?: ProgressSize;
		class?: string;
	}

	let { value, max = 100, size = 'md', class: className, ...rest }: Props = $props();
	const percentage = $derived(value === undefined ? undefined : Math.min(100, Math.max(0, (value / Math.max(max, 1)) * 100)));
</script>

<div
	{...rest}
	class="atelier-progress {className ?? ''}"
	data-size={size}
	data-indeterminate={value === undefined || undefined}
	role="progressbar"
	aria-valuemin="0"
	aria-valuemax={max}
	aria-valuenow={value}
>
	<span class="atelier-progress__bar" style={`inline-size: ${percentage ?? 100}%`}></span>
</div>
