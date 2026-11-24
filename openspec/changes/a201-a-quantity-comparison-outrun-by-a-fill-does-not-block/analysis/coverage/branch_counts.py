#!/usr/bin/env python3
"""a201 2.4 — 커버 프로파일(mode set)에서 함수 분기마다 몸이 실행됐는지 셈(읽기 전용, 표준 출력만).

분기 몸 = 프로파일에서 분기 위치(줄, 열) **바로 뒤에 시작하는** 첫 블록(if 몸 · else 몸 · range 몸 · case 절).
여러 프로파일은 합집합(어느 하나라도 1 이면 실행됨). 함수 범위 안의 블록 줄도 그대로 출력해 판정을 다시 볼 수 있게 함.

사용: python3 branch_counts.py <ast.json> <profile> [<profile> ...]
"""
import json
import pathlib
import re
import sys

LINE = re.compile(r"^(?P<file>.+):(?P<sl>\d+)\.(?P<sc>\d+),(?P<el>\d+)\.(?P<ec>\d+) (?P<n>\d+) (?P<c>\d+)$")


def main() -> None:
    ast = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding="utf-8"))
    suffix = "/" + ast["file"]
    start, end = ast["start"]["line"], ast["end"]["line"]
    blocks: dict[tuple[int, int, int, int], int] = {}
    for profile in sys.argv[2:]:
        for raw in pathlib.Path(profile).read_text(encoding="utf-8").splitlines():
            m = LINE.match(raw)
            if not m or not m["file"].endswith(suffix):
                continue
            key = (int(m["sl"]), int(m["sc"]), int(m["el"]), int(m["ec"]))
            if not (start <= key[0] <= end):
                continue
            blocks[key] = max(blocks.get(key, 0), int(m["c"]))
    ordered = sorted(blocks)
    print(f"# {ast['function']} {ast['file']} ({start}-{end}) sha {ast['source_sha256'][:12]} — profiles: {len(sys.argv) - 2}")
    for b in ast.get("branches") or []:
        at = (b["at"]["line"], b["at"]["column"])
        body = next((k for k in ordered if (k[0], k[1]) > at), None)
        if body is None:
            print(f"{b['id']} {b['kind']}@{at[0]} body=? (블록 없음)")
            continue
        print(f"{b['id']} {b['kind']}@{at[0]} body={blocks[body]} ({body[0]}.{body[1]},{body[2]}.{body[3]})")
    print("# blocks")
    for k in ordered:
        print(f"{k[0]}.{k[1]},{k[2]}.{k[3]} {blocks[k]}")


if __name__ == "__main__":
    main()
