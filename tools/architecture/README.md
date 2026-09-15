# architecture

Import checks for the declared native layer graph. `check_imports.py` reads
every Go or Rust source file, classifies it by path into leaf, context layer,
host or root, and refuses imports the rule table in [RULES.md](RULES.md)
forbids. Test files may reach their own context's infra; nothing else is
relaxed.

```sh
python3 tools/architecture/check_imports.py
python3 -m unittest discover -s tools/architecture -p 'test_*.py'
```

Source-text analysis only: no compilation, no re-export following. The
frontend tree is not covered. Standard library only.
