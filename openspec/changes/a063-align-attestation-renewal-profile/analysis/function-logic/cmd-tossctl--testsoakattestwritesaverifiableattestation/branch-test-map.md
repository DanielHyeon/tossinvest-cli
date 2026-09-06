# Branch Test Map: `TestSoakAttestWritesAVerifiableAttestation`

| B-id | E scenario | Evidence limit |
|---|---|---|
| B1 | Seeded-record CLI issuance errors. | Deleted E test's own `Fatalf`; immutable source only. |
| B2 | Saved artifact cannot load. | Deleted E test's own `Fatalf`; no current rerun claim. |
| B3 | Loaded artifact fails verification. | Deleted E test's own `Fatalf`; no current rerun claim. |
| B4 | Soak days below three. | Deleted E test's own `Errorf`; no current rerun claim. |
| B5 | Iterate claimed endpoints. | Deleted E test's loop; test semantics only. |
| B6 | Claim is not GET. | Deleted E test's `Errorf`; no current rerun claim. |
| B7 | Iterate live-only requirements. | Deleted E test's loop; test semantics only. |
| B8 | Output misses a remaining live-only endpoint. | Deleted E test's `Errorf`; no current rerun claim. |
| B9 | Output leaks full account suffix. | Deleted E test's `Error`; no current rerun claim. |
