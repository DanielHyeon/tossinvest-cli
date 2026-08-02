## Why

The dashboard only summarizes holdings by market, while the positions page mixes prices and exit lines with lifecycle, provenance, and diagnostic prose. Operators cannot scan the facts they use most—position size, cost, current price, protection, value, and unrealized PnL—without reading a dense row or navigating between screens.

## What Changes

- Give `/dashboard` and `/positions` one shared, read-only holdings projection and information hierarchy.
- Keep name/symbol, quantity, average/current price, take-profit, stop, trailing recovery, baseline, high-water, market value, and unrealized PnL visible in the primary row.
- Move lifecycle provenance, journal/snapshot identities, detailed management reasons, and secondary diagnostics into an accessible per-row disclosure.
- Preserve unavailable/stale evidence and safety warnings above the fold; missing facts remain `—` with their reason rather than being inferred.
- Apply the same presentation to KR and US holdings without adding inputs, scripts, mutations, or cross-currency totals.
- Keep the holdings surface usable at 375 CSS pixels through the existing responsive card transformation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `operator-console`: tighten the holdings information hierarchy and require dashboard/positions projection parity while preserving read-only, evidence, accessibility, and rate-budget contracts.

## Impact

- `internal/console`: shared position-row enrichment, dashboard account projection, holdings templates, responsive styles, and rendering tests.
- No API, journal schema, broker call cadence, policy calculation, order path, or operating setting changes.
