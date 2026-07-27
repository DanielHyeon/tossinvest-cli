#!/usr/bin/env python3
"""Verify the minimal Claude/Codex safety bootstrap and workflow routing."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ".claude/CLAUDE.md"
MIRROR = ".codex/agents.md"
WORKFLOW = "docs/WORKFLOW.md"
START = "<!-- SDD_SHARED_START -->"
END = "<!-- SDD_SHARED_END -->"
REQUIRED_BOOTSTRAP = (
    "사람 승인 없는 LIVE 주문",
    "mutating: true",
    "토글 OFF",
    "손절·비상 청산",
    "High-risk",
    "운영 토글 flip",
    "docs/WORKFLOW.md",
    "OpenSpec",
    "CodeGraph hard evidence",
    "Function Logic Map",
    "make sdd-sync",
    "make sdd-check",
    "make gate CHANGE=<change-id>",
)
REQUIRED_PATHS = (
    "docs/WORKFLOW.md",
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
        "# TossOS — agent safety bootstrap (Codex)\n\n"
        "이 파일은 `.claude/CLAUDE.md`의 최소 안전 부트스트랩 mirror다.\n"
        "상세 개발 절차의 단일 정본은 `docs/WORKFLOW.md`이며 개발 작업 전에 반드시 읽는다.\n\n"
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
        errors.append("minimal safety bootstrap markers missing")
        return errors
    if source_block != mirror_block:
        errors.append("Claude/Codex minimal safety bootstrap drift; run --generate")
    for value in REQUIRED_BOOTSTRAP:
        if value not in source_block:
            errors.append(f"required bootstrap contract missing: {value}")
    for path, text in ((SOURCE, source_text), (MIRROR, mirror_value)):
        if WORKFLOW not in text:
            errors.append(f"{path} does not route to {WORKFLOW}")
    for value in REQUIRED_PATHS:
        if not (root / value).exists():
            errors.append(f"referenced path missing: {value}")

    root_claude = (root / "CLAUDE.md").read_text(encoding="utf-8")
    root_agents = (root / "AGENTS.md").read_text(encoding="utf-8")
    if SOURCE not in root_claude or WORKFLOW not in root_claude:
        errors.append("root CLAUDE.md does not route to bootstrap and workflow")
    if SOURCE not in root_agents or WORKFLOW not in root_agents:
        errors.append("root AGENTS.md does not route to bootstrap and workflow")

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
    print("[agent-config] Claude/Codex safety bootstrap and workflow routing are synchronized")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
