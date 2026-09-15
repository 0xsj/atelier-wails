<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import './layout.css';

	type ResizeDirection = 'horizontal' | 'vertical';

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'class'> {
		first: Snippet;
		second: Snippet;
		value?: number;
		direction?: ResizeDirection;
		min?: number;
		max?: number;
		step?: number;
		class?: string;
	}

	let { first, second, value = $bindable(42), direction = 'horizontal', min = 20, max = 80, step = 2, class: className, ...rest }: Props = $props();
	let root: HTMLDivElement;
	let dragging = false;

	function setValue(next: number): void {
		value = Math.min(max, Math.max(min, next));
	}

	function startResize(event: PointerEvent): void {
		if (event.button !== 0) return;
		dragging = true;
		root.setPointerCapture(event.pointerId);
	}

	function resize(event: PointerEvent): void {
		if (!dragging) return;
		const bounds = root.getBoundingClientRect();
		const position = direction === 'horizontal' ? event.clientX - bounds.left : event.clientY - bounds.top;
		const total = direction === 'horizontal' ? bounds.width : bounds.height;
		if (total > 0) setValue((position / total) * 100);
	}

	function endResize(event: PointerEvent): void {
		if (!dragging) return;
		dragging = false;
		if (root.hasPointerCapture(event.pointerId)) root.releasePointerCapture(event.pointerId);
	}

	function handleKeydown(event: KeyboardEvent): void {
		const decrease = direction === 'horizontal' ? event.key === 'ArrowLeft' : event.key === 'ArrowUp';
		const increase = direction === 'horizontal' ? event.key === 'ArrowRight' : event.key === 'ArrowDown';
		if (!decrease && !increase) return;
		event.preventDefault();
		setValue(value + (increase ? step : -step));
	}
</script>

<div {...rest} bind:this={root} class="atelier-resizable {className ?? ''}" data-direction={direction} style={`--split-size: ${value}%`} onpointermove={resize} onpointerup={endResize} onpointercancel={endResize}>
	<div class="atelier-resizable__pane">{@render first()}</div>
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div class="atelier-resizable__handle" role="separator" tabindex="0" aria-label="Resize panels" aria-orientation={direction} aria-valuemin={min} aria-valuemax={max} aria-valuenow={Math.round(value)} onpointerdown={startResize} onkeydown={handleKeydown}><span></span></div>
	<div class="atelier-resizable__pane">{@render second()}</div>
</div>
