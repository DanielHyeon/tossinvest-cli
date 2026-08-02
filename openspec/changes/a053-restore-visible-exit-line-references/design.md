## Context

a052 already carries legacy journal facts in a separate `StoredExit` view, but the table still leads with `현재 보호선 —` and renders the useful baseline as a subordinate warning. For a position without any exit state, including a US candidate waiting on reconciliation, the table offers no protection reference beyond a dash.

The missing canonical snapshot and the missing exit state are different conditions. A legacy managed position has durable price evidence but not a new effective snapshot. A candidate has neither because adoption has not established t0. The UI must make both conditions useful without converting either into order authority.

## Goals / Non-Goals

**Goals:**

- Make stored baseline and initial-stop evidence immediately visible for legacy managed KR and US rows.
- Explain unestablished KR/US candidate lines and show the effective policy percentage without inventing a price.
- Preserve canonical freshness, lifecycle, market identity, and read-only behavior.

**Non-Goals:**

- Backfill or recompute canonical snapshots.
- Derive a hypothetical stop price from broker average/current price.
- Resolve reconcile blocks, adopt positions, place orders, or change operating settings.
- Make a breaking HTTP API change or add mutation/authority fields; the nullable additive reference field is in scope.

## Decisions

### 1. The table has three explicit evidence states

1. **실효 기준선**: a fresh canonical effective snapshot. Existing price/action fields remain unchanged.
2. **원장 기준선 · 실효 미확인**: a canonical snapshot is absent but same-lifecycle legacy exit-state price evidence exists. The table leads with the stored baseline and initial stop, while next target/protection remain unavailable and the values never populate `ExitLine`.
3. **기준선 미생성**: no exit state exists. For `ADOPTION_PENDING` or `RECONCILE_BLOCKED`, the row explains that t0 and price lines are established only after adoption and the first effective evaluation. If running settings are known, it shows the effective initial-stop percentage as policy context, not as a price.

Stale canonical snapshots retain the existing fail-closed `오래된 평가` state and do not promote raw prices into the primary reference stack. Released lifecycle evidence stays non-effective and is labelled as historical rather than current protection.

### 2. Market-neutral projection covers KR and US

The projector branches only on management/evidence state, never on market. Tests use both KR and US fixtures. No US price is derived from a prior closed generation or another market/symbol row.

### 3. A shared reference projector keeps console and API aligned

`operatorview.ExitLineView` is unchanged. A new pure `ExitLineReferenceView` produces `LEGACY_RAW`, `ADOPTION_PLAN`, or typed generation/runtime/lifecycle-unknown explanations. It never returns actionable prices and always reports `effectiveKnown=false`. The console and HTTP API both consume this projector; `StoredExitEvidence` remains as a compatibility field for allowlisted raw evidence.

Raw prices are permitted only for `legacy_snapshot_absent` and `not_evaluated_yet`. Partial, invalid, corrupt, or otherwise ambiguous snapshot states hide raw prices. If a lifecycle lookup was required but could not verify the row, prices are also hidden.

When position-policy lifecycle evidence is known, `exit_state.lifecycle_generation` must equal `positionpolicy.State.AdoptionGeneration` before either canonical or raw prices are exposed. A mismatch clears both projections and reports a typed reason. Legacy NULL generations already normalize to generation 1 at the journal read boundary.

## Risks / Trade-offs

- [legacy baseline mistaken for active order line] → use `원장 기준선 · 실효 미확인`, non-green styling, and keep actionable `ExitLine` fields unchanged.
- [candidate percentage mistaken for an exact stop] → show a percentage only, never a calculated price, and state that price lines are created after adoption.
- [stale canonical values hidden] → preserve the existing stale status and reason; do not widen stale display semantics in this fix.

## Migration Plan

No data migration. Deploy the console binary/image, verify KR legacy and US blocked/pending fixtures plus live read-only canaries, and roll back to the prior image if necessary.

## Open Questions

없음.
