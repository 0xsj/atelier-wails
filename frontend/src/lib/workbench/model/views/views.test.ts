import { describe, expect, it } from 'vitest';
import {
	activateView,
	activeView,
	addGroup,
	closeGroupViews,
	closeOthers,
	closeToRight,
	closeView,
	createViewsState,
	dirtyViews,
	groupOf,
	moveView,
	openView,
	pinView,
	removeGroup,
	reorderView,
	setDirty,
	viewsIn,
	type ViewDescriptor
} from './views';

const file = (id: string, extra: Partial<ViewDescriptor> = {}): ViewDescriptor => ({ id, kind: 'file', title: `${id}.ts`, ...extra });

describe('views model', () => {
	it('opens a view into the active group and activates it', () => {
		const state = openView(createViewsState(), file('a'));
		expect(viewsIn(state, 'group-1').map((view) => view.id)).toEqual(['a']);
		expect(activeView(state)?.id).toBe('a');
		expect(state.groups[0].history).toEqual(['a']);
	});

	it('does not mutate the input state', () => {
		const initial = createViewsState();
		const frozen = JSON.stringify(initial);
		openView(initial, file('a'));
		expect(JSON.stringify(initial)).toBe(frozen);
	});

	it('re-opening an open view activates it instead of duplicating', () => {
		let state = openView(createViewsState(), file('a'));
		state = openView(state, file('b'));
		state = openView(state, file('a'));
		expect(state.groups[0].viewIds).toEqual(['a', 'b']);
		expect(activeView(state)?.id).toBe('a');
	});

	it('a preview view is replaced by the next preview and promoted when opened for real', () => {
		let state = openView(createViewsState(), file('a'), { preview: true });
		state = openView(state, file('b'), { preview: true });
		expect(state.groups[0].viewIds).toEqual(['b']);
		expect(state.views.a).toBeUndefined();
		state = openView(state, file('b'));
		expect(state.views.b.preview).toBe(false);
		state = openView(state, file('c'), { preview: true });
		expect(state.groups[0].viewIds).toEqual(['b', 'c']);
	});

	it('closing the active view activates the most recently used remaining view', () => {
		let state = ['a', 'b', 'c'].reduce((next, id) => openView(next, file(id)), createViewsState());
		state = activateView(state, 'a');
		state = activateView(state, 'c');
		state = closeView(state, 'c');
		expect(activeView(state)?.id).toBe('a');
		expect(state.views.c).toBeUndefined();
	});

	it('falls back to the neighbor when no history remains', () => {
		let state = ['a', 'b', 'c'].reduce((next, id) => openView(next, file(id)), createViewsState());
		state = { ...state, groups: [{ ...state.groups[0], history: [], activeViewId: 'b' }] };
		state = closeView(state, 'b');
		expect(activeView(state)?.id).toBe('c');
		state = closeView(state, 'c');
		expect(activeView(state)?.id).toBe('a');
		state = closeView(state, 'a');
		expect(activeView(state)).toBeNull();
	});

	it('refuses to close a non-closable view', () => {
		const state = openView(createViewsState(), file('settings', { closable: false }));
		expect(closeView(state, 'settings')).toBe(state);
	});

	it('close others and close to the right keep pinned views', () => {
		let state = ['a', 'b', 'c', 'd'].reduce((next, id) => openView(next, file(id)), createViewsState());
		state = pinView(state, 'd', true);
		expect(state.groups[0].viewIds).toEqual(['d', 'a', 'b', 'c']);
		const others = closeOthers(state, 'b');
		expect(others.groups[0].viewIds).toEqual(['d', 'b']);
		const right = closeToRight(state, 'a');
		expect(right.groups[0].viewIds).toEqual(['d', 'a']);
		expect(closeGroupViews(state, 'group-1').groups[0].viewIds).toEqual([]);
	});

	it('pinning moves a view to the pinned zone and unpinning moves it after it', () => {
		let state = ['a', 'b', 'c'].reduce((next, id) => openView(next, file(id)), createViewsState());
		state = pinView(state, 'c', true);
		state = pinView(state, 'b', true);
		expect(state.groups[0].viewIds).toEqual(['c', 'b', 'a']);
		state = pinView(state, 'c', false);
		expect(state.groups[0].viewIds).toEqual(['b', 'c', 'a']);
	});

	it('reorder clamps into the pinned or unpinned zone', () => {
		let state = ['a', 'b', 'c'].reduce((next, id) => openView(next, file(id)), createViewsState());
		state = pinView(state, 'a', true);
		expect(reorderView(state, 'group-1', 2, 0).groups[0].viewIds).toEqual(['a', 'c', 'b']);
		expect(reorderView(state, 'group-1', 0, 2).groups[0].viewIds).toEqual(['a', 'b', 'c']);
		expect(reorderView(state, 'group-1', 1, 2).groups[0].viewIds).toEqual(['a', 'c', 'b']);
	});

	it('dirty views drop preview and are listed', () => {
		let state = openView(createViewsState(), file('a'), { preview: true });
		state = setDirty(state, 'a', true);
		expect(state.views.a.preview).toBe(false);
		expect(dirtyViews(state).map((view) => view.id)).toEqual(['a']);
		expect(setDirty(state, 'missing', true)).toBe(state);
	});

	it('adds a group after the active one and moves views between groups', () => {
		let state = ['a', 'b'].reduce((next, id) => openView(next, file(id)), createViewsState());
		state = addGroup(state);
		expect(state.groups.map((group) => group.id)).toEqual(['group-1', 'group-2']);
		expect(state.activeGroupId).toBe('group-2');
		state = moveView(state, 'a', 'group-2');
		expect(groupOf(state, 'a')?.id).toBe('group-2');
		expect(activeView(state)?.id).toBe('a');
		expect(viewsIn(state, 'group-1').map((view) => view.id)).toEqual(['b']);
		expect(state.groups[0].activeViewId).toBe('b');
	});

	it('removing a group merges its views into the neighbor and never removes the last group', () => {
		let state = openView(createViewsState(), file('a'));
		state = addGroup(state);
		state = openView(state, file('b'));
		state = removeGroup(state, 'group-2');
		expect(state.groups.map((group) => group.id)).toEqual(['group-1']);
		expect(state.groups[0].viewIds).toEqual(['a', 'b']);
		expect(state.activeGroupId).toBe('group-1');
		expect(removeGroup(state, 'group-1')).toBe(state);
	});
});
