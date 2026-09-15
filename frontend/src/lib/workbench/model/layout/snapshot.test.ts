import { describe, expect, it } from 'vitest';
import { createLayoutState, selectActivity } from './layout';
import { createSnapshot, readSnapshot, SNAPSHOT_VERSION } from './snapshot';
import { addGroup, createViewsState, openView, pinView, setDirty } from '../views/views';

describe('workbench snapshot', () => {
	it('round-trips layout and views, dropping dirty state and history', () => {
		let views = openView(createViewsState(), { id: 'a', kind: 'file', title: 'a.ts' });
		views = openView(views, { id: 'b', kind: 'file', title: 'b.ts', description: 'src/b.ts' }, { preview: true });
		views = pinView(views, 'a', true);
		views = setDirty(views, 'a', true);
		views = addGroup(views);
		const layout = selectActivity(createLayoutState(), 'files');
		const snapshot = createSnapshot(layout, views);
		expect(snapshot.version).toBe(SNAPSHOT_VERSION);
		expect('dirty' in snapshot.views.views.a).toBe(false);

		const outcome = readSnapshot(JSON.parse(JSON.stringify(snapshot)));
		expect(outcome.kind).toBe('restored');
		if (outcome.kind !== 'restored') return;
		expect(outcome.dropped).toEqual([]);
		expect(outcome.layout).toEqual(layout);
		expect(outcome.views.views.a).toMatchObject({ pinned: true, dirty: false });
		expect(outcome.views.views.b).toMatchObject({ preview: true, description: 'src/b.ts' });
		expect(outcome.views.groups.map((group) => group.id)).toEqual(['group-1', 'group-2']);
		expect(outcome.views.activeGroupId).toBe('group-2');
		expect(outcome.views.groups[0].history).toEqual(['b']);
	});

	it('rejects non-objects and unsupported versions', () => {
		expect(readSnapshot(null)).toEqual({ kind: 'invalid', reason: 'snapshot is not an object' });
		expect(readSnapshot('nope').kind).toBe('invalid');
		expect(readSnapshot({ version: 2 })).toEqual({ kind: 'unsupported', version: 2 });
		expect(readSnapshot({ version: 1 }).kind).toBe('invalid');
	});

	it('prunes dangling references, duplicates and malformed views while keeping the rest', () => {
		const outcome = readSnapshot({
			version: 1,
			layout: { sidebar: { width: 'wide', collapsedPanes: ['outline', 3] }, editorSizes: [2, 'x'] },
			views: {
				views: {
					a: { id: 'a', kind: 'file', title: 'a.ts' },
					b: { id: 'b', kind: 'file' },
					c: { id: 'c', kind: 'file', title: 'c.ts' }
				},
				groups: [
					{ id: 'g1', viewIds: ['a', 'missing', 'a'], activeViewId: 'missing' },
					{ id: 'g1', viewIds: [] },
					{ id: 'g2', viewIds: [] }
				],
				activeGroupId: 'nope'
			}
		});
		expect(outcome.kind).toBe('restored');
		if (outcome.kind !== 'restored') return;
		expect(outcome.views.groups.map((group) => group.id)).toEqual(['g1', 'g2']);
		expect(outcome.views.groups[0]).toMatchObject({ viewIds: ['a'], activeViewId: 'a' });
		expect(Object.keys(outcome.views.views)).toEqual(['a']);
		expect(outcome.views.activeGroupId).toBe('g1');
		expect(outcome.dropped).toEqual(['view b', 'reference missing', 'reference a', 'group g1', 'orphan c']);
		expect(outcome.layout.sidebar.width).toBe(240);
		expect(outcome.layout.sidebar.collapsedPanes).toEqual(['outline']);
		expect(outcome.layout.editorSizes).toEqual([2, 0]);
	});

	it('always restores at least one group', () => {
		const outcome = readSnapshot({ version: 1, views: { views: {}, groups: [] } });
		expect(outcome.kind).toBe('restored');
		if (outcome.kind === 'restored') expect(outcome.views.groups).toHaveLength(1);
	});
});
