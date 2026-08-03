# Status — a073-operate-multi-market-strategy-lanes

- Updated: 2026-08-04
- Overall: IN PROGRESS
- Current wave: shared KR/US operational projection, Unix read transport, console and private REST/SSE/OpenAPI GREEN
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
| Strict OpenSpec validation | PASS |
| `git diff --check` | PASS |

## Pending

- Pre/post logic-map completion across the performance and Compose scopes (tasks 1.1, 1.2 and 5.1).
- Lane-performance lineage, fill/close accounting and conservation (tasks 2.5, 2.6 and 3.5).
- Compose immutable preimage, replacement and partial rollback guards (tasks 4.2, 4.4 and 4.5).
- Repository-wide gates, independent final implementation review and dormant deployment (tasks 5 and 6).

No market was activated, no entry runtime was started, and no LIVE order or operating setting changed.
