// Browser preview transport for Preferences: an in-memory store that honors
// the desktop wire contract's shapes, decoding vocabulary, compare-and-replace
// semantics and commit rule, so the gallery works honestly without a native
// host. It can inject the contract's W11 faults to exercise reconciliation.
// It is a development double, not the native domain.

import type { PreferencesTransport } from '../../services/preferences/preferences';
import type { PreferencesOperation } from '../../services/preferences/model';
import { KEY_PATTERN, TEXT_LIMIT_BYTES, WRITE_OPERATIONS, textByteLength } from '../../services/preferences/model';

export type PreviewFault = {
	readonly kind: 'unavailable' | 'timeout';
	readonly mode: 'fail-before' | 'lose-acknowledgement';
	/** Limits the fault to the next call of this operation; other operations pass through untouched. */
	readonly operation?: PreferencesOperation;
};

export interface PreviewPreferencesTransport extends PreferencesTransport {
	/** Queues one fault for the next call, mirroring the contract's FailBefore and LoseAcknowledgement wrappers. */
	failNext(fault: PreviewFault | null): void;
	/** Wire-shaped entries for inspection, sorted by scope then key. */
	dump(): readonly Record<string, unknown>[];
	reset(): void;
}

type WireScope = { kind: 'global' } | { kind: 'workspace'; workspace_id: string };
type WireValue = { kind: 'text'; text: string } | { kind: 'bool'; bool: boolean } | { kind: 'int'; int: string };
type Stored = { scope: WireScope; key: string; value: WireValue; revision: bigint };

class Refusal extends Error {
	constructor(
		readonly kind: 'invalid' | 'conflict' | 'unavailable' | 'timeout',
		message: string,
		readonly type: string | null,
		readonly fields: Record<string, string> = {}
	) {
		super(message);
	}
}

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const INT64_MIN = -(2n ** 63n);
const INT64_MAX = 2n ** 63n - 1n;
const REVISION_MAX = 2n ** 64n - 1n;

