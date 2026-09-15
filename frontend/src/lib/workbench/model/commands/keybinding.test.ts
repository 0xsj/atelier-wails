import { describe, expect, it } from 'vitest';
import { formatKeybinding, matchesKeybinding, parseKeybinding, type KeyEventLike } from './keybinding';

const press = (key: string, mods: Partial<Omit<KeyEventLike, 'key'>> = {}): KeyEventLike => ({ key, metaKey: false, ctrlKey: false, shiftKey: false, altKey: false, ...mods });

describe('keybinding', () => {
	it('parses modifiers and keys, rejecting malformed chords', () => {
		expect(parseKeybinding('mod+shift+p')).toEqual({ key: 'p', mod: true, ctrl: false, shift: true, alt: false });
		expect(parseKeybinding(' Ctrl + Alt + Delete ')).toEqual({ key: 'delete', mod: false, ctrl: true, shift: false, alt: true });
		expect(parseKeybinding('cmd+enter')?.key).toBe('enter');
		expect(parseKeybinding('mod')).toBeNull();
		expect(parseKeybinding('mod+mod+p')).toBeNull();
		expect(parseKeybinding('a+b')).toBeNull();
		expect(parseKeybinding('')).toBeNull();
	});

	it('maps mod to meta on mac and ctrl elsewhere, and requires exact modifiers', () => {
		const binding = parseKeybinding('mod+b');
		if (!binding) throw new Error('parse failed');
		expect(matchesKeybinding(binding, press('b', { metaKey: true }), 'mac')).toBe(true);
		expect(matchesKeybinding(binding, press('b', { ctrlKey: true }), 'mac')).toBe(false);
		expect(matchesKeybinding(binding, press('b', { ctrlKey: true }), 'windows')).toBe(true);
		expect(matchesKeybinding(binding, press('b', { metaKey: true }), 'linux')).toBe(false);
		expect(matchesKeybinding(binding, press('B', { metaKey: true, shiftKey: true }), 'mac')).toBe(false);
		expect(matchesKeybinding(binding, press('b'), 'mac')).toBe(false);
	});

	it('treats explicit ctrl as control on mac and as ctrl elsewhere', () => {
		const binding = parseKeybinding('ctrl+`');
		if (!binding) throw new Error('parse failed');
		expect(matchesKeybinding(binding, press('`', { ctrlKey: true }), 'mac')).toBe(true);
		expect(matchesKeybinding(binding, press('`', { metaKey: true }), 'mac')).toBe(false);
		expect(matchesKeybinding(binding, press('`', { ctrlKey: true }), 'windows')).toBe(true);
	});

	it('formats for each platform', () => {
		const binding = parseKeybinding('mod+shift+p');
		if (!binding) throw new Error('parse failed');
		expect(formatKeybinding(binding, 'mac')).toBe('⇧⌘P');
		expect(formatKeybinding(binding, 'windows')).toBe('Ctrl+Shift+P');
		expect(formatKeybinding(parseKeybinding('alt+arrowup')!, 'mac')).toBe('⌥↑');
		expect(formatKeybinding(parseKeybinding('ctrl+alt+enter')!, 'linux')).toBe('Ctrl+Alt+Enter');
		expect(formatKeybinding(parseKeybinding('escape')!, 'windows')).toBe('Esc');
	});
});
