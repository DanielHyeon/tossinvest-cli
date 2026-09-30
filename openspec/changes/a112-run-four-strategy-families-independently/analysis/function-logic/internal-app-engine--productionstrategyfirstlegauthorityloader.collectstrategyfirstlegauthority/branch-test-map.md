# Branch Test Map: `collectStrategyFirstLegAuthority`

- Source SHA-256: `26bca2d7d1c0ae0ff24a70ec64eb9a660ddf0ac27a60cd4947a40487cbe303be`; AST branch locations are authoritative.
- Revision: **modified (a112 6.2 봉인 로트, 2026-10-01).** 편집 전 9 분기 → 11: 편집 전 B2(준비 + 개수)를 준비(B2)와 개수(B4)로 나누고 그 사이에 소유자 범위 선택 실패(B3)를 더했다. 편집 전 B3~B9 는 B5~B11(조건 불변). 편집 전 번들(5.5 의 B3 실측 기록 포함)은 `analysis/measurements/lot-6.2-seal/pre-edit/`. 변이 원장 `analysis/measurements/lot-6.2-seal/mutation-6.2-seal.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 211:2 — loader · ctx · 시계 · 원장 · Guardian 부재 → 발급 불가 | 편집 전과 같음(분기 불변) — 이 로트의 시험 없음 | no | 측정 안 함 |
| B2 | if at 217:2 — 위험 · 환율 · 계좌 · 일정 준비 · 활성화 부재 → `paired production authority is incomplete for market`. **편집: 개수 조건을 떼어 냄**(아래 B4) | 편집 전 B2 의 준비 조건 절반 — 이 로트의 새 시험 없음(준비 상태는 형제 시험들이 이미 잰다) | no | 측정 안 함 |
| B3 | if at 224:2 — **(새) 소유자 범위 선택 실패** — 조립의 권한 쌍에 accepted 범위가 정확히 하나가 아님(0: 미선택 범위 · 다른 시장, 2+: 범위당 하나 붕괴) → `production proposal identity changed: owner scope is not uniquely authorized by the assembly` | `a112_first_leg_owner_scope_seal_test.go` `TestTheFirstLegSealRefusesEveryForgeryAxis/unselected_scope` · `/other_market` · `TestTheFirstLegSealRefusesAnOwnerScopeTheAssemblyHoldsTwice` | yes — `red-6.2-seal.log`(셋 FAIL: 편집 전에는 identity 문구로 거절) | yes |
| B4 | if at 228:2 — **(새, 편집 전 B2 에서 분리) 시장 단위 개수 관문** `len(proposal.entries) != 1` — 봉인이 아니라 시장당 하나 상한, 걷어 내는 일은 5.2.2.2 | `a112_first_leg_owner_scope_seal_test.go` `TestTheFirstLegSealSelectsByScopeBeforeTheMarketCountGate`(범위 선택이 먼저 성공해야 이 문구) · `a112_owner_scope_handoff_test.go` `TestTwoOwnerScopesStillPlaceNothingBecauseTheFirstLegGuardRefuses`(두 순서) | 편집 전에도 같은 문구로 거절(선택 기제 변이 S04 · S06 이 이 시험을 빨갛게 함) | yes |
| B5 | if at 232:2 — 봉인된 identity 대조(편집 전 B3, 조건 불변) — 선택된 항목의 `Lineage.Identity` · `ExecutionTerms.Identity()` 와 accepted 비교 | `a112_first_leg_owner_scope_seal_test.go` ① 같은 범위 패자 · ② 게이트된 레인 · ⑤ 조건 재작성(선택 실패 문구 없이 정확히 이 문구) + backstop 셋 `TestFirstLegAuthorityRefusesAProposalItDidNotAuthorize` · `…ASiblingCampaignOnTheSameSymbol` · `…RewrittenExecutionTermsUnderTheSameLineage` | 편집 전 번들의 5.5 실측(변이) — 이 로트 변이 S01 · S02 · S03 | yes |
| B6 | if at 236:2 — 위험 권한 범위 불일치(편집 전 B4) | 분기 불변 — 편집 전 번들 서술 | no | 측정 안 함 |
| B7 | if at 241:2 — 포지션 캠페인 CAS 변경(편집 전 B5) | 분기 불변 — 편집 전 번들 서술 | no | 측정 안 함 |
| B8 | range at 248:2 — 위험 버킷 항목 순회(편집 전 B6) | 분기 불변 | no | 측정 안 함 |
| B9 | if at 259:2 — 가격 단위 무효(편집 전 B7) | 분기 불변 | no | 측정 안 함 |
| B10 | if at 268:3 — 노출 스냅숏 만료(collect 클로저, 편집 전 B8) | 분기 불변 | no | 측정 안 함 |
| B11 | if at 272:3 — 예약 버전 읽기 실패(collect 클로저, 편집 전 B9) | 분기 불변 | no | 측정 안 함 |
