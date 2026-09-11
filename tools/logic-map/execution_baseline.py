"""Fixed a063 execution-baseline adoption helpers."""

from __future__ import annotations

import hashlib
import json
import os
import stat
import subprocess
import argparse
from pathlib import Path

CHANGE = "a063-align-attestation-renewal-profile"
P = "da80ce31b6a1ab5d443016768f970a82bab102db"
E = "e65e394bf84b3c6e4559a219e816af96d341d75d"


class AdoptionError(ValueError):
    pass


def canonical(value: object) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=True).encode("utf-8")


def digest(value: object) -> str:
    return hashlib.sha256(canonical(value)).hexdigest()


def _git(root: Path, *args: str, text: bool = True) -> subprocess.CompletedProcess:
    process = subprocess.run(["git", *args], cwd=root, capture_output=True, text=text, check=False)
    if process.returncode:
        error = process.stderr if text else process.stderr.decode("utf-8", "replace")
        raise AdoptionError(error.strip() or "git command failed")
    return process


def full_commit(root: Path, value: str) -> str:
    if not isinstance(value, str) or len(value) != 40 or any(char not in "0123456789abcdef" for char in value):
        raise AdoptionError("commit must be a full lowercase SHA-1")
    return _git(root, "rev-parse", "--verify", value + "^{commit}").stdout.strip()


def _pairs(pairs: list[tuple[str, object]]) -> dict:
    result: dict = {}
    for key, value in pairs:
        if key in result:
            raise AdoptionError("duplicate JSON key")
        result[key] = value
    return result


def strict_json(raw: bytes) -> dict:
    try:
        value = json.loads(raw, object_pairs_hook=_pairs)
    except (UnicodeDecodeError, ValueError) as error:
        raise AdoptionError(f"invalid JSON: {error}") from error
    if not isinstance(value, dict):
        raise AdoptionError("JSON document must be an object")
    return value


def _schema_one(value: object, document: str) -> None:
    # bool is an int subclass; JSON true is never schema version 1.
    if type(value) is not int or value != 1:
        raise AdoptionError(f"{document} schema must be exact integer 1")


def _sha256(value: object, field: str) -> None:
    if not isinstance(value, str) or len(value) != 64 or any(char not in "0123456789abcdef" for char in value):
        raise AdoptionError(f"{field} must be a lowercase SHA-256")


def _change_analysis_path(value: object, analysis: str, field: str) -> str:
    if not isinstance(value, str):
        raise AdoptionError(f"{field} must be a string")
    candidate = Path(value)
    if candidate.is_absolute() or ".." in candidate.parts or value == analysis.rstrip("/") or not value.startswith(analysis):
        raise AdoptionError("adoption evidence path is outside current change analysis")
    return value


def ancestry(root: Path, older: str, newer: str, *, strict: bool = False) -> None:
    result = subprocess.run(["git", "merge-base", "--is-ancestor", older, newer], cwd=root, capture_output=True, check=False)
    if result.returncode or (strict and older == newer):
        raise AdoptionError("required commit ancestry is absent")


def history(root: Path, older: str, newer: str) -> list[dict]:
    commits = _git(root, "rev-list", "--topo-order", "--reverse", newer, "--not", older).stdout.splitlines()
    output: list[dict] = []
    empty = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
    for oid in commits:
        parents = _git(root, "show", "-s", "--format=%P", oid).stdout.strip().split()
        parent_diffs = []
        for parent in (parents or [empty]):
            raw = _git(root, "diff-tree", "--no-commit-id", "--no-renames", "--name-status", "-z", "-r", parent, oid, text=False).stdout
            fields = raw.split(b"\0")
            changes = []
            for index in range(0, len(fields) - 1, 2):
                try:
                    changes.append({"status": fields[index].decode("utf-8", "strict"), "path": fields[index + 1].decode("utf-8", "strict")})
                except UnicodeDecodeError as error:
                    raise AdoptionError("Git filename is not UTF-8") from error
            parent_diffs.append({"parent": parent, "changes": sorted(changes, key=lambda item: (item["path"], item["status"]))})
        output.append({"commit": oid, "parents": parents, "parent_diffs": parent_diffs})
    return output


