<script lang="ts">
	import type { Snippet } from 'svelte';
	import './layout.css';

	type SplitDirection = 'horizontal' | 'vertical';

	interface Props {
		sizes?: number[];
		direction?: SplitDirection;
		minSize?: number;
		collapsible?: boolean;
		step?: number;
		pane: Snippet<[number]>;
		ariaLabel?: string;
		onChange?: (sizes: number[]) => void;
		class?: string;
	}

	let { sizes = $bindable([50, 50]), direction = 'horizontal', minSize = 10, collapsible = false, step = 2, pane, ariaLabel = 'Split panes', onChange, class: className }: Props = $props();

	let paneRefs: HTMLDivElement[] = [];
	let drag: { index: number; start: number; before: number; pairPx: number } | null = null;
	const remembered = new Map<number, number>();

	const total = $derived(sizes.reduce((sum, size) => sum + Math.max(size, 0), 0) || 1);
	const template = $derived(sizes.map((size, index) => `minmax(0, ${Math.max(size, 0)}fr)${index < sizes.length - 1 ? ' var(--split-handle-size)' : ''}`).join(' '));

	function commit(next: number[]): void {
		sizes = next;
		onChange?.(next);
	}

	function resolve(index: number, desiredBefore: number): void {
		const pair = sizes[index] + sizes[index + 1];
		const min = Math.min((minSize / 100) * total, pair / 2);
		let before = desiredBefore;
		if (collapsible && before < min / 2) before = 0;
		else if (collapsible && before > pair - min / 2) before = pair;
		else before = Math.min(Math.max(before, min), pair - min);
		const after = pair - before;
		if (before > 0) remembered.set(index, before);
		if (after > 0) remembered.set(index + 1, after);
		const next = [...sizes];
		next[index] = before;
		next[index + 1] = after;
		commit(next);
	}

	function toggleCollapse(index: number): void {
		if (!collapsible) return;
		const pair = sizes[index] + sizes[index + 1];
		if (sizes[index] === 0) {
			resolve(index, Math.min(remembered.get(index) ?? pair / 2, pair));
		} else if (sizes[index + 1] === 0) {
			resolve(index, pair - Math.min(remembered.get(index + 1) ?? pair / 2, pair));
		} else {
			remembered.set(index, sizes[index]);
			resolve(index, 0);
		}
	}

	function startResize(event: PointerEvent, index: number): void {
		if (event.button !== 0) return;
		const first = paneRefs[index]?.getBoundingClientRect();
		const second = paneRefs[index + 1]?.getBoundingClientRect();
		if (!first || !second) return;
		const pairPx = direction === 'horizontal' ? first.width + second.width : first.height + second.height;
		drag = { index, start: direction === 'horizontal' ? event.clientX : event.clientY, before: sizes[index], pairPx };
		(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
	}

	function resize(event: PointerEvent): void {
		if (!drag || drag.pairPx <= 0) return;
		const position = direction === 'horizontal' ? event.clientX : event.clientY;
		const pair = sizes[drag.index] + sizes[drag.index + 1];
		resolve(drag.index, drag.before + ((position - drag.start) / drag.pairPx) * pair);
	}

	function endResize(event: PointerEvent): void {
		if (!drag) return;
		drag = null;
		const target = event.currentTarget as HTMLElement;
		if (target.hasPointerCapture(event.pointerId)) target.releasePointerCapture(event.pointerId);
	}

	function handleKeydown(event: KeyboardEvent, index: number): void {
		const decrease = direction === 'horizontal' ? event.key === 'ArrowLeft' : event.key === 'ArrowUp';
		const increase = direction === 'horizontal' ? event.key === 'ArrowRight' : event.key === 'ArrowDown';
		if (decrease || increase) {
			event.preventDefault();
			resolve(index, sizes[index] + ((increase ? step : -step) / 100) * total);
		} else if (event.key === 'Enter' && collapsible) {
			event.preventDefault();
			toggleCollapse(index);
		}
	}
</script>

<div class="atelier-split-group {className ?? ''}" data-direction={direction} style={`--split-template: ${template}`} role="group" aria-label={ariaLabel}>
	{#each sizes as size, index (index)}
		<div bind:this={paneRefs[index]} class="atelier-split-group__pane" data-collapsed={size === 0 || undefined}>{@render pane(index)}</div>
		{#if index < sizes.length - 1}
			<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<div class="atelier-split-group__handle" role="separator" tabindex="0" aria-orientation={direction === 'horizontal' ? 'vertical' : 'horizontal'} aria-label={`Resize pane ${index + 1}`} aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round((size / total) * 100)} data-collapsed={size === 0 || sizes[index + 1] === 0 || undefined} onpointerdown={(event) => startResize(event, index)} onpointermove={resize} onpointerup={endResize} onpointercancel={endResize} ondblclick={() => toggleCollapse(index)} onkeydown={(event) => handleKeydown(event, index)}><span></span></div>
		{/if}
	{/each}
</div>
