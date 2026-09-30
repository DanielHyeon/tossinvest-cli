# Function Logic Map: `ProjectWholeShares`

- Source: `internal/exitpolicy/snapshot.go` (`73`–`94`)
- Qualified: `ProjectWholeShares`
- AST evidence: `ast.json` (`source_sha256` 5aaed0f721c2931c…) — 편집 뒤 `3ec1efd2` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 5 · 반환 3

**편집.** 편집하지 않는다 — M1: 투영 수량의 철자 출처.

**역할.** 잔여 × 비율(상한 1)을 유리수로 곱해 내림한 **정수**를 `units.String()`(`:93`)로 낸다. 음수면 0(B5).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `remaining` | 양의 유리수 정규형 | `canonicalSnapshotContext` 의 `positive` + `RatString` | 파싱 실패 → 오류(B1) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `3ec1efd2` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:75` `if !ok {` | 아니오 |
| B2 | if | `:79` `if r := strings.TrimSpace(ratio); r != "" {` | 예 |
| B3 | if | `:81` `if !ok {` | 아니오 |
| B4 | if | `:85` `if share.Cmp(one) > 0 {` | 예 |
| B5 | if | `:90` `if units.Sign() < 0 {` | 아니오 |

Exact AST return positions: `76:3`, `82:4`, `93:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(21) — 근거는 반환 철자(`:93`) 하나다.

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `units.String` | `:93` | 정수 정규형 철자 | 순수 — 0 은 `"0"` |

## State mutations and fallbacks

없음.

## Safety conclusion

- 반환은 `big.Int` 의 10진 철자 — 선행 0 · 소수점 · 공백이 없다.
