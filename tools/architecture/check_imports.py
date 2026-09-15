"""Enforce the native layer graph on source imports. See RULES.md.

Standard library only. Exit 0 with a summary when no violation exists,
exit 1 listing every violation otherwise.
"""
import re
import sys
from dataclasses import dataclass
from pathlib import Path

CONTEXT_LAYERS = ("domain", "app", "infra", "transport")
PURE_LEAVES = {"errors", "id"}
INFRA_LEAVES = {"errors", "id", "clock", "fileio"}


@dataclass(frozen=True)
class Layer:
    kind: str  # leaf | context | host | root | other
    name: str = ""  # leaf name or context name
    layer: str = ""  # context layer or "" for the context root
    sub: str = ""  # first segment under the layer, e.g. memory / persistence

    def describe(self):
        if self.kind == "leaf":
            return f"leaf {self.name}"
        if self.kind == "context":
            where = self.layer or "root"
            return f"{self.name} {where}" + (f" {self.sub}" if self.sub else "")
        return self.kind


def classify(segments):
    """Map module-relative path segments to a Layer."""
    if not segments:
        return Layer("other")
    head = segments[0]
    if head in ("pkg", "shared"):
        return Layer("leaf", segments[1]) if len(segments) > 1 else Layer("other")
    if head == "host" or (head == "internal" and len(segments) > 1 and segments[1] == "host"):
        return Layer("host")
    if head == "root":
        return Layer("root")
    if head in ("internal", "domains") and len(segments) > 1:
        ctx = segments[1]
        layer = segments[2] if len(segments) > 2 and segments[2] in CONTEXT_LAYERS else ""
        sub = segments[3] if layer and len(segments) > 3 else ""
        return Layer("context", ctx, layer, sub)
    return Layer("other")


def violation(source, target, is_test):
    """Return the rule text a source layer breaks by importing target, or None."""
    if source.kind == "root":
        return None
    if target.kind == "other":
        return None
    if target.kind == "root":
        return "nothing below root imports root"
    if source.kind == "leaf":
        return None if target.kind == "leaf" else "shared leaves import only shared leaves"
    if source.kind == "host":
        if target.kind in ("leaf", "host"):
            return None
        if target.kind == "context" and target.layer in ("app", "transport"):
            return None
        return "host imports only leaves, application ports and transports"
    # source is a context layer or context root
    if target.kind == "host":
        return "contexts do not import host"
    if target.kind == "context" and target.name != source.name:
        return "peer contexts do not import one another"
    if target.kind == "leaf":
        allowed = INFRA_LEAVES if source.layer == "infra" else PURE_LEAVES
        if source.layer == "" or target.name in allowed:
            return None
        return f"{source.layer} imports only leaves {sorted(allowed)}"
    # same context
    if source.layer == "":
        return None
    allowed_layers = {
        "domain": {"domain"},
        "app": {"domain", "app"},
        "infra": {"domain", "app", "infra"},
        "transport": {"domain", "app", "transport"},
    }[source.layer]
    if is_test and source.layer in ("app", "transport"):
        allowed_layers = allowed_layers | {"infra"}
    if target.layer and target.layer not in allowed_layers:
        return f"{source.layer} imports only {sorted(allowed_layers)} of its own context"
    if source.layer == "infra" and source.sub == "persistence" and target.layer == "infra" and target.sub == "memory":
        return "persistence does not import the memory adapter"
    return None


# ---------------------------------------------------------------- Go flavor

GO_IMPORT_BLOCK = re.compile(r"^import\s*\((.*?)^\)", re.S | re.M)
GO_IMPORT_LINE = re.compile(r'^import\s+(?:\w+\s+)?"([^"]+)"', re.M)
GO_QUOTED = re.compile(r'"([^"]+)"')


def go_module(root):
    for line in (root / "go.mod").read_text().splitlines():
        if line.startswith("module "):
            return line.split()[1].strip()
    raise ValueError("go.mod has no module line")


def go_files(root):
    for path in root.rglob("*.go"):
        rel = path.relative_to(root)
        if rel.parts[0] in ("frontend", "build", "node_modules") or ".git" in rel.parts:
            continue
        yield path


def go_imports(text):
    found = []
    for block in GO_IMPORT_BLOCK.findall(text):
        found.extend(GO_QUOTED.findall(block))
    found.extend(GO_IMPORT_LINE.findall(text))
    return found


