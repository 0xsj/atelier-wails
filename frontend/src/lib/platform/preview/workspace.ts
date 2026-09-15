// In-memory Workspace registry for browser previews and tests. It implements
// the registry contract behind the same JSON boundary the future desktop
// facade will use; no filesystem inspection or host APIs belong here.

import { parseWorkspaceId, type WorkspaceId } from '../../kernel';
import { readExpected } from '../codecs/workspace';
import { isCanonicalWorkspaceLocation, isValidWorkspaceName, type WorkspaceEvent, type WorkspaceOperation, type WorkspaceSnapshot, type WorkspaceStatus } from '../../services/workspace/model';
import type { WorkspaceTransport } from '../../services/workspace';

export interface PreviewWorkspaceTransport extends WorkspaceTransport {
	reset(): void;
	dump(): readonly WorkspaceSnapshot[];
}

const NOW = '2026-09-14T00:00:00.000Z';
const MAX_REVISION = (2n ** 64n) - 1n;
const SAMPLE_ID = '01900000-0000-7000-8000-000000000101';
const ARCHIVED_ID = '01900000-0000-7000-8000-000000000102';

export const PREVIEW_WORKSPACES: readonly WorkspaceSnapshot[] = [
	{ id: parseWorkspaceId(SAMPLE_ID)!, name: 'Atelier Core', location: '/workspace/atelier-core', status: 'active', revision: '1', createdAt: NOW, updatedAt: NOW },
	{ id: parseWorkspaceId(ARCHIVED_ID)!, name: 'Field Notes', location: '/workspace/field-notes', status: 'archived', revision: '2', createdAt: NOW, updatedAt: '2026-09-14T00:00:02.000Z' }
];

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function failure(kind: 'invalid' | 'conflict' | 'not_found', type: string, commit: 'not_applied' | 'none' = 'not_applied'): Record<string, unknown> {
	return { ok: false, failure: { kind, message: kind === 'not_found' ? 'workspace not found' : kind === 'conflict' ? 'workspace conflict' : 'invalid workspace request', type, fields: {}, commit } };
}

function parseRequest(document: string): Record<string, unknown> | null {
	try {
		const value: unknown = JSON.parse(document);
		return isRecord(value) ? value : null;
	} catch {
		return null;
	}
}

function readId(value: unknown): WorkspaceId | null {
	return parseWorkspaceId(value);
}

function readAt(value: unknown): string | null {
	return typeof value === 'string' && Number.isFinite(Date.parse(value)) ? value : null;
}

function sameLocation(items: Map<WorkspaceId, WorkspaceSnapshot>, location: string, except: WorkspaceId): boolean {
	return [...items.values()].some((item) => item.id !== except && item.status === 'active' && item.location === location);
}

function event(name: WorkspaceEvent['name'], workspace: WorkspaceSnapshot): Record<string, unknown> {
	return { name, workspace_id: workspace.id, revision: workspace.revision, title: workspace.name, location: workspace.location, status: workspace.status };
}

function expectedMatches(current: WorkspaceSnapshot | undefined, raw: unknown): boolean {
	const expected = readExpected(raw);
	if (!expected.ok) return false;
	return expected.value.kind === 'absent' ? current === undefined : current?.revision === expected.value.revision;
}

function nextRevision(current: string): string | null {
	const revision = BigInt(current);
	return revision >= MAX_REVISION ? null : (revision + 1n).toString(10);
}

