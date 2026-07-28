# Function Logic Map: `Collect`

- Source: `internal/candidate/scan.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

본문이 바뀐 기존 함수다. 이 change가 넣은 것은 셋이다 — (1) `opts.NotAsked`에서 `heard` 집합을 만들고 냉각 판정을 `seen`이 아니라 `heard`로 하기(§5 리뷰 P1-1), (2) `firstPriced`/`firstRanked` 패널 순서 선점과 `recordFirsts` 호출(D17과 그 반쪽), (3) `Vouched` 목록 — 행을 나른 응답만 커버리지로 세기(§5 리뷰 P2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Market` | 비어 있지 않은 시장 코드 | 호출자(`Cycle`/CLI) | 공백이면 즉시 에러, 아무것도 읽지 않는다 |
| `opts.At` | 0이 아닌 순간 | 호출자 — **스캔은 시계를 직접 부르지 않는다**(D5) | 0이면 에러. 시각은 주입된 clock 또는 호출자 인자에서만 온다. `TestNothingInThisPackageAsksTheWallClockWhatTimeItIs`가 `time` import를 **경로로** 해석해 `Now/Since/Until/After/Tick/NewTimer/NewTicker/AfterFunc/Sleep` 9개 이름과 alias import·dot import 두 형태, 합쳐 11가지 철자를 막는다. |
| `opts.Sources` | id가 서로 다른 패널 | 호출자 | id 충돌은 **읽기 전에** 거부한다. 응답한 원천이 응답 못 한 원천을 대신 보증하는 것이 §2-1이 실행으로 재현한 결함이다 |
| `opts.NotAsked` | 패널에 속하지만 이번 pass에서 일정상 읽지 않은 원천 | `Cycle`이 panel−due로 계산 | 패널과 겹치면 에러. 이것은 원천이 아니므로 `Attempted`·`Missing`에 들어가지 않는다 |
| `Row.Price`/`TradingValue` 등 십진값 | 문자열, 빈 문자열은 **원천이 나르지 않음** | 원천 어댑터 | 0이 아니다. `firstPriced`는 `TrimSpace(r.Price) != ""`로만 선점한다 |
| `r.Rank`/`r.RankTotal` | 둘 다 양수여야 순위 | 원천 | 하나라도 0이면 `firstRanked` 선점 대상이 아니다. rank/0은 +Inf이고 모든 임계를 통과한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `market == ""` | 없음 | `ScanResult{}, error` | 직접 테스트 없음(호출자 `Cycle`이 시장을 항상 채운다) |
| B2 | `opts.At.IsZero()` | 없음 | `ScanResult{}, error` | 직접 테스트 없음 |
| B3 | 패널 순회로 `seen` 채우기 | 지역 map | — | `TestTwoSourcesCannotShareAnID` |
| B4 | `seen[id]` 중복 id | 없음 | `ScanResult{}, error` — **읽기 전에** | `TestTwoSourcesCannotShareAnID` |
| B5 | `seen`을 `heard`로 복사 | 지역 map | — | `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` |
| B6 | `opts.NotAsked` 순회 | `heard[id]=true` | — | 같은 테스트(`Cycle` 경유) |
| B7 | `seen[id]` — 패널과 not-asked에 동시에 있음 | 없음 | `ScanResult{}, error` | 직접 테스트 없음. 유일한 프로덕션 호출자 `Cycle`이 panel−due로 만들어 구성상 도달 불가 |
| B8 | 패널 원천 순회 | 원천 Read 1회씩 | — | `TestOfficialSourcesAloneProduceCandidates` |
| B9 | `src.Read` 에러 | `result.Missing` 추가, 429면 `RateLimited` | 계속 | `TestWTSFailingEntirelyStillYieldsCandidates`, `TestARankingFailureIsALossAndNotADegradedFallback` |
| B10 | `reading.ViaFallback && !id.hasFallback()` | `Missing` + `misWired` | 계속(마지막에 에러 동반 반환) | `TestARankingCannotClaimToHaveFallenBack` |
| B11 | `len(reading.Rows) == 0` | `Missing`에 사유, `responded`·`Vouched` **미등록** | 계속 | `TestAnEmptyReadingIsNotEvidenceOfAbsence`, `TestOnlyAReadingWithRowsInItClearsARetreat` |
| B12 | 행 순회 | `observations`·`covered`·`raisedBy` 누적 | — | `TestTwoSourcesRaisingOneSymbolMakeOneCandidate` |
| B13 | `symbol == ""` | 없음 | 그 행만 건너뜀 | 직접 테스트 없음 |
| B14 | 아직 선점 안 됨 && 가격이 공백 아님 | `firstPriced[symbol] = o` | — | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` |
| B15 | 아직 선점 안 됨 && `Rank>0 && RankTotal>0` | `firstRanked[symbol] = o` | — | `TestAScanDoesNotInventAFirstRankForASourceThatCarriesNone` |
| B16 | `!anyRead` — 아무도 응답 안 함 | 없음 | `result, ErrNoSourceAnswered` | `TestEverySourceFailingIsAnError` |
| B17 | 그 안에서 `len(misWired) > 0` | 없음 | `result, ErrFallbackNotPossible` — 증상보다 원인을 낸다 | `TestARankingCannotClaimToHaveFallenBack` |
| B18 | `store.RecordObservations` 에러 | 없음(트랜잭션 롤백) | `result, err` | 직접 테스트 없음 |
| B19 | `raisedBy` 키 수집 | 지역 slice | — | `TestOfficialSourcesAloneProduceCandidates` |
| B20 | 정렬된 심볼 순회 | `Promote` + `NoteSources` **저장소 write** | — | `TestOneRejectedSymbolDoesNotAbortTheMarket` |
| B21 | `store.Promote` 에러 | `result.Rejected` 추가 | `continue` — 시장 전체를 멈추지 않는다 | `TestOneRejectedSymbolDoesNotAbortTheMarket` |
| B22 | `store.NoteSources` 에러 | `result.Rejected` 추가 | `continue` | 같은 테스트 |
| B23 | `recordFirsts` 에러 | `result` 유지 | `result, err` | 직접 테스트 없음 — `Summaries` 실패 경로 |
| B24 | `coolAbsent` 에러 | `result.Cooled`는 이미 설정 | `result, err` | 직접 테스트 없음 |
| B25 | `len(misWired) > 0` (정상 종료 직전) | 없음 | 완전한 `result` + `ErrFallbackNotPossible` | `TestARankingCannotClaimToHaveFallenBack` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `src.Read` | 원천 1회 조회 | 에러는 `Missing`으로 흡수, 재시도 없음. 429는 `isRateLimited`로 분류 | `Source` 인터페이스는 읽기 메서드 하나뿐이라 주문을 표현할 수 없다 |
| `store.RecordObservations` | 이번 pass의 원 관측 append | 실패는 pass 중단 | ast.json calls |
| `store.Promote` / `store.NoteSources` | 후보 수명·출처 갱신 | 심볼 단위 실패는 `Rejected`로 모으고 계속 | `TestOneRejectedSymbolDoesNotAbortTheMarket` |
| `recordFirsts` | D17과 그 반쪽의 1회성 write | 심볼 실패는 `Rejected`, `Summaries` 실패는 pass 중단 | 같은 파일 |
| `coolAbsent` | 증거가 있는 후보만 냉각 | 실패 시 지금까지의 `Cooled`와 함께 반환 | 같은 파일 |
| `isRateLimited` | 429 분류 | 앵커드 매칭. 맨 `429` 부분문자열 매칭은 §2-7이 종목코드·trace id로 재현해 제거했다 | `TestARequestIDContainingTheDigits429IsNotARateLimit` |
| (주문 경로) | — | — | 없음. 이 패키지의 모듈 내부 의존 폐포는 `{internal/clock}` 하나다(`go list -deps ./internal/candidate`, `TestDiscoveryDependsOnNothingItHasNotArguedFor`). 주문 경로에 닿는 간선이 없고, 이 함수도 예외가 아니다. |

## State mutations and fallbacks

- 저장소 write 4종: `RecordObservations`(observations append), `Promote`(candidates upsert), `NoteSources`(sources 병합 + 완전성 갱신), `recordFirsts` 경유 `NoteFirstPrice`/`NoteFirstRank`(각 1회성), `coolAbsent` 경유 `Cool`(cooled_at).
- 부분 실패가 남기는 것: 심볼 하나가 `Promote`에서 거절되면 그 심볼만 `promoted`에 없고 이후 first-write·냉각 대상에서 빠지며 `Rejected`에 이름이 남는다. 나머지 시장은 끝까지 간다 — §2-3이 실행으로 재현한, 정렬된 루프 하나가 watchlist 전체를 만료시키던 결함의 수리다.
- `RecordObservations` 실패는 관측을 **하나도** 남기지 않는다(단일 트랜잭션). 승격·냉각도 실행되지 않는다.
- `heard`는 순수 지역 계산이고 영속되지 않는다. 냉각 판정에만 쓰인다.
- `Vouched`는 행을 나른 응답만 담는다. 빈 200은 `Responded`에는 세어지지만 `Vouched`에는 없고, 백오프 사다리의 회복 신호는 후자다(§5 리뷰 P2).

## Safety conclusion

- Safe edit boundary: `heard`를 다시 `seen`으로 되돌리기, 빈 200을 `responded`에 넣기, 심볼 실패에서 루프를 중단하기, 패널 중복 id 검사를 읽기 뒤로 미루기는 모두 이미 실행으로 재현된 파괴 경로의 복원이라 금지
- High-risk impact: no — 읽기 전용 발굴 경로다. 주문·손절·사이징·원장 어디에도 닿지 않고 의존 폐포는 `{internal/clock}`이다. 다만 이 함수의 write는 **원장과 같은 파일시스템**의 `candidates.db`에 들어가고(D2·D16), 잘못된 냉각은 `first_seen_at`을 지워 D3의 회계를 파괴한다. 안전 불변식에는 닿지 않지만 이 change에서 가장 파괴적인 함수다.
