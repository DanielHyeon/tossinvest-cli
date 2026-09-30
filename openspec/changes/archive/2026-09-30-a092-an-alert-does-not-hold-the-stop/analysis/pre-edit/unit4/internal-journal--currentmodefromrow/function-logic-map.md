# Function Logic Map: `currentModeFromRow`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 전**, :655–672, 분기 3 · 반환 3 · 호출 2, source_sha256 `ed35ea67c26b…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): 스냅숏에 순번을 옮긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` (:657) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B2 | `case errors.Is(err, errNoOperatingMode):` (:658) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B3 | `case err != nil:` (:661) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.Is`, `scanOperatingMode` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: Low.
