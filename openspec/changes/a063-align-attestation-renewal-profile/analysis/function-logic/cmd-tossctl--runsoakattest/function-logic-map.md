# Function Logic Map: `runSoakAttest`

- Source: `cmd/tossctl/soak.go`; E-based current S snapshot, SHA-256 `5602b691caadfdf4694fabe6f475e4a8fbc5e5195a4a4cd87ad63132c783a4c2` (retrospective, not pre-edit evidence).
- AST evidence: `ast.json`.

## Inputs and invariants

`root` and `opts` determine the local profile/record/attestation paths, criteria overrides, notes, verifier, and optional supervised-proof records. Normal issuance requires `loadSoakSummary`, supervised proof validation, and `BuildAttestation` to succeed before directory creation and `attest.Save`. With `recordRenewalStatus`, `--record`/`--out` are forbidden, the standard attestation path is used, and a deferred local renewal status is attempted for every outcome. The function prints that endpoint coverage does not itself start the automation gate.

## Branches and early returns

| B-id | Source condition and return/effect | Focused test evidence |
|---|---|---|
| B1 | `recordRenewalStatus` enables the special status-path/defer flow. | `TestSoakAttestRecordsBoundedRefusalStatus` enables it and checks refused local status. |
| B2 | Nonblank `opts.record` or `opts.out` with that flag returns a flag-conflict error. | unmeasured by focused test. |
| B3 | Failure resolving the default attestation path returns that error before summary loading. | unmeasured by focused test. |
| B4 | Deferred `result == nil` treats the command result as issued, clears reason codes, and tries to load the saved attestation. | `TestSoakAttestWritesAVerifiableAttestation` establishes ordinary successful local issuance, but does not enable this deferred-status path. |
| B5 | Deferred `attest.Load` failure replaces an otherwise nil result with a contextual error. | unmeasured by focused test. |
| B6 | Else after successful deferred load copies `ExpiresAt` to the status. | unmeasured by focused test. |
| B7 | Deferred `saveRenewalStatus` failure replaces only a nil result; an earlier result is retained. | unmeasured by focused test. |
| B8 | `loadSoakSummary` error returns immediately (defer still records status when enabled). | unmeasured by focused test. |
| B9 | Positive `opts.validity` overrides `criteria.Validity`; nonpositive leaves loaded/default value. | unmeasured by focused test. |
| B10 | Nonblank surveyed base appends trimmed survey provenance to notes. | unmeasured by focused test. |
| B11 | `supervisedProofs` error returns immediately. | unmeasured by focused test. |
| B12 | `BuildAttestation` error returns without directory/save; status defer records refusal if enabled. | `TestSoakAttestRecordsBoundedRefusalStatus` supplies an unfinished soak and checks refusal status plus no attestation. |
| B13 | Without renewal-status mode, resolve the requested/default output path. | `TestSoakAttestWritesAVerifiableAttestation` uses ordinary output resolution. |
| B14 | Output path resolution error returns. | unmeasured by focused test. |
| B15 | `os.MkdirAll` failure returns a wrapped creation error. | unmeasured by focused test. |
| B16 | `attest.Save` failure returns its error. | unmeasured by focused test. |
| B17 | Ranges `SupervisedBy` to print each proof. | unmeasured by focused test. |
| B18 | Empty proof market prints `—`; otherwise prints its market. | unmeasured by focused test. |
| B19 | Nonempty missing live-only endpoints prints each and returns nil after successful local save. | `TestSoakAttestWritesAVerifiableAttestation` checks output names every `LiveOnlyEndpoints` entry. |
| B20 | Ranges missing endpoints for the warning list; empty list falls through to the final still-refused message and nil return. | The named verifiable-attestation test establishes the nonempty warning-list behavior; empty-list fallthrough is unmeasured. |

## Calls and live bindings

Calls resolve local paths, load the local soak summary, collect supervised proofs, build/load/save the attestation and optional renewal status, create the output directory, and format CLI output. `time.Now().UTC()` timestamps the decision, but this function contains no retry/timeout policy.

## State mutations and fallbacks

Errors are returned immediately at validation/loading/building/persistence boundaries; the named result lets the defer preserve an original error over status-write failure. Durable mutations are limited to the local attestation and optional renewal-status artifacts.

## Safety conclusion

It does not place/cancel/amend an order, change a runtime toggle, restart a service, or satisfy the protective-order interlock.
