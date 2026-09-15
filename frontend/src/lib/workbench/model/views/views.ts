// Pure view model: identity, grouping, activation and open/close state.
// No Svelte, DOM or native imports. Every transition returns a new state and
// leaves the input untouched; impossible transitions return the input state.

export type ViewId = string;
export type GroupId = string;

export interface ViewDescriptor {
	readonly id: ViewId;
	readonly kind: string;
	readonly title: string;
	readonly description?: string;
	readonly closable?: boolean;
}

export interface ViewInstance extends ViewDescriptor {
	readonly pinned: boolean;
	readonly preview: boolean;
	readonly dirty: boolean;
}

export interface ViewGroup {
	readonly id: GroupId;
	readonly viewIds: readonly ViewId[];
	readonly activeViewId: ViewId | null;
	/** Most recently activated first. Drives which view becomes active on close. */
	readonly history: readonly ViewId[];
}

export interface ViewsState {
	readonly views: Readonly<Record<ViewId, ViewInstance>>;
	readonly groups: readonly ViewGroup[];
	readonly activeGroupId: GroupId;
}

export interface OpenViewOptions {
	readonly groupId?: GroupId;
	readonly preview?: boolean;
	readonly activate?: boolean;
}

export const DEFAULT_GROUP_ID: GroupId = 'group-1';

export function createViewsState(groupId: GroupId = DEFAULT_GROUP_ID): ViewsState {
	return { views: {}, groups: [emptyGroup(groupId)], activeGroupId: groupId };
}

function emptyGroup(id: GroupId): ViewGroup {
	return { id, viewIds: [], activeViewId: null, history: [] };
}

function groupIndex(state: ViewsState, groupId: GroupId): number {
	return state.groups.findIndex((group) => group.id === groupId);
}

function replaceGroup(state: ViewsState, group: ViewGroup): ViewsState {
	return { ...state, groups: state.groups.map((entry) => (entry.id === group.id ? group : entry)) };
}

function withHistoryFront(group: ViewGroup, viewId: ViewId): ViewGroup {
	return { ...group, activeViewId: viewId, history: [viewId, ...group.history.filter((id) => id !== viewId)] };
}

function pinnedCount(state: ViewsState, group: ViewGroup): number {
	let count = 0;
	for (const id of group.viewIds) {
		if (state.views[id]?.pinned) count += 1;
		else break;
	}
	return count;
}

export function groupOf(state: ViewsState, viewId: ViewId): ViewGroup | null {
	return state.groups.find((group) => group.viewIds.includes(viewId)) ?? null;
}

export function isOpen(state: ViewsState, viewId: ViewId): boolean {
	return groupOf(state, viewId) !== null;
}

export function viewsIn(state: ViewsState, groupId: GroupId): readonly ViewInstance[] {
	const group = state.groups.find((entry) => entry.id === groupId);
	if (!group) return [];
	return group.viewIds.flatMap((id) => (state.views[id] ? [state.views[id]] : []));
}

export function activeView(state: ViewsState): ViewInstance | null {
	const group = state.groups.find((entry) => entry.id === state.activeGroupId);
	if (!group?.activeViewId) return null;
	return state.views[group.activeViewId] ?? null;
}

export function dirtyViews(state: ViewsState): readonly ViewInstance[] {
	return Object.values(state.views).filter((view) => view.dirty);
}

export function nextGroupId(state: ViewsState): GroupId {
	let highest = 0;
	for (const group of state.groups) {
		const match = /^group-(\d+)$/.exec(group.id);
		if (match) highest = Math.max(highest, Number(match[1]));
	}
	return `group-${highest + 1}`;
}

