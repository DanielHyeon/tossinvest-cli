# Branch Test Map: `Collect`

- Source: `internal/candidate/scan.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

B1·B2·B7·B13·B18·B23·B24는 직접 테스트가 없다. 앞의 넷은 유일한 프로덕션 호출자 `Cycle`의 구성상 도달 불가하거나 방어적 가드이고, 뒤의 셋은 저장소 I/O 실패 경로다. 결함으로 판정하지 않고 미커버로 명시한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 시장 없는 스캔 | 없음 | 아니오(사후 증거) | 미커버 |
| B2 | 순간 없는 스캔 | 없음 | 아니오(사후 증거) | 미커버 |
| B3 | 패널 id 수집 | `TestTwoSourcesCannotShareAnID` | 기록됨(§2-1 실행 재현) | yes |
| B4 | 두 원천이 한 id를 공유 — 응답한 쪽이 못 한 쪽을 보증 | `TestTwoSourcesCannotShareAnID` | 기록됨(§2-1 실행 재현) | yes |
| B5 | 패널을 `heard`로 복사 | `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` | 기록됨(§5 리뷰 P1-1) | yes |
| B6 | 일정이 넘긴 원천이 `heard`에 들어간다 | 같은 테스트 + `TestASourceHeldByTheBackoffIsNotAskedAndDoesNotVouch` | 기록됨(§5 리뷰 P1-1) | yes |
| B7 | 한 원천이 패널과 not-asked 양쪽에 | 없음 — `Cycle`이 panel−due로 만들어 도달 불가 | 아니오(사후 증거) | 미커버 |
| B8 | 패널을 순서대로 읽는다 | `TestOfficialSourcesAloneProduceCandidates` | 아니오(사후 증거) | yes |
| B9 | 원천 실패 — 429와 그 외 | `TestWTSFailingEntirelyStillYieldsCandidates`, `TestARankingFailureIsALossAndNotADegradedFallback` | 기록됨(§1 D14 정정) | yes |
| B10 | 우회가 불가능한 원천이 우회를 주장 | `TestARankingCannotClaimToHaveFallenBack` | 기록됨(§1 3절 — D14 사실 오류) | yes |
| B11 | 행 0개의 200은 커버리지가 아니다 | `TestAnEmptyReadingIsNotEvidenceOfAbsence`, `TestOnlyAReadingWithRowsInItClearsARetreat` | 기록됨(§2-2, §5 리뷰 P2) | yes |
| B12 | 한 원천의 행을 관측으로 | `TestTwoSourcesRaisingOneSymbolMakeOneCandidate` | 아니오(사후 증거) | yes |
| B13 | 빈 심볼 행 | 없음 | 아니오(사후 증거) | 미커버 |
| B14 | 패널 순서상 가격을 처음 나른 원천이 기준선을 가져간다 | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` | 기록됨(issues.md 4) | yes |
| B15 | 순위를 나르지 않는 원천은 순위를 만들어내지 않는다 | `TestAScanDoesNotInventAFirstRankForASourceThatCarriesNone` | 기록됨(§4 P1 — D20) | yes |
| B16 | 패널 전멸 | `TestEverySourceFailingIsAnError` | 아니오(사후 증거) | yes |
| B17 | 전멸 + 오배선이면 오배선을 낸다 | `TestARankingCannotClaimToHaveFallenBack` | 아니오(사후 증거) | yes |
| B18 | 관측 write 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B19 | 승격 대상 심볼 정렬 | `TestOfficialSourcesAloneProduceCandidates` | 아니오(사후 증거) | yes |
| B20 | 정렬된 심볼을 승격하고 출처를 기록 | `TestOneRejectedSymbolDoesNotAbortTheMarket` | 기록됨(§2-3 실행 재현) | yes |
| B21 | 역행 시각으로 한 심볼이 거절 — 시장은 계속 | `TestOneRejectedSymbolDoesNotAbortTheMarket` | 기록됨(§2-3) | yes |
| B22 | 출처 기록 실패 — 시장은 계속 | 같은 테스트 | 아니오(사후 증거) | yes |
| B23 | `Summaries` 읽기 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B24 | 냉각 중 저장소 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B25 | 완전한 결과와 함께 오배선을 알린다 | `TestARankingCannotClaimToHaveFallenBack` | 기록됨(§1 3절) | yes |
