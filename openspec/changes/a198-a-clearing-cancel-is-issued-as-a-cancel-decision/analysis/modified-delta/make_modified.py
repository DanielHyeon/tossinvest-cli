#!/usr/bin/env python3
"""a198 — 정본 order-execution 의 Requirement 블록들을 MODIFIED 로 뜸.

정본 원문을 그대로 떠서 의도한 치환(EDITS)만 한다. 치환마다 원문 블록에 정확히 1회 있음을 단언하고, 결과에서 새 문자열이
정확히 1회 있음을 단언한 뒤 역치환으로 원문 블록이 바이트 그대로 돌아옴을 확인한다(= 의도 편집 외 차이 0). 대조 diff 를
남긴다. 재실행하면 같은 결과가 나와야 한다(멱등 — 출력은 원문에서 매번 다시 만듦).

블록:
  1. 「취소의 일시적 거절 처리」 — D7(보강 1).
  2. 「IN_DOUBT 해소」 — D9·D10(보강 2): 기동 ACKED 문단, 세션 해소 「대상」 문단, 시나리오 하나 수정 + 시나리오 하나 추가.

사용: python3 make_modified.py <repo-root>
출력: analysis/modified-delta/{modified-block.md, compare.diff, edits.md} — delta 파일에는 MODIFIED 블록들 + added.md.
"""
import difflib
import os
import sys

root = sys.argv[1]
canon = os.path.join(root, "openspec/specs/order-execution/spec.md")
change = os.path.join(root, "openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision")
here = os.path.join(change, "analysis/modified-delta")

BLOCKS = [
    ("### Requirement: 취소의 일시적 거절 처리\n", [
        (
            "- **WHEN** 게이트웨이의 취소가 409 로 거절되고 본문 code 가 `already-processing` 이 아니거나 두 code 자리가 다르면\n"
            "- **THEN** 이 처리 이전과 같이 결과 불명으로 기록된다\n",
            "- **WHEN** 게이트웨이의 취소가 409 로 거절되고 본문 code 가 `already-processing` 이 아니거나 두 code 자리가 다르면"
            "(엔진 게이트웨이의 취소가 `already-filled`·`already-canceled`·`already-rejected` 로 거절되는 경우는 "
            "「취소가 이미 끝난 주문을 이유로 거절되면 확정 거절로 분류한다」가 정한다)\n"
            "- **THEN** 그 밖의 경우는 이 처리 이전과 같이 결과 불명으로 기록된다\n",
            "Scenario 「다른 사유의 409」: 엔진 취소의 세 code 예외를 WHEN 에, THEN 을 「그 밖의 경우」로",
        ),
    ]),
    ("### Requirement: IN_DOUBT 해소\n", [
        (
            "**기동은 ACKED로 남은 발주를 기록된 번호로만 확정하고, 나머지는 명명해 알려야 한다.**",
            "**기동은 ACKED로 남은 발주를 기록된 번호로만 확정하고, ACKED로 남은 취소는 기록된 번호로 쓰기를 마치며, "
            "나머지는 명명해 알려야 한다.**",
            "D9: 기동 ACKED 문단 제목에 취소",
        ),
        (
            "그 외의 경우(번호·종목 불일치, 읽기 실패, 응답 해석 불가, 기록 번호 없음, 취소·정정 attempt, 확정 전이의 원장 쓰기 실패)에는",
            "재시작 뒤 ACKED로 남은 취소 attempt에 기록된 브로커 주문 번호가 있으면, 기동은 브로커를 읽지 않고 그 번호로 "
            "접수 확정으로 종결해야 한다(SHALL — 취소에는 접수 뒤 확인 읽기가 없으므로 그 상태는 브로커가 취소 접수를 답했고 "
            "최종 쓰기만 남은 것이며, 기동은 디스패치가 하려던 쓰기를 마칠 뿐 추론하지 않는다; 그 번호는 취소가 새로 받은 "
            "식별자로 대상 주문의 판정에 쓰이지 않는다). 그 확정 전이의 원장 쓰기가 실패하면 아래 그 외의 경우와 같이 다룬다. "
            "그 외의 경우(번호·종목 불일치, 읽기 실패, 응답 해석 불가, 기록 번호 없음, 정정 attempt, 확정 전이의 원장 쓰기 실패)에는",
            "D9: 기록 번호 있는 ACKED 취소는 그 번호로 접수 확정, 「그 외」 목록에서 취소 제거",
        ),
        (
            "바이트 일치 외의 어떤 추론으로도 ACKED attempt를 종결하거나 모호로 올려서는 안 되며(SHALL NOT)",
            "바이트 일치 외의 어떤 추론으로도 ACKED 발주·정정 attempt를 종결하거나 모호로 올려서는 안 되며(SHALL NOT)",
            "D9: 추론 금지의 대상을 발주·정정으로(취소는 위 문장 — 추론 없는 쓰기 완료)",
        ),
        (
            "- **대상**: 세션 중 해소의 대상은 `IN_DOUBT`이면서 발주인 attempt뿐이다(SHALL). 취소·정정 attempt는 기동이 해소한다.",
            "- **대상**: 세션 중 해소(조회 대조)의 대상은 `IN_DOUBT`이면서 발주인 attempt뿐이다(SHALL). 정정 attempt는 기동이 "
            "해소한다. `IN_DOUBT`인 취소 attempt는 조회 대조에 넣지 않고 「세션은 IN_DOUBT 로 남은 취소를 대상 주문의 종결 기록으로 "
            "닫는다」가 브로커 조회 없이 닫으며, 그것이 닫지 못한 취소는 기동이 해소한다. 실행자는 취소 attempt에 이 Requirement의 "
            "처음 봄·오래 남음·park 알림을 내지 않는다.",
            "D10: 세션 해소 대상 문단 — 취소는 원장 종결 기록으로 닫음, 알림은 발주만",
        ),
        (
            "- **WHEN** 세션 중 해소 실행자가 RECORDED·DISPATCH_STARTED·ACKED 상태이거나 취소·정정인 미종결 attempt를 읽으면\n"
            "- **THEN** 그 attempt의 상태는 바뀌지 않고 해소 절차에 넣어지지 않는다\n",
            "- **WHEN** 세션 중 해소 실행자가 RECORDED·DISPATCH_STARTED·ACKED 상태이거나 정정인 미종결 attempt, 또는 대상 주문의 "
            "종결 기록이 없는 `IN_DOUBT` 취소 attempt를 읽으면\n"
            "- **THEN** 그 attempt의 상태는 바뀌지 않고 조회 대조 해소 절차에 넣어지지 않는다\n",
            "D10: 시나리오 「세션 중에는 진행 중인 전송을 종결하지 않는다」",
        ),
        (
            "#### Scenario: 운영자가 park된 attempt를 비수용으로 닫는다\n",
            "#### Scenario: 재시작 뒤 ACKED로 남은 취소\n"
            "- **WHEN** 브로커가 취소를 접수해 번호가 기록된 뒤 결과를 쓰지 못한 채 프로세스가 멈췄다가 재시작하면\n"
            "- **THEN** 기동은 브로커를 읽지 않고 그 번호로 그 취소를 접수 확정으로 종결하고, 정정 attempt의 ACKED는 종전대로 "
            "상태를 바꾸지 않고 명명해 알린다\n"
            "\n"
            "#### Scenario: 운영자가 park된 attempt를 비수용으로 닫는다\n",
            "D9: 시나리오 「재시작 뒤 ACKED로 남은 취소」 추가(기존 시나리오 앞)",
        ),
    ]),
]

