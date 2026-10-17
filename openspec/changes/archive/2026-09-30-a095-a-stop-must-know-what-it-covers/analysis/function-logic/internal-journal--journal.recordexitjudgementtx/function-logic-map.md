# Function Logic Map: `Journal.recordExitJudgementTx`

- Source: `internal/journal/exit_state.go` (`392`–`625`)
- Qualified: `Journal.recordExitJudgementTx`
- AST evidence: `ast.json` (`source_sha256` 99c96bdba4a08f7c…)
- Risk scan: `risk-pattern-report.md`
- 분기 51 · return 31 · 호출 69

**역할.** exit 판정을 기록한다. `baseline_price`를 **값이 바뀌게** 쓰는 정상 경로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `recomputed` | 재계산 스냅샷 | 호출자 | B23 — nil이면 옛 스칼라 단조 검사(B24 · B25) |
| `current.Baseline` | 저장된 기준선 | 원장 | B25 — `notBelow("baseline", …)` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:395` `if id == "" {` | `fmt.Errorf` | :396 | 아니오 |
| B2 | if | `:398` `if err := judgement.Provenance.validate(); err != nil {` | `judgement.Provenance.validate` | :399 | 예 |
| B3 | if | `:401` `if judgement.Provenance.zero() && strings.TrimSpace(judgement.ArmSuppressedReason) != "" {` | `fmt.Errorf`, `judgement.Provenance.zero`, `strings.TrimSpace` | :402 | 예 |
| B4 | if | `:404` `if judgement.Proposal != nil {` | — | — | 예 |
| B5 | if | `:405` `if err := validateProposal(*judgement.Proposal); err != nil {` | `validateProposal` | :406 | 예 |
| B6 | if | `:408` `if err := judgement.Proposal.Provenance.validate(); err != nil {` | `judgement.Proposal.Provenance.validate` | :409 | 예 |
| B7 | if | `:411` `if judgement.Provenance.zero() != judgement.Proposal.Provenance.zero() \|\|` | `fmt.Errorf`, `judgement.Proposal.Provenance.zero`, `judgement.Provenance.zero`, `sameExitDecisionProvenance` | :414 | 예 |
| B8 | if | `:418` `if !judgement.Provenance.zero() {` | `Format`, `judgement.ObservedAt.UTC`, `judgement.Provenance.zero`, `strings.TrimSpace` | — | 예 |
| B9 | if | `:422` `if err := validateJudgementSnapshot(id, judgement, candidate); err != nil {` | `fmt.Errorf`, `j.db.BeginTx`, `j.nowString`, `validateJudgementSnapshot` | :423 | 예 |
| B10 | if | `:430` `if err != nil {` | `fmt.Errorf`, `scanExitProgress`, `tx.Rollback` | :431 | 아니오 |
| B11 | if | `:436` `if err != nil {` | — | :437 | 아니오 |
| B12 | if | `:439` `if current.Completed {` | `fmt.Errorf` | :440 | 예 |
| B13 | if | `:443` `if expectedLifecycle == 0 {` | — | — | 예 |
| B14 | if | `:446` `if expectedLifecycle != current.LifecycleGeneration {` | `Scan`, `fmt.Errorf`, `tx.QueryRowContext` | :447 | 예 |
| B15 | if | `:455` `if errors.Is(err, sql.ErrNoRows) {` | `errors.Is` | — | 예 |
| B16 | else | `:457` `} else if err != nil {` | — | — | 아니오 |
| B17 | if | `:457` `} else if err != nil {` | `fmt.Errorf` | :458 | 아니오 |
| B18 | if | `:460` `if lifecycleStatus != positionpolicy.StatusManaged \|\| lifecycleGeneration != expectedLifecycle {` | `fmt.Errorf` | :461 | 예 |
| B19 | if | `:464` `if recomputed != nil && recomputed.Line.PositionGeneration != current.PositionGeneration {` | `fmt.Errorf` | :465 | 예 |
| B20 | if | `:468` `if recomputed != nil {` | `Scan`, `tx.QueryRowContext` | — | 예 |
| B21 | if | `:472` `if err == nil {` | — | :477 | 예 |
| B22 | if | `:479` `if !errors.Is(err, sql.ErrNoRows) {` | `errors.Is`, `fmt.Errorf` | :480 | 예 |
| B23 | if | `:487` `if recomputed == nil {` | — | — | 예 |
| B24 | if | `:488` `if err := notBelow("high water", id, judgement.HighWater, current.HighWater); err != nil {` | `notBelow` | :489 | 예 |
| B25 | if | `:491` `if err := notBelow("baseline", id, judgement.Baseline, current.Baseline); err != nil {` | `notBelow`, `strings.TrimSpace` | :492 | 예 |
| B26 | if | `:496` `if level == "" {` | — | — | 아니오 |
| B27 | if | `:502` `if current.Effective != nil {` | — | — | 예 |
| B28 | if | `:506` `if recomputed != nil {` | — | — | 예 |
| B29 | if | `:508` `if saved != nil {` | `exitpolicy.SelectRecoverySnapshot` | — | 예 |
| B30 | if | `:513` `if selectErr != nil {` | — | — | 아니오 |
| B31 | if | `:514` `if _, qerr := quarantineExitSnapshotTx(ctx, tx, id, recomputed.Line.PositionGeneration,` | `quarantineExitSnapshotTx`, `selectErr.Error` | :516 | — |
| B32 | if | `:518` `if err := tx.Commit(); err != nil {` | `fmt.Errorf`, `tx.Commit` | :519, :521 | 아니오 |
| B33 | if | `:527` `if err := releaseReJudgedQuarantineTx(ctx, tx, judgement.ReJudgingVersion, id,` | `releaseReJudgedQuarantineTx` | :529 | 예 |
| B34 | if | `:531` `if source == exitpolicy.RecoverySavedMonotone {` | — | — | 예 |
| B35 | else | `:534` `} else {` | `string` | — | 예 |
| B36 | if | `:544` `if effectiveSource == EffectiveSourceSaved {` | `nullableRung`, `strings.TrimSpace` | — | 예 |
| B37 | if | `:557` `if effective != nil {` | `encodeStoredSnapshot` | — | 예 |
| B38 | if | `:559` `if err != nil {` | `boolInt`, `nullableRung`, `nullableString`, `string` | :560 | 아니오 |
| B39 | if | `:575` `if _, err := tx.ExecContext(ctx, updateSQL, args...); err != nil {` | `fmt.Errorf`, `tx.ExecContext` | :576 | 예 |
| B40 | if | `:578` `if err := j.runExitWriteHook("after_state"); err != nil {` | `j.runExitWriteHook` | :579 | 예 |
| B41 | if | `:584` `if judgement.Proposal != nil {` | — | — | 예 |
| B42 | if | `:586` `if err := armExitProposalTx(ctx, tx, id, *judgement.Proposal, now); err != nil {` | `armExitProposalTx` | :587 | 예 |
| B43 | if | `:589` `if err := j.runExitWriteHook("after_arm"); err != nil {` | `j.runExitWriteHook`, `levelAfter` | :590 | 예 |
| B44 | if | `:600` `if recomputed != nil && effective != nil {` | `evaluationForEvent` | — | 예 |
| B45 | if | `:603` `if err := appendExitEventTx(ctx, tx, event); err != nil {` | `appendExitEventTx` | :604 | 예 |
| B46 | if | `:606` `if err := j.runExitWriteHook("after_event"); err != nil {` | `j.runExitWriteHook` | :607 | 예 |
| B47 | if | `:609` `if err := tx.Commit(); err != nil {` | `fmt.Errorf`, `tx.Commit` | :610 | 예 |
| B48 | switch | `:613` `switch {` | — | — | — |
| B49 | case | `:614` `case effectiveSource == EffectiveSourceSaved:` | — | — | 예 |
| B50 | case | `:616` `case judgement.Proposal != nil:` | — | — | 예 |
| B51 | case | `:620` `case judgement.ArmSuppressedReason == ArmSuppressedWorkingOrder,` | — | :624 | — |

## Calls and live bindings

`notBelow`(B24 · B25) · `exitpolicy.SelectRecoverySnapshot`(B29 창) · `tx.ExecContext`(B39 — UPDATE).

결과에 `error`가 있다 — 원장(트랜잭션 · 질의) 호출의 오류와 입력 검증 실패를 되던진다(위 표의 return 열이 그 자리다). 브로커 호출은 없다(9판 r8 N5 — 값 단위 재전수).

## State mutations and fallbacks

`exit_states.baseline_price` 외 판정 열 · 제안 무장 · `exit_events`.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** 판정마다 `baseline_price`를 UPDATE한다(B39). 하향에 대한 보장은 경로마다 **전제가 다르다(4판, r3 N5)**: 옛 경로(B23 `recomputed == nil`)는 B25의 `notBelow`가 **스칼라** `baseline_price`와 비교하고, B37(`effective != nil`)이 거짓이면 스칼라 열만 쓰고 effective JSON은 다시 쓰지 않는다. 스냅샷 경로는 B29 창의 `SelectRecoverySnapshot`이 **저장된 effective 스냅샷**과 비교하며, 저장 스냅샷이 없으면(그 함수 B2) 비교 없이 재계산값을 받는다 — 스칼라와 비교하지 않는다. 475150의 57,900은 이 함수의 산물이다.
- **High-risk impact**: yes — 손절선의 정상 갱신 자리다.
