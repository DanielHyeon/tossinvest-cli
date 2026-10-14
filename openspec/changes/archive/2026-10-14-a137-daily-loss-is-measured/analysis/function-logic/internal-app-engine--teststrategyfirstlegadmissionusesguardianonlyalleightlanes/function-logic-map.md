# Function Logic Map: `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes`

- Source: `internal/app/engine/strategy_first_leg_admission_test.go` (32-58, 편집 뒤)
- Qualified: `TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` ae208ecf8187b894…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 7b2eca4d50c6fb8a…, 편집 전 분기수 4) — 현재 branches 4 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** 첫 다리 승인이 Guardian 경계만 쓰는지 여덟 레인으로 확인하는 시험.

**a137 편집.** 기대 사건 열을 `measure,precheck,issue` 로 — 승인이 precheck 직전에 측정함(a137). 분기 수 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 여덟 레인 결과 | 시험 | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, lane := range allPairedStrategyFirstLegResults(t) {` (range 34) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B2 | `if outcome.Code != StrategyFirstLegAdmitted \|\| loader.calls != 1 \|\| issuer.precheckCalls != 1 \|\| issuer.issueCalls != 1 {` (if 44) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B3 | `if issuer.events != "measure,precheck,issue" \|\| issuer.requests[0].Result.Lineage.Identity != lane.result.Lineage.Identity \|\|` (if 47) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B4 | `if markets["KR"] != 4 \|\| markets["US"] != 4 {` (if 55) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `admit` | 승인 | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 기대 문자열 하나.
- High-risk impact: no — 시험.
