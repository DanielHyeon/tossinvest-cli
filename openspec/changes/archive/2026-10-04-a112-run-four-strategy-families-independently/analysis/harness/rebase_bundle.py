#!/usr/bin/env python3
"""a112 게이트 위생 — 분기 구조(id · 종류 · return 수)는 같은데 호출이 늘거나 본문 줄이 고르지 않게 밀린 번들을 재기준화한다.

shift_same_file_bundles.py 는 「순수 줄 이동」만 받는다(호출 텍스트 순서 · 균일한 줄 차이). 이 스크립트는 그보다 한 걸음 넓다:
- 판정: 옛 · 새 ast 의 분기 (id, 종류) 목록과 return 수가 같아야 한다 — 다르면 멈춘다(분기가 바뀐 번들은 손으로 다시 쓴다).
- 동작: 시작 · 끝 · 분기 · return 좌표를 **같은 id/순번끼리** 옛 → 새로 사상해 md 두 장의 `줄:열` 토큰과 `(start-end)` 범위 · 파일 SHA-256 을 바꾸고,
  FLM 의 「Calls and live bindings」 표를 새 AST 호출 목록으로 다시 쓴다(호출 표는 게이트가 AST 와 한 줄씩 대조한다). 좌표를 손으로 적지 않는다.

사용: rebase_bundle.py <bundle-dir> <new-ast.json> --note "…"
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path


def anchors(ast: dict) -> list[tuple[int, int]]:
    out = [(ast["start"]["line"], ast["start"]["column"]), (ast["end"]["line"], ast["end"]["column"])]
    for key in ("branches", "returns"):
        out += [(item["at"]["line"], item["at"]["column"]) for item in ast.get(key) or []]
    return out


def main() -> int:
    bundle, new_path = Path(sys.argv[1]), Path(sys.argv[2])
    note = sys.argv[4] if len(sys.argv) > 4 and sys.argv[3] == "--note" else ""
    old, new = json.loads((bundle / "ast.json").read_text()), json.loads(new_path.read_text())
    shape = lambda a: ([(b["id"], b["kind"]) for b in a.get("branches") or []], len(a.get("returns") or []), a["function"], a.get("receiver"))
    if shape(old) != shape(new):
        print(f"STOP {bundle.name}: 분기 구조가 다르다 — 손으로 다시 쓸 것", file=sys.stderr)
        return 2
    mapping = {f"{a}:{b}": f"{c}:{d}" for (a, b), (c, d) in zip(anchors(old), anchors(new))}
    pos = lambda n: f"{n['at']['line']}:{n['at']['column']}"
    calls = [f"| `{c.get('text', '(unnamed)')}` | {pos(c)} |" for c in new.get("calls") or []]
    token = re.compile(r"(?<![\d:])(\d+):(\d+)(?![\d:])")
    for name in ("function-logic-map.md", "branch-test-map.md"):
        path = bundle / name
        text = path.read_text()
        text = text.replace(old["source_sha256"], new["source_sha256"])
        text = text.replace(f"({old['start']['line']}-{old['end']['line']})", f"({new['start']['line']}-{new['end']['line']})")
        if name == "function-logic-map.md":
            head, sep, rest = text.partition("## Calls and live bindings")
            if sep:
                table_start = rest.index("|---|---|") + len("|---|---|")
                after = rest[table_start:]
                end = re.search(r"\n(?!\|)", after[1:])
                tail = after[1 + end.start():] if end else ""
                rest = rest[:table_start] + "\n" + "\n".join(calls) + tail
                text = head + sep + rest
        lines = text.split("\n")
        out = []
        for line in lines:
            if line.startswith("| `") and name == "function-logic-map.md" and line in calls:
                out.append(line)  # 새 호출 표 행은 이미 새 좌표
                continue
            out.append(token.sub(lambda m: mapping.get(m.group(0), m.group(0)), line))
        text = "\n".join(out)
        if note and name == "branch-test-map.md":
            text = text.rstrip("\n") + f"\n\n> 재기준화: {note} — 분기 (id · 종류) · return 수 동일, 좌표는 id 끼리 사상, 호출 표는 새 AST.\n"
        path.write_text(text)
    (bundle / "ast.json").write_text(new_path.read_text())
    print(f"rebased {bundle.name}: anchors {len(mapping)}, calls {len(calls)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