def check_go(root):
    module = go_module(root)
    prefix = module + "/"
    violations, files, imports = [], 0, 0
    for path in sorted(go_files(root)):
        rel = path.relative_to(root)
        if rel.name == "main.go" and len(rel.parts) == 1:
            source = Layer("root")
        else:
            source = classify(list(rel.parts[:-1]))
        if source.kind == "other":
            continue
        files += 1
        is_test = rel.name.endswith("_test.go")
        for imported in go_imports(path.read_text()):
            if not imported.startswith(prefix):
                continue
            imports += 1
            target = classify(imported[len(prefix):].split("/"))
            rule = violation(source, target, is_test)
            if rule:
                violations.append(f"{rel}: imports {imported} ({source.describe()} -> {target.describe()}): {rule}")
    return violations, files, imports


# -------------------------------------------------------------- Rust flavor

RUST_USE = re.compile(r"^\s*(?:pub(?:\([^)]*\))?\s+)?use\s+(.*?);", re.S | re.M)
RUST_PATH = re.compile(r"\b(crate|super)((?:::\w+)+)")
RUST_CFG_TEST = re.compile(r"#\[cfg\(test\)\]\s*(?:pub(?:\([^)]*\))?\s+)?mod\s+\w+\s*\{")


def strip_inline_tests(text):
    """Remove `#[cfg(test)] mod name { … }` blocks by brace matching."""
    out, position = [], 0
    for match in RUST_CFG_TEST.finditer(text):
        if match.start() < position:
            continue
        out.append(text[position:match.start()])
        depth, index = 1, match.end()
        while index < len(text) and depth:
            depth += {"{": 1, "}": -1}.get(text[index], 0)
            index += 1
        position = index
    out.append(text[position:])
    return "".join(out)


def expand_use(tree):
    """Expand `a::{b, c::{d, e}}` into flat paths."""
    tree = "".join(tree.split())
    results, stack, current = [], [], ""
    index = 0
    while index < len(tree):
        char = tree[index]
        if char == "{":
            stack.append(current)
            current = ""
        elif char == "}":
            if current:
                results.append("".join(stack) + current)
            current = ""
            stack.pop()
        elif char == ",":
            if current:
                results.append("".join(stack) + current)
            current = ""
        else:
            current += char
        index += 1
    if current:
        results.append("".join(stack) + current)
    return [item.split("as")[0] if "::" in item else item for item in results]


def rust_module(rel):
    """Module path segments of a file under src-tauri/src."""
    parts = list(rel.parts)
    if parts[-1] in ("mod.rs", "lib.rs", "main.rs"):
        return parts[:-1]
    return parts[:-1] + [parts[-1][:-3]]


def resolve(module, path):
    segments = path.split("::")
    if segments[0] == "crate":
        return segments[1:]
    if segments[0] == "super":
        base = list(module)
        while segments and segments[0] == "super":
            base = base[:-1]
            segments = segments[1:]
        return base + segments
    if segments[0] == "self":
        return list(module) + segments[1:]
    return None


def rust_imports(text, module):
    found = []
    for use in RUST_USE.findall(text):
        for path in expand_use(use):
            resolved = resolve(module, path)
            if resolved is not None:
                found.append((path, resolved))
    without_uses = RUST_USE.sub("", text)
    for match in RUST_PATH.finditer(without_uses):
        path = match.group(1) + match.group(2)
        resolved = resolve(module, path)
        if resolved is not None:
            found.append((path, resolved))
    return found


def check_rust(root):
    src = root / "src-tauri" / "src"
    violations, files, imports = [], 0, 0
    for path in sorted(src.rglob("*.rs")):
        rel = path.relative_to(src)
        module = rust_module(rel)
        if rel.name in ("lib.rs", "main.rs") and len(rel.parts) == 1:
            source = Layer("root")
        else:
            source = classify(module)
        if source.kind == "other":
            continue
        files += 1
        is_test = rel.name == "tests.rs" or rel.name.endswith(("_test.rs", "_tests.rs"))
        text = strip_inline_tests(path.read_text())
        seen = set()
        for path_text, resolved in rust_imports(text, module):
            target = classify(resolved)
            if target.kind == "other":
                continue
            key = (path_text, target)
            if key in seen:
                continue
            seen.add(key)
            imports += 1
            rule = violation(source, target, is_test)
            if rule:
                violations.append(f"{rel}: imports {path_text} ({source.describe()} -> {target.describe()}): {rule}")
    return violations, files, imports


def check(root):
    root = Path(root)
    if (root / "go.mod").is_file():
        return check_go(root)
    if (root / "src-tauri" / "Cargo.toml").is_file():
        return check_rust(root)
    raise ValueError("neither go.mod nor src-tauri/Cargo.toml found")


def main():
    root = Path(__file__).resolve().parents[2]
    violations, files, imports = check(root)
    if violations:
        print("\n".join(violations), file=sys.stderr)
        print(f"{len(violations)} architecture violation(s)", file=sys.stderr)
        return 1
    print(f"Architecture check passed: {files} files, {imports} internal imports, 0 violations.")
    print("Source-text analysis only; behavior and re-exports are not verified.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
