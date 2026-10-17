#!/usr/bin/env python3
"""a112 base 재고정(1e25b3a3, 사용자 승인 2026-10-04) — 새 창에서 처음 요구되는 시험 함수 셋의 경량 번들(render_5222_bundles.main 재사용).

옛 base(aeeb209e) 창에서는 이 셋이 a112 가 만든 파일의 「새 함수」라 요구되지 않았다. 새 base(1e25b3a3)에서는 파일이 이미 있으므로 8.5 응답 로트(6f5b0df6)의
편집이 「기존 함수의 편집」이 된다. 편집 전 번들: `analysis/measurements/repin-1e25b3a3/pre-edit/`(178cc196 워크트리에서 render_pre_edit.py).
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

LEDGER = "`analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`"
T = "이 시험 자신(시험 코드) — `go test -tags tossos_testseams ./internal/app/engine` GREEN(`lot-8.5-R/verify-8.5-R.log`)"
NEW = "yes — 8.5 응답 로트에서 새 갈래(편집 전 없음)"
SAME = "no — 시험 코드, 갈래 불변"
FUNCS = [
    {"bundle": "internal-app-engine--testamarketwithmorescopesthanthequeueholdsclosesinsteadofdroppingone", "file": "internal/app/engine/a112_coordinator_test.go",
     "func": "TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne", "title": "TestAMarketWithMoreScopesThanTheQueueHoldsClosesInsteadOfDroppingOne (시험)",
     "revision": "편집 전 5 분기 → 7(B6 · B7 새로, 끝에 덧붙음 — 앞 다섯은 번호 그대로). 8.5 응답 로트 ⑦: 검증된 관문 아래에서 같은 넘침을 다시 돌려 QUEUE_OVERFLOW 닫힘이 판정 "
                 "활성화를 싣는지 단언(보이스 3 P2-1 — 변이 E7).",
     "scen": {
         "B1": ("종목 Capacity+1 개 생성", T, SAME, "yes"),
         "B2": ("미선언 실행의 사유 ≠ QUEUE_OVERFLOW → Fatal", T, SAME, "yes"),
         "B3": ("닫힌 시장이 열림 · 항목 있음 → Fatal", T, SAME, "yes"),
         "B4": ("버린 수 0 → Fatal(조용한 유실 금지)", T, SAME, "yes"),
         "B5": ("넘침이 중재 코드를 빌림 → Fatal", T, SAME, "yes"),
         "B6": ("**(새)** 관문 아래 실행의 사유 ≠ QUEUE_OVERFLOW → Fatal", T, NEW, "yes"),
         "B7": ("**(새)** QUEUE_OVERFLOW 닫힘이 판정 활성화를 안 실음 → Fatal", T, f"{NEW}; 변이 E7 CAUGHT({LEDGER})", "yes"),
     },
     "invariants": ["큐 넘침은 시장을 닫고(항목 0) 버린 수를 세며 중재 코드를 빌리지 않는다 — 관문 아래에서도 같고, 그 닫힘이 판정 활성화를 싣는다."],
     "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]},
    {"bundle": "internal-app-engine--collectoverflowing", "file": "internal/app/engine/a112_coordinator_test.go", "func": "collectOverflowing",
     "title": "collectOverflowing (시험 도우미)",
     "revision": "편집 전 5 분기 → 6(B6 새로, 끝에 덧붙음). 8.5 응답 로트 ⑦: 가변 인자 configure 로 수집 직전 적재기를 바꿀 수 있게(관문 아래 실행).",
     "scen": {
         "B1": ("종목마다 경로 항목 생성", T, SAME, "yes"),
         "B2": ("소유자 열쇠 생성 실패 → Fatal", T, SAME, "도달 불가(고정 입력)"),
         "B3": ("후보 경로 픽스처 실패 → Fatal", T, SAME, "도달 불가(고정 입력)"),
         "B4": ("제안 적재 스텁의 대상 순회", T, SAME, "yes"),
         "B5": ("수락 결과 픽스처 실패 → Fatal", T, SAME, "도달 불가(고정 입력)"),
         "B6": ("**(새)** configure 적용", T, NEW, "yes"),
     },
     "invariants": ["KR 에 종목을 원하는 수만큼 두고 각 종목이 지속형 한 레인으로만 제안하게 하는 시험 도우미 — configure 는 수집 직전에만 적재기를 바꾼다."],
     "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]},
    {"bundle": "internal-app-engine--testagatedfamilymustnotshrinkthemarketintotheexactlyonevalve", "file": "internal/app/engine/a112_family_gate_test.go",
     "func": "TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve", "title": "TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve (시험)",
     "revision": "편집 전 8 분기 → 9(B9 새로, 끝에 덧붙음). 8.5 응답 로트 ⑦: FAMILY_GATE_CLOSED 닫힘이 판정 활성화를 싣는지 단언(보이스 3 P2-1 — 변이 E9).",
     "scen": {
         "B1": ("KR 레인 순회", T, SAME, "yes"),
         "B2": ("지속형이 아니면 건너뜀", T, SAME, "yes"),
         "B3": ("비정상 실패로도 안 잠기면 → Fatal", T, SAME, "yes"),
         "B4": ("잠근 레인 수 ≠ 1 → Fatal", T, SAME, "yes"),
         "B5": ("시장이 열림 → Fatal(고장이 시스템을 관대하게)", T, SAME, "yes"),
         "B6": ("사유 ≠ FAMILY_GATE_CLOSED → Fatal", T, SAME, "yes"),
         "B7": ("거절 수 ≠ 경로 수 → Fatal", T, SAME, "yes"),
         "B8": ("닫힌 시장이 dispatch 에 건넴 → Fatal", T, SAME, "yes"),
         "B9": ("**(새)** FAMILY_GATE_CLOSED 닫힘이 판정 활성화를 안 실음 → Fatal", T, f"{NEW}; 변이 E9 CAUGHT({LEDGER})", "yes"),
     },
     "invariants": ["잠긴 가족이 소유자 범위 하나를 통째로 지우면 시장을 닫는다(정확히-하나 관문 만족 금지) — 그 닫힘이 판정 활성화를 싣는다."],
     "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]},
]


if __name__ == "__main__":
    for spec in FUNCS:
        (base.FL / spec["bundle"]).mkdir(parents=True, exist_ok=True)
    base.main(FUNCS, tag="a112 repin-1e25b3a3", pre="repin-1e25b3a3", ledger=LEDGER)
