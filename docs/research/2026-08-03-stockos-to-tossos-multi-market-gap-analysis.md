# StockOS → TossOS KR/US Multi-Market Strategy Gap Analysis

Date: 2026-08-03  
Feature: `FEAT-TOS-013`  
Status: OpenSpec proposal-freeze input  
Source input: user-provided StockOS strategy discussion (`pasted-text.txt`, 3,041 lines) plus the current TossOS code, specs, tests, and read-only local runtime observations

## Executive Finding

TossOS has most of the safety and transport primitives needed for automated trading, but it does not yet have a production strategy loop. The current console accurately shows that state: `/strategy-runtime` exposes one KR Parker descriptor, reports `strategy-runtime seam 미배선`, and keeps runtime observation, lane, autostart, entry, Guardian, reconciliation, protection, and activation evidence dormant or unavailable. `/strategy-runtime/market-schedule` knows KR and US as selectable scopes, but the scheduler is disabled with no selected market and no verified exchange calendar.

The missing work is therefore not a single connection flag. It is the full evidence-to-execution ownership chain:

```text
point-in-time evidence
  → market/horizon router
  → market-specific lane
  → campaign/leg state
  → multi-horizon risk cap
  → Guardian + protection readiness
  → durable dispatch lease
  → official execution gateway
```

KR and US SHALL be implemented in the same delivery wave. KR runtime stabilization is not a prerequisite for US lane design or implementation. The markets share contracts and safety infrastructure, but not evidence lineage, calendars, budgets, campaign ownership, readiness, or activation.

## What the StockOS Material Contributes

The user-provided material contributes four strategy contracts that are not present in TossOS today.

1. A strategy-neutral `PositionCampaign`/`CampaignLeg` core. It must not know the 8:4:2, 2:4:8, or seven-leg policy constants. It owns idempotence, partial fills, restart reconstruction, EXIT FIRST, and a non-retreating effective stop.
2. A continuation family using 8:4:2. Additional legs follow confirmed continuation/reclaim evidence, not fixed price percentages alone.
3. A reversal family using 2:4:8. The final and largest leg requires sweep, market-structure break, and reclaim; price decline alone is never sufficient.
4. A weekly value family with at most seven legs and at most one filled leg per week. Every leg revalidates point-in-time fundamentals, disclosure revisions, dilution, structure, and post-trade risk.

The material also fixes two important architecture mistakes seen in earlier StockOS plans:

- investor-flow evidence must be source-neutral; WTS can be optional enrichment but cannot be the sole required source;
- multi-horizon risk buckets must exist before live scale-in and must only reduce an upstream candidate quantity: `q_final <= q_candidate`.

These are adopted in a064–a070. The StockOS suggestion to move directly into small live orders is not treated as implementation authorization in TossOS. TossOS keeps LIVE approval, lane activation, autostart, and risk-limit changes behind explicit human approval.

## Current TossOS Baseline

### Present and Reusable

| Area | Current evidence | Reuse decision |
|---|---|---|
| KR/US candidates | `internal/candidate` models and CLI accept `MarketKR` and `MarketUS`; rankings, prices, candles, threshold scopes, and independent market descriptors already exist | Extend with normalized strategy evidence rather than create another candidate store |
| Market clocks | `internal/marketclock` and scheduler specs already model KR and US IANA time zones, regular sessions, DST, early close, and market-scoped activation | Reuse and add horizon routing plus independent worker/budget isolation |
| Pure strategy evaluation | `internal/strategyengine/ParkerConservativeLane.Evaluate` is a pure KR evaluator with extensive refusal tests | Preserve its safety ordering; add separate KR/US lane identities and evidence inputs |
| Guardian/risk | `internal/risk` has fixed-order Guardian checks, reservation authority, cost-aware KR/US RR, entry loss latches, and reduce-only handling | Add conservative market/horizon/sector/symbol caps; do not create a parallel sizing authority |
| Dispatch | `internal/strategydispatch.Dispatch` validates immutable decision provenance and official gateway submission dependencies | Reuse only after runtime lease revalidation; it currently has no production strategy caller |
| KR/US official order support | trading and verification paths support both KR and US official order contracts | Keep as the only submission boundary; no WTS order path |
| Scheduler and runtime descriptors | console/API already expose desired/effective/refusal concepts and preserve OFF defaults | Replace dormant single-KR projection with server-owned market-scoped projection |
| Exit/reconcile/protection code | position ledger, fill detection, reconciliation, exit policy, and protection saga primitives exist | Preserve their priority and authority; wire readiness only from attested broker capability |

