# Function Logic Map: `TestTheTracerDrivesEntryRatchetAndExit`

- Source: `internal/app/engine/tracer_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

스택 생성자를 newEntryCapableE2EStack으로 바꿨다. 이 슬라이스는 매수를 내므로 보호 배선된 게이트웨이가 필요하다 — 보호주문 change가 landed된 뒤의 형태로 도는 것이고, 그것을 이름으로 말한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/app/engine/tracer_test.go:128 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/app/engine/tracer_test.go:131 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/app/engine/tracer_test.go:134 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/app/engine/tracer_test.go:137 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B5 | `if` @ internal/app/engine/tracer_test.go:140 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B6 | `if` @ internal/app/engine/tracer_test.go:143 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
