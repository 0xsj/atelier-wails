<script lang="ts">
	import { Switch as Primitive } from 'bits-ui';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import './forms.css';

	interface Props extends Omit<HTMLButtonAttributes, 'children' | 'id'> {
		id?: string;
		checked?: boolean;
		disabled?: boolean;
		required?: boolean;
		name?: string;
		value?: string;
		class?: string;
	}

	let { checked = $bindable(false), class: className, ...rest }: Props = $props();
</script>

<Primitive.Root {...rest} bind:checked class="atelier-switch {className ?? ''}">
	<Primitive.Thumb class="atelier-switch__thumb" />
</Primitive.Root>
