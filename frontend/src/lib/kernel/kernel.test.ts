import { describe, expect, it } from 'vitest';
import { FAILURE_KINDS, INTERNAL_MESSAGE, failure, fieldProblems, internalFailure, mayHaveCommitted, parseCommitState, parseFailureKind, wasRefused } from './failure';
import { isWorkspaceId, parseWorkspaceId } from './identity';
import { absent, found, mapPresence, presenceOr } from './presence';
import { all, andThen, err, map, mapError, match, ok, unwrapOr } from './result';

describe('result', () => {
	it('maps, chains and matches without throwing', () => {
		const doubled = map(ok(2), (n) => n * 2);
		expect(doubled).toEqual({ ok: true, value: 4 });
		expect(andThen(ok(2), (n) => (n > 1 ? err('too big') : ok(n)))).toEqual({ ok: false, error: 'too big' });
		expect(mapError(err('a'), (e) => e.toUpperCase())).toEqual({ ok: false, error: 'A' });
		expect(match(err('x'), { ok: () => 'ok', err: (e) => `err:${e}` })).toBe('err:x');
		expect(unwrapOr(err('x') as ReturnType<typeof err<string>> | ReturnType<typeof ok<number>>, 7)).toBe(7);
		expect(all([ok(1), ok(2)])).toEqual({ ok: true, value: [1, 2] });
		expect(all([ok(1), err('stop'), err('later')])).toEqual({ ok: false, error: 'stop' });
	});
});

describe('failure', () => {
	it('parses exactly the ten kinds without trimming or folding', () => {
		expect(FAILURE_KINDS).toHaveLength(10);
		for (const kind of FAILURE_KINDS) expect(parseFailureKind(kind)).toBe(kind);
		expect(parseFailureKind('Invalid')).toBeNull();
		expect(parseFailureKind(' invalid')).toBeNull();
		expect(parseFailureKind('')).toBeNull();
		expect(parseCommitState('not_applied')).toBe('not_applied');
		expect(parseCommitState('applied')).toBeNull();
	});

	it('builds the internal fallback and copies fields', () => {
		expect(internalFailure('unknown')).toEqual({ kind: 'internal', message: INTERNAL_MESSAGE, type: null, fields: {}, commit: 'unknown' });
		const fields = { 'value.int': 'invalid', key: 'missing' };
		const refused = failure('invalid', 'invalid request', { type: 'desktop.invalid_request', fields, commit: 'not_applied' });
		expect(refused.fields).not.toBe(fields);
		expect(fieldProblems(refused)).toEqual([['key', 'missing'], ['value.int', 'invalid']]);
		expect(wasRefused(refused)).toBe(true);
		expect(mayHaveCommitted(refused)).toBe(false);
		expect(mayHaveCommitted(failure('timeout', 'timed out', { commit: 'unknown' }))).toBe(true);
	});
});

describe('presence', () => {
	it('keeps found and absent distinct and maps only found', () => {
		expect(found(1)).toEqual({ present: true, value: 1 });
		expect(absent()).toBe(absent());
		expect(mapPresence(found(1), (n) => n + 1)).toEqual({ present: true, value: 2 });
		expect(mapPresence(absent(), (n: number) => n + 1)).toEqual({ present: false });
		expect(presenceOr(absent(), 'fallback')).toBe('fallback');
	});
});

describe('identity', () => {
	it('parses canonical UUID text and normalizes case', () => {
		expect(parseWorkspaceId('01900000-0000-7000-8000-000000000001')).toBe('01900000-0000-7000-8000-000000000001');
		expect(parseWorkspaceId('01900000-0000-7000-8000-00000000000A')).toBe('01900000-0000-7000-8000-00000000000a');
		expect(parseWorkspaceId('not-a-uuid')).toBeNull();
		expect(parseWorkspaceId('00000000-0000-0000-0000-000000000000')).toBeNull();
		expect(parseWorkspaceId(42)).toBeNull();
		expect(isWorkspaceId('01900000-0000-7000-8000-00000000000A')).toBe(false);
		expect(isWorkspaceId('01900000-0000-7000-8000-00000000000a')).toBe(true);
	});
});
