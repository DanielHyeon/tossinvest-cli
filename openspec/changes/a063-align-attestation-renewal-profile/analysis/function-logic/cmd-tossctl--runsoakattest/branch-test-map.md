# Branch Test Map: `runSoakAttest`

| B-id | Scenario | Focused test and bounded claim |
|---|---|---|
| B1 | Enable renewal-status mode. | `TestSoakAttestRecordsBoundedRefusalStatus`. |
| B2 | Reject record/out override in renewal-status mode. | unmeasured by focused test. |
| B3 | Default-path resolution failure. | unmeasured by focused test. |
| B4 | Deferred successful issuance status handling. | `TestSoakAttestWritesAVerifiableAttestation` covers ordinary issuance, not this status branch. |
| B5 | Deferred issued-attestation load failure. | unmeasured by focused test. |
| B6 | Copy issued expiry into status. | unmeasured by focused test. |
| B7 | Status-save failure only supersedes nil result. | unmeasured by focused test. |
| B8 | Summary load error. | unmeasured by focused test. |
| B9 | Positive validity override. | unmeasured by focused test. |
| B10 | Surveyed-base note augmentation. | unmeasured by focused test. |
| B11 | Supervised-proof error. | unmeasured by focused test. |
| B12 | Incomplete soak returns before writing attestation. | `TestSoakAttestRecordsBoundedRefusalStatus`. |
| B13 | Ordinary output-path resolution. | `TestSoakAttestWritesAVerifiableAttestation`. |
| B14 | Ordinary output-path resolution error. | unmeasured by focused test. |
| B15 | Directory creation error. | unmeasured by focused test. |
| B16 | Attestation save error. | unmeasured by focused test. |
| B17 | Print supervised proofs. | unmeasured by focused test. |
| B18 | Blank proof market becomes em dash. | unmeasured by focused test. |
| B19 | Warn and return when live-only coverage is missing. | `TestSoakAttestWritesAVerifiableAttestation`. |
| B20 | Print each missing requirement. | Same test covers nonempty list; empty-list path unmeasured. |
