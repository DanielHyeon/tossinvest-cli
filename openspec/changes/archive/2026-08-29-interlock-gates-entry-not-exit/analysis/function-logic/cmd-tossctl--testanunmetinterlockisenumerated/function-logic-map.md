# Function Logic Map: `TestAnUnmetInterlockIsEnumerated`

- Source: `cmd/tossctl/engine_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

열거를 확인할 때 쓰던 조항을 6에서 8(ErrKeylessTransport)로 바꿨다. 조항 6은 더 이상 기동을 거부하지 않으므로, 그것으로 열거를 시험하면 운영자가 볼 수 없는 메시지를 단언하게 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ cmd/tossctl/engine_test.go:166 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ cmd/tossctl/engine_test.go:169 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ cmd/tossctl/engine_test.go:172 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
