# a072 strategyflow to q_final bridge — authority blocker map

## Scope and implementation decision

This checkpoint traced the normal-build production path from `strategyflow.Result` / sealed
`strategyflow.Lineage` to `execgw.QFinalEntryIssuance` for both KR and US. It changes no Go code,
runtime wiring, journal schema, activation, Guardian, Gateway, broker call or operating setting.
Consequently no existing-function Function Logic Map is required for this checkpoint.

The implementation decision is **map-only, fail closed**. An authority-complete bridge cannot be
constructed from the current production surface. Adding a public DTO or constructor that accepts
the missing values from an arbitrary caller would turn caller strings into price, policy, snapshot,
FX, account or owner authority. A Go bridge that always returns a refusal would add a second dormant
API without closing any missing authority, so this checkpoint does not add one.

## Hard call-path evidence

CodeGraph resolution found `strategyflow.Result`, `execgw.QFinalEntryIssuance`,
`riskbucket.Owner` and `execgw.CollectExposure`, but a trace from `strategyflow.Result` to
`execgw.QFinalEntryIssuance` returned no direct static call path. Current-source search confirms:

1. `strategyflow.Evaluate` has no normal-build production caller.
2. `QFinalEntryIssuance` is consumed only by `RiskGuardian.PrecheckQFinalEntry`,
   `IssuePrecheckedQFinalEntry` and `IssueQFinalEntry`.
3. The only request builder is test helper `qFinalKRRequest` in
   `internal/execgw/riskguardian_qfinal_test.go`. It fabricates fixed prices, account state, five
   bucket policies/snapshots, owner generation and exposure. It is not production authority.
4. Every a072 strategyflow descriptor remains `OFF/UNOBSERVED` in
   `internal/strategyflow/registry.go`.
5. The v25 journal persistence shapes are deliberately dormant. Every authority-bearing API returns
   `journal.ErrStrategyDispatchDormant`, and no production lease/market-authority mint exists.

The concrete current graph therefore stops here:

```text
ApprovedSnapshot -> strategyrouter.Route -> one pure KR/US lane evaluator
                 -> strategyflow.Result { Quantity, sealed Lineage }
                 -> [NO PRODUCTION CALLER]

QFinalEntryIssuance <- [TEST-ONLY qFinalKRRequest]
                    -> RiskGuardian.PrecheckQFinalEntry
                    -> RiskGuardian.IssuePrecheckedQFinalEntry
                    -> Journal atomic q_final decision + reservations
```

## Required authority inventory

| QFinal input / authority | Current source | Exact loss or mismatch | KR | US |
| --- | --- | --- | --- | --- |
| `QCandidate` | `strategyflow.Result.Quantity` after `Code == RefusalNone`, complete valid lineage | No production caller transports it, but the value itself is present | present | present |
| account / market / symbol | sealed `strategyflow.Lineage` | Values are present and integrity-checkable; no later durable authority generation is bound | partial | partial |
| lane / campaign / leg | sealed `strategyflow.Lineage` | Present, but no v25 durable lease mint or current-authority revalidation exists | partial | partial |
| owner prospective generation | `strategyflow.Lineage.PositionGeneration` is a router generation; a065 stores the actual authority in `position_campaigns.prospective_token` | `riskbucket.OwnerKey` requires that exact durable token. It must not be derived by stringifying or hashing the router generation | missing | missing |
| entry price | weekly request contains `EntryPriceMinor`; continuation/reversal outputs do not | `strategyflow.Result` and `Lineage` contain no entry price. The tagged-union `LaneInput` payload is private and is not returned | missing | missing |
| effective stop | continuation and weekly outcomes compute one; reversal validates a private candidate but returns none | `strategyflow.adapters.go` deliberately projects only quantity and lineage, discarding all effective stops | missing | missing |
| target price | weekly request contains `StagedTargetMinor`; other result shapes expose none | No lane outcome or `strategyflow.Result` carries the target required by QFinal | missing | missing |
| account state | private tracer `accountState` only | It assumes a flat account and fabricates zero exposure from tracer preconditions; there is no production strategyflow/QFinal account-state collector | missing | missing |
| `CollectExposure` | private `(*engine.Tracer).collectExposure` | It is inaccessible outside tracer, returns zero open exposure, and is tied to the legacy tracer flat-account path rather than authoritative KR/US gross exposure | missing | missing |
| reserve worst-price evidence | none | No normal-build adapter freezes an official executable quote with source/version/digest/observed/fresh-until | missing | missing |
| reserve fee policy | none | No normal-build versioned fee authority builds `riskbucket.FeePolicy` | missing | missing |
| five bucket policies | only pure `NewPolicyProvenance` constructor | No production authority reads/builds horizon, market, strategy, sector and symbol policies | missing | missing |
| five bucket snapshots | only pure `NewSnapshotProvenance` constructor | No production authority reads limit/filled/held, snapshot version/digest/freshness and matching journal references | missing | missing |
| exact owner claim | lineage has account/market/symbol/lane/campaign | Prospective generation is unresolved and there is no current durable active-owner read/build transaction | missing | missing |
| Guardian policy version / limits digest | methods on a concrete `RiskGuardian` | Obtainable only after a market-compatible Guardian has been selected; no strategy bridge wiring exists | partial | partial |
| official FX | `OfficialReads.ExchangeRate` -> `official.Client.ExchangeRate` | `domain.ExchangeRate` retains numeric rate but discards API `validFrom`/`validUntil`; no adapter freezes source/version/digest/freshness/haircut as `riskbucket.FXEvidence` | identity only | missing |
| existing Guardian cap | `PrecheckQFinalEntry` recomputes `risk.StrategyEntryQuantity` | QFinal requires Guardian `LimitCurrency == order quote currency`. There is no field or sealed authority for an officially converted existing cap | KR only when Guardian is KRW | US only with a separate USD Guardian; unavailable for one KRW account-wide Guardian |
| durable dispatch authority | v25 journal shapes | Every authority-bearing mutation is `ErrStrategyDispatchDormant`; no central owner validation/Gateway CAS path exists | missing | missing |

