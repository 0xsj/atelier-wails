# desktop

The Wails-host adapter. `preferences.ts` implements the service-owned
`PreferencesTransport` port by calling the bound `Preferences` methods `Read`
through `Resolve` with one document string; the outcome envelope comes back as
plain data and the codecs validate it. The binding is reached through the
runtime global Wails installs, so generated binding files are not a
type-check dependency. `workbench.ts` handles restoration and `workspace.ts`
bridges the Workspace registry methods. `host.ts` detects the Wails runtime for
root. No SDK type crosses a port; a rejected call is mapped by the service, not
here.
