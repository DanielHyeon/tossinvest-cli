#!/usr/bin/env python3
"""`READ_CAP` 이 거절할 **정상 입력**의 열거표를 다시 찍는다 (a122 task 7.5.2.4).

`check_analysis.py:READ_CAP` 옆의 주석이 이 숫자들을 인용한다. 7.5.2.3 은 그 숫자를 세션에서
한 번 재고 하네스를 안 남겼고, 그 사이 한 문장에 **모집단 둘**이 섞였다(활성 번들의 최대를
전수 0/12,193 과 나란히 적었다). 숫자는 저장소와 함께 움직이므로 세는 자리를 남긴다.

    python3 7524_census.py
"""
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
BUNDLE = ("ast.json", "function-logic-map.md", "branch-test-map.md", "risk-pattern-report.md")
CAP = 16 << 20


def sizes(paths) -> list[tuple[int, Path]]:
    return sorted((path.stat().st_size, path) for path in paths if path.is_file())


def main() -> int:
    bundle = [path for path in (ROOT / "openspec" / "changes").rglob("*") if path.name in BUNDLE]
    every = sizes(bundle)
    active = sizes(path for path in bundle if "archive" not in path.parts)
    go = sizes(path for path in ROOT.rglob("*.go") if ".git" not in path.parts)
    print(f"번들 파일 전수        {len(every):6d} · 최대 {every[-1][0]:,} B  {every[-1][1].relative_to(ROOT)}")
    print(f"아카이브 아닌 것만    {len(active):6d} · 최대 {active[-1][0]:,} B  {active[-1][1].relative_to(ROOT)}")
    print(f"`*.go` 전수           {len(go):6d} · 최대 {go[-1][0]:,} B  {go[-1][1].relative_to(ROOT)}")
    over = [path for size, path in every if size > CAP]
    print(f"{CAP:,} B 를 넘는 번들 파일: **{len(over)}** / {len(every)}")
    # 상한이 거절할 정상 입력이 하나라도 있으면 그것은 열거표가 아니라 사고다.
    return 1 if over else 0


if __name__ == "__main__":
    raise SystemExit(main())
