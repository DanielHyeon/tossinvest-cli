#!/usr/bin/env python3
"""File-based episodic memory for TossOS SDD."""

from __future__ import annotations

import argparse
import copy
import fcntl
import json
import os
import re
import sqlite3
import tempfile
from contextlib import contextmanager
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DEFAULT_MEMORY_ROOT = ROOT / "docs" / "memory"
REQUIRED = ("id", "change_id", "status", "created_at", "summary")
FORBIDDEN = ("password", "secret", "api_key", "authorization:", "cookie:")
SECRET_SHAPES = (
    re.compile(r"(?i)\bbearer\s+[A-Za-z0-9._~+/=-]+"),
    re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----"),
    re.compile(r"(?<![\w.+-])[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}(?![\w.-])"),
    re.compile(r"\b[A-Za-z0-9_-]{32,}\b"),
)
SECRET_ASSIGNMENT = re.compile(
    r"(?i)\b(api[-_]?key|token|secret|password|authorization|cookie|session|private[-_]?key)"
    r"\s*[:=]\s*(?:\"[^\"]*\"|'[^']*'|[^\s,;]+)"
)


def sensitive_text_errors(text: str) -> list[str]:
    lowered = text.lower()
    errors = [f"forbidden sensitive marker: {word}" for word in FORBIDDEN if word in lowered]
    if SECRET_ASSIGNMENT.search(text):
        errors.append("forbidden sensitive assignment")
    errors.extend(
        f"forbidden sensitive shape: {pattern.pattern}"
        for pattern in SECRET_SHAPES
        if pattern.search(text)
    )
    return errors


def parse_frontmatter(text: str) -> tuple[dict, str]:
    if not text.startswith("---\n"):
        return {}, text
    _, frontmatter, body = text.split("---", 2)
    data: dict[str, object] = {}
    for raw in frontmatter.splitlines():
        if ":" not in raw:
            continue
        key, value = raw.split(":", 1)
        value = value.strip()
        if value == "[]":
            data[key.strip()] = []
        elif value:
            data[key.strip()] = value.strip("\"'")
    return data, body.strip()


def validate_episode(path: Path) -> list[str]:
    try:
        text = path.read_text(encoding="utf-8")
    except OSError as exc:
        return [f"read failed: {exc}"]
    meta, _ = parse_frontmatter(text)
    errors = [f"missing field: {name}" for name in REQUIRED if not meta.get(name)]
    errors.extend(sensitive_text_errors(text))
    if meta.get("status") not in ("episodic", "canonical"):
        errors.append("status must be episodic or canonical")
    return errors


def ledger_path(memory_root: Path) -> Path:
    return memory_root / "ledger.jsonl"


@contextmanager
def ledger_lock(memory_root: Path):
    memory_root.mkdir(parents=True, exist_ok=True)
    lock_path = memory_root / ".ledger.lock"
    with lock_path.open("a+", encoding="utf-8") as handle:
        fcntl.flock(handle.fileno(), fcntl.LOCK_EX)
        try:
            yield
        finally:
            fcntl.flock(handle.fileno(), fcntl.LOCK_UN)


def load_ledger(memory_root: Path = DEFAULT_MEMORY_ROOT) -> list[dict]:
    path = ledger_path(memory_root)
    if not path.exists():
        return []
    records = []
    for line in path.read_text(encoding="utf-8").splitlines():
        if line.strip():
            record = json.loads(line)
            raw = Path(record.get("source_path", ""))
            if raw.is_absolute():
                continue
            try:
                safe_source_path(memory_root / raw, memory_root)
            except ValueError:
                continue
            records.append(record)
    return records


def write_ledger(records: list[dict], memory_root: Path) -> None:
    path = ledger_path(memory_root)
    path.parent.mkdir(parents=True, exist_ok=True)
    content = "".join(
        json.dumps(record, ensure_ascii=False, sort_keys=True) + "\n"
        for record in records
    )
    with tempfile.NamedTemporaryFile(
        "w",
        encoding="utf-8",
        dir=path.parent,
        prefix=".ledger-",
        delete=False,
    ) as handle:
        handle.write(content)
        temporary = Path(handle.name)
    temporary.replace(path)


