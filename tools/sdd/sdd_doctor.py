#!/usr/bin/env python3
"""Report whether the Full SDD toolchain and repository wiring are usable."""

from __future__ import annotations

import argparse
import json
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


def report(root: Path = ROOT) -> dict:
    tools = {name: command_status(name, command) for name, command in TOOLS.items()}
    sdd_python = root / ".sdd" / ".venv" / "bin" / "python"
    if sdd_python.exists():
        typedb_driver = command_status(
            "typedb-driver",
            (
                str(sdd_python),
                "-c",
                "import importlib.metadata as m; print('typedb-driver ' + m.version('typedb-driver'))",
            ),
        )
    else:
        typedb_driver = {"ok": False, "detail": "run `make sdd-infra`"}
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
