# Function Logic Map: `IssueEntry`

- Source: `internal/execgw/riskguardian.go` (411-559, 편집 뒤)
- Qualified: `RiskGuardian.IssueEntry`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 88cf7c183018aae8…, 편집 전 분기 수 13) — 현재 branches 13 · returns 12
- Risk scan: `risk-pattern-report.md`

**역할.** 일반 진입 발급 문(체인 1회 → 원자 발급).

**a136 편집.** collectIssue 클로저의 ReserveRequest 에 `DailyBuyLimit: g.dailyBuyLimit()` 한 칸.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req` | BUY | 호출자 | 오류 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if req.Collect == nil {` (if 412) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B2 | `if err != nil {` (if 416) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B3 | `if intent.Side != risk.SideBuy {` (if 419) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B4 | `if verdict := evaluateChain(in); !verdict.Allowed {` (if 433) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B5 | `if !verdict.Allowed {` (if 442) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B6 | `if err != nil {` (if 477) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B7 | `if err != nil {` (if 481) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B8 | `if req.strategyPlan == nil {` (if 513) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B9 | `} else {` (else 515) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B10 | `if collectErr != nil {` (if 519) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B11 | `if strategyErr == nil {` (if 524) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B12 | `if err != nil {` (if 530) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |
| B13 | `if refused, ok := refusal.(*IssueRefusal); ok && refused.Stage == StageIssuance {` (if 537) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run RiskGuardian` |

Exact AST return positions: 413 · 417 · 420 · 439 · 444 · 478 · 482 · 484 · 520 · 522 · 541 · 552

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.journal.RecordDecisionAndReserveWithRecollection` | 원자 발급 | 한도 거부 즉시 | AST |

## State mutations and fallbacks

- journal tx.

## Safety conclusion

- Safe edit boundary: 요청 칸 하나.
- High-risk impact: yes
