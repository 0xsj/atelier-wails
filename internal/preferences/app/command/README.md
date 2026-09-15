# command

Preferences bounded context. Mutations (replace, remove) and the ports they
consume: a conditional store and an in-process event publisher.

Implemented against [../CONTRACT.md](../CONTRACT.md) revision 1 (task
preferences-app). Scenario tests live in spec_test.go and run over the memory
adapter draft with a recording publisher and a fault store.
