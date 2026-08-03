# Status — a073-operate-multi-market-strategy-lanes

- Updated: 2026-08-04
- Overall: IN PROGRESS
- Current wave: shared KR/US projection, pure lane-performance attribution and immutable deployment guard GREEN
- Runtime authority: read-only and dormant by default; no lane, activation, Guardian, Gateway, broker or operating-toggle writer

## Completed in the operational projection wave

- One versioned server-owned snapshot always contains the exact `KR` and `US` keys and validates every
  per-market lane, evidence, campaign/leg, horizon-risk, scheduler/calendar, activation, protection,
  reconciliation, typed first refusal and observed-at field.
- A single failed market becomes typed `UNKNOWN` with exact `UNWIRED` readiness and cannot erase, copy or
  default the peer market. Dormant output is exact OFF, unobserved/not-configured truth for both markets.
- The runtime-only Unix service is authenticated, bounded and strict about method, body, query, descriptor,
  schema and unknown fields. Its client exposes only `Read`; no preview/apply/order/activation/protection
  capability crosses the transport.
- `/strategy-runtime` now renders independent KR and US cards from the shared projection. It remains
  authenticated GET/HEAD-only, responsive and free of input, order, gate, LIVE, activation and protection
  mutation controls.
- `/api/v1/strategy-runtime`, SSE snapshot projection and OpenAPI use the same Go snapshot. OpenAPI declares
  an exact paired market object, strict object schemas and only the GET operation.
- A real Unix-socket integration fixture proves paired current state, US-only failure with KR preservation,
  stale-epoch SSE reconnect to a complete recovered snapshot and console/API value parity.
- Dormant health tests prove authenticated Unix connectivity, console/API schema health, both markets OFF and
  not configured, and zero console broker mutations.

## Verification checkpoint

| Check | Result |
|---|---|
| Four affected package tests | PASS |
| Four affected package race tests | PASS |
| Four affected package vet | PASS |
| OpenAPI and strategy-runtime contract tests | PASS |
| Deployment-guard package tests (`-count=1`, `-count=25`) | PASS |
| Deployment-guard race tests and vet | PASS |
| Existing Compose/API separation static test | PASS |
| Full performance package tests and race tests | PASS |
| Performance package vet | PASS |
| Strict OpenSpec validation | PASS |
| `git diff --check` | PASS |

## Completed in the deployment-guard wave

- `internal/deployguard` freezes exact `httpapi` → `tossos` current/target image digests, rendered
  Compose/config/activation/lane/autostart/automation/LIVE/protection/journal digests, canonical environment
  keys and mount identities, schema ranges, recent healthy baseline and exact dormant KR/US truth.
- Mutable image tags, missing services/evidence, reordered manifests or incompatible schema ranges produce no
  first replacement action. Rendered release images must match the frozen exact immutable targets.
- Each plain action carries an exact service/image, UTC issued-at/deadline and a positive timeout no greater
  than five minutes. Canonical observation SHA-256 binds that action window to image, schema, health and all
  preservation evidence; stale, future, replayed or post-seal-mutated observations cannot advance.
- Applied failures and timeouts reverse only the replaced subset, including the current applied attempt.
  Unhealthy/timed-out compatibility reads emit no destructive rollback; incompatible rollback retains the
  exact new digest with proven entry OFF. Drift recovery reports observed common `ON|OFF`, or `UNKNOWN` when
  KR/US or preservation evidence is incomplete/different; it never invents OFF. Rollback timeout becomes
  typed recovery.
- The package returns data only and has no Docker/process/engine/broker/config/journal/protection writer.
  Compose and operations documentation keep `tossos:local` development-only and require a separate
  digest-pinned release override.

## Completed in the lane-performance wave

- A new immutable derived view inside the already isolated `internal/performance` package consumes only
  caller-supplied authoritative position/cost-policy evidence and signed fill deltas. Existing
  `performance.db`, 90-day retention and bounded 500-row pruning remain unchanged.
- Attribution requires exact market, candidate, lane/version, campaign/leg, decision/attempt, order/fill,
  position/close/close-leg and policy/version lineage. Ticker is display-only; ticker-only queries require a
  market, and cross-lane/campaign/market corrections are refused.
- Event and fill replay is idempotent only for byte-equivalent evidence. Corrections are input-order
  independent, cite the exact original composite fill and cannot cumulatively reverse more quantity or money
  than that fill.
- Partial entry and staged close projection proves acquired = closed + authoritative residual quantity and
  total basis = allocated close basis + residual basis. Open positions and missing lineage/measurements remain
  `link_missing` or `not_measured`, never numeric zero.
- Source and reporting gross-to-net PnL retain entry/exit fee, tax, persisted FX cost, FX source/rate/as-of,
  quote currency and authoritative rounding policy/version. Missing fee/FX evidence does not invoke a current
  rate or an implicit same-currency rate of one.

## Pending

- Remaining repository-wide pre/post logic-map completion (tasks 1.1, 1.2 and 5.1); the new performance
  leaf has its hard map and changes no existing function body.
- Repository-wide gates and final independent implementation review (tasks 5.1–5.3).
- Actual immutable preimage collection/verification and dormant replacement/post-deploy checks (tasks 5.4
  and 6). No Docker or Compose mutation was run in this wave.

No market was activated, no entry runtime was started, and no LIVE order or operating setting changed.
