# a110 — Design

## Context

`Tracker.Observe` currently stores one `failures int` for the account. Every comparison for which
`Diff.BlocksEntry()` is true increments it, regardless of which quantity or missing-order dispute made
that comparison blocking. At three it writes an account-wide, operator-only permanent block.

The 2026-08-07 incident is the counterexample. Three successive cycles each had a blocking quantity
diff, but the durable symbol blocks show changing symbols and the first two later released through
`ADJUSTMENT_APPLIED`. The third observation nevertheless promoted the shared scalar. From then on,
`ReconcileDriver.blocked` returned true for every candidate before quote collection; the three current
unmanaged positions therefore have no adoption record, no exit state and no canonical exit snapshot.
The console's `—` lines are truthful consequences, not a price-calculator failure.

Current hard evidence is frozen under `analysis/function-logic/`. `Tracker.Observe` has 21 AST-enumerated
branches/ranges and writes the gate both before and after persistence. That durable-first ordering is not
part of the defect and must not move.

## Goals / Non-Goals

**Goals:**

- Make “three consecutive failures” mean three consecutive observations of the same canonical blocking
  dispute.
- Preserve first-observation symbol blocking and every existing release/persistence invariant.
- Preserve the existing public `Failures` observation as a scalar without using it as identity.
- Pin the production incident shape, exact quantity identity and canonical missing-order identity.

**Non-Goals:**

- Automatically release any permanent block, including the one currently active in production.
- Change automatic-adoption settings, candidate rules, stop percentages or exit policy math.
- Change broker calls, orders, schema, reconciliation cadence or the account-wide scope of a legitimately
  earned permanent promotion.
- Add a web reconcile mutation surface. The existing local audited recovery command remains the only door.

## Decisions

### D1. Track streaks by a comparable canonical dispute key

Add an unexported comparable key with a kind and fixed fields, and keep
`map[promotionDisputeKey]int` in `Tracker`.

- Quantity key: kind `quantity`, trimmed account reference, uppercased symbol, and promotion 전용 exact
  finite-decimal canonical local/broker values. 이 helper는 `riskcalc.CanonicalDecimal`처럼 float64를
  거치지 않고 `1`과 `1.0`은 합치되 blank, malformed, `NaN`, `Inf`는 `(key, false)`로 거절한다.
  snapshot digest/comparer의 기존 float vocabulary는 a110에서 바꾸지 않는다. 승격 identity는
  이미 만들어진 `Diff` tuple을 exact하게 분류하며, 상위 comparer의 수치 허용 계약을 넓히지 않는다.
- Missing-order key: kind `missing-order` plus all fields from `LocalOrder.Identity()`:
  account, market, market-local trading day, symbol, side and opaque order ID. A promotion-only validator
  reuses that method's canonicalization vocabulary but requires all six canonical components to be
  non-empty; an incomplete identity returns `(key, false)` and earns no permanent evidence.

The key is process-local and is never logged or persisted. A comparable struct avoids delimiter ambiguity
and does not invent a second order-identity vocabulary.

Alternative rejected: key by `Block.Key()`. Quantity blocks intentionally collapse to symbol scope, so a
new local/broker tuple on the same symbol would inherit an old streak. Missing orders also need the complete
canonical order identity that a symbol block discards.

Alternative rejected: fingerprint the whole diff. One persistent dispute would lose its streak whenever an
unrelated dispute joined or left the set. Per-dispute state answers the actual question: which disagreement
has survived repeated authoritative reads?

### D2. Consecutive means present in this and the immediately previous blocking comparison

For each blocking observation:

1. derive the current unique promotion keys;
2. build a fresh streak map containing only those keys;
3. set each count to `previous[key] + 1`;
4. replace the old map, so an absent key loses its streak;
5. derive the compatibility scalar `failures` as the maximum current count.

A non-blocking comparison clears the map and scalar. Duplicate entries inside one diff count once. A changed
quantity tuple is a new key and begins at one. If canonicalization unexpectedly fails, the ordinary block is
still entered immediately but that item earns no permanent-streak increment; inability to prove sameness
cannot be converted into evidence that it is the same dispute.

한 diff에 valid와 unclassifiable item이 함께 있으면 valid key만 streak를 얻는다. 큰 정수의 float64
collision, non-finite 또는 malformed spelling, required field가 빈 missing-order identity는 서로 같은
영구승격 증거로 간주하지 않는다.

This last rule is fail-closed for exposure: the symbol stays blocked. It is conservative about escalation:
an operator-only account outage is not created from an identity the tracker could not prove.

### D3. Promotion is earned when any one current key reaches the existing threshold

