<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './forms.css';

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'class'> { leading?: Snippet; trailing?: Snippet; children: Snippet; class?: string; }

	let { leading, trailing, children, class: className, ...rest }: Props = $props();
</script>

<div {...rest} class="atelier-input-group {className ?? ''}">{#if leading}<span class="atelier-input-group__leading">{@render leading()}</span>{/if}<div class="atelier-input-group__control">{@render children()}</div>{#if trailing}<span class="atelier-input-group__trailing">{@render trailing()}</span>{/if}</div>