“Partial” is not sufficient for bridge construction. QFinal admission is an intersection: any missing
price, snapshot, policy, owner, currency, account or exposure authority must produce zero issuance.

## Lane execution-term loss

| Lane | Evaluator has enough information to validate | Normal outcome exports | Strategyflow preserves |
| --- | --- | --- | --- |
| continuation KR/US | effective stop plus sealed lane plan/cap | quantity, effective stop, lineage | quantity, lineage |
| reversal KR/US | stop candidate/saved stop are validated | quantity, action, lineage | quantity, lineage |
| weekly value KR/US | entry, staged target, effective stop and RR inputs | quantity, effective stop, lineage | quantity, lineage |

The bridge cannot recover discarded terms from lineage digests. A digest proves identity; it is not
the price preimage and must not be reverse-inferred or looked up by symbol/time.

## KR/US currency blocker

Production engine wiring constructs one `RiskGuardian` from one automation-gate `LimitCurrency`.
`PrecheckQFinalEntry` refuses when that currency differs from the exact market quote currency, before
risk-bucket admission:

- KR requires `KRW` and works only with a KRW Guardian.
- US requires `USD` and works only with a USD Guardian.
- A single current engine profile cannot authorize simultaneous KRW and USD entries with one
  Guardian.
- The reserve layer can represent quote-to-account FX, but the existing Guardian cap cannot. Passing
  converted bucket reserve while comparing raw USD prices to KRW Guardian limits would create false
  headroom, which the code correctly refuses.

Before a US production bridge exists, the design must choose and specify one sealed model: separate
market-currency Guardians whose account-wide limits are themselves authoritative, or an account-base
Guardian whose quantity/cash/notional calculations consume frozen official FX. The current API
implements neither model for concurrent KR+US.

## Prospective-token sequencing blocker

The missing owner value is not a formatting problem. a065 defines the prospective token as a durable
journal CAS authority minted while creating a `PositionCampaign`. `Journal.PositionCampaign` can read
the token back, but the current creation transaction requires an already-persisted exposure-raising
decision with immutable strategy lineage. Conversely, a066 q_final admission currently requires the
prospective token and campaign identifier before it can atomically persist its Guardian decision and
risk-bucket owner. The normal-build code has no transaction that resolves this ordering.

```text
CreatePositionCampaign
  requires persisted exposure-raising decision + strategy lineage

QFinal atomic issuance
  requires campaign ID + a065 prospective token
  creates the exposure-raising Guardian decision
```

Therefore the bridge must not invent a token from `Lineage.PositionGeneration`, candidate identity,
symbol/time, or a caller-provided random string. The production integration needs one reviewed atomic
ordering contract. The safe target is a journal-owned prepare/finalize protocol or one transaction that
reserves the a065 prospective token and binds campaign/leg, q_final decision, five monetary reservations
and a066 owner without exposing a reusable intermediate decision. Until that exists, both KR and US
owner authority remain unresolved.

## Minimum authority-complete implementation sequence

1. Define a sealed lane execution-term result that preserves the exact validated entry, effective
   stop and target preimages for all six KR/US bindings. Keep `strategyflow` pure and mutation-free.
2. Resolve the a065/a066 transaction ordering so the journal-owned prospective token is reserved and
   atomically bound to router owner, lane, campaign, q_final decision and market lineage. Do not derive
   or stringify a token from router generation and do not expose a reusable intermediate decision.
3. Implement a read-only authoritative risk snapshot service that returns exactly five versioned,
   frozen policy/snapshot attestations plus matching immutable journal references.
4. Implement official price, fee and FX evidence adapters that preserve raw decimal strings,
   source/version/digest and API validity bounds. For same-currency KR, freeze explicit identity FX;
   for US, freeze the specified official conversion and haircut policy.
5. Implement authoritative account state and retryable `CollectExposure` from current account and
   reservation-ledger reads; never substitute tracer flat-account zeroes.
6. Resolve the account-base versus per-market Guardian model, then expose a sealed market-compatible
   Guardian cap that QFinal can intersect without unit mismatch.
7. Only after 1-6, add the narrow bridge that consumes a valid complete `strategyflow.Result` and the
   opaque authorities above, calls QFinal precheck, and returns no broker capability.
8. Connect that precheck to the still-missing fenced central dispatch owner and v25 non-revivable
   lease transaction. KR and US evaluation remain concurrent; only mutation ownership is serialized.

Steps 1-6 can be developed as independent KR/US authority adapters where their sources are truly
independent. Step 7 is the first point at which an authority-complete QFinal request can exist. Step 8
must not activate either market; activation and LIVE approval remain separate human authority.
