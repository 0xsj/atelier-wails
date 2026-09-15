# architecture-checks: executable forbidden-import rules

Contract: tools/architecture/RULES.md, revision 1. Dependencies: none; the
rules restate DOMAIN_GUIDE.md. State authority: work/manifest.json.

Implement C01–C05: the layer classifier, the rule table, the Go and Rust
import readers, the inline test-block skip, and the fixture tests. Own
tools/architecture and work/handoffs/architecture-checks.md. Coordinator
owns manifests and the documentation that previously said import rules were
reviewed only. No frontend rules, no CI wiring, no compilation.

Checks:

```sh
python3 tools/architecture/check_imports.py
python3 -m unittest discover -s tools/architecture -p 'test_*.py'
```
