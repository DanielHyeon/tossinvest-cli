#!/usr/bin/env python3
"""a201 2.4 — 편집 전/뒤 ast.json 의 분기 목록을 정렬해 분기 번호 대응을 출력함(읽기 전용, 파일을 쓰지 않음).

분기 번호는 위치다 — 함수 안에 분기가 끼면 그 뒤 모든 번호가 다른 분기를 가리킴. 손으로 옮기지 않고, 각 분기의
(종류, 그 줄의 소스 글자)를 difflib 로 정렬해 대응을 셈.

사용: python3 renumber.py <옛 ast.json> <옛 소스 rev> <새 ast.json>
  옛 소스는 `git show <rev>:<ast 의 file>` 로, 새 소스는 작업 트리에서 읽음.
"""
import difflib
import json
import pathlib
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parents[5]


def lines_at(rev: str | None, path: str) -> list[str]:
    if rev is None:
        return (ROOT / path).read_text(encoding="utf-8").splitlines()
    out = subprocess.run(["git", "-C", str(ROOT), "show", f"{rev}:{path}"], check=True,
                         capture_output=True, text=True).stdout
    return out.splitlines()


def keyed(ast: dict, source: list[str]) -> list[tuple[str, str]]:
    return [(b["kind"], source[b["at"]["line"] - 1].strip()) for b in ast.get("branches") or []]


def main() -> None:
    old_ast = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
    rev = sys.argv[2]
    new_ast = json.loads(pathlib.Path(sys.argv[3]).read_text(encoding="utf-8"))
    old_b, new_b = old_ast.get("branches") or [], new_ast.get("branches") or []
    old_k = keyed(old_ast, lines_at(rev, old_ast["file"]))
    new_k = keyed(new_ast, lines_at(None, new_ast["file"]))
    matcher = difflib.SequenceMatcher(a=old_k, b=new_k, autojunk=False)
    for tag, i1, i2, j1, j2 in matcher.get_opcodes():
        if tag == "equal":
            for o, n in zip(range(i1, i2), range(j1, j2)):
                print(f"{old_b[o]['id']} {old_b[o]['kind']}@{old_b[o]['at']['line']} -> "
                      f"{new_b[n]['id']} {new_b[n]['kind']}@{new_b[n]['at']['line']}  {new_k[n][1][:90]}")
        else:
            for o in range(i1, i2):
                print(f"{old_b[o]['id']} {old_b[o]['kind']}@{old_b[o]['at']['line']} -> ({tag}: 사라짐)  {old_k[o][1][:90]}")
            for n in range(j1, j2):
                print(f"(새 분기 — {tag}) -> {new_b[n]['id']} {new_b[n]['kind']}@{new_b[n]['at']['line']}  {new_k[n][1][:90]}")


if __name__ == "__main__":
    main()
