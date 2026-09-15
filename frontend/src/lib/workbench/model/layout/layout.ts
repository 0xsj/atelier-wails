// Pure layout model: region visibility and sizing rules. No DOM measurement.
// Sizes are plain numbers (CSS pixels for fixed regions, relative weights for
// splits); the UI decides how to render them.

export interface Bounds {
	readonly width: number;
	readonly height: number;
}

export interface SidebarLayout {
	readonly visible: boolean;
	readonly width: number;
	readonly activityId: string | null;
	readonly collapsedPanes: readonly string[];
	readonly paneWeights: Readonly<Record<string, number>>;
}

export interface PanelLayout {
	readonly visible: boolean;
	readonly height: number;
}

export interface LayoutState {
	readonly sidebar: SidebarLayout;
	readonly panel: PanelLayout;
	readonly editorSizes: readonly number[];
	readonly statusVisible: boolean;
}

export const LAYOUT_LIMITS = {
	sidebarMin: 170,
	sidebarMax: 640,
	sidebarMaxFraction: 0.5,
	panelMin: 96,
	panelMaxFraction: 0.8
} as const;

export const DEFAULT_LAYOUT: LayoutState = {
	sidebar: { visible: true, width: 240, activityId: null, collapsedPanes: [], paneWeights: {} },
	panel: { visible: false, height: 200 },
	editorSizes: [1],
	statusVisible: true
};

export function createLayoutState(overrides: Partial<LayoutState> = {}): LayoutState {
	return fitToBounds({ ...DEFAULT_LAYOUT, ...overrides, sidebar: { ...DEFAULT_LAYOUT.sidebar, ...overrides.sidebar }, panel: { ...DEFAULT_LAYOUT.panel, ...overrides.panel }, editorSizes: normalizeSizes(overrides.editorSizes ?? DEFAULT_LAYOUT.editorSizes) }, null);
}

function clamp(value: number, min: number, max: number): number {
	if (!Number.isFinite(value)) return min;
	return Math.min(Math.max(value, min), Math.max(min, max));
}

export function clampSidebarWidth(width: number, bounds: Bounds | null): number {
	const max = bounds ? Math.min(LAYOUT_LIMITS.sidebarMax, bounds.width * LAYOUT_LIMITS.sidebarMaxFraction) : LAYOUT_LIMITS.sidebarMax;
	return Math.round(clamp(width, LAYOUT_LIMITS.sidebarMin, max));
}

export function clampPanelHeight(height: number, bounds: Bounds | null): number {
	const max = bounds ? bounds.height * LAYOUT_LIMITS.panelMaxFraction : Number.POSITIVE_INFINITY;
	return Math.round(clamp(height, LAYOUT_LIMITS.panelMin, max));
}

export function normalizeSizes(sizes: readonly number[]): readonly number[] {
	const cleaned = sizes.map((size) => (Number.isFinite(size) && size > 0 ? size : 0));
	if (cleaned.length === 0) return [];
	if (cleaned.every((size) => size === 0)) return cleaned.map(() => 1);
	return cleaned;
}

export function toggleSidebar(state: LayoutState, visible = !state.sidebar.visible): LayoutState {
	if (state.sidebar.visible === visible) return state;
	return { ...state, sidebar: { ...state.sidebar, visible } };
}

export function setSidebarWidth(state: LayoutState, width: number, bounds: Bounds | null = null): LayoutState {
	const next = clampSidebarWidth(width, bounds);
	if (next === state.sidebar.width) return state;
	return { ...state, sidebar: { ...state.sidebar, width: next } };
}

/** Selecting the active activity again hides the sidebar; selecting it while hidden shows it. */
export function selectActivity(state: LayoutState, activityId: string): LayoutState {
	if (state.sidebar.activityId === activityId) return toggleSidebar(state);
	return { ...state, sidebar: { ...state.sidebar, activityId, visible: true } };
}

export function setPaneCollapsed(state: LayoutState, paneId: string, collapsed: boolean): LayoutState {
	const has = state.sidebar.collapsedPanes.includes(paneId);
	if (has === collapsed) return state;
	const collapsedPanes = collapsed ? [...state.sidebar.collapsedPanes, paneId] : state.sidebar.collapsedPanes.filter((id) => id !== paneId);
	return { ...state, sidebar: { ...state.sidebar, collapsedPanes } };
}

export function togglePane(state: LayoutState, paneId: string): LayoutState {
	return setPaneCollapsed(state, paneId, !state.sidebar.collapsedPanes.includes(paneId));
}

export function setPaneWeights(state: LayoutState, weights: Readonly<Record<string, number>>): LayoutState {
	const paneWeights: Record<string, number> = {};
	for (const [id, weight] of Object.entries(weights)) if (Number.isFinite(weight) && weight > 0) paneWeights[id] = weight;
	return { ...state, sidebar: { ...state.sidebar, paneWeights } };
}

export function togglePanel(state: LayoutState, visible = !state.panel.visible): LayoutState {
	if (state.panel.visible === visible) return state;
	return { ...state, panel: { ...state.panel, visible } };
}

export function setPanelHeight(state: LayoutState, height: number, bounds: Bounds | null = null): LayoutState {
	const next = clampPanelHeight(height, bounds);
	if (next === state.panel.height) return state;
	return { ...state, panel: { ...state.panel, height: next } };
}

export function toggleStatus(state: LayoutState, visible = !state.statusVisible): LayoutState {
	if (state.statusVisible === visible) return state;
	return { ...state, statusVisible: visible };
}

export function setEditorSizes(state: LayoutState, sizes: readonly number[]): LayoutState {
	return { ...state, editorSizes: normalizeSizes(sizes) };
}

/** Keeps one size per editor group: new groups take the average share, removed groups drop their share. */
export function syncEditorSizes(state: LayoutState, groupCount: number): LayoutState {
	const current = state.editorSizes;
	if (groupCount === current.length) return state;
	if (groupCount <= 0) return { ...state, editorSizes: [] };
	if (groupCount < current.length) return { ...state, editorSizes: normalizeSizes(current.slice(0, groupCount)) };
	const positive = current.filter((size) => size > 0);
	const share = positive.length ? positive.reduce((sum, size) => sum + size, 0) / positive.length : 1;
	return { ...state, editorSizes: normalizeSizes([...current, ...Array.from({ length: groupCount - current.length }, () => share)]) };
}

export function fitToBounds(state: LayoutState, bounds: Bounds | null): LayoutState {
	const width = clampSidebarWidth(state.sidebar.width, bounds);
	const height = clampPanelHeight(state.panel.height, bounds);
	if (width === state.sidebar.width && height === state.panel.height) return state;
	return { ...state, sidebar: { ...state.sidebar, width }, panel: { ...state.panel, height } };
}
