<script lang="ts">
	import type { Snippet } from 'svelte';
	import { DropdownMenu as Primitive } from 'bits-ui';
	import './navigation.css';

	interface Props { label?: string; shortcut?: string; disabled?: boolean; closeOnSelect?: boolean; onSelect?: (event: Event) => void; children?: Snippet; class?: string; }
	let { label, shortcut, disabled = false, closeOnSelect = true, onSelect, children, class: className }: Props = $props();
</script>
<Primitive.Item {disabled} {closeOnSelect} {onSelect} textValue={label} class="atelier-menu__item {className ?? ''}"><span class="atelier-menu__item-label">{#if children}{@render children()}{:else}{label}{/if}</span>{#if shortcut}<kbd>{shortcut}</kbd>{/if}</Primitive.Item>
