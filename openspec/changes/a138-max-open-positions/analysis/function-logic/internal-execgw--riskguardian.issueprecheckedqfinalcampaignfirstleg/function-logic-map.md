# Function Logic Map: `IssuePrecheckedQFinalCampaignFirstLeg`

- Source: `internal/execgw/riskguardian_first_leg.go` (117-214)
- Qualified: `RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg`
- AST evidence: `ast.json`(현재) — 현재 AST branches 13 · returns 14
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 2afe54a55664d8af…) — 편집 전 갈래 수 13
- Risk scan: `risk-pattern-report.md`

**역할.** 진입 발급 문 3 — 전략 첫 다리.

**a138 편집.** ReserveRequest 에 상한 한 칸(분기 무변경).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil \|\| precheck.entry.decision.QFinal == 0 \|\| precheck.entry.request.Collect == nil \|\|` (if 118) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B2 | `if precheck.issuedAt.IsZero() \|\| now.Before(precheck.issuedAt) \|\| !now.Before(precheck.issuedAt.Add(g.ttl)) {` (if 124) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B3 | `if (err != nil && !legacyTestIdentity) \|\| (!legacyTestIdentity && !sameAccountBaseFX(precheck.entry.accountFX, accountFX)) {` (if 131) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B4 | `if err != nil {` (if 136) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B5 | `if !legacyTestIdentity {` (if 140) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B6 | `if err != nil {` (if 142) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B7 | `if err != nil {` (if 160) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B8 | `if !legacyTestIdentity && !sameQFinalFXAuthority(accountFX, policy) {` (if 164) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B9 | `if err != nil {` (if 176) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B10 | `if !verdict.Allowed {` (if 183) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B11 | `if err != nil {` (if 187) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B12 | `if err != nil {` (if 191) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |
| B13 | `if err != nil {` (if 210) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'FirstLeg|A138'` · `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` |

Exact AST return positions: 120 · 125 · 132 · 137 · 143 · 161 · 165 · 177 · 184 · 188 · 192 · 194 · 211 · 213

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 25종 — 대표: `UTC`, `chainRefusal`, `cloneGuardianWeeklyBinding`, `encodeQFinalLimits`, `exposureUsage`, `fmt.Errorf`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
