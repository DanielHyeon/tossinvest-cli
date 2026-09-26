#!/usr/bin/env python3
"""task 6.4 보수(P2) — `check` 한 번에 `_head_commit` 이 몇 번 불리는가, 활성 change 전수.

`_head_commit` docstring 은 7.5.22 실측 "3회 24건 · 0회 3건(head 를 묻기 전에 거절)" 을 적었다. 6.4(b) 가 `head` 를
`resolve_base` **앞에서** 풀게 순서를 바꿔 "0회" 의 뜻이 바뀌었다 — 다시 잰다. 모집단: `openspec/changes/` 의 활성
디렉터리 전부(아카이브 제외). 계측은 모듈의 `_head_commit` 을 세는 감싸개로 바꾸는 것뿐이고 판정은 그대로다.

    python3 64r_head_calls.py
"""
import collections
import sys
import time
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))

import check_analysis  # noqa: E402

real = check_analysis._head_commit
calls = [0]


def counted(root: Path) -> str:
    calls[0] += 1
    return real(root)


check_analysis._head_commit = counted
head = real(REPO)
changes = sorted(p.name for p in (REPO / "openspec" / "changes").iterdir() if p.is_dir() and p.name != "archive")
tally = collections.Counter()
for name in changes:
    calls[0] = 0
    started = time.monotonic()
    try:
        errors = check_analysis.check(name, REPO)
        first = errors[0][:90] if errors else "[]"
    except Exception as exc:  # 결함도 답이다
        first = f"{type(exc).__name__}: {exc}"[:90]
    tally[calls[0]] += 1
    print(f"{calls[0]}회 {time.monotonic() - started:6.1f}s {name} · {first}", flush=True)
print(f"\nHEAD {head} · 활성 change {len(changes)} · " + " · ".join(f"{k}회 {v}건" for k, v in sorted(tally.items())))
