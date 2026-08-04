# Function Logic and Branch Test Map — isolated multi-market runtime core

This wave adds `internal/strategyruntime` only. Existing engine, journal, Gateway and scheduler functions are
unchanged; their maps and integration tests remain pending.

## Paired coordinator

1. Construct sealed KR and US workers in one release, both `OFF/UNOBSERVED`, with all safety classes on.
2. A sealed ready worker binds only its market calendar generation/digest, activation generation/digest,
   evidence cursor/digest and budget key.
3. Apply each market cycle independently. WAIT/BUDGET/refusal never becomes peer input.
4. Panic/abnormal/cycle failure latches only that market effective OFF, freezes its first typed refusal and
   advances saturating bounded restart state; invalid refusal and time overflow remain fail-closed.
5. Fill, reconciliation, protection, reduce-only exit and emergency reduction remain enabled.

Branches: paired default, KR wait/US allowed, one-market cycle input, abnormal return, peer-state preservation,
bounded restart and no combined authority.

## Lease claim and pre-transport validation

1. Issue only from complete sealed lineage, exact authority snapshot, owner epoch/token and bounded expiry.
2. `Claim` accepts only current `ISSUED`, sealed trusted time, exact scope/fence and byte-exact authority
   generations/digests. It returns expected/next revision for durable CAS.
3. Missing/drifted/expired/cross-market/stale-owner input consumes the lease as `REFUSED+RELEASED` with
   transport unauthorized and broker request count zero.
4. Valid claim becomes `CLAIMED`; a second validation marks `SUBMITTING` and only then returns a pure
   transport-authorization bit. The package itself has no transport dependency.
5. Terminal replay cannot alter the lease. A separately sealed exact retry HELD reservation may be released,
   with its ID and reservation-transition count reported independently.

Branches: all success states, activation/Guardian/protection/reconciliation generation drift, A→B→A digest
return with higher generation, scope mismatch, exact expiry boundary, owner restart fencing, terminal replay,
retry-only release and incomplete lineage.

Every identity crossing a future journal/adapter boundary is canonical UTF-8, non-empty, at most 256 bytes and
contains no whitespace/control character. Validation repeats after mint/seal so internal record corruption cannot
be made valid merely by recomputing the seal.

## Crash/outcome/reservation classification

1. Recover `CLAIMED` without a transport marker as `REFUSED+RELEASED`.
2. Classify `SUBMITTING` only with sealed exact operation ID, query/lookup/response digests, broker order ID where
   applicable, authoritative/complete flags, acceptance, fill count, pending/terminal presence and observed time
   no earlier than lease issue. Contradictory proof combinations are invalid.
3. Acceptance becomes `SUBMITTED+TRANSFERRED`; definitive rejection or authoritative not-submitted becomes
   `REFUSED+RELEASED`; transport uncertainty alone becomes `AMBIGUOUS+HELD`.
4. Missing outcome evidence cannot synthesize a terminal state.
5. Exact later reconciliation changes only HELD disposition to RELEASED or TRANSFERRED; lease remains terminal
   `AMBIGUOUS` and keeps its operation identity.
6. Same-key resubmit is limited to `AMBIGUOUS+HELD` and requires exact lease-bound current authority/owner,
   current exact protection generation/serial/digest, full broker
   client-key/lookup/identity/pending/terminal/cancel/dedup/idempotency capability and attempts below the cap.

An out-of-order nonterminal submit or classification invocation consumes the exact lease as
`REFUSED+RELEASED`; terminal replay remains immutable and `CLAIMED` crash recovery remains separately classified.

## Central safety fallback

1. Accept only sealed central fault evidence and a versioned manifest with `0 < RTO ≤ 60s`.
2. Require a replacement owner with strictly larger epoch and a different token before the deadline.
3. Successful plan is safety-only: all safety classes on, entry and lease issuance off, old owner fenced.
4. Invalid or late startup is `SAFETY_FALLBACK_UNAVAILABLE`, entry remains off, a critical alert persists and
   broker-resident protection is preserved.

Static and external-package tests reject broker/live transport and runtime writer imports, mutation symbols,
public authority constructors and one-market release descriptors. Fuzz properties prove terminal leases never
transition and matching digests cannot revive a lease after any generation change.
