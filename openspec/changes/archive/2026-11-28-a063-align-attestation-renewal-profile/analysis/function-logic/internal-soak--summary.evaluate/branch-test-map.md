# Branch Test Map: `Summary.Evaluate`

| B-id | Scenario | Focused test and bounded claim |
|---|---|---|
| B1 | An empty summary produces issues, so the range appends its refusal message. | `TestEvaluateRefusesAnEmptyRecord` checks false and nonempty reasons; the zero-iteration case is unmeasured by focused test. |
| B2 | A qualifying record returns `(true, nil)`; an empty record returns false with reasons. | `TestEvaluatePreservesNilReasonsForAQualifyingSoak` and `TestEvaluateRefusesAnEmptyRecord` establish those outcomes. |
