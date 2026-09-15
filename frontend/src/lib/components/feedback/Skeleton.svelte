<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import './feedback.css';

	type SkeletonVariant = 'text' | 'rect' | 'circle';

	interface Props extends Omit<HTMLAttributes<HTMLSpanElement>, 'class' | 'style'> {
		variant?: SkeletonVariant;
		width?: string;
		height?: string;
		class?: string;
	}

	let { variant = 'text', width, height, class: className, ...rest }: Props = $props();
</script>

<span {...rest} class="atelier-skeleton {className ?? ''}" data-variant={variant} aria-hidden="true" style={`--skeleton-width: ${width ?? '100%'}; --skeleton-height: ${height ?? (variant === 'text' ? '12px' : '72px')}`}></span>
