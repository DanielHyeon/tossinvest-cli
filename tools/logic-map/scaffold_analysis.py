#!/usr/bin/env python3
"""Create the required Function Logic Map artifact skeleton."""

from __future__ import annotations

import argparse
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def slug(value: str) -> str:
    return re.sub(r"[^a-z0-9._-]+", "-", value.lower()).strip("-")


def artifact_dir(change: str, source: str, function: str, root: Path = ROOT) -> Path:
    if not re.fullmatch(r"[a-z0-9][a-z0-9-]*", change):
        raise ValueError(f"invalid OpenSpec change id: {change!r}")
    changes_root = (root / "openspec" / "changes").resolve()
    package = Path(source).parent.as_posix().replace("/", "-")
    output = changes_root / change / "analysis" / "function-logic" / (
        f"{slug(package)}--{slug(function)}"
    )
    resolved = output.resolve()
    if not resolved.is_relative_to(changes_root):
        raise ValueError("analysis output escapes openspec/changes")
    return resolved


def render(source: str, function: str) -> dict[str, str]:
    return {
        "ast.json": "{}\n",
        "function-logic-map.md": f"""# Function Logic Map: `{function}`

- Source: `{source}`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| TODO | TODO | TODO | TODO |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | TODO | TODO | TODO | TODO |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| TODO | TODO | TODO | CodeGraph + AST |

## State mutations and fallbacks

- TODO

## Safety conclusion

- Safe edit boundary: TODO
- High-risk impact: yes/no
""",
        "branch-test-map.md": f"""# Branch Test Map: `{function}`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | TODO | TODO | no | no |
""",
        "risk-pattern-report.md": f"""# Risk Pattern Report: `{function}`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml {source}
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| TODO | TODO | defect/reviewed-safe/not-applicable | TODO |
""",
    }


def scaffold(change: str, source: str, function: str, root: Path = ROOT) -> Path:
    out = artifact_dir(change, source, function, root)
    out.mkdir(parents=True, exist_ok=True)
    for name, content in render(source, function).items():
        path = out / name
        if not path.exists():
            path.write_text(content, encoding="utf-8")
    return out


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--change", required=True)
    parser.add_argument("--file", required=True)
    parser.add_argument("--func", required=True)
    parser.add_argument("--root", default=str(ROOT))
    args = parser.parse_args()
    out = scaffold(args.change, args.file, args.func, Path(args.root))
    print(out)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
