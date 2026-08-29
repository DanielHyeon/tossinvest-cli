# Function Logic Map: `ExitSnapshotQuarantine.NeedsReJudgement`

- Source: `internal/journal/exit_snapshot.go` (lines 770–773)
- AST evidence: `ast.json` (`source_sha256: 0e376d1ca4f6e29b27308c540088d35ad9725304b6fc5cff20c8c7eed9780524`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

이 격리가 재판정 한 번을 벌었는지 보고한다. 두 조건이 모두 필요하다:
사유가 `ambiguous_recovery`여야 하고 (선택기에 관한 사실이므로 선택기가 만든 사유에만 적용된다),
각인된 revision이 지금 빌드의 것보다 *엄격히 낮아야* 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `q.Reason` | 격리 사유 문자열 | `exit_snapshot_quarantines.reason` | `ambiguous_recovery`가 아니면 false — 운영자만 풀 수 있는 행을 자동 해제하지 않는다 |
| `q.SelectorRevision` | 0(미각인) 이상 | `selector_revision` 열, v30 마이그레이션에서 nullable 추가 | 0은 stamp 이전 행 — 어떤 revision보다 낮으므로 backfill 없이 한 번을 번다 |
| `exitpolicy.RecoverySelectorRevision` | 빌드 상수 | `exitpolicy` 패키지 | 더 *높은* revision이 각인된 행은 false — 혼합 revision 프로세스 간 무한 재작성 방지 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | 분기 없음 | 단일 반환식 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | 호출 없음. 두 비교 | — | AST `calls` 없음 |

## State mutations and fallbacks

- 없음. 순수 조회.

## Safety conclusion

- Safe edit boundary: 두 비교. `<`를 `!=`로 바꾸면 여러 프로세스가 서로의 행을 매 사이클 재작성하며 critical 알림을 재무장한다.
- High-risk impact: yes — 격리 해제 자격 판정.
