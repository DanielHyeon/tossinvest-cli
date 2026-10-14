# Function Logic Map: `RiskGuardian.PrecheckQFinalEntry`

- Source: `internal/execgw/riskguardian_qfinal.go` (89-177, 편집 뒤)
- Qualified: `RiskGuardian.PrecheckQFinalEntry`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 201f9a1b04a9ecfd…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` cc8dcaf1c7a196af…, 편집 전 분기수 17) — 현재 branches 17 · returns 16
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 전략 첫 다리의 mutation-free precheck(q_final 계산 → 체인 1회). 살아 있는 유일한 진입 입구.

**a137 편집.** 체인 앞에서 `g.dailyLossFloor(request.Account, request.DailyLoss, now, qFinalManualBuyFX(request, market))` — 토큰(첫 다리 승인이 precheck 직전에 잼)과 요청의 봉인 FX(자기 시장 + 같은 파도 짝 FX, a136 PairedFXAuthority 재사용). 원장 · 브로커 호출 없음(계약 유지). B16 조건 식만 바뀜, 분기 수 불변(17).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `request.DailyLoss` | 이 Guardian 의 토큰 · 10초 안 · 같은 일 키 · 봉인 오류 없음 | `MeasureDailyLoss` | 어긋나면 측정 실패 → daily_loss 단 거절 |
| `request.PairedFXAuthority` | 같은 파도 다른 시장 봉인 권한 | 첫 다리 조립(a136) | 없으면 다른 통화 손실 측정 불가 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil {` (if 90) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B2 | `if request.ExpectedPolicyVersion != g.policyVersion \|\| request.ExpectedLimitsDigest != g.LimitsDigest() {` (if 93) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B3 | `if err != nil {` (if 99) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B4 | `if symbol == "" \|\| symbol != request.Symbol {` (if 103) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B5 | `if request.QCandidate == 0 {` (if 106) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B6 | `if admission.Owner.Key.AccountID != g.accountRef \|\| admission.Owner.Key.Market != bucketMarket \|\| admission.Owner.Key.Symbol != symbol {` (if 120) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B7 | `if policy.AccountCurrency != limitCurrency \|\| policy.QuoteCurrency != exactCurrency {` (if 124) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B8 | `if err != nil {` (if 128) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B9 | `if err != nil && !legacyTestIdentity {` (if 134) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B10 | `if !legacyTestIdentity && !sameQFinalFXAuthority(accountFX, policy) {` (if 137) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B11 | `if legacyTestIdentity {` (if 142) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B12 | `} else {` (else 144) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B13 | `if err != nil {` (if 147) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B14 | `if err != nil \|\| guardianCap == 0 {` (if 151) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B15 | `if decision.Refusal != nil {` (if 156) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |
| B16 | `if verdict := evaluateChainWithDailyLoss(in, measureErr); !verdict.Allowed {` (if 168) | — | 분기 본문 참조 | `TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne` · `TestA137AForeignLossConvertsOnlyThroughThePairedAuthority` · `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock`(fixture 토큰 경유) · 기존 q_final 시험(fixture 가 토큰을 실음) |
| B17 | `if _, verdict := risk.EntryExposureValue(in); !verdict.Allowed {` (if 171) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run QFinal` |

Exact AST return positions: 91 · 94 · 100 · 104 · 107 · 121 · 125 · 129 · 135 · 138 · 148 · 152 · 157 · 169 · 172 · 176

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.dailyLossFloor` | 바닥 규칙(a137) | 측정 실패 → 쓸 수 없는 값 | AST calls |
| `evaluateChainWithDailyLoss` | 체인 1회 | 거절 → chainRefusal(B16) | AST calls |

## State mutations and fallbacks

- 없음(precheck 는 mutation-free). 강화는 호출자(첫 다리 승인)가 함.

## Safety conclusion

- Safe edit boundary: 체인 입력의 일일 손실 칸.
- High-risk impact: yes — 진입을 더 막는 방향만.
