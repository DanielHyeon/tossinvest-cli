#!/usr/bin/env python3
"""Persist a small, private Codex handoff without blocking tool execution."""

from __future__ import annotations

import fcntl
import json
import os
import re
import subprocess
import sys
import tempfile
import uuid
from collections import deque
from datetime import datetime, timezone
from pathlib import Path
from typing import BinaryIO


CONTEXT_DIRECTORY = ".codex-context"
LATEST_NAME = "session-summary.md"
LOCK_NAME = ".save-session.lock"
BACKUP_LIMIT = 5
MESSAGE_LIMIT = 30
MESSAGE_CHARACTER_LIMIT = 2_000
TRANSCRIPT_BYTE_LIMIT = 2 * 1024 * 1024
TRANSCRIPT_LINE_LIMIT = 256 * 1024
GIT_OUTPUT_LIMIT = 12_000
REDACTION = "[REDACTED]"

SECRET_PATTERNS = (
    re.compile(
        r"(?i)((?:[\"'])?\b(?:[a-z0-9]+[_-])*(?:api[_-]?key|access[_-]?key|"
        r"access[_-]?token|auth[_-]?token|password|passwd|client[_-]?secret|"
        r"private[_-]?key|secret|token)\b(?:[\"'])?\s*[:=]\s*)"
        r"(?:\"[^\"\n]*\"|'[^'\n]*'|[^\s,;]+)"
    ),
    re.compile(r"(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{12,}"),
    re.compile(r"\bsk-(?:proj-)?[A-Za-z0-9_-]{16,}"),
    re.compile(r"\bgh[pousr]_[A-Za-z0-9]{20,}"),
    re.compile(r"\bAKIA[0-9A-Z]{16}\b"),
    re.compile(r"\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b"),
)


def redact(value: str) -> str:
    """자주 쓰는 시크릿 모양을 저장 전에 가린다."""
    result = value
    for index, pattern in enumerate(SECRET_PATTERNS):
        if index == 0:
            result = pattern.sub(lambda match: match.group(1) + REDACTION, result)
        else:
            result = pattern.sub(REDACTION, result)
    return result


def bounded(value: str, limit: int = MESSAGE_CHARACTER_LIMIT) -> str:
    cleaned = redact(value.replace("\x00", "")).strip()
    if len(cleaned) <= limit:
        return cleaned
    return cleaned[:limit].rstrip() + "\n… [truncated]"


def read_payload() -> dict:
    if sys.stdin.isatty():
        return {}
    try:
        value = json.load(sys.stdin)
    except Exception:
        return {}
    return value if isinstance(value, dict) else {}


def command_cwd(payload: dict) -> Path:
    raw = payload.get("cwd")
    if isinstance(raw, str) and raw:
        try:
            candidate = Path(raw).resolve()
            if candidate.is_dir():
                return candidate
        except OSError:
            pass
    return Path.cwd().resolve()


def repository_root(payload: dict) -> Path:
    cwd = command_cwd(payload)
    try:
        result = subprocess.run(
            ["git", "-C", str(cwd), "rev-parse", "--show-toplevel"],
            check=False,
            capture_output=True,
            text=True,
            timeout=3,
        )
        if result.returncode == 0 and result.stdout.strip():
            root = Path(result.stdout.strip()).resolve()
            if root.is_dir():
                return root
    except (OSError, subprocess.SubprocessError):
        pass
    return cwd


def forbidden_path(candidate: Path) -> bool:
    """다른 에이전트가 소유한 저장소는 열지 않는다."""
    return candidate.name == "save-session.sh" or any(
        part in {".claude", ".ai-context"} for part in candidate.parts
    )


def transcript_file(payload: dict, root: Path) -> Path | None:
    raw = payload.get("transcript_path")
    if not isinstance(raw, str) or not raw:
        return None
    try:
        candidate = Path(raw)
        if not candidate.is_absolute():
            candidate = root / candidate
        candidate = candidate.resolve(strict=True)
        if forbidden_path(candidate) or not candidate.is_file():
            return None
        return candidate
    except (OSError, RuntimeError):
        return None


def recent_transcript_bytes(path: Path) -> bytes:
    with path.open("rb") as handle:
        size = handle.seek(0, os.SEEK_END)
        offset = max(0, size - TRANSCRIPT_BYTE_LIMIT)
        handle.seek(offset)
        if offset:
            handle.readline(TRANSCRIPT_LINE_LIMIT + 1)
        return handle.read(TRANSCRIPT_BYTE_LIMIT)


