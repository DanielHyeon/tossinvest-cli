#!/usr/bin/env python3
"""a112 6.2 봉인 로트 — collectStrategyFirstLegAuthority 편집 번들(FLM/BTM)을 ast.json 에서 채운다. 좌표 · 호출 · return 은 AST 에서만 읽고,
저자는 분기의 뜻 · 시험 이름 · 편집 설명만 쓴다. 편집 전 번들은 analysis/measurements/lot-6.2-seal/pre-edit/."""
from __future__ import annotations

import json
from pathlib import Path

A = Path(__file__).resolve().parents[1]
B = A / "function-logic" / "internal-app-engine--productionstrategyfirstlegauthorityloader.collectstrategyfirstlegauthority"
ast = json.loads((B / "ast.json").read_text())
pos = lambda n: f"{n['at']['line']}:{n['at']['column']}"
br = {b["id"]: b for b in ast["branches"]}

SEAL = "`a112_first_leg_owner_scope_seal_test.go`"
SCEN = {
    "B1": ("loader · ctx · 시계 · 원장 · Guardian 부재 → 발급 불가", "편집 전과 같음(분기 불변) — 이 로트의 시험 없음", "no", "측정 안 함"),
    "B2": ("위험 · 환율 · 계좌 · 일정 준비 · 활성화 부재 → `paired production authority is incomplete for market`. **편집: 개수 조건을 떼어 냄**(아래 B4)",
           "편집 전 B2 의 준비 조건 절반 — 이 로트의 새 시험 없음(준비 상태는 형제 시험들이 이미 잰다)", "no", "측정 안 함"),
    "B3": ("**(새) 소유자 범위 선택 실패** — 조립의 권한 쌍에 accepted 범위가 정확히 하나가 아님(0: 미선택 범위 · 다른 시장, 2+: 범위당 하나 붕괴) → "
           "`production proposal identity changed: owner scope is not uniquely authorized by the assembly`",
           f"{SEAL} `TestTheFirstLegSealRefusesEveryForgeryAxis/unselected_scope` · `/other_market` · `TestTheFirstLegSealRefusesAnOwnerScopeTheAssemblyHoldsTwice`",
           "yes — `red-6.2-seal.log`(셋 FAIL: 편집 전에는 identity 문구로 거절)", "yes"),
    "B4": ("**(새, 편집 전 B2 에서 분리) 시장 단위 개수 관문** `len(proposal.entries) != 1` — 봉인이 아니라 시장당 하나 상한, 걷어 내는 일은 5.2.2.2",
           f"{SEAL} `TestTheFirstLegSealSelectsByScopeBeforeTheMarketCountGate`(범위 선택이 먼저 성공해야 이 문구) · "
           "`a112_owner_scope_handoff_test.go` `TestTwoOwnerScopesStillPlaceNothingBecauseTheFirstLegGuardRefuses`(두 순서)",
           "편집 전에도 같은 문구로 거절(선택 기제 변이 S04 · S06 이 이 시험을 빨갛게 함)", "yes"),
    "B5": ("봉인된 identity 대조(편집 전 B3, 조건 불변) — 선택된 항목의 `Lineage.Identity` · `ExecutionTerms.Identity()` 와 accepted 비교",
           f"{SEAL} ① 같은 범위 패자 · ② 게이트된 레인 · ⑤ 조건 재작성(선택 실패 문구 없이 정확히 이 문구) + backstop 셋 "
           "`TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize` · `…ASiblingCampaignOnTheSameSymbol` · `…RewrittenExecutionTermsUnderTheSameLineage`",
           "편집 전 번들의 5.5 실측(변이) — 이 로트 변이 S01 · S02 · S03", "yes"),
    "B6": ("위험 권한 범위 불일치(편집 전 B4)", "분기 불변 — 편집 전 번들 서술", "no", "측정 안 함"),
    "B7": ("포지션 캠페인 CAS 변경(편집 전 B5)", "분기 불변 — 편집 전 번들 서술", "no", "측정 안 함"),
    "B8": ("위험 버킷 항목 순회(편집 전 B6)", "분기 불변", "no", "측정 안 함"),
    "B9": ("가격 단위 무효(편집 전 B7)", "분기 불변", "no", "측정 안 함"),
    "B10": ("노출 스냅숏 만료(collect 클로저, 편집 전 B8)", "분기 불변", "no", "측정 안 함"),
    "B11": ("예약 버전 읽기 실패(collect 클로저, 편집 전 B9)", "분기 불변", "no", "측정 안 함"),
}

