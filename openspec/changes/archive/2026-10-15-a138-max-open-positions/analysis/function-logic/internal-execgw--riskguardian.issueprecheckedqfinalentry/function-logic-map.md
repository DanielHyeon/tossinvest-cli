# Function Logic Map: `IssuePrecheckedQFinalEntry`

- Source: `internal/execgw/riskguardian_qfinal.go` (187-275)
- Qualified: `RiskGuardian.IssuePrecheckedQFinalEntry`
- AST evidence: `ast.json`(현재) — 현재 AST branches 12 · returns 13
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 0e68a5c27cad92a5…) — 편집 전 갈래 수 12
- Risk scan: `risk-pattern-report.md`

**역할.** 진입 발급 문 2 — q_final precheck 뒤 원자 발급.

**a138 편집.** ReserveRequest 에 상한 한 칸(분기 무변경).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil \|\| precheck.decision.QFinal == 0 \|\| precheck.request.Collect == nil \|\| precheck.admission.TransactionID == "" {` (if 188) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B2 | `if (err != nil && !legacyTestIdentity) \|\| (!legacyTestIdentity && !sameAccountBaseFX(precheck.accountFX, accountFX)) {` (if 194) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B3 | `if err != nil {` (if 198) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B4 | `if !legacyTestIdentity {` (if 202) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B5 | `if err != nil {` (if 204) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B6 | `if err != nil {` (if 226) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B7 | `if !legacyTestIdentity && !sameQFinalFXAuthority(accountFX, policy) {` (if 229) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B8 | `if err != nil {` (if 240) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B9 | `if !verdict.Allowed {` (if 244) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B10 | `if err != nil {` (if 248) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B11 | `if err != nil {` (if 252) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |
| B12 | `if err != nil {` (if 267) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run 'QFinal|A138'` · `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` |

Exact AST return positions: 189 · 195 · 199 · 205 · 227 · 230 · 241 · 245 · 249 · 253 · 255 · 268 · 270

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 22종 — 대표: `UTC`, `chainRefusal`, `encodeQFinalLimits`, `exposureUsage`, `g.clk.Now`, `g.dailyBuyLimit`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
