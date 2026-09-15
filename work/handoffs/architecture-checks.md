# architecture-checks handoff

Contract revision: 1 (`tools/architecture/RULES.md`). Implemented and
coordinator-integrated on 2026-09-12.

Files: `tools/architecture/check_imports.py`, `test_check_imports.py`,
`RULES.md`, `README.md`. Standard library only. The same script and tests
are carried in the Tauri repository; each repository stays independent.

Scenario coverage: C01 is the run against the real tree; C02 and C03 are
`GoRules.test_c02_refusals` and `test_c03_allowed` plus the narrower checks
for an effectful leaf in a domain, persistence importing memory, and the
test exemption not reaching a peer; C04 is `RustRules.test_c04_…` with the
nested-use and inline-block unit checks; C05 is implied by both fixture
classes sharing one rule table.

Verification: `python3 tools/architecture/check_imports.py`: passed, 74
files, 142 internal imports, 0 violations. `python3 -m unittest discover -s
tools/architecture -p 'test_*.py'`: 10 tests passed. A probe file placed in
the preferences domain importing the memory adapter was refused with one
violation naming the file, the import and the rule, then removed.

Limitations: source-text analysis; no compilation, re-export following or
frontend coverage. The check is now part of the recorded checks for future
tasks but is not wired into CI.