def safe_source_path(path: Path, memory_root: Path) -> tuple[Path, str]:
    resolved = path.resolve()
    root = memory_root.resolve()
    allowed = ((root / "episodes").resolve(), (root / "canonical").resolve())
    if not any(resolved.is_relative_to(directory) for directory in allowed):
        raise ValueError("memory source must stay under episodes/ or canonical/")
    return resolved, resolved.relative_to(root).as_posix()


def retain(path: Path, memory_root: Path = DEFAULT_MEMORY_ROOT) -> dict:
    try:
        resolved, source_path = safe_source_path(path, memory_root)
    except ValueError as exc:
        return {"ok": False, "errors": [str(exc)]}
    if not resolved.is_relative_to((memory_root.resolve() / "episodes").resolve()):
        return {"ok": False, "errors": ["retain accepts episodes/ sources only"]}
    errors = validate_episode(resolved)
    if errors:
        return {"ok": False, "errors": errors}
    meta, _ = parse_frontmatter(resolved.read_text(encoding="utf-8"))
    if meta.get("status") != "episodic":
        return {"ok": False, "errors": ["retained source status must be episodic"]}
    record = {
        "id": str(meta["id"]),
        "project": "tossos",
        "change_id": str(meta["change_id"]),
        "status": "episodic",
        "summary": str(meta["summary"]),
        "created_at": str(meta["created_at"]),
        "source_path": source_path,
        "promotion_evidence": None,
    }
    with ledger_lock(memory_root):
        existing = [
            item
            for item in load_ledger(memory_root)
            if item.get("id") == record["id"]
        ]
        if any(item.get("status") == "canonical" for item in existing):
            return {
                "ok": False,
                "errors": ["canonical memory id cannot be replaced or demoted"],
            }
        records = [
            item for item in load_ledger(memory_root) if item.get("id") != record["id"]
        ]
        records.append(record)
        write_ledger(records, memory_root)
        _rebuild(memory_root)
    return {"ok": True, "id": record["id"]}


def promote(record_id: str, evidence: str, memory_root: Path = DEFAULT_MEMORY_ROOT) -> dict:
    if not evidence.strip():
        return {"ok": False, "errors": ["promotion evidence is required"]}
    if sensitive_text_errors(evidence):
        return {"ok": False, "errors": ["promotion evidence contains sensitive data"]}
    with ledger_lock(memory_root):
        records = load_ledger(memory_root)
        for record in records:
            if record.get("id") == record_id:
                if record.get("status") != "episodic":
                    return {"ok": False, "errors": ["only episodic memory can be promoted"]}
                try:
                    raw = Path(record.get("source_path", ""))
                    if raw.is_absolute():
                        raise ValueError("absolute memory source paths are forbidden")
                    source, _ = safe_source_path(memory_root / raw, memory_root)
                    canonical_dir = memory_root.resolve() / "canonical"
                    if canonical_dir.is_symlink():
                        raise ValueError("canonical memory directory cannot be a symlink")
                    canonical_dir.mkdir(parents=True, exist_ok=True)
                    canonical_dir = canonical_dir.resolve()
                    if not canonical_dir.is_relative_to(memory_root.resolve()):
                        raise ValueError("canonical memory directory escapes memory root")
                    target = canonical_dir / source.name
                    resolved_target = target.resolve(strict=False)
                    if not resolved_target.is_relative_to(canonical_dir):
                        raise ValueError("canonical target escapes canonical memory directory")
                    if target.exists() or target.is_symlink():
                        raise ValueError(f"canonical target exists: {target.name}")
                except (OSError, ValueError) as exc:
                    return {"ok": False, "errors": [str(exc)]}
                text = source.read_text(encoding="utf-8")
                updated, count = re.subn(
                    r"(?m)^status:\s*episodic\s*$",
                    "status: canonical",
                    text,
                    count=1,
                )
                if count != 1:
                    return {"ok": False, "errors": ["episode frontmatter is not episodic"]}
                committed = False
                original_records = copy.deepcopy(records)
                try:
                    with target.open("x", encoding="utf-8") as handle:
                        handle.write(updated)
                    record["status"] = "canonical"
                    record["promotion_evidence"] = evidence
                    record["source_path"] = target.relative_to(
                        memory_root.resolve()
                    ).as_posix()
                    write_ledger(records, memory_root)
                    _rebuild(memory_root)
                    committed = True
                except Exception as exc:
                    if not committed:
                        target.unlink(missing_ok=True)
                    try:
                        write_ledger(original_records, memory_root)
                    except Exception as rollback_exc:
                        return {
                            "ok": False,
                            "errors": [
                                f"promotion failed: {exc}",
                                f"ledger rollback failed: {rollback_exc}",
                            ],
                        }
                    return {"ok": False, "errors": [f"promotion failed: {exc}"]}
                try:
                    source.unlink()
                except OSError as exc:
                    return {
                        "ok": True,
                        "id": record_id,
                        "warnings": [f"episodic source cleanup failed: {exc}"],
                    }
                return {"ok": True, "id": record_id}
    return {"ok": False, "errors": [f"unknown id: {record_id}"]}


