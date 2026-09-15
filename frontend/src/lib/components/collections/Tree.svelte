<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import './collections.css';

	export type TreeNode = { id: string; label: string; children?: readonly TreeNode[]; disabled?: boolean };
	type VisibleNode = TreeNode & { depth: number; parentId?: string; hasChildren: boolean; expanded: boolean };
	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'class' | 'role'> { items: readonly TreeNode[]; selectedId?: string; expandedIds?: string[]; ariaLabel?: string; onSelect?: (item: TreeNode) => void; class?: string; }
	let { items, selectedId = $bindable(''), expandedIds = $bindable<string[]>([]), ariaLabel = 'Tree', onSelect, class: className, ...rest }: Props = $props();
	let treeRefs: HTMLButtonElement[] = [];

	function flatten(nodes: readonly TreeNode[], depth = 0, parentId?: string): VisibleNode[] {
		const output: VisibleNode[] = [];
		for (const node of nodes) {
			const hasChildren = Boolean(node.children?.length);
			const expanded = expandedIds.includes(node.id);
			output.push({ ...node, depth, parentId, hasChildren, expanded });
			if (hasChildren && expanded) output.push(...flatten(node.children ?? [], depth + 1, node.id));
		}
		return output;
	}

	const visibleNodes = $derived(flatten(items));

	function toggle(node: VisibleNode): void { if (!node.hasChildren) return; expandedIds = node.expanded ? expandedIds.filter((id) => id !== node.id) : [...expandedIds, node.id]; }
	function select(node: VisibleNode): void { if (node.disabled) return; selectedId = node.id; onSelect?.(node); }
	function focusIndex(index: number): void { treeRefs[index]?.focus(); }
	function handleKeydown(event: KeyboardEvent, index: number, node: VisibleNode): void {
		if (event.key === 'ArrowDown') { event.preventDefault(); focusIndex(Math.min(index + 1, visibleNodes.length - 1)); }
		else if (event.key === 'ArrowUp') { event.preventDefault(); focusIndex(Math.max(index - 1, 0)); }
		else if (event.key === 'ArrowRight') { event.preventDefault(); if (node.hasChildren && !node.expanded) toggle(node); else if (node.hasChildren) focusIndex(Math.min(index + 1, visibleNodes.length - 1)); }
		else if (event.key === 'ArrowLeft') { event.preventDefault(); if (node.expanded) toggle(node); else if (node.parentId) focusIndex(visibleNodes.findIndex((item) => item.id === node.parentId)); }
		else if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); select(node); }
	}
</script>

<div {...rest} class="atelier-tree {className ?? ''}" role="tree" aria-label={ariaLabel}>
	{#each visibleNodes as node, index (node.id)}
		<button bind:this={treeRefs[index]} type="button" class="atelier-tree__item" disabled={node.disabled} role="treeitem" aria-level={node.depth + 1} aria-selected={selectedId === node.id} aria-expanded={node.hasChildren ? node.expanded : undefined} data-state={selectedId === node.id ? 'selected' : 'idle'} style={`--tree-depth: ${node.depth}`} onclick={() => { select(node); toggle(node); }} onkeydown={(event) => handleKeydown(event, index, node)}>
			<span class="atelier-tree__chevron" data-expanded={node.expanded || undefined} aria-hidden="true">{node.hasChildren ? '›' : '·'}</span><span class="atelier-tree__label">{node.label}</span>
		</button>
	{/each}
</div>
