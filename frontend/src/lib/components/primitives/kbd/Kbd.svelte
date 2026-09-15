<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './kbd.css';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> { children?: Snippet; class?: string; }
	let { children, class: className, ...rest }: Props = $props();
</script>
<kbd {...rest} class="atelier-kbd {className ?? ''}">{@render children?.()}</kbd>
