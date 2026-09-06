# Function Logic Map: `Console.readAttestation`

- Source: `internal/console/data.go`; E-based current S snapshot, SHA-256 `b121d8b15fc6d4d997c0d4dc8752d89c9a1b51bf0c7de77de18aec4f5655fa7f` (retrospective, not pre-edit evidence).
- AST evidence: `ast.json`.

## Inputs and invariants

`c.opts.Attestation` is the configured local attestation path, `c.opts.RequiredEndpoints` is the dashboard's required set, and `now` drives expiry status. It builds an advisory `attestView`: it masks the account, reports missing/invalid data, calls `readRenewalStatus` only in the subsequent load-error and loaded-success flows after a nonblank trimmed path, and never changes attestation usability merely because a renewal diagnostic is refused.

## Branches and early returns

| B-id | Source condition and effect | Focused test evidence |
|---|---|---|
| B1 | Blank trimmed path returns the initialized unknown view. | unmeasured by the named renewal-diagnostic test. |
| B2 | `attest.Load` error enters missing/unreadable handling, invokes `readRenewalStatus`, and returns. | unmeasured by the named renewal-diagnostic test. |
| B3 | A load error other than `attest.ErrMissing` appends the Korean unreadable-file reason; missing errors do not. | unmeasured by focused test. |
| B4 | Switches over invalid loaded attestation content. | `TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability` supplies a valid account/endpoints, so invalid cases are unmeasured. |
| B5 | Blank account appends the account-missing reason. | unmeasured by focused test. |
| B6 | Empty endpoint list appends the no-proven-endpoints reason. | unmeasured by focused test. |
| B7 | Expired attestation appends the re-soak/re-attest reason. | unmeasured by focused test. |
| B8 | Nonzero future expiry computes `ExpiryHorizon`; the warning assignment independently covers expiry within 72 hours. | The named test uses a 48-hour expiry and asserts the 72-hour UI text; direct horizon value is not asserted. |
| B9 | Ranges `v.Missing`, appending one reason per required endpoint not present. | The named test uses all required endpoints, so the nonempty range case is unmeasured. |

## Calls and live bindings

For a nonblank trimmed path, calls load/mask/inspect the local attestation and runs `readRenewalStatus` in the load-error or loaded-success flow; the blank-path early return does not run that diagnostic. There is no retry/timeout, file write, broker call, live binding, or gate mutation.

## State mutations and fallbacks

Errors are represented as advisory fields rather than propagated. State mutation is limited to the returned local view.

## Safety conclusion

The advisory view never changes attestation usability merely because a renewal diagnostic is refused.
