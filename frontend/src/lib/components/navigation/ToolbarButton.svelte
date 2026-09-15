<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Toolbar as Primitive } from 'bits-ui';
	import './navigation.css';

	interface Props {
		label?: string;
		disabled?: boolean;
		onclick?: (event: MouseEvent) => void;
		children?: Snippet;
		class?: string;
	}

	let { label, disabled = false, onclick, children, class: className }: Props = $props();
</script>

<Primitive.Button {disabled} {onclick} aria-label={label} title={label} class="atelier-toolbar__button {className ?? ''}">
	{@render children?.()}
</Primitive.Button>
