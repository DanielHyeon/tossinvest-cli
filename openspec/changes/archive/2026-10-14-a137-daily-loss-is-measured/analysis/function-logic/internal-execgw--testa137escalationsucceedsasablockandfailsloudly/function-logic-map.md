# Function Logic Map: `TestA137EscalationSucceedsAsABlockAndFailsLoudly`

- Source: `internal/execgw/a137_daily_loss_internal_test.go` (328-346, base 리비전)
- Qualified: `TestA137EscalationSucceedsAsABlockAndFailsLoudly`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` bca3eba9aebc69d6…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` bca3eba9aebc69d6…, 편집 전 분기수 4) — 현재 branches 4 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** 강화 입구의 성공(표지) · 실패(시끄러움) 시험.

**a137 편집.** 본문 무편집 — 바로 뒤에 새 시험 둘을 덧붙인 hunk 가 이 함수 끝 줄에 닿음. 분기 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| g · j | 시험 Guardian · 원장 | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `t.Fatal(err)` (if 331) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137EscalationSucceeds` |
| B2 | `err = gw2.withStrategyEntryGateAuthority(ctx, expected2, "kr", "005930", func() error { return nil })` (if 334) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137EscalationSucceeds` |
| B3 | `t.Fatalf("fenced = %+v, want a plain fenced rejection", fenced)` (if 339) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137EscalationSucceeds` |
| B4 | `// 강화 입구: 성공하면 계좌 차단, 실패하면 차단이 아닌 오류(조용히 삼켜지지 않음).` (if 343) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137EscalationSucceeds` |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EscalateDailyLossLimit` | 강화 | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 시험.
- High-risk impact: no — 시험.
