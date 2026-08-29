# Branch Test Map: `Journal.recordExitJudgementTx`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (380) `if` — if id == "" | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B2 | (383) `if` — if err := judgement.Provenance.validate(); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B3 | (386) `if` — if judgement.Provenance.zero() && strings.TrimSpace(judgement.ArmSuppressedRea… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B4 | (389) `if` — if judgement.Proposal != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B5 | (390) `if` — if err := validateProposal(*judgement.Proposal); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B6 | (393) `if` — if err := judgement.Proposal.Provenance.validate(); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B7 | (396) `if` — if judgement.Provenance.zero() != judgement.Proposal.Provenance.zero() || | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B8 | (403) `if` — if !judgement.Provenance.zero() | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B9 | (407) `if` — if err := validateJudgementSnapshot(id, judgement, candidate); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B10 | (415) `if` — if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B11 | (421) `if` — if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B12 | (424) `if` — if current.Completed | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B13 | (428) `if` — if expectedLifecycle == 0 | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B14 | (431) `if` — if expectedLifecycle != current.LifecycleGeneration | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B15 | (440) `if` — if errors.Is(err, sql.ErrNoRows) | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B16 | (442) `else` — } else if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B17 | (442) `if` — } else if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B18 | (445) `if` — if lifecycleStatus != positionpolicy.StatusManaged || lifecycleGeneration != e… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B19 | (449) `if` — if recomputed != nil && recomputed.Line.PositionGeneration != current.Position… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B20 | (453) `if` — if recomputed != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B21 | (457) `if` — if err == nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B22 | (464) `if` — if !errors.Is(err, sql.ErrNoRows) | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B23 | (472) `if` — if recomputed == nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B24 | (473) `if` — if err := notBelow("high water", id, judgement.HighWater, current.HighWater); … | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B25 | (476) `if` — if err := notBelow("baseline", id, judgement.Baseline, current.Baseline); err … | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B26 | (481) `if` — if level == "" | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B27 | (487) `if` — if current.Effective != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B28 | (491) `if` — if recomputed != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B29 | (493) `if` — if saved != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B30 | (498) `if` — if selectErr != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B31 | (499) `if` — if _, qerr := quarantineExitSnapshotTx(ctx, tx, id, recomputed.Line.PositionGe… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B32 | (503) `if` — if err := tx.Commit(); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B33 | (512) `if` — if err := releaseReJudgedQuarantineTx(ctx, tx, id, | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B34 | (516) `if` — if source == exitpolicy.RecoverySavedMonotone | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B35 | (519) `else` — } else | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B36 | (529) `if` — if effectiveSource == EffectiveSourceSaved | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B37 | (542) `if` — if effective != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B38 | (544) `if` — if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B39 | (560) `if` — if _, err := tx.ExecContext(ctx, updateSQL, args...); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B40 | (563) `if` — if err := j.runExitWriteHook("after_state"); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B41 | (569) `if` — if judgement.Proposal != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B42 | (571) `if` — if err := armExitProposalTx(ctx, tx, id, *judgement.Proposal, now); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B43 | (574) `if` — if err := j.runExitWriteHook("after_arm"); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B44 | (585) `if` — if recomputed != nil && effective != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B45 | (588) `if` — if err := appendExitEventTx(ctx, tx, event); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B46 | (591) `if` — if err := j.runExitWriteHook("after_event"); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B47 | (594) `if` — if err := tx.Commit(); err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B48 | (598) `switch` — switch | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B49 | (599) `case` — case effectiveSource == EffectiveSourceSaved: | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B50 | (601) `case` — case judgement.Proposal != nil: | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B51 | (605) `case` — case judgement.ArmSuppressedReason == ArmSuppressedWorkingOrder: | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |

추가 근거: `TestANewQuarantineRecordsTheSelectorThatJudgedIt`,
`TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown`,
`TestAQuarantineFromTheCurrentSelectorIsNotReJudged` (journal 계층).
