# Function Logic Map: `ConfirmedFloorQuantity`

- Source: `internal/riskcalc/confirmed_floor.go` (`129`–`195`)
- Qualified: `ConfirmedFloorQuantity`
- AST evidence: `ast.json` (`source_sha256` 05cd90fe5dece3e1…) — base `b30318d6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 11 · 반환 10

**편집.** 편집하지 않는다 — M1: `floor.Quantity` 의 철자 출처.

**역할.** 확정 하한 = max(0, min(보유, 매도가능) − 로컬 미체결 매도). 스냅숏 부재 · 낡음은 `zeroFloor`(리터럴 `"0"`, B4 · B6), 계산 경로는 `MaxDecimal("0", …)` → `CanonicalDecimal`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.Holdings` · `in.Sellable` | 스냅숏(나이 ≤ 한계) | `reconcileFloor` 의 두 브로커 읽기 | 부재 · 낡음 → 0 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 base `b30318d6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`). engine 패키지 실행은 `-trimpath` 로 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다 — 커버리지 프로파일은 그대로 쓰인다(두 시험은 이 함수들과 무관한 AST 핀).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:130` `if in.Now.IsZero() {` | 예 |
| B2 | if | `:142` `if err != nil {` | 예 |
| B3 | if | `:147` `if err != nil {` | 예 |
| B4 | if | `:150` `if !ok {` | 예 |
| B5 | if | `:154` `if err != nil {` | 예 |
| B6 | if | `:157` `if !ok {` | 예 |
| B7 | if | `:165` `if err != nil {` | 아니오 |
| B8 | if | `:169` `if base == sellable && sellable != holdings {` | 예 |
| B9 | if | `:174` `if err != nil {` | 아니오 |
| B10 | if | `:181` `if err != nil {` | 아니오 |
| B11 | if | `:184` `if floor != base {` | 예 |

Exact AST return positions: `131:3`, `143:3`, `148:3`, `151:3`, `155:3`, `158:3`, `166:3`, `175:3`, `182:3`, `188:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(19).

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `MaxDecimal` | `:180` | 0 하한 | 순수 — 정규 철자 반환(`decimal.go:92-101`) |

## State mutations and fallbacks

없음.

## Safety conclusion

- 반환 `Quantity` 의 0 은 `"0"` 한 철자다. 첫 리뷰 M1 의 전제(정규형)는 여기서 성립한다.
