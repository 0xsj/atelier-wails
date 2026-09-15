// The portable workbench store: binds the pure models together, owns the
// close policy for dirty views, dispatches commands and shortcuts, and produces
// snapshots for restoration. It implements the store contract (subscribe with an
// immediate call, returning an unsubscribe) without importing Svelte.

import * as layoutModel from '../model/layout/layout';
import { createSnapshot, type WorkbenchSnapshot } from '../model/layout/snapshot';
import * as viewsModel from '../model/views/views';
import type { CommandContext, CommandDefinition, CommandId } from '../model/commands/commands';
import type { Platform } from '../model/commands/keybinding';
import { CommandDispatcher, type CommandHandler, type DispatchOutcome } from './commands/dispatcher';
import { resolveShortcut, type ShortcutEventLike, type ShortcutResolution } from './keybindings/keybindings';
import type { LoadOutcome } from './restoration/restoration';
import { EMPTY_HISTORY, canGoBack, canGoForward, currentEntry, forgetView, goBack, goForward, recordVisit, type NavigationHistory } from './navigation/history';

export interface WorkbenchState {
	readonly layout: layoutModel.LayoutState;
	readonly views: viewsModel.ViewsState;
	readonly history: NavigationHistory;
}

export interface WorkbenchStoreOptions {
	readonly platform: Platform;
	readonly layout?: layoutModel.LayoutState;
	readonly views?: viewsModel.ViewsState;
	readonly dispatcher?: CommandDispatcher;
	readonly bounds?: layoutModel.Bounds | null;
}

export type CloseRequest =
	| { readonly kind: 'closed'; readonly view: viewsModel.ViewInstance }
	| { readonly kind: 'needs-confirmation'; readonly view: viewsModel.ViewInstance }
	| { readonly kind: 'refused'; readonly viewId: viewsModel.ViewId };

export type KeyOutcome =
	| { readonly kind: 'dispatched'; readonly outcome: DispatchOutcome }
	| { readonly kind: 'ignored'; readonly resolution: ShortcutResolution };

export class WorkbenchStore {
	readonly platform: Platform;
	readonly dispatcher: CommandDispatcher;
	#state: WorkbenchState;
	#bounds: layoutModel.Bounds | null;
	readonly #subscribers = new Set<(state: WorkbenchState) => void>();

	constructor(options: WorkbenchStoreOptions) {
		this.platform = options.platform;
		this.dispatcher = options.dispatcher ?? new CommandDispatcher();
		this.#bounds = options.bounds ?? null;
		const views = options.views ?? viewsModel.createViewsState();
		const layout = layoutModel.syncEditorSizes(options.layout ?? layoutModel.createLayoutState(), views.groups.length);
		this.#state = { layout, views, history: EMPTY_HISTORY };
	}

	get state(): WorkbenchState {
		return this.#state;
	}

