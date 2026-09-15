<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './content.css';

	interface Props extends Omit<HTMLAttributes<HTMLElement>, 'children' | 'class'> {
		children?: Snippet;
		class?: string;
	}

	let { children, class: className, ...rest }: Props = $props();
</script>

<code {...rest} class="atelier-inline-code {className ?? ''}">{@render children?.()}</code>
