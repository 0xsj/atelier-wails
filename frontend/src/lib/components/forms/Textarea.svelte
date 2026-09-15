<script lang="ts">
	import type { HTMLTextareaAttributes } from 'svelte/elements';
	import './forms.css';

	type ControlSize = 'sm' | 'md' | 'lg';

	interface Props extends Omit<HTMLTextareaAttributes, 'class' | 'value'> {
		size?: ControlSize;
		invalid?: boolean;
		value?: string;
		class?: string;
	}

	let { size = 'md', invalid = false, value = $bindable(''), class: className, ...rest }: Props = $props();
</script>

<textarea
	{...rest}
	bind:value
	class="atelier-textarea {className ?? ''}"
	data-size={size}
	aria-invalid={invalid || undefined}
></textarea>