def message_text(payload: dict) -> str:
    content = payload.get("content")
    if isinstance(content, str):
        return content
    if not isinstance(content, list):
        return ""
    chunks: list[str] = []
    for item in content:
        if not isinstance(item, dict):
            continue
        if item.get("type") not in {"input_text", "output_text", "text"}:
            continue
        text = item.get("text")
        if isinstance(text, str) and text:
            chunks.append(text)
    return "\n".join(chunks)


def transcript_messages(path: Path | None) -> list[tuple[str, str]]:
    if path is None:
        return []
    messages: deque[tuple[str, str]] = deque(maxlen=MESSAGE_LIMIT)
    try:
        data = recent_transcript_bytes(path).decode("utf-8", errors="replace")
    except OSError:
        return []
    for raw_line in data.splitlines():
        if not raw_line or len(raw_line) > TRANSCRIPT_LINE_LIMIT:
            continue
        try:
            record = json.loads(raw_line)
        except (json.JSONDecodeError, UnicodeDecodeError):
            continue
        if not isinstance(record, dict) or record.get("type") != "response_item":
            continue
        payload = record.get("payload")
        if not isinstance(payload, dict) or payload.get("type") != "message":
            continue
        role = payload.get("role")
        if role not in {"user", "assistant"}:
            continue
        text = bounded(message_text(payload))
        if text:
            messages.append((role, text))
    return list(messages)


def run_git(root: Path, *arguments: str) -> str:
    try:
        result = subprocess.run(
            ["git", "-C", str(root), *arguments],
            check=False,
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
            timeout=3,
        )
    except (OSError, subprocess.SubprocessError):
        return ""
    if result.returncode != 0:
        return ""
    return redact(result.stdout[:GIT_OUTPUT_LIMIT].strip())


def repository_pathspecs(transcript_relative: str | None = None) -> tuple[str, ...]:
    exclusions = [
        ":(exclude).claude/**",
        ":(exclude).ai-context/**",
        ":(exclude).codex-context/**",
        ":(exclude)save-session.sh",
    ]
    if transcript_relative:
        exclusions.append(f":(exclude,literal){transcript_relative}")
    return ("--", ".", *exclusions)


def excluded_repository_name(name: str, transcript_relative: str | None = None) -> bool:
    normalized = name.replace("\\", "/")
    excluded = (
        ".codex-context/",
        ".claude/",
        ".ai-context/",
    )
    if normalized == "save-session.sh" or any(token in normalized for token in excluded):
        return True
    return bool(transcript_relative and transcript_relative in normalized)


def git_status(root: Path, transcript_relative: str | None) -> str:
    output = run_git(
        root,
        "status",
        "--short",
        "--untracked-files=normal",
        *repository_pathspecs(transcript_relative),
    )
    lines = [
        line
        for line in output.splitlines()
        if not excluded_repository_name(line, transcript_relative)
    ]
    return "\n".join(lines) or "clean"


def git_statistics(root: Path, transcript_relative: str | None) -> tuple[str, str]:
    pathspecs = repository_pathspecs(transcript_relative)
    unstaged = run_git(root, "diff", "--stat", *pathspecs) or "none"
    staged = run_git(root, "diff", "--cached", "--stat", *pathspecs) or "none"
    return unstaged, staged


def recent_files(root: Path, transcript_relative: str | None) -> list[str]:
    output = run_git(root, "ls-files", "-z", *repository_pathspecs(transcript_relative))
    candidates: list[tuple[float, str]] = []
    for relative in output.split("\x00"):
        if not relative or excluded_repository_name(relative, transcript_relative):
            continue
        try:
            path = (root / relative).resolve(strict=True)
            if forbidden_path(path) or not path.is_file():
                continue
            candidates.append((path.stat().st_mtime, relative))
        except (OSError, RuntimeError):
            continue
    candidates.sort(reverse=True)
    return [redact(relative) for _, relative in candidates[:10]]


def indent_block(value: str) -> list[str]:
    return [f"    {line}" if line else "    " for line in value.splitlines()]


