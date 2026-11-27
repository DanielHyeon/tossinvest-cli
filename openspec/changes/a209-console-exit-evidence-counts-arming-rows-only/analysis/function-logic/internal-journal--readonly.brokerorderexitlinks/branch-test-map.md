# Branch Test Map: `ReadOnly.BrokerOrderExitLinks`

Source: `internal/journal/account_views.go` (373-593) — AST branches 28. a209 가 뜻을 바꾼 것은 B21 · B22 의 입력(SQL 이 세는 행)뿐임.
RED/GREEN 은 `analysis/tdd/red-journal.txt` · `green-journal.txt` 의 실제 실행으로만 yes. 기존 시험은 journal 패키지 선택 실행으로 통과만 확인함(커버리지 count 미측정).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | scope 0 | `TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL` | no | no |
| B2 | scope 상한 | 같음 | no | no |
| B3 | 정규화 루프 | 모든 시험 | no | no |
| B4 | market 오류 | `TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL` | no | no |
| B5 | 빈 주문 · 계좌 | 같음 | no | no |
| B6 | 거래일 오류 | 같음 | no | no |
| B7 | 인코딩 실패 | 도달 불가 | no | no |
| B8 | 질의 실패 | 기존(무변경) | no | no |
| B9 | 틀 초기화 | 모든 시험 | no | no |
| B10 | 행 루프 | 모든 시험 | no | no |
| B11 | 행 상한 | `TestBrokerOrderExitLinksRowSentinelIsStructurallyAboveTheQueryMaximum` | no | no |
| B12 | Scan 실패 | 기존(무변경) | no | no |
| B13 | 신원 불일치 | 기존(무변경) | no | no |
| B14 | 엔진 주문 | `TestReadOnlyLinksBrokerOrdersToExitSnapshotsOnlyThroughTheAttemptIntent` | no | no |
| B15 | 사유 switch | 아래 | no | no |
| B16 | 순환 | `TestBrokerOrderExitLinksFailsClosedForPureSingleParentCycle` | no | no |
| B17 | 부모 다수 | `TestBrokerOrderExitLinksDoesNotExpandWideBranchingLineage` | no | no |
| B18 | 범위 밖 부모 | `TestBrokerOrderExitLinksRefusesIncompleteAmendEvidence` | no | no |
| B19 | 깊이 상한 | `TestBrokerOrderExitLinksFailsClosedOnlyWhenLineageExceedsDepthBound` | no | no |
| B20 | PLACE 다수 | `TestBrokerOrderExitLinksAttributesOnlyConfirmedPlaceAttempts` | no | no |
| B21 | 무장 행 + 훅 행 → 안 걸림 / 무장 행 둘 → 걸림 | `TestA209AFilledExitOrderLinksItsArmingRowDespiteTheFillHook` · `TestA209AReleasedExitOrderLinksItsArmingRowDespiteTheReleaseHook` / `TestA209TwoArmingRowsStayAmbiguousWithAHookRow` · `TestReadOnlyMarksDuplicateExitEvidenceAmbiguous` | yes(FILLED · REFUSED · CANCELLED 가 이 case 로 모호) | yes |
| B22 | 무장 행 하나 → 값 / 둘 → NULL | 같음 | yes | yes |
| B23 | hydrate 실패 | 기존(무변경) | no | no |
| B24 | rows.Err | 기존(무변경) | no | no |
| B25 | 조립 루프 | 모든 시험 | no | no |
| B26 | 모호 표시 | `TestA209TwoArmingRowsStayAmbiguousWithAHookRow` | yes(RED 에서는 a209 세 시험이 이 갈래로 감) | yes |
| B27 | 사유 기본값 | `TestReadOnlyMarksDuplicateExitEvidenceAmbiguous` | no | no |
| B28 | 단일 사건 | `TestA209AFilledExitOrderLinksItsArmingRowDespiteTheFillHook` | yes(RED 에서 미도달) | yes |
