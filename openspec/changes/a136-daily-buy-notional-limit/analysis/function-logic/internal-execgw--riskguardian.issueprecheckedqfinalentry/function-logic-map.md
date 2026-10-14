# Function Logic Map: `IssuePrecheckedQFinalEntry`

- Source: `internal/execgw/riskguardian_qfinal.go` (177-260, 편집 뒤)
- Qualified: `RiskGuardian.IssuePrecheckedQFinalEntry`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 3e5cc7b86303b458…, 편집 전 분기 수 11) — 현재 branches 11 · returns 12
- Risk scan: `risk-pattern-report.md`

**역할.** q_final 진입 발급 문(봉인 precheck → 원자 발급).

**a136 편집.** collect 클로저의 ReserveRequest 에 `DailyBuyLimit: g.dailyBuyLimit()` 한 칸.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `precheck` | 봉인 완료 | PrecheckQFinalEntry | qFinalRefusal |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil \|\| precheck.decision.QFinal == 0 \|\| precheck.request.Collect == nil \|\| precheck.admission.TransactionID == "" {` (if 178) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B2 | `if (err != nil && !legacyTestIdentity) \|\| (!legacyTestIdentity && !sameAccountBaseFX(precheck.accountFX, accountFX)) {` (if 184) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B3 | `if err != nil {` (if 188) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B4 | `if !legacyTestIdentity {` (if 192) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B5 | `if err != nil {` (if 194) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B6 | `if err != nil {` (if 216) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B7 | `if !legacyTestIdentity && !sameQFinalFXAuthority(accountFX, policy) {` (if 219) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B8 | `if err != nil {` (if 230) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B9 | `if !verdict.Allowed {` (if 234) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B10 | `if err != nil {` (if 238) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |
| B11 | `if err != nil {` (if 252) | — | 분기 본문 참조 | `TestA136EveryEntryDoorCarriesTheDailyBuyCeiling` · `go test ./internal/execgw -run QFinal` |

Exact AST return positions: 179 · 185 · 189 · 195 · 217 · 220 · 231 · 235 · 239 · 241 · 253 · 255

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.journal.RecordQFinalDecisionAndReserveWithRecollection` | 원자 발급 | issuanceRefusal | AST |

## State mutations and fallbacks

- journal tx.

## Safety conclusion

- Safe edit boundary: 요청 칸 하나.
- High-risk impact: yes
