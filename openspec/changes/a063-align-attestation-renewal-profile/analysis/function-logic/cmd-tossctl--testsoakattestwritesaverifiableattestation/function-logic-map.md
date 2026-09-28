# Function Logic Map: `TestSoakAttestWritesAVerifiableAttestation`

- Source: immutable P blob `da80ce31b6a1ab5d443016768f970a82bab102db:cmd/tossctl/soak_test.go`, revision `base`, SHA-256 `158aa423021f8e9143a354acc6e08c8911418dc9274a37890c34df14bde173c2` (retrospective, not pre-edit evidence). a125 (2026-09-29): the adoption exception was retired, so the comparison base is P, not E; the E blob hash `562e7914…` was replaced with the P blob hash — branches and coordinates are identical between the two blobs (a125 `analysis/python-function-logic` receipt).
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
