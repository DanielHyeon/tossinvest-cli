#!/usr/bin/env python3
"""task 7.5.6 — 저장소의 커밋된 번들에서 BOM · UTF-16 · 좌표 없는 행 때문에 열거형 호출 감사가 **실제로** 꺼진 change 가 있는가.

사용자 결정(2026-09-27): 실제 오판정 사례가 있으면 그것만 고치고, 없으면 backlog. 그래서 **먼저 잰다**.

change 디렉터리(활성 + 아카이브)마다 게이트와 같은 입력으로 감사 스위치(`role_check.call_enumeration_in_use`)를 돌리고, 같은
번들을 **관대하게 읽은** 판(BOM 을 지우고 · 좌표 없는 행을 빼고)으로도 돌린다. 둘이 갈리는 change 가 "그 모양 때문에 감사가 꺼진"
사례다. 곁들여 번들 파일 중 BOM · UTF-16 BOM · NUL 바이트 · UTF-8 이 아닌 파일을 센다. 저장소는 읽기만 한다.

    python3 756_census.py
"""
import subprocess
import sys
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))
import check_analysis  # noqa: E402
import role_check  # noqa: E402


def lenient(bundles) -> bool:
    """감사 스위치의 관대한 판 — 글에서 BOM 을 지우고, 표의 행 중 좌표가 없는 행을 빼고 견준다."""
    for text, value in bundles:
        calls = role_check._coords((value or {}).get("calls"))
        if not calls:
            continue
        for table in role_check._tables(text.replace("﻿", "").splitlines()):
            cited = [coord for coord in (role_check._first_coord(row) for row in table[1:]) if coord is not None]
            if cited and cited == calls:
                return True
    return False


def main() -> None:
    head = subprocess.run(["git", "rev-parse", "--short=12", "HEAD"], cwd=REPO, capture_output=True, text=True).stdout.strip()
    changes = REPO / "openspec" / "changes"
    homes = sorted([p for p in changes.iterdir() if p.is_dir() and p.name != "archive"]
                   + [p for p in (changes / "archive").iterdir() if p.is_dir()])
    files = boms = utf16 = nuls = not_utf8 = bundles_total = 0
    differing = []
    audited = []
    for home in homes:
        analysis = home / "analysis" / "function-logic"
        if not analysis.is_dir():
            continue
        pairs = []
        for target in sorted(p for p in analysis.iterdir() if p.is_dir()):
            bundles_total += 1
            for item in target.iterdir():
                if not item.is_file():
                    continue
                raw = item.read_bytes()
                files += 1
                boms += raw.startswith(b"\xef\xbb\xbf")
                utf16 += raw.startswith((b"\xff\xfe", b"\xfe\xff"))
                nuls += b"\0" in raw
                try:
                    raw.decode("utf-8")
                except UnicodeDecodeError:
                    not_utf8 += 1
            ast_path = target / "ast.json"
            held = ast_path.read_bytes() if ast_path.is_file() else None
            try:
                text = check_analysis._bundle_text(target, held)
            except OSError:
                continue
            pairs.append((text, check_analysis._parsed(held)))
        gate, loose = role_check.call_enumeration_in_use(pairs), lenient(pairs)
        if gate:
            audited.append(home.name)
        if gate != loose:
            differing.append((home.name, gate, loose))
    print(f"HEAD {head} · change 디렉터리 {len(homes)} · 번들 {bundles_total} · 번들 파일 {files}")
    print(f"UTF-8 BOM {boms} · UTF-16 BOM {utf16} · NUL 든 파일 {nuls} · UTF-8 아닌 파일 {not_utf8}")
    print(f"감사가 켜진 change {len(audited)}: {', '.join(audited)} (계측기 대조 — 0 이면 이 측정은 아무것도 못 잰 것이다)")
    print(f"감사 스위치가 관대한 판과 갈리는 change {len(differing)}")
    for row in differing:
        print("  ", row)


if __name__ == "__main__":
    main()
