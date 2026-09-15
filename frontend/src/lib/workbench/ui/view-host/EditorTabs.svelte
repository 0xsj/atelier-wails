<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ContextMenu, MenuItem, MenuSeparator } from '../../../components/navigation';
	import EditorTab from './EditorTab.svelte';
	import './view-host.css';

	export type EditorTabItem = {
		id: string;
		label: string;
		description?: string;
		dirty?: boolean;
		pinned?: boolean;
		preview?: boolean;
		closable?: boolean;
	};

	interface Props {
		tabs: readonly EditorTabItem[];
		activeId?: string;
		ariaLabel?: string;
		icon?: Snippet<[EditorTabItem]>;
		trailing?: Snippet;
		onActivate?: (tab: EditorTabItem) => void;
		onClose?: (tab: EditorTabItem) => void;
		onCloseOthers?: (tab: EditorTabItem) => void;
		onCloseRight?: (tab: EditorTabItem) => void;
		onCloseAll?: () => void;
		onPin?: (tab: EditorTabItem, pinned: boolean) => void;
		onReorder?: (fromIndex: number, toIndex: number) => void;
		class?: string;
	}

	let { tabs, activeId = $bindable(''), ariaLabel = 'Open editors', icon, trailing, onActivate, onClose, onCloseOthers, onCloseRight, onCloseAll, onPin, onReorder, class: className }: Props = $props();

	let list = $state<HTMLDivElement | null>(null);
	let dragIndex = $state<number | null>(null);
	let dropIndex = $state<number | null>(null);
	const focusIndex = $derived(Math.max(tabs.findIndex((tab) => tab.id === activeId), 0));

	$effect(() => {
		if (!list || !activeId) return;
		list.querySelector<HTMLElement>('[data-active]')?.scrollIntoView({ block: 'nearest', inline: 'nearest' });
	});

	function activate(tab: EditorTabItem): void {
		activeId = tab.id;
		onActivate?.(tab);
	}

	function focusTab(index: number): void {
		const target = tabs[index];
		if (!target) return;
		activate(target);
		list?.querySelectorAll<HTMLElement>('[role="tab"]')[index]?.focus();
	}

	function handleKeydown(event: KeyboardEvent): void {
		if (!tabs.length) return;
		const current = tabs.findIndex((tab) => tab.id === activeId);
		if (event.key === 'ArrowRight') {
			event.preventDefault();
			focusTab((current + 1) % tabs.length);
		} else if (event.key === 'ArrowLeft') {
			event.preventDefault();
			focusTab((current - 1 + tabs.length) % tabs.length);
		} else if (event.key === 'Home') {
			event.preventDefault();
			focusTab(0);
		} else if (event.key === 'End') {
			event.preventDefault();
			focusTab(tabs.length - 1);
		} else if (event.key === 'Delete' && current >= 0 && tabs[current].closable !== false) {
			event.preventDefault();
			onClose?.(tabs[current]);
		}
	}

	function handleDragStart(event: DragEvent, index: number): void {
		dragIndex = index;
		event.dataTransfer?.setData('text/plain', tabs[index].id);
		if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
	}

	function handleDragOver(event: DragEvent, index: number): void {
		if (dragIndex === null) return;
		event.preventDefault();
		dropIndex = index;
	}

	function handleDrop(event: DragEvent, index: number): void {
		event.preventDefault();
		if (dragIndex !== null && dragIndex !== index) onReorder?.(dragIndex, index);
		dragIndex = null;
		dropIndex = null;
	}

	function handleDragEnd(): void {
		dragIndex = null;
		dropIndex = null;
	}
</script>

<div class="atelier-editor-tabs {className ?? ''}">
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div bind:this={list} class="atelier-editor-tabs__list" role="tablist" aria-label={ariaLabel} aria-orientation="horizontal" onkeydown={handleKeydown}>
		{#each tabs as tab, index (tab.id)}
			<ContextMenu>
				{#snippet trigger()}
					<EditorTab
						{tab}
						{icon}
						active={tab.id === activeId}
						focusable={index === focusIndex}
						draggable={Boolean(onReorder)}
						dropBefore={dropIndex === index && dragIndex !== null && dragIndex !== index}
						onActivate={() => activate(tab)}
						onClose={onClose ? () => onClose(tab) : undefined}
						onDragStart={(event) => handleDragStart(event, index)}
						onDragOver={(event) => handleDragOver(event, index)}
						onDrop={(event) => handleDrop(event, index)}
						onDragEnd={handleDragEnd}
					/>
				{/snippet}
				<MenuItem label="Close" disabled={tab.closable === false || !onClose} onSelect={() => onClose?.(tab)} />
				<MenuItem label="Close others" disabled={!onCloseOthers || tabs.length < 2} onSelect={() => onCloseOthers?.(tab)} />
				<MenuItem label="Close to the right" disabled={!onCloseRight || index === tabs.length - 1} onSelect={() => onCloseRight?.(tab)} />
				<MenuItem label="Close all" disabled={!onCloseAll} onSelect={() => onCloseAll?.()} />
				<MenuSeparator />
				<MenuItem label={tab.pinned ? 'Unpin' : 'Pin'} disabled={!onPin} onSelect={() => onPin?.(tab, !tab.pinned)} />
			</ContextMenu>
		{/each}
	</div>
	{#if trailing}<div class="atelier-editor-tabs__trailing">{@render trailing()}</div>{/if}
</div>
