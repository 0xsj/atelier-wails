import { describe, expect, it } from 'vitest';
import { createPreviewPreferencesTransport } from '../../platform/preview/preferences';
import { GLOBAL_SCOPE, bool, int, text } from './model';
import { clearPreference, listPreferences, readPreference, replacePreference, resolvePreference, setPreference, type PreferencesTransport } from './preferences';

describe('preferences service', () => {
	it('reads absence and presence as distinct successes', async () => {
		const transport = createPreviewPreferencesTransport();
		expect(await readPreference(transport, GLOBAL_SCOPE, 'ui.theme')).toEqual({ ok: true, value: { present: false } });
		await replacePreference(transport, GLOBAL_SCOPE, 'ui.theme', text('dark'), { kind: 'absent' });
		const found = await readPreference(transport, GLOBAL_SCOPE, 'ui.theme');
		expect(found.ok && found.value.present && found.value.value).toEqual({ scope: { kind: 'global' }, key: 'ui.theme', value: { kind: 'text', text: 'dark' }, revision: '1' });
		expect(await resolvePreference(transport, GLOBAL_SCOPE, 'missing', bool(true))).toEqual({ ok: true, value: { value: { kind: 'bool', value: true }, stored: false } });
	});

	it('maps a rejected transport promise per the contract', async () => {
		const rejecting: PreferencesTransport = { call: async () => { throw new Error('bridge down'); } };
		expect(await readPreference(rejecting, GLOBAL_SCOPE, 'k')).toEqual({ ok: false, error: { kind: 'internal', message: 'internal error', type: null, fields: {}, commit: 'none' } });
		expect(await replacePreference(rejecting, GLOBAL_SCOPE, 'k', bool(true), { kind: 'absent' })).toMatchObject({ ok: false, error: { kind: 'internal', commit: 'unknown' } });
	});

	it('setPreference creates, updates against the observed revision, and reports unchanged', async () => {
		const transport = createPreviewPreferencesTransport();
		expect(await setPreference(transport, GLOBAL_SCOPE, 'editor.tab', int(4n))).toMatchObject({ ok: true, value: { status: 'changed', entry: { revision: '1' }, attempts: 1 } });
		expect(await setPreference(transport, GLOBAL_SCOPE, 'editor.tab', int(2n))).toMatchObject({ ok: true, value: { status: 'changed', entry: { revision: '2' } } });
		expect(await setPreference(transport, GLOBAL_SCOPE, 'editor.tab', int(2n))).toMatchObject({ ok: true, value: { status: 'unchanged', entry: { revision: '2' } } });
		const list = await listPreferences(transport, GLOBAL_SCOPE);
		expect(list.ok && list.value.map((entry) => entry.key)).toEqual(['editor.tab']);
	});

	it('reconciles an uncertain commit by re-reading instead of retrying blindly', async () => {
		const transport = createPreviewPreferencesTransport();
		transport.failNext({ kind: 'timeout', mode: 'lose-acknowledgement', operation: 'replace' });
		const outcome = await setPreference(transport, GLOBAL_SCOPE, 'ui.theme', text('dark'));
		expect(outcome).toMatchObject({ ok: true, value: { status: 'reconciled', entry: { revision: '1' }, attempts: 1 } });
		transport.failNext({ kind: 'unavailable', mode: 'fail-before', operation: 'replace' });
		const retried = await setPreference(transport, GLOBAL_SCOPE, 'ui.theme', text('light'));
		expect(retried).toMatchObject({ ok: true, value: { status: 'changed', entry: { revision: '2' }, attempts: 2 } });
	});

	it('gives up after the attempt budget and surfaces the last failure', async () => {
		const transport = createPreviewPreferencesTransport();
		const flaky: PreferencesTransport = {
			call(operation, document) {
				if (operation === 'replace') transport.failNext({ kind: 'unavailable', mode: 'fail-before' });
				return transport.call(operation, document);
			}
		};
		const outcome = await setPreference(flaky, GLOBAL_SCOPE, 'k', bool(true), { attempts: 2 });
		expect(outcome).toMatchObject({ ok: false, error: { kind: 'unavailable', commit: 'unknown' } });
	});

	it('clearPreference removes with the observed revision and treats absence as a no-op', async () => {
		const transport = createPreviewPreferencesTransport();
		expect(await clearPreference(transport, GLOBAL_SCOPE, 'k')).toEqual({ ok: true, value: { removed: false, revision: null, reconciled: false, attempts: 1 } });
		await setPreference(transport, GLOBAL_SCOPE, 'k', bool(true));
		expect(await clearPreference(transport, GLOBAL_SCOPE, 'k')).toEqual({ ok: true, value: { removed: true, revision: '2', reconciled: false, attempts: 1 } });
		await setPreference(transport, GLOBAL_SCOPE, 'k', bool(false));
		transport.failNext({ kind: 'timeout', mode: 'lose-acknowledgement', operation: 'remove' });
		expect(await clearPreference(transport, GLOBAL_SCOPE, 'k')).toMatchObject({ ok: true, value: { removed: true, reconciled: true } });
	});
});
