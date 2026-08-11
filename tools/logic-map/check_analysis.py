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


# Prose coordinates drift silently: the bundle keeps its files, its hash stays
# current and every AST branch stays covered while the map underneath describes
# a revision the function no longer has. Both claims below are optional in the
# corpus — most maps cite neither — so they are checked only where stated.
SOURCE_RANGE = re.compile(r"Source:\s*`[^`]+`\s*\((\d+)\s*[-–]\s*(\d+)\)")
# Anchored to AST on purpose: "미테스트 분기 5개" is prose about coverage, not a
# claim about what the extractor found, and must not be read as one.
AST_BRANCH_COUNT = (
    re.compile(r"AST[^\n]*?\bbranches\s+(\d+)"),
    re.compile(r"AST[^\n]*?분기\s+(\d+)"),
)


# B-T1: until 19판 the checker never opened the file a row pointed at. It
# confirmed the bundle's four files existed, that the AST hash was current and
# that every branch ID had a row -- every one of which a fabricated test name
# satisfies. Ten of a092's rows carried false coverage claims through eighteen
# rounds on exactly that gap.
CITED_TEST = re.compile(r"`(Test[A-Za-z0-9_]+)`")
CITED_TEST_LINE = re.compile(r"`([A-Za-z0-9_./-]*_test\.go):(\d+)`")
GO_TEST_DECL = re.compile(r"^func\s+(Test[A-Za-z0-9_]+)\s*\(")


def test_spans(path: Path) -> dict[str, tuple[int, int]]:
    """Map each Test function in one file to the line range of its body."""
    spans: dict[str, tuple[int, int]] = {}
    try:
        lines = path.read_text(encoding="utf-8", errors="replace").splitlines()
    except OSError:
        return spans
    for index, line in enumerate(lines):
        declared = GO_TEST_DECL.match(line)
        if not declared:
            continue
        depth = 0
        for offset in range(index, len(lines)):
            depth += lines[offset].count("{") - lines[offset].count("}")
            if depth <= 0 and offset > index:
                spans[declared.group(1)] = (index + 1, offset + 1)
                break
        else:
            spans[declared.group(1)] = (index + 1, len(lines))
    return spans


def test_index(root: Path) -> dict[str, list[tuple[Path, int, int]]]:
    """Every Test function in the tree, by name. Built once per check() run."""
    index: dict[str, list[tuple[Path, int, int]]] = {}
    for path in root.rglob("*_test.go"):
        if ".git" in path.parts:
            continue
        for name, (start, end) in test_spans(path).items():
            index.setdefault(name, []).append((path, start, end))
    return index


def resolve_test_file(cited: str, package_dir: Path, root: Path) -> Path | None:
    """Resolve a cited test file: a qualified path wins, then the package, then the tree.

    A citation that carries a directory means it, and honouring that is the
    whole point -- `replay_test.go` exists in both internal/execgw and
    internal/journal, and resolving the bare name against the package under test
    silently answered for the wrong one. Telling an author to qualify the path
    is useless if the qualification is then discarded.

    An ambiguous bare name resolves to nothing rather than to a guess.
    """
    if "/" in cited:
        qualified = root / cited
        return qualified if qualified.is_file() else None
    local = package_dir / cited
    if local.is_file():
        return local
    matches = [path for path in root.rglob(cited) if ".git" not in path.parts]
    return matches[0] if len(matches) == 1 else None


def test_citation_errors(
    target: str, text: str, package_dir: Path, root: Path, index: dict
) -> list[str]:
    errors: list[str] = []
    for name in sorted(set(CITED_TEST.findall(text))):
        if name not in index:
            errors.append(
                f"{target}: branch test map cites {name}, which is not a Go test "
                f"function anywhere in the tree"
            )
    # What is NOT checked, and why -- because a check nobody can satisfy gets
    # worked around, and a check that fires on true rows is worse than none.
    #
    # 19판 built the stronger version first: the cited line must belong to the
    # test named beside it. It was measured against the corpus and withdrawn.
    # Rows legitimately cite a test's doc comment (a092 cites two), the shared
    # harness a test is built from (`obs_test.go:338` is `newNotifier`, a helper),
    # and -- decisively -- a named test in ONE file alongside independent call
    # sites in ANOTHER. a091's `severityof` row does exactly that and is correct:
    #
    #   `TestAQuarantineCreationIsCritical` `a074_quarantine_event_test.go:16`
    #     · `exitloop_test.go:910`, `:1013`
    #
    # Which coordinate answers for which claim is carried by the prose, not by
    # co-occurrence on a line. So the line check stays at what a coordinate can
    # be wrong about on its own: pointing past the end of the file it names.
    for line in text.splitlines():
        for basename, raw in CITED_TEST_LINE.findall(line):
            path = resolve_test_file(basename, package_dir, root)
            if path is None:
                continue
            number = int(raw)
            total = len(path.read_text(encoding="utf-8", errors="replace").splitlines())
            if number > total:
                errors.append(
                    f"{target}: branch test map cites {basename}:{number}, "
                    f"past the end of a {total}-line file"
                )
    return errors


def coordinate_errors(target: str, texts: dict[str, str], value: dict, branches: list) -> list[str]:
    errors: list[str] = []
    start = (value.get("start") or {}).get("line")
    end = (value.get("end") or {}).get("line")
    for name in ("function-logic-map.md", "branch-test-map.md"):
        text = texts.get(name, "")
        cited = SOURCE_RANGE.search(text)
        if cited and (int(cited.group(1)), int(cited.group(2))) != (start, end):
            errors.append(
                f"{target}: {name} cites line range "
                f"{cited.group(1)}-{cited.group(2)} but ast.json is {start}-{end}"
            )
        for pattern in AST_BRANCH_COUNT:
            for claim in pattern.finditer(text):
                if int(claim.group(1)) != len(branches):
                    errors.append(
                        f"{target}: {name} claims AST branch count "
                        f"{claim.group(1)} but ast.json has {len(branches)}"
                    )
    return errors


def validate_target(
    target: Path, root: Path, index: dict | None = None
) -> tuple[list[str], tuple[str, str] | None]:
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
    # Covering every AST branch says nothing about rows the AST never produced.
    # A map left over from a wider revision keeps naming branches that are gone,
    # and each one carries a coverage claim no source line answers for.
    unexpected = set(mapped) - (expected if branches else {"B1"})
    if unexpected:
        errors.append(
            f"{target.name}: branch test map has rows for branch IDs absent "
            f"from the AST {sorted(unexpected)}"
        )
    errors.extend(coordinate_errors(target.name, texts, value, branches))
    errors.extend(
        test_citation_errors(
            target.name,
            branch_map,
            (root / relative).parent,
            root,
            test_index(root) if index is None else index,
        )
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
    # Built once: the tree has thousands of test functions and every target
    # would otherwise rescan them.
    index = test_index(root)
    for target in targets:
        target_errors, binding = validate_target(target, root, index)
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
