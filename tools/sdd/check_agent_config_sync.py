#!/usr/bin/env python3
"""Verify the shared Full SDD contract across Claude, Codex, and root entrypoints."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ".claude/CLAUDE.md"
MIRROR = ".codex/agents.md"
START = "<!-- SDD_SHARED_START -->"
END = "<!-- SDD_SHARED_END -->"
REQUIRED_SECTIONS = (
    "최상위 안전 불변식",
    "계약(OpenSpec)",
    "증거(CodeGraph hard evidence)",
    "증거 보조(CodeGraphContext)",
    "실행(Superpowers TDD)",
    "게이트(gstack)",
    "파일기반 episodic",
    "Function Logic Map",
    "SDD Control Graph",
    "PM 조정 계층",
    "Enterprise Scale Addendum",
)
REQUIRED_COMMANDS = (
    "scripts/memory-recall.sh",
    "tools/sdd/gbrain_project.py search",
    "tools/sdd/capture_change_base.py --change",
    "openspec validate <change-id> --strict",
    "codegraph callers <symbol>",
    "codegraphcontext update . --quiet",
    "go run ./tools/logic-map",
    "make sdd-check",
    "make gate CHANGE=<change-id>",
    "tools/pm/generate_master_tracker.py --check",
)
REQUIRED_PATHS = (
    "tools/logic-map/extract_go_ast.go",
    "tools/logic-map/check_analysis.py",
    "tools/sdd-history/agent_save_hook.py",
    "tools/sdd-history/refresh_indexes.py",
    "tools/pm/generate_master_tracker.py",
    "scripts/memory_index.py",
    "tools/sdd/check_index_freshness.py",
    ".sdd/control-graph/typedb/schema.tql",
    ".mcp.json",
)


def block(text: str) -> str | None:
    if START not in text or END not in text:
        return None
    begin = text.index(START) + len(START)
    finish = text.index(END, begin)
    return text[begin:finish].strip("\n")


def mirror_text(source_text: str) -> str:
    shared = block(source_text)
    if shared is None:
        raise ValueError(f"{SOURCE}: shared markers missing")
    return (
        "# TossOS — Full SDD agent contract (Codex)\n\n"
        "이 파일은 `.claude/CLAUDE.md`의 공유 블록 mirror다.\n\n"
        f"{START}\n{shared}\n{END}\n"
    )


def check(root: Path = ROOT) -> list[str]:
    errors: list[str] = []
    source_path = root / SOURCE
    mirror_path = root / MIRROR
    if not source_path.exists() or not mirror_path.exists():
        return [f"missing agent config: {SOURCE if not source_path.exists() else MIRROR}"]
    source_text = source_path.read_text(encoding="utf-8")
    mirror_value = mirror_path.read_text(encoding="utf-8")
    source_block = block(source_text)
    mirror_block = block(mirror_value)
    if source_block is None or mirror_block is None:
        errors.append("shared SDD markers missing")
        return errors
    if source_block != mirror_block:
        errors.append("Claude/Codex shared SDD block drift; run --generate")
    for value in REQUIRED_SECTIONS:
        if value not in source_block:
            errors.append(f"required section missing: {value}")
    for value in REQUIRED_COMMANDS:
        if value not in source_block:
            errors.append(f"required command missing: {value}")
    for value in REQUIRED_PATHS:
        if not (root / value).exists():
            errors.append(f"referenced path missing: {value}")
    root_claude = (root / "CLAUDE.md").read_text(encoding="utf-8")
    root_agents = (root / "AGENTS.md").read_text(encoding="utf-8")
    if ".claude/CLAUDE.md" not in root_claude:
        errors.append("root CLAUDE.md does not route to Full SDD source")
    if ".claude/CLAUDE.md" not in root_agents:
        errors.append("root AGENTS.md does not route to Full SDD source")
    try:
        settings = json.loads(
            (root / ".claude" / "settings.json").read_text(encoding="utf-8")
        )
        allowed = settings.get("permissions", {}).get("allow", [])
    except (OSError, ValueError, TypeError):
        errors.append(".claude/settings.json is missing or invalid")
        allowed = []
    forbidden = {
        "Bash(rtk *)",
        "Bash(tossctl *)",
        "Bash(openspec *)",
        "Bash(codegraph *)",
        "Bash(codegraphcontext *)",
        "Bash(python3 tools/sdd/gbrain_project.py *)",
        "mcp__codegraph__*",
        "mcp__gbrain__*",
    }
    for permission in allowed:
        if permission in forbidden:
            errors.append(f"broad agent auto-approval is forbidden: {permission}")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", default=str(ROOT))
    parser.add_argument("--generate", action="store_true")
    args = parser.parse_args()
    root = Path(args.root)
    if args.generate:
        source = (root / SOURCE).read_text(encoding="utf-8")
        path = root / MIRROR
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(mirror_text(source), encoding="utf-8")
        print(f"[agent-config] generated {MIRROR}")
    errors = check(root)
    if errors:
        print("[agent-config] invalid:")
        for error in errors:
            print(f"  - {error}")
        return 1
    print("[agent-config] Claude/Codex/root Full SDD contracts are synchronized")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
