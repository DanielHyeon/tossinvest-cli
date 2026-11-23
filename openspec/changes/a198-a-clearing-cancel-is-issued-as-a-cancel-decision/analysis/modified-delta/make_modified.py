#!/usr/bin/env python3
"""a198 — 정본 order-execution 「취소의 일시적 거절 처리」 를 MODIFIED 블록으로 뜸.

정본 원문을 그대로 떠서 의도한 치환(EDITS)만 한다. 치환마다 원문에 정확히 1회 있음을 단언하고, 결과와 원문의 차이가
EDITS 의 줄뿐임을 difflib 으로 확인해 대조 diff 를 남긴다. 재실행하면 같은 결과가 나와야 한다(멱등 — 출력은 원문에서 매번 다시 만듦).

사용: python3 make_modified.py <repo-root>
출력: analysis/modified-delta/{modified-block.md, compare.diff, edits.md} — delta 파일에는 modified-block.md 를 MODIFIED 절로 넣음.
"""
import difflib
import os
import sys

root = sys.argv[1]
canon = os.path.join(root, "openspec/specs/order-execution/spec.md")
here = os.path.join(root, "openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/modified-delta")
HEADER = "### Requirement: 취소의 일시적 거절 처리\n"

EDITS = [
    (
        "- **WHEN** 게이트웨이의 취소가 409 로 거절되고 본문 code 가 `already-processing` 이 아니거나 두 code 자리가 다르면\n"
        "- **THEN** 이 처리 이전과 같이 결과 불명으로 기록된다\n",
        "- **WHEN** 게이트웨이의 취소가 409 로 거절되고 본문 code 가 `already-processing` 이 아니거나 두 code 자리가 다르면"
        "(엔진 게이트웨이의 취소가 `already-filled`·`already-canceled`·`already-rejected` 로 거절되는 경우는 "
        "「취소가 이미 끝난 주문을 이유로 거절되면 확정 거절로 분류한다」가 정한다)\n"
        "- **THEN** 그 밖의 경우는 이 처리 이전과 같이 결과 불명으로 기록된다\n",
        "Scenario 「다른 사유의 409」: 엔진 취소의 세 code 예외를 WHEN 에, THEN 을 「그 밖의 경우」로",
    ),
]

text = open(canon, encoding="utf-8").read()
start = text.index(HEADER)
assert text.count(HEADER) == 1
end = text.find("\n### Requirement:", start + len(HEADER))
block = text[start:end + 1] if end != -1 else text[start:]
assert block.endswith("\n")

out = block
for old, new, _ in EDITS:
    assert out.count(old) == 1, ("anchor not exactly once", old[:60])
    out = out.replace(old, new)

diff = list(difflib.unified_diff(block.splitlines(True), out.splitlines(True), "canonical", "modified"))
changed_old = [l[1:] for l in diff if l.startswith("-") and not l.startswith("---")]
changed_new = [l[1:] for l in diff if l.startswith("+") and not l.startswith("+++")]
expect_old = [l for o, _, _ in EDITS for l in o.splitlines(True)]
expect_new = [l for _, n, _ in EDITS for l in n.splitlines(True)]
assert changed_old == expect_old and changed_new == expect_new, "diff contains more than the intended edits"

os.makedirs(here, exist_ok=True)
open(os.path.join(here, "modified-block.md"), "w", encoding="utf-8").write(out)
open(os.path.join(here, "compare.diff"), "w", encoding="utf-8").writelines(diff)
with open(os.path.join(here, "edits.md"), "w", encoding="utf-8") as f:
    f.write("# MODIFIED 의도 편집 목록\n\n")
    f.write(f"- 원문: openspec/specs/order-execution/spec.md 의 `{HEADER.strip()}` 블록({block.count(chr(10))} 줄)\n")
    for i, (_, _, why) in enumerate(EDITS, 1):
        f.write(f"- 편집 {i}: {why}\n")
    f.write(f"- 결과 diff: 바뀐 줄 원문 {len(changed_old)} → {len(changed_new)}, 그 밖의 차이 0(스크립트 단언)\n")
# delta 파일 = MODIFIED(생성) + ADDED(손으로 쓴 added.md). 재실행하면 같은 파일이 나와야 함.
added = open(os.path.join(here, "added.md"), encoding="utf-8").read()
delta = "## MODIFIED Requirements\n\n" + out.rstrip("\n") + "\n\n" + added
delta_path = os.path.join(root, "openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/specs/order-execution/spec.md")
open(delta_path, "w", encoding="utf-8").write(delta)
# 되읽어 MODIFIED 부분이 원문과 의도 편집 외 차이 0 인지 다시 확인
back = open(delta_path, encoding="utf-8").read()
mod = back[len("## MODIFIED Requirements\n\n"):back.index("\n## ADDED Requirements")]
assert mod.rstrip("\n") == out.rstrip("\n"), "delta file drifted from the generated MODIFIED block"
print("ok", len(changed_old), len(changed_new))
