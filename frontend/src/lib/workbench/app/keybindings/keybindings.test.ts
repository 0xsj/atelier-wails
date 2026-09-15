import { describe, expect, it } from 'vitest';
import { EMPTY_REGISTRY, registerCommands } from '../../model/commands/commands';
import { resolveShortcut, scopeForTarget, type ShortcutEventLike } from './keybindings';

const { registry } = registerCommands(EMPTY_REGISTRY, [
	{ id: 'view.toggleSidebar', title: 'Toggle Sidebar', keybinding: 'mod+b' },
	{ id: 'editor.next', title: 'Next Editor', keybinding: 'alt+arrowright' },
	{ id: 'search.focus', title: 'Focus Search', keybinding: '/' },
	{ id: 'palette.close', title: 'Close', keybinding: 'escape' }
]);

const press = (key: string, extra: Partial<ShortcutEventLike> = {}): ShortcutEventLike => ({ key, metaKey: false, ctrlKey: false, shiftKey: false, altKey: false, ...extra });
const input = { tagName: 'INPUT', getAttribute: () => 'text' };

describe('keybinding scopes', () => {
	it('classifies targets', () => {
		expect(scopeForTarget(null)).toBe('workbench');
		expect(scopeForTarget({ tagName: 'DIV' })).toBe('workbench');
		expect(scopeForTarget(input)).toBe('text-input');
		expect(scopeForTarget({ tagName: 'textarea' })).toBe('text-input');
		expect(scopeForTarget({ tagName: 'INPUT', getAttribute: () => 'checkbox' })).toBe('workbench');
		expect(scopeForTarget({ tagName: 'DIV', isContentEditable: true })).toBe('text-input');
	});

	it('lets modifier chords through a text field but reserves plain keys for it', () => {
		expect(resolveShortcut(registry, press('b', { metaKey: true, target: input }), 'mac')).toMatchObject({ kind: 'match', scope: 'text-input' });
		expect(resolveShortcut(registry, press('/', { target: input }), 'mac')).toEqual({ kind: 'reserved-for-text', scope: 'text-input' });
		expect(resolveShortcut(registry, press('/', { target: { tagName: 'DIV' } }), 'mac')).toMatchObject({ kind: 'match' });
		expect(resolveShortcut(registry, press('Escape', { target: input }), 'mac')).toMatchObject({ kind: 'match' });
		expect(resolveShortcut(registry, press('ArrowRight', { altKey: true, target: input }), 'windows')).toMatchObject({ kind: 'match' });
	});

	it('ignores composition and unbound keys', () => {
		expect(resolveShortcut(registry, press('b', { metaKey: true, isComposing: true }), 'mac')).toEqual({ kind: 'composing' });
		expect(resolveShortcut(registry, press('q', { metaKey: true }), 'mac')).toEqual({ kind: 'none', scope: 'workbench' });
	});
});
