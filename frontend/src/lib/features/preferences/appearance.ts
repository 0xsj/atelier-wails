// Theme and density as stored preferences. The first real consumer of the
// Preferences round trip: values load from the store with fallbacks, apply
// as root data attributes through the runtime, and write back with
// compare-and-replace. No Svelte import; the gallery binds the result.

import type { Failure, Result } from '../../kernel';
import { DENSITIES, THEMES, applyDensity, applyTheme, type Density, type Theme } from '../../runtime';
import { GLOBAL_SCOPE, resolvePreference, setPreference, text, type PreferencesTransport } from '../../services/preferences';

export const APPEARANCE_KEYS = { theme: 'ui.theme', density: 'ui.density' } as const;

export type AppearanceSource = 'stored' | 'fallback' | 'failed';

export interface AppearanceState {
	readonly theme: Theme;
	readonly density: Density;
	readonly themeSource: AppearanceSource;
	readonly densitySource: AppearanceSource;
	readonly failure: Failure | null;
}

function pick<T extends string>(candidates: readonly T[], value: unknown, fallback: T): T {
	return typeof value === 'string' && (candidates as readonly string[]).includes(value) ? (value as T) : fallback;
}

export interface AppearanceSync {
	load(): Promise<AppearanceState>;
	setTheme(theme: Theme): Promise<Result<unknown, Failure>>;
	setDensity(density: Density): Promise<Result<unknown, Failure>>;
}

export function createAppearanceSync(transport: PreferencesTransport): AppearanceSync {
	return {
		async load() {
			const [theme, density] = await Promise.all([
				resolvePreference(transport, GLOBAL_SCOPE, APPEARANCE_KEYS.theme, text('system')),
				resolvePreference(transport, GLOBAL_SCOPE, APPEARANCE_KEYS.density, text('comfortable'))
			]);
			const themeValue = theme.ok ? pick(THEMES, theme.value.value.kind === 'text' ? theme.value.value.text : null, 'system') : 'system';
			const densityValue = density.ok ? pick(DENSITIES, density.value.value.kind === 'text' ? density.value.value.text : null, 'comfortable') : 'comfortable';
			applyTheme(themeValue);
			applyDensity(densityValue);
			return {
				theme: themeValue,
				density: densityValue,
				themeSource: theme.ok ? (theme.value.stored ? 'stored' : 'fallback') : 'failed',
				densitySource: density.ok ? (density.value.stored ? 'stored' : 'fallback') : 'failed',
				failure: !theme.ok ? theme.error : !density.ok ? density.error : null
			};
		},
		setTheme(theme) {
			applyTheme(theme);
			return setPreference(transport, GLOBAL_SCOPE, APPEARANCE_KEYS.theme, text(theme));
		},
		setDensity(density) {
			applyDensity(density);
			return setPreference(transport, GLOBAL_SCOPE, APPEARANCE_KEYS.density, text(density));
		}
	};
}
