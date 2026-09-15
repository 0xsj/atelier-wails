<script lang="ts">
	import { Select as Primitive } from 'bits-ui';
import './forms.css';

	export type SelectOption = {
		value: string;
		label: string;
		disabled?: boolean;
	};

	type ControlSize = 'sm' | 'md' | 'lg';

	interface Props {
		options: readonly SelectOption[];
		value?: string;
		placeholder?: string;
		size?: ControlSize;
		disabled?: boolean;
		required?: boolean;
		name?: string;
		id?: string;
		'aria-label'?: string;
		'aria-labelledby'?: string;
		class?: string;
	}

	let {
		options,
		value = $bindable(''),
		placeholder = 'Select an option',
		size = 'md',
		disabled = false,
		required = false,
		name,
		id,
		'aria-label': ariaLabel,
		'aria-labelledby': ariaLabelledby,
		class: className
	}: Props = $props();
	const items = $derived(options.map((option) => ({ ...option })));
</script>

<Primitive.Root
	type="single"
	bind:value
	items={items}
	{disabled}
	{required}
	{name}
>
	<Primitive.Trigger
		{id}
		class="atelier-select__trigger {className ?? ''}"
		data-size={size}
		aria-label={ariaLabel}
		aria-labelledby={ariaLabelledby}
	>
		<Primitive.Value {placeholder} />
		<span class="atelier-select__chevron" aria-hidden="true">⌄</span>
	</Primitive.Trigger>
	<Primitive.Content class="atelier-select__content">
		<Primitive.Viewport class="atelier-select__viewport">
			{#each options as option (option.value)}
				<Primitive.Item
					value={option.value}
					label={option.label}
					disabled={option.disabled}
					class="atelier-select__item"
				>
					<span>{option.label}</span>
					<span class="atelier-select__check" aria-hidden="true">✓</span>
				</Primitive.Item>
			{/each}
		</Primitive.Viewport>
	</Primitive.Content>
</Primitive.Root>
