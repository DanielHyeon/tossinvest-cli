# Function Logic Map: `httpAPIReader.Snapshot`

- Source: `cmd/tossctl/httpapi_reader.go` (450-513)
- AST evidence: `ast.json` — AST branches 10 · returns 8 · defers 0
  (source_sha256 `7abe45c66b68edbcb716f97feacbd2330bd2dbbde9c3e053ba66f7c67200397b`,
  **Fix 라운드 6.8② 편집 후 생성**)
- Risk scan: `risk-pattern-report.md`
- 편집 대상: **B8·B9·B10** (겹3 Fix, design D4-2). 기준(base) 판은 분기 9개·return 9개였고,
  사라진 return 이 전략 `Read` 실패의 `return nil, err` 다 — 여섯 조회가 전부 성공했는데도
  집계 화면 전체를 비우던 줄.

## 이 함수가 하는 일 (AST 열거 기준)

`ast.json` 의 `calls` 가 여덟 개의 읽기를 순서대로 열거한다: `r.Engine` → `r.Positions`
→ `r.Orders` → `r.Candidates` → `r.Performance` → `r.Settings` → `r.Optimization` →
`r.strategyRuntime.Read`. 앞의 일곱은 각각 `if err != nil { return nil, err }`
(B1~B7, returns 1~7) 로 **fail-closed** 다. 여덟 번째만 흡수한다.

이 비대칭이 설계다. 앞의 일곱은 **이 데몬의 자기 저장소**(원장·브로커 판독·성과 DB·
최적화 DB)이고, 그중 하나가 실패했다는 것은 스냅샷이 무엇을 말하든 믿을 수 없다는
뜻이다. 여덟 번째는 **다른 프로세스(엔진)가 소유한 조회 전용 export** 이고, 그것이
사라지는 것은 엔진 재시작·잔재 회수 중 정상적으로 일어난다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.engineMarker` | 빈 문자열이면 조회 불가 | `httpAPIReader` 조립 | 비었으면 오류 → 집계 실패 (B1) |
| `r.holdings` / `r.accountRef` | 필요 | 브로커 판독 seam | 실패는 집계 실패 (B2) |
| `r.orders` | 필요 | 브로커 판독 seam | 실패는 집계 실패 (B3) |
| `r.signals` | 필요 | 후보 저장소 | 실패는 집계 실패 (B4) |
| `r.performance` | 필요 | 성과 DB (읽기 전용) | 실패는 집계 실패 (B5) |
| `r.optimization` + `r.adoptionDesired` | 필요 | 최적화 DB·설정 | 실패는 집계 실패 (B6·B7) |
| **`r.strategyRuntime`** | **nil 이어도 된다** | `runHTTPAPI` 의 강등 결과 | **nil → dormant(NOT_CONFIGURED) (B8 의 거짓 가지)** |
| **`r.strategyRuntime.Read`** | **실패해도 된다 (a108 6.8②)** | 엔진의 unix projection | **unavailable(RUNTIME_UNAVAILABLE) (B9)** |
| `r.clockNow()` | 항상 UTC | `r.now` 또는 `time.Now` | 없음 |

**관통 불변식:** 집계 스냅샷은 **자기 저장소**의 실패에는 fail-closed 이고,
**남의 조회 표면**의 실패에는 degrade 한다. 두 실패를 같은 규칙으로 다루면
「엔진 재시작 3초」가 「운영자 화면 전체 3초 정전」이 된다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `r.Engine` 실패 | 없음 | `nil, err` | 미고정(조립 불변 — `validate()` 가 앞을 막는다) |
| B2 | `r.Positions` 실패 | 없음 | `nil, err` | 미고정 |
| B3 | `r.Orders` 실패 | 없음 | `nil, err` | 미고정 |
| B4 | `r.Candidates` 실패 | 없음 | `nil, err` | 미고정 |
| B5 | `r.Performance` 실패 | 없음 | `nil, err` | 미고정 |
| B6 | `r.Settings` 실패 | 없음 | `nil, err` | 미고정 |
| B7 | `r.Optimization` 실패 | 없음 | `nil, err` | 미고정 |
| **B8** | **`r.strategyRuntime != nil`** | 없음 | 없음 (거짓 가지 = dormant 유지) | `TestAnAbsentStrategyReaderStaysDormantRatherThanUnavailable` |
| **B9** | **`Read` 실패 → `UnavailableSnapshot`** | 지역 변수 대입만 | **없음 — 집계는 계속된다** | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` |
| **B10** | **`Read` 성공 → 읽은 값 대입** | 지역 변수 대입만 | 없음 | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` (대조군 절반) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.Engine` … `r.Optimization` (7건) | 집계의 나머지 여섯 자원 | 전부 fail-closed | AST calls · httpapi_reader.go:451-475 |
| `strategyprojection.DormantSnapshot` | reader 부재의 기본값 | 순수 함수 | AST · model.go:174 |
| **`r.strategyRuntime.Read`** | 전략 화면 | **실패는 흡수 (a108 D4-2)**. 재시도 없음 — 다음 스냅샷 주기가 재시도다 | AST · a108 T2 테스트 |
| `strategyprojection.UnavailableSnapshot` | 읽기 실패의 정직한 투영 | 순수 함수 | AST · model.go:182 |
| `json.Marshal` | 봉투 직렬화 | 실패는 그대로 올라간다(returns[8]) | AST |

## State mutations and fallbacks

- **이 함수는 아무것도 쓰지 않는다.** 부작용은 하위 read 들의 캐시 갱신
  (`positionsCache`·`ordersCache`)뿐이고 이 편집은 거기 닿지 않았다.
- 흡수의 결과는 **두 값 중 하나**이며 둘은 다른 사실을 말한다:
  - `DormantSnapshot` → `NOT_CONFIGURED`: 이 배포는 전략 화면을 안 쓴다.
  - `UnavailableSnapshot` → `RUNTIME_UNAVAILABLE`: 붙었는데 못 읽는다(엔진이 없다).
  접으면 운영자는 「기능을 안 켰다」와 「엔진이 죽었다」를 구별할 수 없다. 콘솔이
  이미 같은 두 값을 쓴다(`internal/console/strategy_runtime_multimarket.go:42-52`) —
  두 소비자가 같은 상태를 다르게 그리던 갈림을 여기서 닫는다.
- `strategyprojection.Validate` 는 **부르지 않는다**(콘솔은 부른다). 편집 범위를
  「Read 오류 흡수」로 좁힌 결과이며, 성공했지만 형식이 깨진 스냅샷의 처리는
  이 change 이전과 같다 — 선언된 생략이다.

## Safety conclusion

- Safe edit boundary: B8·B9·B10. B1~B7 의 fail-closed 판정과 **순서**는 건드리지
  않았다 — 순서가 바뀌면 캐시 갱신 순서가 바뀐다.
- High-risk impact: **no.** 이 함수는 조회 전용 데몬의 직렬화 경로이고 주문·손절·
  사이징에 닿지 않는다. 변경 방향은 「화면이 더 자주 보이는」 쪽이며, 화면이
  **거짓으로** 보이는 방향은 아니다: 실패는 지워지지 않고 `RUNTIME_UNAVAILABLE`
  이라는 이름으로 스냅샷 안에 그대로 실린다.