function invalidRequest(path: string, problem: 'missing' | 'unknown' | 'invalid'): Refusal {
	return new Refusal('invalid', 'invalid request', 'desktop.invalid_request', { [path]: problem });
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function decodeScope(input: unknown, path: string): WireScope {
	const raw = isRecord(input) ? input : {};
	if (raw.kind === undefined || raw.kind === '') throw invalidRequest(`${path}.kind`, 'missing');
	if (raw.kind === 'global') return { kind: 'global' };
	if (raw.kind !== 'workspace') throw invalidRequest(`${path}.kind`, 'unknown');
	if (typeof raw.workspace_id !== 'string' || raw.workspace_id === '') throw invalidRequest(`${path}.workspace_id`, 'missing');
	if (!UUID.test(raw.workspace_id)) throw invalidRequest(`${path}.workspace_id`, 'invalid');
	return { kind: 'workspace', workspace_id: raw.workspace_id.toLowerCase() };
}

function decodeKey(input: unknown): string {
	if (typeof input !== 'string' || !KEY_PATTERN.test(input)) throw new Refusal('invalid', 'invalid preference key', 'preferences.invalid_key');
	return input;
}

function decodeValue(input: unknown, path: string): WireValue {
	const raw = isRecord(input) ? input : {};
	if (raw.kind === undefined || raw.kind === '') throw invalidRequest(`${path}.kind`, 'missing');
	switch (raw.kind) {
		case 'text': {
			if (typeof raw.text !== 'string') throw invalidRequest(`${path}.text`, 'missing');
			if (textByteLength(raw.text) > TEXT_LIMIT_BYTES) throw new Refusal('invalid', 'invalid preference value', 'preferences.invalid_value');
			return { kind: 'text', text: raw.text };
		}
		case 'bool':
			if (typeof raw.bool !== 'boolean') throw invalidRequest(`${path}.bool`, 'missing');
			return { kind: 'bool', bool: raw.bool };
		case 'int': {
			if (typeof raw.int !== 'string') throw invalidRequest(`${path}.int`, 'missing');
			if (!/^-?[0-9]+$/.test(raw.int)) throw invalidRequest(`${path}.int`, 'invalid');
			const parsed = BigInt(raw.int);
			if (parsed < INT64_MIN || parsed > INT64_MAX) throw invalidRequest(`${path}.int`, 'invalid');
			return { kind: 'int', int: parsed.toString(10) };
		}
		default:
			throw invalidRequest(`${path}.kind`, 'unknown');
	}
}

function decodeExpected(input: unknown, path: string): { kind: 'absent' } | { kind: 'revision'; revision: bigint } {
	const raw = isRecord(input) ? input : {};
	if (raw.kind === undefined || raw.kind === '') throw invalidRequest(`${path}.kind`, 'missing');
	if (raw.kind === 'absent') return { kind: 'absent' };
	if (raw.kind !== 'revision') throw invalidRequest(`${path}.kind`, 'unknown');
	if (typeof raw.revision !== 'string' || raw.revision === '') throw invalidRequest(`${path}.revision`, 'missing');
	if (!/^[0-9]+$/.test(raw.revision)) throw invalidRequest(`${path}.revision`, 'invalid');
	const revision = BigInt(raw.revision);
	if (revision === 0n) throw new Refusal('invalid', 'invalid preference revision', 'preferences.invalid_revision');
	return { kind: 'revision', revision };
}

function valuesEqual(a: WireValue, b: WireValue): boolean {
	return JSON.stringify(a) === JSON.stringify(b);
}

function scopeKey(scope: WireScope): string {
	return scope.kind === 'global' ? 'global' : `workspace:${scope.workspace_id}`;
}

function encodeEntry(entry: Stored): Record<string, unknown> {
	return { scope: entry.scope, key: entry.key, value: entry.value, revision: entry.revision.toString(10) };
}

function changedEvent(entry: Stored): Record<string, unknown> {
	return { name: 'preference.changed', scope: entry.scope, key: entry.key, revision: entry.revision.toString(10), value: entry.value };
}

function compareBytes(a: string, b: string): number {
	const left = new TextEncoder().encode(a);
	const right = new TextEncoder().encode(b);
	const length = Math.min(left.length, right.length);
	for (let i = 0; i < length; i += 1) if (left[i] !== right[i]) return left[i] - right[i];
	return left.length - right.length;
}

export function createPreviewPreferencesTransport(): PreviewPreferencesTransport {
	const store = new Map<string, Map<string, Stored>>();
	let queued: PreviewFault | null = null;

	function bucket(scope: WireScope): Map<string, Stored> {
		const id = scopeKey(scope);
		let entries = store.get(id);
		if (!entries) {
			entries = new Map();
			store.set(id, entries);
		}
		return entries;
	}

	function decodeDocument(document: string): Record<string, unknown> {
		let parsed: unknown;
		try {
			parsed = JSON.parse(document);
		} catch {
			throw invalidRequest('request', 'invalid');
		}
		if (!isRecord(parsed)) throw invalidRequest('request', 'invalid');
		return parsed;
	}

	function fault(kind: 'unavailable' | 'timeout'): Refusal {
		return new Refusal(kind, kind === 'timeout' ? 'preference store timed out' : 'preference store unavailable', `preview.${kind}`);
	}

	function perform(operation: PreferencesOperation, request: Record<string, unknown>): unknown {
		const scope = decodeScope(request.scope, 'scope');
		const entries = bucket(scope);
		const active = queued && (!queued.operation || queued.operation === operation) ? queued : null;
		if (active) queued = null;
		if (active?.mode === 'fail-before') throw fault(active.kind);
		const loseAcknowledgement = active?.mode === 'lose-acknowledgement';
		let response: unknown;
		switch (operation) {
			case 'read': {
				const entry = entries.get(decodeKey(request.key));
				response = entry ? { found: true, entry: encodeEntry(entry) } : { found: false, entry: null };
				break;
			}
			case 'list':
				response = { entries: [...entries.values()].sort((a, b) => compareBytes(a.key, b.key)).map(encodeEntry) };
				break;
			case 'replace': {
				const key = decodeKey(request.key);
				const value = decodeValue(request.value, 'value');
				const expected = decodeExpected(request.expected, 'expected');
				const current = entries.get(key);
				if (expected.kind === 'absent' ? current !== undefined : current === undefined || current.revision !== expected.revision) throw new Refusal('conflict', 'preference revision conflict', 'preferences.conflict');
				if (current && valuesEqual(current.value, value)) {
					response = { status: 'unchanged', entry: encodeEntry(current), event: null };
					break;
				}
				if (current && current.revision === REVISION_MAX) throw new Refusal('conflict', 'preference revision exhausted', 'preferences.revision_exhausted');
				const next: Stored = { scope, key, value, revision: current ? current.revision + 1n : 1n };
				entries.set(key, next);
				response = { status: 'changed', entry: encodeEntry(next), event: changedEvent(next) };
				break;
			}
			case 'remove': {
				const key = decodeKey(request.key);
				const expected = decodeExpected(request.expected, 'expected');
				const current = entries.get(key);
				if (!current) {
					if (expected.kind === 'revision') throw new Refusal('conflict', 'preference revision conflict', 'preferences.conflict');
					response = { removed: false, revision: null, event: null };
					break;
				}
				if (expected.kind === 'absent' || current.revision !== expected.revision) throw new Refusal('conflict', 'preference revision conflict', 'preferences.conflict');
				if (current.revision === REVISION_MAX) throw new Refusal('conflict', 'preference revision exhausted', 'preferences.revision_exhausted');
				entries.delete(key);
				const revision = (current.revision + 1n).toString(10);
				response = { removed: true, revision, event: { name: 'preference.removed', scope, key, revision, value: null } };
				break;
			}
			case 'resolve': {
				const key = decodeKey(request.key);
				const fallback = decodeValue(request.fallback, 'fallback');
				const entry = entries.get(key);
				response = entry ? { value: entry.value, stored: true, revision: entry.revision.toString(10) } : { value: fallback, stored: false, revision: null };
				break;
			}
		}
		if (loseAcknowledgement && active) throw fault(active.kind);
		return response;
	}

	return {
		async call(operation, document) {
			const write = WRITE_OPERATIONS.has(operation);
			try {
				const request = decodeDocument(document);
				return { ok: true, value: perform(operation, request) };
			} catch (error) {
				if (!(error instanceof Refusal)) throw error;
				const commit = !write ? 'none' : error.kind === 'invalid' || error.kind === 'conflict' ? 'not_applied' : 'unknown';
				return { ok: false, failure: { kind: error.kind, message: error.message, type: error.type, fields: { ...error.fields }, commit } };
			}
		},
		failNext(fault) {
			queued = fault;
		},
		dump() {
			return [...store.entries()].sort(([a], [b]) => a.localeCompare(b)).flatMap(([, entries]) => [...entries.values()].sort((a, b) => compareBytes(a.key, b.key)).map(encodeEntry));
		},
		reset() {
			store.clear();
			queued = null;
		}
	};
}
