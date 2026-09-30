# Function Logic Map: `ExitLineSnapshot.ExecutableProposal`

- Source: `internal/exitpolicy/snapshot.go` (`54`–`59`)
- Qualified: `ExitLineSnapshot.ExecutableProposal`
- AST evidence: `ast.json` (`source_sha256` 5aaed0f721c2931c…) — base `b30318d6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 1 · 반환 2

**편집.** 편집하지 않는다 — M1 증거.

**역할.** 주문 가능하지 않거나 투영 수량이 `"0"` 이면 빈 제안(B1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.ProjectedQuantity` | `big.Int.String` 정규형 | `ProjectWholeShares` | `"0"` → 빈 제안 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 base `b30318d6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`). engine 패키지 실행은 `-trimpath` 로 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다 — 커버리지 프로파일은 그대로 쓰인다(두 시험은 이 함수들과 무관한 AST 핀).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:55` `if !s.Orderable \|\| s.ProjectedQuantity == "0" {` | 예 |

Exact AST return positions: `56:3`, `58:2`

## Calls and live bindings

호출 없음(`ast.json` `calls` 0).

## State mutations and fallbacks

없음.

## Safety conclusion

- 문자열 `"0"` 비교이지만 입력이 `big.Int.String` 이라 0 의 철자는 `"0"` 하나뿐이다(ProjectWholeShares 번들).
