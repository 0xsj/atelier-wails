<script lang="ts">
	import type { ButtonProps } from './button';
import './button.css';

	let {
		variant = 'secondary',
		size = 'md',
		loading = false,
		disabled = false,
		type = 'button',
		children,
		class: className,
		...rest
	}: ButtonProps = $props();
</script>

<button
	{...rest}
	{type}
	class="atelier-button {className ?? ''}"
	data-variant={variant}
	data-size={size}
	data-loading={loading || undefined}
	disabled={disabled || loading}
	aria-busy={loading || undefined}
>
	{#if loading}<span class="atelier-button__spinner" aria-hidden="true"></span>{/if}
	<span class="atelier-button__content">{@render children?.()}</span>
</button>
