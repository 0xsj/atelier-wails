// Total readers and writers for the Preferences desktop wire contract
// (revision 1). Readers accept unknown and never trust a cast; a malformed
// response becomes a frontend failure with type `codec.invalid_response`.

import { err, failure, internalFailure, ok, parseCommitState, parseFailureKind, parseWorkspaceId, type Failure, type Result } from '../../kernel';
import { WRITE_OPERATIONS, type Expected, type PreferenceEntry, type PreferenceEvent, type PreferenceScope, type PreferenceValue, type PreferencesOperation } from '../../services/preferences/model';

export type WireScope = { kind: 'global' } | { kind: 'workspace'; workspace_id: string };
export type WireValue = { kind: 'text'; text: string } | { kind: 'bool'; bool: boolean } | { kind: 'int'; int: string };
export type WireExpected = { kind: 'absent' } | { kind: 'revision'; revision: string };

export interface ReadRequest { scope: PreferenceScope; key: string }
export interface ListRequest { scope: PreferenceScope }
export interface ReplaceRequest { scope: PreferenceScope; key: string; value: PreferenceValue; expected: Expected }
export interface RemoveRequest { scope: PreferenceScope; key: string; expected: Expected }
export interface ResolveRequest { scope: PreferenceScope; key: string; fallback: PreferenceValue }

export type ReadResponse = { found: true; entry: PreferenceEntry } | { found: false };
export type ListResponse = { entries: readonly PreferenceEntry[] };
export type ReplaceResponse = { status: 'changed' | 'unchanged'; entry: PreferenceEntry; event: PreferenceEvent | null };
export type RemoveResponse = { removed: true; revision: string; event: PreferenceEvent } | { removed: false };
export type ResolveResponse = { value: PreferenceValue; stored: true; revision: string } | { value: PreferenceValue; stored: false };

export const CODEC_INVALID_RESPONSE = 'codec.invalid_response';

const INT_PATTERN = /^-?[0-9]+$/;
const REVISION_PATTERN = /^[0-9]+$/;

// Encoding

export function encodeScope(scope: PreferenceScope): WireScope {
	return scope.kind === 'global' ? { kind: 'global' } : { kind: 'workspace', workspace_id: scope.workspaceId };
}

export function encodeValue(value: PreferenceValue): WireValue {
	switch (value.kind) {
		case 'text':
			return { kind: 'text', text: value.text };
		case 'bool':
			return { kind: 'bool', bool: value.value };
		case 'int':
			return { kind: 'int', int: value.value.toString(10) };
	}
}

export function encodeExpected(expected: Expected): WireExpected {
	return expected.kind === 'absent' ? { kind: 'absent' } : { kind: 'revision', revision: expected.revision };
}

export function encodeRequest(operation: 'read', request: ReadRequest): string;
export function encodeRequest(operation: 'list', request: ListRequest): string;
export function encodeRequest(operation: 'replace', request: ReplaceRequest): string;
export function encodeRequest(operation: 'remove', request: RemoveRequest): string;
export function encodeRequest(operation: 'resolve', request: ResolveRequest): string;
export function encodeRequest(operation: PreferencesOperation, request: ReadRequest | ListRequest | ReplaceRequest | RemoveRequest | ResolveRequest): string {
	const scope = encodeScope(request.scope);
	switch (operation) {
		case 'read':
			return JSON.stringify({ scope, key: (request as ReadRequest).key });
		case 'list':
			return JSON.stringify({ scope });
		case 'replace': {
			const r = request as ReplaceRequest;
			return JSON.stringify({ scope, key: r.key, value: encodeValue(r.value), expected: encodeExpected(r.expected) });
		}
		case 'remove': {
			const r = request as RemoveRequest;
			return JSON.stringify({ scope, key: r.key, expected: encodeExpected(r.expected) });
		}
		case 'resolve': {
			const r = request as ResolveRequest;
			return JSON.stringify({ scope, key: r.key, fallback: encodeValue(r.fallback) });
		}
	}
}

