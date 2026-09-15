import { describe, expect, it } from 'vitest';
import { parseWorkspaceId } from '../../kernel';
import { readWorkspace, listWorkspaces, registerWorkspace, renameWorkspace, archiveWorkspace, restoreWorkspace, forgetWorkspace } from '../../services/workspace';
import { createPreviewWorkspaceTransport } from './workspace';

const ID = parseWorkspaceId('01900000-0000-7000-8000-000000000201')!;
const AT = '2026-09-14T00:00:10.000Z';

describe('preview workspace registry', () => {
	it('distinguishes absent reads from an empty active list', async () => {
		const transport = createPreviewWorkspaceTransport([]);
		expect(await readWorkspace(transport, ID)).toEqual({ ok: true, value: { present: false } });
		expect(await listWorkspaces(transport, 'active')).toEqual({ ok: true, value: [] });
	});

	it('runs register, rename, archive, restore and forget with revisions', async () => {
		const transport = createPreviewWorkspaceTransport([]);
		expect((await registerWorkspace(transport, ID, 'Studio', '/workspace/studio', AT)).ok).toBe(true);
		const registered = await readWorkspace(transport, ID);
		expect(registered.ok && registered.value.present && registered.value.value.revision).toBe('1');
		expect(await renameWorkspace(transport, ID, 'Studio North', { kind: 'revision', revision: '1' }, AT)).toMatchObject({ ok: true, value: { status: 'changed', workspace: { revision: '2', name: 'Studio North' } } });
		expect(await archiveWorkspace(transport, ID, { kind: 'revision', revision: '2' }, AT)).toMatchObject({ ok: true, value: { status: 'changed', workspace: { revision: '3', status: 'archived' } } });
		expect(await listWorkspaces(transport, 'active')).toEqual({ ok: true, value: [] });
		expect(await restoreWorkspace(transport, ID, { kind: 'revision', revision: '3' }, AT)).toMatchObject({ ok: true, value: { status: 'changed', workspace: { revision: '4', status: 'active' } } });
		expect(await forgetWorkspace(transport, ID, { kind: 'revision', revision: '4' })).toMatchObject({ ok: false, error: { kind: 'conflict', type: 'workspace.active_forget' } });
		expect(await archiveWorkspace(transport, ID, { kind: 'revision', revision: '4' }, AT)).toMatchObject({ ok: true, value: { workspace: { revision: '5' } } });
		expect(await forgetWorkspace(transport, ID, { kind: 'revision', revision: '5' })).toMatchObject({ ok: true, value: { revision: '6', event: { name: 'workspace.forgotten' } } });
		expect(await readWorkspace(transport, ID)).toEqual({ ok: true, value: { present: false } });
	});

	it('enforces active location uniqueness and conditional revisions', async () => {
		const transport = createPreviewWorkspaceTransport([]);
		const second = parseWorkspaceId('01900000-0000-7000-8000-000000000202')!;
		await registerWorkspace(transport, ID, 'One', '/workspace/shared', AT);
		expect(await registerWorkspace(transport, second, 'Two', '/workspace/shared', AT)).toMatchObject({ ok: false, error: { kind: 'conflict', type: 'workspace.location_taken' } });
		expect(await renameWorkspace(transport, ID, 'Renamed', { kind: 'revision', revision: '9' }, AT)).toMatchObject({ ok: false, error: { kind: 'conflict', type: 'workspace.conflict' } });
	});

	it('refuses mutations when the positive revision is exhausted', async () => {
		const max = '18446744073709551615';
		const transport = createPreviewWorkspaceTransport([{ id: ID, name: 'Maxed', location: '/workspace/maxed', status: 'archived', revision: max, createdAt: AT, updatedAt: AT }]);
		expect(await renameWorkspace(transport, ID, 'Maxed renamed', { kind: 'revision', revision: max }, AT)).toMatchObject({ ok: false, error: { kind: 'conflict', type: 'workspace.revision_exhausted' } });
		expect(await restoreWorkspace(transport, ID, { kind: 'revision', revision: max }, AT)).toMatchObject({ ok: false, error: { kind: 'conflict', type: 'workspace.revision_exhausted' } });
		expect(await forgetWorkspace(transport, ID, { kind: 'revision', revision: max })).toMatchObject({ ok: false, error: { kind: 'conflict', type: 'workspace.revision_exhausted' } });
	});
});
