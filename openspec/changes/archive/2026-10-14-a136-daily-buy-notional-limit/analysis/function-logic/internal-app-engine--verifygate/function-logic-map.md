# Function Logic Map: `verifyGate`

- Source: `internal/app/engine/interlock.go` (484-562, 편집 뒤)
- Qualified: `verifyGate`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` a2277d4bd3a76152…, 편집 전 분기 수 9) — 현재 branches 10 · returns 11
- Risk scan: `risk-pattern-report.md`

**역할.** 게이트 인터록 전체 검사, 첫 실패 정지.

**a136 편집.** 7번(단일 출처) 직후 `checkGuardianDailyBuyLimit(facts.guardian, status)` — 미설정(USD 무값)이면 ErrLimitsRequired, Guardian 미진술 · 불일치면 ErrGuardianLimitsMismatch.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `status` · `facts` | — | 기동 | 거부 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if facts.guardian == nil {` (if 486) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B2 | `if err := status.Limits.Validate(); err != nil {` (if 494) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B3 | `if err := checkTradingPolicy(facts.trading); err != nil {` (if 501) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B4 | `if err != nil {` (if 507) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B5 | `if accountRef == "" {` (if 519) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B6 | `if err := att.Verify(facts.now, accountRef, RequiredEndpoints()); err != nil {` (if 524) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B7 | `if err := checkGuardianLimits(facts.guardian, status.Limits); err != nil {` (if 530) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B8 | `if err := checkGuardianDailyBuyLimit(facts.guardian, *status); err != nil {` (if 534) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B9 | `if err := checkGateway(facts.gateway); err != nil {` (if 539) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |
| B10 | `if err := checkKeyCapableTransport(facts.transport); err != nil {` (if 542) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'Interlock / Gate / TestA136'` |

Exact AST return positions: 487 · 495 · 502 · 508 · 520 · 525 · 531 · 535 · 540 · 543 · 561

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkGuardianLimits` | 5필드 단일 출처 | 오류 | AST |
| `checkGuardianDailyBuyLimit`(새 leaf) | 하루 매수 상한 단일 출처 | 오류 | daily_buy_interlock.go |

## State mutations and fallbacks

- status.AttestationExpiresAt.

## Safety conclusion

- Safe edit boundary: 검사 한 단계 추가(거부 방향).
- High-risk impact: yes(인터록)
