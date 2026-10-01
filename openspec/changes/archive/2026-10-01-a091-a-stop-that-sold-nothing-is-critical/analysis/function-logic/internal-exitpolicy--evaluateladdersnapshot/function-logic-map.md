# Function Logic Map: `EvaluateLadderSnapshot`

- Source: `internal/exitpolicy/snapshot.go` (`96`–`167`)
- Qualified: `EvaluateLadderSnapshot`
- AST evidence: `ast.json` (`source_sha256` 5aaed0f721c2931c…) — 편집 뒤 `540aebe6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 11 · 반환 7

**편집.** 편집하지 않는다 — M1: 사다리 스냅숏의 `orderable = projected != "0"`(`:143`).

**역할.** 사다리 판정의 스냅숏. 주문 가능 액션이면 `ProjectWholeShares` 로 투영하고 0 이면 state-only.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.Context.RemainingQuantity` | 양수 | canonicalSnapshotContext | — |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `540aebe6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:98` `if err != nil {` | 예 |
| B2 | if | `:102` `if err != nil {` | 아니오 |
| B3 | if | `:111` `if strings.TrimSpace(eval.State.PolicyID) == "" {` | 아니오 |
| B4 | if | `:114` `if strings.TrimSpace(eval.State.PolicyVersion) == "" {` | 예 |
| B5 | if | `:117` `if strings.TrimSpace(eval.State.PolicyDigest) == "" {` | 예 |
| B6 | if | `:121` `if err != nil {` | 예 |
| B7 | if | `:126` `if err != nil {` | 아니오 |
| B8 | if | `:132` `if !transition.Proposal.Zero() {` | 예 |
| B9 | if | `:138` `if action.Orderable() {` | 예 |
| B10 | if | `:140` `if err != nil {` | 아니오 |
| B11 | if | `:147` `if err != nil {` | 아니오 |

Exact AST return positions: `99:3`, `103:3`, `122:3`, `127:3`, `141:4`, `148:3`, `166:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(15).

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `ProjectWholeShares` | `:139` | 투영 | 순수 |

## State mutations and fallbacks

없음.

## Safety conclusion

- 0 투영은 `orderable=false` — 제출 경로에 닿지 않는다.
