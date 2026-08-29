# Function Logic Map: `openE2EStack`

- Source: `internal/app/engine/exit_e2e_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

본문이 buildE2EStack 위임으로 바뀌었다. 동작은 동일 — protectionReady=false, 즉 출하 형태.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 픽스처 | 함수 본문이 세운다 | 테스트 자신 | 단언 실패 시 t.Error/t.Fatal |

## Branches and early returns

분기는 전부 단언 가드다. 각 가드는 실패 시 이 테스트 자신이 보고한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — happy path | 없음 | 정상 반환 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 테스트 대상 API | 이 테스트가 무엇을 보는지 | 단언으로 처리 | AST |

## State mutations and fallbacks

- 격리된 임시 디렉터리와 임시 journal 외에는 없다.

## Safety conclusion

- Safe edit boundary: 테스트 함수. 프로덕션 동작을 바꾸지 않는다.
- High-risk impact: no
