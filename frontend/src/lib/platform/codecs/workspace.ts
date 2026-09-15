// Plain JSON request/response mapping for the Workspace registry. Native
// transport is intentionally not installed yet; this codec fixes the shape
// that the preview adapter and future desktop facade will share.

import { err, failure, internalFailure, ok, parseCommitState, parseFailureKind, parseWorkspaceId, type Failure, type Result, type WorkspaceId } from '../../kernel';
import { isCanonicalWorkspaceLocation, isValidWorkspaceName, isValidWorkspaceRevision, WORKSPACE_WRITE_OPERATIONS, type WorkspaceEvent, type WorkspaceExpected, type WorkspaceFilter, type WorkspaceForgotten, type WorkspaceMutation, type WorkspaceOperation, type WorkspaceSnapshot, type WorkspaceStatus } from '../../services/workspace/model';

export type WireExpected = { readonly kind: 'absent' } | { readonly kind: 'revision'; readonly revision: string };
export type WireWorkspace = { readonly id: string; readonly name: string; readonly location: string; readonly status: WorkspaceStatus; readonly revision: string; readonly created_at: string; readonly updated_at: string };
export type WireEvent = { readonly name: WorkspaceEvent['name']; readonly workspace_id: string; readonly revision: string; readonly title: string; readonly location: string; readonly status: WorkspaceStatus };

export interface ReadRequest { readonly id: WorkspaceId }
export interface ListRequest { readonly filter: WorkspaceFilter }
export interface RegisterRequest { readonly id: WorkspaceId; readonly name: string; readonly location: string; readonly at: string }
export interface RenameRequest { readonly id: WorkspaceId; readonly name: string; readonly expected: WorkspaceExpected; readonly at: string }
export interface LifecycleRequest { readonly id: WorkspaceId; readonly expected: WorkspaceExpected; readonly at?: string }

export type ReadResponse = { readonly found: true; readonly workspace: WorkspaceSnapshot } | { readonly found: false };
export type ListResponse = { readonly workspaces: readonly WorkspaceSnapshot[] };
export type RegisteredResponse = { readonly workspace: WorkspaceSnapshot; readonly event: WorkspaceEvent };
export type MutationResponse = WorkspaceMutation;
export type ForgetResponse = { readonly found: true; readonly workspace: WorkspaceForgotten } | { readonly found: false };

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function problem(path: string, message: string): string {
	return `${path}: ${message}`;
}

function readId(value: unknown, path: string): Result<WorkspaceId, string> {
	const id = parseWorkspaceId(value);
	return id ? ok(id) : err(problem(path, 'invalid'));
}

function readStatus(value: unknown, path: string): Result<WorkspaceStatus, string> {
	return value === 'active' || value === 'archived' ? ok(value) : err(problem(path, 'unknown'));
}

function readTimestamp(value: unknown, path: string): Result<string, string> {
	return typeof value === 'string' && value.length > 0 && Number.isFinite(Date.parse(value)) ? ok(value) : err(problem(path, 'invalid'));
}

export function encodeExpected(expected: WorkspaceExpected): WireExpected {
	return expected.kind === 'absent' ? { kind: 'absent' } : { kind: 'revision', revision: expected.revision };
}

export function encodeRequest(operation: 'read', request: ReadRequest): string;
export function encodeRequest(operation: 'list', request: ListRequest): string;
export function encodeRequest(operation: 'register', request: RegisterRequest): string;
export function encodeRequest(operation: 'rename', request: RenameRequest): string;
export function encodeRequest(operation: 'archive' | 'restore', request: LifecycleRequest): string;
export function encodeRequest(operation: 'forget', request: LifecycleRequest): string;
export function encodeRequest(operation: string, request: ReadRequest | ListRequest | RegisterRequest | RenameRequest | LifecycleRequest): string {
	switch (operation) {
		case 'read':
			return JSON.stringify({ id: (request as ReadRequest).id });
		case 'list':
			return JSON.stringify({ filter: (request as ListRequest).filter });
		case 'register': {
			const value = request as RegisterRequest;
			return JSON.stringify({ id: value.id, name: value.name, location: value.location, at: value.at });
		}
		case 'rename': {
			const value = request as RenameRequest;
			return JSON.stringify({ id: value.id, name: value.name, expected: encodeExpected(value.expected), at: value.at });
		}
		case 'archive':
		case 'restore': {
			const value = request as LifecycleRequest;
			return JSON.stringify({ id: value.id, expected: encodeExpected(value.expected), at: value.at });
		}
		case 'forget': {
			const value = request as LifecycleRequest;
			return JSON.stringify({ id: value.id, expected: encodeExpected(value.expected) });
		}
		default:
			return JSON.stringify(request);
	}
}

