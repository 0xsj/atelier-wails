# workspace

`WorkspacePicker` lists the registry's active or archived metadata, reads the
selected record through the `Presence` boundary, and reports the runtime Open
action to its consumer. It never inspects or mutates project files. Root
supplies the transport; the current adapter is the in-memory browser preview
registry while native Workspace application and transport remain reserved.
