# Branch Test Map: `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing`

| B-id | E scenario | Evidence limit |
|---|---|---|
| B1 | Fixture soak run fails. | The deleted E test's own `Fatalf` assertion; immutable source evidence only. |
| B2 | One-cycle attestation unexpectedly succeeds. | The deleted E test's own `Fatal` assertion; no current rerun claim. |
| B3 | Refusal omits `consecutive`. | The deleted E test's own `Errorf` assertion; no current rerun claim. |
| B4 | Refusal leaves an attestation path present/stat-error other than missing. | The deleted E test's own `Fatal` assertion; no current rerun claim. |
