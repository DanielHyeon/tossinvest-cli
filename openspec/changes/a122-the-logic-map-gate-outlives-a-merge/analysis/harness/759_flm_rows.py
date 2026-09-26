#!/usr/bin/env python3
"""task 7.5.9 · 7.5.10 · 7.5.11 · 7.5.15 · 7.5.16 — FLM 표의 행을 **열거 산출물에서** 찍는다(손으로 옮기지 않는다).

편집 전/후 `ast.*.json`(`python-function-logic/enumerate.py` 가 뽑은 것)을 `difflib.SequenceMatcher` 로 (종류, 소스)
열에서 정렬해 옛 id → 새 id 표를 찍는다 — 분기 번호는 위치라서 손 재번호는 편집 지점 뒤를 전부 어긋나게 한다
([[positional-branch-ids-break-hand-renumbering]]). 편집 전 판이 없으면(새 함수) 편집 후 표만 찍는다.

    python3 759_flm_rows.py <함수> <편집 전 꼬리표|-> <편집 후 꼬리표>
    python3 759_flm_rows.py _listing_outcome 7510 7510
"""
import difflib
import json
import sys
from pathlib import Path

BUNDLES = Path(__file__).resolve().parents[1] / "python-function-logic"


def load(function: str, stem: str) -> dict:
    return json.loads((BUNDLES / f"tools-logic-map--{function}" / f"{stem}.json").read_text(encoding="utf-8"))


def cell(text: str) -> str:
    return text.strip().replace("|", "\\|")


def main() -> None:
    function, before_tag, after_tag = sys.argv[1:4]
    after = load(function, f"ast.after-{after_tag}")
    counts = lambda d: f"분기 {len(d['branches'])} · 반환 {len(d['returns'])} · raise {len(d['raises'])} · 호출 {len(d['calls'])}"
    print(f"편집 후 `{after['file']}:{after['start']['line']}-{after['end']['line']}` · {counts(after)} "
          f"(`ast.after-{after_tag}.json`, source sha `{after['source_sha256'][:12]}`)")
    if before_tag == "-":
        print("\n| id | 줄 | 종류 | 소스 |\n|---|---|---|---|")
        for branch in after["branches"]:
            print(f"| {branch['id']} | {branch['line']} | {branch['kind']} | `{cell(branch['source'])}` |")
        return
    before = load(function, f"ast.before-{before_tag}")
    print(f"편집 전 `{before['file']}:{before['start']['line']}-{before['end']['line']}` · {counts(before)} "
          f"(`ast.before-{before_tag}.json`, revision `{before['revision'][:8]}`, source sha `{before['source_sha256'][:12]}`)")
    old = [(b["kind"], b["source"]) for b in before["branches"]]
    new = [(b["kind"], b["source"]) for b in after["branches"]]
    print("\n| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |\n|---|---|---|---|---|---|")
    for op, i1, i2, j1, j2 in difflib.SequenceMatcher(a=old, b=new, autojunk=False).get_opcodes():
        if op == "equal":
            for i, j in zip(range(i1, i2), range(j1, j2)):
                b = after["branches"][j]
                same = "같음" if before["branches"][i]["id"] == b["id"] else "번호만"
                print(f"| {before['branches'][i]['id']} | {b['id']} | {b['line']} | {b['kind']} | `{cell(b['source'])}` | {same} |")
            continue
        for i in range(i1, i2):
            a = before["branches"][i]
            print(f"| {a['id']} | — | — | {a['kind']} | `{cell(a['source'])}` (옛) | **빠짐** |")
        for j in range(j1, j2):
            b = after["branches"][j]
            print(f"| — | {b['id']} | {b['line']} | {b['kind']} | `{cell(b['source'])}` | **새** |")


if __name__ == "__main__":
    main()
