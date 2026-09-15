import { describe, expect, it } from 'vitest';
import { CODEC_INVALID_RESPONSE, encodeRequest, readEntry, readFailure, readListResponse, readOutcome, readReadResponse, readRemoveResponse, readReplaceResponse, readResolveResponse, readScope, readValue } from './preferences';
import { GLOBAL_SCOPE, int, text } from '../../services/preferences/model';
import { parseWorkspaceId } from '../../kernel';

const CREATE_FIXTURE = '{"ok":true,"value":{"status":"changed","entry":{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"revision":"1"},"event":{"name":"preference.changed","scope":{"kind":"global"},"key":"editor.theme","revision":"1","value":{"kind":"text","text":"dark"}}}}';
const CONFLICT_FIXTURE = '{"ok":false,"failure":{"kind":"conflict","message":"preference revision conflict","type":"preferences.conflict","fields":{},"commit":"not_applied"}}';

describe('preferences codecs', () => {
	it('encodes requests in the contract shape with ints and revisions as decimal strings', () => {
		const workspaceId = parseWorkspaceId('01900000-0000-7000-8000-000000000001')!;
		expect(encodeRequest('read', { scope: GLOBAL_SCOPE, key: 'editor.theme' })).toBe('{"scope":{"kind":"global"},"key":"editor.theme"}');
		expect(encodeRequest('list', { scope: { kind: 'workspace', workspaceId } })).toBe('{"scope":{"kind":"workspace","workspace_id":"01900000-0000-7000-8000-000000000001"}}');
		expect(encodeRequest('replace', { scope: GLOBAL_SCOPE, key: 'editor.theme', value: text('dark'), expected: { kind: 'absent' } })).toBe('{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"expected":{"kind":"absent"}}');
		expect(encodeRequest('remove', { scope: GLOBAL_SCOPE, key: 'k', expected: { kind: 'revision', revision: '3' } })).toBe('{"scope":{"kind":"global"},"key":"k","expected":{"kind":"revision","revision":"3"}}');
		expect(encodeRequest('resolve', { scope: GLOBAL_SCOPE, key: 'k', fallback: int(-9223372036854775808n) })).toBe('{"scope":{"kind":"global"},"key":"k","fallback":{"kind":"int","int":"-9223372036854775808"}}');
	});

	it('decodes the create fixture byte-identically after re-encoding its parts', () => {
		const outcome = readOutcome(JSON.parse(CREATE_FIXTURE), 'replace', readReplaceResponse);
		expect(outcome.ok).toBe(true);
		if (!outcome.ok) return;
		expect(outcome.value.status).toBe('changed');
		expect(outcome.value.entry).toEqual({ scope: { kind: 'global' }, key: 'editor.theme', value: { kind: 'text', text: 'dark' }, revision: '1' });
		expect(outcome.value.event?.name).toBe('preference.changed');
		expect(outcome.value.event?.value).toEqual({ kind: 'text', text: 'dark' });
	});

	it('decodes the conflict fixture into a kernel failure', () => {
		const outcome = readOutcome(JSON.parse(CONFLICT_FIXTURE), 'replace', readReplaceResponse);
		expect(outcome).toEqual({ ok: false, error: { kind: 'conflict', message: 'preference revision conflict', type: 'preferences.conflict', fields: {}, commit: 'not_applied' } });
	});

	it('keeps int64 extremes exact and refuses malformed numbers', () => {
		expect(readValue({ kind: 'int', int: '9223372036854775807' })).toEqual({ ok: true, value: { kind: 'int', value: 9223372036854775807n } });
		expect(readValue({ kind: 'int', int: '1.5' }).ok).toBe(false);
		expect(readValue({ kind: 'int', int: '+1' }).ok).toBe(false);
		expect(readValue({ kind: 'int', int: 42 }).ok).toBe(false);
		expect(readValue({ kind: 'text', text: '' })).toEqual({ ok: true, value: { kind: 'text', text: '' } });
		expect(readValue({ kind: 'bool', bool: 'true' })).toEqual({ ok: false, error: 'value.bool: missing' });
		expect(readValue({ kind: 'float', float: 1 })).toEqual({ ok: false, error: 'value.kind: unknown' });
		expect(readScope({ kind: 'workspace', workspace_id: 'nope' })).toEqual({ ok: false, error: 'scope.workspace_id: invalid' });
		expect(readEntry({ scope: { kind: 'global' }, key: 'k', value: { kind: 'bool', bool: true }, revision: '01' }).ok).toBe(true);
		expect(readEntry({ scope: { kind: 'global' }, key: 'k', value: { kind: 'bool', bool: true }, revision: 1 })).toEqual({ ok: false, error: 'entry.revision: invalid' });
	});

	it('decodes the remaining response shapes including absence', () => {
		expect(readReadResponse({ found: false, entry: null })).toEqual({ ok: true, value: { found: false } });
		expect(readReadResponse({ found: 'yes' }).ok).toBe(false);
		expect(readListResponse({ entries: [] })).toEqual({ ok: true, value: { entries: [] } });
		expect(readListResponse({ entries: [{ scope: { kind: 'global' } }] })).toEqual({ ok: false, error: 'value.entries[0].key: missing' });
		expect(readRemoveResponse({ removed: false, revision: null, event: null })).toEqual({ ok: true, value: { removed: false } });
		const removed = readRemoveResponse({ removed: true, revision: '2', event: { name: 'preference.removed', scope: { kind: 'global' }, key: 'k', revision: '2', value: null } });
		expect(removed.ok && removed.value.removed && removed.value.event.value).toBeNull();
		expect(readResolveResponse({ value: { kind: 'bool', bool: true }, stored: false, revision: null })).toEqual({ ok: true, value: { value: { kind: 'bool', value: true }, stored: false } });
		expect(readResolveResponse({ value: { kind: 'bool', bool: true }, stored: true, revision: '3' })).toEqual({ ok: true, value: { value: { kind: 'bool', value: true }, stored: true, revision: '3' } });
	});

	it('projects malformed failures and envelopes to the internal fallback with the operation commit', () => {
		expect(readFailure({ kind: 'weird', message: 'x', commit: 'none' }, 'read')).toEqual({ kind: 'internal', message: 'internal error', type: null, fields: {}, commit: 'none' });
		expect(readFailure({ kind: 'internal', message: 'secret cause', type: 'x', fields: { a: 'b' }, commit: 'unknown' }, 'replace')).toEqual({ kind: 'internal', message: 'internal error', type: null, fields: {}, commit: 'unknown' });
		expect(readFailure({ kind: 'invalid', message: '', type: null, fields: { 'value.int': 'invalid', bad: 1 }, commit: 'not_applied' }, 'replace')).toEqual({ kind: 'invalid', message: 'request failed', type: null, fields: { 'value.int': 'invalid' }, commit: 'not_applied' });
		expect(readOutcome('nope', 'remove', readRemoveResponse)).toMatchObject({ ok: false, error: { kind: 'internal', type: CODEC_INVALID_RESPONSE, commit: 'unknown', fields: { response: 'outcome: not an object' } } });
		expect(readOutcome({ ok: true, value: { found: true, entry: {} } }, 'read', readReadResponse)).toMatchObject({ ok: false, error: { type: CODEC_INVALID_RESPONSE, commit: 'none' } });
		expect(readOutcome({ ok: 'maybe' }, 'list', readListResponse).ok).toBe(false);
	});
});
