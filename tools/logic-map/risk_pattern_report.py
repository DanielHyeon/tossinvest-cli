#!/usr/bin/env python3
"""Run ast-grep and render a compact markdown evidence report."""

from __future__ import annotations

import argparse
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def scan(path: str, root: Path = ROOT) -> list[dict]:
    proc = subprocess.run(
        [
            "ast-grep",
            "scan",
            "--json=compact",
            "-c",
            str(root / "tools" / "logic-map" / "sgconfig.yml"),
            path,
        ],
        cwd=root,
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode not in (0, 1):
        raise RuntimeError(proc.stderr.strip() or f"ast-grep exit {proc.returncode}")
    if not proc.stdout.strip():
        return []
    value = json.loads(proc.stdout)
    return value if isinstance(value, list) else [value]


def markdown(path: str, findings: list[dict]) -> str:
    lines = [f"# Risk Pattern Report: `{path}`", "", "| Rule | Location | Message |", "|---|---|---|"]
    for item in findings:
        start = item.get("range", {}).get("start", {})
        line = int(start.get("line", 0)) + 1
        lines.append(
            f"| {item.get('ruleId', 'unknown')} | `{path}:{line}` | "
            f"{str(item.get('message', '')).replace('|', '\\|')} |"
        )
    if not findings:
        lines.append("| — | — | No configured risk pattern matched |")
    lines.extend(["", "> Findings are review candidates, not automatic defect verdicts.", ""])
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("path")
    parser.add_argument("--output")
    args = parser.parse_args()
    rendered = markdown(args.path, scan(args.path))
    if args.output:
        Path(args.output).write_text(rendered, encoding="utf-8")
    else:
        print(rendered)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
