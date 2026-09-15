import { describe, expect, it } from 'vitest';
import { parseWorkspaceId } from '../../kernel';
import { createPreviewWorkspaceTransport } from '../../platform/preview/workspace';
import { createWorkspaceOpenService, type WorkspaceOpenPort } from './runtime';
import type { WorkspaceSnapshot } from './model';

const ID = parseWorkspaceId('01900000-0000-7000-8000-000000000101')!;
const ARCHIVED_ID = parseWorkspaceId('01900000-0000-7000-8000-000000000102')!;
const SELECTED: WorkspaceSnapshot = {
	id: ID,
	name: 'Stale name',
	location: '/workspace/stale',
	status: 'active',
	revision: '1',
	createdAt: '2026-09-14T00:00:00.000Z',
	updatedAt: '2026-09-14T00:00:00.000Z'
};

describe('workspace runtime-open service', () => {
	it('re-reads an active record and returns the refreshed Workbench session', async () => {
		const open: WorkspaceOpenPort = createWorkspaceOpenService(createPreviewWorkspaceTransport());
		await expect(open.open(SELECTED)).resolves.toEqual({
			ok: true,
			value: {
				workspace: expect.objectContaining({ name: 'Atelier Core', status: 'active', revision: '1' }),
				scope: `workspace:${ID}`
			}
		});
	});

	it('refuses archived records and preserves the explicit failure', async () => {
		const transport = createPreviewWorkspaceTransport();
		await expect(createWorkspaceOpenService(transport).open({ ...SELECTED, id: ARCHIVED_ID, status: 'archived' })).resolves.toMatchObject({
			ok: false,
			error: { kind: 'conflict', type: 'workspace.archived', commit: 'not_applied' }
		});
	});

	it('maps a selected record that disappeared before activation to not_found', async () => {
		const transport = createPreviewWorkspaceTransport([]);
		await expect(createWorkspaceOpenService(transport).open(SELECTED)).resolves.toMatchObject({
			ok: false,
			error: { kind: 'not_found', type: 'workspace.not_found' }
		});
	});
});
