// Closed failure vocabulary mirroring the native errors contract (revision 2)
// and the desktop wire contract's commit rule. Kinds are exact lowercase
// names; parsing never trims or folds case.

export const FAILURE_KINDS = ['unauthenticated', 'forbidden', 'rate_limited', 'unavailable', 'timeout', 'canceled', 'internal', 'not_found', 'invalid', 'conflict'] as const;
export type FailureKind = (typeof FAILURE_KINDS)[number];

export const COMMIT_STATES = ['none', 'not_applied', 'unknown'] as const;
/** Whether a request's write was applied: never attempted, refused without effect, or possibly committed. */
export type CommitState = (typeof COMMIT_STATES)[number];

export interface Failure {
	readonly kind: FailureKind;
	/** Public-safe text; internal and unclassified failures carry exactly INTERNAL_MESSAGE. */
	readonly message: string;
	/** Module-owned condition identifier such as `preferences.conflict`, or null. */
	readonly type: string | null;
	/** Public field problems keyed by request path; never the offending value. */
	readonly fields: Readonly<Record<string, string>>;
	readonly commit: CommitState;
}

export const INTERNAL_MESSAGE = 'internal error';

export function parseFailureKind(text: string): FailureKind | null {
	return (FAILURE_KINDS as readonly string[]).includes(text) ? (text as FailureKind) : null;
}

export function parseCommitState(text: string): CommitState | null {
	return (COMMIT_STATES as readonly string[]).includes(text) ? (text as CommitState) : null;
}

export function failure(kind: FailureKind, message: string, options: { type?: string | null; fields?: Readonly<Record<string, string>>; commit?: CommitState } = {}): Failure {
	return { kind, message, type: options.type ?? null, fields: { ...(options.fields ?? {}) }, commit: options.commit ?? 'none' };
}

/** The fallback projection: no type, no fields, fixed message. */
export function internalFailure(commit: CommitState = 'none'): Failure {
	return { kind: 'internal', message: INTERNAL_MESSAGE, type: null, fields: {}, commit };
}

export function isKind(value: Failure, ...kinds: readonly FailureKind[]): boolean {
	return kinds.includes(value.kind);
}

export function hasType(value: Failure, type: string): boolean {
	return value.type === type;
}

/** True when the write may have committed and the caller must reconcile before retrying. */
export function mayHaveCommitted(value: Failure): boolean {
	return value.commit === 'unknown';
}

/** True when the request was refused before any effect. */
export function wasRefused(value: Failure): boolean {
	return value.commit === 'not_applied';
}

export function fieldProblems(value: Failure): readonly (readonly [path: string, problem: string])[] {
	return Object.entries(value.fields).sort(([a], [b]) => a.localeCompare(b));
}
