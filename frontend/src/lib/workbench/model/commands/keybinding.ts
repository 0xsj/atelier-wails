// Pure keybinding values: parse a chord like "mod+shift+p", match it against a
// key event shape, and format it for the platform. No DOM types are required;
// KeyEventLike is the subset the browser event provides.

export type Platform = 'mac' | 'windows' | 'linux';

export interface Keybinding {
	readonly key: string;
	readonly mod: boolean;
	readonly ctrl: boolean;
	readonly shift: boolean;
	readonly alt: boolean;
}

export interface KeyEventLike {
	readonly key: string;
	readonly metaKey: boolean;
	readonly ctrlKey: boolean;
	readonly shiftKey: boolean;
	readonly altKey: boolean;
}

const MODIFIERS: Record<string, keyof Omit<Keybinding, 'key'>> = {
	mod: 'mod',
	cmd: 'mod',
	meta: 'mod',
	ctrl: 'ctrl',
	control: 'ctrl',
	shift: 'shift',
	alt: 'alt',
	option: 'alt'
};

const KEY_ALIASES: Record<string, string> = {
	esc: 'escape',
	return: 'enter',
	del: 'delete',
	up: 'arrowup',
	down: 'arrowdown',
	left: 'arrowleft',
	right: 'arrowright',
	' ': 'space',
	spacebar: 'space',
	plus: '+'
};

export function normalizeKey(key: string): string {
	const lower = key.toLowerCase();
	return KEY_ALIASES[lower] ?? lower;
}

export function parseKeybinding(text: string): Keybinding | null {
	const tokens = text.trim().toLowerCase().split('+').map((token) => token.trim());
	if (tokens.some((token) => token === '') && text.trim() !== '+') return null;
	const binding = { key: '', mod: false, ctrl: false, shift: false, alt: false };
	for (const token of tokens.length ? tokens : ['']) {
		const modifier = MODIFIERS[token];
		if (modifier) {
			if (binding[modifier]) return null;
			binding[modifier] = true;
			continue;
		}
		if (binding.key) return null;
		binding.key = normalizeKey(token);
	}
	if (!binding.key) return null;
	return binding;
}

export function matchesKeybinding(binding: Keybinding, event: KeyEventLike, platform: Platform): boolean {
	if (normalizeKey(event.key) !== binding.key || event.shiftKey !== binding.shift || event.altKey !== binding.alt) return false;
	if (platform === 'mac') return event.metaKey === binding.mod && event.ctrlKey === binding.ctrl;
	// Outside mac there is no separate command key: mod and ctrl both mean Ctrl, and the meta key never matches.
	return !event.metaKey && event.ctrlKey === (binding.mod || binding.ctrl);
}

const MAC_KEY_GLYPHS: Record<string, string> = {
	enter: '↩',
	escape: '⎋',
	backspace: '⌫',
	delete: '⌦',
	tab: '⇥',
	space: 'Space',
	arrowup: '↑',
	arrowdown: '↓',
	arrowleft: '←',
	arrowright: '→'
};

const KEY_LABELS: Record<string, string> = {
	enter: 'Enter',
	escape: 'Esc',
	backspace: 'Backspace',
	delete: 'Delete',
	tab: 'Tab',
	space: 'Space',
	arrowup: 'Up',
	arrowdown: 'Down',
	arrowleft: 'Left',
	arrowright: 'Right'
};

function keyLabel(key: string, platform: Platform): string {
	if (platform === 'mac' && MAC_KEY_GLYPHS[key]) return MAC_KEY_GLYPHS[key];
	if (KEY_LABELS[key]) return KEY_LABELS[key];
	return key.length === 1 ? key.toUpperCase() : key.charAt(0).toUpperCase() + key.slice(1);
}

export function formatKeybinding(binding: Keybinding, platform: Platform): string {
	if (platform === 'mac') {
		return `${binding.ctrl ? '⌃' : ''}${binding.alt ? '⌥' : ''}${binding.shift ? '⇧' : ''}${binding.mod ? '⌘' : ''}${keyLabel(binding.key, platform)}`;
	}
	const parts: string[] = [];
	if (binding.mod || binding.ctrl) parts.push('Ctrl');
	if (binding.alt) parts.push('Alt');
	if (binding.shift) parts.push('Shift');
	parts.push(keyLabel(binding.key, platform));
	return parts.join('+');
}
