// Frontend values for the Workspace registry contract. Workspaces are
// metadata records; files, documents and filesystem inspection stay outside
// this slice. Revisions remain decimal strings at the transport boundary.

import type { WorkspaceId } from '../../kernel';

export type WorkspaceStatus = 'active' | 'archived';
export type WorkspaceFilter = 'active' | 'all';
export type WorkspaceRevision = string;

export type WorkspaceExpected =
	| { readonly kind: 'absent' }
	| { readonly kind: 'revision'; readonly revision: WorkspaceRevision };

export interface WorkspaceSnapshot {
	readonly id: WorkspaceId;
	readonly name: string;
	readonly location: string;
	readonly status: WorkspaceStatus;
	readonly revision: WorkspaceRevision;
	readonly createdAt: string;
	readonly updatedAt: string;
}

export type WorkspaceEventName =
	| 'workspace.registered'
	| 'workspace.renamed'
	| 'workspace.archived'
	| 'workspace.restored'
	| 'workspace.forgotten';

export interface WorkspaceEvent {
	readonly name: WorkspaceEventName;
	readonly workspaceId: WorkspaceId;
	readonly revision: WorkspaceRevision;
	readonly title: string;
	readonly location: string;
	readonly status: WorkspaceStatus;
}

export interface WorkspaceRegistered {
	readonly workspace: WorkspaceSnapshot;
	readonly event: WorkspaceEvent;
}

export type WorkspaceMutation = {
	readonly status: 'changed' | 'unchanged';
	readonly workspace: WorkspaceSnapshot;
	readonly event: WorkspaceEvent | null;
};

export interface WorkspaceForgotten {
	readonly revision: WorkspaceRevision;
	readonly event: WorkspaceEvent;
}

export type WorkspaceOperation = 'read' | 'list' | 'register' | 'rename' | 'archive' | 'restore' | 'forget';

export const WORKSPACE_WRITE_OPERATIONS: ReadonlySet<WorkspaceOperation> = new Set([
	'register',
	'rename',
	'archive',
	'restore',
	'forget'
]);

export const WORKSPACE_NAME_MAX_LENGTH = 120;

/** The registry accepts Unicode scalar names without controls or edge whitespace. */
export function isValidWorkspaceName(name: string): boolean {
	const scalarLength = Array.from(name).length;
	return scalarLength > 0 && scalarLength <= WORKSPACE_NAME_MAX_LENGTH && name === name.trim() && !/[\u0000-\u001f\u007f]/u.test(name);
}

/** Locations are already canonical by the time they enter the registry. */
export function isCanonicalWorkspaceLocation(location: string): boolean {
	return location.length > 0 && !location.includes('\u0000') && (location.startsWith('/') || /^[A-Za-z]:[\\/]/u.test(location) || location.startsWith('\\\\'));
}

export function isValidWorkspaceRevision(revision: string): boolean {
	return /^[1-9][0-9]*$/u.test(revision);
}