def render_summary(payload: dict, root: Path) -> str:
    transcript = transcript_file(payload, root)
    transcript_relative: str | None = None
    if transcript is not None:
        try:
            transcript_relative = transcript.relative_to(root).as_posix()
        except ValueError:
            transcript_relative = None

    session = payload.get("session_id")
    model = payload.get("model")
    saved_at = datetime.now(timezone.utc).isoformat(timespec="seconds")
    status = git_status(root, transcript_relative)
    unstaged, staged = git_statistics(root, transcript_relative)
    commits = run_git(root, "log", "-5", "--date=short", "--pretty=format:%h %ad %s")
    files = recent_files(root, transcript_relative)
    messages = transcript_messages(transcript)

    lines = [
        "# Codex Session Handoff",
        "",
        f"- Saved: {saved_at}",
        f"- Session: {bounded(session, 200) if isinstance(session, str) else 'unavailable'}",
        f"- Model: {bounded(model, 200) if isinstance(model, str) else 'unavailable'}",
        "",
        "## Git working state",
        "",
        *indent_block(status),
        "",
        "### Unstaged statistics",
        "",
        *indent_block(unstaged),
        "",
        "### Staged statistics",
        "",
        *indent_block(staged),
        "",
        "## Recent commits",
        "",
        *indent_block(commits or "unavailable"),
        "",
        "## Recent repository files",
        "",
    ]
    lines.extend(f"- {name}" for name in files)
    if not files:
        lines.append("- unavailable")
    lines.extend(["", "## Recent Codex conversation", ""])
    if not messages:
        lines.append("_No readable user/assistant messages._")
    for role, text in messages:
        lines.extend([f"### {role.title()}", "", *indent_block(text), ""])
    return "\n".join(lines).rstrip() + "\n"


def acquire_lock(context: Path) -> BinaryIO | None:
    if context.is_symlink():
        raise OSError("context directory cannot be a symbolic link")
    context.mkdir(parents=True, exist_ok=True, mode=0o700)
    try:
        context.chmod(0o700)
    except OSError:
        pass
    lock_path = context / LOCK_NAME
    handle = lock_path.open("a+b")
    try:
        lock_path.chmod(0o600)
    except OSError:
        pass
    try:
        fcntl.flock(handle.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
    except (BlockingIOError, OSError):
        handle.close()
        return None
    return handle


def release_lock(handle: BinaryIO) -> None:
    try:
        fcntl.flock(handle.fileno(), fcntl.LOCK_UN)
    finally:
        handle.close()


def atomic_write(path: Path, content: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
    try:
        path.parent.chmod(0o700)
    except OSError:
        pass
    temporary: Path | None = None
    try:
        with tempfile.NamedTemporaryFile(
            mode="wb",
            dir=path.parent,
            prefix=f".{path.name}.",
            suffix=".tmp",
            delete=False,
        ) as handle:
            temporary = Path(handle.name)
            os.chmod(temporary, 0o600)
            handle.write(content)
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(temporary, path)
        temporary = None
        try:
            descriptor = os.open(path.parent, os.O_RDONLY)
            try:
                os.fsync(descriptor)
            finally:
                os.close(descriptor)
        except OSError:
            pass
    finally:
        if temporary is not None:
            temporary.unlink(missing_ok=True)


def backup_name() -> str:
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S.%fZ")
    return f"session-summary-{timestamp}-{uuid.uuid4().hex}.md"


def prune_backups(backups: Path) -> None:
    existing = sorted(backups.glob("session-summary-*.md"), reverse=True)
    for obsolete in existing[BACKUP_LIMIT:]:
        obsolete.unlink(missing_ok=True)


def publish(context: Path, summary: str) -> None:
    latest = context / LATEST_NAME
    backups = context / "backups"
    if latest.is_symlink() or backups.is_symlink():
        raise OSError("session output cannot be a symbolic link")
    if latest.is_file():
        previous = latest.read_bytes()
        atomic_write(backups / backup_name(), previous)
    atomic_write(latest, summary.encode("utf-8"))
    if backups.is_dir():
        prune_backups(backups)


def run(payload: dict | None = None, root: Path | None = None) -> int:
    try:
        hook_payload = read_payload() if payload is None else payload
        if not isinstance(hook_payload, dict):
            hook_payload = {}
        project = root.resolve() if root is not None else repository_root(hook_payload)
        context = project / CONTEXT_DIRECTORY
        lock = acquire_lock(context)
        if lock is None:
            return 0
        try:
            publish(context, render_summary(hook_payload, project))
        finally:
            release_lock(lock)
    except Exception as exc:
        # Hook 실패가 원래 도구 결과를 막지 않도록 예외 형식만 짧게 알린다.
        print(
            f"[codex-session-save] warning: save skipped ({type(exc).__name__})",
            file=sys.stderr,
        )
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
