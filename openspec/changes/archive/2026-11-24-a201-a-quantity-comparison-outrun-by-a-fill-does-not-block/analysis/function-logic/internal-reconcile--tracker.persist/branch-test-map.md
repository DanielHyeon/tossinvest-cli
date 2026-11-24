# Branch Test Map: `Tracker.persist` (reconcile) — post-edit(v3)

- Source: `internal/reconcile/mismatch.go` (890-973). 번호는 `../renumber.py` 정렬(B1~B9 같음 · 새 B10 · 옛 B10/B11 → B11/B12). 커버: `../../coverage/post-internal-reconcile--tracker.persist.txt`(mode set). RED/GREEN 원문 `../../tdd/red.txt` · `../../tdd/green.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 장부 없음 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B2 | pending 순회 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B3 | 빈 종목 블록 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B4 | 삽입 오류 | `TestA201AnObserveErrorCycleGrantsNoFillCredit` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B5 | 다른 원인 소유 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
| B6 | Cleared 순회 | `TestA201W2FillAfterTheLocalReadIsCreditedLikeAnAdjustment` | yes — 해제 없음 | yes — body=1, PASS |
| B7 | a192 표시 | `TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B8 | a192 해제 오류 | 기존 a192 | 해당 없음(기존 분기) | yes — body=1 |
| B9 | a192 재판정 해제 | `TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit` | 해당 없음(편집 전에도 통과하는 고정) | yes — body=1, PASS |
| B10 | 체결 credit 해제 → FILL_APPLIED | `TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison` · `TestA201ALaterAdjustmentNamesTheRelease`(거짓 쪽) | yes — build failed(새 API) / 엔진 판은 해제 없음 | yes — body=1, PASS |
| B11 | 해제 쓰기 오류 | 기존(a083 해제 실패 계열) | 해당 없음(기존 분기) | yes — body=1 |
| B12 | 활성 행 없음 | 기존 | 해당 없음(기존 분기) | yes — body=1 |
