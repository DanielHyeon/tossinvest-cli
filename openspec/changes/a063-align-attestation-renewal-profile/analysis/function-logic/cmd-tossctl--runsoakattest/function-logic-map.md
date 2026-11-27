# Function Logic Map: `runSoakAttest`

- Source: `cmd/tossctl/soak.go` lines 477-588; SHA-256 `65a048a5cbc10cc8ebf7513d8b8785a58106878436d19112e32939d595dcde8c` (2026-11-28 a063 D6 post-edit current snapshot).
- AST evidence: `ast.json` (21 branches). Pre-edit AST at HEAD `baeddfe31`: `../../scope-absorb/pre-edit-ast/runSoakAttest.json` (20 branches).
- D6 edit: one new `if` (B7) inside the deferred successful-load branch. Every later branch shifts by one ID; no other predicate, message or return changed.

## Inputs and invariants

`root` and `opts` determine the local profile/record/attestation paths, criteria overrides, notes, verifier, and optional supervised-proof records. Normal issuance requires `loadSoakSummary`, supervised proof validation, and `BuildAttestation` to succeed before directory creation and `attest.Save`. With `recordRenewalStatus`, `--record`/`--out` are forbidden, the standard attestation path is used, and a deferred local renewal status is attempted for every outcome. Invariant kept by D6: the command result (exit status), the attestation bytes and the issuer criteria do not depend on the new branch; it only changes the reason codes of an `issued` diagnostic.

## Branches and early returns

| B-id | Source condition and return/effect | Focused test evidence |
|---|---|---|
| B1 | `recordRenewalStatus` enables the special status-path/defer flow. | `TestSoakAttestRecordsBoundedRefusalStatus`. |
| B2 | Nonblank `opts.record` or `opts.out` with that flag returns a flag-conflict error. | `TestSoakAttestRenewalStatusRejectsPathOverrides`. |
| B3 | Failure resolving the default attestation path returns that error before summary loading. | unmeasured by focused test. |
| B4 | Deferred `result == nil` sets outcome issued, clears reason codes, and loads the saved attestation. | `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation`. |
| B5 | Deferred `attest.Load` failure replaces an otherwise nil result with a contextual error. | unmeasured by focused test. |
| B6 | Else after successful deferred load copies `ExpiresAt` to the status. | `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation` checks the expiry equality. |
| B7 | (D6, new) The loaded attestation misses any `soak.LiveOnlyEndpoints()` entry → reason codes become `[live_endpoint_uncovered]`; outcome stays issued, result untouched. False → codes stay empty. | true: `TestA063RenewalWithoutSupervisedEvidenceRecordsTheUncoveredScope`; false: `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation` (supervised record seeded, empty codes asserted). |
| B8 | Deferred `saveRenewalStatus` failure replaces only a nil result; an earlier result is retained. | `TestSoakAttestStatusWriteFailureKeepsIssuedAttestationAndFailsCommand`. |
| B9 | `loadSoakSummary` error returns immediately (defer still records status when enabled). | unmeasured by focused test. |
| B10 | Positive `opts.validity` overrides `criteria.Validity`; nonpositive leaves loaded/default value. | unmeasured by focused test. |
| B11 | Nonblank surveyed base appends trimmed survey provenance to notes. | unmeasured by focused test. |
| B12 | `supervisedProofs` error returns immediately. | `TestSoakAttestRefusesSupervisedEvidenceFromAnotherAccount`. |
| B13 | `BuildAttestation` error returns without directory/save; status defer records refusal if enabled. | `TestSoakAttestRecordsBoundedRefusalStatus`. |
| B14 | Without renewal-status mode, resolve the requested/default output path. | `TestSoakAttestWritesAVerifiableAttestation`. |
| B15 | Output path resolution error returns. | unmeasured by focused test. |
| B16 | `os.MkdirAll` failure returns a wrapped creation error. | unmeasured by focused test. |
| B17 | `attest.Save` failure returns its error. | unmeasured by focused test. |
| B18 | Ranges `SupervisedBy` to print each proof. | `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun`. |
| B19 | Empty proof market prints `—`; otherwise prints its market. | unmeasured by focused test (market is always set by the fixture). |
| B20 | Nonempty missing live-only endpoints prints each and returns nil after successful local save. | `TestSoakAttestWritesAVerifiableAttestation`. |
| B21 | Ranges missing endpoints for the warning list; empty list falls through to the final message and nil return. | nonempty: `TestSoakAttestWritesAVerifiableAttestation`; empty: `TestSoakAttestCoversTheEngineSetOnceTheSupervisedCheckHasRun`. |

## Calls and live bindings

Calls resolve local paths, load the local soak summary, collect supervised proofs, build/load/save the attestation and optional renewal status, create the output directory, and format CLI output. D6 adds `Attestation.MissingEndpoints(soak.LiveOnlyEndpoints())` on the already loaded attestation — the same set and method B20 uses, pure, no I/O. `time.Now().UTC()` timestamps the decision; no retry/timeout policy.

## State mutations and fallbacks

Errors are returned immediately at validation/loading/building/persistence boundaries; the named result lets the defer preserve an original error over status-write failure. Durable mutations are limited to the local attestation and optional renewal-status artifacts. B7 mutates only `status.ReasonCodes` before the save.

## Safety conclusion

- It does not place/cancel/amend an order, change a runtime toggle, restart a service, or satisfy the protective-order interlock.
- High-risk impact: no — the renewal-status file is read only by the console (`attestView.readRenewalStatus`), never by the engine or the startup interlock.
