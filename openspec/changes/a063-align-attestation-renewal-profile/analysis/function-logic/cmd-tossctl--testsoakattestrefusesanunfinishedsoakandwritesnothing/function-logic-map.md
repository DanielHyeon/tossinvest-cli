# Function Logic Map: `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing`

- Source: immutable P blob `da80ce31b6a1ab5d443016768f970a82bab102db:cmd/tossctl/soak_test.go`, revision `base`, SHA-256 `158aa423021f8e9143a354acc6e08c8911418dc9274a37890c34df14bde173c2` (retrospective, not pre-edit evidence). a125 (2026-09-29): the adoption exception was retired, so the comparison base is P, not E; the E blob hash `562e7914…` was replaced with the P blob hash — branches and coordinates are identical between the two blobs (a125 `analysis/harness/p_blob_check.py` → `4.2-p-blob-check.txt`).
- This function is deleted from current source. The map records its E behavior only and makes no current runtime or execution claim.

## Inputs and invariants

This deleted function is immutable historical E evidence only, not a current runtime or execution claim. The E test isolated a config directory, pointed its soak client to the test server, ran exactly one zero-interval cycle, then invoked `soak attest`. Its contract was refusal for insufficient consecutive days and absence of the local attestation file; it was not a live broker or installed-service scenario.

## Branches and early returns

| B-id | E-source condition and effect | E focused-test behavior |
|---|---|---|
| B1 | Failure of the preparatory one-cycle `runCLI` calls `t.Fatalf`. | Requires the fixture soak run to complete before testing refusal. |
| B2 | `err == nil` after `soak attest` calls `t.Fatal`. | Requires refusal. |
| B3 | Error text lacking `consecutive` calls `t.Errorf`. | Requires a missing-days explanation. |
| B4 | `os.Stat` is not `ErrNotExist` calls `t.Fatal`. | Requires no attestation artifact after refusal. |

## Calls and live bindings

Calls create an isolated fixture, invoke the local CLI, inspect error text, and stat the temporary path. There is no retry/timeout or live binding.

## State mutations and fallbacks

The only mutation was fixture-local record/output state. `Fatal/Fatalf` are assertion fallbacks.

## Safety conclusion

The immutable E test records no live broker or installed-service scenario.
