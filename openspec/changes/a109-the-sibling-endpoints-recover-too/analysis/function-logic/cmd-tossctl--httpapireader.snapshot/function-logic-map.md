# Function Logic Map: `httpAPIReader.Snapshot`

- Source: `cmd/tossctl/httpapi_reader.go` (517-591)
- AST evidence: `ast.json` — AST 분기 10 · return 8 · defer 0
  (source_sha256 `876f5a33600f0de326089dd9b15176ec38c30f4183db8c4983b14fd0e8b1fba2`,
  a109 §2.3 편집 후 생성)
- Risk scan: `risk-pattern-report.md`
- **a109 편집 대상: B8 의 조건 한 줄** — 전략 reader 의 부재 판정을 `!= nil` 에서
  `!httpapi.StrategyRuntimeAbsent(...)` 로 바꾼다(freeze P1-4). 앞의 일곱 조회와
  a108 의 흡수 구조는 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 일곱 조회(engine·positions·orders·candidates·performance·settings·optimization) | 전부 성공해야 봉투가 선다 | 각 어댑터 | 하나라도 실패하면 **집계 전체가 오류** (B1–B7) |
| 전략 projection | **없어도 되고 실패해도 된다** | `r.strategyRuntime` | 부재 → dormant, 있는데 실패 → unavailable (B8·B9·B10) |
| 부재 판정 | nil 이거나 wrapper 가 부재라고 말함 | `httpapi.StrategyRuntimeAbsent` | 판정이 갈리면 화면 값이 틀린다 |

**불변식 1 (a108 D4-2)**: 전략 화면 하나가 집계 스냅샷 전체를 죽이지 않는다. 앞의 여섯
조회가 전부 성공했는데 화면이 통째로 비는 것이 a108 이 지운 비대칭이다.

**불변식 2 (a108 D4-2, a109 가 보존)**: 부재(dormant/NOT_CONFIGURED)와 실패
(unavailable/RUNTIME_UNAVAILABLE)를 **한 값으로 접지 않는다**. 재부착 wrapper 는 정의상
non-nil 이므로 nil 검사로는 그 구분을 유지할 수 없다 — 그래서 상태 신호로 물어야 한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:519) | engine 조회 실패 | 없음 | 오류 | 기존 httpapi reader 테스트 |
| B2 (:523) | positions 실패 | 없음 | 오류 | 기존 |
| B3 (:527) | orders 실패 | 없음 | 오류 | 기존 |
| B4 (:531) | candidates 실패 | 없음 | 오류 | 기존 |
| B5 (:535) | performance 실패 | 없음 | 오류 | 기존 |
| B6 (:539) | settings 실패 | 없음 | 오류 | 기존 |
| B7 (:543) | optimization 실패 | 없음 | 오류 | 기존 |
| **B8 (:570)** | **전략 reader 가 부재가 아니다** | 없음 | 아래 둘 | `TestAnAbsentStrategyReaderStaysDormantRatherThanUnavailable` · `TestTheDaemonAttachesWhenTheEngineComesUpLater` |
| B9 (:572) | 전략 읽기 실패 | 없음 | **unavailable 스냅샷** (오류 아님) | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` |
| B10 (:574) | 전략 읽기 성공 | 없음 | 읽은 값 | 위와 같음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 일곱 read 어댑터 | 집계의 본문 | 실패는 전체 오류 (B1–B7) | AST · :518–545 |
| `httpapi.StrategyRuntimeAbsent` | 부재 판정 **한 벌** | 순수 함수 | AST · :570 |
| `r.strategyRuntime.Read` | 전략 읽기 | 실패는 **흡수**되어 unavailable 값이 된다 | AST · :571 |
| `strategyprojection.DormantSnapshot` / `UnavailableSnapshot` | 두 화면 값 | 순수 함수 | AST · :569·573 |

## State mutations and fallbacks

- 상태 변경 없음 — 조회 집계다.
- fallback 둘: 부재 → dormant, 읽기 실패 → unavailable. **둘을 접지 않는 것**이 계약이다.

## Safety conclusion

- Safe edit boundary: **B8 의 조건**. 흡수 구조(B9·B10)와 앞의 일곱 조회는 그대로.
- High-risk impact: **no** — 조회 전용 집계다.
- 금지: 전략 읽기 실패를 오류로 올리는 것(집계 전체가 죽는다). 부재 판정의 두 번째 사본.
