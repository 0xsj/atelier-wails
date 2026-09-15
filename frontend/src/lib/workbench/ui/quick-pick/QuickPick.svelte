<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Dialog as Primitive } from 'bits-ui';
	import { Spinner } from '../../../components/feedback';
	import './quick-pick.css';

	export type QuickPickAction = { id: string; label: string; glyph?: string };
	export type QuickPickItem = {
		id: string;
		label: string;
		description?: string;
		detail?: string;
		shortcut?: string;
		group?: string;
		recent?: boolean;
		disabled?: boolean;
		actions?: readonly QuickPickAction[];
	};
	export type QuickPickMode = { prefix: string; label: string; placeholder: string };

	interface Props {
		open?: boolean;
		items: readonly QuickPickItem[];
		modes?: readonly QuickPickMode[];
		mode?: string;
		query?: string;
		title?: string;
		placeholder?: string;
		emptyText?: string;
		recentLabel?: string;
		groupOrder?: readonly string[];
		loading?: boolean;
		filter?: (item: QuickPickItem, query: string, mode: string) => boolean;
		icon?: Snippet<[QuickPickItem]>;
		onSelect?: (item: QuickPickItem, mode: string) => void;
		onAction?: (item: QuickPickItem, action: QuickPickAction, mode: string) => void;
		onModeChange?: (mode: string) => void;
		class?: string;
	}

	let {
		open = $bindable(false),
		items,
		modes = [],
		mode = $bindable(''),
		query = $bindable(''),
		title = 'Quick pick',
		placeholder = 'Type to filter',
		emptyText = 'No matching results',
		recentLabel = 'Recently used',
		groupOrder = [],
		loading = false,
		filter,
		icon,
		onSelect,
		onAction,
		onModeChange,
		class: className
	}: Props = $props();

	const listId = $props.id();
	let selectedIndex = $state(0);
	let input = $state<HTMLInputElement | null>(null);
	let list = $state<HTMLDivElement | null>(null);

	const activeMode = $derived(modes.find((entry) => entry.prefix === mode));
	const prefixedModes = $derived(modes.filter((entry) => entry.prefix !== ''));
	const resolvedPlaceholder = $derived(activeMode?.placeholder ?? placeholder);

	function defaultFilter(item: QuickPickItem, text: string): boolean {
		if (!text) return true;
		const haystack = `${item.label} ${item.description ?? ''} ${item.detail ?? ''}`.toLowerCase();
		return text.toLowerCase().split(/\s+/).filter(Boolean).every((token) => haystack.includes(token));
	}

	const filtered = $derived(items.filter((item) => (filter ? filter(item, query.trim(), mode) : defaultFilter(item, query.trim()))));

	type Group = { name: string | undefined; items: { item: QuickPickItem; index: number }[] };
	const groups = $derived.by<Group[]>(() => {
		const map = new Map<string | undefined, Group>();
		const orderedNames: (string | undefined)[] = [];
		const named = (item: QuickPickItem) => item.group ?? (item.recent && !query.trim() ? recentLabel : undefined);
		filtered.forEach((item, index) => {
			const name = named(item);
			if (!map.has(name)) {
				map.set(name, { name, items: [] });
				orderedNames.push(name);
			}
			map.get(name)?.items.push({ item, index });
		});
		const rank = (name: string | undefined) => {
			if (name === recentLabel) return -1;
			if (name === undefined) return groupOrder.length + 1;
			const position = groupOrder.indexOf(name);
			return position === -1 ? groupOrder.length : position;
		};
		return orderedNames.map((name) => map.get(name) as Group).sort((a, b) => rank(a.name) - rank(b.name));
	});
	const ordered = $derived(groups.flatMap((group) => group.items));

	$effect(() => {
		query;
		mode;
		selectedIndex = 0;
	});

	$effect(() => {
		if (selectedIndex >= ordered.length) selectedIndex = Math.max(ordered.length - 1, 0);
	});

	$effect(() => {
		if (!open) return;
		selectedIndex;
		list?.querySelector<HTMLElement>('[aria-selected="true"]')?.scrollIntoView({ block: 'nearest' });
	});

	function setMode(next: string): void {
		if (next === mode) return;
		mode = next;
		onModeChange?.(next);
	}

	function handleInput(event: Event): void {
		const value = (event.currentTarget as HTMLInputElement).value;
		const match = prefixedModes.find((entry) => value.startsWith(entry.prefix));
		if (match && mode === '') {
			setMode(match.prefix);
			query = value.slice(match.prefix.length);
			return;
		}
		query = value;
	}

	function select(position: number): void {
		const entry = ordered[position];
		if (!entry || entry.item.disabled) return;
		onSelect?.(entry.item, mode);
		open = false;
	}

	function runAction(entry: QuickPickItem, action: QuickPickAction, event: MouseEvent): void {
		event.stopPropagation();
		onAction?.(entry, action, mode);
	}

	function handleKeydown(event: KeyboardEvent): void {
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			selectedIndex = ordered.length ? (selectedIndex + 1) % ordered.length : 0;
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			selectedIndex = ordered.length ? (selectedIndex - 1 + ordered.length) % ordered.length : 0;
		} else if (event.key === 'Home') {
			event.preventDefault();
			selectedIndex = 0;
		} else if (event.key === 'End') {
			event.preventDefault();
			selectedIndex = Math.max(ordered.length - 1, 0);
		} else if (event.key === 'Enter') {
			event.preventDefault();
			select(selectedIndex);
		} else if (event.key === 'Backspace' && query === '' && mode !== '') {
			event.preventDefault();
			setMode('');
		}
	}

	function optionId(position: number): string {
		return `${listId}-option-${position}`;
	}
