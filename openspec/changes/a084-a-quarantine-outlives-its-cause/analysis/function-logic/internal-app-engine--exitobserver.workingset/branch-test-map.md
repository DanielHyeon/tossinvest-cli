# Branch Test Map: `ExitObserver.workingSet`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (478) `if` — if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B2 | (482) `if` — if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B3 | (486) `range` — for _, result := range stateResults | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B4 | (491) `range` — for _, p := range positions | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B5 | (492) `if` — if p.State == journal.PositionClosed || isZeroQuantity(p.Quantity) | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B6 | (495) `if` — if !p.ExitEligible() | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B7 | (506) `if` — if !ok | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B8 | (508) `if` — if err != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B9 | (509) `if` — if cycle.Err == nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B10 | (514) `if` — if opened.PositionID == "" | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B11 | (523) `if` — if result.Corruption != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B12 | (526) `if` — if qerr != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B13 | (527) `if` — if cycle.Err == nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B14 | (537) `if` — if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p… | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B15 | (542) `else` — } else if active && !q.NeedsReJudgement() | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B16 | (538) `if` — if cycle.Err == nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B17 | (542) `if` — } else if active && !q.NeedsReJudgement() | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B18 | (546) `else` — } else if active | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B19 | (546) `if` — } else if active | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B20 | (563) `if` — if identityErr != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B21 | (566) `if` — if qerr != nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |
| B22 | (567) `if` — if cycle.Err == nil | `TestASupersededQuarantineIsReJudgedAndReleased`, `TestAQuarantineThisSelectorWroteIsStillSkipped` | yes | yes |

추가 근거: `TestANewQuarantineRecordsTheSelectorThatJudgedIt`,
`TestAQuarantineWrittenBeforeTheColumnReadsAsUnknown`,
`TestAQuarantineFromTheCurrentSelectorIsNotReJudged` (journal 계층).
