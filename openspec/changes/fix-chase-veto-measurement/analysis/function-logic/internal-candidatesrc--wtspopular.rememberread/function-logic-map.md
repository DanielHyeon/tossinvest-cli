# Function Logic Map: `wtsPopular.rememberRead`

- Source: `internal/candidatesrc/candidatesrc.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L510–535, 분기 5개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. `officialRanking.rememberRead`를 이 소스의 행 타입에 대해 반복한 것이고,
두 조건(온전함, `previousReadingTTL`)과 그 이유가 같다.

두 어댑터가 **같이** 이 규칙을 들어야 한다. 한쪽만 들면 화면이 설명하지 않는 이유로 패널의
절반만 측정 가능해진다 — `TestTheWTSMemoryCarriesTheSameTwoConditions`가 그 대칭을 고정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | 임의 문자열 | `Read`(KR만 통과) | 정규화해 키로 쓴다 |
| `stocks` | 이번 읽기의 행 | WTS | `wtsSymbol`이 빈 것은 제외 |
| `requested` | `w.size` — 요청 행 수 | `Read` | 이 함수가 `current`와 비교해 `whole`을 정한다(issues.md I16) |
| `w.seen` / `w.now` | 시장별 기억 / 주입 clock | 이 함수 / `WTSPopular` | nil 맵은 여기서 생성 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, s := range stocks` | `current` 구성 | — | `TestTheWTSPopularityListReportsTheSameTwoFacts` |
| B2 | `wtsSymbol(s) != ""` | 집합에 넣지 않고, 그래서 `len(current) < len(stocks)`가 되어 B4가 교체를 거부한다 | — | `TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps` |
| B3 | `w.seen == nil` | 맵 생성 | — | `TestTheWTSPopularityListReportsTheSameTwoFacts`(첫 읽기) |
| B4 | `whole`(이 함수가 계산) | `w.seen[key]` 교체 | — | `TestTheWTSMemoryCarriesTheSameTwoConditions`(1행/2행) · `TestTheWTSMemoryIsAlsoBuiltFromTheRowsItKeeps`(식별자 없는 행) |
| B5 | `!previous.usableAt(at)` | 없음 | `nil, false` | `TestTheWTSMemoryCarriesTheSameTwoConditions`(TTL 경과) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `w.now.Now()` | 이번 읽기의 instant | 주입 clock | ast.json calls |
| `w.mu.Lock`/`defer Unlock` | 취득과 교체를 한 임계구역에 | 다른 잠금 없음 | ast.json calls/defers |
| `wtsSymbol` | 집합 키 | 순수 — 행과 같은 함수 | ast.json calls |
| `previous.usableAt(at)` | 자격 판정 | 순수 | ast.json calls |

## State mutations and fallbacks

- `w.seen[key]`: `whole`일 때만 교체. 프로세스 내 메모리.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음. 2026-07-28 편집은 공식 어댑터와 동일하다 — `whole bool` → `requested int`, 그리고 함수 안 한 줄.
- High-risk impact: no (메모리 상태). 공식 어댑터와 같은 위험 형태이며, 이 소스는 KR 전용이라 영향 범위가 한 시장이다.
