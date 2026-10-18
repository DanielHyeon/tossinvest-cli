# a087 task 0.5 upstream session census

2026-10-18; worktree /tmp/tossos-a148-impl-20261018. Read-only source inspection, no tests, source/spec edits, live requests or a157 work. Source hashes below bind this census. Existing a155 CodeGraph status was read (worktree correctly named), but it does not contain this complete chain; no fresh graph/AST execution or graph-derived completeness claimed. This is evidence for Manager task 0.5 disposition, not automatic checkbox completion.

## Source path

1. cmd/tossctl/engine.go:655 engineRuntime constructs observer via Context.ExitObserver at 671 and registers observer.Run as supervised loop `exit` at 734. Engine startup/automation/recovery requirements apply, but this registration is not conditional on market hours.
2. internal/app/engine/exitwiring.go:319 Context.ExitObserver requires Automation.Verified and ReductionIssuer, injects official prices, journal, Guardian, Gateway and exit-side retrier/floor. No calendar or regular-session input introduced.
3. internal/app/engine/runtime.go:288 Runtime.Run completes Recover first, then starts all supplied loops. Recovery failure blocks startup; no date/time-of-day filter selects the exit loop.
4. internal/app/engine/exitloop.go:395 ExitObserver.Run continuously calls ObserveOnce until cancellation and sleeps Interval. ObserveOnce:457 defers on fill-detection SLO, obtains workingSet, observes prices, then judges. Identity/protection/state/floor guards still apply. Timestamp use does not itself restrict market sessions.
5. exitloop.go:842 observe fetches Prices; ignores nonpositive/nonfinite or unwanted prices and rejects future/expired fetch evidence. Missing FetchedAt receives a local observation time. official/market_reads.go:142 Prices calls the price endpoint; adaptPrices:168 sets FetchedAt to time.Now().UTC, not exchange last-trade time. Therefore a newly retrieved unchanged after-hours price is not excluded merely by its session.
6. exitloop.go:1179 quoteUsable checks wall-clock rollback and monotonic QueryPriceEvidenceDuration. It is a local evidence lease, not proof of exchange-session membership. judge:961 consumes quote and cycle observation; snapshotContext:1191 binds position/generation/observation. record:1288 performs existing proposal persistence and submission flow.
7. exitloop.go:1485 IssueReduction authorizes bounded sells, attaches the intent, constructs sellIntent:1788 and calls Submit.Place at 1521. sellIntent uses observed price or baseline and emits LIMIT; no market-calendar argument.
8. execgw/riskguardian.go:587 IssueReduction uses scoped sell intent and risk.Evaluate. risk/chain.go:131 routes sells to evaluateReduction:227, whose quantity/held checks prevent shorts. Entry rules are not the reduction chain. Decision TTL is a lifetime constraint, not regular-hours permission.
9. execgw/gateway.go:358 Place builds plan, invokes ordinary safety/shape/idempotency/decision/transport checks. checkEntry:850 returns nil for !raisesExposure; protection.go:89 likewise exempts reductions. Thus entry schedule/readiness gates do not establish a protective-exit session restriction. failclosed.go:53 checkOrderShape rejects nonfractional MARKET today, permits valid LIMIT shape, no session branch.
10. trading/service.go:186 Place checks supported intent, sell/fractional policy, execute/confirmation/live guard and broker availability, then calls broker.PlacePendingOrder. placeIntentSupported:269 currently limits nonfractional KR/US to LIMIT. internal/app/engine/broker.go:23 officialBroker forwards to official.Client.PlaceOrder without hybrid/web fallback. official/orders_write.go:182 PlaceOrder builds body/account headers and invokes POST /api/v1/orders; no calendar decision in that call path.

Current repository-wide direct InRegularSession production use found only internal/verifylive/hours.go:130, besides definition internal/clock/market.go:146. This negative symbol search alone is not a universal proof; the positive path inspection above supports the narrower conclusion.

## Conclusion and unresolved D7