export function readExpected(value: unknown, path = 'expected'): Result<WorkspaceExpected, string> {
	if (!isRecord(value)) return err(problem(path, 'not an object'));
	if (value.kind === 'absent') return ok({ kind: 'absent' });
	if (value.kind !== 'revision' || typeof value.revision !== 'string' || !isValidWorkspaceRevision(value.revision)) return err(problem(`${path}.revision`, 'invalid'));
	return ok({ kind: 'revision', revision: value.revision });
}

export function readWorkspace(value: unknown, path = 'workspace'): Result<WorkspaceSnapshot, string> {
	if (!isRecord(value)) return err(problem(path, 'not an object'));
	const id = readId(value.id, `${path}.id`);
	if (!id.ok) return id;
	if (typeof value.name !== 'string' || !isValidWorkspaceName(value.name)) return err(problem(`${path}.name`, 'invalid'));
	if (typeof value.location !== 'string' || !isCanonicalWorkspaceLocation(value.location)) return err(problem(`${path}.location`, 'invalid'));
	const status = readStatus(value.status, `${path}.status`);
	if (!status.ok) return status;
	if (typeof value.revision !== 'string' || !isValidWorkspaceRevision(value.revision)) return err(problem(`${path}.revision`, 'invalid'));
	const createdAt = readTimestamp(value.created_at, `${path}.created_at`);
	if (!createdAt.ok) return createdAt;
	const updatedAt = readTimestamp(value.updated_at, `${path}.updated_at`);
	if (!updatedAt.ok) return updatedAt;
	return ok({ id: id.value, name: value.name, location: value.location, status: status.value, revision: value.revision, createdAt: createdAt.value, updatedAt: updatedAt.value });
}

export function readEvent(value: unknown, path = 'event'): Result<WorkspaceEvent, string> {
	if (!isRecord(value)) return err(problem(path, 'not an object'));
	const id = readId(value.workspace_id, `${path}.workspace_id`);
	if (!id.ok) return id;
	if (typeof value.name !== 'string' || !['workspace.registered', 'workspace.renamed', 'workspace.archived', 'workspace.restored', 'workspace.forgotten'].includes(value.name)) return err(problem(`${path}.name`, 'unknown'));
	if (typeof value.revision !== 'string' || !isValidWorkspaceRevision(value.revision)) return err(problem(`${path}.revision`, 'invalid'));
	if (typeof value.title !== 'string' || !isValidWorkspaceName(value.title)) return err(problem(`${path}.title`, 'invalid'));
	if (typeof value.location !== 'string' || !isCanonicalWorkspaceLocation(value.location)) return err(problem(`${path}.location`, 'invalid'));
	const status = readStatus(value.status, `${path}.status`);
	if (!status.ok) return status;
	return ok({ name: value.name as WorkspaceEvent['name'], workspaceId: id.value, revision: value.revision, title: value.title, location: value.location, status: status.value });
}

export function readReadResponse(value: unknown): Result<ReadResponse, string> {
	if (!isRecord(value)) return err(problem('value', 'not an object'));
	if (value.found === false) return ok({ found: false });
	if (value.found !== true) return err(problem('value.found', 'invalid'));
	const workspace = readWorkspace(value.workspace, 'value.workspace');
	return workspace.ok ? ok({ found: true, workspace: workspace.value }) : workspace;
}

