// Frontend values for the Preferences context, as the wire contract defines
// them. Revisions stay decimal strings; int values are bigint so nothing is
// rounded. These are plain data, not the native domain.

import type { WorkspaceId } from '../../kernel';

export type PreferenceScope = { readonly kind: 'global' } | { readonly kind: 'workspace'; readonly workspaceId: WorkspaceId };

export type PreferenceValue =
	| { readonly kind: 'text'; readonly text: string }
	| { readonly kind: 'bool'; readonly value: boolean }
	| { readonly kind: 'int'; readonly value: bigint };

/** Decimal digits as the wire carries them; compared for equality, never parsed to number. */
export type Revision = string;

export type Expected = { readonly kind: 'absent' } | { readonly kind: 'revision'; readonly revision: Revision };

export interface PreferenceEntry {
	readonly scope: PreferenceScope;
	readonly key: string;
	readonly value: PreferenceValue;
	readonly revision: Revision;
}

export interface PreferenceEvent {
	readonly name: 'preference.changed' | 'preference.removed';
	readonly scope: PreferenceScope;
	readonly key: string;
	readonly revision: Revision;
	readonly value: PreferenceValue | null;
}

export type PreferencesOperation = 'read' | 'list' | 'replace' | 'remove' | 'resolve';

export const WRITE_OPERATIONS: ReadonlySet<PreferencesOperation> = new Set(['replace', 'remove']);

export const GLOBAL_SCOPE: PreferenceScope = { kind: 'global' };

export const KEY_PATTERN = /^[a-z][a-z0-9_.-]{0,127}$/;

export const TEXT_LIMIT_BYTES = 4096;

export function text(value: string): PreferenceValue {
	return { kind: 'text', text: value };
}

export function bool(value: boolean): PreferenceValue {
	return { kind: 'bool', value };
}

export function int(value: bigint): PreferenceValue {
	return { kind: 'int', value };
}

export function valuesEqual(a: PreferenceValue, b: PreferenceValue): boolean {
	if (a.kind !== b.kind) return false;
	if (a.kind === 'text') return a.text === (b as { text: string }).text;
	return a.value === (b as { value: boolean | bigint }).value;
}

export function scopesEqual(a: PreferenceScope, b: PreferenceScope): boolean {
	if (a.kind !== b.kind) return false;
	return a.kind === 'global' || a.workspaceId === (b as { workspaceId: WorkspaceId }).workspaceId;
}

export function isValidKey(key: string): boolean {
	return KEY_PATTERN.test(key);
}

export function textByteLength(value: string): number {
	return new TextEncoder().encode(value).length;
}
