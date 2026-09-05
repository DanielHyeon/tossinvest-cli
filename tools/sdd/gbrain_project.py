#!/usr/bin/env python3
"""Run the shared GBrain CLI against TossOS' project-local data home."""

from __future__ import annotations

import fcntl
import json
import os
import shutil
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
GBRAIN_HOME = ROOT / ".sdd" / "gbrain-home"
PROCESS_LOCK = GBRAIN_HOME / ".gbrain" / "tossos-process.lock"
CONFIG_FILE = GBRAIN_HOME / ".gbrain" / "config.json"
BRAIN_PATH = GBRAIN_HOME / ".gbrain" / "brain.pglite"
PGLITE_LOCK = BRAIN_PATH / ".gbrain-lock" / "lock"
BUSY_EXIT = 75
BUSY_MARKER = "[gbrain-project] busy:"
DEFAULT_PGLITE_STEAL_GRACE_SECONDS = 600
SAFE_DIAGNOSTIC_SUBCOMMANDS = frozenset(
    {
        "ask",
        "doctor",
        "init",
        "query",
        "search",
        "serve",
        "sources",
        "sync",
        "version",
    }
)


def project_environment() -> dict[str, str]:
    env = os.environ.copy()
    env["GBRAIN_HOME"] = str(GBRAIN_HOME)
    return env


def read_metadata(path: Path) -> dict[str, object]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
        return value if isinstance(value, dict) else {}
    except (OSError, ValueError):
        return {}


def process_alive(pid: object) -> bool:
    if not isinstance(pid, int) or isinstance(pid, bool) or pid <= 0:
        return False
    try:
        os.kill(pid, 0)
    except ProcessLookupError:
        return False
    except PermissionError:
        return True
    return True


def pglite_steal_grace_seconds() -> int:
    try:
        value = int(os.environ.get("GBRAIN_PGLITE_LOCK_STEAL_GRACE_SECONDS", ""))
        return value if value > 0 else DEFAULT_PGLITE_STEAL_GRACE_SECONDS
    except ValueError:
        return DEFAULT_PGLITE_STEAL_GRACE_SECONDS


def heartbeat_fresh(metadata: dict[str, object]) -> bool:
    refreshed = metadata.get("refreshed_at", metadata.get("acquired_at"))
    if (
        not isinstance(refreshed, (int, float))
        or isinstance(refreshed, bool)
    ):
        return False
    age_ms = time.time() * 1000 - refreshed
    return age_ms <= pglite_steal_grace_seconds() * 1000


def owner_summary(metadata: dict[str, object]) -> str:
    pid = metadata.get("pid", "unknown")
    command_parts = str(metadata.get("command", "unknown")).split()
    subcommand = next(
        (part for part in command_parts if part in SAFE_DIAGNOSTIC_SUBCOMMANDS),
        None,
    )
    command = f"gbrain {subcommand}" if subcommand else "gbrain"
    home = metadata.get("home", str(GBRAIN_HOME))
    return f"owner pid={pid} command={command!r} home={home}"


def legacy_owner() -> dict[str, object] | None:
    metadata = read_metadata(PGLITE_LOCK)
    pid = metadata.get("pid")
    if pid == os.getpid() or not process_alive(pid) or not heartbeat_fresh(metadata):
        return None
    return metadata


def acquire_process_lock(arguments: list[str]) -> int:
    """Acquire the one kernel-owned entry lock for this project GBrain.

    The descriptor is deliberately inheritable so `execve(gbrain)` retains the
    flock. The kernel releases it on every exit path, including SIGKILL.
    """

    PROCESS_LOCK.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
    descriptor = os.open(PROCESS_LOCK, os.O_RDWR | os.O_CREAT, 0o600)
    os.fchmod(descriptor, 0o600)
    try:
        fcntl.flock(descriptor, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError:
        metadata = read_metadata(PROCESS_LOCK)
        os.close(descriptor)
        raise RuntimeError(owner_summary(metadata)) from None

    existing = legacy_owner()
    if existing is not None:
        fcntl.flock(descriptor, fcntl.LOCK_UN)
        os.close(descriptor)
        raise RuntimeError(
            "legacy PGLite "
            + owner_summary(
                {
                    **existing,
                    "home": str(GBRAIN_HOME),
                }
            )
        )

    metadata = {
        "pid": os.getpid(),
        "command": "gbrain " + arguments[0],
        "home": str(GBRAIN_HOME),
    }
    payload = (json.dumps(metadata, ensure_ascii=False, sort_keys=True) + "\n").encode(
        "utf-8"
    )
    os.ftruncate(descriptor, 0)
    os.lseek(descriptor, 0, os.SEEK_SET)
    os.write(descriptor, payload)
    os.fsync(descriptor)
    os.set_inheritable(descriptor, True)
    return descriptor


def align_database_path() -> None:
    """저장된 브레인 경로를 이 체크아웃의 홈으로 되돌린다.

    체크아웃 이름이 바뀌면 `GBRAIN_HOME` 은 `__file__` 에서 다시 계산되어 따라오는데,
    그 홈 **안의** `config.json` 이 쥔 절대경로는 혼자 뒤처진다. 그러면 gbrain 은
    옛 자리에 브레인을 새로 판다. 이것이 조용한 이유는 오류가 아니라 **다른 자리의
    성공**이기 때문이다 (2026-09-05 실발생: 저장소 밖에 42MB 부분 재색인이 생기고
    진짜 459MB 브레인은 다섯 시간 동안 아무도 안 썼다).

    상대경로로는 못 고친다. `database_path` 는 PGLite 로 그대로 넘어가 프로세스 cwd
    기준으로 풀리므로 같은 고장을 더 예측하기 어렵게 만들 뿐이고, 아예 빼면
    in-memory 로 떨어져 조용히 휘발한다. gbrain 은 `GBRAIN_HOME` 자체도 절대경로만
    받는다. 그래서 이 값은 저장하는 대신 **자기 위치를 아는 이 래퍼가 유도한다.**

    이미 있는 값만 고치고 없는 칸은 만들지 않는다. 그 칸의 부재는 in-memory 를
    뜻하고, 그 선택은 이 래퍼가 아니라 `gbrain init` 이 내리는 것이다.
    """

    configured = read_metadata(CONFIG_FILE)
    if "database_path" not in configured:
        return
    if configured["database_path"] == str(BRAIN_PATH):
        return
    configured["database_path"] = str(BRAIN_PATH)
    staged = CONFIG_FILE.with_name(CONFIG_FILE.name + ".tmp")
    staged.write_text(
        json.dumps(configured, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    os.chmod(staged, 0o600)
    os.replace(staged, CONFIG_FILE)


def main() -> int:
    executable = shutil.which("gbrain")
    if executable is None:
        print("gbrain is not installed; run make sdd-doctor", file=sys.stderr)
        return 127
    if len(sys.argv) == 1:
        print("usage: gbrain_project.py <gbrain args...>", file=sys.stderr)
        return 2
    try:
        descriptor = acquire_process_lock(sys.argv[1:])
    except RuntimeError as exc:
        print(f"{BUSY_MARKER} {exc}", file=sys.stderr)
        return BUSY_EXIT
    align_database_path()
    try:
        os.execve(executable, [executable, *sys.argv[1:]], project_environment())
    except OSError as exc:
        print(f"gbrain execution failed: {exc}", file=sys.stderr)
        os.close(descriptor)
        return 126


if __name__ == "__main__":
    raise SystemExit(main())
