<script lang="ts">
	import { Combobox as Primitive } from 'bits-ui';
	import './forms.css';

	export type ComboboxOption = { value: string; label: string; disabled?: boolean };

	interface Props {
		options: readonly ComboboxOption[];
		value?: string;
		placeholder?: string;
		size?: 'sm' | 'md' | 'lg';
		disabled?: boolean;
		ariaLabel?: string;
		class?: string;
	}

	let { options, value = $bindable(''), placeholder = 'Search options', size = 'md', disabled = false, ariaLabel = 'Choose an option', class: className }: Props = $props();
	const items = $derived(options.map((option) => ({ ...option })));
</script>

<div class="atelier-combobox {className ?? ''}">
<Primitive.Root type="single" bind:value {items} {disabled}>
	<Primitive.Input class="atelier-combobox__input" data-size={size} placeholder={placeholder} aria-label={ariaLabel} />
	<Primitive.Content class="atelier-combobox__content">
		<Primitive.Viewport class="atelier-combobox__viewport">
			{#each options as option (option.value)}
				<Primitive.Item value={option.value} label={option.label} disabled={option.disabled} class="atelier-combobox__item"><span>{option.label}</span><span class="atelier-combobox__check" aria-hidden="true">✓</span></Primitive.Item>
			{/each}
		</Primitive.Viewport>
	</Primitive.Content>
</Primitive.Root>
</div>
