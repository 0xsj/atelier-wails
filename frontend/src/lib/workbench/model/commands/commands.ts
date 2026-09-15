// Pure command registry: ids, titles, availability and keybinding lookup.
// UI actions dispatched here are distinct from native application commands.

import { formatKeybinding, matchesKeybinding, parseKeybinding, type KeyEventLike, type Keybinding, type Platform } from './keybinding';

export type CommandId = string;

export interface CommandContext {
	readonly [key: string]: boolean | string | number | null | undefined;
}

export interface CommandDefinition {
	readonly id: CommandId;
	readonly title: string;
	readonly category?: string;
	readonly keybinding?: string;
	readonly when?: (context: CommandContext) => boolean;
}

export interface RegisteredCommand extends CommandDefinition {
	readonly binding: Keybinding | null;
}

export type CommandRegistry = Readonly<Record<CommandId, RegisteredCommand>>;

export type RegisterOutcome =
	| { readonly kind: 'registered'; readonly registry: CommandRegistry }
	| { readonly kind: 'duplicate'; readonly id: CommandId; readonly registry: CommandRegistry }
	| { readonly kind: 'invalid-keybinding'; readonly id: CommandId; readonly keybinding: string; readonly registry: CommandRegistry };

export type CommandResolution =
	| { readonly kind: 'available'; readonly command: RegisteredCommand }
	| { readonly kind: 'unavailable'; readonly command: RegisteredCommand }
	| { readonly kind: 'unknown'; readonly id: CommandId };

export interface CommandListing {
	readonly command: RegisteredCommand;
	readonly available: boolean;
}

export interface PaletteEntry {
	readonly id: CommandId;
	readonly label: string;
	readonly group?: string;
	readonly shortcut?: string;
	readonly disabled: boolean;
}

export const EMPTY_REGISTRY: CommandRegistry = {};

export function registerCommand(registry: CommandRegistry, definition: CommandDefinition): RegisterOutcome {
	if (registry[definition.id]) return { kind: 'duplicate', id: definition.id, registry };
	let binding: Keybinding | null = null;
	if (definition.keybinding !== undefined) {
		binding = parseKeybinding(definition.keybinding);
		if (!binding) return { kind: 'invalid-keybinding', id: definition.id, keybinding: definition.keybinding, registry };
	}
	return { kind: 'registered', registry: { ...registry, [definition.id]: { ...definition, binding } } };
}

export function registerCommands(registry: CommandRegistry, definitions: readonly CommandDefinition[]): { registry: CommandRegistry; rejected: readonly Exclude<RegisterOutcome, { kind: 'registered' }>[] } {
	const rejected: Exclude<RegisterOutcome, { kind: 'registered' }>[] = [];
	let next = registry;
	for (const definition of definitions) {
		const outcome = registerCommand(next, definition);
		if (outcome.kind === 'registered') next = outcome.registry;
		else rejected.push(outcome);
	}
	return { registry: next, rejected };
}

export function unregisterCommand(registry: CommandRegistry, id: CommandId): CommandRegistry {
	if (!registry[id]) return registry;
	const { [id]: _removed, ...rest } = registry;
	return rest;
}

export function isAvailable(command: RegisteredCommand, context: CommandContext): boolean {
	return command.when ? command.when(context) : true;
}

export function resolveCommand(registry: CommandRegistry, id: CommandId, context: CommandContext): CommandResolution {
	const command = registry[id];
	if (!command) return { kind: 'unknown', id };
	return isAvailable(command, context) ? { kind: 'available', command } : { kind: 'unavailable', command };
}

function compareCommands(a: RegisteredCommand, b: RegisteredCommand): number {
	const category = (a.category ?? '').localeCompare(b.category ?? '');
	return category !== 0 ? category : a.title.localeCompare(b.title);
}

export function listCommands(registry: CommandRegistry, context: CommandContext): readonly CommandListing[] {
	return Object.values(registry)
		.sort(compareCommands)
		.map((command) => ({ command, available: isAvailable(command, context) }));
}

export function findByKeybinding(registry: CommandRegistry, event: KeyEventLike, platform: Platform, context: CommandContext): readonly RegisteredCommand[] {
	return Object.values(registry)
		.filter((command) => command.binding && matchesKeybinding(command.binding, event, platform))
		.filter((command) => isAvailable(command, context))
		.sort(compareCommands);
}

export function paletteEntries(registry: CommandRegistry, context: CommandContext, platform: Platform): readonly PaletteEntry[] {
	return listCommands(registry, context).map(({ command, available }) => ({
		id: command.id,
		label: command.title,
		group: command.category,
		shortcut: command.binding ? formatKeybinding(command.binding, platform) : undefined,
		disabled: !available
	}));
}
