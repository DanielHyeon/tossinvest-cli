# Function Logic Map: `officialRanking.Read`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L244–299, 분기 4개)
- Risk scan: `risk-pattern-report.md`

공식 순위 한 판을 읽어 `candidate.Reading`으로 옮긴다. 이 change가 **두 사실을 더 싣는다** —
`RankRequested`(요청 행 수)와 `NewlyListed`(3-상태 신규 진입). 둘 다 이 어댑터만 알고
downstream이 복원할 수 없는 값이고, 지금까지는 여기서 버려졌다.

`rememberRead` 호출 위치가 요점이다: **행을 하나도 만들기 전에**, 한 번의 lock 아래에서
직전 집합을 가져오고 교체한다. 같은 시장을 동시에 읽는 두 호출이 서로 교체해 버린 집합에서
각자 답하는 것을 막는다.

`rememberRead`에 넘어가는 것은 **요청 행 수**(`o.count`)이고 비교가 아니다.
2026-07-28 이전에는 여기서 `o.count == total`을 계산해 넘겼고, `total`은
`rememberRead`가 버리는 행(빈 심볼)까지 세는 수였다 — 즉 판정의 분모와 기억의 내용이 서로
다른 집합이었다. 비교는 이제 `current`가 존재하는 자리에서 이루어진다(issues.md I16).
어느 쪽이든 기준은 소스 자신이 선언한 숫자이지 지어낸 하한이 아니다(design D4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | 대문자 시장 코드 | 호출자(스캔) | 그대로 엔드포인트와 기억 키로 간다 |
| `raw.Items` | 0..count 행 | 공식 API | 0행도 정상 — `Collect`가 coverage로 세지 않는다 |
| `total` | `len(raw.Items)` — **도착한** 행 수 | 이 함수 | `RankTotal`이 되고 백분위의 분모다 |
| `o.count` | 요청 행 수(상한 적용 후) | `OfficialRanking` | `RankRequested`가 된다 |
| `previous`/`hadPrevious` | 직전 읽기 집합과 그 자격 | `rememberRead` | `hadPrevious=false`면 모든 행이 `unknown` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `err != nil` — 읽기 실패 | **없음** — 기억을 건드리기 전에 반환한다 | `Reading{}, error` | `TestARateLimitedRankingIsReportedAsOne` |
| B2 | `errors.Is(err, official.ErrRateLimited)` | 없음 | `candidate.ErrRateLimited` 래핑 | `TestARateLimitedRankingIsReportedAsOne` · `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer`(429 3회) |
| B3 | `for _, item := range raw.Items` | `rows` append | — | `TestTheRankTotalIsTheListWeActuallyReceived` 등 전부 |
| B4 | `symbol == ""` — 심볼 없는 행 | 행을 만들지 않는다. `RankTotal`(=`total`)에는 여전히 포함된다 — 엔드포인트가 서브한 행 수가 백분위의 분모다 | continue | `TestAReadingThatLostRowsToBlankSymbolsIsNotAWholeReading` · `TestAReadingOfNothingButBlankSymbolsDoesNotBecomeThePreviousReading` |

B1이 기억보다 앞이라는 것이 F1의 절반이다 — 실패한 읽기는 직전 집합을 **교체하지 않으므로**
장애를 그대로 살아남는다. 그래서 자격 판정이 `usableAt`의 나이 상한으로 옮겨졌다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.reader.Rankings(ctx, o.typ, market, "realtime", false, o.count)` | 순위 한 판 | 429는 `ErrRateLimited`로 사상, 그 밖은 래핑 | ast.json calls |
| `o.rememberRead(market, raw.Items, o.count)` | 직전 집합 취득 + 조건부 교체(온전 여부는 그 안에서 판정) | 오류 없음, 자체 mutex | ast.json calls |
| `newlyListed(previous, hadPrevious, symbol)` | 행별 3-상태 | 오류 없음 | ast.json calls |
| `o.rateBudget()` | rate 예산 | 미보고는 zero value | ast.json calls |
| `decimal` | float → 소수 문자열 | NaN/Inf는 빈 문자열(부재) | ast.json calls |

## State mutations and fallbacks

- `o.seen[market]`을 `rememberRead`가 조건부로 교체한다 — 이 함수 자신은 상태를 갖지 않는다.
- fallback 없음. 실패는 오류이고 부분 결과를 만들지 않는다.
- 주문 경로 무접촉 — `RankingReader`는 `Rankings` 한 메서드만 선언한다.

## Safety conclusion

- Safe edit boundary: 행 리터럴에 필드 2개 가산 + `rememberRead` 호출 1줄. 삭제·재배치 0.
- High-risk impact: no (조회 전용). 재는 성질은 High-risk 인접이다 — `NewlyListed`가 `yes`로 잘못 채워지면 세션 시작마다 패널 전체가 신규 진입으로 기록되고 `seen_late`가 우리 프로세스를 잰다.
