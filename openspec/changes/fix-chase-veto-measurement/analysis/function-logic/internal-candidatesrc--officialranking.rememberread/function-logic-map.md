# Function Logic Map: `officialRanking.rememberRead`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L308–332, 분기 5개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. 이번 읽기의 심볼 집합을 `market` 키에 swap해 넣고, 교체한 집합과 **그것을
써도 되는지**를 함께 돌려준다.

두 반환이 3-상태의 출처다. `usable == false`는 "빈 직전 읽기"가 **아니다** — 빈 집합은
"여기 있는 모든 심볼이 새로 들어왔다"는 증거이고, 자격 있는 직전 읽기가 없는 것은 아무
증거도 아니다.

`whole == false`면 **교체하지 않는다**(리뷰 F2). 100행을 요청해 3행이 온 읽기는 3행짜리
목록이 아니라 100행 목록을 3행만 본 것이다. 그것을 기억으로 삼으면 다음 온전한 읽기가
97개 심볼을 "직전 읽기에 없었다"로 보고하고, 그 97개는 **측정된 사실**로 저장되어
`candidates.first_rank_newly_listed`와 화면의 `신규 진입`까지 간다. 빈 읽기도 같은 비교로
짧으므로 같은 규칙을 받는다 — `TestAnEmptyReadingIsStillAReadingOfThisList`가 이 change에서
뒤집힌 이유다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | 임의 문자열 | 호출자 | `ToUpper(TrimSpace(...))`로 정규화해 키로 쓴다 |
| `items` | 이번 읽기의 행 | API 응답 | 빈 심볼은 집합에서 제외 |
| `whole` | `o.count == len(items)` | `Read` | false면 기억을 남긴다 |
| `o.seen` | 시장별 `previousReading` | 이 함수 | nil이면 여기서 생성 |
| `o.now` | 주입 clock | `OfficialRanking` | 기억의 날짜와 나이 비교의 기준 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, item := range items` | `current` 맵 구성 | — | `TestTheSecondReadingSeparatesTheSymbolsThatJoinedFromTheOnesThatStayed` |
| B2 | `symbol != ""` — 빈 심볼 제외 | 제외 시 집합에 넣지 않는다 | — | **없음** (아래 주석) |
| B3 | `o.seen == nil` | `o.seen = map[...]{}` | — | 모든 첫 읽기 테스트 |
| B4 | `whole` | `o.seen[key]`를 이번 읽기로 교체 | — | `TestAShortReadingDoesNotBecomeTheYardstickForTheNextWholeOne` · `TestAnEmptyReadingDoesNotBecomeThePreviousReading` |
| B5 | `!previous.usableAt(at)` | 없음 | `nil, false` — 미상 | `TestTheMemoryOfAReadingBeforeAnOutageIsNotAnAnswer` · `TestTheMemoryExpiresAtTheStalenessTTLAndNotBefore` |

`previous`는 값 복사이므로 B4의 교체가 B5의 판정을 오염시키지 않는다. 순서가 뒤집히면
"방금 넣은 것을 직전 읽기로 읽는" 상태가 되고, 그러면 모든 심볼이 영구히 `no`가 된다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.now.Now()` | 이번 읽기의 instant | 주입 clock, 오류 없음 | ast.json calls |
| `o.mu.Lock`/`defer Unlock` | 취득과 교체를 한 임계구역에 | 교착 없음 — 다른 잠금을 들고 있지 않다 | ast.json calls/defers |
| `previous.usableAt(at)` | 자격 판정 | 순수 | ast.json calls |
| `strings.ToUpper`/`TrimSpace` | 시장 키 정규화 | 순수 | ast.json calls |

## State mutations and fallbacks

- `o.seen[key]`: `whole`일 때만 교체. 프로세스 내 메모리이며 디스크·계좌 무관.
- `o.mu` 아래에서만 읽고 쓴다. `Read`는 이 함수 밖에서 `o.seen`을 만지지 않는다.
- fallback 없음 — 자격 없는 기억은 대체되지 않고 `unknown`이 된다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (메모리 상태). 재는 성질은 High-risk 인접 — 이 함수가 짧은 읽기를 기억으로 삼으면 다음 온전한 읽기의 패널 전체가 `신규 진입`으로 **측정**되고, 그 사실은 첫 관측 칼럼에 영구 기록된다.
