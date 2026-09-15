# preview

Explicit browser adapters for development. `preferences.ts` is an in-memory
Preferences transport that honors the desktop wire contract: the same request
decoding vocabulary and paths, compare-and-replace and revision rules, the
outcome envelope, the failure projection and the commit rule, plus injectable
W11 faults so reconciliation can be exercised. `workspace.ts` is the matching
in-memory registry double for Workspace metadata and lifecycle transitions.
Tests pin both adapters to their contract fixtures and scenarios. These are
development doubles; they do not stand in for native-domain evidence, and
anything they cannot honor must be reported as a failure rather than faked.
