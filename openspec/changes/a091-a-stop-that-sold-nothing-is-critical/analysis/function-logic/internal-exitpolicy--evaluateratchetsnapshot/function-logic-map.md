# Function Logic Map: `EvaluateRatchetSnapshot`

- Source: `internal/exitpolicy/snapshot.go` (`176`–`239`)
- Qualified: `EvaluateRatchetSnapshot`
- AST evidence: `ast.json` (`source_sha256` 5aaed0f721c2931c…) — 편집 뒤 `3ec1efd2` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 10 · 반환 8

**편집.** 편집하지 않는다 — M1: 래칫 스냅숏의 `orderable = projected != "0"`(`:208`).

**역할.** 래칫 판정의 스냅숏. 사다리와 같은 투영 · 주문 가능성 규칙.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.Context.RemainingQuantity` | 양수 | canonicalSnapshotContext | — |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `3ec1efd2` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:178` `if in.Input.Config != nil {` | 예 |
| B2 | if | `:182` `if err != nil {` | 아니오 |
| B3 | if | `:186` `if err != nil {` | 아니오 |
| B4 | if | `:190` `if err != nil {` | 아니오 |
| B5 | if | `:194` `if err != nil {` | 아니오 |
| B6 | if | `:203` `if action.Orderable() {` | 예 |
| B7 | if | `:205` `if err != nil {` | 아니오 |
| B8 | if | `:212` `if err != nil {` | 아니오 |
| B9 | if | `:216` `if !previousLevel.Valid() {` | 예 |
| B10 | if | `:218` `if err != nil {` | 아니오 |

Exact AST return positions: `183:3`, `187:3`, `191:3`, `195:3`, `206:4`, `213:3`, `219:4`, `238:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(13).

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `ProjectWholeShares` | `:204` | 투영 | 순수 |

## State mutations and fallbacks

없음.

## Safety conclusion

- 0 투영은 `orderable=false`.
