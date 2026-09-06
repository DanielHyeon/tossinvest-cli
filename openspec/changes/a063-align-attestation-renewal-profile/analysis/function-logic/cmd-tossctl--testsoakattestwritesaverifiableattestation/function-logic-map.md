# Function Logic Map: `TestSoakAttestWritesAVerifiableAttestation`

- Source: immutable E blob `e65e394bf84b3c6e4559a219e816af96d341d75d:cmd/tossctl/soak_test.go`, revision `base`, SHA-256 `562e7914c5b47738c7cd03418e0ae872a56a227fd45776faffbd011e0b96200f` (retrospective, not pre-edit evidence).
- This function is deleted from current source. It is immutable E evidence, not a current runtime/execution claim.

## Inputs and invariants

This deleted function is immutable historical E evidence only, not a current runtime or execution claim. The E test isolated a config directory, seeded a qualifying record, invoked local `soak attest --verified-by tester`, then loaded and verified the written local attestation against the fixture account and required endpoints. It required read-only endpoint claims, a three-day minimum, printed remaining live-only requirements, and account masking.

## Branches and early returns

| B-id | E-source condition/effect | E focused-test behavior |
|---|---|---|
| B1 | CLI error calls `t.Fatalf`. | Attestation issuance must succeed for the seeded record. |
| B2 | Load error calls `t.Fatalf`. | Written artifact must load. |
| B3 | `a.Verify` error calls `t.Fatalf`. | Artifact must verify for fixture account/endpoints. |
| B4 | `a.SoakDays < 3` records `t.Errorf`. | Enforces day minimum. |
| B5 | Ranges attested endpoints. | Iterates each claimed endpoint. |
| B6 | A claimed endpoint without `GET ` records `t.Errorf`. | Enforces read-only claims. |
| B7 | Ranges `LiveOnlyEndpoints`. | Iterates every still-required live endpoint. |
| B8 | Output omitting one remaining endpoint records `t.Errorf`. | Requires the warning list. |
| B9 | Output containing full account suffix `678901` calls `t.Error`. | Requires masking. |

## Calls and live bindings

The E test invokes the local CLI, then loads and verifies the written local attestation against fixture data. There is no retry, timeout, or live-account action in the test.

## State mutations and fallbacks

Mutations were confined to the temporary fixture artifact. `Fatal/Fatalf/Errorf/Error` are assertion fallbacks.

## Safety conclusion

The immutable E test records local fixture behavior only and makes no current runtime or execution claim.