def _regular_committed(root: Path, relative: str, head: str) -> bytes:
    candidate = Path(relative)
    if candidate.is_absolute() or ".." in candidate.parts:
        raise AdoptionError("evidence path escapes repository")
    path = root / candidate
    try:
        probe = root
        for part in candidate.parts:
            probe = probe / part
            if stat.S_ISLNK(probe.lstat().st_mode):
                raise AdoptionError(f"evidence path contains symlink: {relative}")
        if not stat.S_ISREG(path.lstat().st_mode):
            raise AdoptionError(f"evidence is not a regular file: {relative}")
    except OSError as error:
        raise AdoptionError(f"missing evidence: {relative}") from error
    committed = _git(root, "show", f"{head}:{relative}", text=False).stdout
    if path.read_bytes() != committed:
        raise AdoptionError(f"evidence differs from HEAD: {relative}")
    return committed


def _allowed_metadata(path: str) -> bool:
    lower = path.lower()
    forbidden = (".go", ".py", ".sh", ".c", ".cc", ".cpp", ".cxx", ".m", ".h", ".hh", ".hpp", ".hxx", ".s", ".f", ".for", ".f90", ".swig", ".swigcxx", ".syso", ".o", ".a", ".so", ".dylib", ".dll")
    if lower.endswith(forbidden):
        return False
    cache = Path(path)
    direct_pyc = cache.suffix == ".pyc" and cache.parent.name == "__pycache__" and (cache.parts[0] in {"tools", "auth-helper"})
    return path.startswith(".codegraph/") or path == ".sdd/index-state.json" or path.startswith(".sdd/gbrain-home/") or path.startswith(".sdd/history/checkpoints/") or path.startswith(".sdd/history/refresh-queue/") or path == ".sdd/history/index-refresh.lock" or (path.startswith(".sdd/history/events/") and path.endswith(".jsonl") and "/" not in path[len(".sdd/history/events/"):]) or direct_pyc


def nul_paths(root: Path, *args: str) -> list[str]:
    raw = _git(root, *args, "-z", text=False).stdout
    result = []
    for value in raw.split(b"\0"):
        if not value:
            continue
        try:
            result.append(value.decode("utf-8", "strict"))
        except UnicodeDecodeError as error:
            raise AdoptionError("Git filename is not UTF-8") from error
    return result


def go_tree_entries(root: Path, commit: str) -> dict[str, tuple[str, str]]:
    """Return immutable Go path -> (Git mode, blob) without lossy path parsing."""
    raw = _git(root, "ls-tree", "-r", "-z", commit, text=False).stdout
    entries: dict[str, tuple[str, str]] = {}
    for value in (item for item in raw.split(b"\0") if item):
        try:
            meta, raw_name = value.split(b"\t", 1)
            mode, kind, blob = meta.decode("ascii").split(" ", 2)
            name = raw_name.decode("utf-8", "strict")
        except (UnicodeDecodeError, ValueError) as error:
            raise AdoptionError("Git Go source path is malformed or not UTF-8") from error
        if not name.endswith(".go"):
            continue
        if kind != "blob" or mode not in {"100644", "100755"}:
            raise AdoptionError("source snapshot contains non-regular Go file")
        entries[name] = (mode, blob)
    return entries


def verify_source_go_lock(root: Path, source: str, head: str) -> None:
    """Lock every tracked Go source path and executable bit to S, bypassing Git config."""
    expected = go_tree_entries(root, source)
    if go_tree_entries(root, head) != expected:
        raise AdoptionError("source snapshot Go tree differs from evidence HEAD")
    for relative, (mode, _) in expected.items():
        # _regular_committed compares the worktree bytes to the immutable H blob.
        _regular_committed(root, relative, head)
        item = root / relative
        try:
            actual_mode = item.lstat().st_mode
        except OSError as error:
            raise AdoptionError(f"missing source snapshot Go file: {relative}") from error
        if not stat.S_ISREG(actual_mode):
            raise AdoptionError(f"source snapshot Go worktree file is not regular: {relative}")
        executable = bool(actual_mode & 0o111)
        if executable != (mode == "100755"):
            raise AdoptionError(f"source snapshot Go worktree mode mismatch: {relative}")


