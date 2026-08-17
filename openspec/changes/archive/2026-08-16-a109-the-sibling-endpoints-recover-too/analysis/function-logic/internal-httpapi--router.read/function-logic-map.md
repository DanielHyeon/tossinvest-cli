# Function Logic Map: `router.read`

- Source: `internal/httpapi/router.go` (118-171)
- AST evidence: `ast.json` — AST 분기 17 · return 12 · defer 0
  (source_sha256 `4860a766f4d8d2dc84a1ce6af2678fbeb24c3f0283ff28616594a3dbbed73e42`,
  a109 §2.3 편집 후 생성)
- Risk scan: `risk-pattern-report.md`
- **a109 편집 대상: B14 한 줄** — `strategy-runtime` case 의 부재 판정을 nil 검사에서
  `StrategyRuntimeAbsent` 상태 신호로 바꾼다. 나머지 16 분기는 건드리지 않는다.
- **왜 이 함수가 편집 목록에 있는가**: 설계 D4 는 소비자 nil 검사를 두 곳으로 열거했고
  (집계 스냅샷·SSE helper) **production REST 경로인 여기가 빠져 있었다**
  (issues.md T2-1). 고치지 않으면 재부착 wrapper 가 꽂힌 순간 이 경로의 부재 응답이
  dormant 스냅샷에서 오류로 바뀐다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `resource` | 8개 이름 중 하나 | 라우팅 표 | 모르는 이름은 오류 (B17 default) |
| `r.reader` | 조립 시 non-nil 보장 | `NewRouter` | 각 read 의 오류를 그대로 올린다 |
| `r.strategyRuntime` | **nil 이거나 부재를 말하는 wrapper 이면 부재** | `StrategyRuntimeAbsent` | 부재 → **200 + dormant 스냅샷**(B14), 있는데 못 읽음 → 오류(B15) |
| `r.now` | UTC 시각원 | `NewRouter` | dormant 스냅샷의 생성 시각 |

**불변식**: 부재(dormant)와 읽기 실패(unavailable)는 **다른 응답**이다. 접으면 운영자는
「이 배포는 전략 화면을 안 쓴다」와 「엔진이 죽었다」를 구별할 수 없다(a108 D4-2).

**불변식 2**: 목록 자원(positions·orders·candidates·settings)은 nil 슬라이스를 빈 배열로
정규화한다(B4·B6·B8·B11) — null 과 [] 은 디코더에게 같은 사실이고 화면에서는 다른 말이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:120) | 자원 이름 분기 | 없음 | 아래 case 들 | `TestStrategyRuntimeRESTUsesSharedProjectionAndStrictGuards` |
| B2 (:121) | `engine` | 없음 | reader 결과 | 기존 REST 계약 테스트 |
| B3 (:123) · B4 (:125) | `positions`, 목록 nil | 빈 배열 정규화 | reader 결과 | 기존 |
| B5 (:129) · B6 (:131) | `orders`, 목록 nil | 빈 배열 정규화 | reader 결과 | 기존 |
| B7 (:135) · B8 (:137) | `candidates`, 목록 nil | 빈 배열 정규화 | reader 결과 | 기존 |
| B9 (:141) | `performance` | 없음 | 변환 결과 | 기존 |
| B10 (:144) · B11 (:146) | `settings`, 목록 nil | 빈 배열 정규화 | reader 결과 | 기존 |
| B12 (:150) | `optimization` | 없음 | 변환 결과 | 기존 |
| B13 (:153) | `strategy-runtime` | 없음 | 아래 셋 | `TestStrategyRuntimeRESTUsesSharedProjectionAndStrictGuards` |
| **B14 (:157)** | **`StrategyRuntimeAbsent(r.strategyRuntime)`** | 없음 | **200 + dormant 스냅샷** | `TestTheRESTRouteStaysDormantForAnUnconfiguredWrapper` |
| B15 (:161) | 읽기 실패 | 없음 | 오류 → unavailable 응답 | 기존 REST 계약 테스트 |
| B16 (:164) | 스냅샷이 계약 위반 | 없음 | 오류 | 기존 |
| B17 (:168) | 모르는 자원 | 없음 | 오류 | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrategyRuntimeAbsent` | 부재 판정 **한 벌** | 순수 함수 — nil 이거나 wrapper 가 부재라고 말하면 true | AST · router.go:157 · strategy_runtime.go |
| `strategyprojection.DormantSnapshot` | 부재의 화면 값 | 순수 함수 | AST · router.go:158 |
| `r.strategyRuntime.Read` | 붙어 있을 때의 실제 읽기 | 실패는 오류로 올린다(B15) | AST · router.go:160 |
| `strategyprojection.Validate` / `Clone` | 계약 검사·복사 | 위반은 오류 | AST · router.go:164·167 |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기 라우터다. 목록 정규화(nil → 빈 배열)만이 값의 변형이다.
- fallback 은 B14 하나: 부재는 오류가 아니라 **dormant 스냅샷**이다.

## Safety conclusion

- Safe edit boundary: **B14 의 조건 한 줄**. 다른 case 와 정규화는 건드리지 않는다.
- High-risk impact: **no** — 조회 전용 REST 다. 다만 잘못 바꾸면 화면 값이 조용히
  틀린다(부재 ↔ 장애 접힘).
- 금지: 이 자리에 `StrategyRuntimeAbsent` 의 **두 번째 사본**을 쓰는 것. 판정이 갈리면
  같은 디스크 상태가 화면마다 다른 값이 된다(a098 D7.1).
