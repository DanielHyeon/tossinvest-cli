# Paired KR/US production dispatch-lease cycle — pre-edit specification

## Scope

This wave connects the already atomic `q_final + campaign + first-leg` commit to the existing
official `execgw.Gateway` path. KR and US are implemented, tested and gated in the same wave. It
does not start a worker, change an operating toggle, approve LIVE trading, or call a broker during
assembly/deployment.

## Safety boundary

The historical public methods `CommitStrategyDispatchMarketAuthority` and
`IssueStrategyDispatchLease` remain permanently dormant. A caller-populated authority or lease plan
therefore remains unissuable. The production path gets a new narrow operation which:

1. accepts the opaque first-leg receipt and the current journal-minted central owner fence;
2. reloads every lineage, Guardian, aggregate reservation and five monetary reservation row;
3. derives the lease plan rather than accepting a caller plan;
4. commits the current per-market authority and the `ISSUED` lease in one transaction;
5. returns a lease CAS only after the durable trigger set accepts the exact KR/US binding.

Protection is sampled through the Gateway's sealed a071 adapter. The returned evidence is opaque
and carries the exact market generation and identity. The Gateway still performs its existing two
fresh checks around submission. Reconciliation is likewise rechecked by the existing entry gate
inside the Gateway. Thus the lease is an audit/fencing prerequisite, not a replacement for either
last-moment gate.

## Function Logic Map

No existing high-risk function is modified for journal issuance or protection observation; both are
implemented as new functions in new files. `Context.NewPairedStrategyEntryProductionAssembly` is
the only existing composition function extended. Its prior map is retained and must be regenerated
after the edit. Its new branch is fail-closed: missing Gateway, Guardian, paired authority, owner,
protection, schedule, account, FX or five-bucket evidence leaves the market non-effective and emits
no lease.

## Branch Test Map

| Branch | KR | US | Expected durable/broker effect |
|---|---:|---:|---|
| exact first-leg + current owner + exact market authority | yes | yes | one `ISSUED` lease, zero broker calls |
| exact replay | yes | yes | same lease/revision, no duplicate authority or reservation |
| cross-market receipt/evidence | yes | yes | refusal, zero lease |
| stale owner epoch/token | yes | yes | refusal, zero lease |
| expired Guardian/lease window | yes | yes | refusal, zero lease |
| missing/released aggregate or any of five holds | yes | yes | refusal, zero lease |
| protection missing/UNWIRED/drift | yes | yes | refusal before transport; exact holds released after claim |
| activation/calendar drift | yes | yes | refusal before lease or irreversible claim refusal |
| claim replay / A→B→A authority drift | yes | yes | consumed lease never revives |
| production assembly only | yes | yes | workers remain OFF; no LIVE/toggle mutation |

## RED / GREEN / VERIFY

RED is a paired test which can commit first-leg rows but cannot call any production lease mint.
GREEN requires derived authority/lease issuance, claim, and Gateway request construction for both
KR/KRW scale 0 and US/USD scale 2. VERIFY includes race, restart, rollback and fault-isolation tests,
the full Go suite, `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a072-wire-multi-market-strategy-runtime`.
