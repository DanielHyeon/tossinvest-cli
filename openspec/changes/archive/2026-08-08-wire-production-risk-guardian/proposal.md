## Why

`tossctl engine run` now reaches the automation interlock, but its production
assembly still omits the Guardian. A gate that is correctly configured therefore
always exits with `ErrGuardianRequired`, even though the engine package already
contains the real `RiskGuardian` and the exit path requires it.

This is a production-wiring defect in a high-risk path. The fix must construct
the Guardian only after the official account and writable journal have been
resolved, derive its policy from the exact audited gate limits, and prove that
the same instance reaches both the startup interlock and reduce-only exit
issuance.

## What Changes

- Add a production Guardian factory to the engine profile. Its default constructs
  one `execgw.RiskGuardian` after account resolution and journal opening.
- Derive the Guardian policy from all five configured gate limits and their one
  currency, with the per-trade risk budget conservatively equal to the configured
  daily-loss amount.
- Pass that one instance to the startup interlock, publish it on the engine
  context, and make the exit observer consume it as its `ReductionIssuer`.
- Preserve explicit test injection without allowing production to silently fall
  back to a nil Guardian.
- Add a command-level regression test that runs the real CLI assembly against an
  isolated config, real SQLite journal, and `httptest` official broker, then
  proves the verified context and exit observer are wired.
- Keep gate OFF, invalid limits, unresolved accounts, journal failures, and
  Guardian construction failures fail-closed. No test may contact a live broker
  or place an order.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `engine-safety`: Production startup assembly must create and inject the real
  Guardian after the account and journal exist, and the interlock must audit the
  same instance the runtime later uses.
- `risk-management`: Configured gate limits become the single policy source for
  the production `RiskGuardian`, including USD operation and reduce-only exit
  issuance.

## Impact

- `internal/app/engine`: construction order, Guardian factory/default, context
  publication, and identity/error-path tests.
- `cmd/tossctl`: real engine assembly seam and command-level regression test.
- `internal/execgw`, `internal/risk`, `internal/journal`: reused contracts only;
  no weakening of order, reservation, or reduction rules.
- Operational behavior: a correctly configured gate can pass the Guardian
  clause; all remaining interlock clauses and live-order approvals remain in
  force.
