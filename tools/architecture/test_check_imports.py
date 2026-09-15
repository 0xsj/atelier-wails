import tempfile
import unittest
from pathlib import Path

import check_imports as checker

MODULE = "github.com/example/app"


def write(root, rel, text):
    path = Path(root) / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text)


def go_file(imports):
    lines = "\n".join(f'\t"{MODULE}/{item}"' for item in imports)
    return f"package x\n\nimport (\n{lines}\n)\n"


class GoFixture:
    def __init__(self):
        self.dir = tempfile.TemporaryDirectory()
        self.root = Path(self.dir.name)
        write(self.root, "go.mod", f"module {MODULE}\n\ngo 1.22\n")

    def add(self, rel, imports):
        write(self.root, rel, go_file(imports))
        return self

    def violations(self):
        return checker.check(self.root)[0]


class RustFixture:
    def __init__(self):
        self.dir = tempfile.TemporaryDirectory()
        self.root = Path(self.dir.name)
        write(self.root, "src-tauri/Cargo.toml", "[package]\nname = \"x\"\n")

    def add(self, rel, text):
        write(self.root, "src-tauri/src/" + rel, text)
        return self

    def violations(self):
        return checker.check(self.root)[0]


class GoRules(unittest.TestCase):
    def test_c02_refusals(self):
        fixture = (
            GoFixture()
            .add("internal/prefs/domain/domain.go", ["internal/prefs/infra/memory"])
            .add("internal/prefs/app/command/command.go", ["internal/workspace/domain"])
            .add("pkg/errors/errors.go", ["internal/prefs/domain"])
            .add("internal/prefs/transport/desktop/handler.go", ["internal/prefs/infra/memory"])
            .add("internal/host/wails/host.go", ["internal/prefs/domain"])
            .add("pkg/id/id.go", ["root"])
        )
        found = fixture.violations()
        self.assertEqual(len(found), 6, found)
        self.assertTrue(any("domain imports only" in v for v in found))
        self.assertTrue(any("peer contexts" in v for v in found))
        self.assertTrue(any("shared leaves import only" in v for v in found))
        self.assertTrue(any("transport imports only" in v for v in found))
        self.assertTrue(any("host imports only" in v for v in found))
        self.assertTrue(any("imports root" in v for v in found))

    def test_c03_allowed(self):
        fixture = (
            GoFixture()
            .add("internal/prefs/app/command/spec_test.go", ["internal/prefs/infra/memory", "internal/prefs/domain"])
            .add("internal/prefs/infra/memory/memory.go", ["internal/prefs/app/command", "pkg/fileio", "pkg/errors"])
            .add("internal/host/wails/host.go", ["internal/prefs/app/command", "internal/prefs/transport/desktop", "pkg/errors"])
            .add("root/root.go", ["internal/host/wails", "internal/prefs/infra/persistence", "pkg/logger"])
            .add("pkg/logger/logger.go", ["pkg/provenance", "pkg/errors"])
            .add("internal/prefs/domain/domain.go", ["pkg/errors", "pkg/id"])
        )
        self.assertEqual(fixture.violations(), [])

    def test_domain_refuses_effectful_leaf(self):
        fixture = GoFixture().add("internal/prefs/domain/domain.go", ["pkg/logger"])
        self.assertEqual(len(fixture.violations()), 1)

    def test_persistence_refuses_memory(self):
        fixture = GoFixture().add("internal/prefs/infra/persistence/store.go", ["internal/prefs/infra/memory"])
        self.assertTrue(any("persistence does not import" in v for v in fixture.violations()))

    def test_test_exemption_is_narrow(self):
        fixture = GoFixture().add("internal/prefs/app/command/spec_test.go", ["internal/workspace/infra/memory"])
        self.assertTrue(any("peer contexts" in v for v in fixture.violations()))


class RustRules(unittest.TestCase):
    def test_c02_refusals(self):
        fixture = (
            RustFixture()
            .add("domains/prefs/domain/mod.rs", "use crate::domains::prefs::infra::memory::Store;\n")
            .add("domains/prefs/app/command.rs", "use crate::domains::workspace::domain::Workspace;\n")
            .add("shared/errors/mod.rs", "use crate::domains::prefs::domain::Key;\n")
            .add("domains/prefs/transport/desktop.rs", "use crate::domains::prefs::infra::memory::Store;\n")
            .add("host/tauri/mod.rs", "use crate::domains::prefs::domain::Key;\n")
            .add("shared/id/mod.rs", "fn f() { crate::root::run(); }\n")
        )
        found = fixture.violations()
        self.assertEqual(len(found), 6, found)

    def test_c04_super_inline_tests_and_persistence(self):
        fixture = (
            RustFixture()
            .add("domains/prefs/app/command.rs", "use super::super::infra::memory::Store;\n")
            .add(
                "domains/prefs/app/query.rs",
                "use crate::domains::prefs::domain::Key;\n\n#[cfg(test)]\nmod tests {\n    use crate::domains::prefs::infra::memory::Store;\n    fn f() { let _ = Store::new(); }\n}\n",
            )
            .add("domains/prefs/infra/persistence.rs", "use crate::domains::prefs::infra::memory::Store;\n")
        )
        found = fixture.violations()
        self.assertEqual(len(found), 2, found)
        self.assertTrue(any("command.rs" in v and "super::super::infra" in v for v in found))
        self.assertTrue(any("persistence does not import" in v for v in found))

    def test_c03_allowed(self):
        fixture = (
            RustFixture()
            .add("domains/prefs/app/tests.rs", "use crate::domains::prefs::infra::memory::Store;\nuse super::command;\n")
            .add("domains/prefs/infra/memory.rs", "use crate::domains::prefs::app::{command, query};\nuse crate::shared::{errors::Failure, fileio};\n")
            .add("host/tauri/preferences.rs", "use crate::domains::prefs::app::command::Discard;\nuse crate::domains::prefs::transport::desktop::Handler;\n")
            .add("root/mod.rs", "fn run() { crate::host::tauri::run(crate::domains::prefs::infra::persistence::Store::new); }\n")
            .add("domains/prefs/tests.rs", "use super::{app::command, infra::memory::Store};\n")
            .add("shared/logger/mod.rs", "use crate::shared::{errors::Failure, provenance::Scope};\n")
        )
        self.assertEqual(fixture.violations(), [])

    def test_nested_use_expansion(self):
        paths = checker.expand_use("crate::shared::{errors::{Failure, Kind}, id::Id}")
        self.assertEqual(paths, ["crate::shared::errors::Failure", "crate::shared::errors::Kind", "crate::shared::id::Id"])

    def test_inline_test_block_skipped_but_declaration_kept(self):
        text = "#[cfg(test)]\nmod tests;\nuse crate::a::b;\n#[cfg(test)]\nmod inline { use crate::c::d; }\nuse crate::e::f;\n"
        stripped = checker.strip_inline_tests(text)
        self.assertIn("crate::a::b", stripped)
        self.assertIn("crate::e::f", stripped)
        self.assertNotIn("crate::c::d", stripped)


if __name__ == "__main__":
    unittest.main()
