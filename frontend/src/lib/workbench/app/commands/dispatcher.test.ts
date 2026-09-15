import { describe, expect, it, vi } from 'vitest';
import { CommandDispatcher } from './dispatcher';

describe('command dispatcher', () => {
	it('executes registered handlers and reports outcomes', async () => {
		const dispatcher = new CommandDispatcher();
		const handler = vi.fn();
		expect(dispatcher.register({ id: 'a', title: 'A' }, handler).kind).toBe('registered');
		expect(dispatcher.register({ id: 'a', title: 'A again' }, handler).kind).toBe('duplicate');
		expect(await dispatcher.dispatch('a', {}, { n: 1 })).toEqual({ kind: 'executed', id: 'a' });
		expect(handler).toHaveBeenCalledWith({ n: 1 });
		expect(await dispatcher.dispatch('missing', {})).toEqual({ kind: 'unknown', id: 'missing' });
	});

	it('refuses unavailable commands and captures handler failures', async () => {
		const dispatcher = new CommandDispatcher();
		dispatcher.register({ id: 'save', title: 'Save', when: (context) => context.dirty === true }, () => {
			throw new Error('disk full');
		});
		expect(await dispatcher.dispatch('save', { dirty: false })).toEqual({ kind: 'unavailable', id: 'save' });
		const failed = await dispatcher.dispatch('save', { dirty: true });
		expect(failed.kind).toBe('failed');
		if (failed.kind === 'failed') expect((failed.error as Error).message).toBe('disk full');
	});

	it('notifies on registry changes and unregisters', () => {
		const dispatcher = new CommandDispatcher();
		const listener = vi.fn();
		const stop = dispatcher.onChange(listener);
		dispatcher.register({ id: 'a', title: 'A' }, () => {});
		dispatcher.unregister('a');
		dispatcher.unregister('a');
		expect(listener).toHaveBeenCalledTimes(2);
		stop();
		dispatcher.register({ id: 'b', title: 'B' }, () => {});
		expect(listener).toHaveBeenCalledTimes(2);
		expect(Object.keys(dispatcher.registry)).toEqual(['b']);
	});
});
