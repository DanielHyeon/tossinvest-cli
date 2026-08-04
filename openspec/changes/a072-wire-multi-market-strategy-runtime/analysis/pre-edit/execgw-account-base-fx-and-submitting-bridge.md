# a072 execgw account-base FX and `SUBMITTING` bridge contract

## Status and scope

Status: **pre-edit design / paired RED target**, 2026-08-04.

This note refines the already-approved a072 contracts for `internal/execgw`. It does not mint a
dispatch lease, activate KR or US, change an operating toggle, call a broker, or change the journal
schema. KR and US are one delivery wave: every test matrix below contains both markets and neither
market can be declared complete independently.

## Current hard evidence

1. `RiskGuardian.PrecheckQFinalEntry` requires `Guardian.Policy.LimitCurrency()` to equal the market
   quote currency. A single KRW Guardian therefore admits KR/KRW and refuses US/USD before consuming
   monetary authority.
2. The opaque `officialfx.Evidence` is retained and revalidated at the Guardian clock, but today it
   feeds only `riskbucket.ReservePolicy`. Guardian sizing, chain exposure and the aggregate
   reservation are still quote-currency calculations.
3. `IssuePrecheckedQFinalEntry` stamps the configuration-only `g.limitsJSON`. The decision row does
   not identify the frozen FX evidence that produced its quantity, aggregate reservation and five
   monetary reservations.
4. `Gateway.checkReservation` proves a HELD aggregate reservation and exact current q_final owner
   plus five HELD bucket reservations. It cannot prove that all six reservations and the Guardian
   quantity used the same FX evidence.
5. `Gateway.submit` repeats decision, protection and reservation reads immediately before
   `plan.call`, but it has no strategy lease input. `Journal.BeginStrategyDispatchSubmitting` remains
   dormant, so no durable `CLAIMED -> SUBMITTING` fence exists in the Gateway callback.
6. `checkLimits` compares the order's raw quote notional with a base-currency maximum and rejects
   `limits.Currency != plan.currency`. A US q_final envelope therefore needs a distinct exact
   account-base branch; merely adding envelope JSON still fails closed before transport.

## Required contract

### Functional requirements

- **EG-FR-1** — One account-base-currency `RiskGuardian` MUST serve both KR/KRW and US/USD. A
  per-market quote-currency Guardian MUST NOT be used to split an account-wide limit.
- **EG-FR-2** — One request-scoped opaque `officialfx.Evidence` MUST be bound at the Guardian clock
  and retained unchanged through sizing, Guardian chain input, aggregate reservation, five-bucket
  reservation, decision envelope and Gateway pre-transport validation.
- **EG-FR-3** — Cash MUST remain in the market quote currency. Order notional, stop-risk budget,
  open exposure, daily loss, equity and all account-wide limits MUST be compared in the account base
  currency. Exposure-raising conversion rounds base reservation upward and quantity downward.
- **EG-FR-4** — A q_final first-leg decision MUST persist an exact versioned account-base FX binding
  in its decision limits envelope: quote currency, account currency, evidence source, version,
  digest and Guardian evaluation time. Scalar rate or haircut fields are not authority.
- **EG-FR-5** — A strategy Gateway request MUST carry only the opaque FX evidence and a lease CAS
  reference. The Gateway MUST read the durable decision/lease/binding and MUST NOT accept caller
  copies of market, owner, router, reservation, generation or FX digest as authority.
- **EG-FR-6** — Immediately before the first broker-send-capable instruction, the Gateway MUST:
  re-read the decision; revalidate protection and exact q_final holds; validate the opaque evidence
  at the Gateway clock against the decision envelope and order market/quote pair; and atomically move
  the exact current lease from `CLAIMED` to `SUBMITTING` with its transport-start marker.
- **EG-FR-7** — Missing, stale, forged, wrong-pair, cross-market, cross-decision, stale owner/fence,
  stale lease revision or non-`CLAIMED` authority MUST produce a typed not-sent refusal and exactly
  zero broker calls. No fallback rate, identity inference, market default or lease synthesis is
  allowed.
- **EG-FR-8** — Ordinary exits/cancels/reducing amends MUST retain their current path and MUST NOT be
  delayed by account-base FX or strategy entry lease checks.

### Minimal API shape

The exact names may follow the final `internal/risk` implementation, but the authority boundary is:

```go
// Existing request, extended without adding caller scalar authority.
type QFinalEntryIssuance struct {
    // existing fields ...
    FXAuthority officialfx.Evidence
}

// Persisted inside decision.limits_json only for q_final exposure-raising decisions.
// Rate and haircut are deliberately absent; EvidenceDigest seals their preimage.
type AccountBaseFXBinding struct {
    SchemaVersion   string    `json:"schema_version"` // exact supported version
    QuoteCurrency   string    `json:"quote_currency"`
    AccountCurrency string    `json:"account_currency"`
    Source          string    `json:"source"`
    Version         string    `json:"version"`
    EvidenceDigest  string    `json:"evidence_digest"`
    EvaluatedAt     time.Time `json:"evaluated_at"`
}

type StrategyPlaceRequest struct {
    Intent      orderintent.PlaceIntent
    Decision    GuardianDecision
    Lease       journal.StrategyDispatchLeaseCAS // exact CLAIMED revision/current owner
    FXAuthority officialfx.Evidence              // opaque, never reconstructed from JSON
}

func (g *Gateway) PlaceClaimedStrategy(context.Context, StrategyPlaceRequest) (Outcome, error)
```

