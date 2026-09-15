import { describe, expect, it } from 'vitest';
import { EMPTY_REGISTRY, findByKeybinding, listCommands, paletteEntries, registerCommand, registerCommands, resolveCommand, unregisterCommand } from './commands';

const definitions = [
	{ id: 'view.toggleSidebar', title: 'Toggle Sidebar', category: 'View', keybinding: 'mod+b' },
	{ id: 'file.save', title: 'Save', category: 'File', keybinding: 'mod+s', when: (context: { activeViewDirty?: boolean | string | number | null }) => context.activeViewDirty === true },
	{ id: 'file.close', title: 'Close Editor', category: 'File', keybinding: 'mod+w', when: (context: { hasActiveView?: boolean | string | number | null }) => context.hasActiveView === true },
	{ id: 'palette.open', title: 'Show All Commands', keybinding: 'mod+shift+p' }
];

describe('command registry', () => {
	it('registers commands, rejecting duplicates and bad keybindings without losing the rest', () => {
		const { registry, rejected } = registerCommands(EMPTY_REGISTRY, [...definitions, { id: 'file.save', title: 'Again' }, { id: 'bad', title: 'Bad', keybinding: 'mod+' }]);
		expect(Object.keys(registry).sort()).toEqual(['file.close', 'file.save', 'palette.open', 'view.toggleSidebar']);
		expect(rejected.map((outcome) => outcome.kind)).toEqual(['duplicate', 'invalid-keybinding']);
		expect(registerCommand(registry, definitions[0]).kind).toBe('duplicate');
		expect(unregisterCommand(registry, 'nope')).toBe(registry);
		expect(Object.keys(unregisterCommand(registry, 'file.save'))).not.toContain('file.save');
	});

	it('resolves availability from context', () => {
		const { registry } = registerCommands(EMPTY_REGISTRY, definitions);
		expect(resolveCommand(registry, 'file.save', { activeViewDirty: false }).kind).toBe('unavailable');
		expect(resolveCommand(registry, 'file.save', { activeViewDirty: true }).kind).toBe('available');
		expect(resolveCommand(registry, 'view.toggleSidebar', {}).kind).toBe('available');
		expect(resolveCommand(registry, 'missing', {})).toEqual({ kind: 'unknown', id: 'missing' });
	});

	it('lists commands sorted by category then title with availability', () => {
		const { registry } = registerCommands(EMPTY_REGISTRY, definitions);
		const listing = listCommands(registry, { hasActiveView: true });
		expect(listing.map((entry) => entry.command.id)).toEqual(['palette.open', 'file.close', 'file.save', 'view.toggleSidebar']);
		expect(listing.map((entry) => entry.available)).toEqual([true, true, false, true]);
	});

	it('finds available commands by key event per platform', () => {
		const { registry } = registerCommands(EMPTY_REGISTRY, definitions);
		const event = { key: 's', metaKey: true, ctrlKey: false, shiftKey: false, altKey: false };
		expect(findByKeybinding(registry, event, 'mac', { activeViewDirty: true }).map((command) => command.id)).toEqual(['file.save']);
		expect(findByKeybinding(registry, event, 'mac', { activeViewDirty: false })).toEqual([]);
		expect(findByKeybinding(registry, event, 'windows', { activeViewDirty: true })).toEqual([]);
		expect(findByKeybinding(registry, { ...event, key: 'P', shiftKey: true }, 'mac', {}).map((command) => command.id)).toEqual(['palette.open']);
	});

	it('produces palette entries with formatted shortcuts', () => {
		const { registry } = registerCommands(EMPTY_REGISTRY, definitions);
		const entries = paletteEntries(registry, { hasActiveView: false }, 'mac');
		expect(entries.find((entry) => entry.id === 'file.close')).toEqual({ id: 'file.close', label: 'Close Editor', group: 'File', shortcut: '⌘W', disabled: true });
		expect(paletteEntries(registry, {}, 'linux').find((entry) => entry.id === 'palette.open')?.shortcut).toBe('Ctrl+Shift+P');
	});
});
