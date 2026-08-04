# Full-suite static guard remediation

## Scope

The paired KR/US production path intentionally adds one second BUY intent
constructor, three additive weekly-reservation tables, build-tagged test-only
opaque-authority constructors, and production lane adapters that consume sealed
FX evidence. Existing structural tests predate those approved boundaries.

## Invariants

- The engine may spell BUY only in the legacy unreachable tracer and the new
  `strategyFirstLegPlaceIntent` path; the latter must still reach the official
  Gateway with a claimed durable lease.
- Test-seam files excluded from a production build must not be classified as
  production authority laundering. Production files remain fully audited.
- `strategy.ApprovedSnapshot` is the audited opaque sanitizer boundary. Importing
  the opaque value must not taint downstream packages back to raw candidate
  authority, while the strategy package itself remains subject to the strict
  pure-boundary audit.
- Pure lane packages must not acquire a transitive dependency on the concrete
  official client or configuration package. `officialfx` therefore accepts a
  minimal read interface instead of importing `internal/official`.
- Schema inventory must include all additive v27 weekly-reservation tables.

## Branch test map

| Branch | Expected verification |
|---|---|
| unexpected third BUY constructor | structural engine test fails |
| strategy BUY outside `strategyFirstLegPlaceIntent` | structural engine test fails |
| production candidate reader reaches authority package | boundary audit fails |
| test-seam-only constructor | ignored under normal production build constraints |
| lane dependency reaches official/config | strategyflow dependency test fails |
| missing v27 table | schema inventory test fails |

