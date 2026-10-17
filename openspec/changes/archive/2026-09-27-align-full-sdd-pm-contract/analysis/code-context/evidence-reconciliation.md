# Evidence reconciliation

Date: 2026-07-31
Verdict: aligned after scope correction

| Claim | StockOS methodology | TossOS hard evidence | Resolution |
|---|---|---|---|
| Story ↔ active change is 1:1 | Mandatory, no permanent bootstrap bypass | 32 pre-existing active changes are allowlisted; this change has a Story | Backfill 32 Stories, then remove allowlist and require exact coverage |
| PM status is evidence-derived | proposal/tasks/archive are authority | Story `status` is manually stored and rendered | Store `intent`; derive status in generator |
| Full SDD includes READY and reconciliation | Explicit gates and artifacts | WORKFLOW abbreviates both | Restore full inputs, outputs, and block conditions |
| Pre-Edit applies before production edits | Explicit dependency/test/rollback declaration | WORKFLOW frames the detailed declaration mainly as high-risk | Require it for non-trivial production edits; keep high-risk non-waivable |
| Completion includes archive + PM | Required before completion report | WORKFLOW compresses gate/archive ordering | Make archive and PM sync explicit completion evidence |
| Project commands are shared | Method only | TossOS is Go/Toss API, StockOS is Python/KIS | Preserve TossOS commands and safety text unchanged in meaning |

## Reconciled implementation boundary

This change edits development governance and deterministic PM tooling only. It does
not alter production trading behavior, runtime configuration, LIVE approval, engine
autostart, Guardian limits, journal semantics, or container deployment.
