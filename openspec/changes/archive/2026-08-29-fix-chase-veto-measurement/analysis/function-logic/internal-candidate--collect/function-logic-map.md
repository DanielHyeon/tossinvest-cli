# Function Logic Map: `Collect`

- Source: `internal/candidate/scan.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L226–497, 분기 27개)
- Risk scan: `risk-pattern-report.md`

한 시장의 패널을 읽어 관측을 기록하고 승격·냉각까지 한 pass에 처리한다. 이 change가 더한
것은 **두 곳**이다.

1. `result.Readings[id]` (B11) — 소스별 요청 행 수와 도착 행 수. 공식 순위 엔드포인트가
   요청한 100행을 실제로 주는지는 **측정된 적이 없고**, 주지 않는다면 후속 change가 임계를
   고를 분포가 존재하지 않는다. 이 한 줄이 스캔 한 번으로 답한다. (증상 문장 2026-07-28
   정정: 계속 짧은 엔드포인트가 만드는 것은 `READING_TRUNCATED`가 아니라 `NO_FIRST_RANK`
   → `NO_FIRST_SIGHTING`이고, 스캔 쪽에서는 내려오지 않는 `FirstRanksHeld`다 — 짧은 읽기는
   기억을 교체하지 못해 자격 자체가 서지 않기 때문이다. issues.md I9.)
2. 행 → 관측 복사에 `RankRequested`가 붙었다(B13 본문).

`Readings`를 행에서 읽는 비용은 숨기지 않고 적혀 있다: `Reading`에 그 수가 없으므로
**행이 0개인 읽기는 여기 나타날 수 없다** — 가장 극단적인 절단이 가장 설명되지 않는다.
빈 읽기는 이미 자기 사유와 함께 `Missing`에 들어간다(B12).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Market` | 비어 있지 않아야 | 호출자 | 빈 값은 오류 |
| `opts.At` | non-zero | `Cycle` | zero는 오류 |
| `opts.Sources` | id가 서로 달라야 | `Panel` | 충돌은 읽기 전에 오류 |
| `opts.NotAsked` | `Sources`와 서로소 | 스케줄 | 교집합은 오류 |
| `reading.Rows[0].RankRequested` | 한 읽기의 모든 행이 같은 값 | 어댑터가 자기 필드에서 설정 | 0이면 `Readings`에 항목 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `market == ""` | 없음 | 오류 | **커버 없음** |
| B2 | `opts.At.IsZero()` | 없음 | 오류 | **커버 없음** |
| B3 | `for _, src := range opts.Sources`(중복 id 검사) | `seen` 구성 | — | `TestTwoSourcesCannotShareAnID` |
| B4 | `seen[id]` — 중복 | **아무것도 읽지 않는다** | 오류 | `TestTwoSourcesCannotShareAnID` |
| B5 | `for id := range seen` | `heard` 구성 | — | `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` |
| B6 | `for _, id := range opts.NotAsked` | `heard` 구성 | — | `TestASourceTheSchedulePassedOverIsNotASourceThatIsGone` |
| B7 | `seen[id]` — 패널과 not-asked 양쪽 | 없음 | 오류 | **커버 없음** |
| B8 | `for _, src := range opts.Sources`(본 루프) | 관측 축적 | — | `scan_test.go` 전반 |
| B9 | `err != nil` — 읽기 실패 | `Missing`에 사유와 429 여부 | continue | `TestARankingFailureIsALossAndNotADegradedFallback` · `TestEverySourceFailingIsAnError` |
| B10 | `reading.ViaFallback && !id.hasFallback()` | `Missing` + `misWired` | continue, 끝에 오류 | `TestARankingCannotClaimToHaveFallenBack` |
| B11 | `len(rows) > 0 && rows[0].RankRequested > 0` **(신규)** | `result.Readings[id]` 설정 | — | `TestTheScanReportSaysWhatEachSourceAskedForAndWhatArrived`(리포트 측) · `TestATruncatedReadingReachesTheVerdictAsTruncated`(배선 측) |
| B12 | `len(reading.Rows) == 0` | `Missing`에 '커버리지 아님' | continue — **냉각 근거가 되지 않는다** | `TestAnEmptyReadingIsNotEvidenceOfAbsence` |
| B13 | `for _, r := range reading.Rows` | 관측·`covered`·`raisedBy` 구성 | — | `scan_test.go` 전반 |
| B14 | `symbol == ""` | 행을 버린다 | continue | **커버 없음** |
| B15 | `!taken && Price != ""` | `firstPriced[symbol]` 고정 | — | `TestTwoSourcesRaisingOneSymbolMakeOneCandidate` |
| B16 | `Rank > 0 && RankTotal > 0` | 없음(B17의 문지기) | — | `TestAQualifiedReadingTakesTheFirstSightingFromAPanelEarlierUnqualifiedOne` |
| B17 | `!taken \|\| (!qualifiesFirstRank(held) && qualifiesFirstRank(o))` **(이번 수리)** | `firstRanked[symbol]` 설정 — 자격 있는 읽기가 자격 없는 것을 **밀어낸다**, 같은 자격끼리는 패널 순서 | — | `TestAQualifiedReadingTakesTheFirstSightingFromAPanelEarlierUnqualifiedOne` · `TestPanelOrderStillDecidesBetweenTwoReadingsOfTheSameStanding` · `TestAQualifiedReadingIsAlsoPreferredWhenItArrivesForOnlySomeSymbols` |
| B18 | `!anyRead` | 없음 | 오류 | `TestEverySourceFailingIsAnError` · `TestAPanelWithNoSourcesInItIsStillAnError` |
| B19 | `len(misWired) > 0`(전 소스 실패 시) | 없음 | mis-wire 오류가 이긴다 | `TestARankingCannotClaimToHaveFallenBack` |
| B20 | `RecordObservations` 오류 | 없음 | 오류 — 부분 결과와 함께 | **커버 없음** |
| B21 | `for symbol := range raisedBy` | 정렬용 슬라이스 | — | `scan_test.go` 전반 |
| B22 | `for _, symbol := range symbols` | 승격 pass | — | `scan_test.go` 전반 |
| B23 | `Promote` 오류 | `Rejected`에 사유 | continue — 한 심볼의 거부가 시장 전체를 멈추지 않는다 | `TestOneRejectedSymbolDoesNotAbortTheMarket` |
| B24 | `NoteSources` 오류 | `Rejected` | continue | **커버 없음** |
| B25 | `recordFirsts` 오류 | 없음 | 오류 | **커버 없음** |
| B26 | `coolAbsent` 오류 | `result.Cooled`는 이미 설정 | 오류 | **커버 없음** |
| B27 | `len(misWired) > 0`(정상 종료 시) | 없음 | 완전한 결과 + 오류 | **커버 없음** — mis-wire와 성공 읽기를 함께 넣는 fixture가 없다 |