def untracked_filesystem_entries(root: Path) -> set[str]:
    """All untracked filesystem entries, without following a directory link."""
    tracked = set(nul_paths(root, "ls-files"))
    found: set[str] = set()
    def visit(directory: Path, relative: Path) -> None:
        for entry in os.scandir(directory):
            if not relative.parts and entry.name == ".git":
                continue
            child_relative = relative / entry.name
            name = child_relative.as_posix()
            mode = entry.stat(follow_symlinks=False).st_mode
            if stat.S_ISLNK(mode):
                if name not in tracked:
                    found.add(name)
                continue
            if stat.S_ISDIR(mode):
                visit(Path(entry.path), child_relative)
            elif name not in tracked:
                found.add(name)
    visit(root, Path())
    return found


def go_inputs(root: Path, tags: str = "") -> set[str]:
    command = ["go", "list", "-deps", "-test", "-json"]
    if tags:
        command.extend(["-tags", tags])
    command.append("./...")
    process = subprocess.run(command, cwd=root, capture_output=True, text=True, check=False)
    if process.returncode:
        raise AdoptionError(process.stderr.strip() or "go package enumeration failed")
    decoder = json.JSONDecoder(); offset = 0; result: set[str] = set()
    fields = ("GoFiles", "CgoFiles", "CompiledGoFiles", "IgnoredGoFiles", "IgnoredOtherFiles", "CFiles", "CXXFiles", "MFiles", "HFiles", "FFiles", "SFiles", "SwigFiles", "SwigCXXFiles", "SysoFiles", "TestGoFiles", "XTestGoFiles", "EmbedFiles", "TestEmbedFiles", "XTestEmbedFiles")
    while offset < len(process.stdout):
        while offset < len(process.stdout) and process.stdout[offset].isspace():
            offset += 1
        if offset >= len(process.stdout):
            break
        try:
            package, offset = decoder.raw_decode(process.stdout, offset)
        except ValueError as error:
            raise AdoptionError(f"invalid go package enumeration JSON: {error}") from error
        if not isinstance(package, dict):
            raise AdoptionError("go package enumeration entry is not an object")
        directory = package.get("Dir")
        if not isinstance(directory, str):
            if any(field in package for field in fields):
                raise AdoptionError("go package enumeration entry has no directory")
            continue
        base = Path(directory)
        try:
            relative_base = base.resolve().relative_to(root.resolve())
        except ValueError:
            continue
        for field in fields:
            value = package.get(field, [])
            if value is None:
                value = []
            if not isinstance(value, list) or not all(isinstance(name, str) for name in value):
                raise AdoptionError(f"go package enumeration {field} is malformed")
            for name in value:
                candidate = Path(name)
                if candidate.is_absolute():
                    try:
                        result.add(candidate.resolve().relative_to(root.resolve()).as_posix())
                    except ValueError:
                        # -test may report generated toolchain files outside the repository.
                        continue
                    continue
                if ".." in candidate.parts:
                    raise AdoptionError(f"go package enumeration {field} path escapes package")
                result.add((relative_base / candidate).as_posix())
    return result


def function_inventory(root: Path, older: str, newer: str) -> list[dict]:
    """Ordinary checker semantics over two immutable commits, without checkout."""
    # Import lazily: check_analysis imports this module to select the base.
    from check_analysis import changed_existing_functions
    required = changed_existing_functions(root, older, newer)
    rows = []
    for (_, _), value in required.items():
        if "current_hash" in value:
            revision, source_hash = "current", value["current_hash"]
        else:
            revision, source_hash = "base", value["base_hash"]
        rows.append({"file": value["file"], "function": value["function"], "revision": revision, "source_sha256": source_hash})
    return sorted(rows, key=lambda row: (row["file"], row["function"]))


def exact_inventory(recorded: object, expected: list[dict]) -> None:
    if not isinstance(recorded, list):
        raise AdoptionError("function inventory must be a list")
    keys = []
    for row in recorded:
        if not isinstance(row, dict) or set(row) != {"file", "function", "revision", "source_sha256"}:
            raise AdoptionError("function inventory row schema is invalid")
        if (not isinstance(row["file"], str) or not isinstance(row["function"], str)
                or row["revision"] not in {"current", "base"}):
            raise AdoptionError("function inventory row types are invalid")
        _sha256(row["source_sha256"], "function inventory source_sha256")
        keys.append((row["file"], row["function"]))
    if len(keys) != len(set(keys)):
        raise AdoptionError("function inventory contains duplicate rows")
    if recorded != expected:
        raise AdoptionError("function inventory differs from immutable recomputation")


