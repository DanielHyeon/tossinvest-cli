# Function Logic Map: `firstLegGuardianSpy.MeasureDailyLoss`

- Source: `internal/app/engine/strategy_first_leg_admission_test.go` (177-180, 편집 뒤)
- Qualified: `firstLegGuardianSpy.MeasureDailyLoss`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` ae208ecf8187b894…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 608c815a0bd96c02…, 편집 전 분기수 0) — 현재 branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 첫 다리 승인 시험의 Guardian 대역 — 측정 호출 기록.

**a137 편집.** 구현 리뷰 P3: 영값 대신 `s.token` 을 돌려줌(실제 Guardian 토큰 운반 시험용). 분기 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| s.token | 시험이 넣은 토큰 | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 | — | — | `go test -tags tossos_testseams ./internal/app/engine -run TestA137AdmissionCarriesTheMeasuredTokenIntoThePrecheck` |

Exact AST return positions: 179

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | — |

## State mutations and fallbacks

- s.events.

## Safety conclusion

- Safe edit boundary: 시험 대역.
- High-risk impact: no — 시험.
