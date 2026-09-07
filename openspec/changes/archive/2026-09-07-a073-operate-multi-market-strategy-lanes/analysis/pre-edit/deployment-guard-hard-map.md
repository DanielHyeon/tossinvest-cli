# Pre-edit hard map — dormant Compose deployment guard

- Captured: 2026-08-04
- Base: `21200aaecff279a368634385101a763e2018d95b`
- Scope: a073 tasks 4.2, 4.4, 4.5 and read-only preparation for 6.x
- Safety boundary: no Docker replace/up/down/stop command is executed or implemented; no runtime, broker,
  journal, Guardian, protection or operating-setting writer is accepted.

## Memory and graph evidence

- File memory and TossOS GBrain search for Compose preimage, immutable digest, compatibility and partial
  rollback returned no prior result.
- `make sdd-sync` refreshed CodeGraph successfully. Advisory CodeGraphContext stalled after its database
  load and was interrupted; current files and CodeGraph hard evidence remain authoritative.
- CodeGraph and repository search found no existing deployment preimage, compatibility gate or bounded
  subset replacement planner. The only Compose production binding is `compose.yaml`; the only current
  deployment instructions are `docs/operations.md` and `docs/httpapi.md`.
- CodeGraph search hits named `Replace` are trading/protection operations and are explicitly out of scope.

## Existing Compose and deployment paths

| Path | Current fact | a073 gap |
|---|---|---|
| `compose.yaml` | two services, `tossos` then `httpapi`; both use mutable `tossos:local`; bind config/data; no Docker socket | no immutable current/target digest or compatibility preimage |
| `compose.yaml` health | console `/healthz`; API `/api/v1/engine`; bounded probe timeout | no paired projection/config/activation/protection preservation evidence |
| `docs/operations.md` | advises `docker compose build` and blanket `up -d`; rollback by old tag | mutable image and blanket replacement/rollback conflict with a073 |
| `docs/httpapi.md` | documents stopping only API and preserving engine | no exact digest/schema rollback compatibility gate |
| `cmd/tossctl/httpapi_static_test.go` | statically proves API service separation and no autostart command | does not prove immutable preimage, frozen order or bounded partial rollback |

## Planned pure boundary

- Add a new leaf package that accepts plain evidence only and returns a sealed preflight plus plain action
  values. It will not import Docker, process execution, HTTP broker clients, journal writers, engine control,
  config writers or protection mutation packages.
- The frozen preimage requires exact SHA-256 image and state digests, sorted environment keys, exact mount
  identity/mode, current/post schema versions, target and rollback read/write ranges and healthy baseline
  evidence for every service.
- The transition model emits one service action at a time in the frozen order with timeout `0 < t <= 5m`.
  A failed later step can inspect rollback compatibility and emit reverse-order rollback action values only
  for the successfully replaced subset. Incompatibility emits typed `ROLLBACK_INCOMPATIBLE`, preserves the
  new service and records entry `OFF`; it never emits a destructive rollback action.
- Post-step evidence must equal the frozen config/activation/lane/autostart/automation/LIVE/protection/
  journal, environment and mount facts. Drift is a failed health result, never a healthy deployment.

## Function Logic Map disposition

`Function Logic Map: not-applicable` for this wave before implementation: production behavior is added in a
new leaf package and existing Go function bodies are not edited. `compose.yaml` and operating documentation
are declarative/read-only surfaces. New leaf functions receive direct branch tests and risk scanning; if an
existing Go function becomes necessary, its AST and Function Logic Map must be captured before editing.

## Pre-edit safety declaration

- High-risk production order/risk/journal/reconciliation functions changed: no.
- Upstream behavior inheritance affected: no; the package has no runtime binding.
- RED tests before implementation: yes.
- Safety invariants: pass — no LIVE call, toggle flip, protection weakening, journal write or Docker action.
