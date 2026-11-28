# Function Logic Map: `TestTheDashboardReportsAnUnstartedMachineWithoutFailing`

- Source: immutable E blob `e65e394bf84b3c6e4559a219e816af96d341d75d:internal/console/console_test.go`, revision `base`, SHA-256 `45cf9c9260cd0dd2a91672339550cd9a601707c7967d4f5c60d2862239aac7f3` (retrospective, not pre-edit evidence).
- Deleted from current source: immutable E behavior only; no current runtime/execution claim.

## Inputs and invariants

This deleted function is immutable historical E evidence only, not a current runtime or execution claim. The E fixture created and authenticated a console harness with no records, got the verify-console page, and required onboarding words plus text stating that the console does not turn on a gate. It was an in-memory/local test harness, not a service or account binding.

## Branches and early returns

| B-id | E-source condition and effect | E focused-test behavior |
|---|---|---|
| B1 | Ranges required onboarding strings. | Iterates `soak`, `attestation`, and `tossctl soak run`. |
| B2 | Missing one required string records `t.Errorf`. | Requires all three strings. |
| B3 | Missing `게이트를 켜지 않는다` calls `t.Error`. | Requires non-enablement wording. |

## Calls and live bindings

Calls create/authenticate the harness and read page body. There is no retry, timeout, file output, service binding, account binding, or live action.

## State mutations and fallbacks

There is no mutation beyond test-local state; assertion methods are the fallback.

## Safety conclusion

The immutable E test records non-enablement wording in a local harness only and makes no current runtime or execution claim.