	subscribe(run: (state: WorkbenchState) => void): () => void {
		this.#subscribers.add(run);
		run(this.#state);
		return () => this.#subscribers.delete(run);
	}

	#set(next: WorkbenchState): void {
		if (next === this.#state) return;
		this.#state = next;
		for (const run of this.#subscribers) run(next);
	}

	#setViews(views: viewsModel.ViewsState, history = this.#state.history): void {
		if (views === this.#state.views && history === this.#state.history) return;
		const layout = layoutModel.syncEditorSizes(this.#state.layout, views.groups.length);
		this.#set({ layout, views, history });
	}

	#setLayout(layout: layoutModel.LayoutState): void {
		if (layout === this.#state.layout) return;
		this.#set({ ...this.#state, layout });
	}

	#visit(views: viewsModel.ViewsState): NavigationHistory {
		const active = viewsModel.activeView(views);
		return active ? recordVisit(this.#state.history, active.id) : this.#state.history;
	}

	// Views

	openView(descriptor: viewsModel.ViewDescriptor, options?: viewsModel.OpenViewOptions): void {
		const views = viewsModel.openView(this.#state.views, descriptor, options);
		this.#setViews(views, this.#visit(views));
	}

	activateView(viewId: viewsModel.ViewId): void {
		const views = viewsModel.activateView(this.#state.views, viewId);
		this.#setViews(views, this.#visit(views));
	}

	setActiveGroup(groupId: viewsModel.GroupId): void {
		this.#setViews(viewsModel.setActiveGroup(this.#state.views, groupId));
	}

	/** Dirty views are not closed silently; the host confirms and calls closeView with force. */
	requestClose(viewId: viewsModel.ViewId): CloseRequest {
		const view = this.#state.views.views[viewId];
		if (!view || view.closable === false) return { kind: 'refused', viewId };
		if (view.dirty) return { kind: 'needs-confirmation', view };
		this.closeView(viewId, { force: true });
		return { kind: 'closed', view };
	}

	closeView(viewId: viewsModel.ViewId, options: { force?: boolean } = {}): void {
		const view = this.#state.views.views[viewId];
		if (!view || (view.dirty && !options.force)) return;
		const views = viewsModel.closeView(this.#state.views, viewId);
		if (views === this.#state.views) return;
		this.#setViews(views, forgetView(this.#state.history, viewId));
	}

	closeOthers(viewId: viewsModel.ViewId): readonly viewsModel.ViewInstance[] {
		return this.#closeMany(viewsModel.groupOf(this.#state.views, viewId)?.viewIds.filter((id) => id !== viewId) ?? []);
	}

	closeToRight(viewId: viewsModel.ViewId): readonly viewsModel.ViewInstance[] {
		const group = viewsModel.groupOf(this.#state.views, viewId);
		if (!group) return [];
		return this.#closeMany(group.viewIds.slice(group.viewIds.indexOf(viewId) + 1));
	}

	closeGroupViews(groupId: viewsModel.GroupId): readonly viewsModel.ViewInstance[] {
		return this.#closeMany(this.#state.views.groups.find((group) => group.id === groupId)?.viewIds ?? []);
	}

	/** Closes what can be closed silently and returns the dirty views left for the host to confirm. */
	#closeMany(viewIds: readonly viewsModel.ViewId[]): readonly viewsModel.ViewInstance[] {
		const skipped: viewsModel.ViewInstance[] = [];
		for (const id of viewIds) {
			const view = this.#state.views.views[id];
			if (!view || view.pinned || view.closable === false) continue;
			if (view.dirty) skipped.push(view);
			else this.closeView(id, { force: true });
		}
		return skipped;
	}

	pinView(viewId: viewsModel.ViewId, pinned: boolean): void {
		this.#setViews(viewsModel.pinView(this.#state.views, viewId, pinned));
	}

	setDirty(viewId: viewsModel.ViewId, dirty: boolean): void {
		this.#setViews(viewsModel.setDirty(this.#state.views, viewId, dirty));
	}

	reorderView(groupId: viewsModel.GroupId, from: number, to: number): void {
		this.#setViews(viewsModel.reorderView(this.#state.views, groupId, from, to));
	}

	moveView(viewId: viewsModel.ViewId, groupId: viewsModel.GroupId, index?: number): void {
		const views = viewsModel.moveView(this.#state.views, viewId, groupId, index);
		this.#setViews(views, this.#visit(views));
	}

	splitGroup(afterGroupId?: viewsModel.GroupId): viewsModel.GroupId {
		const id = viewsModel.nextGroupId(this.#state.views);
		this.#setViews(viewsModel.addGroup(this.#state.views, { id, afterGroupId }));
		return id;
	}

	removeGroup(groupId: viewsModel.GroupId): void {
		this.#setViews(viewsModel.removeGroup(this.#state.views, groupId));
	}

	// Layout

	setBounds(bounds: layoutModel.Bounds | null): void {
		this.#bounds = bounds;
		this.#setLayout(layoutModel.fitToBounds(this.#state.layout, bounds));
	}

	toggleSidebar(visible?: boolean): void {
		this.#setLayout(layoutModel.toggleSidebar(this.#state.layout, visible));
	}

	setSidebarWidth(width: number): void {
		this.#setLayout(layoutModel.setSidebarWidth(this.#state.layout, width, this.#bounds));
	}

	selectActivity(activityId: string): void {
		this.#setLayout(layoutModel.selectActivity(this.#state.layout, activityId));
	}

	togglePane(paneId: string): void {
		this.#setLayout(layoutModel.togglePane(this.#state.layout, paneId));
	}

	setPaneCollapsed(paneIds: readonly string[]): void {
		let layout = this.#state.layout;
		const all = new Set([...layout.sidebar.collapsedPanes, ...paneIds]);
		for (const id of all) layout = layoutModel.setPaneCollapsed(layout, id, paneIds.includes(id));
		this.#setLayout(layout);
	}

	setPaneWeights(weights: Readonly<Record<string, number>>): void {
		this.#setLayout(layoutModel.setPaneWeights(this.#state.layout, weights));
	}

	setEditorSizes(sizes: readonly number[]): void {
		this.#setLayout(layoutModel.setEditorSizes(this.#state.layout, sizes));
	}

	togglePanel(visible?: boolean): void {
		this.#setLayout(layoutModel.togglePanel(this.#state.layout, visible));
	}

	setPanelHeight(height: number): void {
		this.#setLayout(layoutModel.setPanelHeight(this.#state.layout, height, this.#bounds));
	}

	// Navigation

	goBack(): void {
		this.#navigate(goBack(this.#state.history));
	}

	goForward(): void {
		this.#navigate(goForward(this.#state.history));
	}

	#navigate(history: NavigationHistory): void {
		if (history === this.#state.history) return;
		const target = currentEntry(history);
		const views = target ? viewsModel.activateView(this.#state.views, target) : this.#state.views;
		this.#setViews(views, history);
	}

	// Commands and shortcuts

	context(): CommandContext {
		const { layout, views, history } = this.#state;
		const active = viewsModel.activeView(views);
		return {
			hasActiveView: active !== null,
			activeViewDirty: active?.dirty ?? false,
			activeViewPinned: active?.pinned ?? false,
			activeViewKind: active?.kind ?? null,
			openViewCount: Object.keys(views.views).length,
			groupCount: views.groups.length,
			sidebarVisible: layout.sidebar.visible,
			panelVisible: layout.panel.visible,
			canGoBack: canGoBack(history),
			canGoForward: canGoForward(history)
		};
	}

	registerCommand(definition: CommandDefinition, handler: CommandHandler): void {
		this.dispatcher.register(definition, handler);
	}

	dispatch(id: CommandId, args?: unknown): Promise<DispatchOutcome> {
		return this.dispatcher.dispatch(id, this.context(), args);
	}

	/** Resolves a key event against the registry; the host prevents default when this returns dispatched. */
	async handleKey(event: ShortcutEventLike): Promise<KeyOutcome> {
		const resolution = resolveShortcut(this.dispatcher.registry, event, this.platform, this.context());
		if (resolution.kind !== 'match') return { kind: 'ignored', resolution };
		return { kind: 'dispatched', outcome: await this.dispatch(resolution.command.id) };
	}

	// Snapshot

	snapshot(): WorkbenchSnapshot {
		return createSnapshot(this.#state.layout, this.#state.views);
	}

	/** Applies a restored snapshot; every other load outcome leaves the store untouched. */
	restore(outcome: LoadOutcome): boolean {
		if (outcome.kind !== 'restored') return false;
		const layout = layoutModel.syncEditorSizes(layoutModel.fitToBounds(outcome.layout, this.#bounds), outcome.views.groups.length);
		const active = viewsModel.activeView(outcome.views);
		this.#set({ layout, views: outcome.views, history: active ? recordVisit(EMPTY_HISTORY, active.id) : EMPTY_HISTORY });
		return true;
	}
}

export function createWorkbenchStore(options: WorkbenchStoreOptions): WorkbenchStore {
	return new WorkbenchStore(options);
}
