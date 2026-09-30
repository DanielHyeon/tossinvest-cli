# Function Logic Map: `canonicalSnapshotContext`

- Source: `internal/exitpolicy/snapshot.go` (`259`–`270`)
- Qualified: `canonicalSnapshotContext`
- AST evidence: `ast.json` (`source_sha256` 5aaed0f721c2931c…) — base `b30318d6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 2 · 반환 3

**편집.** 편집하지 않는다 — M1: 잔여 수량은 양수여야 하고 `RatString` 정규형으로 넘어간다.

**역할.** 스냅숏 문맥을 정규화한다. 잔여 수량은 `positive("remaining quantity", …)`(`:263`) — 0 · 음수 · 비수치는 오류(B2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx.RemainingQuantity` | 포지션 수량(원장 투영) | `ConvergeQuantities` 가 쓴 계좌 값 등 | 0/음수/비수치 → 오류 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 base `b30318d6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`). engine 패키지 실행은 `-trimpath` 로 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다 — 커버리지 프로파일은 그대로 쓰인다(두 시험은 이 함수들과 무관한 AST 핀).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:260` `if strings.TrimSpace(ctx.PositionID) == "" \|\| ctx.PositionGeneration < 0 \|\| strings.TrimSpace(ctx.ObservationID) == "" {` | 아니오 |
| B2 | if | `:264` `if err != nil {` | 아니오 |

Exact AST return positions: `261:3`, `265:3`, `268:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(8).

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `positive` | `:263` | 양수 강제 | 순수 |
| `quantity.RatString` | `:267` | 정규형 | 순수 |

## State mutations and fallbacks

없음.

## Safety conclusion

- 포지션 수량의 철자(원장에 어떤 모양으로 있든)는 여기서 유리수로 파싱되고 `RatString` 으로 바뀐다 — 철자가 0 판정으로 새지 않는다.
