# Function Logic Map: `BuildAttestation`

- Source: `internal/soak/attest.go`; E-based current S snapshot, SHA-256 `d693c9dbd9583789b6372aa63a126e7e4817df874c74ff76c597633eb60d29d5` (retrospective, not pre-edit evidence).
- AST evidence: `ast.json`.

## Inputs and invariants

`s` is a soaked record, `c` is defaulted, `now` fixes issuance/expiry, and `verifiedBy`, `notes`, and `supervised` become audit data. An issued attestation may contain read-only successful endpoints plus accepted supervised live-only proofs; a non-GET endpoint in the read-only record is rejected. This function constructs a value only: it does not save it or bind a live runtime setting.

## Branches and early returns

| B-id | Source condition and effect | Focused test evidence |
|---|---|---|
| B1 | `len(issues) != 0` returns an empty attestation and typed `IncompleteError`. | `TestBuildAttestationRefusesAnIncompleteSoak` checks error, `ErrIncomplete`, typed error, message, and streak reason code. |
| B2 | Iterates successful record endpoints; no endpoints gives zero iterations. | `TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise` supplies endpoints; zero-iteration case is unmeasured by focused test. |
| B3 | A trimmed, uppercased endpoint not prefixed `GET ` returns a refusal error. | `TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise` establishes rejection of injected `POST /api/v1/orders`. |
| B4 | `acceptSupervised` error returns empty attestation and that error. | unmeasured by focused test; the incomplete fixture returns before this call. |
| B5 | Iterates accepted supervised proofs and appends their endpoints. | unmeasured by focused test; the named incomplete test returns before acceptance. |
| B6 | Nonblank trimmed `notes` prefixes the generated note; blank notes leave the generated note unchanged. | unmeasured by focused test. |

## Calls and live bindings

It calls `withDefaults`, `evaluateIssues`, `SuccessfulEndpoints`, string normalization, `acceptSupervised`, formatting/time helpers, and `mutationNote`. There is no retry/timeout, filesystem write, broker request, order action, or gate enablement; persistence is the caller's responsibility.

## State mutations and fallbacks

Qualification and proof-validation errors are returned before value construction. The only mutation is local slice/string construction.

## Safety conclusion

This function constructs a value only: it does not save it or bind a live runtime setting.