def range_payload(root: Path, older: str, newer: str) -> dict:
    functions = function_inventory(root, older, newer)
    return {"commits": history(root, older, newer), "function_inventory": functions, "function_inventory_sha256": digest(functions)}


def _fixed_draft_paths(root: Path, change: str, ledger_path: Path, record_path: Path) -> tuple[Path, Path]:
    root = root.resolve()
    change_root = root / "openspec" / "changes" / change
    expected_ledger = change_root / "analysis" / "execution-baseline-ledger.json"
    expected_record = change_root / "execution-baseline.json"
    ledger = ledger_path if ledger_path.is_absolute() else root / ledger_path
    record = record_path if record_path.is_absolute() else root / record_path
    if ledger != expected_ledger or record != expected_record:
        raise AdoptionError("draft output paths must be the fixed change record and analysis ledger")
    for path in (ledger, record):
        relative = path.relative_to(root)
        probe = root
        for part in relative.parts[:-1]:
            probe = probe / part
            try:
                mode = probe.lstat().st_mode
            except FileNotFoundError:
                continue
            if stat.S_ISLNK(mode) or not stat.S_ISDIR(mode):
                raise AdoptionError(f"draft output parent is unsafe: {relative.as_posix()}")
        try:
            path.lstat()
        except FileNotFoundError:
            continue
        raise AdoptionError(f"draft output already exists: {relative.as_posix()}")
    return ledger, record


def _exclusive_write(path: Path, contents: bytes) -> None:
    try:
        with path.open("xb") as output:
            output.write(contents)
    except FileExistsError as error:
        raise AdoptionError(f"draft output already exists: {path}") from error


def draft(root: Path, change: str, source: str, ledger_path: Path, record_path: Path) -> None:
    """Write review drafts once; never create an approval or alter a checkout/base."""
    if change != CHANGE:
        raise AdoptionError("draft generation only supports the fixed a063 change")
    root = root.resolve()
    ledger_path, record_path = _fixed_draft_paths(root, change, ledger_path, record_path)
    source = full_commit(root, source)
    ancestry(root, P, E); ancestry(root, E, source)
    pe = range_payload(root, P, E)
    es = range_payload(root, E, source)
    go_paths = sorted(path for path in nul_paths(root, "diff", "--name-only", E, source) if path.endswith(".go"))
    ledger = {"schema": 1, "planning_to_execution": pe, "execution_to_source": {**es, "go_paths": go_paths}, "inherited_history_disposition": "committed historical work; missing original analysis remains debt"}
    ledger_path.parent.mkdir(parents=True, exist_ok=True)
    ledger_raw = canonical(ledger)
    relative_ledger = ledger_path.relative_to(root).as_posix()
    record = {"schema": 1, "change": CHANGE, "planning_base": P, "execution_base": E, "source_commit": source, "source_tree": _git(root, "rev-parse", source + "^{tree}").stdout.strip(), "pre_edit_provenance": "retrospective-exception", "ledger_path": relative_ledger, "ledger_sha256": hashlib.sha256(ledger_raw).hexdigest(), "adversarial_review_path": "", "adversarial_review_sha256": "", "gstack_review_path": "", "gstack_review_sha256": "", "inherited_history_disposition": ledger["inherited_history_disposition"]}
    created_ledger = False
    try:
        _exclusive_write(ledger_path, ledger_raw)
        created_ledger = True
        _exclusive_write(record_path, canonical(record))
    except Exception:
        if created_ledger:
            try:
                ledger_path.unlink()
            except OSError:
                pass
        raise


