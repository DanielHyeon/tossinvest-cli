# Function Logic Map: `Journal.recordExitJudgementTx`

- Source: `internal/journal/exit_state.go` (lines 377–609)
- AST evidence: `ast.json` (`source_sha256: dae066216761246c204c834d48351225135f817633e2db0af949f5398ede1691`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 판정 기록과 격리 생성이 같은 트랜잭션에 있다.

## What it does

한 판정을 기록한다. a084는 복구 선택 성공 분기에 `releaseReJudgedQuarantineTx` 한 줄을 더해, 재판정이 성공하면 그것을 허용한 격리가 같은 트랜잭션에서 닫히게 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement | 관측·제안·snapshot | 호출자(exit loop) | invalid은 `ErrInvalidRequest` |
| current | 저장된 exit progress | 같은 트랜잭션의 SELECT | completed는 거부 |
| saved/recomputed | 복구 후보 두 개 | 저장 effective snapshot + 재계산 | 선택 실패는 격리 |
| 활성 격리 | 세대별 최대 1행 | 같은 트랜잭션 | 현재 개정이면 해제하지 않는다 |

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
| `{'kind': 'call', 'at': {'line': 379, 'column': 8}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 381, 'column': 10}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 383, 'column': 12}, 'text': 'judgement.Provenance.validate'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 386, 'column': 36}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 386, 'column': 5}, 'text': 'judgement.Provenance.zero'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 387, 'column': 10}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 390, 'column': 13}, 'text': 'validateProposal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 393, 'column': 13}, 'text': 'judgement.Proposal.Provenance.validate'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 396, 'column': 37}, 'text': 'judgement.Proposal.Provenance.zero'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 396, 'column': 6}, 'text': 'judgement.Provenance.zero'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 397, 'column': 6}, 'text': 'judgement.Provenance.zero'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 398, 'column': 6}, 'text': 'sameExitDecisionProvenance'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 399, 'column': 11}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 403, 'column': 6}, 'text': 'judgement.Provenance.zero'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- `exit_states` UPDATE, `exit_events` INSERT, 격리 INSERT 또는 (a084) 격리 UPDATE(해제). 전부 한 트랜잭션.

## Safety conclusion

- Safe edit boundary: 선택 성공 분기의 해제 호출 한 줄. 선택 규칙·격리 조건·제안·arm 억제는 그대로다.
- High-risk impact: yes — 격리 해제를 자동화한다. 원자성이 요구사항이고 `releaseReJudgedQuarantineTx`는 현재 개정 행에 대해 no-op이므로, 재판정되지 않은 격리는 해제되지 않는다.
