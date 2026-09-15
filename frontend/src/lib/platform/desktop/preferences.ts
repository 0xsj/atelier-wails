// Wails adapter for the Preferences transport port. The only frontend module
// that reaches the bound Go struct for this capability. Bound methods take one
// document string and resolve with the outcome envelope as plain data; the
// binding lives on the runtime global that Wails installs, so this file does
// not depend on generated binding files existing at type-check time.

import type { PreferencesOperation } from '../../services/preferences/model';
import type { PreferencesTransport } from '../../services/preferences/preferences';

const METHODS: Readonly<Record<PreferencesOperation, string>> = {
	read: 'Read',
	list: 'List',
	replace: 'Replace',
	remove: 'Remove',
	resolve: 'Resolve'
};

type BoundPreferences = Readonly<Record<string, (document: string) => Promise<unknown>>>;

function boundPreferences(): BoundPreferences | null {
	const go = (globalThis as { go?: { wailshost?: { Preferences?: BoundPreferences } } }).go;
	return go?.wailshost?.Preferences ?? null;
}

export function createDesktopPreferencesTransport(): PreferencesTransport {
	return {
		call(operation, document) {
			const bound = boundPreferences();
			const method = bound?.[METHODS[operation]];
			if (!method) return Promise.reject(new Error('Wails preferences binding is unavailable'));
			return method(document);
		}
	};
}
