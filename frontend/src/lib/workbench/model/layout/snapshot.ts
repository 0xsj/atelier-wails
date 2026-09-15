// Serializable workbench snapshot: what restoration persists and reads back.
// The reader is total over unknown input; it never trusts a cast.

import { createLayoutState, normalizeSizes, type LayoutState } from './layout';
import { DEFAULT_GROUP_ID, type GroupId, type ViewDescriptor, type ViewGroup, type ViewId, type ViewInstance, type ViewsState } from '../views/views';

export const SNAPSHOT_VERSION = 1 as const;

export interface SnapshotView extends ViewDescriptor {
	readonly pinned: boolean;
	readonly preview: boolean;
}

export interface SnapshotGroup {
	readonly id: GroupId;
	readonly viewIds: readonly ViewId[];
	readonly activeViewId: ViewId | null;
}

export interface WorkbenchSnapshot {
	readonly version: typeof SNAPSHOT_VERSION;
	readonly layout: LayoutState;
	readonly views: {
		readonly views: Readonly<Record<ViewId, SnapshotView>>;
		readonly groups: readonly SnapshotGroup[];
		readonly activeGroupId: GroupId;
	};
}

export type RestoreOutcome =
	| { readonly kind: 'restored'; readonly layout: LayoutState; readonly views: ViewsState; readonly dropped: readonly string[] }
	| { readonly kind: 'invalid'; readonly reason: string }
	| { readonly kind: 'unsupported'; readonly version: unknown };

/** Dirty state and activation history are runtime facts, not layout; they are not snapshotted. */
export function createSnapshot(layout: LayoutState, views: ViewsState): WorkbenchSnapshot {
	const snapshotViews: Record<ViewId, SnapshotView> = {};
	for (const view of Object.values(views.views)) {
		const { dirty: _dirty, ...rest } = view;
		snapshotViews[view.id] = rest;
	}
	return {
		version: SNAPSHOT_VERSION,
		layout,
		views: {
			views: snapshotViews,
			groups: views.groups.map((group) => ({ id: group.id, viewIds: [...group.viewIds], activeViewId: group.activeViewId })),
			activeGroupId: views.activeGroupId
		}
	};
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isString(value: unknown): value is string {
	return typeof value === 'string' && value.length > 0;
}

function optionalString(value: unknown): string | undefined {
	return typeof value === 'string' ? value : undefined;
}

function optionalBoolean(value: unknown, fallback: boolean): boolean {
	return typeof value === 'boolean' ? value : fallback;
}

function finiteNumber(value: unknown, fallback: number): number {
	return typeof value === 'number' && Number.isFinite(value) ? value : fallback;
}

function readLayout(input: unknown): LayoutState {
	if (!isRecord(input)) return createLayoutState();
	const sidebar = isRecord(input.sidebar) ? input.sidebar : {};
	const panel = isRecord(input.panel) ? input.panel : {};
	const paneWeights: Record<string, number> = {};
	if (isRecord(sidebar.paneWeights)) for (const [id, weight] of Object.entries(sidebar.paneWeights)) if (typeof weight === 'number' && Number.isFinite(weight) && weight > 0) paneWeights[id] = weight;
	return createLayoutState({
		sidebar: {
			visible: optionalBoolean(sidebar.visible, true),
			width: finiteNumber(sidebar.width, 240),
			activityId: typeof sidebar.activityId === 'string' ? sidebar.activityId : null,
			collapsedPanes: Array.isArray(sidebar.collapsedPanes) ? sidebar.collapsedPanes.filter(isString) : [],
			paneWeights
		},
		panel: { visible: optionalBoolean(panel.visible, false), height: finiteNumber(panel.height, 200) },
		editorSizes: Array.isArray(input.editorSizes) ? normalizeSizes(input.editorSizes.map((size) => finiteNumber(size, 0))) : [1],
		statusVisible: optionalBoolean(input.statusVisible, true)
	});
}

export function readSnapshot(input: unknown): RestoreOutcome {
	if (!isRecord(input)) return { kind: 'invalid', reason: 'snapshot is not an object' };
	if (input.version !== SNAPSHOT_VERSION) return { kind: 'unsupported', version: input.version };
	const section = input.views;
	if (!isRecord(section)) return { kind: 'invalid', reason: 'views section is missing' };
	if (!isRecord(section.views)) return { kind: 'invalid', reason: 'views map is missing' };
	if (!Array.isArray(section.groups)) return { kind: 'invalid', reason: 'groups list is missing' };

	const dropped: string[] = [];
	const views: Record<ViewId, ViewInstance> = {};
	for (const [id, raw] of Object.entries(section.views)) {
		if (!isRecord(raw) || !isString(raw.id) || raw.id !== id || !isString(raw.kind) || !isString(raw.title)) {
			dropped.push(`view ${id}`);
			continue;
		}
		views[id] = {
			id,
			kind: raw.kind,
			title: raw.title,
			description: optionalString(raw.description),
			closable: typeof raw.closable === 'boolean' ? raw.closable : undefined,
			pinned: optionalBoolean(raw.pinned, false),
			preview: optionalBoolean(raw.preview, false),
			dirty: false
		};
	}

	const seen = new Set<ViewId>();
	const groups: ViewGroup[] = [];
	for (const raw of section.groups) {
		if (!isRecord(raw) || !isString(raw.id) || !Array.isArray(raw.viewIds)) {
			dropped.push('group');
			continue;
		}
		if (groups.some((group) => group.id === raw.id)) {
			dropped.push(`group ${raw.id}`);
			continue;
		}
		const viewIds: ViewId[] = [];
		for (const viewId of raw.viewIds) {
			if (!isString(viewId) || !views[viewId] || seen.has(viewId)) {
				dropped.push(`reference ${String(viewId)}`);
				continue;
			}
			seen.add(viewId);
			viewIds.push(viewId);
		}
		const activeViewId = isString(raw.activeViewId) && viewIds.includes(raw.activeViewId) ? raw.activeViewId : (viewIds[0] ?? null);
		groups.push({ id: raw.id, viewIds, activeViewId, history: activeViewId ? [activeViewId] : [] });
	}
	for (const id of Object.keys(views)) {
		if (!seen.has(id)) {
			delete views[id];
			dropped.push(`orphan ${id}`);
		}
	}
	if (groups.length === 0) groups.push({ id: DEFAULT_GROUP_ID, viewIds: [], activeViewId: null, history: [] });
	const requestedGroup = section.activeGroupId;
	const activeGroupId = isString(requestedGroup) && groups.some((group) => group.id === requestedGroup) ? requestedGroup : groups[0].id;

	return { kind: 'restored', layout: readLayout(input.layout), views: { views, groups, activeGroupId }, dropped };
}
