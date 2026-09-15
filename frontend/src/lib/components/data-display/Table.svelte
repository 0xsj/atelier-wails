<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import './data-display.css';

	export type TableCellValue = string | number | boolean | null | undefined;
	export type TableColumn = { key: string; label: string; align?: 'start' | 'center' | 'end'; width?: string };
	export type TableRow = Record<string, TableCellValue> & { id?: string };

	interface Props extends Omit<HTMLAttributes<HTMLDivElement>, 'children' | 'class'> {
		columns: readonly TableColumn[];
		rows: readonly TableRow[];
		caption?: string;
		ariaLabel?: string;
		emptyLabel?: string;
		compact?: boolean;
		class?: string;
	}

	let { columns, rows, caption, ariaLabel = 'Data table', emptyLabel = 'No data available', compact = false, class: className, ...rest }: Props = $props();

	function cellText(value: TableCellValue): string {
		return value === null || value === undefined || value === '' ? '—' : String(value);
	}
</script>

<div {...rest} class="atelier-table-wrap {className ?? ''}">
	<table class="atelier-table" data-compact={compact} aria-label={ariaLabel}>
		{#if caption}<caption>{caption}</caption>{/if}
		<thead>
			<tr>{#each columns as column}<th scope="col" data-align={column.align ?? 'start'} style={column.width ? `inline-size: ${column.width}` : undefined}>{column.label}</th>{/each}</tr>
		</thead>
		<tbody>
			{#if rows.length}
				{#each rows as row, rowIndex (row.id ?? rowIndex)}
					<tr>{#each columns as column}<td data-align={column.align ?? 'start'}>{cellText(row[column.key])}</td>{/each}</tr>
				{/each}
			{:else}
				<tr><td class="atelier-table__empty" colspan={columns.length}>{emptyLabel}</td></tr>
			{/if}
		</tbody>
	</table>
</div>
