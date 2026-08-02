## Context

At 1440 CSS pixels, the StockOS reference table measures 1,152 px wide with header/body sizes of 10/12 px, 15/18 px line heights, 8 px × 12 px cell padding, and column widths of 355/72/122/130/201/143/129 px. Its holding row is 96 px high. The deployed TossOS table measures 1,115 px wide with 14.4 px headers and body, 23.04 px line height, 12 px × 10.4 px padding, seven equal 159 px columns, and rows up to 268 px high.

The equal-width defect is structural: `table-layout: fixed` derives widths from the first row, but a057 attached width rules to `.position-row` body cells only. The header row therefore fixes seven equal columns before body widths can participate.

## Goals / Non-Goals

**Goals:**

- Make the seven columns share one stable axis across header and body.
- Match the StockOS scan density without changing TossOS's evidence meaning.
- Keep concise safety state visible and fold only verbose supporting prose.
- Preserve the existing mobile card layout and native disclosure.

**Non-Goals:**

- Copying StockOS's dark visual theme or page shell.
- Changing management/adoption, reconciliation, exit-policy, broker, or journal behavior.
- Inferring missing exit prices or weakening stale/unknown evidence.
- Adding JavaScript, inputs, external assets, or new calls.

## Decisions

### 1. Define widths with `colgroup`

The table will declare seven columns at 31/6/11/11/18/12/11 percent. At TossOS's current 1,115 px content width this yields approximately 346/67/123/123/201/134/123 px, closely matching the StockOS measurement while preserving the full container width. Header and body necessarily share those widths.

Alternative: continue applying widths to body cells. Rejected because fixed table layout already committed the widths from the header row.

### 2. Scope compact type to the holdings table

Only `.positions-table` receives StockOS-like 10 px headers, 12 px body text, 18 px line height, and 8 px × 12 px padding. The rest of the console retains its current type scale. Numeric headers and cells align right with tabular numerals; the instrument and line columns align left/right according to their scan task.

### 3. Visible verdict, folded explanation

The primary row retains one management badge and one protection/evidence badge plus the five line values. Long reconciliation text, pending/excluded explanations, and exit reason prose move into the already-existing native `details` content. Missing/stale/mismatch remains unmistakable through the concise visible verdict and `—` values.

### 4. Preserve mobile cards

The existing 720 px breakpoint continues to replace table layout with labelled cards. Desktop-only column rules do not override the mobile `width: 100% !important` contract. The detail summary remains a 44 px target.

## Risks / Trade-offs

- [Compact text can become too small] → apply the 10 px size only to short uppercase-like column headers and keep body values at 12 px/18 px, matching the explicit StockOS reference.
- [A long management badge may wrap] → the instrument column grows to 31%; detailed reason prose is folded.
- [Folding prose could hide a safety state] → the visible badge still names blocked/pending/excluded and the line badge still names missing/stale/mismatch; only the explanation moves.
- [Dashboard and positions could drift] → both continue to call the same `holdingstable` template.

## Migration Plan

1. Add a rendering regression test and observe RED.
2. Add the explicit column grid and compact holdings-only styling; relocate verbose prose.
3. Run focused/full tests, SDD gate, and browser measurement at desktop/mobile widths.
4. Merge, push, rebuild the current container deployment, and verify both routes.

Rollback is a single commit/image rollback. There is no data migration.

## Open Questions

None. The user explicitly selected the StockOS positions table as the reference and requested a simpler input-free presentation.
