# Function Logic Map: `TestSoakAttestRefusesAnUnfinishedSoakAndWritesNothing`

- Source: immutable E blob `e65e394bf84b3c6e4559a219e816af96d341d75d:cmd/tossctl/soak_test.go`, revision `base`, SHA-256 `562e7914c5b47738c7cd03418e0ae872a56a227fd45776faffbd011e0b96200f` (retrospective, not pre-edit evidence).
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
