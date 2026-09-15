# storetest

Reusable logical scenario suite (M01–M12) and deterministic fault wrappers
for preference store adapters. Imported only by adapter and application
tests; never a production dependency. Specified in
[../memory/CONTRACT.md](../memory/CONTRACT.md). A persistent adapter reruns
`Run` unchanged and adds its own restart and corruption evidence.
