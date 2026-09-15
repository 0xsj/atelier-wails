// Linear navigation history of view activations for back and forward.
// Pure values; the store prunes entries when views close.

import type { ViewId } from '../../model/views/views';

export interface NavigationHistory {
	readonly entries: readonly ViewId[];
	readonly index: number;
}

export const EMPTY_HISTORY: NavigationHistory = { entries: [], index: -1 };

export const HISTORY_LIMIT = 50;

/** Records an activation, discarding any forward entries and consecutive duplicates. */
export function recordVisit(history: NavigationHistory, viewId: ViewId): NavigationHistory {
	if (history.entries[history.index] === viewId) return history;
	const entries = [...history.entries.slice(0, history.index + 1), viewId].slice(-HISTORY_LIMIT);
	return { entries, index: entries.length - 1 };
}

export function canGoBack(history: NavigationHistory): boolean {
	return history.index > 0;
}

export function canGoForward(history: NavigationHistory): boolean {
	return history.index >= 0 && history.index < history.entries.length - 1;
}

export function goBack(history: NavigationHistory): NavigationHistory {
	return canGoBack(history) ? { ...history, index: history.index - 1 } : history;
}

export function goForward(history: NavigationHistory): NavigationHistory {
	return canGoForward(history) ? { ...history, index: history.index + 1 } : history;
}

export function currentEntry(history: NavigationHistory): ViewId | null {
	return history.entries[history.index] ?? null;
}

/** Drops a closed view, merging neighbors that become consecutive duplicates. */
export function forgetView(history: NavigationHistory, viewId: ViewId): NavigationHistory {
	if (!history.entries.includes(viewId)) return history;
	const entries: ViewId[] = [];
	let index = history.index;
	history.entries.forEach((entry, position) => {
		const dropped = entry === viewId || entries[entries.length - 1] === entry;
		if (dropped) {
			if (position <= history.index) index -= 1;
			return;
		}
		entries.push(entry);
	});
	return { entries, index: Math.min(Math.max(index, entries.length ? 0 : -1), entries.length - 1) };
}
