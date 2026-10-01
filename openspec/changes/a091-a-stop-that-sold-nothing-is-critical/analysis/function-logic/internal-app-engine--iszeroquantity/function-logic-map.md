# Function Logic Map: `isZeroQuantity`

- Source: `internal/app/engine/exitloop.go` (`1889`–`1896`)
- Qualified: `isZeroQuantity`
- AST evidence: `ast.json` (`source_sha256` 0733bd8641ed8c41…) — 편집 뒤 `540aebe6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 1 · 반환 2

**편집.** 편집하지 않는다 — M1(0주 판정의 모양) 증거.

**역할.** 0 판정. 공백 제거 뒤 빈 문자열이면 0(B1), 아니면 **수치 비교** `CompareDecimal(q, "0") <= 0`, 파싱 실패도 0.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `q` | 십진 문자열 | `applyFloor` 반환 · 포지션 수량 | 파싱 실패 → 0(fail-closed) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `540aebe6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1891` `if q == "" {` | 아니오 |

Exact AST return positions: `1892:3`, `1895:2`

## Calls and live bindings

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `strings.TrimSpace` | `:1890` | — | 순수 |
| `riskcalc.CompareDecimal` | `:1894` | 수치 비교 | 순수 |

## State mutations and fallbacks

없음.

## Safety conclusion

- 첫 리뷰 M1 이 적은 「정확히 `"0"` 문자열 비교」는 base b30318d6 에서 **더 이상 참이 아니다** — `"0.0"` · `" 0"` 도 0 이다. 0주 판정은 철자에 의존하지 않는다.
