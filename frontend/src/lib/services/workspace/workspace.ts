// Workspace application operations over a consumer-owned raw transport. The
// codec validates every response; this service turns a successful missing
// record into Presence.absent instead of conflating it with a failure.

import { absent, err, found, internalFailure, ok, type Failure, type Lookup, type Result, type WorkspaceId } from '../../kernel';
import {
	encodeRequest,
	readOutcome,
	readForgetResponse,
	readListResponse,
	readMutationResponse,
	readReadResponse,
	readRegisteredResponse,
} from '../../platform/codecs/workspace';
import { WORKSPACE_WRITE_OPERATIONS, type WorkspaceExpected, type WorkspaceFilter, type WorkspaceForgotten, type WorkspaceMutation, type WorkspaceOperation, type WorkspaceRegistered, type WorkspaceSnapshot } from './model';

export interface WorkspaceTransport {
	/** Sends a plain JSON request and resolves with the raw result envelope. */
	call(operation: WorkspaceOperation, document: string): Promise<unknown>;
}

async function invoke<T>(transport: WorkspaceTransport, operation: WorkspaceOperation, document: string, reader: (value: unknown) => Result<T, string>): Promise<Result<T, Failure>> {
	try {
		return readOutcome(await transport.call(operation, document), operation, reader);
	} catch {
		return err(internalFailure(WORKSPACE_WRITE_OPERATIONS.has(operation) ? 'unknown' : 'none'));
	}
}

export function readWorkspace(transport: WorkspaceTransport, id: WorkspaceId): Promise<Lookup<WorkspaceSnapshot, Failure>> {
	return invoke(transport, 'read', encodeRequest('read', { id }), readReadResponse).then((result) => (result.ok ? ok(result.value.found ? found(result.value.workspace) : absent()) : result));
}

export function listWorkspaces(transport: WorkspaceTransport, filter: WorkspaceFilter = 'active'): Promise<Result<readonly WorkspaceSnapshot[], Failure>> {
	return invoke(transport, 'list', encodeRequest('list', { filter }), readListResponse).then((result) => (result.ok ? ok(result.value.workspaces) : result));
}

export function registerWorkspace(transport: WorkspaceTransport, id: WorkspaceId, name: string, location: string, at: string): Promise<Result<WorkspaceRegistered, Failure>> {
	return invoke(transport, 'register', encodeRequest('register', { id, name, location, at }), readRegisteredResponse);
}

export function renameWorkspace(transport: WorkspaceTransport, id: WorkspaceId, name: string, expected: WorkspaceExpected, at: string): Promise<Result<WorkspaceMutation, Failure>> {
	return invoke(transport, 'rename', encodeRequest('rename', { id, name, expected, at }), readMutationResponse);
}

export function archiveWorkspace(transport: WorkspaceTransport, id: WorkspaceId, expected: WorkspaceExpected, at: string): Promise<Result<WorkspaceMutation, Failure>> {
	return invoke(transport, 'archive', encodeRequest('archive', { id, expected, at }), readMutationResponse);
}

export function restoreWorkspace(transport: WorkspaceTransport, id: WorkspaceId, expected: WorkspaceExpected, at: string): Promise<Result<WorkspaceMutation, Failure>> {
	return invoke(transport, 'restore', encodeRequest('restore', { id, expected, at }), readMutationResponse);
}

export function forgetWorkspace(transport: WorkspaceTransport, id: WorkspaceId, expected: WorkspaceExpected): Promise<Result<WorkspaceForgotten | { readonly absent: true }, Failure>> {
	return invoke(transport, 'forget', encodeRequest('forget', { id, expected }), readForgetResponse).then((result) => (result.ok ? ok(result.value.found ? result.value.workspace : { absent: true }) : result));
}
