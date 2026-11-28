# Branch Test Map: `runSoakAttest`

| B-id | Scenario | Focused test and bounded claim | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Enable renewal-status mode. | `TestSoakAttestRecordsBoundedRefusalStatus`. | n/a | yes |
| B2 | Reject record/out override in renewal-status mode. | `TestSoakAttestRenewalStatusRejectsPathOverrides`. | n/a | yes |
| B3 | Default-path resolution failure. | unmeasured by focused test. | n/a | n/a |
| B4 | Deferred successful issuance status handling. | `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation`. | n/a | yes |
| B5 | Deferred issued-attestation load failure. | unmeasured by focused test. | n/a | n/a |
| B6 | Copy issued expiry into status. | `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation`. | n/a | yes |
| B7 | (D6) Issued file lacks a live-only endpoint → `live_endpoint_uncovered`; covered → no codes. | true `TestA063RenewalWithoutSupervisedEvidenceRecordsTheUncoveredScope`; false `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation`. | yes (`analysis/scope-absorb/red.txt`) | yes |
| B8 | Status-save failure only supersedes nil result. | `TestSoakAttestStatusWriteFailureKeepsIssuedAttestationAndFailsCommand`. | n/a | yes |
| B9 | Summary load error. | unmeasured by focused test. | n/a | n/a |
| B10 | Positive validity override. | unmeasured by focused test. | n/a | n/a |
| B11 | Surveyed-base note augmentation. | unmeasured by focused test. | n/a | n/a |
| B12 | Supervised-proof error. | `TestSoakAttestRefusesSupervisedEvidenceFromAnotherAccount`. | n/a | yes |
| B13 | Incomplete soak returns before writing attestation. | `TestSoakAttestRecordsBoundedRefusalStatus`. | n/a | yes |
| B14 | Ordinary output-path resolution. | `TestSoakAttestWritesAVerifiableAttestation`. | n/a | yes |
| B15 | Ordinary output-path resolution error. | unmeasured by focused test. | n/a | n/a |
| B16 | Directory creation error. | unmeasured by focused test. | n/a | n/a |
| B17 | Attestation save error. | unmeasured by focused test. | n/a | n/a |
| B18 | Print supervised proofs. | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun`. | n/a | yes |
| B19 | Blank proof market becomes em dash. | unmeasured by focused test. | n/a | n/a |
| B20 | Warn and return when live-only coverage is missing. | `TestSoakAttestWritesAVerifiableAttestation`. | n/a | yes |
| B21 | Print each missing requirement; empty list falls through. | `TestSoakAttestWritesAVerifiableAttestation` · `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun`. | n/a | yes |
