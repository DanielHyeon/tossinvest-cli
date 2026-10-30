"""a087 0.10 증거 — evidence/*.ast.json 의 분기·반환 좌표를 원문 줄과 함께 전부 열거함.

실행: 저장소 루트에서 `python3 -I openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/evidence/enumerate.py`
- 각 ast.json 의 source_sha256 을 현재 파일 바이트와 대조하고 다르면 그 파일을 STALE 로 표시함(낡은 AST 로 주장하지 않기 위함).
- 손으로 고른 줄이 아니라 AST 가 낸 좌표 전부를 찍음.
"""
import hashlib
import json
import pathlib
import sys

HERE = pathlib.Path(__file__).resolve().parent
ROOT = HERE.parents[4]


def main() -> int:
    stale = 0
    for path in sorted(HERE.glob("*.ast.json")):
        ast = json.loads(path.read_text())
        src = ROOT / ast["file"]
        raw = src.read_bytes()
        fresh = hashlib.sha256(raw).hexdigest() == ast["source_sha256"]
        stale += 0 if fresh else 1
        lines = raw.decode().splitlines()
        name = ast["function"] if not ast.get("receiver") else f'{ast["receiver"]}.{ast["function"]}'
        print(f'## {name} — {ast["file"]}:{ast["start"]["line"]}-{ast["end"]["line"]} '
              f'({"fresh" if fresh else "STALE"}, sha {ast["source_sha256"][:12]})')
        events = [(b["at"]["line"], b["id"], b["kind"]) for b in ast.get("branches") or []]
        events += [(r["at"]["line"], "R", r["kind"]) for r in ast.get("returns") or []]
        for line, ident, kind in sorted(events):
            print(f'  {ident:>4} {kind:<7} :{line:<5} {lines[line - 1].strip()}')
        print(f'  branches={len(ast.get("branches") or [])} returns={len(ast.get("returns") or [])}')
        print()
    print(f"stale={stale}")
    return 1 if stale else 0


if __name__ == "__main__":
    sys.exit(main())
