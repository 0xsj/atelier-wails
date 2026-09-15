// Typed Preferences operations over a narrow transport port. Capabilities
// arrive as arguments; responses are validated by the codecs; every outcome is
// a kernel Result. The transport port is owned here because only this service
// consumes it.

import { absent, err, found, internalFailure, ok, type Failure, type Lookup, type Result } from '../../kernel';
import { encodeRequest, readListResponse, readOutcome, readReadResponse, readRemoveResponse, readReplaceResponse, readResolveResponse, type ReplaceResponse, type RemoveResponse, type ResolveResponse } from '../../platform/codecs/preferences';
import { WRITE_OPERATIONS, valuesEqual, type Expected, type PreferenceEntry, type PreferenceScope, type PreferenceValue, type PreferencesOperation, type Revision } from './model';

export interface PreferencesTransport {
	/** Sends one request document and resolves with the raw outcome envelope. A rejection means the facade was not reached or did not answer. */
	call(operation: PreferencesOperation, document: string): Promise<unknown>;
}

async function invoke<T>(transport: PreferencesTransport, operation: PreferencesOperation, document: string, reader: (value: unknown) => Result<T, string>): Promise<Result<T, Failure>> {
	let raw: unknown;
	try {
		raw = await transport.call(operation, document);
	} catch {
		return err(internalFailure(WRITE_OPERATIONS.has(operation) ? 'unknown' : 'none'));
	}
	return readOutcome(raw, operation, reader);
}

export function readPreference(transport: PreferencesTransport, scope: PreferenceScope, key: string): Promise<Lookup<PreferenceEntry, Failure>> {
	return invoke(transport, 'read', encodeRequest('read', { scope, key }), readReadResponse).then((result) => (result.ok ? ok(result.value.found ? found(result.value.entry) : absent()) : result));
}

export function listPreferences(transport: PreferencesTransport, scope: PreferenceScope): Promise<Result<readonly PreferenceEntry[], Failure>> {
	return invoke(transport, 'list', encodeRequest('list', { scope }), readListResponse).then((result) => (result.ok ? ok(result.value.entries) : result));
}

export function replacePreference(transport: PreferencesTransport, scope: PreferenceScope, key: string, value: PreferenceValue, expected: Expected): Promise<Result<ReplaceResponse, Failure>> {
	return invoke(transport, 'replace', encodeRequest('replace', { scope, key, value, expected }), readReplaceResponse);
}

export function removePreference(transport: PreferencesTransport, scope: PreferenceScope, key: string, expected: Expected): Promise<Result<RemoveResponse, Failure>> {
	return invoke(transport, 'remove', encodeRequest('remove', { scope, key, expected }), readRemoveResponse);
}

export function resolvePreference(transport: PreferencesTransport, scope: PreferenceScope, key: string, fallback: PreferenceValue): Promise<Result<ResolveResponse, Failure>> {
	return invoke(transport, 'resolve', encodeRequest('resolve', { scope, key, fallback }), readResolveResponse);
}

export type SetOutcome = { readonly status: 'changed' | 'unchanged' | 'reconciled'; readonly entry: PreferenceEntry; readonly attempts: number };
export type ClearOutcome = { readonly removed: boolean; readonly revision: Revision | null; readonly reconciled: boolean; readonly attempts: number };

function expectedFor(lookup: Lookup<PreferenceEntry, Failure>): Expected | null {
	if (!lookup.ok) return null;
	return lookup.value.present ? { kind: 'revision', revision: lookup.value.value.revision } : { kind: 'absent' };
}

/**
 * Sets a value with compare-and-replace: read the current revision, replace
 * against it, and on an uncertain commit re-read to see whether the write
 * landed instead of retrying blindly. A conflict re-reads and retries while
 * attempts remain. Never uses a blind overwrite.
 */
export async function setPreference(transport: PreferencesTransport, scope: PreferenceScope, key: string, value: PreferenceValue, options: { attempts?: number } = {}): Promise<Result<SetOutcome, Failure>> {
	const maxAttempts = Math.max(1, options.attempts ?? 2);
	let attempt = 0;
	let lastFailure: Failure | null = null;
	while (attempt < maxAttempts) {
		attempt += 1;
		const current = await readPreference(transport, scope, key);
		if (!current.ok) return current;
		if (current.value.present && valuesEqual(current.value.value.value, value)) return ok({ status: attempt === 1 ? 'unchanged' : 'reconciled', entry: current.value.value, attempts: attempt });
		const expected = expectedFor(current);
		if (!expected) return err(internalFailure('none'));
		const replaced = await replacePreference(transport, scope, key, value, expected);
		if (replaced.ok) return ok({ status: replaced.value.status, entry: replaced.value.entry, attempts: attempt });
		lastFailure = replaced.error;
		if (replaced.error.commit === 'unknown') {
			const after = await readPreference(transport, scope, key);
			if (!after.ok) return after;
			if (after.value.present && valuesEqual(after.value.value.value, value)) return ok({ status: 'reconciled', entry: after.value.value, attempts: attempt });
			continue;
		}
		if (replaced.error.kind === 'conflict') continue;
		return replaced;
	}
	return err(lastFailure ?? internalFailure('unknown'));
}

/** Removes a key with compare-and-remove and the same reconcile rule; an absent key is a successful no-op. */
export async function clearPreference(transport: PreferencesTransport, scope: PreferenceScope, key: string, options: { attempts?: number } = {}): Promise<Result<ClearOutcome, Failure>> {
	const maxAttempts = Math.max(1, options.attempts ?? 2);
	let attempt = 0;
	let lastFailure: Failure | null = null;
	while (attempt < maxAttempts) {
		attempt += 1;
		const current = await readPreference(transport, scope, key);
		if (!current.ok) return current;
		if (!current.value.present) return ok({ removed: false, revision: null, reconciled: attempt > 1, attempts: attempt });
		const removed = await removePreference(transport, scope, key, { kind: 'revision', revision: current.value.value.revision });
		if (removed.ok) return ok({ removed: removed.value.removed, revision: removed.value.removed ? removed.value.revision : null, reconciled: false, attempts: attempt });
		lastFailure = removed.error;
		if (removed.error.commit === 'unknown') {
			const after = await readPreference(transport, scope, key);
			if (!after.ok) return after;
			if (!after.value.present) return ok({ removed: true, revision: null, reconciled: true, attempts: attempt });
			continue;
		}
		if (removed.error.kind === 'conflict') continue;
		return removed;
	}
	return err(lastFailure ?? internalFailure('unknown'));
}
