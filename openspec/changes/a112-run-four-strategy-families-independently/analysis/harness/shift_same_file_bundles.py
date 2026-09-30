#!/usr/bin/env python3
"""a112 — 같은 파일의 다른 함수 편집으로 **본문은 그대로이고 줄만 밀린** 번들을 재기준화한다.

입력: 번들 디렉터리와 새로 뽑은 ast.json(`go run ./tools/logic-map --file … --func …`).
판정: 옛 ast 와 새 ast 의 구조(분기 id·종류, return 수, 호출 텍스트 순서)가 같아야 한다 — 다르면 멈춘다(본문이 바뀐
번들은 이 스크립트의 대상이 아니다, 손으로 다시 쓴다).
동작: 옛 좌표 → 새 좌표 사상을 AST 에서 만들고(분기 · return · 호출 · 대입 · 시작/끝), md 두 장의 `줄:열` 토큰 중 **옛 AST 에
있는 좌표만** 새 좌표로 바꾼다. 열 없는 `(:줄)` · `:줄–` 인용도 옛 AST 의 줄일 때만 옮긴다. `(start-end)` 범위와 파일 SHA-256 도 바꾼다. 손으로 좌표를 적지 않는다.

사용: shift_same_file_bundles.py <bundle-dir> <new-ast.json> [--note "…"]
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path


def positions(ast: dict) -> list[tuple[int, int]]:
    out = [(ast["start"]["line"], ast["start"]["column"]), (ast["end"]["line"], ast["end"]["column"])]
    for key in ("branches", "returns", "calls", "assignments", "defers", "go_statements"):
        out += [(item["at"]["line"], item["at"]["column"]) for item in ast.get(key) or []]
    return out


def shape(ast: dict) -> tuple:
    return ([(b["id"], b["kind"]) for b in ast.get("branches") or []], len(ast.get("returns") or []),
            [c.get("text") for c in ast.get("calls") or []], ast["function"], ast.get("receiver"))


def main() -> int:
    bundle, new_path = Path(sys.argv[1]), Path(sys.argv[2])
    note = sys.argv[4] if len(sys.argv) > 4 and sys.argv[3] == "--note" else ""
    old, new = json.loads((bundle / "ast.json").read_text()), json.loads(new_path.read_text())
    if shape(old) != shape(new):
        print(f"STOP {bundle.name}: 본문 구조가 다르다 — 이 스크립트의 대상이 아님", file=sys.stderr)
        return 2
    mapping = {f"{a}:{b}": f"{c}:{d}" for (a, b), (c, d) in zip(positions(old), positions(new))}
    # 새 좌표는 전부 옛 좌표 + 같은 줄 차이여야 한다(순수 이동). 아니면 이동이 아니다.
    deltas = {int(v.split(":")[0]) - int(k.split(":")[0]) for k, v in mapping.items()}
    if len(deltas) != 1:
        print(f"STOP {bundle.name}: 줄 차이가 하나가 아님 {sorted(deltas)}", file=sys.stderr)
        return 2
    delta = deltas.pop()
    token = re.compile(r"(?<![\d.])(\d+):(\d+)(?![\d])")
    line_ref = re.compile(r"(\(:|(?<![\w\d]):)(\d{2,5})(\)|–)")
    old_lines = {line for line, _ in positions(old)}
    rng = f"({old['start']['line']}-{old['end']['line']})"
    new_rng = f"({new['start']['line']}-{new['end']['line']})"
    for name in ("function-logic-map.md", "branch-test-map.md"):
        path = bundle / name
        if not path.exists():
            continue
        text = path.read_text()
        text = token.sub(lambda m: mapping.get(m.group(0), m.group(0)), text)
        # 열 없는 줄 범위(`:616–694`)는 양 끝이 옛 AST 의 시작 · 끝일 때 둘 다 옮긴다.
        text = text.replace(f":{old['start']['line']}–{old['end']['line']}", f":{new['start']['line']}–{new['end']['line']}")
        # 열 없는 줄 인용(`(:618)` · `:616–`)은 그 줄이 옛 AST 의 줄일 때만 옮긴다 — 다른 함수 · 옛 측정 블록 인용은 그대로.
        text = line_ref.sub(lambda m: m.group(1) + (str(int(m.group(2)) + delta) if int(m.group(2)) in old_lines else m.group(2))
                            + m.group(3), text)
        text = text.replace(rng, new_rng).replace(old["source_sha256"], new["source_sha256"])
        text = text.replace(f"`{old['source_sha256'][:12]}…`", f"`{new['source_sha256'][:12]}…`")  # 줄인 표기
        if name == "function-logic-map.md" and note and note not in text:
            text = text.rstrip("\n") + "\n\n" + note + "\n"
        path.write_text(text)
    (bundle / "ast.json").write_text(new_path.read_text())
    print(f"shifted {bundle.name}: Δ{delta:+d}, 좌표 {len(mapping)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
