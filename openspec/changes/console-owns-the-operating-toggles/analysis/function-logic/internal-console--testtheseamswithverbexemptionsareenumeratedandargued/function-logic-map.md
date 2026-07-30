# Function Logic Map: `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued`

- Source: `internal/console/orders_static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

위의 대체본. 예외를 가질 수 있는 seam을 열거하고, 각 이름에 논증이 적혀 있을 것을 요구하며, 열거됐는데 예외가 없는 stale 항목도 잡는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` @ internal/console/orders_static_test.go:194 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/console/orders_static_test.go:195 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/console/orders_static_test.go:199 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `range` @ internal/console/orders_static_test.go:205 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/console/orders_static_test.go:206 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/console/orders_static_test.go:210 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `range` @ internal/console/orders_static_test.go:215 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B8 | `if` @ internal/console/orders_static_test.go:216 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B9 | `range` @ internal/console/orders_static_test.go:222 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B10 | `if` @ internal/console/orders_static_test.go:223 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B11 | `range` @ internal/console/orders_static_test.go:233 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B12 | `range` @ internal/console/orders_static_test.go:235 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B13 | `if` @ internal/console/orders_static_test.go:236 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B14 | `if` @ internal/console/orders_static_test.go:240 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B15 | `range` @ internal/console/orders_static_test.go:246 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B16 | `if` @ internal/console/orders_static_test.go:247 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B17 | `if` @ internal/console/orders_static_test.go:251 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
