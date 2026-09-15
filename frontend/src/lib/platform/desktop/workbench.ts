// Wails adapter for the workbench restoration document. The bound Go facade
// owns the config path and atomic file replacement; this module only bridges
// the JSON document to the consumer-owned SnapshotPort.

import { createJsonSnapshotPort, type SnapshotPort } from '../../workbench/app';

type BoundWorkbench = Readonly<Record<string, (...args: string[]) => Promise<unknown>>>;

function boundWorkbench(): BoundWorkbench | null {
	const go = (globalThis as { go?: { wailshost?: { Workbench?: BoundWorkbench } } }).go;
	return go?.wailshost?.Workbench ?? null;
}

export function createDesktopSnapshotPort(): SnapshotPort {
	return createJsonSnapshotPort({
		async load() {
			const method = boundWorkbench()?.Load;
			if (!method) throw new Error('Wails workbench binding is unavailable');
			return (await method()) as string | null;
		},
		async save(document) {
			const method = boundWorkbench()?.Save;
			if (!method) throw new Error('Wails workbench binding is unavailable');
			await method(document);
		}
	});
}
