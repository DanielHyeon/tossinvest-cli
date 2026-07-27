# Function Logic Map: `verifyPage.RefreshSeconds`

- Source: `internal/console/pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

change refresh-positions-screen에서 신설된 무상태 leaf 함수다(diff 교차로 evidence 요구). dashboardPage와 동일한 2초 보존.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (없음 — 수신자 무상태) | — | head 템플릿 `{{.RefreshSeconds}}` 소비 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 | `2` 상수 | `TestTheVerificationScreensKeepTheirTwoSecondReload` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 없음(상수 반환 순수 함수). 템플릿 전용.

## Safety conclusion

- Safe edit boundary: 상수 값만 — 2초는 기존 하드코드의 보존이다
- High-risk impact: no (콘솔 read-only 렌더 경로 — 주문·위험·원장 코드 무접촉)