</script>

<Primitive.Root bind:open>
	<Primitive.Portal>
		<Primitive.Overlay class="atelier-quick-pick__overlay" />
		<Primitive.Content class="atelier-quick-pick {className ?? ''}" onkeydown={handleKeydown}>
			<Primitive.Title class="atelier-quick-pick__sr">{activeMode?.label ?? title}</Primitive.Title>
			<Primitive.Description class="atelier-quick-pick__sr">Type to filter, use the arrow keys to move, Enter to choose, Escape to close.</Primitive.Description>
			<div class="atelier-quick-pick__field">
				{#if activeMode && activeMode.prefix !== ''}<span class="atelier-quick-pick__mode">{activeMode.label}</span>{/if}
				<!-- svelte-ignore a11y_autofocus -->
				<input
					bind:this={input}
					class="atelier-quick-pick__input"
					type="text"
					role="combobox"
					value={query}
					placeholder={resolvedPlaceholder}
					aria-label={activeMode?.label ?? title}
					aria-expanded="true"
					aria-controls={listId}
					aria-activedescendant={ordered.length ? optionId(selectedIndex) : undefined}
					aria-autocomplete="list"
					autocomplete="off"
					spellcheck="false"
					autofocus
					oninput={handleInput}
				/>
				{#if loading}<Spinner size="sm" label="Loading results" />{/if}
			</div>
			{#if prefixedModes.length}
				<div class="atelier-quick-pick__modes" aria-label="Modes">
					<button type="button" class="atelier-quick-pick__mode-chip" data-active={mode === '' || undefined} onclick={() => { setMode(''); input?.focus(); }}>{modes.find((entry) => entry.prefix === '')?.label ?? 'All'}</button>
					{#each prefixedModes as entry (entry.prefix)}<button type="button" class="atelier-quick-pick__mode-chip" data-active={mode === entry.prefix || undefined} onclick={() => { setMode(entry.prefix); input?.focus(); }}><kbd>{entry.prefix}</kbd>{entry.label}</button>{/each}
				</div>
			{/if}
			<div bind:this={list} id={listId} class="atelier-quick-pick__list" role="listbox" aria-label={activeMode?.label ?? title}>
				{#each groups as group (group.name ?? '')}
					{#if group.name}<div class="atelier-quick-pick__group" role="presentation">{group.name}</div>{/if}
					{#each group.items as { item, index } (item.id)}
						{@const position = ordered.findIndex((entry) => entry.index === index)}
						<!-- Focus stays in the input; the active option is announced through aria-activedescendant. -->
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<!-- svelte-ignore a11y_interactive_supports_focus -->
						<div id={optionId(position)} class="atelier-quick-pick__item" role="option" aria-selected={selectedIndex === position} aria-disabled={item.disabled || undefined} data-disabled={item.disabled || undefined} onmousemove={() => (selectedIndex = position)} onclick={() => select(position)}>
							{#if icon}<span class="atelier-quick-pick__icon" aria-hidden="true">{@render icon(item)}</span>{/if}
							<span class="atelier-quick-pick__copy"><span class="atelier-quick-pick__label">{item.label}{#if item.description}<span class="atelier-quick-pick__description">{item.description}</span>{/if}</span>{#if item.detail}<span class="atelier-quick-pick__detail">{item.detail}</span>{/if}</span>
							{#if item.shortcut}<kbd class="atelier-quick-pick__shortcut">{item.shortcut}</kbd>{/if}
							{#if item.actions?.length}<span class="atelier-quick-pick__actions">{#each item.actions as action (action.id)}<button type="button" class="atelier-quick-pick__action" aria-label={action.label} title={action.label} tabindex="-1" onclick={(event) => runAction(item, action, event)}>{action.glyph ?? '•'}</button>{/each}</span>{/if}
						</div>
					{/each}
				{:else}
					<p class="atelier-quick-pick__empty">{loading ? 'Searching…' : emptyText}</p>
				{/each}
			</div>
		</Primitive.Content>
	</Primitive.Portal>
</Primitive.Root>
