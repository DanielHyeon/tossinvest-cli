# Function Logic Map: `Gateway.submit`

- Source: `internal/execgw/gateway.go`
- AST evidence: `ast.json` (re-extracted Wave 2A 2026-09-25 at HEAD `648df8ef`; 466–763, 57 branches, 44 returns)
- Risk scan: `risk-pattern-report.md`

## Wave 2A re-extraction (2026-09-25)

The bundle was written on 2026-08-04 against the a066 q_final checkpoint (`a37d97f5`, 25 branches). Commit
`8022f578` (the squashed dormant KR/US strategy-lane landing, owned by a072 and archived with its own bundle
`archive/2026-08-29-a072-wire-multi-market-strategy-runtime/.../internal-execgw--gateway.submit`, whose
`ast.json` is byte-identical to this one) added 32 branches. `difflib` alignment of old/new branches by
`(kind, source line)`: every one of the 25 old branches survives unchanged —
old B1–B9 → new B6–B14, old B10–B15 → B19–B24, old B16 → B28, old B17–B20 → B30–B33, old B21 → B36,
old B22–B23 → B53–B54, old B24–B25 → B56–B57. Inserted (all `8022f578`): B1–B5, B15–B18, B25–B27, B29,
B34–B35, B37–B52, B55.

a066 owns two branch sites here, both calls to `checkReservation` for the ordinary (non-strategy) plan:
B28 (initial revalidation) and B36 (last-moment revalidation inside the dispatch callback). B41 is the same
barrier for the strategy plan and was added by `8022f578`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan/ref | normalized mutation plus durable Guardian reference | public Place/Cancel/Amend adapters and journal | mismatch/refusal settles NOT_DISPATCHED |
| exposure class | derived from mutation shape and the durable decision's `SafetyClass`, never a caller label | `mutationPlan.raisesExposure`, `journal.Decision.SafetyClass` | entry-only gates; risk-reducing decisions return early inside `checkEntry`/`checkReservation` |
| q_final authority (a066) | exact decision quantity, active owner, aggregate HELD hold and five HELD bucket reservations | journal rows via `RevalidateQFinalAdmission` | any mismatch/read error refuses before `plan.call`; no repair |
| strategy lease (8022f578) | exact CLAIMED/SUBMITTING lease and final authority check | journal strategy dispatch tables | not-sent refusal; claimed lease released through the pre-transport refusal path |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage at `648df8ef` |
|---|---|---|---|---|
| B1 | if at 475:3 | `if strategyCrossedTransport \|\| returnedErr == nil \|\| preTransportClaimFinalizationAttempted {`; then `return` (line last changed by `8022f578`) | not a066 | covered |
| B2 | if at 478:3 | `if preTransportClaim == nil && plan.strategy != nil && !preTransportClaimAttempted &&`; then `preTransportClaimAttempted = true` (line last changed by `8022f578`) | not a066 | NOT covered |
| B3 | if at 482:4 | `if claimRejected != nil {`; then `returnedErr = errors.Join(returnedErr, claimRejected)` (line last changed by `8022f578`) | not a066 | NOT covered |
| B4 | if at 488:3 | `if preTransportClaim == nil {`; then `return` (line last changed by `8022f578`) | not a066 | covered |
| B5 | if at 492:3 | `if !errors.As(returnedErr, &rejected) {`; then `rejected = reject(ReasonStrategyDispatchFenced, "%v", returnedErr)` (line last changed by `8022f578`) | not a066 | NOT covered |
| B6 | if at 506:2 | `if err != nil {`; then `return Outcome{Reason: ReasonInvalidRequest, Detail: err.Error()}, err` (line last changed by `d295555a`) | not a066 | NOT covered |
| B7 | if at 519:2 | `if !g.claimSymbol(symbolKey) {`; then `return Outcome{Reason: ReasonSymbolInFlight, Detail: "another mutation on " + plan.symbol + " is in flight"},` (line last changed by `9dfbfd03`) | not a066 | covered |
| B8 | if at 525:2 | `if rejected, err := g.checkSymbolFree(ctx, plan); err != nil {`; then `return Outcome{Reason: ReasonInvalidRequest, Detail: err.Error()}, err` (line last changed by `9dfbfd03`) | not a066 | covered |
| B9 | else at 527:9 | `} else if rejected != nil {`; then `return Outcome{Reason: rejected.Reason, Detail: rejected.Detail}, rejected` (line last changed by `9dfbfd03`) | not a066 | covered |
| B10 | if at 527:9 | `} else if rejected != nil {`; then `return Outcome{Reason: rejected.Reason, Detail: rejected.Detail}, rejected` (line last changed by `9dfbfd03`) | not a066 | covered |
| B11 | if at 535:2 | `if rejected, err := g.checkDecisionUnspent(ctx, prep, decision); err != nil {`; then `return Outcome{Reason: ReasonInvalidRequest, Detail: err.Error()}, err` (line last changed by `122985d9`) | not a066 | NOT covered |
| B12 | else at 537:9 | `} else if rejected != nil {`; then `return Outcome{Reason: rejected.Reason, Detail: rejected.Detail}, rejected` (line last changed by `658d4b0a`) | not a066 | covered |
| B13 | if at 537:9 | `} else if rejected != nil {`; then `return Outcome{Reason: rejected.Reason, Detail: rejected.Detail}, rejected` (line last changed by `658d4b0a`) | not a066 | covered |
| B14 | if at 545:2 | `if err != nil {`; then `return Outcome{IntentID: prep.Intent.ID, Reason: ReasonInvalidRequest, Detail: err.Error()},` (line last changed by `d295555a`) | not a066 | covered |
| B15 | if at 552:3 | `if plan.strategy != nil && preTransportClaim == nil && !preTransportClaimAttempted {`; then `preTransportClaimAttempted = true` (line last changed by `8022f578`) | not a066 | covered |
| B16 | if at 555:4 | `if claimRejected != nil {`; then `refused, refusalErr := g.refuseWithStrategySettlement(ctx, attempt, out, rejected, true)` (line last changed by `8022f578`) | not a066 | NOT covered |
| B17 | if at 561:3 | `if plan.strategy != nil && preTransportClaim != nil {`; then `preTransportClaimFinalizationAttempted = true` (line last changed by `8022f578`) | not a066 | covered |
| B18 | if at 564:4 | `if err := g.refusePreparedClaimedStrategyPreTransport(ctx, attempt, *preTransportClaim, rejected); err != nil {`; then `return out, errors.Join(rejected,` (line last changed by `8022f578`) | not a066 | covered |
| B19 | if at 583:2 | `if rejected != nil {`; then `return refusePrepared(rejected)` (line last changed by `171739a4`) | not a066 | covered |
| B20 | if at 588:2 | `if rejected := g.checkEntry(plan); rejected != nil {`; then `return refusePrepared(rejected)` (line last changed by `03a226bb`) | not a066 | covered |
| B21 | if at 594:2 | `if plan.preflight != nil {`; then `if rejected := plan.preflight(ctx); rejected != nil {` (line last changed by `0c332b80`) | not a066 | covered |
| B22 | if at 595:3 | `if rejected := plan.preflight(ctx); rejected != nil {`; then `return refusePrepared(rejected)` (line last changed by `0c332b80`) | not a066 | covered |
| B23 | if at 604:2 | `if loadRejected != nil {`; then `return refusePrepared(loadRejected)` (line last changed by `658d4b0a`) | not a066 | covered |
| B24 | if at 607:2 | `if rejected := g.checkDecision(decision, ref, plan, g.clk.Now()); rejected != nil {`; then `return refusePrepared(rejected)` (line last changed by `658d4b0a`) | not a066 | covered |
| B25 | if at 610:2 | `if plan.strategy != nil {`; then `preTransportClaimAttempted = true` (line last changed by `8022f578`) | not a066 | covered |
| B26 | if at 613:3 | `if rejected != nil {`; then `return refusePrepared(rejected)` (line last changed by `8022f578`) | not a066 | covered |
| B27 | if at 617:3 | `if rejected := g.checkClaimedStrategyLease(ctx, decision, prep.ClientOrderID, plan); rejected != nil {`; then `return refusePrepared(rejected)` (line last changed by `8022f578`) | not a066 | NOT covered |
| B28 | if at 625:2 | `if rejected := g.checkReservation(ctx, decision); rejected != nil {`; then `return refusePrepared(rejected)` (line last changed by `a6a396ab`) | a066: a066 initial q_final/aggregate-hold revalidation refuses before any broker call (true branch entered by both named tests, per-test coverprofile) | covered |
| B29 | if at 636:3 | `if plan.strategy != nil {`; then `strategyCrossedTransport = true` (line last changed by `8022f578`) | not a066 | covered |
| B30 | if at 648:3 | `if rejected != nil {`; then `return notSent(rejected)` (line last changed by `658d4b0a`) | not a066 | NOT covered |
| B31 | if at 651:3 | `if rejected := g.checkDecision(fresh, ref, plan, now); rejected != nil {`; then `return notSent(rejected)` (line last changed by `658d4b0a`) | not a066 | covered |
| B32 | if at 654:3 | `if fresh.ClientOrderID != prep.ClientOrderID {`; then `return notSent(reject(ReasonGuardianKeyMismatch,` (line last changed by `658d4b0a`) | not a066 | NOT covered |
| B33 | if at 658:3 | `if _, rejected := g.checkProtection(dctx, plan, protectionCheckpoint); rejected != nil {`; then `return notSent(rejected)` (line last changed by `171739a4`) | not a066 | covered |
| B34 | if at 665:3 | `if rejected := g.checkEntry(plan); rejected != nil {`; then `return notSent(rejected)` (line last changed by `8022f578`) | not a066 | covered |
| B35 | if at 672:3 | `if plan.strategy == nil {`; then `if rejected := g.checkReservation(dctx, fresh); rejected != nil {` (line last changed by `8022f578`) | not a066 | covered |
| B36 | if at 673:4 | `if rejected := g.checkReservation(dctx, fresh); rejected != nil {`; then `return notSent(rejected)` (line last changed by `8022f578`) | a066: a066 last-moment q_final barrier (ordinary plan): hold released after the initial check is refused before send tracking (true branch entered, per-test coverprofile) | covered |
| B37 | if at 685:4 | `if err := dctx.Err(); err != nil {`; then `return err` (line last changed by `8022f578`) | not a066 | NOT covered |
| B38 | if at 688:4 | `if plan.strategy != nil {`; then `if plan.strategy.finalAuthorityCheck == nil {` (line last changed by `8022f578`) | not a066 | covered |
| B39 | if at 689:5 | `if plan.strategy.finalAuthorityCheck == nil {`; then `return errors.New("strategy scheduler authority has no final source-backed revalidator")` (line last changed by `8022f578`) | not a066 | NOT covered |
| B40 | if at 692:5 | `if err := plan.strategy.finalAuthorityCheck(dctx); err != nil {`; then `return fmt.Errorf("strategy scheduler authority changed before transport: %w", err)` (line last changed by `8022f578`) | not a066 | covered |
| B41 | if at 695:5 | `if rejected := g.checkReservation(dctx, fresh); rejected != nil {`; then `return rejected` (line last changed by `8022f578`) | not a066 | NOT covered |
| B42 | if at 699:5 | `if requireTransport == nil {`; then `requireTransport = g.journal.RequireCurrentStrategyDispatchTransportAuthority` (line last changed by `8022f578`) | not a066 | NOT covered |
| B43 | if at 702:5 | `if err := requireTransport(dctx, plan.strategy.lease); err != nil {`; then `return fmt.Errorf("strategy owner or submitting lease changed before transport: %w", err)` (line last changed by `8022f578`) | not a066 | covered |
| B44 | if at 711:3 | `if plan.strategy != nil && !g.skipStrategyEntryABAForTest {`; then `if err := g.withStrategyEntryGateAuthority(dctx, plan.strategy.entryGateAuthority,` (line last changed by `8022f578`) | not a066 | covered |
| B45 | else at 717:10 | `} else {`; then `if err := call(); err != nil {` (line last changed by `8022f578`) | not a066 | covered |
| B46 | if at 712:4 | `if err := g.withStrategyEntryGateAuthority(dctx, plan.strategy.entryGateAuthority,`; then `return notSent(reject(ReasonStrategyDispatchFenced,` (line last changed by `8022f578`) | not a066 | covered |
| B47 | if at 718:4 | `if err := call(); err != nil {`; then `return notSent(reject(ReasonStrategyDispatchFenced,` (line last changed by `8022f578`) | not a066 | covered |
| B48 | if at 723:3 | `if tracker == nil {`; then `return notSent(reject(ReasonStrategyDispatchFenced, "strategy send tracker was not created"))` (line last changed by `8022f578`) | not a066 | NOT covered |
| B49 | if at 729:2 | `if plan.strategy == nil {`; then `res, err = attempt.DispatchVerified(ctx, dispatch, g.roundTripFor(plan))` (line last changed by `8022f578`) | not a066 | covered |
| B50 | else at 731:9 | `} else if g.dispatchStrategyVerified != nil {`; then `res, err = g.dispatchStrategyVerified(ctx, attempt, plan.strategy.lease, dispatch, g.roundTripFor(plan))` (line last changed by `8022f578`) | not a066 | covered |
| B51 | if at 731:9 | `} else if g.dispatchStrategyVerified != nil {`; then `res, err = g.dispatchStrategyVerified(ctx, attempt, plan.strategy.lease, dispatch, g.roundTripFor(plan))` (line last changed by `8022f578`) | not a066 | covered |
| B52 | else at 733:9 | `} else {`; then `res, err = attempt.DispatchStrategyVerified(ctx, plan.strategy.lease, dispatch, g.roundTripFor(plan))` (line last changed by `8022f578`) | not a066 | NOT covered |
| B53 | if at 736:2 | `if err != nil {`; then `if errors.Is(err, journal.ErrNonceSpent) {` (line last changed by `d295555a`) | not a066 | covered |
| B54 | if at 740:3 | `if errors.Is(err, journal.ErrNonceSpent) {`; then `return refusePrepared(reject(ReasonGuardianNonceReused,` (line last changed by `122985d9`) | not a066 | NOT covered |
| B55 | if at 744:3 | `if plan.strategy != nil && !strategyCrossedTransport {`; then `return refusePrepared(strategySubmittingRefusal(err))` (line last changed by `8022f578`) | not a066 | covered |
| B56 | if at 756:2 | `if res.Final == journal.StateConfirmed {`; then `return out, nil` (line last changed by `d295555a`) | not a066 | covered |
| B57 | if at 759:2 | `if res.Err != nil {`; then `return out, res.Err` (line last changed by `d295555a`) | not a066 | covered |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkReservation` | a066 q_final/aggregate-hold revalidation at B28 (initial) and B36 (last moment); B41 for the strategy plan | read-only; any error is a refusal; risk-reducing decisions return nil before any read | CodeGraph `callers checkReservation` = 1 (`submit`); per-test coverprofile |
| `checkEntry` | entry gate at B20 and again at B34 | exposure-raising only | current source |
| `checkProtection` | protection readiness at B19 path and B33 | refusal is not-sent | current source |
| `DispatchVerified` / `DispatchStrategyVerified` | durable dispatch transition around the broker closure | exactly once, never retried | current source |
| `plan.call` | the sole broker mutation | reached only after every final check | broker-spy tests |

The table above is a hand summary, not an enumeration of the 70 call sites CodeGraph reports.

## State mutations and fallbacks

- Journal Prepare/settle, send tracker and broker call ordering are unchanged by a066; a066 only adds read-only
  refusals (B28, B36) before the send tracker exists.
- Risk-reducing mutations never reach a q_final read: `checkReservation` returns at its B1 for any non
  exposure-raising decision (`TestAnExitNeedsNoReservation`, `TestACancelNeedsNoReservation`).

## Safety conclusion

- Safe edit boundary: unchanged — a066 inserted only conservative refusals before `plan.call`.
- High-risk impact: yes — sole broker boundary.
- Residual risk measured in Wave 2A: 16 of 57 branch bodies are not executed by `./internal/execgw` (untagged and
  `tossos_testseams`), including B41 (strategy-plan last-moment q_final barrier) and B30/B32 (fresh-decision
  reload refusal and idempotency-key drift inside the dispatch callback). None is an a066 site; B41 is an a066
  contract reached through the a072 strategy path.
