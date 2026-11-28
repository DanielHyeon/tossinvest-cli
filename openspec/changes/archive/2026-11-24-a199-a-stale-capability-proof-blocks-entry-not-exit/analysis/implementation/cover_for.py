#!/usr/bin/env python3
"""coverprofile(set 모드)에서 한 함수 범위의 블록과 count 를 뽑음 — BTM 의 「GREEN observed」 칸 근거.

사용: python3 cover_for.py <coverprofile> <파일 접미사> <시작줄> <끝줄>
"""

from __future__ import annotations

import sys


def main() -> None:
    profile, suffix, lo, hi = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4])
    for line in open(profile, encoding="utf-8"):
        if line.startswith("mode:"):
            continue
        loc, _stmts, count = line.rsplit(" ", 2)
        path, span = loc.rsplit(":", 1)
        if not path.endswith(suffix):
            continue
        start = int(span.split(",")[0].split(".")[0])
        if lo <= start <= hi:
            print(f"{span} count {count.strip()}")


if __name__ == "__main__":
    main()
