// Detects whether the frontend runs inside the Wails host. Root uses this to
// choose native adapters over preview adapters; nothing below root reads it.

export type HostKind = 'wails' | 'browser';

export function detectHost(): HostKind {
	const scope = globalThis as { go?: unknown; runtime?: unknown };
	return typeof scope.go === 'object' && scope.go !== null ? 'wails' : 'browser';
}