The production engine must not call standalone `IssuePrecheckedQFinalEntry` and then try to attach a
campaign. The first-leg bridge is a separate opaque precheck whose mutation half calls only the
journal's atomic first-leg recorder:

```go
type QFinalCampaignFirstLegIssuance struct {
    Entry    QFinalEntryIssuance
    Strategy journal.StrategyPlanRequest
    Campaign journal.FirstLegCampaignRequest
}

type QFinalCampaignFirstLegPrecheck struct { /* package-private sealed fields */ }

func (g *RiskGuardian) PrecheckQFinalCampaignFirstLeg(
    QFinalCampaignFirstLegIssuance,
) (QFinalCampaignFirstLegPrecheck, error)

func (g *RiskGuardian) IssuePrecheckedQFinalCampaignFirstLeg(
    context.Context, QFinalCampaignFirstLegPrecheck,
) (journal.QFinalCampaignFirstLegReceipt, error)
```

The public request contains no prospective token and no router selector. Execgw writes the fixed
production `strategyrouter.RouterID/RouterRelease`; the journal mints the token inside its transaction.
The issue method calls `RecordQFinalCampaignFirstLegWithRecollection` directly and creates no
standalone decision, dispatch lease, execution-lineage `DISPATCH_START`, Gateway or broker capability.

`PlaceClaimedStrategy` is not a second broker path. It only constructs the existing `mutationPlan`
with a private strategy-dispatch capability; the actual call remains `Gateway.submit`. The private
plan data must be inaccessible to `Place`, `Cancel` and `Amend` callers. A q_final/first-leg
exposure-raising decision submitted through ordinary `Place` is refused for missing lease authority.

The risk boundary expected by execgw is an opaque result equivalent to:

```go
fx, err := risk.BindAccountBaseFX(now, market, guardianPolicy, opaqueEvidence)
quantity, err := risk.AccountBaseStrategyEntryQuantity(guardianPolicy, entry, stop, fx)
input.AccountBaseFX = fx
```

Execgw requires read-only accessors for the quote/base pair, source, version, digest and evaluation
time. It must not require or expose a constructor from public scalar fields.

### Decision envelope rules

- Legacy `Limits` JSON remains readable unchanged.
- The q_final strategy envelope is a separately versioned exact schema. Unknown versions and unknown
  fields fail closed.
- `ExpectedLimitsDigest` continues to identify the configuration-only Guardian policy. The persisted
  decision envelope additionally binds the request-scoped FX evidence; it must not change the
  startup/config digest on every quote.
- The aggregate reservation currency and all five `reserved_minor` values are the account base
  currency and are derived from the same bound evidence digest.
- Gateway comparison uses opaque-evidence validation at `Gateway.Now()` plus exact pair/source/
  version/digest equality with the persisted envelope. It does not require `EvaluatedAt` to equal the
  later Gateway clock; that field records the Guardian boundary, while freshness is re-evaluated.

## Paired RED matrix

| ID | KR case | US case | Expected result |
|---|---|---|---|
| EG-AC-1 | KRW identity evidence, KRW cash | USD/KRW official evidence, USD cash | same KRW Guardian admits both through one code path; base reservations use the same evidence as sizing |
| EG-AC-2 | stale identity snapshot | stale official rate or haircut policy | q_final refusal, zero journal writes, zero broker calls |
| EG-AC-3 | forged zero/public FX DTO | forged zero/public FX DTO | opaque authority refusal, zero journal writes, zero broker calls |
| EG-AC-4 | USD/KRW evidence on KR order | KRW/KRW identity or reversed KRW/USD on US order | exact pair refusal, zero broker calls |
| EG-AC-5 | KR decision with US owner/bucket/lease scope | US decision with KR owner/bucket/lease scope | cross-market refusal before transport |
| EG-AC-6 | decision envelope digest differs from opaque evidence | same | Gateway not-sent refusal, lease cannot reach `SUBMITTING`, broker count 0 |
| EG-AC-7 | current CLAIMED revision/current owner | current CLAIMED revision/current owner | final callback records `SUBMITTING` immediately before exactly one broker call |
| EG-AC-8 | stale revision/old owner/non-CLAIMED | stale revision/old owner/non-CLAIMED | typed fenced refusal, zero broker calls, no lease revival |
| EG-AC-9 | frozen FX expires after precheck but before issue | same | issue refuses with zero first-leg/decision/reservation rows |
| EG-AC-10 | frozen FX expires after issue but before Gateway | same | Gateway refuses before `SUBMITTING`/broker; existing safety loops and exits unaffected |

## Failure and crash ordering

```text
CLAIMED lease
  -> existing Gateway attempt RECORDED
  -> decision/protection/q_final hold re-read
  -> opaque FX current + exact envelope/pair/digest check
  -> Journal.BeginStrategyDispatchSubmitting(exact CLAIMED CAS)
  -> send tracker creation
  -> official broker call (at most once)
```

A crash before the `SUBMITTING` commit is provably pre-transport and recovers as
`REFUSED + RELEASED`. A crash after that commit requires the already-specified exact official broker
outcome classification; it must never be guessed as not sent. The journal outcome/recovery
implementation is outside this execgw-owned note.

## Out of scope

- Production identity/haircut/account/snapshot loaders or human activation authority.
- Journal schema and `BeginStrategyDispatchSubmitting` implementation.
- Broker outcome attestation and recovery settlement implementation.
- Lane/autostart/automation toggle changes, LIVE order execution or deployment.