btm = [f"# Branch Test Map: `collectStrategyFirstLegAuthority`", "",
       f"- Source SHA-256: `{ast['source_sha256']}`; AST branch locations are authoritative.",
       "- Revision: **modified (a112 6.2 봉인 로트, 2026-10-01).** 편집 전 9 분기 → 11: 편집 전 B2(준비 + 개수)를 준비(B2)와 개수(B4)로 나누고 그 사이에 "
       "소유자 범위 선택 실패(B3)를 더했다. 편집 전 B3~B9 는 B5~B11(조건 불변). 편집 전 번들(5.5 의 B3 실측 기록 포함)은 "
       "`analysis/measurements/lot-6.2-seal/pre-edit/`. 변이 원장 `analysis/measurements/lot-6.2-seal/mutation-6.2-seal.tsv`.", "",
       "| Branch | Scenario anchor | Test | RED observed | GREEN observed |", "|---|---|---|---|---|"]
for bid in [b["id"] for b in ast["branches"]]:
    scenario, test, red, green = SCEN[bid]
    btm.append(f"| {bid} | {br[bid]['kind']} at {pos(br[bid])} — {scenario} | {test} | {red} | {green} |")
(B / "branch-test-map.md").write_text("\n".join(btm) + "\n")

flm = [f"# Function Logic Map: `collectStrategyFirstLegAuthority`", "",
       f"- Source: `{ast['file']}`", f"- Source SHA-256: `{ast['source_sha256']}`",
       f"- Signature: `{ast['signature']}`", f"- Source range: `{ast['start']['line']}:1`–`{ast['end']['line']}:2`",
       "- AST evidence: `ast.json` — **편집 뒤**(a112 6.2 봉인 로트).", "- Risk scan: `risk-pattern-report.md`.", "",
       "## Inputs and invariants", "",
       "- 입력: loader 자기 권한 쌍(조립이 새로 고침 때 중재한 제안 · 위험 · 환율 · 계좌 · 일정)과 건너온 `accepted`(봉투가 나른 결과).",
       "- **봉인 불변식(6.2):** 발급하는 1차 레그의 결과는 건너온 값이 아니라 **조립의 권한 쌍에서 소유자 범위로 다시 꺼낸 항목**이고, 그 항목의 봉인된 "
       "identity 가 accepted 와 같아야 한다. 범위는 accepted 계보에서 읽되 선택 기준일 뿐 대조 대상이 아니다(자기 참조 함정 회피 — "
       "`authorityForOwnerScope` 머리말).",
       "- 시장 단위 개수 관문(B4)은 봉인이 아니라 상한이고 5.2.2.2 가 걷어 낸다. 이 편집은 판정을 더 엄격하게만 한다(새로 통과하는 입력 0).", "",
       "## Branches and early returns", "",
       "- Exact AST return nodes: `" + ", ".join(pos(r) for r in ast["returns"]) + "`.", "",
       "| Branch | AST kind | Source location | Meaning |", "|---|---|---|---|"]
for bid in [b["id"] for b in ast["branches"]]:
    flm.append(f"| {bid} | {br[bid]['kind']} | {pos(br[bid])} | {SCEN[bid][0]} |")
flm += ["", "## Calls and live bindings", "", "| Callee expression | Position |", "|---|---|"]
flm += [f"| `{c['text']}` | {pos(c)} |" for c in ast["calls"]]
flm += ["", "## State mutations and fallbacks", "",
        "- 이 함수 자신은 원장을 쓰지 않는다(읽기: 캠페인 CAS · 예약 버전). 발급 값은 호출자가 q_final 입장에서 쓴다.", "",
        "## Safety conclusion", "",
        "- High-risk impact: yes(1차 레그 발급). 편집은 거절 갈래를 하나 더하고(B3) 선택 기준을 위치에서 소유자 범위로 바꿨다 — 편집 전 통과하던 입력 중 "
        "새로 통과하는 것은 없다(단일 항목 쌍에서 범위가 같으면 편집 전과 같은 항목을 고르고, 다르면 편집 전에도 identity 대조에서 거절됐다).",
        "- 위조 축 다섯 · 기제 둘은 행동 시험, 선택 기제 · 대조 약화 · 재유도 제거는 변이 S01~S07 이 CAUGHT."]
(B / "function-logic-map.md").write_text("\n".join(flm) + "\n")
print("rendered")
