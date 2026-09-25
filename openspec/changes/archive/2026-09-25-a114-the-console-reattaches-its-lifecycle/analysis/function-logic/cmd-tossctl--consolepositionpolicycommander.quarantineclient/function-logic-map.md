# Function Logic Map: `consolePositionPolicyCommander.quarantineClient`

- Source: `cmd/tossctl/exit_quarantine_commander.go`
- AST evidence: `ast.json` (base `634cf3c5`, 분기 1)
- Risk scan: `risk-pattern-report.md`

**a114 는 이 함수를 편집하지 않는다.** design 이 이 함수의 분기를 근거로 쓰므로(FLM-before-claiming)
AST 를 먼저 만들었다. 근거: B1(:28) `c.lifecycle.(exitQuarantineClient)` 단언이 실패하면 `exitquarantine.ErrUnwired` — lifecycle 자리를 감싸는 wrapper 가 세 격리 메서드를 구현하지 않으면 격리 해제가 사라진다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.lifecycle` | positionPolicyLifecycleClient | commander 생성자 | 세 메서드 없으면 ErrUnwired |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if !ok {` (:28) | — | — | **근거**: 격리 해제 발견의 단언 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.opts.PositionPolicies` 의 메서드 | 엔진 lifecycle 읽기/명령 | 오류는 화면 값으로 | AST calls |

## State mutations and fallbacks

- 편집 없음. a114 이후 이 함수가 받는 commander 는 engineDir 가 있으면 non-nil wrapper 이고, 부착 전
  호출은 연결 없는 detached 오류를 돌려준다.

## Safety conclusion

- Safe edit boundary: 편집 없음(0줄).
- High-risk impact: no.
