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

Current hard evidence is frozen under `analysis/function-logic/`. The post-edit `Tracker.Observe` has 28 AST-enumerated
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

- Quantity key: kind `quantity`, non-empty canonical `Diff.AccountRef`가 canonical tracker account와
  일치하는 계좌, non-empty uppercased symbol, and promotion 전용 exact
  finite-decimal canonical local/broker values. 이 helper는 `riskcalc.CanonicalDecimal`처럼 float64를
  거치지 않고 `1`과 `1.0`은 합치되 blank, malformed, `NaN`, `Inf`는 `(key, false)`로 거절한다.
  `Comparer`가 `Diff`를 만들기 전에 수량 문자열을 float64로 반올림하거나 identical invalid strings를
  equality shortcut으로 받아들이면 promotion helper까지 fail-closed evidence가 도달하지 못하므로,
  snapshot/comparer의 공용 decimal canonicalizer도 finite decimal은
  `riskcalc.CanonicalDecimal`로 보존한다. blank를 zero로, malformed spelling을 원문으로 유지하는 기존
  가시성 계약은 유지한다. `quantitiesAgree`는 equality shortcut보다 먼저 양쪽이 finite decimal인지
  검증한다. invalid/non-finite는 문자열이 같아도 mismatch다. canonical strings가 다르지만 float64가 정확히 같은
  값으로 충돌하면 mismatch로 판정한다. 서로 다른 float64로 표현되는 문서화된 binary round-trip
  artifact(예: 짧은 canonical `0.3` 대 그 값을 계산 경로가 확장한 `0.30000000000000004`)만, 한쪽이
  짧은 decimal spelling이고 다른 쪽이 그 spelling의 binary round-trip 확장임을 증명하면서 larger
  magnitude에서 한 float64 ULP 이내일 때 일치로 허용한다. 단지 두 parsed 값이 한 ULP 이내라는 사실은
  증명이 아니다. `9007199254740992` 대 `9007199254740994`처럼 exact integer가 실제로 다르면 mismatch다.
  largest finite float에서 `Nextafter(+Inf)`로 계산한 ULP가 infinity가 되면 tolerance를
  허용하지 않고 exact canonical equality만 인정한다. 기존 `1e-9 * scale` relative bound는 실제 decimal
  차이를 숨기므로 폐기한다.
  raw broker/local quantity의 blank 또는 invalid 여부는 `canonicalDecimal`이 blank를 zero로 바꾸거나
  external-position branch가 실행되기 전에 보존·검사한다. 정상 positive broker-only holding은 계속
  external provenance이지만 unreadable 또는 long-only projection에 불가능한 negative broker-only quantity는
  ordinary fail-closed mismatch이고 promotion evidence는 아니다.
  이 보존은 직접 구성한 `Snapshot`뿐 아니라 production `Collector.holdings`의 `RawPositionsReader` 경계부터
  적용한다. raw holding quantity는 blank/invalid 원문을 유지해 `Comparer` validation까지 전달하며, digest나
  다른 optional decimal field의 기존 blank vocabulary를 전역 변경하지 않는다. 다만 holdings quantity의
  stabilisation digest는 같은 dedicated evidence vocabulary를 사용해 blank와 exact zero를 구별한다. 서로 다른
  두 account read를 동일 snapshot 증거로 세어서는 안 되기 때문이다.
- Missing-order key: kind `missing-order` plus all fields from `LocalOrder.Identity()`:
  account, market, market-local trading day, symbol, side and opaque order ID. A promotion-only validator
  reuses that method's canonicalization vocabulary but requires all six canonical components to be
  non-empty; an incomplete identity returns `(key, false)` and earns no permanent evidence. The order's
  canonical account and the comparison's non-empty canonical `Diff.AccountRef` must both equal the tracker's
  canonical account. An unknown/foreign snapshot cannot prove this account's order is missing. `OrderID` is checked for whitespace-only
  invalidity but its original opaque bytes are retained in the key; `"id"` and `" id "` are distinct IDs.

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

한 diff에 valid와 unclassifiable item이 함께 있으면 valid key만 streak를 얻는다. 비어 있는 account
또는 normalized symbol, 큰 정수의 float64
collision, non-finite 또는 malformed spelling, required field가 빈 missing-order identity는 서로 같은
영구승격 증거로 간주하지 않는다.

This last rule is fail-closed for exposure: the symbol stays blocked. It is conservative about escalation:
an operator-only account outage is not created from an identity the tracker could not prove.

빈 normalized symbol의 ordinary block은 `EntryGate`와 `Tracker.EntryAllowed` 모두에서 account-safe하게
cover하되, real account-permanent block과 in-memory key를 충돌시키지 않는다. journal의 empty-symbol
`QUANTITY_MISMATCH`로 쓰지 않는다. 그 durable 모양은 기존 schema/restore 계약상 account-wide permanent와
구분할 수 없기 때문이다. representable symbol scope가 없는 write는 error를 반환하고 in-memory pending
block/gate를 유지한다. 재시작 중에는 recovery gate가 첫 authoritative reconcile까지 진입을 막으며, 이
invalid ordinary observation을 operator-only permanent row로 제조해서는 안 된다.

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