### Missing or Unwired

| Gap | Hard evidence | Consequence |
|---|---|---|
| No production lane caller | CodeGraph/current-code lookup finds no production caller of `ParkerConservativeLane.Evaluate` | Candidate evidence never becomes a production strategy decision |
| No production dispatch caller | `internal/strategydispatch/dispatch.go` defines `Dispatch`, but no production strategy loop calls it | Even a valid lane decision cannot reach the official gateway |
| KR-only lane semantics | `internal/strategyengine/lane.go` mints the current Parker decision for KR scope and KRW lineage | No US strategy identity, refusal lineage, or attribution |
| No campaign/leg domain | the ledger represents position/fill/exit truth but not multi-leg campaign intent and ownership | Scale-in retry, restart, partial-fill, and competing-lane behavior are undefined |
| No multi-horizon buckets | Guardian has global/order limits but no horizon/market/sector/symbol intersection for a campaign | Short/weekly and KR/US scale-in can compete for the same risk or symbol |
| No point-in-time fundamental domain | no OpenDART/EDGAR adapter or immutable disclosure revision model exists | Weekly value decisions cannot be reproduced without future-information leakage |
| Protection readiness hardcoded unwired | `internal/execgw/protection.go` sets `ProfileProtection = ProtectionUnwired` | Exposure-raising runtime must remain blocked |
| Runtime seam dormant | read-only runtime response reports `UNOBSERVED`, lane/autostart/entry `OFF`, candidate/scheduler/Guardian/reconciliation missing, protection `UNWIRED`, activation absent | The console is a descriptor, not a running strategy engine |
| Scheduler dormant | market-schedule response reports scheduler `DISABLED`, no selected market, unverified calendar, `NOT_ACTIVATED` | Neither KR nor US evaluation is currently scheduled |
| No market-isolated operator projection | current page is titled “한국 주식 전략 lane” and only lists `krx_parker_vwap_conservative_v1` | Operators cannot distinguish KR/US evidence, readiness, campaign state, refusal, or performance |

## Separate Lane Model

TossOS SHALL not copy a single KR strategy and merely change the market code. It will expose six independently versioned lanes:

| Horizon | KR lane | US lane | Shared policy | Market-specific evidence |
|---|---|---|---|---|
| Short continuation | KR flow continuation | US participation continuation | 8:4:2, reclaim/continuation confirmation, stop non-retreat | KR investor-flow pressure vs US participation/price-volume evidence |
| Short reversal | KR absorption reversal | US dislocation reversal | 2:4:8, EXIT FIRST, final-leg structural confirmation | KR flow absorption vs US liquidity/dislocation structure |
| Weekly value | KR OpenDART weekly value | US EDGAR weekly value | max seven legs, max one/week, revalidation, structural stop cap | OpenDART revisions/XBRL vs SEC submissions/companyfacts |

Each lane has a unique `(market, lane_id, version)` and produces an entry decision, typed refusal,
or typed invalidation. An invalidation is evidence consumed by the common exit engine; only that
engine owns exit decisions. Lanes do not call brokers, write the journal, modify operating settings,
or activate themselves.

## Evidence Boundary

### Common Envelope

Every evidence item must preserve at least:

```text
market, symbol, source, source_record_id,
effective_at, source_available_at, observed_at, ingested_at,
revision_id, currency, quality_state, digest
```

Historical lookup applies two independent time gates:
`source_available_at <= evaluation_at` prevents a later-published/backfilled filing from leaking into
an earlier evaluation, and `ingested_at <= ingestion_cutoff` reproduces what this installation had
actually received. Evidence rows live in a bounded append-only `evidence.db`; the single-writer
trading journal stores only the immutable snapshot ID/digest consumed by a decision, so ingestion
backpressure cannot delay exit, reconciliation, or protection work.

Fatal evidence and lane scoring evidence are separate. Trading suspension, irreconcilable account state, unusable liquidity, or an attested fatal disclosure may veto all exposure-raising lanes. Valuation, flow, participation, or technical evidence that is not universal belongs to its consuming lane and must not silently become a common hard gate.

