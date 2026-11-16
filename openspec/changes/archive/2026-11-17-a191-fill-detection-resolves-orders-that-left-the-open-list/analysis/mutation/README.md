# a191 변이 원장

- 하네스: `mutate.py`(이 폴더) — `go test -overlay` 로 변이 사본을 끼움. **워크트리 파일은 고치지 않음**(시작·끝 sha256 대조, 각 원문 끝줄).
  사본은 `<scratch>/<pid>/<변이>/` — 판마다 pid, 한 번에 한 판. 무변이 대조군 m0 이 GREEN 이 아니면 멈춤.
- 명령: `GOCACHE=~/.cache/a191-gocache python3 mutate.py <repo> <scratch> [변이…]`(go test `-run 'TestA191|TestJournalTrackedCarriesTheIntentID' -timeout 45s`).
- 원문: `mutation-run-1.txt`(m0~m26), `mutation-run-2-buildfix.txt`(m9·m20·m23 — 1판은 사본의 미사용 import/변수로 빌드 실패, 하네스 수정 뒤 재실행),
  `mutation-run-3-wiring.txt`(m27~m29 — 구현 뒤 작성한 시험의 대체 증거),
  **`mutation-run-4-after-code-review-1.txt`(현재 코드 기준 전수 재실행: m0 GREEN, m1~m26·m30~m32 — m4 만 SURVIVED(동등), 나머지 CAUGHT)**.

| 변이 | 무엇 | 판정 | 잡은 시험(대표) |
|---|---|---|---|
| m1 | 후보 등록 삭제 | CAUGHT | 핵심 RED 3 포함 21개 |
| m2 | 단건 종결도 후보 | CAUGHT | `TestA191TerminalSingleReadIsNeverLookedUp` |
| m3 | OPEN 에 있는 주문도 후보 | CAUGHT | `TestA191OrderStillInTheOpenListIsNeverLookedUp` |
| m4 | 행 파생의 `!FailClosed` 삭제 | **SURVIVED — 동등 변이** | — (아래) |
| m5 | 행 파생의 `Terminal` 삭제 | CAUGHT | `TestA191PartialFilledInTheClosedGroupIsNotTerminal` |
| m6 | 체결 수량 일치 삭제 | CAUGHT | `TestA191ClosedListFillDisagreementNeverRecordsAFill` |
| m7 | 주문 수량 일치 삭제 | CAUGHT | 같음 |
| m8 | 범위 일치 삭제 | CAUGHT | `TestA191ClosedRowFromAnotherScopeIsNotAccepted` |
| m9 | 중복 중 첫 행 채택 | CAUGHT(2판) | `…DuplicateClosedRows…`, `…SameScopeRows…` |
| m10 | 조회 실패로 주기 실패 | CAUGHT | `TestA191ClosedListReadFailureDoesNotFailTheCycle` |
| m11 | 간격 제거 | CAUGHT | `TestA191ClosedLookupIsThrottledPerOrder` |
| m12 | 창 from = 거래일 | CAUGHT | `TestA191ClosedQueryNamesTheSymbolAndTheTradingDayWindow` |
| m13 | 종목 필터 누락 | CAUGHT | 같음 |
| m14 | 즉시 보고 | CAUGHT | `TestA191UnresolvedIsReportedOnlyAfterItPersists` |
| m15 | id 비교 완화 | CAUGHT | `TestA191ClosedRowIDMustMatchByteForByte` |
| m16 | 수량 불일치 시 행의 체결 수량 채택 | CAUGHT | `TestA191ClosedListFillDisagreementNeverRecordsAFill` |
| m17 | 조회를 기존 범위 가드 앞으로 | CAUGHT | `TestA191ExistingTrackedReadGuardsRunBeforeTheLookup` 외 |
| m18 | 정리 삭제 | CAUGHT | `TestA191JustPlacedOrderMissingFromTheOpenListIsNotEscalated` |
| m19 | 행의 실행 필드 사용 | CAUGHT | `TestA191AcceptedClosedRowKeepsTheSingleReadExecution` |
| m20 | 중복을 범위로만 판정 | CAUGHT(2판) | `TestA191SameScopeRowsWithDifferentPayloadsAreAmbiguous` |
| m21 | UNKNOWN 단건도 후보(M2 위반) | CAUGHT | `TestA191UnknownSingleReadIsNeverLookedUp` |
| m22 | execution 존재 요구 삭제(M6) | CAUGHT | `TestA191ClosedRowWithoutExecutionIsNotAccepted` |
| m23 | 주기당 1건 상한 삭제 | CAUGHT(2판) | `TestA191OnlyOneClosedLookupPerCycle` |
| m24 | 조회 시한 삭제 | CAUGHT(시험 시한 45s 초과) | `TestA191SlowClosedLookupIsCutAtTheDeadline` |
| m25 | 경보에 브로커 주문번호 | CAUGHT | `TestA191UnresolvedAlertCarriesOnlyTheInternalIdentifier` |
| m26 | 미뤄진 후보의 lastLookup 갱신 | CAUGHT | `TestA191OnlyOneClosedLookupPerCycle` |
| m27 | IntentID 옮김 줄 삭제 | CAUGHT | `TestJournalTrackedCarriesTheIntentID` |
| m28 | 경보 배선 줄 삭제 | CAUGHT | `TestA191TheDetectorIsWiredToTheEngineNotifier` |
| m29 | nil 알림기 거름 삭제 | CAUGHT | `TestA191FillAlertsFilterTheNilNotifier` |
| m30 | 실패 주기 후보 버림 삭제(코드 리뷰 1 #1) | CAUGHT | `TestA191StaleCandidatesDoNotAccumulate` |
| m31 | 수용 즉시 간격 상태 삭제(코드 리뷰 1 #2) | CAUGHT | `TestA191AcceptedLookupKeepsTheIntervalWhenTheCycleFailsLater` |
| m32 | 파싱 실패의 M6 사유 가름 삭제(코드 리뷰 1 #3) | CAUGHT | `TestA191ClosedRowWithoutExecutionIsNotAccepted` |
| m33 | collect B12 `> 0` → `>= 0`(기존 갈래, `mutation-run-5-b12.txt`) | CAUGHT | `TestA191IncompleteOpenOrderNotTrackedIsStillApplied` |

**m4 가 동등인 이유**: `brokerstate.Derive` 의 모든 return 은 `d.settle`·`d.unknown`·`d.pendingMutation`·`d.deriveGroupLabel` 을 거치고
(cited AST `internal-brokerstate--derive` calls), 앞 둘만 `Terminal`/`FailClosed` 를 정한다 — `settle` 은 `FailClosed=false`(derive.go:580),
`unknown` 은 `Terminal=false`(derive.go:590). 그래서 `FailClosed ⇒ !Terminal` 이고 `!FailClosed` 검사는 지금 판정을 바꾸지 않는 방어용 중복이다.
`Derived` 가 바뀌어 그 함의가 깨지면 이 검사가 살아남는다. **Manager 결정(2026-11-17)**: 동등 변이로 기록, 새 시험 없음, 방어 검사는 지우지 않음. 근거는 derive.go 의 두 생성자 `settle`(577~586)·`unknown`(588~595).
