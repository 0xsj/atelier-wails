// Runtime-open is an orchestration workflow, not a registry mutation. It
// re-reads the selected record so a stale picker cannot activate an archived
// or forgotten workspace, then returns the context the Workbench consumes.

import { err, failure, ok, type Failure, type Result } from '../../kernel';
import { readWorkspace, type WorkspaceTransport } from './workspace';
import type { WorkspaceSnapshot } from './model';

export interface WorkspaceSession {
	readonly workspace: WorkspaceSnapshot;
	readonly scope: string;
}

export interface WorkspaceOpenPort {
	open(selected: WorkspaceSnapshot): Promise<Result<WorkspaceSession, Failure>>;
}

export function createWorkspaceOpenService(transport: WorkspaceTransport): WorkspaceOpenPort {
	return {
		async open(selected) {
			const current = await readWorkspace(transport, selected.id);
			if (!current.ok) return current;
			if (!current.value.present) return err(failure('not_found', 'workspace not found', { type: 'workspace.not_found' }));
			if (current.value.value.status !== 'active') {
				return err(failure('conflict', 'workspace is archived', { type: 'workspace.archived', commit: 'not_applied' }));
			}
			return ok({ workspace: current.value.value, scope: `workspace:${current.value.value.id}` });
		}
	};
}
