<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import './forms.css';

	type ControlSize = 'sm' | 'md' | 'lg';

	interface Props extends Omit<HTMLInputAttributes, 'class' | 'size' | 'value'> {
		size?: ControlSize;
		invalid?: boolean;
		value?: string;
		class?: string;
	}

	let { size = 'md', invalid = false, value = $bindable(''), class: className, ...rest }: Props = $props();
</script>

<input
	{...rest}
	bind:value
	class="atelier-input {className ?? ''}"
	data-size={size}
	aria-invalid={invalid || undefined}
 />
