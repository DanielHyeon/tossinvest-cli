#!/usr/bin/env python3
"""a126 FLM 분기 표 생성기 — ast.json · 소스 · 커버리지 프로파일에서 분기 행을 기계로 만든다.

사용(저장소 루트): python3 <this> <bundle dir> <coverprofile>
출력: 분기 표 Markdown(표준 출력). 조건은 소스 그 줄 원문, 창의 return 은 [분기 줄, 다음 분기 줄) 안의 AST return 좌표,
진입 실측은 그 줄로 시작하는 커버리지 블록 count(없으면 `—`). 분기의 **의미**는 산문이 적는다 — 이 표는 위치다.
"""
import json
import sys
from pathlib import Path


def blocks(profile: str, relative: str) -> dict[int, int]:
    out: dict[int, int] = {}
    for line in Path(profile).read_text().splitlines()[1:]:
        loc, _, count = line.rpartition(" ")
        path, _, span = loc.rpartition(":")
        if not path.endswith(relative):
            continue
        start = int(span.split(".")[0])
        out[start] = max(out.get(start, 0), int(count))
    return out


def main() -> int:
    bundle, profile = Path(sys.argv[1]), sys.argv[2]
    ast = json.loads((bundle / "ast.json").read_text())
    relative = ast["file"]
    lines = Path(relative).read_text().splitlines()
    cov = blocks(profile, relative)
    branches = ast.get("branches") or []
    returns = ast.get("returns") or []
    end = ast["end"]["line"] + 1
    print("| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |")
    print("|---|---|---|---|---|")
    for i, b in enumerate(branches):
        line = b["at"]["line"]
        upper = branches[i + 1]["at"]["line"] if i + 1 < len(branches) else end
        rets = [f":{r['at']['line']}" for r in returns if line <= r["at"]["line"] < max(upper, line + 1)]
        entered = "—" if line not in cov else ("예" if cov[line] > 0 else "아니오")
        text = lines[line - 1].strip().replace("|", "\\|")
        if len(text) > 140:
            text = text[:137] + "…"
        print(f"| {b['id']} | {b['kind']} | `:{line}` `{text}` | {', '.join(rets) or '—'} | {entered} |")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