export function createPreviewWorkspaceTransport(initial: readonly WorkspaceSnapshot[] = PREVIEW_WORKSPACES): PreviewWorkspaceTransport {
	const items = new Map<WorkspaceId, WorkspaceSnapshot>(initial.map((item) => [item.id, { ...item }]));

	return {
		async call(operation: WorkspaceOperation, document: string): Promise<unknown> {
			const request = parseRequest(document);
			if (!request) return failure('invalid', 'workspace.invalid_request');
			if (operation === 'list') {
				if (request.filter !== 'active' && request.filter !== 'all') return failure('invalid', 'workspace.invalid_filter');
				const workspaces = [...items.values()].filter((item) => request.filter === 'all' || item.status === 'active').sort((a, b) => a.id.localeCompare(b.id));
				return { ok: true, value: { workspaces: workspaces.map(toWire) } };
			}
			const id = readId(request.id);
			if (!id) return failure('invalid', 'workspace.invalid_id');

			switch (operation) {
				case 'read': {
					const workspace = items.get(id);
					return { ok: true, value: workspace ? { found: true, workspace: toWire(workspace) } : { found: false } };
				}
				case 'register': {
					const name = typeof request.name === 'string' ? request.name : '';
					const location = typeof request.location === 'string' ? request.location : '';
					const at = readAt(request.at);
					if (!isValidWorkspaceName(name) || !isCanonicalWorkspaceLocation(location) || !at) return failure('invalid', 'workspace.invalid_value');
					if (items.has(id)) return failure('conflict', 'workspace.id_taken');
					if (sameLocation(items, location, id)) return failure('conflict', 'workspace.location_taken');
					const workspace: WorkspaceSnapshot = { id, name, location, status: 'active', revision: '1', createdAt: at, updatedAt: at };
					items.set(id, workspace);
					return { ok: true, value: { workspace: toWire(workspace), event: event('workspace.registered', workspace) } };
				}
				case 'rename': {
					const current = items.get(id);
					const name = typeof request.name === 'string' ? request.name : '';
					const at = readAt(request.at);
					if (!current) return failure('not_found', 'workspace.not_found');
					if (!isValidWorkspaceName(name) || !at) return failure('invalid', 'workspace.invalid_name');
					if (!expectedMatches(current, request.expected)) return failure('conflict', 'workspace.conflict');
					if (current.name === name) return { ok: true, value: { status: 'unchanged', workspace: toWire(current), event: null } };
					const revision = nextRevision(current.revision);
					if (!revision) return failure('conflict', 'workspace.revision_exhausted');
					const next = { ...current, name, revision, updatedAt: at };
					items.set(id, next);
					return { ok: true, value: { status: 'changed', workspace: toWire(next), event: event('workspace.renamed', next) } };
				}
				case 'archive':
				case 'restore': {
					const current = items.get(id);
					const at = readAt(request.at);
					if (!current) return failure('not_found', 'workspace.not_found');
					if (!at || !expectedMatches(current, request.expected)) return failure('conflict', 'workspace.conflict');
					const status: WorkspaceStatus = operation === 'archive' ? 'archived' : 'active';
					if (status === 'active' && sameLocation(items, current.location, id)) return failure('conflict', 'workspace.location_taken');
					if (current.status === status) return { ok: true, value: { status: 'unchanged', workspace: toWire(current), event: null } };
					const revision = nextRevision(current.revision);
					if (!revision) return failure('conflict', 'workspace.revision_exhausted');
					const next = { ...current, status, revision, updatedAt: at };
					items.set(id, next);
					return { ok: true, value: { status: 'changed', workspace: toWire(next), event: event(operation === 'archive' ? 'workspace.archived' : 'workspace.restored', next) } };
				}
				case 'forget': {
					const current = items.get(id);
					if (!current) return { ok: true, value: { found: false } };
					if (current.status !== 'archived') return failure('conflict', 'workspace.active_forget');
					if (!expectedMatches(current, request.expected)) return failure('conflict', 'workspace.conflict');
					const revision = nextRevision(current.revision);
					if (!revision) return failure('conflict', 'workspace.revision_exhausted');
					items.delete(id);
					return { ok: true, value: { found: true, workspace: { revision }, event: event('workspace.forgotten', { ...current, revision }) } };
				}
			}
		},
		reset() {
			items.clear();
			for (const item of initial) items.set(item.id, { ...item });
		},
		dump() {
			return [...items.values()].sort((a, b) => a.id.localeCompare(b.id)).map((item) => ({ ...item }));
		}
	};
}

function toWire(workspace: WorkspaceSnapshot): Record<string, unknown> {
	return { id: workspace.id, name: workspace.name, location: workspace.location, status: workspace.status, revision: workspace.revision, created_at: workspace.createdAt, updated_at: workspace.updatedAt };
}
