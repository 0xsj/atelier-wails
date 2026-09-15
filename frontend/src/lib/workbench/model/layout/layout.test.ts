import { describe, expect, it } from 'vitest';
import { LAYOUT_LIMITS, createLayoutState, fitToBounds, selectActivity, setEditorSizes, setPanelHeight, setPaneWeights, setSidebarWidth, syncEditorSizes, togglePane, toggleSidebar } from './layout';

describe('layout model', () => {
	it('clamps the sidebar width to limits and to half the window', () => {
		const state = createLayoutState();
		expect(setSidebarWidth(state, 10).sidebar.width).toBe(LAYOUT_LIMITS.sidebarMin);
		expect(setSidebarWidth(state, 5000).sidebar.width).toBe(LAYOUT_LIMITS.sidebarMax);
		expect(setSidebarWidth(state, 500, { width: 800, height: 600 }).sidebar.width).toBe(400);
		expect(setSidebarWidth(state, Number.NaN).sidebar.width).toBe(LAYOUT_LIMITS.sidebarMin);
		expect(setSidebarWidth(state, state.sidebar.width)).toBe(state);
	});

	it('selecting the active activity toggles the sidebar, another activity shows it', () => {
		let state = selectActivity(createLayoutState(), 'files');
		expect(state.sidebar).toMatchObject({ activityId: 'files', visible: true });
		state = selectActivity(state, 'files');
		expect(state.sidebar.visible).toBe(false);
		state = selectActivity(state, 'search');
		expect(state.sidebar).toMatchObject({ activityId: 'search', visible: true });
		expect(toggleSidebar(state, true)).toBe(state);
	});

	it('tracks collapsed panes and keeps only positive weights', () => {
		let state = togglePane(createLayoutState(), 'outline');
		expect(state.sidebar.collapsedPanes).toEqual(['outline']);
		state = togglePane(state, 'outline');
		expect(state.sidebar.collapsedPanes).toEqual([]);
		state = setPaneWeights(state, { explorer: 2, outline: 0, timeline: Number.NaN });
		expect(state.sidebar.paneWeights).toEqual({ explorer: 2 });
	});

	it('normalizes editor sizes and syncs them to the group count', () => {
		let state = setEditorSizes(createLayoutState(), [0, -3, Number.NaN]);
		expect(state.editorSizes).toEqual([1, 1, 1]);
		state = setEditorSizes(state, [60, 40]);
		expect(syncEditorSizes(state, 3).editorSizes).toEqual([60, 40, 50]);
		expect(syncEditorSizes(state, 1).editorSizes).toEqual([60]);
		expect(syncEditorSizes(state, 0).editorSizes).toEqual([]);
		expect(syncEditorSizes(state, 2)).toBe(state);
		expect(setEditorSizes(state, [70, 0]).editorSizes).toEqual([70, 0]);
	});

	it('fits the panel and sidebar to smaller bounds', () => {
		let state = setPanelHeight(createLayoutState(), 900);
		expect(state.panel.height).toBe(900);
		state = fitToBounds(state, { width: 400, height: 500 });
		expect(state.panel.height).toBe(400);
		expect(state.sidebar.width).toBe(200);
		expect(fitToBounds(state, { width: 400, height: 500 })).toBe(state);
	});
});
