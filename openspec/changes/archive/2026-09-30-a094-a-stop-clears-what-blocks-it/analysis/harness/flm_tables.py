#!/usr/bin/env python3
"""a094 FLM/BTM 표 생성기 — 번들의 ast.json 과 소스 원문, 커버리지 프로파일에서 분기 표를 **측정으로** 만든다.

사람이 채우는 것은 BTM 의 Test 열(시험 이름)과 FLM 의 산문뿐이다. 조건 원문 · 줄 · 진입 실측은 이 스크립트가 쓴다
(손으로 옮긴 분기 주장은 선택적이다 — .claude/CLAUDE.md 「단계 건너뛰기 금지」).

사용:
  python3 flm_tables.py <bundle-dir> <coverprofile> [tests.json]
    tests.json = {"B1": ["TestX", ...], ...} (없으면 Test 열이 비어 있어 check_analysis 가 거절한다)
출력: 표준 출력으로 두 표(Branches · BTM 행)를 낸다.
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]


def coverage(profile: Path, source: str) -> dict[int, int]:
    """그 파일의 블록 시작 줄 → count 최댓값."""
    out: dict[int, int] = {}
    if not profile.exists():
        return out
    pat = re.compile(r"^(?:.*/)?" + re.escape(source) + r":(\d+)\.\d+,\d+\.\d+ \d+ (\d+)$")
    for line in profile.read_text().splitlines():
        m = pat.match(line)
        if m:
            start, count = int(m.group(1)), int(m.group(2))
            out[start] = max(out.get(start, 0), count)
    return out


def main() -> int:
    bundle = Path(sys.argv[1])
    profile = Path(sys.argv[2])
    tests = json.loads(Path(sys.argv[3]).read_text()) if len(sys.argv) > 3 else {}
    ast = json.loads((bundle / "ast.json").read_text())
    source = ast["file"]
    lines = (ROOT / source).read_text().splitlines()
    cov = coverage(profile, source)
    branches = ast.get("branches") or []
    print("| Branch | 종류 | 조건 (원문) | 진입 실측 |")
    print("|---|---|---|---|")
    rows = []
    for b in branches:
        n = b["at"]["line"]
        text = lines[n - 1].strip().replace("|", "\\|")
        entered = "예" if cov.get(n, 0) > 0 else ("아니오" if n in cov else "—")
        print(f"| {b['id']} | {b['kind']} | `:{n}` `{text}` | {entered} |")
        rows.append((b["id"], n, text, entered))
    print()
    print("| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |")
    print("|---|---|---|---|---|---|")
    if not branches:
        cited = " · ".join(f"`{t}`" for t in tests.get("B1", []))
        print(f"| B1 | 분기 없음 — 유일한 경로 | — | {cited} | n/a | yes |")
    for bid, n, text, entered in rows:
        cited = " · ".join(f"`{t}`" for t in tests.get(bid, []))
        print(f"| {bid} | `:{n}` `{text}` | {entered} | {cited} | {tests.get(bid + ':red', 'n/a')} | yes |")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
