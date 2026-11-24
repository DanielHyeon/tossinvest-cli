# Branch Test Map: `Tracker.Observe` (reconcile) — post-edit(v3)

- Source: `internal/reconcile/mismatch.go` (635-868). 번호는 `../renumber.py` 정렬로 편집 전과 같은 번호(equal 29). 커버: `../../coverage/post-internal-reconcile--tracker.observe.txt`(mode set, body=1 은 실행됨). RED/GREEN 원문 `../../tdd/red.txt` · `../../tdd/green.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 블록 맵 초기화 | 기존(a083 · a110 · mismatch 계열) | 해당 없음(기존 분기) | yes — body=1 |
| B2 | 깨끗한 비교 | `TestA201W2FillAfterTheLocalReadIsCreditedLikeAnAdjustment` | yes — 다음 주기 해제 없음 | yes — body=1, PASS |
| B3 | 차단 비교 | `TestA201TheFillCycleStillCountsAndBreaksTheStreak` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B4 | 깨끗함 ∧ 대기 승격 | `TestA201AnAuthorityErrorOnACleanComparisonRefutesAnAdjustmentCredit` · 기존 a110 | 해당 없음(기존 분기) | yes — body=1 |
| B5 | 권위 확인 오류(깨끗함) → V4b 반박 | `TestA201AnAuthorityErrorOnACleanComparisonRefutesAnAdjustmentCredit` · `TestA201AnAuthorityErrorOnACleanComparisonRefutesAFillCredit` | yes — 조정 판 FAIL(편집 전 credit 이 살아 해제), 체결 판 build failed(새 API) | yes — body=1, PASS |
| B6 | 활성 블록 순회 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B7 | QM ∧ adjusted_reconcile 아님 | `TestA201AConvergedSymbolGetsNoFillCredit` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B8 | 쓸 credit 없음 → 대기 | `TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison` · `TestA201ALaterDisputeRefutesTheFillCredit` | 해당 없음(같은 as-of 대기는 편집 전 동작) | yes — body=1, PASS |
| B9 | a192 종결 증명 후보 | `TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B10 | credit 있음 ∧ 분쟁 → 대기 | `TestA201WithoutTheAuthorityErrorTheSameInputIsUnchanged` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B11 | 대기 승격 ∧ 승격 키 사라짐 | 기존 a110 | 해당 없음(기존 분기) | yes — body=1 |
| B12 | 현재 블록 순회(latch) | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B13 | 없는 블록 latch | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B14 | 권위 확인 오류(차단) | 기존 a110 | 해당 없음(기존 분기) | yes — body=1 |
| B15 | 연속 횟수 진행 | `TestA201TheFillCycleStillCountsAndBreaksTheStreak` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B16 | latch 안 됨 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B17 | 현재 블록 순회 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B18 | 새 블록 추가(그 주기 차단은 선다) | `TestA201W2FillAfterTheLocalReadIsCreditedLikeAnAdjustment`(`assertStuck`) | 해당 없음(편집 전과 같음) | yes — body=1, PASS |
| B19 | 임계 도달 | 기존 a110 · V4b 픽스처 | 해당 없음(기존 분기) | yes — body=1 |
| B20 | 계좌 영구 블록 추가 | 기존 a110 · V4b 픽스처 | 해당 없음(기존 분기) | yes — body=1 |
| B21 | 영속 블록 반영 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B22 | 영구 영속 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B23 | 권위 블록 반영 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B24 | 계좌 권위 | 없음(선택 범위 — 편집 무관) | 해당 없음(기존 분기) | 미실행 — body=0 |
| B25 | 해제 블록 삭제 | `TestA201AnotherSymbolsDisputeKeepsTheFillReleaseWaiting` | yes — A 미해제 | yes — body=1, PASS |
| B26 | 해제 커밋 표시 | 기존 · FILL_APPLIED 해제 | 해당 없음(기존 분기) | yes — body=1 |
| B27 | credit 순회 | `TestA201ALaterDisputeRefutesTheFillCredit` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B28 | 같은 비교 · 앞선 credit 건너뜀 | `TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison` | yes — build failed(새 API) | yes — body=1, PASS |
| B29 | 분쟁 · 커밋 · 블록 없음 → 삭제 | `TestA201ALaterDisputeRefutesTheFillCredit` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