def source_body(record: dict, memory_root: Path) -> str:
    try:
        raw = Path(record.get("source_path", ""))
        if raw.is_absolute():
            return ""
        path, _ = safe_source_path(memory_root / raw, memory_root)
        _, body = parse_frontmatter(path.read_text(encoding="utf-8"))
        return body
    except (OSError, ValueError):
        return ""


def check_memory(memory_root: Path = DEFAULT_MEMORY_ROOT) -> list[str]:
    errors: list[str] = []
    records: list[dict] = []
    path = ledger_path(memory_root)
    if path.exists():
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            if not line.strip():
                continue
            try:
                value = json.loads(line)
            except ValueError as exc:
                errors.append(f"ledger line {number}: invalid JSON: {exc}")
                continue
            if not isinstance(value, dict):
                errors.append(f"ledger line {number}: record must be an object")
                continue
            records.append(value)
    seen: set[str] = set()
    source_files: set[str] = set()
    for directory in ("episodes", "canonical"):
        for path in sorted((memory_root / directory).glob("*.md")):
            errors.extend(f"{path}: {error}" for error in validate_episode(path))
            try:
                _, relative = safe_source_path(path, memory_root)
                source_files.add(relative)
            except ValueError as exc:
                errors.append(f"{path}: {exc}")
    referenced: set[str] = set()
    for record in records:
        record_id = str(record.get("id", ""))
        if not record_id or record_id in seen:
            errors.append(f"ledger contains duplicate or empty id: {record_id!r}")
        seen.add(record_id)
        try:
            raw = Path(record.get("source_path", ""))
            if raw.is_absolute():
                raise ValueError("absolute memory source paths are forbidden")
            source, relative = safe_source_path(memory_root / raw, memory_root)
            meta, _ = parse_frontmatter(source.read_text(encoding="utf-8"))
        except (OSError, ValueError) as exc:
            errors.append(f"{record_id}: invalid source: {exc}")
            continue
        referenced.add(relative)
        errors.extend(f"{record_id}: {error}" for error in validate_episode(source))
        if str(meta.get("id", "")) != record_id:
            errors.append(f"{record_id}: source frontmatter id mismatch")
        if meta.get("status") != record.get("status"):
            errors.append(f"{record_id}: source/ledger status mismatch")
    for orphan in sorted(source_files - referenced):
        errors.append(f"{orphan}: memory source is not linked from ledger")
    return errors


