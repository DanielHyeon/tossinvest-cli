# Branch Test Map: `ApplyPositionCampaignFill`

2026-09-07 (task 6.4): 이 함수를 편집해 분기가 54 → 49 개가 됐다. 아래 표는 옛 표를
손으로 옮긴 것이 아니라 옛/새 `ast.json` 의 분기 열을 difflib 으로 **정렬해서** 재번호했다.
측정 결과: B1–B31 동일, 옛 B32·B33·B34(`requestedCmp >= 0` / `else` / `fill.Terminal`)가
새 B32 하나로, 옛 B36–B39(`!hasSuccessor` / `requestedCmp > 0` / `reconcile` / terminal 보존)가
새 B34 하나로 접혔다 — 그 넷이 하던 판정이 `positioncampaign.LegStateAfterFill` 로 옮겨갔다.
옛 B35→B33, B40→B35, B41→B36, B42–B54→B37–B49 는 위치만 이동했다.
새로 쓴 두 행(B32·B34)의 RED 는 수정을 되돌리는 뮤테이션으로 확인했다(`issues.md` §2).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | duplicate/lower/restart | `TestCampaignFillRollbackAndRestartAreDeterministic` | yes | yes |
| B2 | predecessor late positive | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` | yes | yes |
| B3 | cap excess | `TestCampaignAggregateFillExcessLatchesReconcileWithoutTruncation` | yes | yes |
| B4 | legacy multiple matches preserves fill | `TestAmbiguousCampaignFillPreservesAuthoritativeFillTransaction` | yes | yes |
| B5 | CLOSED late fill | `TestClosedCampaignLateFillStaysClosedAndLatchesReconcile` | yes | yes |
| B6 | zero/partial terminal | `TestCampaignZeroAndPartialUnchangedTerminalObservationsCancelResidualIdempotently` | yes | yes |
| B7 | zero-fill all terminal closes/releases | `TestZeroFillTerminalClosesCampaignAndReleasesClaim` | yes | yes |
| B8 | live tx check | fill tests | yes | yes |
| B9 | authoritative query | scope tests | yes | yes |
| B10 | row scan | fill tests | yes | yes |
| B11 | row iteration error | query failure tests | yes | yes |
| B12 | no match | unrelated fill tests | yes | yes |
| B13 | multiple match | ambiguous-fill test | yes | yes |
| B14 | cumulative validation | fill validation tests | yes | yes |
| B15 | watermark compare | duplicate/lower tests | yes | yes |
| B16 | lower no-op | duplicate/lower tests | yes | yes |
| B17 | terminal change | terminal tests | yes | yes |
| B18 | duplicate no-op | restart tests | yes | yes |
| B19 | positive delta | partial tests | yes | yes |
| B20 | CLOSED/terminal reconcile | late-fill tests | yes | yes |
| B21 | AppliedFill delta mismatch | ambiguity tests | yes | yes |
| B22 | cap compare | cap test | yes | yes |
| B23 | first generation bind | first-fill tests | yes | yes |
| B24 | generation query | first-fill tests | yes | yes |
| B25 | expected successor | mismatch tests | yes | yes |
| B26 | claim bind | first-fill tests | yes | yes |
| B27 | generation mismatch | mismatch tests | yes | yes |
| B28 | existing generation check | restart tests | yes | yes |
| B29 | bound generation mismatch | reconcile tests | yes | yes |
| B30 | watermark update | fill tests | yes | yes |
| B31 | leg aggregate query | fill tests | yes | yes |
| B32 | `requestedCmp > 0` — leg 요청을 넘긴 합계는 RECONCILE 을 건다 (1115) | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` | yes | yes |
| B33 | `hasCampaignSuccessor` 조회 오류 (1119) | replacement tests | yes | yes |
| B34 | `legErr != nil` — 표가 모르는 사실이면 이전 상태 유지 + RECONCILE (1129) | `TestCampaignFillHasExactlyOneLegStateJudgement` · `TestResidualCancelInOneObservationAgreesWithReconstruction` | yes | yes |
| B35 | `campaign_legs` 수량/상태 갱신 실패 (1135) | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` + restart test | yes | yes |
| B36 | leg 잔여를 넘긴 successor remaining 재계산 실패 (1140) | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` | yes | yes |
| B37 | reconcile latch | ambiguity tests | yes | yes |
| B38 | PLANNED activation | first-fill tests | yes | yes |
| B39 | campaign CAS update | race tests | yes | yes |
| B40 | all-terminal query | zero-fill close test | yes | yes |
| B41 | campaign close | zero-fill close test | yes | yes |
| B42 | claim release | zero-fill close test | yes | yes |
| B43 | event append | replay tests | yes | yes |
| B44 | digest append | drift tests | yes | yes |
| B45 | ambiguous retry | ambiguous-fill retry test | yes | yes |
| B46 | durable account reconcile | CLOSED/ambiguous tests | yes | yes |
| B47 | authoritative fill preserved | ambiguity test | yes | yes |
| B48 | final nil | all fill tests | yes | yes |
| B49 | current per-order remaining calculation failure | fail-closed tx rollback contract + watermark/restart suite | yes | yes |
