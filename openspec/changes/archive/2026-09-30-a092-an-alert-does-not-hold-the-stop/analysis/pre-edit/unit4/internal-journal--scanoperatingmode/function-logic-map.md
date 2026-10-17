# Function Logic Map: `scanOperatingMode`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 전**, :693–712, 분기 3 · 반환 4 · 호출 5, source_sha256 `ed35ea67c26b…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): `rowid` 를 함께 읽는다(`operatingModeSelect` 에 열 추가).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := row.Scan(&record.ID, &record.AccountRef, &record.Mode,` (:698) | — | — | (미실행) |
| B2 | `if errors.Is(err, sql.ErrNoRows)` (:700) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B3 | `if err != nil` (:706) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.Is`, `fmt.Errorf`, `parseJournalTime`, `row.Scan` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: Low.
