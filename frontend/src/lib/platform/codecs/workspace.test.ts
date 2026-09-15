import { describe, expect, it } from 'vitest';
import { encodeRequest, readListResponse, readMutationResponse, readOutcome, readReadResponse, readWorkspace } from './workspace';
import { parseWorkspaceId } from '../../kernel';

const ID = parseWorkspaceId('01900000-0000-7000-8000-000000000101')!;
const WIRE = { id: ID, name: 'Atelier Core', location: '/workspace/atelier-core', status: 'active' as const, revision: '1', created_at: '2026-09-14T00:00:00.000Z', updated_at: '2026-09-14T00:00:00.000Z' };

describe('workspace codecs', () => {
	it('encodes registry requests with explicit expected revisions', () => {
		expect(encodeRequest('read', { id: ID })).toBe('{"id":"01900000-0000-7000-8000-000000000101"}');
		expect(encodeRequest('list', { filter: 'all' })).toBe('{"filter":"all"}');
		expect(encodeRequest('register', { id: ID, name: 'Atelier Core', location: '/workspace/atelier-core', at: WIRE.created_at })).toContain('"name":"Atelier Core"');
		expect(encodeRequest('rename', { id: ID, name: 'Studio', expected: { kind: 'revision', revision: '3' }, at: WIRE.updated_at })).toContain('"expected":{"kind":"revision","revision":"3"}');
	});

	it('reads workspace snapshots and preserves all fields', () => {
		expect(readWorkspace(WIRE)).toEqual({ ok: true, value: { id: ID, name: 'Atelier Core', location: '/workspace/atelier-core', status: 'active', revision: '1', createdAt: WIRE.created_at, updatedAt: WIRE.updated_at } });
		expect(readWorkspace({ ...WIRE, id: '00000000-0000-0000-0000-000000000000' }).ok).toBe(false);
		expect(readWorkspace({ ...WIRE, location: 'relative/project' })).toEqual({ ok: false, error: 'workspace.location: invalid' });
	});

	it('keeps read absence and empty lists successful', () => {
		expect(readReadResponse({ found: false })).toEqual({ ok: true, value: { found: false } });
		expect(readListResponse({ workspaces: [] })).toEqual({ ok: true, value: { workspaces: [] } });
		expect(readListResponse({ workspaces: [WIRE] })).toMatchObject({ ok: true, value: { workspaces: [{ id: ID, createdAt: WIRE.created_at, updatedAt: WIRE.updated_at }] } });
	});

	it('reads changed and unchanged lifecycle responses', () => {
		const changed = readMutationResponse({ status: 'changed', workspace: WIRE, event: { name: 'workspace.renamed', workspace_id: ID, revision: '2', title: 'Studio', location: WIRE.location, status: 'active' } });
		expect(changed).toMatchObject({ ok: true, value: { status: 'changed', workspace: { revision: '1' }, event: { name: 'workspace.renamed', workspaceId: ID, title: 'Studio' } } });
		expect(readMutationResponse({ status: 'unchanged', workspace: WIRE, event: null })).toMatchObject({ ok: true, value: { status: 'unchanged', event: null } });
		expect(readOutcome({ ok: true, value: { found: false } }, 'read', readReadResponse)).toEqual({ ok: true, value: { found: false } });
		expect(readOutcome({ ok: false, failure: { kind: 'unexpected', message: 'leak', commit: 'applied' } }, 'rename', readMutationResponse)).toEqual({ ok: false, error: { kind: 'internal', message: 'internal error', type: null, fields: {}, commit: 'unknown' } });
	});
});
