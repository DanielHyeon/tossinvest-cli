# Function Logic Map: `Journal.recordExitJudgementTx`

- Source: `internal/journal/exit_state.go` (lines 392–625)
- AST evidence: `ast.json` (`source_sha256: 99c96bdba4a08f7ccb5eb09597d93d3b230086d5507f5d9d4367be151c079ff2`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 판정 기록과 격리 생성이 같은 트랜잭션에 있다.

## What it does

한 판정을 기록한다. a084는 복구 선택 성공 분기에 `releaseReJudgedQuarantineTx` 한 줄을 더해,
재판정이 성공하면 그것을 허용한 격리가 같은 트랜잭션에서 닫히게 한다.

개정 3·4가 그 한 줄의 인자를 두 번 바꿨다. 개정 2는 재판정 여부를 활성 행의 *사유*로
추론했는데, 각인이 이 트랜잭션보다 먼저 일어나므로 행은 어느 쪽이든 현재 개정을 달고
있어 판별이 되지 않았다 — 지금 도는 선택기가 방금 쓴 격리가 다음 성공 판정에 닫혔다.
개정 3이 호출자의 사실(`ReJudging bool`)로 게이트했고, 개정 4가 그 사실을
`ReJudgingVersion int64`로 좁혔다: "재판정이 있었다"와 "*이* 행이 재판정되었다"는 다른
주장이고 해제를 정당화하는 것은 두 번째뿐이다. 각인과 이 트랜잭션 사이에 운영자 해제와
병행 관측의 새 행이 끼어들 수 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement | 관측·제안·snapshot | 호출자(exit loop) | invalid은 `ErrInvalidRequest` |
| current | 저장된 exit progress | 같은 트랜잭션의 SELECT | completed는 거부 |
| saved/recomputed | 복구 후보 두 개 | 저장 effective snapshot + 재계산 | 선택 실패는 격리 |
| 활성 격리 | 세대별 최대 1행 | 같은 트랜잭션 | `judgement.ReJudgingVersion`과 버전이 일치하고 사유가 `ambiguous_recovery`일 때만 해제한다. 각인(`selector_revision`)은 판별자가 아니다 |
| `judgement.ReJudgingVersion` | 0 = 재판정 아님, 그 외 = 재시도를 소비한 격리 버전 | 호출자(`ExitObserver.record`) | 0이면 해제 경로 전체가 no-op |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (380) `if` — if id == "" | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (383) `if` — if err := judgement.Provenance.validate(); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (386) `if` — if judgement.Provenance.zero() && strings.TrimSpace(judgement.ArmSuppressedRea… | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (389) `if` — if judgement.Proposal != nil | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (390) `if` — if err := validateProposal(*judgement.Proposal); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B6 | (393) `if` — if err := judgement.Proposal.Provenance.validate(); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B7 | (396) `if` — if judgement.Provenance.zero() != judgement.Proposal.Provenance.zero() || | 본문 참조 | — | 아래 Branch Test Map |
| B8 | (403) `if` — if !judgement.Provenance.zero() | 본문 참조 | — | 아래 Branch Test Map |
| B9 | (407) `if` — if err := validateJudgementSnapshot(id, judgement, candidate); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B10 | (415) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B11 | (421) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B12 | (424) `if` — if current.Completed | 본문 참조 | — | 아래 Branch Test Map |
| B13 | (428) `if` — if expectedLifecycle == 0 | 본문 참조 | — | 아래 Branch Test Map |
| B14 | (431) `if` — if expectedLifecycle != current.LifecycleGeneration | 본문 참조 | — | 아래 Branch Test Map |
| B15 | (440) `if` — if errors.Is(err, sql.ErrNoRows) | 본문 참조 | — | 아래 Branch Test Map |
| B16 | (442) `else` — } else if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B17 | (442) `if` — } else if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B18 | (445) `if` — if lifecycleStatus != positionpolicy.StatusManaged || lifecycleGeneration != e… | 본문 참조 | — | 아래 Branch Test Map |
| B19 | (449) `if` — if recomputed != nil && recomputed.Line.PositionGeneration != current.Position… | 본문 참조 | — | 아래 Branch Test Map |
| B20 | (453) `if` — if recomputed != nil | 본문 참조 | — | 아래 Branch Test Map |
| B21 | (457) `if` — if err == nil | 본문 참조 | — | 아래 Branch Test Map |
| B22 | (464) `if` — if !errors.Is(err, sql.ErrNoRows) | 본문 참조 | — | 아래 Branch Test Map |
| B23 | (472) `if` — if recomputed == nil | 본문 참조 | — | 아래 Branch Test Map |
| B24 | (473) `if` — if err := notBelow("high water", id, judgement.HighWater, current.HighWater); … | 본문 참조 | — | 아래 Branch Test Map |
| B25 | (476) `if` — if err := notBelow("baseline", id, judgement.Baseline, current.Baseline); err … | 본문 참조 | — | 아래 Branch Test Map |
| B26 | (481) `if` — if level == "" | 본문 참조 | — | 아래 Branch Test Map |
| B27 | (487) `if` — if current.Effective != nil | 본문 참조 | — | 아래 Branch Test Map |
| B28 | (491) `if` — if recomputed != nil | 본문 참조 | — | 아래 Branch Test Map |
| B29 | (493) `if` — if saved != nil | 본문 참조 | — | 아래 Branch Test Map |
| B30 | (498) `if` — if selectErr != nil | 본문 참조 | — | 아래 Branch Test Map |
| B31 | (499) `if` — if _, qerr := quarantineExitSnapshotTx(ctx, tx, id, recomputed.Line.PositionGe… | 본문 참조 | — | 아래 Branch Test Map |
| B32 | (503) `if` — if err := tx.Commit(); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B33 | (512) `if` — if err := releaseReJudgedQuarantineTx(ctx, tx, id, | 본문 참조 | — | 아래 Branch Test Map |
| B34 | (516) `if` — if source == exitpolicy.RecoverySavedMonotone | 본문 참조 | — | 아래 Branch Test Map |
| B35 | (519) `else` — } else | 본문 참조 | — | 아래 Branch Test Map |
| B36 | (529) `if` — if effectiveSource == EffectiveSourceSaved | 본문 참조 | — | 아래 Branch Test Map |
| B37 | (542) `if` — if effective != nil | 본문 참조 | — | 아래 Branch Test Map |
| B38 | (544) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B39 | (560) `if` — if _, err := tx.ExecContext(ctx, updateSQL, args...); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B40 | (563) `if` — if err := j.runExitWriteHook("after_state"); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B41 | (569) `if` — if judgement.Proposal != nil | 본문 참조 | — | 아래 Branch Test Map |
| B42 | (571) `if` — if err := armExitProposalTx(ctx, tx, id, *judgement.Proposal, now); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B43 | (574) `if` — if err := j.runExitWriteHook("after_arm"); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B44 | (585) `if` — if recomputed != nil && effective != nil | 본문 참조 | — | 아래 Branch Test Map |
| B45 | (588) `if` — if err := appendExitEventTx(ctx, tx, event); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B46 | (591) `if` — if err := j.runExitWriteHook("after_event"); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B47 | (594) `if` — if err := tx.Commit(); err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B48 | (598) `switch` — switch | 본문 참조 | — | 아래 Branch Test Map |
| B49 | (599) `case` — case effectiveSource == EffectiveSourceSaved: | 본문 참조 | — | 아래 Branch Test Map |
| B50 | (601) `case` — case judgement.Proposal != nil: | 본문 참조 | — | 아래 Branch Test Map |
| B51 | (605) `case` — case judgement.ArmSuppressedReason == ArmSuppressedWorkingOrder: | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | (394) id := strings.TrimSpace(judgement.PositionID) | 호출부 계약 유지 | AST `calls` |
| `fmt.Errorf` | (396) return fmt.Errorf("%w: a judgement needs a position", ErrInvalidRequest) | 호출부 계약 유지 | AST `calls` |
| `judgement.Provenance.validate` | (398) if err := judgement.Provenance.validate(); err != nil { | 호출부 계약 유지 | AST `calls` |
| `judgement.Provenance.zero` | (401) if judgement.Provenance.zero() && strings.TrimSpace(judgement.ArmSuppressedReason) != "" { | 호출부 계약 유지 | AST `calls` |
| `validateProposal` | (405) if err := validateProposal(*judgement.Proposal); err != nil { | 호출부 계약 유지 | AST `calls` |
| `judgement.Proposal.Provenance.validate` | (408) if err := judgement.Proposal.Provenance.validate(); err != nil { | 호출부 계약 유지 | AST `calls` |
| `judgement.Proposal.Provenance.zero` | (411) if judgement.Provenance.zero() != judgement.Proposal.Provenance.zero() \|\| | 호출부 계약 유지 | AST `calls` |
| `sameExitDecisionProvenance` | (413) !sameExitDecisionProvenance(judgement.Provenance, judgement.Proposal.Provenance)) { | 호출부 계약 유지 | AST `calls` |
| `Format` | (421) ObservedAt:        judgement.ObservedAt.UTC().Format(time.RFC3339Nano)} | 호출부 계약 유지 | AST `calls` |
| `judgement.ObservedAt.UTC` | (421) ObservedAt:        judgement.ObservedAt.UTC().Format(time.RFC3339Nano)} | 호출부 계약 유지 | AST `calls` |
| `validateJudgementSnapshot` | (422) if err := validateJudgementSnapshot(id, judgement, candidate); err != nil { | 호출부 계약 유지 | AST `calls` |
| `j.nowString` | (428) now := j.nowString() | 호출부 계약 유지 | AST `calls` |
| `j.db.BeginTx` | (429) tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE | 호출부 계약 유지 | AST `calls` |
| `tx.Rollback` | (433) defer tx.Rollback() | 호출부 계약 유지 | AST `calls` |
| `scanExitProgress` | (435) current, err := scanExitProgress(ctx, tx, id) | 호출부 계약 유지 | AST `calls` |
| `Scan` | (452) err = tx.QueryRowContext(ctx, `SELECT adoption_generation,status | 호출부 계약 유지 | AST `calls` |
| `tx.QueryRowContext` | (452) err = tx.QueryRowContext(ctx, `SELECT adoption_generation,status | 호출부 계약 유지 | AST `calls` |
| `errors.Is` | (455) if errors.Is(err, sql.ErrNoRows) { | 호출부 계약 유지 | AST `calls` |
| `notBelow` | (488) if err := notBelow("high water", id, judgement.HighWater, current.HighWater); err != nil { | 호출부 계약 유지 | AST `calls` |
| `exitpolicy.SelectRecoverySnapshot` | (512) selected, source, selectErr := exitpolicy.SelectRecoverySnapshot(savedLine, recomputed.Line) | 호출부 계약 유지 | AST `calls` |
| `quarantineExitSnapshotTx` | (514) if _, qerr := quarantineExitSnapshotTx(ctx, tx, id, recomputed.Line.PositionGeneration, | 호출부 계약 유지 | AST `calls` |
| `selectErr.Error` | (515) QuarantineReasonAmbiguousRecovery, selectErr.Error(), now); qerr != nil { | 호출부 계약 유지 | AST `calls` |
| `tx.Commit` | (518) if err := tx.Commit(); err != nil { | 호출부 계약 유지 | AST `calls` |
| `releaseReJudgedQuarantineTx` | (527) if err := releaseReJudgedQuarantineTx(ctx, tx, judgement.ReJudgingVersion, id, | 호출부 계약 유지 | AST `calls` |
| `nullableRung` | (556) nullableRung(judgement.ActiveRung), now, id} | 호출부 계약 유지 | AST `calls` |
| `encodeStoredSnapshot` | (558) raw, err := encodeStoredSnapshot(*effective) | 호출부 계약 유지 | AST `calls` |
| `nullableString` | (570) line.DecisionID, line.ObservationID, line.PositionGeneration, nullableString(line.NextTarget), | 호출부 계약 유지 | AST `calls` |
| `boolInt` | (573) boolInt(line.StateOnly), nullableString(line.Suppressed), raw, id} | 호출부 계약 유지 | AST `calls` |
| `tx.ExecContext` | (575) if _, err := tx.ExecContext(ctx, updateSQL, args...); err != nil { | 호출부 계약 유지 | AST `calls` |
| `j.runExitWriteHook` | (578) if err := j.runExitWriteHook("after_state"); err != nil { | 호출부 계약 유지 | AST `calls` |
| `armExitProposalTx` | (586) if err := armExitProposalTx(ctx, tx, id, *judgement.Proposal, now); err != nil { | 호출부 계약 유지 | AST `calls` |
| `levelAfter` | (596) LevelAfter: levelAfter(level, judgement.ActiveRung), Action: action, | 호출부 계약 유지 | AST `calls` |
| `evaluationForEvent` | (601) event.Evaluation = evaluationForEvent(saved, recomputed, effective, effectiveSource) | 호출부 계약 유지 | AST `calls` |
| `appendExitEventTx` | (603) if err := appendExitEventTx(ctx, tx, event); err != nil { | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- `exit_states` UPDATE, `exit_events` INSERT, 격리 INSERT 또는 (a084) 격리 UPDATE(해제). 전부 한 트랜잭션.

## Safety conclusion

- Safe edit boundary: 선택 성공 분기의 해제 호출 한 줄. 선택 규칙·격리 조건·제안·arm 억제는 그대로다.
- High-risk impact: yes — 격리 해제를 자동화한다. 원자성이 요구사항이고, 해제는 호출자가 전달한 버전과 활성 행이 일치할 때만 일어난다. 재판정되지 않은 격리도, 재판정된 행을 대체한 행도 이 경로로는 풀리지 않는다.
