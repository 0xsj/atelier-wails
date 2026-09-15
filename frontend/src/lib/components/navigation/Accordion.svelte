<script lang="ts">
	import { Accordion as Primitive } from 'bits-ui';
	import './navigation.css';

	export type AccordionItem = { value: string; title: string; content: string; disabled?: boolean };

	interface Props {
		items: readonly AccordionItem[];
		type?: 'single' | 'multiple';
		value?: string | string[];
		ariaLabel?: string;
		class?: string;
	}

	let { items, type = 'single', value = $bindable(''), ariaLabel = 'Accordion', class: className }: Props = $props();
</script>

{#if type === 'multiple'}
	<Primitive.Root type="multiple" value={Array.isArray(value) ? value : []} onValueChange={(next) => (value = next)} aria-label={ariaLabel} class="atelier-accordion {className ?? ''}">
		{#each items as item (item.value)}
			<Primitive.Item value={item.value} disabled={item.disabled} class="atelier-accordion__item"><Primitive.Header level={3}><Primitive.Trigger class="atelier-accordion__trigger"><span>{item.title}</span><span class="atelier-accordion__chevron" aria-hidden="true">⌄</span></Primitive.Trigger></Primitive.Header><Primitive.Content class="atelier-accordion__content">{item.content}</Primitive.Content></Primitive.Item>
		{/each}
	</Primitive.Root>
{:else}
	<Primitive.Root type="single" value={typeof value === 'string' ? value : ''} onValueChange={(next) => (value = next)} aria-label={ariaLabel} class="atelier-accordion {className ?? ''}">
		{#each items as item (item.value)}
			<Primitive.Item value={item.value} disabled={item.disabled} class="atelier-accordion__item"><Primitive.Header level={3}><Primitive.Trigger class="atelier-accordion__trigger"><span>{item.title}</span><span class="atelier-accordion__chevron" aria-hidden="true">⌄</span></Primitive.Trigger></Primitive.Header><Primitive.Content class="atelier-accordion__content">{item.content}</Primitive.Content></Primitive.Item>
		{/each}
	</Primitive.Root>
{/if}
