# Branch Test Map: `Journal.recordExitJudgementTx`

- Source: `internal/journal/exit_state.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:395` `if id == "" {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:398` `if err := judgement.Provenance.validate(); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:401` `if judgement.Provenance.zero() && strings.TrimSpace(judgement.ArmSuppressedReason) != "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:404` `if judgement.Proposal != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:405` `if err := validateProposal(*judgement.Proposal); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:408` `if err := judgement.Proposal.Provenance.validate(); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:411` `if judgement.Provenance.zero() != judgement.Proposal.Provenance.zero() \|\|` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:418` `if !judgement.Provenance.zero() {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:422` `if err := validateJudgementSnapshot(id, judgement, candidate); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:430` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:436` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:439` `if current.Completed {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:443` `if expectedLifecycle == 0 {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:446` `if expectedLifecycle != current.LifecycleGeneration {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B15 | `:455` `if errors.Is(err, sql.ErrNoRows) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B16 | `:457` `} else if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B17 | `:457` `} else if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B18 | `:460` `if lifecycleStatus != positionpolicy.StatusManaged \|\| lifecycleGeneration != expectedLifecycle {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B19 | `:464` `if recomputed != nil && recomputed.Line.PositionGeneration != current.PositionGeneration {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B20 | `:468` `if recomputed != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B21 | `:472` `if err == nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B22 | `:479` `if !errors.Is(err, sql.ErrNoRows) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B23 | `:487` `if recomputed == nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B24 | `:488` `if err := notBelow("high water", id, judgement.HighWater, current.HighWater); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B25 | `:491` `if err := notBelow("baseline", id, judgement.Baseline, current.Baseline); err != nil {` | 예 | **a095 5.3** — 하향 거부가 이미 있다는 사실을 issues I1에 인용 | no | no |
| B26 | `:496` `if level == "" {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B27 | `:502` `if current.Effective != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B28 | `:506` `if recomputed != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B29 | `:508` `if saved != nil {` | 예 | **a095 5.3** — 스냅샷 경로의 선택을 issues I1에 인용 | no | no |
| B30 | `:513` `if selectErr != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B31 | `:514` `if _, qerr := quarantineExitSnapshotTx(ctx, tx, id, recomputed.Line.PositionGeneration,` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B32 | `:518` `if err := tx.Commit(); err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B33 | `:527` `if err := releaseReJudgedQuarantineTx(ctx, tx, judgement.ReJudgingVersion, id,` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B34 | `:531` `if source == exitpolicy.RecoverySavedMonotone {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B35 | `:534` `} else {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B36 | `:544` `if effectiveSource == EffectiveSourceSaved {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B37 | `:557` `if effective != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B38 | `:559` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B39 | `:575` `if _, err := tx.ExecContext(ctx, updateSQL, args...); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B40 | `:578` `if err := j.runExitWriteHook("after_state"); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B41 | `:584` `if judgement.Proposal != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B42 | `:586` `if err := armExitProposalTx(ctx, tx, id, *judgement.Proposal, now); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B43 | `:589` `if err := j.runExitWriteHook("after_arm"); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B44 | `:600` `if recomputed != nil && effective != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B45 | `:603` `if err := appendExitEventTx(ctx, tx, event); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B46 | `:606` `if err := j.runExitWriteHook("after_event"); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B47 | `:609` `if err := tx.Commit(); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B48 | `:613` `switch {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B49 | `:614` `case effectiveSource == EffectiveSourceSaved:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B50 | `:616` `case judgement.Proposal != nil:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B51 | `:620` `case judgement.ArmSuppressedReason == ArmSuppressedWorkingOrder,` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 9개**: B1, B10, B11, B16, B17, B26, B30, B32, B38
**자체 블록 없는 분기 3개**: B31, B48, B51 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
