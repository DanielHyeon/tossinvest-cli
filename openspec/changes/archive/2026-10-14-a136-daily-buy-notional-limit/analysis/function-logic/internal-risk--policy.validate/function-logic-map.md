# Function Logic Map: `Validate`

- Source: `internal/risk/input.go` (257-309, 편집 뒤)
- Qualified: `Policy.Validate`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 3ef6826797400f67…, 편집 전 분기 수 11) — 현재 branches 11 · returns 11
- Risk scan: `risk-pattern-report.md`

**역할.** 체인 · Guardian 생성 전 정책 사용 가능성 검사(전부 fail-closed).

**a136 편집.** 돈 한도 루프 목록에 `{"maximum daily buy notional", p.MaxDailyBuyNotional}` 한 항목 추가 → 없음 · 0 · 음수 · 다른 통화 거부(USD 무값 기동 거부 D5 의 실제 지점). FLM 전에 편집됨(P1-3 기록).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p` 의 돈 한도 | 양수 · 같은 통화 | 정책 | 오류 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 259) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B2 | `for _, m := range []struct {` (range 262) | — | 분기 본문 참조 | `TestA136PolicyValidationRefusesAnUnusableDailyBuyCeiling` · contract_test mixed currencies |
| B3 | `if _, err := requirePositiveMoney(m.label, m.value, currency); err != nil {` (if 271) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B4 | `if err != nil {` (if 277) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B5 | `if quantity.Sign() <= 0 {` (if 280) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B6 | `if err != nil {` (if 286) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B7 | `if ratio.Sign() <= 0 \|\| ratio.Cmp(big.NewRat(1, 1)) > 0 {` (if 289) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B8 | `if err != nil {` (if 294) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B9 | `if minRR.Sign() <= 0 {` (if 297) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B10 | `if p.Reentry.MaxEntriesPerSymbol < 1 {` (if 301) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |
| B11 | `if p.Reentry.Cooldown < 0 {` (if 305) | — | 분기 본문 참조 | `go test ./internal/risk -run 'Validate / Policy / TestA136'` |

Exact AST return positions: 260 · 272 · 278 · 281 · 287 · 290 · 295 · 298 · 302 · 306 · 308

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `requirePositiveMoney` | 양수 · 통화 | 오류 | AST |
| `parseWholeNumber` · `parseDecimal` | 수 해석 | 오류 | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 루프 항목 하나 — 분기 수 무변경.
- High-risk impact: yes(정책 검증 — 더 엄격해지는 방향만)
