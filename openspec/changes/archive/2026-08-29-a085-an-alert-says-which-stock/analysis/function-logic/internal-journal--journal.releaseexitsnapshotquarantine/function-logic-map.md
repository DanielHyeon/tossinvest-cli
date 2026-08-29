# Function Logic Map: `Journal.ReleaseExitSnapshotQuarantine`

- Source: `internal/journal/exit_snapshot.go` (lines 893–920)
- AST evidence: `ast.json` (`source_sha256: 0e376d1ca4f6e29b27308c540088d35ad9725304b6fc5cff20c8c7eed9780524`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 운영자 해제 경로.

## What it does

운영자(a079) 해제. a084는 허용 kind 목록에 `SELECTOR_REVISED`를 더할 뿐, 검증·낙관적 잠금·stale 판정은 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| kind | HUMAN_REPAIR / AUTHORITATIVE_RECONCILE / (a084) SELECTOR_REVISED | 호출자 | 그 밖은 `ErrInvalidRequest` |
| expectedVersion | 양수, 활성 행과 일치 | 호출자 | 불일치는 `ErrExitSnapshotReleaseStale` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (849) `if` — if strings.TrimSpace(positionID) == "" || generation < 0 || expectedVersion <=… | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (852) `if` — if kind != QuarantineReleaseHumanRepair && kind != QuarantineReleaseAuthoritat… | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (860) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (864) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (867) `if` — if changed != 1 | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | (895) kind = strings.TrimSpace(kind) | 호출부 계약 유지 | AST `calls` |
| `fmt.Errorf` | (897) return fmt.Errorf("%w: release needs position, generation, positive version, and evidence", ErrInvalidRequest) | 호출부 계약 유지 | AST `calls` |
| `j.db.ExecContext` | (905) result, err := j.db.ExecContext(ctx, `UPDATE exit_snapshot_quarantines | 호출부 계약 유지 | AST `calls` |
| `j.nowString` | (908) j.nowString(), kind, strings.TrimSpace(evidence), strings.TrimSpace(positionID), generation, expectedVersion) | 호출부 계약 유지 | AST `calls` |
| `result.RowsAffected` | (912) changed, err := result.RowsAffected() | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 활성 행 1건 UPDATE.

## Safety conclusion

- Safe edit boundary: kind 검증 조건 하나.
- High-risk impact: no — 새 kind를 받아들일 뿐 해제 조건을 넓히지 않는다. 여전히 정확히 한 활성 행만 닫는다.
