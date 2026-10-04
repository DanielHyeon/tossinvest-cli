#!/usr/bin/env python3
"""a112 게이트 준비(2026-10-04, Manager 판정 — base 무관 로트) — 8.8.4 로트 A 가 편집한 시험 함수의 경량 번들(render_5222_bundles.main 재사용).

`TestProductionWorkersAreExactlyTheEightTheGoldenFroze` 는 현재 base(aeeb209e) 창에서 「새 함수」라 요구되지 않았지만, 권고 base(1e25b3a3 등)에서는 기존 함수의
편집이라 요구된다(분류 영수증 `measurements/gateprep-2026-10-04/`). 로트 A(f473d815)가 desired/effective 절을 빼고 Runtime 대조만 남겼다.
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import render_5222_bundles as base  # noqa: E402

LEDGER = "`analysis/measurements/lot-8.8.4-A/mutation-8.8.4-A.tsv`"
T = "이 시험 자신(시험 코드) — `go test ./internal/strategyworker` 무태그 · 태그 GREEN(`lot-8.8.4-A/verify-8.8.4-A.log`)"
FUNCS = [
    {"bundle": "internal-strategyworker--testproductionworkersareexactlytheeightthegoldenfroze", "file": "internal/strategyworker/golden_contract_test.go",
     "func": "TestProductionWorkersAreExactlyTheEightTheGoldenFroze", "title": "TestProductionWorkersAreExactlyTheEightTheGoldenFroze (시험)",
     "revision": "8.8.4 로트 A(f473d815): 공허한 desired/effective 절(영값 활성화 — `lookup` 이 첫 줄에서 OFF 를 돌려줘 상수-대-상수)을 빼고 Runtime 대조만 남겼다. "
                 "그 행동 단언은 태그 파일 `a112_golden_desired_effective_test.go` `TestTheGoldenOffIsEachWorkersDefaultAndOnlyItsSignedActivationFlipsIt` 로 옮겼다.",
     "scen": {
         "B1": ("골든 자기 모순(worker_count ≠ descriptors 수) → Fatal", T, "no — 시험 코드", "yes"),
         "B2": ("생산 worker 수 ≠ 골든 → Fatal", T, "no — 시험 코드", "yes"),
         "B3": ("골든 서술자 순회", T, "no — 시험 코드", "yes"),
         "B4": ("열쇠(시장 · 가족 · 레인 · 버전) 드리프트 → Error", T, "no — 시험 코드", "yes"),
         "B5": ("horizon 드리프트 → Error", T, "no — 시험 코드", "yes"),
         "B6": ("runtime 드리프트 → Error(로트 A — desired/effective 절 제거 뒤 남은 대조)", T,
                f"no — 시험 코드; 옮겨 간 desired/effective 행동 단언은 변이 A2 CAUGHT({LEDGER})", "yes"),
     },
     "invariants": ["골든 서술자 순서 · 열쇠 · horizon · runtime 을 생산 worker 와 대조한다 — desired/effective 는 태그 시험이 활성화로 잰다."],
     "state": "시험 코드.", "safety": ["시험 코드 — 생산 경로 없음."]},
]


if __name__ == "__main__":
    for spec in FUNCS:
        (base.FL / spec["bundle"]).mkdir(parents=True, exist_ok=True)
    base.main(FUNCS, tag="a112 gate-prep", pre="gateprep-2026-10-04", ledger=LEDGER)