export function openView(state: ViewsState, descriptor: ViewDescriptor, options: OpenViewOptions = {}): ViewsState {
	const activate = options.activate !== false;
	const existing = groupOf(state, descriptor.id);
	if (existing) {
		let next = state;
		if (!options.preview && state.views[descriptor.id].preview) next = promotePreview(next, descriptor.id);
		return activate ? activateView(next, descriptor.id) : next;
	}
	const targetId = options.groupId && groupIndex(state, options.groupId) >= 0 ? options.groupId : state.activeGroupId;
	let group = state.groups[groupIndex(state, targetId)];
	let views = state.views;
	let viewIds = [...group.viewIds];
	let history = group.history;
	if (options.preview) {
		const previewId = group.viewIds.find((id) => views[id]?.preview);
		if (previewId) {
			const position = viewIds.indexOf(previewId);
			viewIds.splice(position, 1, descriptor.id);
			const { [previewId]: _dropped, ...rest } = views;
			views = rest;
			history = history.filter((id) => id !== previewId);
		} else {
			viewIds.push(descriptor.id);
		}
	} else {
		viewIds.push(descriptor.id);
	}
	views = { ...views, [descriptor.id]: { ...descriptor, pinned: false, preview: Boolean(options.preview), dirty: false } };
	group = { ...group, viewIds, history, activeViewId: group.activeViewId && viewIds.includes(group.activeViewId) ? group.activeViewId : null };
	let next: ViewsState = replaceGroup({ ...state, views }, group);
	if (activate) next = activateView(next, descriptor.id);
	else if (!group.activeViewId) next = replaceGroup(next, withHistoryFront(group, descriptor.id));
	return next;
}

export function activateView(state: ViewsState, viewId: ViewId): ViewsState {
	const group = groupOf(state, viewId);
	if (!group) return state;
	if (group.activeViewId === viewId && state.activeGroupId === group.id && group.history[0] === viewId) return state;
	return { ...replaceGroup(state, withHistoryFront(group, viewId)), activeGroupId: group.id };
}

export function setActiveGroup(state: ViewsState, groupId: GroupId): ViewsState {
	if (groupIndex(state, groupId) < 0 || state.activeGroupId === groupId) return state;
	return { ...state, activeGroupId: groupId };
}

export function promotePreview(state: ViewsState, viewId: ViewId): ViewsState {
	const view = state.views[viewId];
	if (!view || !view.preview) return state;
	return { ...state, views: { ...state.views, [viewId]: { ...view, preview: false } } };
}

export function setDirty(state: ViewsState, viewId: ViewId, dirty: boolean): ViewsState {
	const view = state.views[viewId];
	if (!view || view.dirty === dirty) return state;
	return { ...state, views: { ...state.views, [viewId]: { ...view, dirty, preview: dirty ? false : view.preview } } };
}

export function pinView(state: ViewsState, viewId: ViewId, pinned: boolean): ViewsState {
	const view = state.views[viewId];
	const group = groupOf(state, viewId);
	if (!view || !group || view.pinned === pinned) return state;
	const others = group.viewIds.filter((id) => id !== viewId);
	const boundary = others.filter((id) => state.views[id]?.pinned).length;
	const viewIds = [...others.slice(0, boundary), viewId, ...others.slice(boundary)];
	const views = { ...state.views, [viewId]: { ...view, pinned, preview: pinned ? false : view.preview } };
	return replaceGroup({ ...state, views }, { ...group, viewIds });
}

function removeFromGroup(state: ViewsState, group: ViewGroup, viewId: ViewId): ViewGroup {
	const index = group.viewIds.indexOf(viewId);
	const viewIds = group.viewIds.filter((id) => id !== viewId);
	const history = group.history.filter((id) => id !== viewId);
	let activeViewId = group.activeViewId;
	if (activeViewId === viewId) {
		activeViewId = history.find((id) => viewIds.includes(id)) ?? viewIds[Math.min(index, viewIds.length - 1)] ?? null;
	}
	void state;
	return { ...group, viewIds, history, activeViewId };
}

export function closeView(state: ViewsState, viewId: ViewId): ViewsState {
	const view = state.views[viewId];
	const group = groupOf(state, viewId);
	if (!view || !group || view.closable === false) return state;
	const { [viewId]: _closed, ...views } = state.views;
	return replaceGroup({ ...state, views }, removeFromGroup(state, group, viewId));
}

export function closeViews(state: ViewsState, viewIds: readonly ViewId[]): ViewsState {
	return viewIds.reduce((next, id) => closeView(next, id), state);
}

export function closeOthers(state: ViewsState, viewId: ViewId): ViewsState {
	const group = groupOf(state, viewId);
	if (!group) return state;
	return closeViews(state, group.viewIds.filter((id) => id !== viewId && !state.views[id]?.pinned));
}

