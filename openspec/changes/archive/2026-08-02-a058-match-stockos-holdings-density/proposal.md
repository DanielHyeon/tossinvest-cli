## Why

The a057 holdings projection contains the right seven columns and facts, but its desktop table does not read like the StockOS reference. The header row determines seven equal columns because the widths are applied only to body rows, all table text renders at 14.4 px, and verbose management/exit reasons expand a normal holding row from StockOS's measured 96 px to as much as 268 px. Operators see columns and values on different visual axes and cannot scan the table quickly.

## What Changes

- Give the shared dashboard/positions holdings table an explicit seven-column desktop grid matching the measured StockOS proportions.
- Match StockOS's compact header/body typography, line height, padding, right-aligned numeric headers, and tabular numeral treatment.
- Keep the concise management and protection verdict visible, but move repeated or verbose management/exit diagnostic prose into the existing native details disclosure.
- Keep the current seven facts, canonical evidence semantics, KR/US parity, input-free operation, and 375 px card transformation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `operator-console`: tighten the holdings density and alignment contract so dashboard and positions render as one compact, truthful seven-column scan surface.

## Impact

- `internal/console/templates.go`: holdings-specific typography, spacing, fixed column grid, and numeric header alignment.
- `internal/console/templates_portfolio.go`: compact primary verdicts and verbose diagnostic placement.
- Rendering contract tests only; no Go handler, broker, journal, policy, order, setting, schema, or deployment topology change.