**B17은 2026-07-28 pre-gate 리뷰 P0/P1-3의 수리다** (issues.md I17). 이전 형태
(`!taken && Rank > 0 && RankTotal > 0`)는 패널 순서에서 처음으로 위치를 실은 소스를
무조건 채택했고, `recordFirsts`가 그것의 자격을 따로 판정해 **보류**했다. 두 규칙이 따로
결정하므로 같은 tick 안의 자격 있는 읽기가 버려졌다 — 그리고 계속 짧게 오는 소스는
자기 기억을 교체하지 않아 `NewlyListed`가 영구히 `unknown`이므로, 패널 첫 번째 소스가
그렇게 되면 시장 전체의 `first_rank`가 무기한 막힌다.

이제 자격 있는 읽기가 자격 없는 것을 밀어낸다. 자격 판정은 `recordFirsts`와 **같은 술어**
(`qualifiesFirstRank`)이므로 두 자리가 어긋날 수 없다. 패널의 어떤 소스도 자격이 없으면
여전히 보류되고(회복 가능, 다음 tick ±TTL 창 안), 같은 자격의 읽기 사이에서는 패널 순서가
그대로 결정한다 — "가장 높은 순위가 이긴다"가 아니다. 그것은 스캔이 여러 참인 답 중 가장
놀라운 것을 고르는 형태다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `src.Read(ctx, market)` | 소스 읽기 | 오류는 `Missing`, 429는 `isRateLimited`로 표시 | ast.json calls |
| `store.RecordObservations` | 관측 append | 검증 실패는 전체 거부 | ast.json calls |
| `store.Promote` / `store.NoteSources` | 승격과 provenance | 실패는 `Rejected` | ast.json calls |
| `recordFirsts(...)` | 두 stored first | 오류는 pass 중단 | ast.json calls |
| `coolAbsent(...)` | 냉각 | 오류는 결과와 함께 반환 | ast.json calls |

## State mutations and fallbacks

- `result.Readings[id]`(신규): 소스별 요청·도착. 판정하지 않고 **기록만 한다**.
- `observations` 테이블 append, `candidates` 승격·provenance·냉각.
- fallback 없음. 실패한 소스는 결과에서 빠지고 사유가 남는다.
- 주문 경로 무접촉 — `Source`는 `Read` 한 메서드다.

## Safety conclusion

- Safe edit boundary: `Readings` 맵 초기화 1줄 + 기록 블록 6줄 + 행 복사에 필드 1개 + `firstRanked` 선택 3줄(B16·B17). 삭제·재배치 0.
- High-risk impact: **yes 인접** — 이 함수가 냉각을 결정하고 냉각의 끝은 만료이며 만료는 first_seen_at을 버린다. 이 change의 편집은 그 판단에 닿지 않는다: `Readings` 기록은 `covered`·`responded`·`raisedBy` 어느 것도 건드리지 않고, 빈 읽기 처리(B12)보다 앞이지만 `continue`를 바꾸지 않는다.
