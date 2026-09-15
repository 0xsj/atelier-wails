<script lang="ts">
	import { RadioGroup as Primitive } from 'bits-ui';
	import './forms.css';

	export type RadioOption = { value: string; label: string; disabled?: boolean };
	interface Props { items: readonly RadioOption[]; value?: string; name?: string; disabled?: boolean; required?: boolean; ariaLabel?: string; class?: string; }
	let { items, value = $bindable(''), name, disabled = false, required = false, ariaLabel, class: className }: Props = $props();
</script>
<Primitive.Root bind:value {name} {disabled} {required} aria-label={ariaLabel} class="atelier-radio-group {className ?? ''}">{#each items as item (item.value)}<Primitive.Item value={item.value} disabled={item.disabled} class="atelier-radio-group__item"><span class="atelier-radio-group__indicator" aria-hidden="true"></span><span>{item.label}</span></Primitive.Item>{/each}</Primitive.Root>
