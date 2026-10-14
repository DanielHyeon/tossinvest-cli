# Function Logic Map: `handleSettingsLimits`

- Source: `internal/console/settings_limits.go` (283-363, 편집 뒤)
- Qualified: `Console.handleSettingsLimits`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 021a96457d9b0b17…, 편집 전 분기 수 9) — 현재 branches 11 · returns 7
- Risk scan: `risk-pattern-report.md`

**역할.** 고급 폼 저장.

**a136 편집.** `max_daily_buy_notional` 파싱(빈 칸 = 0 미설정, 숫자 아님 → 거부) → `DailyBuy` 실음, 저장 뒤 문장에 결과.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c.opts.Limits == nil {` (if 284) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B2 | `for _, target := range targets {` (range 301) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B3 | `if raw == "" {` (if 303) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B4 | `if err != nil {` (if 312) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B5 | `if raw := strings.TrimSpace(r.PostFormValue("max_daily_buy_notional")); raw != "" {` (if 320) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B6 | `if err != nil {` (if 322) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B7 | `if err := next.Validate(); err != nil {` (if 330) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B8 | `if _, err := config.GuardianCeiling(next.Currency); err != nil {` (if 339) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B9 | `if violations := next.CeilingViolations(); len(violations) > 0 {` (if 347) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B10 | `if err := c.opts.Limits.Save(next); err != nil {` (if 352) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |
| B11 | `if id := config.MatchingGuardianTier(next); id != "" {` (if 358) | — | 분기 본문 참조 | `TestA136TheAdvancedFormSavesTheDailyBuyCeiling` · `TestTheAdvancedFormSavesIndividualValues` |

Exact AST return positions: 287 · 314 · 324 · 332 · 342 · 350 · 354

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 설정 파일/audit(해당 시) — 기존과 같음.

## Safety conclusion

- Safe edit boundary: 칸 · 문장 하나 추가 — 기존 분기 무변경.
- High-risk impact: yes(한도 설정 쓰기 입구)
