# workspace

The Workspace application boundary owns registry metadata operations over a
consumer-owned raw transport. Reads return `Lookup<WorkspaceSnapshot, Failure>`
so a successful absent record stays distinct from an unavailable transport.
The codec keeps IDs, lifecycle status, UTC timestamps and decimal revisions
explicit. `createWorkspaceOpenService` re-reads a selected record, rejects
archived or absent records, and returns an active Workbench session without
inspecting files or orchestrating windows.
