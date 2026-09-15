<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Collapsible as Primitive } from 'bits-ui';
	import './navigation.css';

	interface Props {
		title: string;
		open?: boolean;
		disabled?: boolean;
		children: Snippet;
		class?: string;
	}

	let { title, open = $bindable(false), disabled = false, children, class: className }: Props = $props();
</script>

<Primitive.Root bind:open {disabled} class="atelier-collapsible {className ?? ''}">
	<Primitive.Trigger class="atelier-collapsible__trigger"><span>{title}</span><span class="atelier-collapsible__chevron" aria-hidden="true">⌄</span></Primitive.Trigger>
	<Primitive.Content class="atelier-collapsible__content">{@render children()}</Primitive.Content>
</Primitive.Root>
