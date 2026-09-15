<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { EditorTabItem } from './EditorTabs.svelte';
	import './view-host.css';

	interface Props {
		tab: EditorTabItem;
		active: boolean;
		focusable: boolean;
		draggable: boolean;
		dropBefore: boolean;
		icon?: Snippet<[EditorTabItem]>;
		onActivate: () => void;
		onClose?: () => void;
		onDragStart?: (event: DragEvent) => void;
		onDragOver?: (event: DragEvent) => void;
		onDrop?: (event: DragEvent) => void;
		onDragEnd?: () => void;
	}

	let { tab, active, focusable, draggable, dropBefore, icon, onActivate, onClose, onDragStart, onDragOver, onDrop, onDragEnd }: Props = $props();
	const closable = $derived(tab.closable !== false && !tab.pinned && Boolean(onClose));

	function handleAuxClick(event: MouseEvent): void {
		if (event.button === 1 && tab.closable !== false && onClose) {
			event.preventDefault();
			onClose();
		}
	}
</script>

<!-- Keyboard handling lives on the tablist; see EditorTabs. -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
	role="tab"
	class="atelier-editor-tab"
	id={`editor-tab-${tab.id}`}
	aria-selected={active}
	tabindex={focusable ? 0 : -1}
	data-active={active || undefined}
	data-dirty={tab.dirty || undefined}
	data-pinned={tab.pinned || undefined}
	data-preview={tab.preview || undefined}
	data-drop-before={dropBefore || undefined}
	title={tab.description ?? tab.label}
	{draggable}
	onclick={onActivate}
	onauxclick={handleAuxClick}
	ondragstart={onDragStart}
	ondragover={onDragOver}
	ondrop={onDrop}
	ondragend={onDragEnd}
>
	{#if icon}<span class="atelier-editor-tab__icon" aria-hidden="true">{@render icon(tab)}</span>{/if}
	<span class="atelier-editor-tab__label">{tab.label}</span>
	{#if tab.dirty}<span class="atelier-editor-tab__sr">, unsaved changes</span>{/if}
	{#if closable}
		<button type="button" class="atelier-editor-tab__close" tabindex="-1" aria-label={`Close ${tab.label}`} onclick={(event) => { event.stopPropagation(); onClose?.(); }}>
			<span class="atelier-editor-tab__dirty" aria-hidden="true"></span>
			<span class="atelier-editor-tab__close-glyph" aria-hidden="true">×</span>
		</button>
	{:else if tab.dirty}
		<span class="atelier-editor-tab__dirty" aria-hidden="true"></span>
	{/if}
</div>
