#!/usr/bin/env python3
"""Bind Function Logic Map evidence to every modified existing Go function."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
REQUIRED = (
    "ast.json",
    "function-logic-map.md",
    "branch-test-map.md",
    "risk-pattern-report.md",
)
EXEMPTION = "Function Logic Map: not-applicable"
HUNK = re.compile(r"^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@")
MAP_SECTIONS = (
    "## Inputs and invariants",
    "## Branches and early returns",
    "## Calls and live bindings",
    "## State mutations and fallbacks",
    "## Safety conclusion",
)


def qualified(value: dict) -> str:
    receiver = str(value.get("receiver", "")).strip()
    function = str(value.get("function", "")).strip()
    return f"{receiver}.{function}" if receiver else function


def go_functions(path: Path, root: Path) -> list[dict]:
    process = subprocess.run(
        [
            "go",
            "run",
            "./tools/logic-map",
            "--file",
            str(path),
            "--list",
            "--pretty=false",
        ],
        cwd=root,
        capture_output=True,
        text=True,
        timeout=60,
        check=False,
    )
    if process.returncode:
        raise RuntimeError(process.stderr.strip() or f"cannot parse {path}")
    value = json.loads(process.stdout)
    return value if isinstance(value, list) else []


def intersects(start: int, end: int, line: int, count: int) -> bool:
    if count == 0:
        return start <= line <= end + 1
    return start <= line + count - 1 and line <= end


def base_file(root: Path, base: str, source: str) -> Path | None:
    process = subprocess.run(
        ["git", "show", f"{base}:{source}"],
        cwd=root,
        capture_output=True,
        check=False,
    )
    if process.returncode:
        return None
    descriptor, name = tempfile.mkstemp(suffix=".go")
    os.close(descriptor)
    path = Path(name)
    path.write_bytes(process.stdout)
    return path


def changed_existing_functions(
    root: Path = ROOT,
    base: str = "",
) -> dict[tuple[str, str], dict]:
    if not base:
        raise ValueError("Function Logic Map comparison base is required")
    process = subprocess.run(
        [
            "git",
            "diff",
            "--no-ext-diff",
            "--find-renames",
            "--unified=0",
            base,
            "--",
            "*.go",
        ],
        cwd=root,
        capture_output=True,
        text=True,
        timeout=30,
        check=False,
    )
    if process.returncode:
        raise RuntimeError(process.stderr.strip() or f"git diff failed for base {base}")
    required: dict[tuple[str, str], dict] = {}
    old_source = ""
    new_source = ""
    hunks: list[tuple[int, int, int, int]] = []

    def flush() -> None:
        nonlocal hunks
        if not old_source or not hunks:
            hunks = []
            return
        temporary = base_file(root, base, old_source)
        if temporary is None:
            raise RuntimeError(
                f"cannot load existing base file {base}:{old_source}"
            )
        try:
            old_functions = go_functions(temporary, root)
            old_qualified = {qualified(function) for function in old_functions}
            for function in old_functions:
                start = int(function["start"]["line"])
                end = int(function["end"]["line"])
                if any(intersects(start, end, old, old_count) for old, old_count, _, _ in hunks):
                    key = (new_source or old_source, qualified(function))
                    required[key] = {
                        "file": new_source or old_source,
                        "function": qualified(function),
                        "base_hash": function.get("source_sha256"),
                    }
            current = root / new_source if new_source else None
            if current and current.exists():
                for function in go_functions(current, root):
                    # Function Logic Maps are required for functions that existed
                    # at the frozen comparison base and whose logic changed. A
                    # newly added function in an existing file has no base logic
                    # to map; treating it as "modified existing" contradicts the
                    # workflow contract and makes added test helpers look like
                    # high-risk edits.
                    if qualified(function) not in old_qualified:
                        continue
                    start = int(function["start"]["line"])
                    end = int(function["end"]["line"])
                    if any(
                        intersects(start, end, new, new_count)
                        for _, _, new, new_count in hunks
                    ):
                        key = (new_source, qualified(function))
                        required[key] = {
                            "file": new_source,
                            "function": qualified(function),
                            "current_hash": function.get("source_sha256"),
                        }
        finally:
            temporary.unlink(missing_ok=True)
        hunks = []

    for line in process.stdout.splitlines():
        if line.startswith("diff --git "):
            flush()
            old_source = ""
            new_source = ""
        elif line.startswith("--- "):
            value = line[4:]
            old_source = "" if value == "/dev/null" else value.removeprefix("a/")
        elif line.startswith("+++ "):
            value = line[4:]
            new_source = "" if value == "/dev/null" else value.removeprefix("b/")
        else:
            match = HUNK.match(line)
            if match:
                hunks.append(
                    (
                        int(match.group(1)),
                        int(match.group(2) or 1),
                        int(match.group(3)),
                        int(match.group(4) or 1),
                    )
                )
    flush()
    return required


def resolve_base(change_dir: Path, root: Path) -> str:
    path = change_dir / "base-commit.txt"
    try:
        candidate = path.read_text(encoding="utf-8").strip()
    except OSError as exc:
        raise ValueError(
            "missing base-commit.txt; run "
            "`python3 tools/sdd/capture_change_base.py --change <change-id>` "
            "at proposal freeze"
        ) from exc

    def resolve(value: str) -> str:
        process = subprocess.run(
            ["git", "rev-parse", "--verify", f"{value}^{{commit}}"],
            cwd=root,
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
        )
        if process.returncode:
            raise ValueError(
                process.stderr.strip() or f"invalid Function Logic Map base: {value}"
            )
        return process.stdout.strip()

    persisted = resolve(candidate)
    override = os.environ.get("SDD_BASE_REF", "").strip()
    if override and resolve(override) != persisted:
        raise ValueError("SDD_BASE_REF must resolve to the persisted base-commit.txt")
    return persisted


def normalized_source(value: str, root: Path) -> tuple[Path, str]:
    raw = Path(value)
    path = raw if raw.is_absolute() else root / raw
    resolved = path.resolve()
    if not resolved.is_relative_to(root.resolve()):
        raise ValueError("AST source escapes repository")
    return resolved, resolved.relative_to(root.resolve()).as_posix()


def branch_ids(text: str) -> list[str]:
    return [
        match.group(1)
        for line in text.splitlines()
        if (match := re.match(r"^\|\s*(B\d+)\s*\|", line))
    ]


def validate_target(target: Path, root: Path) -> tuple[list[str], tuple[str, str] | None]:
    errors: list[str] = []
    texts: dict[str, str] = {}
    for name in REQUIRED:
        path = target / name
        if not path.exists():
            errors.append(f"{target.name}: missing {name}")
            continue
        texts[name] = path.read_text(encoding="utf-8")
        if "TODO" in texts[name]:
            errors.append(f"{target.name}: {name} still contains TODO")
    if "ast.json" not in texts:
        return errors, None
    try:
        value = json.loads(texts["ast.json"])
    except ValueError:
        return errors + [f"{target.name}: ast.json is invalid"], None
    keys = ("file", "source_sha256", "package", "function", "signature", "start", "end")
    if not isinstance(value, dict) or any(not value.get(key) for key in keys):
        return errors + [f"{target.name}: ast.json is placeholder evidence"], None
    try:
        source, relative = normalized_source(str(value["file"]), root)
    except ValueError as exc:
        return errors + [f"{target.name}: {exc}"], None
    revision = value.get("revision", "current")
    if revision == "current":
        if not source.is_file():
            errors.append(f"{target.name}: AST source is missing: {relative}")
        elif hashlib.sha256(source.read_bytes()).hexdigest() != value["source_sha256"]:
            errors.append(f"{target.name}: AST source hash is stale: {relative}")
    elif revision != "base":
        errors.append(f"{target.name}: unsupported AST revision {revision!r}")
    function = qualified(value)
    logic = texts.get("function-logic-map.md", "")
    for section in MAP_SECTIONS:
        if section not in logic:
            errors.append(f"{target.name}: function map missing section {section}")
    if relative not in logic or function not in logic:
        errors.append(f"{target.name}: function map is not source/function-bound")
    # The Go extractor marshals a nil slice as JSON null, so a branchless
    # function arrives as "branches": null rather than []. Treat both as empty.
    branches = value.get("branches") or []
    branch_map = texts.get("branch-test-map.md", "")
    if "# Branch Test Map" not in branch_map:
        errors.append(f"{target.name}: branch test map header missing")
    mapped = branch_ids(branch_map)
    if len(mapped) != len(set(mapped)):
        errors.append(f"{target.name}: branch test map contains duplicate branch IDs")
    expected = {
        str(branch.get("id"))
        for branch in branches
        if isinstance(branch, dict) and branch.get("id")
    }
    if branches and len(expected) != len(branches):
        errors.append(f"{target.name}: AST branches do not have stable unique IDs")
    missing = expected - set(mapped)
    if missing:
        errors.append(
            f"{target.name}: branch test map is missing AST branches {sorted(missing)}"
        )
    if not branches and "B1" not in mapped:
        errors.append(f"{target.name}: branchless function still needs one happy-path row")
    risk = texts.get("risk-pattern-report.md", "")
    if "# Risk Pattern Report" not in risk or relative not in risk:
        errors.append(f"{target.name}: risk report is not source-bound")
    return errors, (relative, function)


def check(change: str, root: Path = ROOT) -> list[str]:
    change_dir = root / "openspec" / "changes" / change
    analysis = change_dir / "analysis" / "function-logic"
    reference_file = change_dir / "analysis" / "function-logic-reference.txt"
    review = change_dir / "review.md"
    review_text = review.read_text(encoding="utf-8") if review.exists() else ""
    try:
        base = resolve_base(change_dir, root)
        required = changed_existing_functions(root, base)
    except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
        return [f"cannot derive modified Go functions: {exc}"]
    if reference_file.exists():
        if analysis.exists() and any(path.is_dir() and any(path.iterdir()) for path in analysis.iterdir()):
            return ["function-logic reference cannot coexist with local function-logic evidence"]
        referenced_change = reference_file.read_text(encoding="utf-8").strip()
        if not re.fullmatch(r"[a-z0-9][a-z0-9-]*", referenced_change) or referenced_change == change:
            return ["function-logic reference names an invalid or recursive change"]
        referenced_dir = root / "openspec" / "changes" / referenced_change
        try:
            referenced_base = resolve_base(referenced_dir, root)
        except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
            return [f"function-logic reference base is invalid: {exc}"]
        if referenced_base != base:
            return ["function-logic reference must share the exact comparison base"]
        analysis = referenced_dir / "analysis" / "function-logic"
    if not analysis.exists():
        if required:
            names = ", ".join(f"{source}:{function}" for source, function in required)
            return [f"missing Function Logic Map for modified existing functions: {names}"]
        return [] if EXEMPTION in review_text else [f"missing analysis or `{EXEMPTION}` review marker"]
    errors: list[str] = []
    targets = sorted(path for path in analysis.iterdir() if path.is_dir())
    if not targets:
        return ["function-logic analysis directory has no targets"]
    covered: dict[tuple[str, str], Path] = {}
    for target in targets:
        target_errors, binding = validate_target(target, root)
        errors.extend(target_errors)
        if binding:
            if binding in covered:
                errors.append(f"{target.name}: duplicate evidence for {binding[0]}:{binding[1]}")
            covered[binding] = target
    for binding, expected in required.items():
        target = covered.get(binding)
        if target is None:
            errors.append(f"missing evidence for modified function {binding[0]}:{binding[1]}")
            continue
        try:
            ast_value = json.loads((target / "ast.json").read_text(encoding="utf-8"))
        except (OSError, ValueError):
            continue
        expected_hash = expected.get("current_hash") or expected.get("base_hash")
        if ast_value.get("source_sha256") != expected_hash:
            errors.append(f"{target.name}: AST hash does not match modified function revision")
        expected_revision = "current" if expected.get("current_hash") else "base"
        if ast_value.get("revision", "current") != expected_revision:
            errors.append(f"{target.name}: AST revision must be {expected_revision}")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--change", required=True)
    parser.add_argument("--root", default=str(ROOT))
    args = parser.parse_args()
    errors = check(args.change, Path(args.root))
    if errors:
        for error in errors:
            print(f"[logic-map] {error}")
        return 1
    print(f"[logic-map] {args.change}: evidence complete or diff-proven exempt")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
