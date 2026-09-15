<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Tabs as Primitive } from 'bits-ui';
	import type { HTMLAttributes } from 'svelte/elements';
	import './navigation.css';

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'class' | 'id'> {
		id?: string;
		value: string;
		children?: Snippet;
		class?: string;
	}

	let { value, children, class: className, ...rest }: Props = $props();
</script>

<Primitive.Content {...rest} {value} class="atelier-tabs__panel {className ?? ''}">
	{@render children?.()}
</Primitive.Content>