// Decoding

type Problem = string;
type Read<T> = Result<T, Problem>;

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function at(path: string, problem: string): Problem {
	return `${path}: ${problem}`;
}

export function readScope(input: unknown, path = 'scope'): Read<PreferenceScope> {
	if (!isRecord(input)) return err(at(path, 'not an object'));
	if (input.kind === 'global') return ok({ kind: 'global' });
	if (input.kind === 'workspace') {
		const workspaceId = parseWorkspaceId(input.workspace_id);
		return workspaceId ? ok({ kind: 'workspace', workspaceId }) : err(at(`${path}.workspace_id`, 'invalid'));
	}
	return err(at(`${path}.kind`, 'unknown'));
}

export function readValue(input: unknown, path = 'value'): Read<PreferenceValue> {
	if (!isRecord(input)) return err(at(path, 'not an object'));
	switch (input.kind) {
		case 'text':
			return typeof input.text === 'string' ? ok({ kind: 'text', text: input.text }) : err(at(`${path}.text`, 'missing'));
		case 'bool':
			return typeof input.bool === 'boolean' ? ok({ kind: 'bool', value: input.bool }) : err(at(`${path}.bool`, 'missing'));
		case 'int':
			return typeof input.int === 'string' && INT_PATTERN.test(input.int) ? ok({ kind: 'int', value: BigInt(input.int) }) : err(at(`${path}.int`, 'invalid'));
		default:
			return err(at(`${path}.kind`, 'unknown'));
	}
}

function readRevision(input: unknown, path: string): Read<string> {
	return typeof input === 'string' && REVISION_PATTERN.test(input) ? ok(input) : err(at(path, 'invalid'));
}

function readKey(input: unknown, path: string): Read<string> {
	return typeof input === 'string' && input.length > 0 ? ok(input) : err(at(path, 'missing'));
}

export function readEntry(input: unknown, path = 'entry'): Read<PreferenceEntry> {
	if (!isRecord(input)) return err(at(path, 'not an object'));
	const scope = readScope(input.scope, `${path}.scope`);
	if (!scope.ok) return scope;
	const key = readKey(input.key, `${path}.key`);
	if (!key.ok) return key;
	const value = readValue(input.value, `${path}.value`);
	if (!value.ok) return value;
	const revision = readRevision(input.revision, `${path}.revision`);
	if (!revision.ok) return revision;
	return ok({ scope: scope.value, key: key.value, value: value.value, revision: revision.value });
}

export function readEvent(input: unknown, path = 'event'): Read<PreferenceEvent> {
	if (!isRecord(input)) return err(at(path, 'not an object'));
	if (input.name !== 'preference.changed' && input.name !== 'preference.removed') return err(at(`${path}.name`, 'unknown'));
	const scope = readScope(input.scope, `${path}.scope`);
	if (!scope.ok) return scope;
	const key = readKey(input.key, `${path}.key`);
	if (!key.ok) return key;
	const revision = readRevision(input.revision, `${path}.revision`);
	if (!revision.ok) return revision;
	let value: PreferenceValue | null = null;
	if (input.value !== null && input.value !== undefined) {
		const decoded = readValue(input.value, `${path}.value`);
		if (!decoded.ok) return decoded;
		value = decoded.value;
	}
	return ok({ name: input.name, scope: scope.value, key: key.value, revision: revision.value, value });
}

export function readReadResponse(input: unknown): Read<ReadResponse> {
	if (!isRecord(input)) return err(at('value', 'not an object'));
	if (input.found === false) return ok({ found: false });
	if (input.found !== true) return err(at('value.found', 'invalid'));
	const entry = readEntry(input.entry, 'value.entry');
	return entry.ok ? ok({ found: true, entry: entry.value }) : entry;
}

