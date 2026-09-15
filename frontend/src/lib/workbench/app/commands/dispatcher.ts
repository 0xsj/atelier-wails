// Command dispatch: one contract for palette, menus and shortcuts. Handlers
// are registered beside definitions; dispatch reports an outcome value.

import { registerCommand, resolveCommand, unregisterCommand, type CommandContext, type CommandDefinition, type CommandId, type CommandRegistry, type RegisterOutcome, EMPTY_REGISTRY } from '../../model/commands/commands';

export type CommandHandler = (args?: unknown) => void | Promise<void>;

export type DispatchOutcome =
	| { readonly kind: 'executed'; readonly id: CommandId }
	| { readonly kind: 'unavailable'; readonly id: CommandId }
	| { readonly kind: 'unknown'; readonly id: CommandId }
	| { readonly kind: 'no-handler'; readonly id: CommandId }
	| { readonly kind: 'failed'; readonly id: CommandId; readonly error: unknown };

export class CommandDispatcher {
	#registry: CommandRegistry = EMPTY_REGISTRY;
	readonly #handlers = new Map<CommandId, CommandHandler>();
	readonly #listeners = new Set<() => void>();

	get registry(): CommandRegistry {
		return this.#registry;
	}

	/** Notifies when the registry changes so palettes can re-list. */
	onChange(listener: () => void): () => void {
		this.#listeners.add(listener);
		return () => this.#listeners.delete(listener);
	}

	register(definition: CommandDefinition, handler: CommandHandler): RegisterOutcome {
		const outcome = registerCommand(this.#registry, definition);
		if (outcome.kind === 'registered') {
			this.#registry = outcome.registry;
			this.#handlers.set(definition.id, handler);
			this.#notify();
		}
		return outcome;
	}

	unregister(id: CommandId): void {
		if (!this.#registry[id]) return;
		this.#registry = unregisterCommand(this.#registry, id);
		this.#handlers.delete(id);
		this.#notify();
	}

	async dispatch(id: CommandId, context: CommandContext, args?: unknown): Promise<DispatchOutcome> {
		const resolution = resolveCommand(this.#registry, id, context);
		if (resolution.kind === 'unknown') return { kind: 'unknown', id };
		if (resolution.kind === 'unavailable') return { kind: 'unavailable', id };
		const handler = this.#handlers.get(id);
		if (!handler) return { kind: 'no-handler', id };
		try {
			await handler(args);
			return { kind: 'executed', id };
		} catch (error) {
			return { kind: 'failed', id, error };
		}
	}

	#notify(): void {
		for (const listener of this.#listeners) listener();
	}
}
