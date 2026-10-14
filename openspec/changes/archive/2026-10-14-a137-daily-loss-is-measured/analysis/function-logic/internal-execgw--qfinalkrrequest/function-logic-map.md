# Function Logic Map: `qFinalKRRequest`

- Source: `internal/execgw/riskguardian_qfinal_test.go` (305-351, 편집 뒤)
- Qualified: `qFinalKRRequest`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 2f2a1e6036fcf9c6…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 76e4a5487df45d5e…, 편집 전 분기수 3) — 현재 branches 3 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** q_final KR 요청 시험 fixture.

**a137 편집.** 끝에 `request.DailyLoss = rig.guardian.MeasureDailyLoss(context.Background())` — precheck 가 토큰을 요구하므로 생산의 첫 다리 승인과 같은 자리를 흉내 냄. 분기 수 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rig | 시험 Guardian | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, dimension := range riskbucket.RequiredDimensionOrder() {` (range 320) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B2 | `if err != nil {` (if 324) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B3 | `if err != nil {` (if 330) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |

Exact AST return positions: 350

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `rig.guardian.MeasureDailyLoss` | 토큰(a137) | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: fixture 한 줄.
- High-risk impact: no — 시험 fixture.