export function readListResponse(input: unknown): Read<ListResponse> {
	if (!isRecord(input) || !Array.isArray(input.entries)) return err(at('value.entries', 'missing'));
	const entries: PreferenceEntry[] = [];
	for (const [index, raw] of input.entries.entries()) {
		const entry = readEntry(raw, `value.entries[${index}]`);
		if (!entry.ok) return entry;
		entries.push(entry.value);
	}
	return ok({ entries });
}

export function readReplaceResponse(input: unknown): Read<ReplaceResponse> {
	if (!isRecord(input)) return err(at('value', 'not an object'));
	if (input.status !== 'changed' && input.status !== 'unchanged') return err(at('value.status', 'unknown'));
	const entry = readEntry(input.entry, 'value.entry');
	if (!entry.ok) return entry;
	let event: PreferenceEvent | null = null;
	if (input.event !== null && input.event !== undefined) {
		const decoded = readEvent(input.event, 'value.event');
		if (!decoded.ok) return decoded;
		event = decoded.value;
	}
	return ok({ status: input.status, entry: entry.value, event });
}

export function readRemoveResponse(input: unknown): Read<RemoveResponse> {
	if (!isRecord(input)) return err(at('value', 'not an object'));
	if (input.removed === false) return ok({ removed: false });
	if (input.removed !== true) return err(at('value.removed', 'invalid'));
	const revision = readRevision(input.revision, 'value.revision');
	if (!revision.ok) return revision;
	const event = readEvent(input.event, 'value.event');
	if (!event.ok) return event;
	return ok({ removed: true, revision: revision.value, event: event.value });
}

export function readResolveResponse(input: unknown): Read<ResolveResponse> {
	if (!isRecord(input)) return err(at('value', 'not an object'));
	const value = readValue(input.value, 'value.value');
	if (!value.ok) return value;
	if (input.stored === false) return ok({ value: value.value, stored: false });
	if (input.stored !== true) return err(at('value.stored', 'invalid'));
	const revision = readRevision(input.revision, 'value.revision');
	if (!revision.ok) return revision;
	return ok({ value: value.value, stored: true, revision: revision.value });
}

/** Reads the projected failure; any malformed part falls back to the internal projection for the operation. */
export function readFailure(input: unknown, operation: PreferencesOperation): Failure {
	const fallback = internalFailure(WRITE_OPERATIONS.has(operation) ? 'unknown' : 'none');
	if (!isRecord(input)) return fallback;
	const kind = typeof input.kind === 'string' ? parseFailureKind(input.kind) : null;
	const commit = typeof input.commit === 'string' ? parseCommitState(input.commit) : null;
	if (!kind || !commit || typeof input.message !== 'string') return fallback;
	const type = typeof input.type === 'string' && input.type.length > 0 ? input.type : null;
	const fields: Record<string, string> = {};
	if (isRecord(input.fields)) for (const [path, problem] of Object.entries(input.fields)) if (typeof problem === 'string') fields[path] = problem;
	if (kind === 'internal') return internalFailure(commit);
	return failure(kind, input.message.length ? input.message : 'request failed', { type, fields, commit });
}

export function codecFailure(operation: PreferencesOperation, problem: Problem): Failure {
	return failure('internal', 'invalid response', { type: CODEC_INVALID_RESPONSE, fields: { response: problem }, commit: WRITE_OPERATIONS.has(operation) ? 'unknown' : 'none' });
}

/** Reads the outcome envelope. Success values go through the operation's reader; failures are projected. */
export function readOutcome<T>(input: unknown, operation: PreferencesOperation, readValueOf: (value: unknown) => Read<T>): Result<T, Failure> {
	if (!isRecord(input)) return err(codecFailure(operation, at('outcome', 'not an object')));
	if (input.ok === true) {
		const value = readValueOf(input.value);
		return value.ok ? value : err(codecFailure(operation, value.error));
	}
	if (input.ok === false) return err(readFailure(input.failure, operation));
	return err(codecFailure(operation, at('outcome.ok', 'invalid')));
}
