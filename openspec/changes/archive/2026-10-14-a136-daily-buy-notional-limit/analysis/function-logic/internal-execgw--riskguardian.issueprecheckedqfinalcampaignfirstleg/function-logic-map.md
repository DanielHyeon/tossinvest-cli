# Function Logic Map: `IssuePrecheckedQFinalCampaignFirstLeg`

- Source: `internal/execgw/riskguardian_first_leg.go` (117-213, 편집 뒤)
- Qualified: `RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 45de8eaa98ad293b…, 편집 전 분기 수 12) — 현재 branches 13 · returns 14
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 첫 다리 발급 문(생산 전략 경로).

**a136 편집.** collect 클로저의 ReserveRequest 에 `DailyBuyLimit: g.dailyBuyLimit()` 한 칸.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `precheck` | 봉인 완료 | PrecheckQFinalCampaignFirstLeg | 거부 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil \|\| precheck.entry.decision.QFinal == 0 \|\| precheck.entry.request.Collect == nil \|\|` (if 118) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B2 | `if precheck.issuedAt.IsZero() \|\| now.Before(precheck.issuedAt) \|\| !now.Before(precheck.issuedAt.Add(g.ttl)) {` (if 124) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B3 | `if (err != nil && !legacyTestIdentity) \|\| (!legacyTestIdentity && !sameAccountBaseFX(precheck.entry.accountFX, accountFX)) {` (if 131) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B4 | `if err != nil {` (if 136) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B5 | `if !legacyTestIdentity {` (if 140) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B6 | `if err != nil {` (if 142) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B7 | `if err != nil {` (if 160) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B8 | `if !legacyTestIdentity && !sameQFinalFXAuthority(accountFX, policy) {` (if 164) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B9 | `if err != nil {` (if 176) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B10 | `if !verdict.Allowed {` (if 183) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B11 | `if err != nil {` (if 187) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B12 | `if err != nil {` (if 191) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |
| B13 | `if err != nil {` (if 209) | — | 분기 본문 참조 | `TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling`(tossos_testseams) · 기존 발급 시험군 |

Exact AST return positions: 120 · 125 · 132 · 137 · 143 · 161 · 165 · 177 · 184 · 188 · 192 · 194 · 210 · 212

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.journal.RecordQFinalCampaignFirstLegWithRecollection` 계열 | 원자 발급 | issuanceRefusal | AST |

## State mutations and fallbacks

- journal tx.

## Safety conclusion

- Safe edit boundary: 요청 칸 하나.
- High-risk impact: yes
