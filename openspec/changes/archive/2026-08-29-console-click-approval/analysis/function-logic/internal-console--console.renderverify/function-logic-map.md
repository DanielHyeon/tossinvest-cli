# Function Logic Map: `Console.renderVerify`

- Source: `internal/console/pages.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 스냅샷 | 기록·soak·attest 판독 | readVerify 등 | 읽기 실패는 화면 안내 |
| 현재 run | nil 가능 | currentRun | 배치가 있으면 요약을 렌더 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if run := c.currentRun(); run != nil {` (internal/console/pages.go:138, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalScreenAsksForNoTypedString, TestTheApprovedFlowRunsExactlyTheApprovedBatch |
| B2 | `if v.Batch != nil {` (internal/console/pages.go:141, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalScreenAsksForNoTypedString, TestTheApprovedFlowRunsExactlyTheApprovedBatch |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| verifylive.Batch.Summary | 승인 화면의 목록 — Prompt(타이핑 꼬리 포함) 대신 | 없음 | AST callees |
| verifylive.WriteSteps | 단계 목록(계좌 무접촉) | 없음 | AST callees |

## State mutations and fallbacks

- 없음 — 렌더링만 한다. 브로커 호출도 하지 않는다.
- Prompt→Summary 교체가 유일한 변경.

## Safety conclusion

- Safe edit boundary: 화면은 계획 전체를 보여줘야 하고(배치 모델의 근거), 화면에서 성립하지 않는 승인 방식을 안내해서는 안 된다.
- High-risk impact: no — 읽기·렌더링.
