# Branch Test Map: `TestBuildAttestationRefusesAnIncompleteSoak`

| B-id | Scenario | Focused test |
|---|---|---|
| B1 | Refusal must produce an error. | `TestBuildAttestationRefusesAnIncompleteSoak` (its intended passing path). |
| B2 | Error wraps `ErrIncomplete`. | Same test. |
| B3 | Error has `*IncompleteError` type. | Same test. |
| B4 | Error reports the refresh qualification failure. | Same test. |
| B5 | Reason codes begin with `ReasonStreak`. | Same test. |
