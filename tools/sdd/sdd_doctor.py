#!/usr/bin/env python3
"""Report whether the Full SDD toolchain and repository wiring are usable."""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TOOLS = {
    "rtk": ("rtk", "--version"),
    "openspec": ("openspec", "--version"),
    "codegraph": ("codegraph", "--version"),
    "codegraphcontext": ("codegraphcontext", "--help"),
    "create-context-graph": ("create-context-graph", "--version"),
    "ast-grep": ("ast-grep", "--version"),
    "gbrain": ("gbrain", "--version"),
    "docker": ("docker", "--version"),
}
FILES = (
    ".mcp.json",
    ".codex/config.toml",
    ".codex/hooks.json",
    ".claude/settings.json",
    ".githooks/pre-commit",
    ".githooks/post-commit",
    ".gbrain-source",
    "tools/sdd/gbrain_project.py",
    "tools/sdd-history/refresh_indexes.py",
)


def command_status(name: str, command: tuple[str, ...]) -> dict:
    binary = shutil.which(command[0])
    if not binary:
        return {"ok": False, "detail": f"missing; install {name}"}
    try:
        proc = subprocess.run(command, capture_output=True, text=True, timeout=8, check=False)
        first = (proc.stdout or proc.stderr).strip().splitlines()
        return {
            "ok": proc.returncode == 0,
            "detail": first[0] if first else binary,
        }
    except Exception as exc:
        return {"ok": False, "detail": str(exc)}


def skill_status() -> dict[str, dict]:
    home = Path.home()
    candidates = {
        "gstack": (home / ".agents" / "skills" / "gstack", home / ".claude" / "skills" / "gstack"),
        "superpowers": (
            home / ".codex" / ".tmp" / "plugins" / "plugins" / "superpowers",
            home / ".claude" / "plugins" / "cache" / "superpowers-dev" / "superpowers",
        ),
    }
    result = {}
    for name, roots in candidates.items():
        matches = [str(path) for root in roots for path in ([root] if root.exists() else root.glob("*")) if path.exists()]
        result[name] = {"ok": bool(matches), "detail": matches[0] if matches else "not installed"}
    return result


def service_status(name: str) -> dict:
    if not shutil.which("docker"):
        return {"ok": False, "detail": "docker missing", "required": False}
    try:
        proc = subprocess.run(
            ["docker", "inspect", "--format", "{{.State.Status}}", name],
            capture_output=True,
            text=True,
            timeout=8,
            check=False,
        )
    except Exception as exc:
        return {
            "ok": False,
            "detail": f"probe unavailable: {exc}",
            "required": False,
        }
    return {
        "ok": proc.returncode == 0,
        "detail": proc.stdout.strip() if proc.returncode == 0 else "not found",
        "required": False,
    }


def _inside(path: Path, root: Path) -> bool:
    return path == root or root in path.parents


def _normalized_path(path: str | Path) -> Path:
    """Collapse lexical dot segments without resolving any symlinks."""
    return Path(os.path.abspath(path))


def _resolved_detail(path: Path) -> str:
    try:
        return str(path.resolve(strict=True))
    except (OSError, RuntimeError, ValueError) as exc:
        return f"unresolved ({exc})"


def _typedb_pin(root: Path) -> tuple[str | None, str | None]:
    requirements = root / "tools" / "sdd" / "requirements.txt"
    try:
        lines = requirements.read_text(encoding="utf-8").splitlines()
    except (OSError, UnicodeDecodeError) as exc:
        return None, f"cannot read requirements: {exc}"
    pins = []
    exact = re.compile(r"typedb-driver==([A-Za-z0-9][A-Za-z0-9._+-]*)$")
    for raw_line in lines:
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        match = exact.fullmatch(line)
        if match:
            pins.append(match.group(1))
        elif line.startswith("typedb-driver"):
            return None, "requirements typedb-driver pin is not exact"
    if not pins:
        return None, "requirements typedb-driver pin is missing"
    if len(pins) != 1:
        return None, "requirements typedb-driver pin is multiple"
    return pins[0], None


