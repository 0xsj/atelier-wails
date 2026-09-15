# shell

`WorkbenchShell` composes a title area, optional activity rail, optional sidebar,
main region and optional status area using named snippets, and exposes window
focus through `windowActive` so chrome can dim when the window is inactive.
`TitleBar` replaces the default title area with a centered title, leading and
trailing controls, a platform inset for native window controls, and drag-region
attributes supplied by the root, so this component never names a native host.
`ActivityRail` and `ActivityItem` provide the vertical rail of selectable
activities with badges and arrow-key movement; the root decides what each
activity opens.

Rendered browser coverage verifies the shell landmarks and the activity rail's
click and arrow-key selection behavior.
