# Function Logic Map: `TestTheOrdersSeamIsTheOnlyCapabilityWithVerbExemptionsAndTheyAreEnumerated`

- Source: `internal/console/orders_static_test.go`
- AST evidence: `ast.json` (revision: base — 이 change에서 삭제된 함수)
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

삭제됨 — TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued로 대체. '정확히 하나'는 당시 코드에 대한 사실을 규칙으로 적은 것이었고, 진짜 보장은 '논증 없는 예외 금지'였다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` @ internal/console/orders_static_test.go:183 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/console/orders_static_test.go:184 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/console/orders_static_test.go:187 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `range` @ internal/console/orders_static_test.go:193 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/console/orders_static_test.go:194 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/console/orders_static_test.go:198 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B7 | `range` @ internal/console/orders_static_test.go:203 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B8 | `if` @ internal/console/orders_static_test.go:204 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B9 | `range` @ internal/console/orders_static_test.go:215 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B10 | `range` @ internal/console/orders_static_test.go:217 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B11 | `if` @ internal/console/orders_static_test.go:218 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B12 | `if` @ internal/console/orders_static_test.go:222 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B13 | `range` @ internal/console/orders_static_test.go:228 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B14 | `if` @ internal/console/orders_static_test.go:229 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B15 | `if` @ internal/console/orders_static_test.go:233 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
