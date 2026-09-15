// Wails adapter for the Workspace transport port. Bound method names stay at
// the platform boundary; the service validates returned plain data.

import type { WorkspaceOperation } from '../../services/workspace/model';
import type { WorkspaceTransport } from '../../services/workspace/workspace';

const METHODS: Readonly<Record<WorkspaceOperation, string>> = {
	read: 'Read',
	list: 'List',
	register: 'Register',
	rename: 'Rename',
	archive: 'Archive',
	restore: 'Restore',
	forget: 'Forget'
};

type BoundWorkspace = Readonly<Record<string, (document: string) => Promise<unknown>>>;

function boundWorkspace(): BoundWorkspace | null {
	const go = (globalThis as { go?: { wailshost?: { Workspace?: BoundWorkspace } } }).go;
	return go?.wailshost?.Workspace ?? null;
}

export function createDesktopWorkspaceTransport(): WorkspaceTransport {
	return {
		call(operation, document) {
			const bound = boundWorkspace();
			const method = bound?.[METHODS[operation]];
			if (!method) return Promise.reject(new Error('Wails workspace binding is unavailable'));
			return method(document);
		}
	};
}
