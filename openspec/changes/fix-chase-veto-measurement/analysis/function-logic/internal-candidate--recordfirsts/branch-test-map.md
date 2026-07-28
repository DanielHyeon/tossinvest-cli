# Branch Test Map: `recordFirsts`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 승격이 없으면 Summaries도 읽지 않는다 | `TestAScanDoesNotCoolASymbolItDidNotLookFor` | — (기존 동작) | yes |
| B2 | Summaries 실패 | **커버 없음** (I/O) | no | no |
| B3 | needPrice/needRank 구성 | `scan_test.go` 전반 | — (기존 동작) | yes |
| B4 | 다른 시장의 summary는 무시 | **커버 없음** | no | no |
| B5 | 승격된 심볼마다 | `scan_test.go` 전반 | — (기존 동작) | yes |
| B6 | 첫 가격 쓰기 | `TestOfficialSourcesAloneProduceCandidates` | — (기존 동작) | yes |
| B7 | 첫 가격 쓰기 실패 | **커버 없음** | no | no |
| B8 | 첫 가격 카운트 | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` | — (기존 동작) | yes |
| B9 | 이미 순위가 있으면 제안조차 하지 않는다 | `TestAPositionStoredBeforeTheFactsExistedIsNotFilledInByALaterScan`(5회 스캔 후에도 불변) | yes | yes |
| B10 | 자격 없는 읽기의 위치는 저장하지 않고 센다 | `TestASessionStartDoesNotStampThePanelAsSeenLate`(첫 tick `NO_FIRST_RANK`, 저장 없음) · `TestTheHeldCountIsRenderedAndSaysWhichCommandCanNeverQualifyAPosition`(카운트 렌더) | yes (보류 전에는 `NEW_ENTRANT_UNKNOWN`이 영구였다 — 그 테스트가 뒤집혔다) | yes |
| B11 | 쓰기 결과 분류 | `TestARePromotionAfterExpiryIsQualifiedByTheReadingThatSawTheSymbolReturn` | yes | yes |
| B12 | 쓰기 실패는 `Rejected` | **커버 없음** | no | no |
| B13 | 저장 성공 카운트 | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne` | — (기존 동작) | yes |
| (B11 세 번째 경우) | 동일성 창 밖 읽기는 조용히 지나간다 | `TestARankFromOutsideTheIdentityWindowIsNotStored`(store 층) | — (기존 동작) | yes |

**정직한 커버리지 기록**: 기존 분기 넷(B4·B7·B12, 그리고 B2)이 커버되지 않는다 —
두 시장 fixture와 두 stored-first 쓰기의 실패 경로가 없다.

그리고 B10의 **절반**만 각각 독립으로 측정된다. `!NewlyListed.Known()`은
세션 시작 테스트가 잡고, `RankRequested <= 0`만 단독으로 참인 fixture — 즉 신규 진입은
측정됐는데 요청 수가 없는 읽기 — 를 `Cycle` 층에서 만드는 테스트는 **없다**. 그 쌍은
`MeasureFirstSighting` 층에서
`TestTheOneRowReadingWithNoRecordedRequestIsTheCaseThatMattered`가 잡는다.
