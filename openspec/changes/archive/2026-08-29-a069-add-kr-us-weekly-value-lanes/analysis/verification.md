# Verification evidence

## RED

The new package initially failed to compile because the schema, immutable plan, reservation, a066
authority, fill-risk, RR, evaluator and registry contracts did not exist. Tests were added before those
implementations. Because no existing function was modified, Function Logic Maps are not applicable.

## GREEN

Commands run from the dedicated `a064-multi-market-strategy-lanes` worktree on 2026-08-04:

```text
go test -count=1 ./internal/weeklyvaluelane
go test -race -count=1 ./internal/weeklyvaluelane
go test -run '^$' -fuzz=FuzzAllocateSevenConservesQuantity -fuzztime=2s ./internal/weeklyvaluelane
go test -run '^$' -fuzz=FuzzMissingFillRetryIsIdempotent -fuzztime=2s ./internal/weeklyvaluelane
go vet ./internal/weeklyvaluelane
```

All passed. The fuzz runs completed 663,807 allocation executions and 369,435 missing-fill retry
executions during the recorded runs.

## Adversarial hardening rerun

The independent review findings were first captured as compile-failing RED contracts in
`hardening_test.go`. GREEN then removed boolean activation, bound cap/FX freshness to evidence time,
scoped reservation state by campaign+market, enforced exact official-week and sequential ordinal
identity, sealed decoded evidence/stop/reservation inputs, and made positive-fill reservation+risk a
single aggregate transition. The following commands passed after the changes:

```text
go test -count=1 ./internal/weeklyvaluelane
go test -race -count=1 ./internal/weeklyvaluelane
go vet ./internal/weeklyvaluelane
go test -run '^$' -fuzz=FuzzAllocateSevenConservesQuantity -fuzztime=2s ./internal/weeklyvaluelane
go test -run '^$' -fuzz=FuzzMissingFillRetryIsIdempotent -fuzztime=2s ./internal/weeklyvaluelane
```

The hardening fuzz rerun completed 721,182 allocation executions and 477,291 missing-fill retry
executions. The selected evidence/campaign/risk/strategy/exit/scheduler regression set also passed.

A second independent review found that exported mutable RiskState fields/maps could be altered after
construction. RED reproduced both a filled-balance reset and latch-map clearing. GREEN made every field
private, added a sorted canonical seal, validated it in both admission and fill application, recalculated
it on every pure transition, and exposed scalar read-only accessors only. The second rerun passed package,
race and vet checks plus 674,056 allocation fuzz and 345,571 missing-fill retry executions.

The full-worktree rerun reached completion. `internal/weeklyvaluelane` and the previously incomplete
`internal/strategyrouter` both passed. The only failure was an unrelated concurrent journal change:
`internal/position.TestExitEligibilityHasOneSpelling` rejected new `entry_decision_id` spellings in
`internal/journal/position_campaign.go` and `internal/journal/strategy_evidence.go`. Those files are
outside a069 ownership and were not modified by this hardening pass.

The selected evidence/campaign/risk/strategy/exit/scheduler regression command also passed:

```text
go test -count=1 ./internal/strategyevidence ./internal/positioncampaign ./internal/risk/... \
  ./internal/riskbucket ./internal/riskcalc ./internal/strategy/... ./internal/strategyengine \
  ./internal/exitpolicy ./internal/scheduler ./internal/weeklyvaluelane
```

Strict OpenSpec validation and `git diff --check` passed. Repository SDD/gate evidence is recorded only
after those commands complete.

## Integration-time observations

`make sdd-sync` and one `make sdd-check` passed after the final RiskState seal. After final documentation
updates, the shared fingerprint check became stale again while hard `codegraph status .` still reported
the index up to date at HEAD `23794f8626a20691431d5452b76e800255b0ee74`. The change gate was invoked;
steps 1-3 passed and step 4 reported only missing Function Logic Maps for concurrent
`internal/journal/**` edits. The a069 logic-map checker emitted no weeklyvaluelane finding. Per ownership,
the journal maps, eligibility-spelling repair, full regression and final shared gate rerun remain with
that workstream/root integration.
