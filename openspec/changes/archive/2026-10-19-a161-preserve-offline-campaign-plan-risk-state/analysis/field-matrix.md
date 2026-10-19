# Native field coverage

The detached evidence is a caller-owned point-in-time inspection snapshot. It does not restore an executable native type, attest completeness, assert current ledger state, or capture weekly reservations. Native digest and seals are copied, not regenerated or authenticated by this codec. The external SHA and scope bind the bytes to expectations supplied by the caller; they do not authenticate the capture source.

| Source | Captured fields |
|---|---|
| All three native plans | Account, market, symbol, campaign, lane and version, position generation; budget, per-share risk, planned quantity, per-leg ceilings, policy/configuration digests, account/quote currencies, optional FX and native plan digest. Continuation copies native 8:4:2 weights; reversal records native fixed 2:4:8 weights; weekly has seven ceilings and no weights. Weekly plan seal copied. |
| Optional FX | Source quote ID, as-of, direction, rate, haircut, digest and native seal. Reversal and weekly also have authority/version/fresh-until; continuation and reversal preserve official/frozen booleans. Weekly timestamps are represented as RFC3339Nano; original instant and offset remain represented. Values are not refreshed against capture or current time. |
| Continuation `RiskState` | Risk-budget digest, budget, filled, held, all latch key/boolean pairs, fill key with applied/risk/known/fingerprint, cancel key with applied/fingerprint. |
| Reversal `RiskState` | Plan digest, filled, held, all latch key/boolean pairs, fill key with applied/risk/fingerprint and cancel key/stored release-value string. Whitespace accepted by native numeric parsing remains unchanged. |
| Weekly `RiskState` | Plan digest, filled, held, all latch key/boolean pairs, fill key/fingerprint/known-identity and risk seal. `Known` is identity knowledge only. No native cancel map; `cancel_status=not_applicable`. Separate `ReservationState` is not exported; `reservation_status=not_captured`. |

Nil and empty maps are separately serialized. Unknown nested fill amounts remain empty instead of being converted to zero. Stored filled/held values can exceed the plan budget and are not treated as an authorization, a recomputed current-risk claim, or a remaining budget.
