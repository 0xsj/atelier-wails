<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
import './separator.css';

	type Orientation = 'horizontal' | 'vertical';

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
		orientation?: Orientation;
		decorative?: boolean;
		class?: string;
	}

	let {
		orientation = 'horizontal',
		decorative = true,
		class: className,
		...rest
	}: Props = $props();
</script>

<div
	{...rest}
	class="atelier-separator {className ?? ''}"
	data-orientation={orientation}
	role={decorative ? undefined : 'separator'}
	aria-orientation={decorative ? undefined : orientation}
	aria-hidden={decorative || undefined}
></div>
