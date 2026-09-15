// Restoration through an injected snapshot port. The port returns plain data;
// the model's total reader validates it. Saving is debounced by the host.

import { readSnapshot, type RestoreOutcome, type WorkbenchSnapshot } from '../../model/layout/snapshot';

export interface SnapshotPort {
	load(): Promise<unknown>;
	save(snapshot: WorkbenchSnapshot): Promise<void>;
}

/** The document-shaped boundary used by native adapters. */
export interface SnapshotDocumentPort {
	load(): Promise<string | null>;
	save(document: string): Promise<void>;
}

/** Turns a native JSON document store into the workbench's unknown-valued port. */
export function createJsonSnapshotPort(port: SnapshotDocumentPort): SnapshotPort {
	return {
		async load() {
			const document = await port.load();
			if (document === null) return null;
			try {
				return JSON.parse(document) as unknown;
			} catch {
				// Keep malformed text as an unknown value so loadSnapshot reports the
				// model's normal invalid outcome instead of an unavailable port.
				return document;
			}
		},
		async save(snapshot) {
			await port.save(JSON.stringify(snapshot));
		}
	};
}

export type LoadOutcome = RestoreOutcome | { readonly kind: 'empty' } | { readonly kind: 'unavailable'; readonly error: unknown };

export async function loadSnapshot(port: SnapshotPort): Promise<LoadOutcome> {
	let raw: unknown;
	try {
		raw = await port.load();
	} catch (error) {
		return { kind: 'unavailable', error };
	}
	if (raw === null || raw === undefined) return { kind: 'empty' };
	return readSnapshot(raw);
}

/** In-memory port for previews and tests. Stores a JSON copy so callers cannot share references. */
export function createMemorySnapshotPort(initial: unknown = null): SnapshotPort & { readonly stored: () => unknown } {
	let stored: unknown = initial === null ? null : JSON.parse(JSON.stringify(initial));
	return {
		async load() {
			return stored;
		},
		async save(snapshot) {
			stored = JSON.parse(JSON.stringify(snapshot));
		},
		stored: () => stored
	};
}

export interface Scheduler {
	setTimeout(callback: () => void, delay: number): unknown;
	clearTimeout(handle: unknown): void;
}

export interface DebouncedSaver {
	request(snapshot: () => WorkbenchSnapshot): void;
	flush(): Promise<void>;
	dispose(): void;
}

/** Coalesces save requests; the last requested snapshot wins. Errors are reported, never thrown. */
export function createDebouncedSaver(port: SnapshotPort, delay: number, onError: (error: unknown) => void = () => {}, scheduler: Scheduler = globalThis): DebouncedSaver {
	let handle: unknown = null;
	let pending: (() => WorkbenchSnapshot) | null = null;
	async function run(): Promise<void> {
		const take = pending;
		pending = null;
		handle = null;
		if (!take) return;
		try {
			await port.save(take());
		} catch (error) {
			onError(error);
		}
	}
	return {
		request(snapshot) {
			pending = snapshot;
			if (handle !== null) scheduler.clearTimeout(handle);
			handle = scheduler.setTimeout(() => void run(), delay);
		},
		async flush() {
			if (handle !== null) scheduler.clearTimeout(handle);
			await run();
		},
		dispose() {
			if (handle !== null) scheduler.clearTimeout(handle);
			handle = null;
			pending = null;
		}
	};
}
