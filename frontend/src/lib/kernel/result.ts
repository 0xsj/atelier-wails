// Explicit expected outcomes. An Err is a value, never a thrown exception.

export interface Ok<T> {
	readonly ok: true;
	readonly value: T;
}

export interface Err<E> {
	readonly ok: false;
	readonly error: E;
}

export type Result<T, E> = Ok<T> | Err<E>;

export function ok<T>(value: T): Ok<T> {
	return { ok: true, value };
}

export function err<E>(error: E): Err<E> {
	return { ok: false, error };
}

export function isOk<T, E>(result: Result<T, E>): result is Ok<T> {
	return result.ok;
}

export function isErr<T, E>(result: Result<T, E>): result is Err<E> {
	return !result.ok;
}

export function map<T, U, E>(result: Result<T, E>, transform: (value: T) => U): Result<U, E> {
	return result.ok ? ok(transform(result.value)) : result;
}

export function mapError<T, E, F>(result: Result<T, E>, transform: (error: E) => F): Result<T, F> {
	return result.ok ? result : err(transform(result.error));
}

export function andThen<T, U, E>(result: Result<T, E>, next: (value: T) => Result<U, E>): Result<U, E> {
	return result.ok ? next(result.value) : result;
}

export function match<T, E, R>(result: Result<T, E>, handlers: { ok: (value: T) => R; err: (error: E) => R }): R {
	return result.ok ? handlers.ok(result.value) : handlers.err(result.error);
}

export function unwrapOr<T, E>(result: Result<T, E>, fallback: T): T {
	return result.ok ? result.value : fallback;
}

/** Collects results in order; the first Err wins. */
export function all<T, E>(results: readonly Result<T, E>[]): Result<T[], E> {
	const values: T[] = [];
	for (const result of results) {
		if (!result.ok) return result;
		values.push(result.value);
	}
	return ok(values);
}
