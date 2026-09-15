// Typed identifiers, introduced only when a consuming operation needs them.
// Preferences scopes reference a workspace by its canonical UUID text.

export type WorkspaceId = string & { readonly __brand: 'WorkspaceId' };

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

/** Accepts canonical UUID text in any case and returns the lowercase form; anything else is null. */
export function parseWorkspaceId(text: unknown): WorkspaceId | null {
	if (typeof text !== 'string') return null;
	const lower = text.toLowerCase();
	if (!UUID.test(lower) || /^0{8}-0{4}-0{4}-0{4}-0{12}$/.test(lower)) return null;
	return lower as WorkspaceId;
}

export function isWorkspaceId(text: unknown): text is WorkspaceId {
	return parseWorkspaceId(text) === text;
}
