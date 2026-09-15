// The workbench's own commands. Hosts register these once and add their own.

import type { WorkbenchStore } from '../workbench-store';

export const WORKBENCH_COMMANDS = {
	toggleSidebar: 'workbench.view.toggleSidebar',
	togglePanel: 'workbench.view.togglePanel',
	closeEditor: 'workbench.editor.close',
	closeOthers: 'workbench.editor.closeOthers',
	pinEditor: 'workbench.editor.togglePin',
	splitEditor: 'workbench.editor.splitRight',
	back: 'workbench.navigate.back',
	forward: 'workbench.navigate.forward'
} as const;

export interface WorkbenchCommandHooks {
	/** Called when a close needs confirmation; the host shows its dialog. */
	onCloseNeedsConfirmation?: (viewId: string) => void;
}

export function registerWorkbenchCommands(store: WorkbenchStore, hooks: WorkbenchCommandHooks = {}): void {
	const activeId = () => {
		const views = store.state.views;
		return views.groups.find((group) => group.id === views.activeGroupId)?.activeViewId ?? null;
	};
	store.registerCommand({ id: WORKBENCH_COMMANDS.toggleSidebar, title: 'Toggle Sidebar', category: 'View', keybinding: 'mod+b' }, () => store.toggleSidebar());
	store.registerCommand({ id: WORKBENCH_COMMANDS.togglePanel, title: 'Toggle Panel', category: 'View', keybinding: 'mod+j' }, () => store.togglePanel());
	store.registerCommand({ id: WORKBENCH_COMMANDS.closeEditor, title: 'Close Editor', category: 'Editor', keybinding: 'mod+w', when: (context) => context.hasActiveView === true }, () => {
		const id = activeId();
		if (!id) return;
		const request = store.requestClose(id);
		if (request.kind === 'needs-confirmation') hooks.onCloseNeedsConfirmation?.(id);
	});
	store.registerCommand({ id: WORKBENCH_COMMANDS.closeOthers, title: 'Close Other Editors', category: 'Editor', when: (context) => typeof context.openViewCount === 'number' && context.openViewCount > 1 }, () => {
		const id = activeId();
		if (id) store.closeOthers(id);
	});
	store.registerCommand({ id: WORKBENCH_COMMANDS.pinEditor, title: 'Pin or Unpin Editor', category: 'Editor', keybinding: 'mod+k', when: (context) => context.hasActiveView === true }, () => {
		const id = activeId();
		if (id) store.pinView(id, !store.state.views.views[id]?.pinned);
	});
	store.registerCommand({ id: WORKBENCH_COMMANDS.splitEditor, title: 'Split Editor Right', category: 'Editor', keybinding: 'mod+\\' }, () => {
		store.splitGroup();
	});
	store.registerCommand({ id: WORKBENCH_COMMANDS.back, title: 'Go Back', category: 'Navigate', keybinding: 'alt+arrowleft', when: (context) => context.canGoBack === true }, () => store.goBack());
	store.registerCommand({ id: WORKBENCH_COMMANDS.forward, title: 'Go Forward', category: 'Navigate', keybinding: 'alt+arrowright', when: (context) => context.canGoForward === true }, () => store.goForward());
}
