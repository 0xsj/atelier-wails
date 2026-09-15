<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Checkbox as Primitive } from 'bits-ui';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import './forms.css';

	interface Props extends Omit<HTMLButtonAttributes, 'children' | 'id'> {
		id?: string;
		checked?: boolean;
		indeterminate?: boolean;
		disabled?: boolean;
		required?: boolean;
		readonly?: boolean;
		name?: string;
		value?: string;
		form?: string;
		children?: Snippet;
		class?: string;
	}

	let {
		checked = $bindable(false),
		indeterminate = false,
		children,
		class: className,
		...rest
	}: Props = $props();
</script>

<Primitive.Root
	{...rest}
	bind:checked
	{indeterminate}
	class="atelier-checkbox {className ?? ''}"
	aria-label={rest['aria-label']}
>
	<span class="atelier-checkbox__indicator" aria-hidden="true"></span>
	{@render children?.()}
</Primitive.Root>
