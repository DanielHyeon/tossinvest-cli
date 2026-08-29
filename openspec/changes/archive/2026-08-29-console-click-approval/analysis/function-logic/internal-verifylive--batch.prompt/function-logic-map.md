# Function Logic Map: `Batch.Prompt`

- Source: `internal/verifylive/confirm.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| b (Batch) | Summary가 요구하는 것 + Nonce·ExpiresAt | NewBatch | Summary 뒤에 타이핑 꼬리를 붙인다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 (internal/verifylive/confirm.go) | 아래 State mutations 참조 | 정상 반환 | TestPromptIsTheSummaryPlusTheTypedTail, TestConfirmBatchAcceptsTheTypedNonce |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Batch.Summary | 목록 부분을 그대로 재사용 — 두 렌더링이 갈라지지 않게 | 없음 | AST callees |

## State mutations and fallbacks

- 없음 — 순수 함수.
- TTY 출력의 의미는 무변경이다: 목록·확인 문자열·만료·입력 지시가 모두 남는다.

## Safety conclusion

- Safe edit boundary: Prompt는 Summary로 시작해야 한다(테스트로 고정). 꼬리만 TTY 전용이다.
- High-risk impact: no — 순수 렌더링. CLI 승인 절차는 변하지 않는다.
