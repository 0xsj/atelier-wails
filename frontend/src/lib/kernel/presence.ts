// A lookup that succeeded may still find nothing. Found, absent and
// unsuccessful stay distinct: a lookup is Result<Presence<T>, Failure>.

import type { Result } from './result';

export interface Found<T> {
	readonly present: true;
	readonly value: T;
}

export interface Absent {
	readonly present: false;
}

export type Presence<T> = Found<T> | Absent;

const ABSENT: Absent = Object.freeze({ present: false });

export function found<T>(value: T): Found<T> {
	return { present: true, value };
}

export function absent(): Absent {
	return ABSENT;
}

export function isFound<T>(presence: Presence<T>): presence is Found<T> {
	return presence.present;
}

export function mapPresence<T, U>(presence: Presence<T>, transform: (value: T) => U): Presence<U> {
	return presence.present ? found(transform(presence.value)) : presence;
}

export function presenceOr<T>(presence: Presence<T>, fallback: T): T {
	return presence.present ? presence.value : fallback;
}

export type Lookup<T, E> = Result<Presence<T>, E>;