export function readListResponse(value: unknown): Result<ListResponse, string> {
	if (!isRecord(value) || !Array.isArray(value.workspaces)) return err(problem('value.workspaces', 'missing'));
	const workspaces: WorkspaceSnapshot[] = [];
	for (const [index, item] of value.workspaces.entries()) {
		const workspace = readWorkspace(item, `value.workspaces[${index}]`);
		if (!workspace.ok) return workspace;
		workspaces.push(workspace.value);
	}
	return ok({ workspaces });
}

export function readRegisteredResponse(value: unknown): Result<RegisteredResponse, string> {
	if (!isRecord(value)) return err(problem('value', 'not an object'));
	const workspace = readWorkspace(value.workspace, 'value.workspace');
	if (!workspace.ok) return workspace;
	const event = readEvent(value.event, 'value.event');
	return event.ok ? ok({ workspace: workspace.value, event: event.value }) : event;
}

export function readMutationResponse(value: unknown): Result<MutationResponse, string> {
	if (!isRecord(value) || (value.status !== 'changed' && value.status !== 'unchanged')) return err(problem('value.status', 'unknown'));
	const workspace = readWorkspace(value.workspace, 'value.workspace');
	if (!workspace.ok) return workspace;
	let event: WorkspaceEvent | null = null;
	if (value.event !== null && value.event !== undefined) {
		const decoded = readEvent(value.event, 'value.event');
		if (!decoded.ok) return decoded;
		event = decoded.value;
	}
	return ok({ status: value.status, workspace: workspace.value, event });
}

export function readForgetResponse(value: unknown): Result<ForgetResponse, string> {
	if (!isRecord(value)) return err(problem('value', 'not an object'));
	if (value.found === false) return ok({ found: false });
	if (value.found !== true) return err(problem('value.found', 'invalid'));
	if (!isRecord(value.workspace) || typeof value.workspace.revision !== 'string' || !isValidWorkspaceRevision(value.workspace.revision)) return err(problem('value.workspace.revision', 'invalid'));
	const event = readEvent(value.event, 'value.event');
	if (!event.ok) return event;
	return ok({ found: true, workspace: { revision: value.workspace.revision, event: event.value } });
}

/** Reads a projected failure; malformed failures fall back to the operation's internal projection. */
export function readFailure(value: unknown, operation: WorkspaceOperation): Failure {
	const fallback = internalFailure(WORKSPACE_WRITE_OPERATIONS.has(operation) ? 'unknown' : 'none');
	if (!isRecord(value)) return fallback;
	const kind = typeof value.kind === 'string' ? parseFailureKind(value.kind) : null;
	const commit = typeof value.commit === 'string' ? parseCommitState(value.commit) : null;
	if (!kind || !commit || typeof value.message !== 'string') return fallback;
	const type = typeof value.type === 'string' && value.type.length > 0 ? value.type : null;
	const fields: Record<string, string> = {};
	if (isRecord(value.fields)) for (const [path, problem] of Object.entries(value.fields)) if (typeof problem === 'string') fields[path] = problem;
	if (kind === 'internal') return internalFailure(commit);
	return failure(kind, value.message.length > 0 ? value.message : 'request failed', { type, fields, commit });
}

export function codecFailure(operation: string, response: string): Failure {
	return failure('internal', 'invalid response', { type: 'codec.invalid_response', fields: { response }, commit: operation === 'read' || operation === 'list' ? 'none' : 'unknown' });
}

export function readOutcome<T>(input: unknown, operation: WorkspaceOperation, reader: (value: unknown) => Result<T, string>): Result<T, Failure> {
	if (!isRecord(input)) return err(codecFailure(operation, 'outcome: not an object'));
	if (input.ok === true) {
		const value = reader(input.value);
		return value.ok ? value : err(codecFailure(operation, value.error));
	}
	if (input.ok === false) return err(readFailure(input.failure, operation));
	return err(codecFailure(operation, 'outcome.ok: invalid'));
}
