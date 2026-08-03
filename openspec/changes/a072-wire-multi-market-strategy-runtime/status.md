# Status — a072-wire-multi-market-strategy-runtime

- Isolated core: GREEN (`internal/strategyruntime`)
- Release/default: paired KR/US, both entry `OFF` / runtime `UNOBSERVED`
- Coordinator: independent market calendar/activation/evidence/budget/refusal/latch and bounded restart state
- Lineage: complete immutable candidate→evidence→router→lane/version→campaign/leg→risk/reservation→Guardian→attempt chain
- Lease: sealed, revisioned, irreversible `ISSUED→CLAIMED→SUBMITTING→terminal`
- Validation: exact authority generations/digests, protection serial, account/market/symbol, owner epoch/fence and expiry
- Outcomes: atomic `REFUSED+RELEASED`, `SUBMITTED+TRANSFERRED`, or `AMBIGUOUS+HELD`
- Recovery: CLAIMED crash is pre-transport refusal; SUBMITTING requires non-contradictory authoritative lookup/response proof observed no earlier than lease issue
- Resubmit: `AMBIGUOUS+HELD` only, same operation ID, exact current authority/owner and complete attested broker capability with bounded attempts
- Ordering: nonterminal out-of-order submit/classification consumes the lease as `REFUSED+RELEASED`, zero broker requests
- Worker hardening: immutable first typed refusal, exact latch/OFF invariant and saturating restart attempt/deadline
- Identity: common canonical UTF-8 boundary, maximum 256 bytes, no whitespace/control; post-mint reseal cannot bypass it
- Fallback: newer fenced safety-only owner, frozen RTO ≤60 seconds, no entry/lease issuance
- Authority: package-private constructors; no broker/live transport, journal/gateway writer, toggle or approval
- Remaining: production coordinator/lineage persistence/Gateway CAS integration, full safety-loop fault injection, repository gates and final independent review

No market was activated, no lease reached a broker, and no live order or operating setting changed.
