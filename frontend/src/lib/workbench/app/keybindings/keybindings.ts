// Keybinding scope and precedence. Text editing and composition input win
// over plain-key shortcuts; modifier chords still reach the workbench.

import { findByKeybinding, type CommandContext, type CommandRegistry, type RegisteredCommand } from '../../model/commands/commands';
import { normalizeKey, type KeyEventLike, type Platform } from '../../model/commands/keybinding';

export type KeybindingScope = 'workbench' | 'text-input';

/** The subset of an event target the scope rule reads; a real DOM element satisfies it and anything else counts as the workbench. */
export interface TargetLike {
	readonly tagName?: string;
	readonly isContentEditable?: boolean;
	getAttribute?(name: string): string | null;
}

function asTarget(value: unknown): TargetLike | null {
	if (typeof value !== 'object' || value === null) return null;
	const candidate = value as { tagName?: unknown; isContentEditable?: unknown; getAttribute?: unknown };
	return {
		tagName: typeof candidate.tagName === 'string' ? candidate.tagName : undefined,
		isContentEditable: candidate.isContentEditable === true,
		getAttribute: typeof candidate.getAttribute === 'function' ? (name: string) => (candidate.getAttribute as (name: string) => string | null).call(value, name) : undefined
	};
}

export interface ShortcutEventLike extends KeyEventLike {
	readonly isComposing?: boolean;
	readonly repeat?: boolean;
	readonly target?: unknown;
}

export type ShortcutResolution =
	| { readonly kind: 'match'; readonly command: RegisteredCommand; readonly scope: KeybindingScope }
	| { readonly kind: 'none'; readonly scope: KeybindingScope }
	| { readonly kind: 'composing' }
	| { readonly kind: 'reserved-for-text'; readonly scope: KeybindingScope };

const TEXT_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT']);
const NON_TEXT_INPUT_TYPES = new Set(['button', 'checkbox', 'radio', 'range', 'submit', 'reset', 'file', 'color']);
const PASSTHROUGH_KEYS = new Set(['escape', 'f1', 'f2', 'f3', 'f4', 'f5', 'f6', 'f7', 'f8', 'f9', 'f10', 'f11', 'f12']);

export function scopeForTarget(value: unknown): KeybindingScope {
	const target = asTarget(value);
	if (!target) return 'workbench';
	if (target.isContentEditable) return 'text-input';
	const tag = target.tagName?.toUpperCase();
	if (!tag || !TEXT_TAGS.has(tag)) return 'workbench';
	if (tag === 'INPUT') {
		const type = (target.getAttribute?.('type') ?? 'text').toLowerCase();
		if (NON_TEXT_INPUT_TYPES.has(type)) return 'workbench';
	}
	return 'text-input';
}

/** In a text field only modifier chords (excluding shift alone) and function/escape keys reach the workbench. */
export function reachesWorkbench(event: KeyEventLike, scope: KeybindingScope): boolean {
	if (scope === 'workbench') return true;
	if (event.metaKey || event.ctrlKey || event.altKey) return true;
	return PASSTHROUGH_KEYS.has(normalizeKey(event.key));
}

export function resolveShortcut(registry: CommandRegistry, event: ShortcutEventLike, platform: Platform, context: CommandContext = {}): ShortcutResolution {
	if (event.isComposing) return { kind: 'composing' };
	const scope = scopeForTarget(event.target);
	if (!reachesWorkbench(event, scope)) return { kind: 'reserved-for-text', scope };
	const [command] = findByKeybinding(registry, event, platform, context);
	return command ? { kind: 'match', command, scope } : { kind: 'none', scope };
}
