# Function Logic Map: `TestTheConsoleDecidesNothingAboutTheGate`

- Source: `internal/console/engineproc_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

면제 파일 목록에 templates_settings.go를 더했다. 금지된 두 단어(Interlock, ProtectionReady)는 모든 파일에서 그대로 금지다 — 콘솔은 여전히 기동 가능 여부를 계산하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range` @ internal/console/engineproc_test.go:209 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/console/engineproc_test.go:212 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `range` @ internal/console/engineproc_test.go:215 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B4 | `if` @ internal/console/engineproc_test.go:216 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
