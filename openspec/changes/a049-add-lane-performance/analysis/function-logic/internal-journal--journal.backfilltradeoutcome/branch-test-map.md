# Branch Test Map: `Journal.BackfillTradeOutcome`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 758: `if _, err := j.TradeOutcomeOf(ctx, id); err == nil {`; invariant: missing/corrupt/alternate path is explicit | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `else` at line 760: `} else if !errors.Is(err, ErrTradeOutcomeNotFound) {`; invariant: missing/corrupt/alternate path is explicit | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B3 | `if` at line 760: `} else if !errors.Is(err, ErrTradeOutcomeNotFound) {`; invariant: missing/corrupt/alternate path is explicit | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B4 | `if` at line 765: `if !ok {`; invariant: missing/corrupt/alternate path is explicit | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B5 | `if` at line 769: `if _, err := j.db.ExecContext(ctx, \``; invariant: missing/corrupt/alternate path is explicit | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B6 | `if` at line 779: `if isUniqueViolation(err) {`; invariant: missing/corrupt/alternate path is explicit | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
