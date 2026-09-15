import { describe, expect, it } from 'vitest';
import { createPreviewPreferencesTransport } from './preferences';

const CREATE_DOCUMENT = '{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"expected":{"kind":"absent"}}';
const CREATE_FIXTURE = '{"ok":true,"value":{"status":"changed","entry":{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"revision":"1"},"event":{"name":"preference.changed","scope":{"kind":"global"},"key":"editor.theme","revision":"1","value":{"kind":"text","text":"dark"}}}}';
const CONFLICT_FIXTURE = '{"ok":false,"failure":{"kind":"conflict","message":"preference revision conflict","type":"preferences.conflict","fields":{},"commit":"not_applied"}}';

const doc = (value: unknown) => JSON.stringify(value);
const global = { kind: 'global' };

describe('preview preferences transport', () => {
	it('W12/W07: the create outcome matches the contract fixture byte for byte, then an equal replace is unchanged', async () => {
		const transport = createPreviewPreferencesTransport();
		expect(JSON.stringify(await transport.call('replace', CREATE_DOCUMENT))).toBe(CREATE_FIXTURE);
		const again = (await transport.call('replace', doc({ scope: global, key: 'editor.theme', value: { kind: 'text', text: 'dark' }, expected: { kind: 'revision', revision: '1' } }))) as { value: { status: string; event: unknown } };
		expect(again.value.status).toBe('unchanged');
		expect(again.value.event).toBeNull();
	});

	it('W10/W12: expected absent on a present key is the conflict fixture; malformed int is a field problem, both not applied', async () => {
		const transport = createPreviewPreferencesTransport();
		await transport.call('replace', CREATE_DOCUMENT);
		expect(JSON.stringify(await transport.call('replace', CREATE_DOCUMENT))).toBe(CONFLICT_FIXTURE);
		const invalid = (await transport.call('replace', doc({ scope: global, key: 'n', value: { kind: 'int', int: 'abc' }, expected: { kind: 'absent' } }))) as { failure: Record<string, unknown> };
		expect(invalid.failure).toEqual({ kind: 'invalid', message: 'invalid request', type: 'desktop.invalid_request', fields: { 'value.int': 'invalid' }, commit: 'not_applied' });
		expect(JSON.stringify(invalid)).not.toContain('abc');
	});

	it('W01–W04: decoding vocabulary matches the contract paths and problem words', async () => {
		const transport = createPreviewPreferencesTransport();
		const failureOf = async (operation: 'read' | 'replace', request: unknown) => ((await transport.call(operation, doc(request))) as { failure: { type: string | null; fields: Record<string, string> } }).failure;
		expect((await failureOf('read', { scope: {}, key: 'k' })).fields).toEqual({ 'scope.kind': 'missing' });
		expect((await failureOf('read', { scope: { kind: 'tenant' }, key: 'k' })).fields).toEqual({ 'scope.kind': 'unknown' });
		expect((await failureOf('read', { scope: { kind: 'workspace' }, key: 'k' })).fields).toEqual({ 'scope.workspace_id': 'missing' });
		expect((await failureOf('read', { scope: { kind: 'workspace', workspace_id: 'nope' }, key: 'k' })).fields).toEqual({ 'scope.workspace_id': 'invalid' });
		expect((await failureOf('read', { scope: global, key: 'Editor' })).type).toBe('preferences.invalid_key');
		expect((await failureOf('replace', { scope: global, key: 'k', value: { kind: 'text' }, expected: { kind: 'absent' } })).fields).toEqual({ 'value.text': 'missing' });
		expect((await failureOf('replace', { scope: global, key: 'k', value: { kind: 'text', text: 'x'.repeat(4097) }, expected: { kind: 'absent' } })).type).toBe('preferences.invalid_value');
		expect((await failureOf('replace', { scope: global, key: 'k', value: { kind: 'bool', bool: true }, expected: { kind: 'revision', revision: '0' } })).type).toBe('preferences.invalid_revision');
		expect((await failureOf('replace', { scope: global, key: 'k', value: { kind: 'bool', bool: true }, expected: { kind: 'revision', revision: 'x' } })).fields).toEqual({ 'expected.revision': 'invalid' });
		expect((await failureOf('replace', { scope: global, key: 'k', value: { kind: 'int', int: '9223372036854775808' }, expected: { kind: 'absent' } })).fields).toEqual({ 'value.int': 'invalid' });
	});

	it('W05/W06/W08/W09: read, list ordering, remove and resolve', async () => {
		const transport = createPreviewPreferencesTransport();
		await transport.call('replace', doc({ scope: global, key: 'z.last', value: { kind: 'bool', bool: true }, expected: { kind: 'absent' } }));
		await transport.call('replace', doc({ scope: global, key: 'a.first', value: { kind: 'int', int: '-42' }, expected: { kind: 'absent' } }));
		const list = (await transport.call('list', doc({ scope: global }))) as { value: { entries: { key: string }[] } };
		expect(list.value.entries.map((entry) => entry.key)).toEqual(['a.first', 'z.last']);
		expect(await transport.call('read', doc({ scope: global, key: 'missing' }))).toEqual({ ok: true, value: { found: false, entry: null } });
		const removed = (await transport.call('remove', doc({ scope: global, key: 'z.last', expected: { kind: 'revision', revision: '1' } }))) as { value: Record<string, unknown> };
		expect(removed.value).toEqual({ removed: true, revision: '2', event: { name: 'preference.removed', scope: global, key: 'z.last', revision: '2', value: null } });
		expect(await transport.call('remove', doc({ scope: global, key: 'z.last', expected: { kind: 'absent' } }))).toEqual({ ok: true, value: { removed: false, revision: null, event: null } });
		expect(await transport.call('resolve', doc({ scope: global, key: 'a.first', fallback: { kind: 'int', int: '0' } }))).toEqual({ ok: true, value: { value: { kind: 'int', int: '-42' }, stored: true, revision: '1' } });
		expect(await transport.call('resolve', doc({ scope: global, key: 'nope', fallback: { kind: 'bool', bool: true } }))).toEqual({ ok: true, value: { value: { kind: 'bool', bool: true }, stored: false, revision: null } });
	});

	it('W11: fail-before and lose-acknowledgement report commit unknown on writes and none on reads', async () => {
		const transport = createPreviewPreferencesTransport();
		transport.failNext({ kind: 'unavailable', mode: 'fail-before' });
		expect(await transport.call('replace', CREATE_DOCUMENT)).toMatchObject({ ok: false, failure: { kind: 'unavailable', commit: 'unknown' } });
		expect(await transport.call('read', doc({ scope: global, key: 'editor.theme' }))).toEqual({ ok: true, value: { found: false, entry: null } });
		transport.failNext({ kind: 'timeout', mode: 'lose-acknowledgement' });
		expect(await transport.call('replace', CREATE_DOCUMENT)).toMatchObject({ ok: false, failure: { kind: 'timeout', commit: 'unknown' } });
		expect(await transport.call('read', doc({ scope: global, key: 'editor.theme' }))).toMatchObject({ ok: true, value: { found: true, entry: { revision: '1' } } });
		expect(await transport.call('replace', CREATE_DOCUMENT)).toMatchObject({ ok: false, failure: { kind: 'conflict', commit: 'not_applied' } });
		transport.failNext({ kind: 'unavailable', mode: 'fail-before' });
		expect(await transport.call('read', doc({ scope: global, key: 'editor.theme' }))).toMatchObject({ ok: false, failure: { kind: 'unavailable', commit: 'none' } });
	});

	it('W13: documents that are not JSON objects refuse request: invalid with commit by operation', async () => {
		const transport = createPreviewPreferencesTransport();
		for (const bad of ['not json', '42', '[]', 'null', '"x"']) {
			expect(await transport.call('replace', bad)).toEqual({ ok: false, failure: { kind: 'invalid', message: 'invalid request', type: 'desktop.invalid_request', fields: { request: 'invalid' }, commit: 'not_applied' } });
			expect(await transport.call('list', bad)).toMatchObject({ ok: false, failure: { fields: { request: 'invalid' }, commit: 'none' } });
		}
	});
});
