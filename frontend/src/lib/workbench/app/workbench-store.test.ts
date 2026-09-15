import { describe, expect, it, vi } from 'vitest';
import { registerWorkbenchCommands, WORKBENCH_COMMANDS } from './commands/defaults';
import { loadSnapshot, createMemorySnapshotPort } from './restoration/restoration';
import { createWorkbenchStore } from './workbench-store';

const file = (id: string) => ({ id, kind: 'file', title: `${id}.ts` });

describe('workbench store', () => {
	it('notifies subscribers immediately and on change, keeping editor sizes in sync with groups', () => {
		const store = createWorkbenchStore({ platform: 'mac' });
		const seen = vi.fn();
		const stop = store.subscribe(seen);
		expect(seen).toHaveBeenCalledTimes(1);
		store.openView(file('a'));
		expect(seen).toHaveBeenCalledTimes(2);
		store.activateView('a');
		expect(seen).toHaveBeenCalledTimes(2);
		store.splitGroup();
		expect(store.state.layout.editorSizes).toHaveLength(2);
		store.removeGroup('group-2');
		expect(store.state.layout.editorSizes).toHaveLength(1);
		stop();
		store.openView(file('b'));
		expect(seen).toHaveBeenCalledTimes(4);
	});

	it('requires confirmation to close dirty views and never closes them silently', () => {
		const store = createWorkbenchStore({ platform: 'mac' });
		store.openView(file('a'));
		store.openView(file('b'));
		store.setDirty('a', true);
		expect(store.requestClose('a')).toMatchObject({ kind: 'needs-confirmation' });
		expect(store.requestClose('b')).toMatchObject({ kind: 'closed' });
		expect(store.requestClose('nope')).toEqual({ kind: 'refused', viewId: 'nope' });
		store.closeView('a');
		expect(store.state.views.views.a).toBeDefined();
		store.closeView('a', { force: true });
		expect(store.state.views.views.a).toBeUndefined();
	});

	it('closing many skips dirty views and returns them for the host', () => {
		const store = createWorkbenchStore({ platform: 'mac' });
		['a', 'b', 'c', 'd'].forEach((id) => store.openView(file(id)));
		store.setDirty('c', true);
		store.pinView('d', true);
		const skipped = store.closeOthers('a');
		expect(skipped.map((view) => view.id)).toEqual(['c']);
		expect(store.state.views.groups[0].viewIds).toEqual(['d', 'a', 'c']);
	});

	it('tracks navigation history across activations and closes', () => {
		const store = createWorkbenchStore({ platform: 'mac' });
		['a', 'b', 'c'].forEach((id) => store.openView(file(id)));
		expect(store.context().canGoBack).toBe(true);
		store.goBack();
		expect(store.state.views.groups[0].activeViewId).toBe('b');
		store.goBack();
		expect(store.state.views.groups[0].activeViewId).toBe('a');
		store.goForward();
		expect(store.state.views.groups[0].activeViewId).toBe('b');
		store.closeView('b', { force: true });
		expect(store.state.history.entries).toEqual(['a', 'c']);
	});

	it('dispatches default commands from shortcuts with availability from context', async () => {
		const store = createWorkbenchStore({ platform: 'mac' });
		const confirm = vi.fn();
		registerWorkbenchCommands(store, { onCloseNeedsConfirmation: confirm });
		const press = (key: string, extra: Partial<{ metaKey: boolean; altKey: boolean; target: { tagName: string } }> = {}) => store.handleKey({ key, metaKey: false, ctrlKey: false, shiftKey: false, altKey: false, ...extra });

		expect(await press('b', { metaKey: true })).toMatchObject({ kind: 'dispatched', outcome: { kind: 'executed' } });
		expect(store.state.layout.sidebar.visible).toBe(false);
		expect((await press('w', { metaKey: true })).kind).toBe('ignored');

		store.openView(file('a'));
		store.setDirty('a', true);
		await press('w', { metaKey: true, target: { tagName: 'INPUT' } });
		expect(confirm).toHaveBeenCalledWith('a');
		expect(store.state.views.views.a).toBeDefined();

		expect(await store.dispatch(WORKBENCH_COMMANDS.splitEditor)).toMatchObject({ kind: 'executed' });
		expect(store.state.views.groups).toHaveLength(2);
		expect((await store.dispatch(WORKBENCH_COMMANDS.forward)).kind).toBe('unavailable');
	});

	it('round-trips through a snapshot port', async () => {
		const port = createMemorySnapshotPort();
		const store = createWorkbenchStore({ platform: 'linux' });
		store.openView(file('a'));
		store.openView(file('b'));
		store.pinView('b', true);
		store.selectActivity('search');
		store.setEditorSizes([70, 30]);
		await port.save(store.snapshot());

		const fresh = createWorkbenchStore({ platform: 'linux', bounds: { width: 300, height: 300 } });
		expect(fresh.restore(await loadSnapshot(port))).toBe(true);
		expect(fresh.state.views.groups[0].viewIds).toEqual(['b', 'a']);
		expect(fresh.state.layout.sidebar.activityId).toBe('search');
		expect(fresh.state.layout.sidebar.width).toBe(170);
		expect(fresh.state.layout.editorSizes).toEqual([70]);
		expect(fresh.context().canGoBack).toBe(false);
		expect(fresh.restore({ kind: 'invalid', reason: 'x' })).toBe(false);
	});
});
