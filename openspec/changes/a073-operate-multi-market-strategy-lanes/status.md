# Status — a073-operate-multi-market-strategy-lanes

- Updated: 2026-08-04
- Overall: COMPLETE — dormant KR/US deployment verified; activation remains a separate human action
- Current wave: exact-digest fix-forward deployment GREEN with existing engine safety loops restored
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

## Dormant deployment evidence

- The initial `8022f578` image replacement exposed one startup regression: the strategy-only exchange-rate
  read had entered the global startup attestation and stopped all engine loops. Replacement stopped, the
  replaced engine service was rolled back, and the old image then correctly refused the already-migrated
  journal (`v29` newer than its supported `v19`). Per the compatibility rule the new dormant image was kept,
  with both markets OFF, while the correction was reviewed and gated.
- Fix commit `171adda8` removed only that global dependency. US strategy entry still fails closed in its local
  FX authority; no fallback rate or activation authority was added. Independent review was CLEAN and the
  full a072 gate passed before the corrected image was built.
- Final fix-forward preimage had both services healthy on
  `sha256:dc2d94d9f745be412295497dd5dd57630a95e2b1cf4fcd76b835dc2e8f743fc0`, rendered
  Compose SHA-256 `d677169959b05c0ea9a7800d9b97d98fa41f6136ed9a86f0485e59450d574440`, config SHA-256
  `e0f6aa2dc7123a0035e5a01d0c673ca5c4d081df2be91ce68ebadb19d668af64`, active journal schema
  `v29`, `attempt_transitions=40`, `mutation_attempts=10`, and zero strategy/protection rows. No activation,
  lane or manifest file existed. Environment keys, mounts and volumes were frozen from the running services.
- `httpapi` then `tossos` were individually replaced in under ten seconds each with exact image
  `tossos@sha256:efba00b51e0d8ce55d48f4991ccd7a692cf670ec63a01c5a6193a6eebddbc6a3` from commit
  `171adda8`. Both became healthy. The private API reports the engine `running`; KR and US independently
  report lane/scheduler/activation OFF, `NOT_CONFIGURED`, protection `UNWIRED` and no campaign/leg.
- Post-deploy config SHA-256 is unchanged. Journal remains `v29` with the same general mutation counts and
  zero strategy/protection rows. The audit grew by exactly one `automation_gate.accepted` assertion with
  `old=true,new=true`; no operating setting, order, protection or activation mutation was recorded. Mounts,
  environment keys and activation-file absence are unchanged.

No market was activated, no strategy lease reached a broker, and no LIVE order, approval or operating setting
changed. Existing reconcile/exit/fill-detection safety loops resumed under the pre-existing autostart setting.