def _rebuild(memory_root: Path = DEFAULT_MEMORY_ROOT) -> list[dict]:
    rows = []
    for record in load_ledger(memory_root):
        raw = Path(record.get("source_path", ""))
        try:
            source, _ = safe_source_path(memory_root / raw, memory_root)
        except ValueError:
            continue
        if validate_episode(source):
            continue
        rows.append(
            {
                "id": record.get("id"),
                "status": record.get("status"),
                "change_id": record.get("change_id"),
                "summary": record.get("summary"),
                "source_path": record.get("source_path"),
                "text": f"{record.get('summary', '')}\n{source_body(record, memory_root)}".strip(),
            }
        )
    memory_root.mkdir(parents=True, exist_ok=True)
    index_content = "".join(
        json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n" for row in rows
    )
    temporary_index: Path | None = None
    temporary_database: Path | None = None
    db = None
    backups: list[tuple[Path, Path]] = []
    published: list[Path] = []
    try:
        with tempfile.NamedTemporaryFile(
            "w",
            encoding="utf-8",
            dir=memory_root,
            prefix=".index-",
            delete=False,
        ) as handle:
            temporary_index = Path(handle.name)
            handle.write(index_content)
        descriptor, database_name = tempfile.mkstemp(
            dir=memory_root,
            prefix=".index-",
            suffix=".db",
        )
        temporary_database = Path(database_name)
        os.close(descriptor)
        db = sqlite3.connect(temporary_database)
        db.execute("CREATE VIRTUAL TABLE memory USING fts5(id, status, change_id, summary, text)")
        db.executemany(
            "INSERT INTO memory VALUES (?, ?, ?, ?, ?)",
            [
                (row["id"], row["status"], row["change_id"], row["summary"], row["text"])
                for row in rows
            ],
        )
        db.commit()
        db.close()
        db = None
        outputs = (
            (temporary_index, memory_root / "index.jsonl"),
            (temporary_database, memory_root / "index.db"),
        )
        for _, output in outputs:
            if output.exists():
                backup = memory_root / f".{output.name}.backup-{os.getpid()}-{id(output)}"
                output.replace(backup)
                backups.append((backup, output))
        for temporary, output in outputs:
            temporary.replace(output)
            published.append(output)
    except Exception:
        for output in published:
            output.unlink(missing_ok=True)
        for backup, output in backups:
            if backup.exists():
                backup.replace(output)
        raise
    finally:
        if db is not None:
            db.close()
        if temporary_index is not None:
            temporary_index.unlink(missing_ok=True)
        if temporary_database is not None:
            temporary_database.unlink(missing_ok=True)
        for backup, _ in backups:
            backup.unlink(missing_ok=True)
    return rows


def rebuild(memory_root: Path = DEFAULT_MEMORY_ROOT) -> list[dict]:
    with ledger_lock(memory_root):
        return _rebuild(memory_root)


def search(query: str, memory_root: Path = DEFAULT_MEMORY_ROOT, limit: int = 50) -> list[dict]:
    terms = re.findall(r"\w+", query, flags=re.UNICODE)
    if not terms:
        return []
    database = memory_root / "index.db"
    if not database.exists():
        rebuild(memory_root)
    expression = " AND ".join(f'"{term.replace(chr(34), chr(34) * 2)}"' for term in terms)
    try:
        db = sqlite3.connect(database)
        db.row_factory = sqlite3.Row
        rows = db.execute(
            """
            SELECT id, status, change_id, summary, text
            FROM memory
            WHERE memory MATCH ?
            LIMIT ?
            """,
            (expression, max(1, min(limit, 200))),
        ).fetchall()
        return [dict(row) for row in rows]
    except sqlite3.DatabaseError:
        rebuild(memory_root)
        return search(query, memory_root, limit)
    finally:
        if "db" in locals():
            db.close()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "command",
        choices=("validate", "check", "retain", "promote", "rebuild", "search"),
    )
    parser.add_argument("arg", nargs="?")
    parser.add_argument("--evidence", default="")
    parser.add_argument("--memory-root", default=str(DEFAULT_MEMORY_ROOT))
    args = parser.parse_args()
    root = Path(args.memory_root)
    if args.command == "validate":
        errors = validate_episode(Path(args.arg or ""))
        print("\n".join(errors) if errors else "valid")
        return 1 if errors else 0
    if args.command == "check":
        errors = check_memory(root)
        print("\n".join(errors) if errors else "valid")
        return 1 if errors else 0
    if args.command == "retain":
        result = retain(Path(args.arg or ""), root)
    elif args.command == "promote":
        result = promote(args.arg or "", args.evidence, root)
    elif args.command == "rebuild":
        result = {"ok": True, "count": len(rebuild(root))}
    else:
        result = {"ok": True, "results": search(args.arg or "", root)}
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if result.get("ok") else 1


if __name__ == "__main__":
    raise SystemExit(main())
