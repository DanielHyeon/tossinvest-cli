# Function Logic Map: `Batch.Expired`

- Source: `internal/verifylive/confirm.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| now | 호출자의 시계 | console.now / runner.now | ExpiresAt 이후면 true |
| b.ExpiresAt | NewBatch 시각 + BatchApprovalTTL(5분) | NewBatch | 경계는 배타적: now==ExpiresAt은 만료다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 (internal/verifylive/confirm.go) | 아래 State mutations 참조 | 정상 반환 | TestExpiredIsTheSameWindowVerifyUses, TestAnExpiredApprovalSendsNothing(console) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| time.Time.Before | 창 판정 | 없음 | AST callees |

## State mutations and fallbacks

- 없음 — 순수 판정. Verify가 이 함수를 호출하므로 창 규칙은 저장소에 한 번만 존재한다.

## Safety conclusion

- Safe edit boundary: Verify와 같은 답을 줘야 한다(테스트로 고정). 콘솔이 자기 시계 규칙을 갖게 되는 것이 이 함수가 막는 실패다.
- High-risk impact: no — 판정만 한다. 다만 승인 창의 정의이므로 느슨해지면 묵은 가격의 계획이 승인될 수 있다.
