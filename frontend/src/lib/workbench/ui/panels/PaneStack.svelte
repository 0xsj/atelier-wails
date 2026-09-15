<script lang="ts">
	import type { Snippet } from 'svelte';
	import './panels.css';

	export type PaneStackItem = { id: string; title: string; collapsible?: boolean };

	interface Props {
		panes: readonly PaneStackItem[];
		collapsedIds?: string[];
		weights?: Record<string, number>;
		minPaneSize?: number;
		pane: Snippet<[PaneStackItem]>;
		actions?: Snippet<[PaneStackItem]>;
		onToggle?: (pane: PaneStackItem, collapsed: boolean) => void;
		onWeightsChange?: (weights: Readonly<Record<string, number>>) => void;
		ariaLabel?: string;
		class?: string;
	}

	let { panes, collapsedIds = $bindable([]), weights = $bindable({}), minPaneSize = 48, pane, actions, onToggle, onWeightsChange, ariaLabel = 'Panels', class: className }: Props = $props();

	const baseId = $props.id();
	let sectionRefs: HTMLElement[] = [];
	let drag: { above: number; below: number; start: number; aboveHeight: number; belowHeight: number; aboveWeight: number; belowWeight: number } | null = null;

	function isCollapsed(item: PaneStackItem): boolean {
		return collapsedIds.includes(item.id);
	}

	function weightOf(item: PaneStackItem): number {
		return weights[item.id] ?? 1;
	}

	function previousExpanded(index: number): number | null {
		for (let cursor = index - 1; cursor >= 0; cursor -= 1) if (!isCollapsed(panes[cursor])) return cursor;
		return null;
	}

	function toggle(item: PaneStackItem): void {
		if (item.collapsible === false) return;
		const collapsed = !isCollapsed(item);
		collapsedIds = collapsed ? [...collapsedIds, item.id] : collapsedIds.filter((id) => id !== item.id);
		onToggle?.(item, collapsed);
	}

	function setWeights(above: number, below: number, aboveHeight: number, belowHeight: number, sum: number): void {
		const pair = aboveHeight + belowHeight;
		if (pair <= 0) return;
		weights = { ...weights, [panes[above].id]: (sum * aboveHeight) / pair, [panes[below].id]: (sum * belowHeight) / pair };
		onWeightsChange?.(weights);
	}

	function startResize(event: PointerEvent, above: number, below: number): void {
		if (event.button !== 0) return;
		const aboveRect = sectionRefs[above]?.getBoundingClientRect();
		const belowRect = sectionRefs[below]?.getBoundingClientRect();
		if (!aboveRect || !belowRect) return;
		drag = { above, below, start: event.clientY, aboveHeight: aboveRect.height, belowHeight: belowRect.height, aboveWeight: weightOf(panes[above]), belowWeight: weightOf(panes[below]) };
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}

	function resize(event: PointerEvent): void {
		if (!drag) return;
		const pair = drag.aboveHeight + drag.belowHeight;
		const aboveHeight = Math.min(Math.max(drag.aboveHeight + (event.clientY - drag.start), minPaneSize), pair - minPaneSize);
		setWeights(drag.above, drag.below, aboveHeight, pair - aboveHeight, drag.aboveWeight + drag.belowWeight);
	}

	function endResize(event: PointerEvent): void {
		if (!drag) return;
		drag = null;
		const target = event.currentTarget as HTMLElement;
		if (target.hasPointerCapture(event.pointerId)) target.releasePointerCapture(event.pointerId);
	}

	function handleResizeKey(event: KeyboardEvent, above: number, below: number): void {
		if (event.key !== 'ArrowUp' && event.key !== 'ArrowDown') return;
		event.preventDefault();
		const aboveRect = sectionRefs[above]?.getBoundingClientRect();
		const belowRect = sectionRefs[below]?.getBoundingClientRect();
		if (!aboveRect || !belowRect) return;
		const pair = aboveRect.height + belowRect.height;
		const delta = pair * 0.05 * (event.key === 'ArrowDown' ? 1 : -1);
		const aboveHeight = Math.min(Math.max(aboveRect.height + delta, minPaneSize), pair - minPaneSize);
		setWeights(above, below, aboveHeight, pair - aboveHeight, weightOf(panes[above]) + weightOf(panes[below]));
	}
</script>

<div class="atelier-pane-stack {className ?? ''}" role="group" aria-label={ariaLabel}>
	{#each panes as item, index (item.id)}
		{@const collapsed = isCollapsed(item)}
		{@const above = collapsed ? null : previousExpanded(index)}
		<section bind:this={sectionRefs[index]} class="atelier-pane" data-collapsed={collapsed || undefined} style={`--pane-weight: ${weightOf(item)}`} aria-labelledby={`${baseId}-${item.id}-title`}>
			{#if above !== null}
				<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
				<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
				<div class="atelier-pane__handle" role="separator" tabindex="0" aria-orientation="horizontal" aria-label={`Resize ${panes[above].title} and ${item.title}`} onpointerdown={(event) => startResize(event, above, index)} onpointermove={resize} onpointerup={endResize} onpointercancel={endResize} onkeydown={(event) => handleResizeKey(event, above, index)}></div>
			{/if}
			<header class="atelier-pane__header">
				<button type="button" class="atelier-pane__toggle" aria-expanded={!collapsed} aria-controls={`${baseId}-${item.id}-body`} disabled={item.collapsible === false} onclick={() => toggle(item)}>
					<span class="atelier-pane__chevron" aria-hidden="true">›</span>
					<span id={`${baseId}-${item.id}-title`} class="atelier-pane__title">{item.title}</span>
				</button>
				{#if actions && !collapsed}<div class="atelier-pane__actions">{@render actions(item)}</div>{/if}
			</header>
			{#if !collapsed}<div id={`${baseId}-${item.id}-body`} class="atelier-pane__body">{@render pane(item)}</div>{/if}
		</section>
	{/each}
</div>
