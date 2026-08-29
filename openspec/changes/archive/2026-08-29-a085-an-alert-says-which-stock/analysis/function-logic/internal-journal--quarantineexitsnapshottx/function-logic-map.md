# Function Logic Map: `quarantineExitSnapshotTx`

- Source: `internal/journal/exit_snapshot.go` (lines 540–567)
- AST evidence: `ast.json` (`source_sha256: 0e376d1ca4f6e29b27308c540088d35ad9725304b6fc5cff20c8c7eed9780524`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

격리 행의 유일한 writer. 개정 2가 여기서 close-and-reopen을 없앴다. 개정 1은
활성 행을 `SELECTOR_REVISED` 근거로 닫고 새 행을 열어 재시도를 revision당 1회로 유지하려
했지만, 재시도 소비는 이제 재판정을 결정하는 곳(`judge`)에서 각인으로 처리하므로 그 대체는
도달 불가였고, 복구 선택기가 판정한 적 없는 행에 "이 세대는 재판정되었다"를 쓰는 것이었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id`, `generation` | 포지션과 세대 | 호출자 | 활성 행이 있으면 그대로 반환 — 멱등 |
| `reason`, `evidence` | 비어 있지 않음 | 호출자가 검증 | — |
| `exitpolicy.RecoverySelectorRevision` | 빌드 상수 | `exitpolicy` | 새 행은 항상 지금 revision으로 각인된다 — 자기가 쓴 행을 스스로 재판정하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (542) `if` — if active, ok, err := activeExitSnapshotQuarantineTx(ctx, tx, id, generation); err != nil { | 본문 참조 | 아래 Branch Test Map |
| B2 | (544) `else` — } else if ok { | 본문 참조 | 아래 Branch Test Map |
| B3 | (544) `if` — } else if ok { | 본문 참조 | 아래 Branch Test Map |
| B4 | (556) `if` — if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(quarantine_version),0)+1 FROM exit_snapshot_quarantines | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `activeExitSnapshotQuarantineTx` | 활성 행 조회 | 오류 전파 | AST `calls` L542 |
| `tx.QueryRowContext` (max version) | 다음 버전 번호 | 오류 전파 | AST `calls` L556 |
| `tx.ExecContext` (INSERT) | 행 기록 | 오류 반환 | AST `calls` L563 |

## State mutations and fallbacks

- `exit_snapshot_quarantines` INSERT — `selector_revision`을 지금 빌드로 각인.
- 활성 행이 있으면 아무것도 쓰지 않는다 (멱등).

## Safety conclusion

- Safe edit boundary: 멱등 반환. close-and-reopen을 되살리면 일어나지 않은 재판정의 근거가 원장에 남는다.
- High-risk impact: yes — 원장 쓰기 경로.