단, write가 timeout/error를 반환했어도 journal에 커밋됐을 수 있다. continuity가 끊겨 pending을
철회하기 전 `ActiveReconcileStates` authority read로 같은 계좌의 durable account-wide quantity
permanent row가 없는지 확인한다. row가 있으면 pending을 durable projection으로 교체하고 gate를
유지한다. authority read가 실패하면 철회하지 않고 fail-closed 상태와 error를 반환한다. `Journal=nil`
테스트 tracker만 durable authority가 없으므로 메모리 pending을 직접 철회할 수 있다.

반대로 failed `EnterReconcile`가 반환된 직후에는 pending account block이 메모리와 account gate에
남아 있어야 한다. 그 다음 authoritative observation이 continuity를 끊어 철회하기 전까지 entry를
잠깐이라도 재개해서는 안 된다.

이 철회는 ordinary symbol block에는 적용하지 않는다. ordinary pending write는 현재처럼 diff가
바뀌어도 durable해질 때까지 fail-closed로 재시도하며 gate를 닫아 둔다. 이미 durable한 permanent
row도 절대 철회하지 않고 operator-only release를 유지한다.

ordinary와 account permanent가 동시에 pending이면 complete retry set은 deterministic하게 정렬하고
account permanent retry를 ordinary retry보다 먼저 journal에 시도한다. store가 첫 error에서 반환하는
현재 계약 아래에서도 continuity가 증명된 permanent retry가 map iteration order 때문에 굶어서는 안 된다.

authority read 실패로 pending permanent gate를 유지하더라도, 이미 관측된 key-loss/clean 사실 자체는
되돌리지 않는다. 즉 error 반환 전에 process-local streak와 retry continuity를 끊어야 하며, 나중에 earning
key가 다시 나타나도 이전 streak 또는 stale permanent proposal을 재사용하지 않는다.
그 관측 자체에 새 ordinary blocking item이 있으면 authority read보다 먼저 memory block과 gate를
fail-closed로 latch해야 한다. authority error 뒤 성공한 `Refresh`가 stale account proposal을 철회하더라도
그 현재 관측의 ordinary pending symbol block을 잃거나 진입을 다시 열어서는 안 된다.
같은 관측이 adjustment credit보다 later이고 그 credited symbol을 여전히 disputed로 보고하면, authority
error로 반환하기 전에 그 credit을 refuted로 만료해야 한다. 이후 `Refresh`와 우연한 clean read가 stale
credit을 재사용해 ordinary block을 자동 해제해서는 안 된다.

journal에 존재하지 않는 것으로 이미 알려진 blank-symbol ordinary pending block은 operator `Resolve`가
durable release request에 포함하지 않고, 다른 durable rows를 원자적으로 해제한 뒤 memory/gate에서 함께
제거할 수 있다. operator identity/note 요구는 그대로이며 실제 durable block의 audited release는 유지한다.

pending persistence의 deterministic priority는 account permanent → representable ordinary blocks →
unrepresentable blank-symbol ordinary block이다. 마지막 항목이 error를 반환해 memory/gate에 남더라도 valid
sibling ordinary rows는 먼저 durable해져야 하며, restart authority를 굶겨서는 안 된다.
같은 이유로 그 blank pending 오류는 adjustment credit과 later clean comparison이 이미 얻은 representable
sibling의 durable release를 굶겨서는 안 된다. representable additions와 earned releases는 journal에 먼저
반영하고, blank block 자체만 memory/gate에 pending으로 남긴 뒤 error를 반환한다.

`promotionDisputeKey` map 자체는 process-local이다. 다만 threshold를 얻은 한 key의 canonical fields는
운영자가 승격 근거를 식별할 수 있도록 permanent journal evidence 문자열에 durable하게 기록한다. 이는
retry map/key 구조를 복원하는 것이 아니며 opaque order ID는 기존 audited reconcile evidence로 취급한다.

### D4. `Failures` remains a derived compatibility view

`Outcome.Failures` and `Tracker.Failures()` continue to return one integer: the maximum live per-dispute
streak. Existing callers and diagnostics do not gain a new API. Tests that only repeat one dispute keep the
same values. The map owns semantics; the scalar is display/test compatibility only.

On restart, transient streaks와 pending promotion identity는 reconstructed되지 않으며, matching current
process-local failure-count behavior.
An already durable permanent account block is restored unchanged and may set the compatibility scalar to the
threshold as today. `Refresh` cannot manufacture a new streak.
Authority-read failure 때문에 fail-closed로 보존한 non-durable pending permanent proposal은 durable permanent가
아니다. 이후 성공한 `Refresh`가 authority에서 durable row 부재를 확인하면 그 pending proposal을
permanent로 재분류하거나 `Failures`를 threshold로 올려서는 안 된다. 관측된 continuity break는 그대로
유지하면서 non-durable proposal을 철회하고 ordinary pending blocks만 보존한다.

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
  observation; authority-read before withdrawing the non-durable account pending row when continuity breaks.
- **[Write committed but acknowledgement was lost]** → journal authority wins; retain/rebuild durable
  permanent and keep the account gate closed. Authority read failure is fail-closed.
- **[Ordinary and permanent writes both remain pending]** → deterministic account-permanent-first retry
  ordering prevents a repeatedly failing ordinary write from starving the required permanent retry.
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
