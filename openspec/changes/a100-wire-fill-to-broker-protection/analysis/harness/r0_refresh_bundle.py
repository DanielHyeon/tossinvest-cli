#!/usr/bin/env python3
"""a100 R0 재동결(2026-10-04) — 소스가 바뀐 편집 전 번들을 현재 소스로 다시 맞춘다.

두 경우:
  shift  : 옛 · 새 AST 의 분기 (id, 종류) 목록이 같다 → ast.json 을 새 추출로 바꾸고 md 두 장의 `(start-end)` 범위만 옮긴다(산문 · 시험 인용 불변).
           옛 ast 는 returns/calls 를 싣지 않던 추출기 판본이라 그 둘은 비교하지 않는다(형식 차이).
  adopt  : 분기 구조가 바뀌었다 → 그 함수를 바꾼 change 의 아카이브 번들(분기 · 시험 인용이 그 편집 뒤 기준)을 통째로 가져오고,
           그 번들의 source 가 현재 파일과 다르면 shift 를 한 번 더 적용한다. 머리에 a100 채택 표지를 단다.
사용: r0_refresh_bundle.py shift <bundle-dir> <new-ast.json>
      r0_refresh_bundle.py adopt <bundle-dir> <archive-bundle-dir> <new-ast.json> <note>
"""
import json, shutil, sys
from pathlib import Path

def kinds(a):
    return [(b["id"], b["kind"]) for b in a.get("branches") or []]

def shift(bundle: Path, new_path: Path) -> None:
    old, new = json.loads((bundle / "ast.json").read_text()), json.loads(new_path.read_text())
    if kinds(old) != kinds(new) or old["function"] != new["function"]:
        sys.exit(f"STOP {bundle.name}: 분기 구조가 다르다")
    a = f"({old['start']['line']}-{old['end']['line']})"
    b = f"({new['start']['line']}-{new['end']['line']})"
    for name in ("function-logic-map.md", "branch-test-map.md"):
        p = bundle / name
        t = p.read_text().replace(a, b).replace(old["source_sha256"], new["source_sha256"])
        p.write_text(t)
    (bundle / "ast.json").write_text(new_path.read_text())
    print(f"shift {bundle.name}: {a} -> {b}")

def adopt(bundle: Path, archive: Path, new_path: Path, note: str) -> None:
    for name in ("ast.json", "function-logic-map.md", "branch-test-map.md", "risk-pattern-report.md"):
        if (archive / name).exists():
            shutil.copyfile(archive / name, bundle / name)
    arch = json.loads((bundle / "ast.json").read_text())
    new = json.loads(new_path.read_text())
    if arch["source_sha256"] != new["source_sha256"]:
        if kinds(arch) != kinds(new):
            sys.exit(f"STOP {bundle.name}: 아카이브 번들 뒤에도 분기 구조가 바뀜 — 손으로 다시 쓸 것")
    p = bundle / "function-logic-map.md"
    t = p.read_text()
    first, _, rest = t.partition("\n")
    p.write_text(first + "\n\n> **a100 R0 재동결(2026-10-04) — 채택 번들.** " + note + "\n" + rest)
    print(f"adopt {bundle.name} <- {archive}")

if __name__ == "__main__":
    mode = sys.argv[1]
    if mode == "shift":
        shift(Path(sys.argv[2]), Path(sys.argv[3]))
    else:
        adopt(Path(sys.argv[2]), Path(sys.argv[3]), Path(sys.argv[4]), sys.argv[5])