With runtime running, automation authorized, valid managed position, usable freshly fetched price, proposal/floor/decision and ordinary safety guards satisfied, reviewed code does not make the outside-regular-session exit path unreachable. Code does not prove that broker will accept it, execute it, or return fresh exchange data. Current whole-share MARKET is locally rejected by two shape/capability layers, independent of time. Changing those layers under a087 could expose the documented MARKET-unavailable/LIMIT-forbidden intersection; do not claim it empty.

Actual §5.1/§5.2 responses by KR regular/after-hours/SOR and US regular/extended sessions remain required under the frozen task prerequisites. Broker acceptance, partial fill, cancellation and authoritative terminal/flat reconciliation cannot be inferred from this census. Human D7 disposition may need a session-scoped order-type design, but this reviewer neither approves it nor changes the existing SHALL NOT. Phase 1 population-zero closure remains; Phase 2 sections 1–3 must wait for all specified measured/disposition/US prerequisites.

## a156 composed fixture constraint

Reuse existing campaign/journal tests, not new framework. Outstanding buy alone does not block later-leg admission universally. internal/positioncampaign/campaign.go:TransitionCampaign leaves normal progress ACTIVE, maps lineage/generation/watermark mismatch to RECONCILE, and AdmitExposure:220 rejects explicit closing/reconcile/closed/risk-reducing pending. PlanCampaignLeg and campaignExposureBlockedInTx also enforce entry-block/CAS/generation-scoped state. A useful composed fixture must use one of those explicit blockers across partial fill/restart and verify existing exit continues. Do not invent a qFinal formula or claim pending buy implies RECONCILE. Existing TestExposureAdmissionReadsPositionAndPendingRiskReduction and TestCampaignFillRollbackAndRestartAreDeterministic already cover components, so add only an uncovered composed transition after writer mapping.

## SHA-256

- `cmd/tossctl/engine.go`: `05babb65018119ec34dcd93b839568fa402c3be5a3f918e9b8bc7ef28ec3c88d`
- `internal/app/engine/runtime.go`: `485dbc57648357c420eb0e599581a6915e5b78fe5af045a2e87f0b41f1089496`
- `internal/app/engine/exitwiring.go`: `36e51db80cd3879305f2c7518552d81169c1ef35ab4e4ca99163b761d91e6453`
- `internal/app/engine/exitloop.go`: `ba30515476584fead62907157bb5ba985271057b2da882c6fda79e2b29951598`
- `internal/app/engine/broker.go`: `10216c3fd848715ef331c21720d0dc448329d3cd03986c627062df5b800f787c`
- `internal/official/market_reads.go`: `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea`
- `internal/execgw/gateway.go`: `b4b020607839345a70e7658ff488aef4282de4335eedd83b90a93520d80e53df`
- `internal/execgw/protection.go`: `71e4923e1301555808b3c65b437d1d20906f9d633d8eef52ac676a1433cd8267`
- `internal/execgw/failclosed.go`: `fbe97b79db74c5e7801fd3e1c6301d13f261d0d6c5e880308b36323ca328ebd4`
- `internal/execgw/riskguardian.go`: `2b33f38749c9a909d19ce016f8c276a667ecce2e625790bc55f1ef0d831b8cce`
- `internal/risk/chain.go`: `43b2ef76b276cc56f96071de62bea02df21bc509ea0cfc08db82a3f3729216fc`
- `internal/trading/service.go`: `631b2f02437e510698ffc06a1124bc11176d005a61956a61b4560e874d1e9f34`
- `internal/official/orders_write.go`: `d89e0bdbffdc2a41343dc195185286e258a1abda007d20d425d550439871bd71`
- `internal/clock/market.go`: `85264b6e9b4e6b13ddee690ca62e0d30cbd73efeb6926e626abc3d20c622f7a9`
- `internal/positioncampaign/campaign.go`: `94c1ceb5addda8bac4001e94dd079cc1f52eab5384112fccec8beb4495cf6495`
- `internal/journal/position_campaign.go`: `0168dc2a026488a43161803e4441b8d40c561a244c0a7043006ff29e8d1e078b`
