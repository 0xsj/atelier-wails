// Launch probes for native verification builds only. When a build sets
// VITE_ATELIER_PROBE=preferences or workbench, the gallery writes one value at
// startup and reads it back, so the native document on disk proves the
// JavaScript-to-native round trip without screen or accessibility access.
// Vite inlines the flag, so normal builds compile this call away.

import { GLOBAL_SCOPE, readPreference, setPreference, text, type PreferencesTransport } from '../lib/services/preferences';
import { loadSnapshot, type SnapshotPort } from '../lib/workbench/app';
import { createLayoutState, createSnapshot } from '../lib/workbench/model';
import { createViewsState, openView } from '../lib/workbench/model/views/views';

export const PROBE_KEY = 'probe.launch';

export async function runLaunchProbe(transport: PreferencesTransport, host: string): Promise<string> {
	const stamp = `${host} ${new Date().toISOString()}`;
	const written = await setPreference(transport, GLOBAL_SCOPE, PROBE_KEY, text(stamp));
	if (!written.ok) return `probe write failed: ${written.error.kind} · commit ${written.error.commit}`;
	const read = await readPreference(transport, GLOBAL_SCOPE, PROBE_KEY);
	if (!read.ok) return `probe read failed: ${read.error.kind}`;
	if (!read.value.present || read.value.value.value.kind !== 'text' || read.value.value.value.text !== stamp) return 'probe read back a different value';
	return `probe ${written.value.status} at revision ${written.value.entry.revision}`;
}

const SNAPSHOT_PROBE_VIEW = 'probe.snapshot';

export async function runSnapshotProbe(port: SnapshotPort, host: string): Promise<string> {
	const views = openView(createViewsState(), { id: SNAPSHOT_PROBE_VIEW, kind: 'file', title: `${host} snapshot probe` });
	await port.save(createSnapshot(createLayoutState(), views));
	const outcome = await loadSnapshot(port);
	if (outcome.kind !== 'restored') return `snapshot probe failed: ${outcome.kind}`;
	if (!outcome.views.views[SNAPSHOT_PROBE_VIEW]) return 'snapshot probe did not restore its view';
	return `snapshot probe restored ${SNAPSHOT_PROBE_VIEW}`;
}
