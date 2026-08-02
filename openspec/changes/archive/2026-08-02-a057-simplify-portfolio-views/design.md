## Context

StockOS presents each holding as seven scan columns: instrument, quantity, average/current price, a compact line stack, market value, and unrealized PnL. TossOS already owns the required broker and canonical exit facts, but `/positions` mixes those facts with policy and journal diagnostics while `/dashboard` keeps only market aggregates. The current implementation also enriches lifecycle and exit evidence only in `handlePositions`, so copying rows to the dashboard without sharing that boundary would reintroduce the policy drift the operator previously reported.

Constraints are read-only operation, no new broker calls from `/dashboard`, no client script or external asset, no cross-currency totals, fail-closed exit evidence, and 375 px usability. KR and US must follow the same structure.

## Goals / Non-Goals

**Goals:**

- Put the requested holding and line facts in a stable primary hierarchy on both pages.
- Reuse one lifecycle/management/exit enrichment path.
- Preserve all current truthfulness, stale-evidence, rate-budget, and accessibility contracts.
- Reduce prose density with native per-row disclosure and no input.

**Non-Goals:**

- Recomputing exit policy or inferring a line from desired settings/raw journal values.
- Changing adoption, reconciliation, order, journal, API, or broker behavior.
- Rebranding TossOS to StockOS styling or introducing JavaScript.
- Combining KRW and USD totals.

## Decisions

### 1. Share the projection, not copied template conditions

Extract the current settings/runtime/lifecycle/exit enrichment in `handlePositions` into one console helper that accepts joined rows. `/positions` calls it after its lazy holdings refresh; `/dashboard` calls it for the rows already built from `holdings.peek`. This keeps dashboard at zero broker calls and makes policy/evidence parity structural.

Alternative: duplicate the handler logic in overview. Rejected because two desired/effective/lifecycle implementations would drift.

### 2. Carry joined rows in the existing account panel

`accountPanelFrom` already builds `positionRow` values to compute market aggregates. Retain those rows on `accountPanel` instead of joining a second time. This avoids a new data model and guarantees aggregate and row membership use the same snapshot.

Alternative: call `positions()` from overview. Rejected because it refreshes holdings and violates the dashboard's zero-call rate-budget contract.

### 3. One reusable holdings template

Define one template for the primary holdings table and invoke it from both pages. The primary row uses the StockOS information hierarchy while retaining TossOS typography/colors. Diagnostic/provenance content moves to `<details>`, but management, blocked/pending state, stale/unknown evidence, and page-level broker/journal notices stay visible.

Line semantics are fixed: next target = 익절, initial stop = 손절, next protection = 추적 회수, current protection = 기준, high-water = 고점. Unknown canonical fields render `—`; stored raw evidence stays in details with `실효 미확인` wording.

Alternative: make the whole row clickable with JavaScript. Rejected because native disclosure is keyboard-accessible, CSP-safe, input-free, and survives without assets.

### 4. Mobile card flow, desktop scan table

Continue using `.data-table` to preserve headers and the existing 720 px card transformation. Primary metrics get numeric alignment and compact line labels. The detail summary receives a 44 px touch target and visible focus. No horizontal page scrolling is required at 375 px.

### 5. Dashboard keeps overview context below holdings

The dashboard adds the shared holdings section directly after its summary and keeps existing engine, market aggregate, today, orders, safety, and recent sections. Those lower-priority panels are not removed in this change; the new holdings block becomes the first concrete portfolio surface.

## Risks / Trade-offs

- [Seven desktop columns can become dense] → compact line stack, tabular numerals, and the existing card breakpoint keep the scan path readable.
- [Native details closes on dashboard auto-reload] → secondary data only is folded; all requested and safety-critical facts remain visible. URL-driven persistence would add navigation noise for per-row details.
- [Legacy raw evidence lacks canonical targets] → show `—` in the primary lines and retain raw evidence only under `실효 미확인` details.
- [Dashboard projection reads settings/runtime each refresh] → these are existing local read seams and add no broker calls or mutations; failures remain unknown rather than defaults.

## Migration Plan

1. Add rendering contracts and observe RED.
2. Extract the shared enrichment boundary, retain overview joined rows, and render the shared table.
3. Run focused console tests, full SDD/gate checks, and desktop/mobile browser QA.
4. Deploy the same binary/container image. Rollback is a single commit/image rollback; no data migration exists.

## Open Questions

None. The user fixed the visible fields, input-free constraint, and StockOS reference; existing evidence contracts fix unknown/stale behavior.
