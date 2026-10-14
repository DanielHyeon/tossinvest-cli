# Function Logic Map: `collectStrategyFirstLegAuthority`

- Source: `internal/app/engine/strategy_account_first_leg_authority.go` (288-399, 편집 뒤)
- Qualified: `productionStrategyFirstLegAuthorityLoader.collectStrategyFirstLegAuthority`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` c29e90e2a1e9f04e…, 편집 전 분기 수 16) — 현재 branches 16 · returns 17
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 전략 첫 다리 발급 요청 조립.

**a136 편집.** `PairedFXAuthority: loader.fx.forMarket(otherStrategyMarket(market)).read.evidence` 한 칸 — 같은 파도 짝 FX 의 다른 시장 authority(FX 조항).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if loader == nil \|\| ctx == nil \|\| loader.clk == nil \|\| loader.journal == nil \|\| loader.guardian == nil {` (if 289) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B2 | `if !riskAuthority.snapshot.Ready \|\| !fx.snapshot.Ready \|\| !account.snapshot.Ready \|\|` (if 295) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B3 | `if !scoped {` (if 302) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B4 | `if result.Lineage.Identity != accepted.result.Lineage.Identity \|\| result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Ident` (if 306) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B5 | `if !proposal.familyActivation().Verified() && len(proposal.entries) != 1 {` (if 311) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B6 | `if !keyed { // 봉인이 정규화를 보장하므로 도달 불가 — 바뀌면 범위 거절이 아니라 결함(리뷰 A #4)` (if 318) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B7 | `if !riskScoped {` (if 324) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B8 | `if !scopeLocal {` (if 326) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B9 | `if !accountScoped {` (if 332) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B10 | `if !currencyKnown {` (if 339) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B11 | `if err := riskBundle.Validate(scope); err != nil \|\| scope.AccountID != result.Lineage.AccountRef \|\|` (if 343) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B12 | `if err != nil \|\| cas.Claimed \|\| cas.State != "FLAT" && cas.State != "CLOSED" \|\| cas.Generation < 0 \|\|` (if 348) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B13 | `for _, entry := range entries {` (range 355) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B14 | `if !entryOK \|\| !stopOK \|\| !targetOK {` (if 366) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B15 | `if readCtx == nil \|\| readCtx.Err() != nil \|\| now.IsZero() \|\| now.After(accountAuthority.FreshUntil()) {` (if 375) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |
| B16 | `if versionErr != nil {` (if 379) | — | 분기 본문 참조 | `TestA136QFinalCarriesTheOtherMarketOfThePairOnly` · 기존 첫 다리 조립 시험 |

Exact AST return positions: 290 · 297 · 303 · 307 · 312 · 319 · 327 · 329 · 334 · 340 · 345 · 350 · 367 · 376 · 380 · 382 · 386

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loader.fx.forMarket` | 짝 FX | — | AST |

## State mutations and fallbacks

- 없음(구성만).

## Safety conclusion

- Safe edit boundary: 칸 · 등록 하나 추가 — 기존 분기 무변경.
- High-risk impact: yes(전략 발급 경로)