def _external_python(root: Path, raw: str) -> tuple[Path | None, str, str | None]:
    if not raw:
        return None, "unresolved", "selection is empty"
    raw_candidate = Path(raw)
    if not raw_candidate.is_absolute():
        return None, "unresolved", "selection is not absolute"
    try:
        candidate = _normalized_path(raw)
        lexical_root = _normalized_path(root)
        canonical_root = root.resolve()
    except (OSError, RuntimeError, ValueError) as exc:
        return None, f"unresolved ({exc})", "repository root cannot be resolved"
    if _inside(candidate, lexical_root) or _inside(candidate, canonical_root):
        return None, _resolved_detail(candidate), "selection is lexically inside repository"
    try:
        resolved = candidate.resolve(strict=True)
    except (OSError, RuntimeError, ValueError) as exc:
        return None, f"unresolved ({exc})", "selection cannot be resolved"
    if _inside(resolved, canonical_root):
        return None, str(resolved), "selection resolves inside repository"
    if not candidate.is_file() or not os.access(candidate, os.X_OK):
        return None, str(resolved), "selection is not an executable file"
    return candidate, str(resolved), None


def _driver_status(root: Path) -> dict:
    if "SDD_PYTHON" not in os.environ:
        selected = root / ".sdd" / ".venv" / "bin" / "python"
        resolved = _resolved_detail(selected)
        if selected.exists():
            status = command_status(
                "typedb-driver",
                (
                    str(selected),
                    "-c",
                    "import importlib.metadata as m; print('typedb-driver ' + m.version('typedb-driver'))",
                ),
            )
        else:
            status = {"ok": False, "detail": "run `make sdd-infra`"}
        return {
            "ok": status["ok"],
            "detail": f"mode=local raw={selected} resolved={resolved}; {status['detail']}",
        }

    raw = os.environ["SDD_PYTHON"]
    selected, resolved, selection_error = _external_python(root, raw)
    raw_detail = raw if raw else "<empty>"
    prefix = f"mode=external raw={raw_detail} resolved={resolved}; "
    if selection_error:
        return {"ok": False, "detail": prefix + f"invalid selection: {selection_error}"}
    pin, pin_error = _typedb_pin(root)
    if pin_error:
        return {"ok": False, "detail": prefix + pin_error}
    status = command_status(
        "typedb-driver",
        (
            str(selected),
            "-c",
            "import importlib.metadata as m; print('typedb-driver ' + m.version('typedb-driver'))",
        ),
    )
    expected = f"typedb-driver {pin}"
    if not status["ok"] or status["detail"] != expected:
        return {"ok": False, "detail": prefix + f"expected typedb-driver=={pin}; observed {status['detail']}"}
    return {"ok": True, "detail": prefix + f"{expected}"}


def report(root: Path = ROOT) -> dict:
    tools = {name: command_status(name, command) for name, command in TOOLS.items()}
    typedb_driver = _driver_status(root)
    files = {
        value: {"ok": (root / value).exists(), "detail": "present" if (root / value).exists() else "missing"}
        for value in FILES
    }
    return {
        "tools": tools,
        "skills": skill_status(),
        "python_modules": {"typedb-driver": typedb_driver},
        "files": files,
        "services": {
            "typedb(shared)": service_status("stockos-sdd-typedb"),
            "neo4j(shared)": service_status("infra-neo4j-1"),
        },
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--root", default=str(ROOT))
    args = parser.parse_args()
    result = report(Path(args.root))
    if args.json:
        print(json.dumps(result, ensure_ascii=False, indent=2))
    else:
        for group, values in result.items():
            print(f"[{group}]")
            for name, status in values.items():
                marker = "OK" if status["ok"] else "WARN" if status.get("required") is False else "FAIL"
                print(f"  {marker:4} {name}: {status['detail']}")
    required = [
        status["ok"]
        for group in ("tools", "skills", "python_modules", "files")
        for status in result[group].values()
    ]
    return 0 if all(required) else 1


if __name__ == "__main__":
    raise SystemExit(main())
