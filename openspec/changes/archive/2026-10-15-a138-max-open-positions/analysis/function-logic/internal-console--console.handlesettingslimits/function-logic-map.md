# Function Logic Map: `handleSettingsLimits`

- Source: `internal/console/settings_limits.go` (307-417)
- Qualified: `Console.handleSettingsLimits`
- AST evidence: `ast.json`(현재) — 현재 AST branches 16 · returns 8
- 편집 전 기준: `ast-base.json`(base, `source_sha256` f3c64ef418118fb8…) — 편집 전 갈래 수 12
- Risk scan: `risk-pattern-report.md`

**역할.** 고급 폼 저장 핸들러.

**a138 편집.** `max_open_positions` 칸이 폼에 **있을 때만** 값을 실음(빈 칸 = 0), 저장 뒤 문장에 한 마디 — 새 분기 셋.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if c.opts.Limits == nil {` (if 308) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B2 | `for _, target := range targets {` (range 325) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B3 | `if raw == "" {` (if 327) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B4 | `if err != nil {` (if 336) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B5 | `if raw := strings.TrimSpace(r.PostFormValue("max_daily_buy_notional")); raw != "" {` (if 344) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B6 | `if err != nil {` (if 346) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B7 | `if current, err := c.opts.Limits.Load(); err == nil &&` (if 356) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B8 | `if _, present := r.PostForm["max_open_positions"]; present {` (if 366) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B9 | `if raw := strings.TrimSpace(r.PostFormValue("max_open_positions")); raw != "" {` (if 368) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B10 | `if err != nil {` (if 370) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B11 | `if err := next.Validate(); err != nil {` (if 380) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B12 | `if _, err := config.GuardianCeiling(next.Currency); err != nil {` (if 389) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B13 | `if violations := next.CeilingViolations(); len(violations) > 0 {` (if 397) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B14 | `if err := c.opts.Limits.Save(next); err != nil {` (if 402) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B15 | `if id := config.MatchingGuardianTier(next); id != "" {` (if 408) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |
| B16 | `if openPositions >= 0 {` (if 412) | — | 분기 본문 참조 | `go test ./internal/console -run 'A136|A138|Limit'` · `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` |

Exact AST return positions: 311 · 338 · 348 · 372 · 382 · 392 · 400 · 404

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 24종 — 대표: `c.engineRunning`, `c.opts.Limits.Load`, `c.opts.Limits.Save`, `c.redirectSettings`, `c.refuse`, `config.GuardianCeiling`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
