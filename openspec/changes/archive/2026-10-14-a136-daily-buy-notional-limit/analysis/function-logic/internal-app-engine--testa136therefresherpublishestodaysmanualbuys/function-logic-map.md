# Function Logic Map: `TestA136TheRefresherPublishesTodaysManualBuys`

- Source: `internal/app/engine/a136_manual_buy_refresh_internal_test.go` (57-82, 편집 뒤)
- Qualified: `TestA136TheRefresherPublishesTodaysManualBuys`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 1f9c34248f1ebae8…, 편집 전 분기 수 3) — 현재 branches 3 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** a136 시험(생산 코드 아님).

**a136 편집.** 844a042c: 픽스처를 실제 OpenAPI 세부 상태 열거로 바꾸거나(OPEN/CLOSED 그룹 라벨 제거) 같은 파일 뒤에 시험을 덧붙임 — 단언 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if !ok \|\| got.Err != nil \|\| got.Day != "2026-10-14" \|\| !got.AsOf.Equal(now) {` (if 72) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` |
| B2 | `if got.ByCurrency["KRW"] != "2500" \|\| got.ByCurrency["USD"] != "3" {` (if 75) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` |
| B3 | `if len(pages.calls) != 3 \|\| pages.calls[0].Status != "OPEN" \|\| pages.calls[1].From != "2026-10-13" \|\|` (if 78) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 픽스처 문자열.
- High-risk impact: no(시험 전용)
