import { describe, expect, it } from 'vitest';
import { EMPTY_HISTORY, HISTORY_LIMIT, canGoBack, canGoForward, currentEntry, forgetView, goBack, goForward, recordVisit } from './history';

describe('navigation history', () => {
	it('records visits, moves back and forward, and truncates forward entries on a new visit', () => {
		let history = ['a', 'b', 'c'].reduce(recordVisit, EMPTY_HISTORY);
		expect(history.entries).toEqual(['a', 'b', 'c']);
		expect(canGoForward(history)).toBe(false);
		history = goBack(history);
		expect(currentEntry(history)).toBe('b');
		history = goBack(history);
		expect(currentEntry(history)).toBe('a');
		expect(canGoBack(history)).toBe(false);
		expect(goBack(history)).toBe(history);
		history = goForward(history);
		expect(currentEntry(history)).toBe('b');
		history = recordVisit(history, 'd');
		expect(history.entries).toEqual(['a', 'b', 'd']);
		expect(recordVisit(history, 'd')).toBe(history);
	});

	it('forgets closed views and repairs the index', () => {
		let history = ['a', 'b', 'a', 'c'].reduce(recordVisit, EMPTY_HISTORY);
		history = goBack(history);
		history = forgetView(history, 'b');
		expect(history.entries).toEqual(['a', 'c']);
		expect(currentEntry(history)).toBe('a');
		history = forgetView(history, 'a');
		expect(history).toEqual({ entries: ['c'], index: 0 });
		history = forgetView(history, 'c');
		expect(history).toEqual({ entries: [], index: -1 });
		expect(forgetView(history, 'zzz')).toBe(history);
	});

	it('caps the number of entries', () => {
		const history = Array.from({ length: HISTORY_LIMIT + 10 }, (_, i) => `v${i}`).reduce(recordVisit, EMPTY_HISTORY);
		expect(history.entries).toHaveLength(HISTORY_LIMIT);
		expect(currentEntry(history)).toBe(`v${HISTORY_LIMIT + 9}`);
	});
});
