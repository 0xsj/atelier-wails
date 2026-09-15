import { describe, expect, it, vi } from 'vitest';
import { createLayoutState } from '../../model/layout/layout';
import { createSnapshot } from '../../model/layout/snapshot';
import { createViewsState, openView } from '../../model/views/views';
import { createDebouncedSaver, createJsonSnapshotPort, createMemorySnapshotPort, loadSnapshot } from './restoration';

const snapshot = () => createSnapshot(createLayoutState(), openView(createViewsState(), { id: 'a', kind: 'file', title: 'a' }));

describe('restoration', () => {
	it('reports empty, unavailable, invalid and restored loads', async () => {
		expect(await loadSnapshot(createMemorySnapshotPort())).toEqual({ kind: 'empty' });
		expect((await loadSnapshot({ load: async () => 'garbage', save: async () => {} })).kind).toBe('invalid');
		const failing = { load: async () => { throw new Error('locked'); }, save: async () => {} };
		expect((await loadSnapshot(failing)).kind).toBe('unavailable');
		const port = createMemorySnapshotPort();
		await port.save(snapshot());
		const outcome = await loadSnapshot(port);
		expect(outcome.kind).toBe('restored');
		if (outcome.kind === 'restored') expect(Object.keys(outcome.views.views)).toEqual(['a']);
	});

	it('the memory port stores copies, not references', async () => {
		const port = createMemorySnapshotPort();
		const value = snapshot();
		await port.save(value);
		expect(port.stored()).toEqual(value);
		expect(port.stored()).not.toBe(value);
	});

	it('adapts JSON documents and leaves malformed text for total model validation', async () => {
		let document: string | null = null;
		const port = createJsonSnapshotPort({
			load: async () => document,
			save: async (next) => { document = next; }
		});
		await port.save(snapshot());
		expect(typeof document).toBe('string');
		expect((await loadSnapshot(port)).kind).toBe('restored');
		document = '{';
		expect((await loadSnapshot(port)).kind).toBe('invalid');
	});

	it('debounces saves so the last request wins and flush writes immediately', async () => {
		vi.useFakeTimers();
		const port = createMemorySnapshotPort();
		const save = vi.spyOn(port, 'save');
		const saver = createDebouncedSaver(port, 100);
		saver.request(() => ({ ...snapshot(), layout: { ...createLayoutState(), statusVisible: false } }));
		saver.request(snapshot);
		expect(save).not.toHaveBeenCalled();
		await vi.advanceTimersByTimeAsync(100);
		expect(save).toHaveBeenCalledTimes(1);
		expect((port.stored() as { layout: { statusVisible: boolean } }).layout.statusVisible).toBe(true);
		saver.request(snapshot);
		await saver.flush();
		expect(save).toHaveBeenCalledTimes(2);
		saver.request(snapshot);
		saver.dispose();
		await vi.advanceTimersByTimeAsync(200);
		expect(save).toHaveBeenCalledTimes(2);
		vi.useRealTimers();
	});

	it('reports save errors instead of throwing', async () => {
		const onError = vi.fn();
		const saver = createDebouncedSaver({ load: async () => null, save: async () => { throw new Error('read-only'); } }, 0, onError);
		saver.request(snapshot);
		await saver.flush();
		expect(onError).toHaveBeenCalledTimes(1);
	});
});
