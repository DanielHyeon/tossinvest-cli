# Branch Test Map: `strategyProposalAuthorityLoader.loadFamilyActivation`

- Source SHA-256: `50c775155bc8a12a5844ddbc30785f72e49b4fab7380146071e458f815d4c377`; AST branch locations are authoritative.
- Revision: **modified (a112 8.5-R, 2026-10-01).** 편집 전 2 분기 → 3(재번호 `lot-8.5-R/renumber.txt`, difflib 정렬): **B1 새로**(getenv nil → Unavailable, 8.5 응답 로트 ② — codex r2 P2). 편집 전 B1(US digest env) → B2, B2(US 위험 정책 env) → B3. 앞 판은 nil getenv 에서 공황했고 collect 의 recover 가 같은 주기 사유를 INTERNAL_FAILURE 로 덮었다.
- 편집 전 번들: `analysis/measurements/lot-8.5-R/pre-edit/internal-app-engine--strategyproposalauthorityloader.loadfamilyactivation/`. 변이 원장 `analysis/measurements/lot-8.5-R/mutation-8.5-R.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 161:2 — **(새)** getenv 가 nil → 맨 `ErrProductionFamilyActivationUnavailable` — 미선언이 아님(핀을 읽을 수 없음), 관문 되돌림(맨 sentinel 인 이유: 이 경로 호출은 `TestTheRollbackPathOnlyReads` 의 읽기 전용 허용 목록으로 묶임) | `a112_gate_decision_recompute_test.go` `TestANilEnvironmentReaderRollsTheGateBackInsteadOfPanicking` | yes — `red-8.5-R.log` — 편집 전 공황(nil 함수 호출) · collect 사유 INTERNAL_FAILURE | yes |
| B2 | if at 165:2 — US 시장이면 US 활성화 digest env | `TestStrategyProposalAuthorityLoadsKRUSConcurrently` · `TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening` | no — 갈래 불변(편집 전 B1) | yes |
| B3 | if at 187:2 — US 시장이면 US 위험 정책 env | `TestStrategyProposalAuthorityLoadsKRUSConcurrently` | no — 갈래 불변(편집 전 B2) | yes |
