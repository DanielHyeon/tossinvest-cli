# Function Logic Map: `Journal.CurrentOperatingMode`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 전**, :553–562, 분기 1 · 반환 2 · 호출 4, source_sha256 `ed35ea67c26b…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): 「현재 모드」를 rowid 내림차순 하나로(C4 · C5 — `modeLatestOrder` 상수). 스냅숏에 rowid(순번)를 싣는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if account == ""` (:555) | — | — | `TestCurrentOperatingModeNeedsAnAccount` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `currentModeFromRow`, `fmt.Errorf`, `j.db.QueryRowContext`, `strings.TrimSpace` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: High-risk — fail-open 방향 제거(벽시계 역행).