If any current streak reaches `MaxFailures` and no permanent block already exists, reuse the existing
account-wide `QUANTITY_MISMATCH`, `reconciliation_mismatch_permanent`, `ReleaseOperatorOnly` block and
durable write path. The evidence text states that one canonical dispute remained for N observations; it
must not claim that arbitrary account comparisons all disagreed about the same fact.

The promotion remains account-wide. A truly stuck mismatch is evidence that the reconciliation process needs
human attention, and changing that blast radius would be a separate safety decision.

### D3-1. Failed permanent persistence may retry only while its earning key is consecutive

Account-wide permanent `EnterReconcile`이 실패한 경우 pending permanent에는 승격을 얻은 canonical
key를 함께 보관한다. 바로 다음 blocking comparison에도 그 key가 존재할 때만 기존 durable-first
write를 재시도한다. 다음 관측이 clean이거나 그 key가 사라졌다면 journal에 기록되지 않은 pending
account block과 그 retry identity를 철회한다. 다른 current key가 독립적으로 threshold에 도달했다면
철회 뒤 그 key의 새 승격으로만 다시 제안할 수 있다.

이 철회는 ordinary symbol block에는 적용하지 않는다. ordinary pending write는 현재처럼 diff가
바뀌어도 durable해질 때까지 fail-closed로 재시도하며 gate를 닫아 둔다. 이미 durable한 permanent
row도 절대 철회하지 않고 operator-only release를 유지한다.

### D4. `Failures` remains a derived compatibility view

`Outcome.Failures` and `Tracker.Failures()` continue to return one integer: the maximum live per-dispute
streak. Existing callers and diagnostics do not gain a new API. Tests that only repeat one dispute keep the
same values. The map owns semantics; the scalar is display/test compatibility only.

On restart, transient streaks와 pending promotion identity는 reconstructed되지 않으며, matching current
process-local failure-count behavior.
An already durable permanent account block is restored unchanged and may set the compatibility scalar to the
threshold as today. `Refresh` cannot manufacture a new streak.

### D5. Release authority does not change

Clean comparisons never release a permanent block. The production block must still be handled by:

1. stop the engine and acquire its exclusive journal lock;
2. collect three identical official snapshots at the bounded spacing;
3. reconstruct local state and require a clean blocking diff;
4. require explicit operator identity and note;
5. commit the exact-cause journal release, then restart the engine.

a110 may document and test this handoff but may not execute it without a new human approval containing the
operator identity and note.

### D6. Verification follows the causal chain, not only the counter

The integration regression replays changing quantity disputes across three cycles, gives the earlier
symbols their normal adjustment/recheck releases, and proves:

- no account-wide permanent block exists;
- ordinary blocks still latch and release durably;
- an uncovered automatic-adoption candidate reaches quote/adoption after the ordinary blocks clear;
- adoption persists its t0 and opens exit state as `SEED`;
- the intermediate operator view remains non-actionable (`not_evaluated_yet`/`—`);
- a separate exit-observer judgement persists an evaluated canonical exit snapshot, after which only the
  UI/API may render actionable prices.

No test calls a LIVE order path. The existing one-share policy behavior (no intermediate take-profit target)
is not changed or treated as a failure.

## Risks / Trade-offs

- **[A real stuck condition changes numeric spelling]** → package canonical decimals keep equivalent values
  in one streak; genuinely changed quantities start new evidence while ordinary blocking remains.
- **[Busy accounts may take longer to become account-wide permanent]** → this is intended. Each current
  symbol remains blocked from its first mismatch; only the operator-only account escalation requires proof
  of the same unresolved dispute.
- **[Map state grows]** → replacement on every observation bounds it to the number of current blocking items.
- **[Missing-order identity drift]** → use `LocalOrder.Identity()` rather than rebuilding its normalization.
- **[Permanent write fails at the threshold]** → bind retry to the earning key's next consecutive blocking
  observation; withdraw only the non-durable account pending row when continuity breaks.
- **[Persistence regression while editing a large function]** → retain the mapped B8–B21 region unchanged,
  run focused mutation/race tests, and have a separate adversarial reviewer inspect gate ordering.

## Migration Plan

1. Land code and tests; no schema or configuration migration.
2. Build and deploy the image. Existing permanent block remains active after restart by design.
3. With separate human approval, run the existing audited reconcile recovery while the engine is stopped.
4. Restart and verify `blocked=0`, adoption of the three candidates, exit-state/snapshot creation and browser
   lines. A one-share managed position may correctly keep its intermediate take-profit as `—`.
5. Rollback: restore the previous image. No data downgrade is needed; any permanent block or ordinary block
   remains journal-authoritative. Do not roll back after a human release without retaining its audit record.

## Open Questions

None for implementation. The current production release remains a later human approval gate, not an a110
design choice.
