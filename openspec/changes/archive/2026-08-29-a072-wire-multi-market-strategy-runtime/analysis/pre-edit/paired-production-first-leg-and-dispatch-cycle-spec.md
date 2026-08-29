# Paired production first-leg and dispatch cycle — pre-edit contract

Date: 2026-08-04

## Scope

This component wave connects KR and US together. Neither market may advance to
production assembly, testing, or release independently of its peer. China is out
of scope.

The cycle is:

`sealed q_candidate → account-base Guardian q_final → atomic campaign first leg
→ fenced dispatch lease → official ExecutionGateway`.

The deployed default remains dormant. This change does not create activation
files, change desired/autostart state, approve LIVE execution, or invoke a broker
mutation.

## Paired price-unit boundary

- KR and US are implemented, tested, and gated in the same wave. There is no
  KR-first/US-later dependency.
- Strategy evidence remains canonical quote-minor integers. Guardian, journal
  `RiskIntent`, and official order intent use canonical major-currency decimals:
  KRW scale 0 is identity; USD scale 2 maps `5000` to `50`, `5050` to `50.5`,
  and `5` to `0.05` without floating-point arithmetic.
- Existing `journal-strategyflow-risk-binding:v2` records keep historical
  literal-minor verification. New projections use v3 and bind sealed minor
  evidence plus major-decimal executable prices without rewriting history.

## Authority rules

1. A production proposal remains the only lane input. A caller cannot construct
   q_final, prospective generation, weekly reservation, dispatch lease, or
   protection authority.
2. The production account/exposure adapter must consume a frozen, externally
   attested official snapshot. Missing daily loss, equity, cash, open exposure,
   pending-order, operating-mode, entry-latch, or symbol facts refuse before the
   Guardian transaction.
3. The five-bucket bundle is converted by value-copy into the Guardian admission
   request. Exact scope, five unique keys, policy, immutable references and
   official FX are revalidated; q_candidate is overwritten from the sealed
   proposal.
4. KR and US weekly lanes carry the exact schema-v27 durable market/week
   reservation through the opaque Guardian precheck. Non-weekly lanes must carry
   no weekly binding. The journal validates this binding in the same first-leg
   transaction.
5. A dispatch lease can be issued only after the exact first-leg receipt exists,
   the current central owner is fenced, signed activation/calendar and current
   protection readiness have been revalidated, and the Gateway operation key is
   already journal-bound. Lease validation failure is terminal and performs zero
   broker requests.
6. Workers never own a broker mutator. They enqueue an immutable plan to the
   single dispatch owner. That owner alone performs `ISSUED→CLAIMED` and calls
   `Gateway.PlaceClaimedStrategy`.

## Paired failure matrix

| KR | US | Required result |
|---|---|---|
| valid | valid | independent admissions may commit in either order; central dispatch serializes transport |
| invalid account/evidence/risk/protection | valid | KR refuses locally; US progresses if its own complete authority is current |
| valid | invalid account/evidence/risk/protection | US refuses locally; KR progresses if its own complete authority is current |
| either OFF | peer valid | OFF market creates no first-leg/lease; peer is unaffected |
| central owner/integrity invalid | any | both entries close; safety-only loops continue |

## RED requirements

- KR and US weekly opaque prechecks preserve their own reservation and reject
  cross-market, divergent, missing, or non-weekly-smuggled bindings before any
  first-leg row commits.
- The production first-leg adapter covers all six lanes, q_final never exceeds
  q_candidate, and one market's collector failure does not cancel its peer.
- Lease/Gateway tests use a fake official broker and prove exactly zero requests
  for missing activation, protection, owner, first-leg, reservation, FX, current
  authority, OFF state, or consumed lease.
- Restart/race tests prove no duplicate submit and non-revival of every consumed
  lease.