### Official Source Policy

- KR corporate disclosure/fundamentals: [OpenDART official API](https://opendart.fss.or.kr/guide/main.do?apiGrpCd=DS003), including receipt/report identifiers and revisions. The API key remains external configuration and must not enter logs or journal payloads.
- KR exchange statistics: [KRX Data Marketplace](https://data.krx.co.kr/contents/MDC/MAIN/main/index.cmd) for exchange-published investor and market statistics where an official bounded adapter is viable.
- US corporate disclosure/fundamentals: [SEC EDGAR data APIs](https://www.sec.gov/search-filings/edgar-application-programming-interfaces), including submissions and companyfacts with a compliant identifying User-Agent and access budget.
- Order submission: Toss official Open API gateway only. WTS is never an execution authority.

If a required source, FX conversion, revision identity, or freshness guarantee is unavailable, the affected exposure-raising lane returns a typed refusal. It does not synthesize a value. A provider outage in one market does not block eligible evaluation in the other market.

## Safety and Ownership Rules

1. One `(account, market, symbol, position_generation)` has at most one owning campaign/lane.
2. EXIT, stop, emergency reduction, fill detection, reconciliation, and protection supervision have priority over evaluation and scale-in.
3. A new leg cannot move the effective stop in the less protective direction.
4. `q_final` is the minimum of the upstream strategy quantity and every applicable risk cap. Missing or invalid arithmetic yields zero/refusal.
5. Entry-only loss latches and lane OFF states cannot be imported into or delay reduce-only paths.
6. Dispatch revalidates candidate/evidence digests, lane/version, activation, calendar, protection readiness, reconciliation health, risk generation, Guardian generation, and lease ownership immediately before official submission. A durable owner epoch/fencing token excludes stale replicas. Every lease claim ends in `SUBMITTED`, `AMBIGUOUS`, or `REFUSED`; drift refusal consumes the lease permanently, so A→B→A authority changes cannot revive it.
7. Deployment keeps lane desired state, autostart, automation gate, LIVE approval, and protection configuration unchanged and OFF/unapproved by default.
8. Protection readiness has only `WIRED` or `UNWIRED` plus a typed refusal. `WIRED` requires a current signed attestation whose trust root, key status, monotonic serial, scope, and exact broker identity/query/idempotency semantics are all valid; code presence is not evidence.

## Change Map

| Change | Closes gap | KR/US delivery rule |
|---|---|---|
| a064 | point-in-time, source-neutral evidence | common envelope; separate KR/US adapters and failure domains |
| a065 | campaign/leg ownership, idempotence, reconstruction | strategy-neutral and shared |
| a066 | horizon/market/sector/symbol caps | independent market buckets; fail-closed US FX until official contract is fixed |
| a067 | continuation strategy family | KR and US lanes implemented and tested in the same wave |
| a068 | reversal strategy family | KR and US lanes implemented and tested in the same wave |
| a069 | weekly value strategy family | OpenDART KR and EDGAR US lanes implemented and tested in the same wave |
| a070 | horizon routing and single owner | independent KR/US calendar, activation, rate budget, and refusal isolation |
| a071 | protection readiness | independent market attestations; no automatic live verification |
| a072 | production supervised loop | concurrent market workers, independent leases/backpressure, common official gateway |
| a073 | operator/API/deployment surface | market-scoped state and dormant deployment verification |

## Definition of Gap Closure

This feature closes the StockOS-to-TossOS gap only when all of the following are true:

- both KR and US have continuation, reversal, and weekly-value lane implementations with deterministic fixtures;
- every production decision has evidence and campaign lineage and every production attempt has a single durable dispatch owner;
- restart, duplicate event, partial fill, market outage, stale evidence, calendar closure, and OFF-state paths are tested;
- protection readiness is computed from current scoped attestation and actual wiring, not a code-presence assumption;
- the local console/private API show each market independently and never infer attribution from symbol/time proximity;
- Compose deployment is healthy with both markets dormant and no activation or live order mutation;
- `make sdd-check`, strict OpenSpec validation, change gates, independent review, build, merge, push, and post-deploy dormant checks pass.

Code completion and operational activation are different states. This feature can be implemented, merged, built, and deployed while entry remains OFF. Enabling a KR or US lane later requires a separate market-scoped human approval and does not automatically enable the other market.
