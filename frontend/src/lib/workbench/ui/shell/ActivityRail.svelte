<script lang="ts">
	import type { Snippet } from 'svelte';
	import ActivityItem from './ActivityItem.svelte';
	import './shell.css';

	export type ActivityEntry = { id: string; label: string; badge?: number | string; disabled?: boolean };

	interface Props {
		items: readonly ActivityEntry[];
		activeId?: string;
		icon: Snippet<[ActivityEntry]>;
		end?: Snippet;
		ariaLabel?: string;
		onSelect?: (item: ActivityEntry) => void;
		class?: string;
	}

	let { items, activeId = $bindable(''), icon, end, ariaLabel = 'Activity', onSelect, class: className }: Props = $props();

	let list = $state<HTMLDivElement | null>(null);
	const focusIndex = $derived(Math.max(items.findIndex((item) => item.id === activeId), 0));

	function select(item: ActivityEntry): void {
		if (item.disabled) return;
		activeId = item.id;
		onSelect?.(item);
	}

	function focusItem(index: number): void {
		const target = items[index];
		if (!target) return;
		select(target);
		list?.querySelectorAll<HTMLElement>('[role="tab"]')[index]?.focus();
	}

	function handleKeydown(event: KeyboardEvent): void {
		if (!items.length) return;
		const current = items.findIndex((item) => item.id === activeId);
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			focusItem((current + 1) % items.length);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			focusItem((current - 1 + items.length) % items.length);
		} else if (event.key === 'Home') {
			event.preventDefault();
			focusItem(0);
		} else if (event.key === 'End') {
			event.preventDefault();
			focusItem(items.length - 1);
		}
	}
</script>

<nav class="atelier-activity-rail {className ?? ''}" aria-label={ariaLabel}>
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div bind:this={list} class="atelier-activity-rail__items" role="tablist" aria-orientation="vertical" onkeydown={handleKeydown}>
		{#each items as item, index (item.id)}
			<ActivityItem label={item.label} badge={item.badge} disabled={item.disabled} active={item.id === activeId} focusable={index === focusIndex} onclick={() => select(item)}>{@render icon(item)}</ActivityItem>
		{/each}
	</div>
	{#if end}<div class="atelier-activity-rail__end">{@render end()}</div>{/if}
</nav>