export function closeToRight(state: ViewsState, viewId: ViewId): ViewsState {
	const group = groupOf(state, viewId);
	if (!group) return state;
	const index = group.viewIds.indexOf(viewId);
	return closeViews(state, group.viewIds.slice(index + 1).filter((id) => !state.views[id]?.pinned));
}

export function closeGroupViews(state: ViewsState, groupId: GroupId): ViewsState {
	const group = state.groups.find((entry) => entry.id === groupId);
	if (!group) return state;
	return closeViews(state, group.viewIds);
}

export function reorderView(state: ViewsState, groupId: GroupId, from: number, to: number): ViewsState {
	const group = state.groups.find((entry) => entry.id === groupId);
	if (!group || from === to || from < 0 || from >= group.viewIds.length) return state;
	const moving = group.viewIds[from];
	const pinned = Boolean(state.views[moving]?.pinned);
	const remaining = group.viewIds.filter((_, index) => index !== from);
	const boundary = remaining.filter((id) => state.views[id]?.pinned).length;
	const lower = pinned ? 0 : boundary;
	const upper = pinned ? boundary : remaining.length;
	const target = Math.min(Math.max(to, lower), upper);
	const viewIds = [...remaining.slice(0, target), moving, ...remaining.slice(target)];
	return replaceGroup(state, { ...group, viewIds });
}

export function moveView(state: ViewsState, viewId: ViewId, targetGroupId: GroupId, index?: number): ViewsState {
	const source = groupOf(state, viewId);
	const target = state.groups.find((entry) => entry.id === targetGroupId);
	if (!source || !target) return state;
	if (source.id === target.id) return index === undefined ? state : reorderView(state, source.id, source.viewIds.indexOf(viewId), index);
	const trimmedSource = removeFromGroup(state, source, viewId);
	const pinned = Boolean(state.views[viewId]?.pinned);
	const boundary = pinnedCount(state, target);
	const position = Math.min(Math.max(index ?? target.viewIds.length, pinned ? 0 : boundary), pinned ? boundary : target.viewIds.length);
	const viewIds = [...target.viewIds.slice(0, position), viewId, ...target.viewIds.slice(position)];
	let next = replaceGroup(replaceGroup(state, trimmedSource), withHistoryFront({ ...target, viewIds }, viewId));
	next = { ...next, activeGroupId: target.id };
	return next;
}

export function addGroup(state: ViewsState, options: { afterGroupId?: GroupId; id?: GroupId; activate?: boolean } = {}): ViewsState {
	const id = options.id ?? nextGroupId(state);
	if (groupIndex(state, id) >= 0) return state;
	const anchor = options.afterGroupId ? groupIndex(state, options.afterGroupId) : groupIndex(state, state.activeGroupId);
	const position = anchor >= 0 ? anchor + 1 : state.groups.length;
	const groups = [...state.groups.slice(0, position), emptyGroup(id), ...state.groups.slice(position)];
	return { ...state, groups, activeGroupId: options.activate === false ? state.activeGroupId : id };
}

export function removeGroup(state: ViewsState, groupId: GroupId): ViewsState {
	const index = groupIndex(state, groupId);
	if (index < 0 || state.groups.length < 2) return state;
	const removed = state.groups[index];
	const neighborIndex = index > 0 ? index - 1 : index + 1;
	const neighbor = state.groups[neighborIndex];
	const boundary = pinnedCount(state, neighbor);
	const pinnedIds = removed.viewIds.filter((id) => state.views[id]?.pinned);
	const unpinnedIds = removed.viewIds.filter((id) => !state.views[id]?.pinned);
	const viewIds = [...neighbor.viewIds.slice(0, boundary), ...pinnedIds, ...neighbor.viewIds.slice(boundary), ...unpinnedIds];
	const merged: ViewGroup = { ...neighbor, viewIds, history: [...neighbor.history, ...removed.history], activeViewId: neighbor.activeViewId ?? removed.activeViewId };
	const groups = state.groups.filter((group) => group.id !== groupId).map((group) => (group.id === neighbor.id ? merged : group));
	return { ...state, groups, activeGroupId: state.activeGroupId === groupId ? neighbor.id : state.activeGroupId };
}
