#!/usr/bin/env python3
"""a112 — 편집 전 FLM 번들(ast.json + function-logic-map.md)을 AST 에서만 렌더한다.

좌표 · 분기 · return · 호출은 `go run ./tools/logic-map` 산출물에서만 읽고, 저자는 편집 계획과 안전 결론 한 줄만 쓴다.
분기 조건 문자열은 AST 가 가리킨 줄의 원문(한 줄)을 그대로 옮긴다 — 손으로 요약하지 않는다.

사용(저장소 루트에서):
  render_pre_edit.py <out-root> <lot-tag> <file> <func> <plan> <safety>
산출: <out-root>/<bundle>/ast.json · function-logic-map.md (bundle = 경로-슬러그--함수 소문자, function-logic/ 관례)
"""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path


def bundle_name(file: str, func: str) -> str:
    pkg = file.rsplit("/", 1)[0].replace("/", "-")
    return f"{pkg}--{func.lower()}"


def main() -> int:
    out_root, tag, file, func, plan, safety = sys.argv[1:7]
    raw = subprocess.run(["go", "run", "./tools/logic-map", "-file", file, "-func", func],
                         capture_output=True, text=True, check=True).stdout
    ast = json.loads(raw)
    lines = Path(file).read_text().split("\n")
    out = Path(out_root) / bundle_name(file, func)
    out.mkdir(parents=True, exist_ok=True)
    (out / "ast.json").write_text(raw)
    pos = lambda n: f"{n['at']['line']}:{n['at']['column']}"
    title = func.rsplit(".", 1)[-1]
    md = [f"# Function Logic Map (편집 전): `{title}`", "",
          f"- Source: `{file}`",
          f"- Source SHA-256: `{ast['source_sha256']}`",
          f"- Signature: `{ast['signature']}`",
          f"- Source range: `{ast['start']['line']}:{ast['start']['column']}`–`{ast['end']['line']}:{ast['end']['column']}`",
          f"- AST evidence: `ast.json` — 편집 **전**({tag}).", "",
          "## Inputs and invariants", "", f"- 편집 계획: {plan}", "",
          "## Branches and early returns", "",
          "- Exact AST return nodes: " + (", ".join(f"`{pos(r)}`" for r in ast.get("returns") or []) or "없음") + ".", ""]
    branches = ast.get("branches") or []
    if branches:
        md += ["| Branch | AST kind | Source location | Condition |", "|---|---|---|---|"]
        for b in branches:
            text = lines[b["at"]["line"] - 1].strip().replace("|", "/")
            md.append(f"| {b['id']} | {b['kind']} | {pos(b)} | `{text}` |")
    else:
        md.append("- 분기 없음(AST 열거 0).")
    md += ["", "## Calls and live bindings", "", "| Callee expression | Position |", "|---|---|"]
    md += [f"| `{c.get('text', '(unnamed)')}` | {pos(c)} |" for c in ast.get("calls") or []]
    md += ["", "## Safety conclusion", "", f"- {safety}", ""]
    (out / "function-logic-map.md").write_text("\n".join(md))
    print(f"{out.name}: branches {len(branches)}, returns {len(ast.get('returns') or [])}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
