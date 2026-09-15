# Implementing with workers

The setup in this repository defines bounded implementation work. It does not
start agents automatically or choose a provider/model. Each Atelier project is
self-contained. Follow [DOMAIN_GUIDE.md](DOMAIN_GUIDE.md) for domain development.

## Dispatch

`work/manifest.json` is the dependency/build manifest. `ready` means specified and
eligible for assignment when its dependencies are complete, not implemented.
`waiting` has a specified task but unfinished prerequisites. `planned` is not
assignable: it needs a contract/task or another design decision. A coordinator
records the assigned worker and changes ready → in_progress before dispatch.

Read [work/BUILD_ORDER.md](work/BUILD_ORDER.md) and the resume
instructions in [work/STATUS.md](work/STATUS.md). The first errors, clock and secret
leaves have no peer dependencies and disjoint files. Their implementation can
run independently; their integrations and root wiring stay with the coordinator.
IDs, config, logger and file IO have since been specified and completed. A directory's existence
is not permission to invent its semantics. These small leaves are a calibration
batch; preferences is the first complete domain slice after its contract is set.

Use [templates/MODULE_CONTRACT.md](templates/MODULE_CONTRACT.md) and
[templates/WORKER_TASK.md](templates/WORKER_TASK.md) for additional work. Hand workers
the task file, contract revision, repository root, and specific baseline revision
or file inventory. Do not describe the whole architecture as one implementation task.

## Ownership and changes

Prefer isolated checkouts when available. In a shared checkout, assign nonoverlapping
files and prohibit cleanup of another worker's changes. Root/module wiring,
manifests, dependency files and lockfiles have one coordinator owner. Workers may
choose private implementation mechanics; public API changes require a revised
contract and dependent-task review by the coordinator. Routine coordination does
not require repeatedly asking the user for permission.

Workers may add tests, never delete or weaken specified scenarios to get green.
Report contradictions with the exact scenario and a proposed correction. Update
contract revision and task references together when the coordinator resolves it.
Do not claim an implementation-visible test suite is independent verification.

## Evidence and integration

Local green results move work to `implemented`, not `complete`. Save a handoff under
`work/handoffs/<id>.md` with revision, scenario-to-test mapping, files and exact
commands/results/test counts. A separate coordinator review checks omissions,
forbidden imports, unexpected effects, error preservation and ownership behavior.
Review can be done by the coordinating agent; no extra agent is mandatory.

Integrate one change set at a time. Check actual changed files against allowed
paths, run focused checks, then the task's integration gate. Do not merge by
matching summaries alone. Record integration evidence before marking complete.
If a dependency changes its API or behavior, re-evaluate dependent evidence.

The manifest verifier validates metadata, references, dependency cycles and task
readiness. It does not execute tests, authenticate evidence or automatically approve
completion. `python3 tools/architecture/check_imports.py` enforces the layer
import rules in tools/architecture/RULES.md; run both before marking complete.

## First-batch limits

Error kinds are native classification values, not a finalized wire protocol.
No automatic retry, authentication subsystem, secret vault or scheduler
is selected by these leaves. The later ID contract selects UUIDv7 generation. Clock and secret have scheduled adapter/process
consumers, not current domain consumers. Do not force them into preference APIs
merely because they have been implemented. The app must retain its working Svelte
Hello world screen throughout this batch.

Task-file state labels describe their initial assignment state. The manifest is
the current status authority. Contracts remain revisioned independently of status.

The later provenance slice is complete; see work/tasks/native-provenance.md.
Its execution model now has explicit logger projection; transport and durable-history adapters remain future work.
