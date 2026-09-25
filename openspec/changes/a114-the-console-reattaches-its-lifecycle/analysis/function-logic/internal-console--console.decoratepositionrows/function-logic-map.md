# Function Logic Map: `Console.decoratePositionRows`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json` (base `634cf3c5`, 분기 12)
- Risk scan: `risk-pattern-report.md`

**a114 는 이 함수를 편집하지 않는다.** design 이 이 함수의 분기를 근거로 쓰므로(FLM-before-claiming)
AST 를 먼저 만들었다. 근거: B1(:99) `PositionPolicies != nil` 이 `runtimeAttempted` 를 켜고, 읽기 실패는 `policyByID` nil(B2 :119 불성립) → 행이 「관리 여부 불명」(fail-closed)으로 간다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.PositionPolicies` | nil 또는 commander | runConsole | nil 이면 관리 판정 생략 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c.opts.PositionPolicies != nil {` (:99) | — | — | **근거**: commander non-nil → runtimeAttempted |
| B2 | `if reading.StatesErr == nil {` (:119) | — | — | **근거**: List 성공일 때만 policyByID |
| B3 | `for _, state := range reading.States {` (:121) | — | — | 편집 없음 |
| B4 | `if c.opts.Settings != nil {` (:126) | — | — | 편집 없음 |
| B5 | `if block, _, err := c.opts.Settings.Load(); err == nil {` (:127) | — | — | 편집 없음 |
| B6 | `for i := range rows {` (:130) | — | — | 편집 없음 |
| B7 | `if runtimeAttempted {` (:143) | — | — | **근거**: runtimeAttempted 일 때만 관리 판정 투영 |
| B8 | `for i := range rows {` (:144) | — | — | 편집 없음 |
| B9 | `if row.InJournal {` (:148) | — | — | 편집 없음 |
| B10 | `if !ok {` (:151) | — | — | 편집 없음 |
| B11 | `} else {` (:153) | — | — | 편집 없음 |
| B12 | `if row.Management.Block != nil {` (:163) | — | — | 편집 없음 |

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