def main() -> int:
    parser = argparse.ArgumentParser(description="Create non-approving fixed a063 execution-baseline drafts")
    parser.add_argument("--root", default=".")
    parser.add_argument("--change", required=True)
    parser.add_argument("--source", required=True)
    parser.add_argument("--ledger")
    parser.add_argument("--record")
    args = parser.parse_args()
    root = Path(args.root).resolve()
    change_dir = root / "openspec" / "changes" / args.change
    ledger = Path(args.ledger) if args.ledger else change_dir / "analysis" / "execution-baseline-ledger.json"
    record = Path(args.record) if args.record else change_dir / "execution-baseline.json"
    try:
        draft(root, args.change, args.source, ledger, record)
    except (AdoptionError, OSError, ValueError) as error:
        print(f"[execution-baseline] {error}")
        return 1
    print(f"[execution-baseline] drafted {ledger} and {record}; reviews are intentionally unbound")
    return 0


def validate(change_dir: Path, root: Path, persisted: str, change_id: str) -> dict | None:
    record_path = change_dir / "execution-baseline.json"
    try:
        record_mode = record_path.lstat().st_mode
    except FileNotFoundError:
        return None
    if not stat.S_ISREG(record_mode):
        raise AdoptionError("execution-baseline record is not a regular file")
    # 신원은 디렉터리 이름이 아니라 **게이트가 요청받은 id** 다 (task 6.2). 아카이브는
    # `changes/<id>` 를 `changes/archive/<YYYY-MM-DD>-<id>` 로 옮겨 이름을 바꾸고, 그
    # 문법으로 id 에서 디렉터리를 찾는 일은 호출자의 해소기가 이미 했다. 여기서 그 문법을
    # 다시 배우면 같은 규칙이 사는 집이 셋이 된다.
    if change_id != CHANGE or persisted != P:
        raise AdoptionError("execution-baseline adoption is not allowed for this change/base")
    if subprocess.run(["git", "symbolic-ref", "-q", "HEAD"], cwd=root, capture_output=True, check=False).returncode == 0:
        raise AdoptionError("adoption requires detached HEAD")
    head = full_commit(root, _git(root, "rev-parse", "HEAD").stdout.strip())
    base_relative = (change_dir.relative_to(root) / "base-commit.txt").as_posix()
    if _regular_committed(root, base_relative, head) != (P + "\n").encode("ascii"):
        raise AdoptionError("planning base file is substituted")
    record = strict_json(_regular_committed(root, record_path.relative_to(root).as_posix(), head))
    required = {"schema", "change", "planning_base", "execution_base", "source_commit", "source_tree", "pre_edit_provenance", "ledger_path", "ledger_sha256", "adversarial_review_path", "adversarial_review_sha256", "gstack_review_path", "gstack_review_sha256", "inherited_history_disposition"}
    if set(record) != required:
        raise AdoptionError("invalid execution-baseline record")
    _schema_one(record["schema"], "execution-baseline record")
    # 기록은 옮기기 **전** 자리의 경로를 적는다(`draft` 가 `openspec/changes/<id>/…` 로
    # 쓴다). 그 경로가 이 change 의 분석 안인지는 **적힌 자리**로 판정하고, 바이트는
    # **지금 자리**에서 읽는다. 아카이브는 내용을 안 바꾸므로 아래 digest 가 그대로 묶는다.
    canonical = f"openspec/changes/{CHANGE}"
    here = change_dir.relative_to(root).as_posix()
    local_analysis = canonical + "/analysis/"
    if not all(isinstance(record[key], str) for key in required - {"schema"}):
        raise AdoptionError("execution-baseline record field types are invalid")
    if record["change"] != CHANGE or record["planning_base"] != P or record["execution_base"] != E or record["pre_edit_provenance"] != "retrospective-exception":
        raise AdoptionError("invalid execution-baseline record")
    if record["inherited_history_disposition"] != "committed historical work; missing original analysis remains debt":
        raise AdoptionError("inherited history disposition is invalid")
    located: dict[str, str] = {}
    for field in ("ledger_path", "adversarial_review_path", "gstack_review_path"):
        located[field] = here + _change_analysis_path(record[field], local_analysis, field)[len(canonical):]
    for field in ("ledger_sha256", "adversarial_review_sha256", "gstack_review_sha256"):
        _sha256(record[field], field)
    if record["adversarial_review_path"] == record["gstack_review_path"]:
        raise AdoptionError("adoption review paths must be distinct")
    source = full_commit(root, record["source_commit"])
    if not isinstance(record["source_tree"], str) or len(record["source_tree"]) != 40 or any(char not in "0123456789abcdef" for char in record["source_tree"]):
        raise AdoptionError("source tree must be a full lowercase SHA-1")
    ancestry(root, P, E)
    ancestry(root, E, source)
    ancestry(root, source, head, strict=True)
    if _git(root, "rev-parse", source + "^{tree}").stdout.strip() != record["source_tree"]:
        raise AdoptionError("source tree mismatch")
    verify_source_go_lock(root, source, head)
    for args in (("diff", "--quiet"), ("diff", "--cached", "--quiet")):
        if subprocess.run(["git", *args], cwd=root, capture_output=True, check=False).returncode:
            raise AdoptionError("adoption requires a clean worktree")
    for path in nul_paths(root, "diff", "--name-only", source, head):
        if not (path.startswith("openspec/") or path.startswith("docs/pm/")):
            raise AdoptionError(f"source-to-evidence drift: {path}")
    candidates: set[str] = set(untracked_filesystem_entries(root))
    for path in sorted(candidates):
        item = root / path
        if not _allowed_metadata(path) or not stat.S_ISREG(item.lstat().st_mode) or os.access(item, os.X_OK):
            raise AdoptionError(f"untracked/ignored input is not allowed: {path}")
        candidates.add(path)
    inputs = go_inputs(root) | go_inputs(root, "tossos_testseams")
    overlap = sorted(candidates & inputs)
    if overlap:
        raise AdoptionError(f"untracked/ignored Go build input: {overlap[0]}")
    ledger_raw = _regular_committed(root, located["ledger_path"], head)
    if hashlib.sha256(ledger_raw).hexdigest() != record["ledger_sha256"]:
        raise AdoptionError("ledger digest mismatch")
    for name in ("adversarial_review", "gstack_review"):
        if hashlib.sha256(_regular_committed(root, located[name + "_path"], head)).hexdigest() != record[name + "_sha256"]:
            raise AdoptionError(f"{name} digest mismatch")
    ledger = strict_json(ledger_raw)
    ledger_required = {"schema", "planning_to_execution", "execution_to_source", "inherited_history_disposition"}
    if set(ledger) != ledger_required:
        raise AdoptionError("invalid execution-baseline ledger")
    _schema_one(ledger["schema"], "execution-baseline ledger")
    if ledger["inherited_history_disposition"] != "committed historical work; missing original analysis remains debt":
        raise AdoptionError("ledger inherited history disposition is invalid")
    if not isinstance(ledger["planning_to_execution"], dict) or not isinstance(ledger["execution_to_source"], dict):
        raise AdoptionError("execution-baseline ledger range types are invalid")
    expected_planning = range_payload(root, P, E)
    if ledger["planning_to_execution"] != expected_planning:
        raise AdoptionError("planning history mismatch")
    recorded_paths = ledger.get("execution_to_source", {}).get("go_paths")
    expected_paths = sorted(path for path in nul_paths(root, "diff", "--name-only", E, source) if path.endswith(".go"))
    if recorded_paths != expected_paths:
        raise AdoptionError("execution-to-source Go path inventory mismatch")
    if any(path not in {
        "cmd/tossctl/soak.go", "cmd/tossctl/soak_test.go", "internal/soak/attest.go",
        "internal/soak/attest_test.go", "internal/soak/renewal_status.go",
        "internal/soak/renewal_status_unix.go", "internal/soak/renewal_status_other.go",
        "internal/soak/renewal_status_test.go", "internal/soak/renewal_status_unix_test.go",
        "internal/console/data.go", "internal/console/templates.go", "internal/console/console_test.go",
    } for path in expected_paths):
        raise AdoptionError("execution-to-source Go path is outside the fixed allowlist")
    expected_execution = {**range_payload(root, E, source), "go_paths": expected_paths}
    if ledger.get("execution_to_source") != expected_execution:
        raise AdoptionError("execution-to-source function inventory mismatch")
    exact_inventory(ledger["planning_to_execution"]["function_inventory"], expected_planning["function_inventory"])
    exact_inventory(ledger["execution_to_source"]["function_inventory"], expected_execution["function_inventory"])
    return {"effective_base": E, "source": source, "ledger": ledger}


if __name__ == "__main__":
    raise SystemExit(main())