text = open(canon, encoding="utf-8").read()
outs, diffs, notes = [], [], []
for header, edits in BLOCKS:
    assert text.count(header) == 1, ("header not exactly once", header)
    start = text.index(header)
    end = text.find("\n### Requirement:", start + len(header))
    block = text[start:end + 1] if end != -1 else text[start:]
    assert block.endswith("\n")
    out = block
    for old, new, _ in edits:
        assert out.count(old) == 1, ("anchor not exactly once", old[:60])
        out = out.replace(old, new)
    # 역치환으로 원문 복원 — 의도 편집 외 차이 0 의 증명.
    back = out
    for old, new, _ in reversed(edits):
        assert back.count(new) == 1, ("new text not exactly once", new[:60])
        back = back.replace(new, old)
    assert back == block, "the block differs from the canonical beyond the intended edits"
    diff = list(difflib.unified_diff(block.splitlines(True), out.splitlines(True), "canonical", "modified"))
    changed_old = sum(1 for l in diff if l.startswith("-") and not l.startswith("---"))
    changed_new = sum(1 for l in diff if l.startswith("+") and not l.startswith("+++"))
    outs.append(out)
    diffs.append(diff)
    notes.append((header.strip(), block.count("\n"), edits, changed_old, changed_new))

os.makedirs(here, exist_ok=True)
modified = "\n".join(o.rstrip("\n") + "\n" for o in outs)
open(os.path.join(here, "modified-block.md"), "w", encoding="utf-8").write(modified)
with open(os.path.join(here, "compare.diff"), "w", encoding="utf-8") as f:
    for d in diffs:
        f.writelines(d)
with open(os.path.join(here, "edits.md"), "w", encoding="utf-8") as f:
    f.write("# MODIFIED 의도 편집 목록\n")
    for header, lines, edits, co, cn in notes:
        f.write(f"\n- 원문: openspec/specs/order-execution/spec.md 의 `{header}` 블록({lines} 줄)\n")
        for i, (_, _, why) in enumerate(edits, 1):
            f.write(f"  - 편집 {i}: {why}\n")
        f.write(f"  - 결과 diff: 바뀐 줄 원문 {co} → {cn}, 그 밖의 차이 0(역치환 단언)\n")
# delta 파일 = MODIFIED(생성) + ADDED(손으로 쓴 added.md). 재실행하면 같은 파일이 나와야 함.
added = open(os.path.join(here, "added.md"), encoding="utf-8").read()
delta = "## MODIFIED Requirements\n\n" + "\n".join(o.rstrip("\n") + "\n" for o in outs) + "\n" + added
delta_path = os.path.join(change, "specs/order-execution/spec.md")
open(delta_path, "w", encoding="utf-8").write(delta)
# 되읽어 MODIFIED 부분이 생성 블록과 같은지 다시 확인
back = open(delta_path, encoding="utf-8").read()
mod = back[len("## MODIFIED Requirements\n\n"):back.index("\n## ADDED Requirements")]
assert mod.rstrip("\n") == modified.rstrip("\n"), "delta file drifted from the generated MODIFIED blocks"
print("ok", [(n[0], n[3], n[4]) for n in notes])
