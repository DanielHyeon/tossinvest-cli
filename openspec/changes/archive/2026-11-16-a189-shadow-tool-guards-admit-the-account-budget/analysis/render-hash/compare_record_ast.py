#!/usr/bin/env python3
"""a189 #2 — ExitObserver.record 의 편집 전·후 tools/logic-map AST 대조.

입력: 같은 디렉터리의 record-ast-9c4bdf4de-parent.json(편집 전), record-ast-9c4bdf4de.json(편집 후),
record-ast-base.json(base 7d1e3163a 워킹트리 = 현재).
각 항목(분기·return·호출·대입)을 「함수 시작 줄 기준 상대 줄 + 종류 + 호출 텍스트」로 정규화하고 difflib 으로 정렬해
더해진 것·빠진 것만 찍음. 줄 번호가 밀리는 것은 상대 줄 + 정렬로 흡수함.
"""
import difflib
import json
import pathlib

HERE = pathlib.Path(__file__).resolve().parent


def items(path):
    d = json.loads((HERE / path).read_text())
    start = d["start"]["line"]
    out = []
    for group in ("branches", "returns", "calls", "assignments"):
        for e in d.get(group) or []:
            out.append((e["at"]["line"] - start, e["at"]["column"], group, e.get("kind", ""), e.get("text", "")))
    out.sort()
    # 정렬 키에서 상대 줄은 빼고 순서만 남김 — 편집으로 줄이 밀려도 같은 항목이 맞물리게 함
    return d, [f"{g}:{k}:{t}" for (_, _, g, k, t) in out], out


def report(a, b):
    da, la, ra = items(a)
    db, lb, rb = items(b)
    print(f"== {a} -> {b}")
    for group in ("branches", "returns", "calls", "assignments"):
        print(f"  {group}: {len(da.get(group) or [])} -> {len(db.get(group) or [])}")
    sm = difflib.SequenceMatcher(a=la, b=lb, autojunk=False)
    for op, i1, i2, j1, j2 in sm.get_opcodes():
        if op == "equal":
            continue
        for i in range(i1, i2):
            print(f"  - (rel {ra[i][0]}) {la[i]}")
        for j in range(j1, j2):
            print(f"  + (rel {rb[j][0]}) {lb[j]}")
    print(f"  equal items: {sum(i2 - i1 for op, i1, i2, _, _ in sm.get_opcodes() if op == 'equal')}")


report("record-ast-9c4bdf4de-parent.json", "record-ast-9c4bdf4de.json")
report("record-ast-9c4bdf4de.json", "record-ast-base.json")
