# KR/US production five-bucket snapshot authority

## Scope

This component wave delivers KR and US together. It adds a read-only production
authority that combines:

- one valid, sealed `strategyflow.Result`;
- the already collected opaque official KR identity or US FX evidence;
- an exact owner-only, Ed25519-signed, externally digest-pinned market policy;
- current committed `filled + HELD` usage read from the schema-v26 journal in
  SQLite read-only/query-only mode; and
- one frozen engine observation time.

It creates no policy writer, signing key, approval, activation, operating
toggle, journal mutation, lease, Gateway capability or broker call.

## Policy authority

The closed filenames are `risk-bucket-policy-KR.json` and
`risk-bucket-policy-US.json`. Each file is regular, non-symlink, owned by the
current process UID, mode `0400`, bounded, strict canonical JSON and pinned by
an environment-supplied SHA-256 digest. Its Ed25519 signature covers the
canonical body and exact market/account/currency scope, generation, validity
window, approver, fee policy, horizon/market limits, lane-to-server-risk mapping
and symbol-to-sector limits.

The policy is read-only input produced by a separate human-controlled process.
TossOS exposes no writer or signer for it. Revoked, expired, future, wrong-key,
wrong-market, wrong-account, cross-currency, duplicate, non-canonical or
unbounded content fails closed.

## Dynamic authority binding

The package validates the complete strategy result and binds the signed lane
mapping to its exact market, horizon, lane ID/version and symbol. The sealed
entry limit is the maximum executable BUY price for the official LIMIT order
contract. Its source, version, digest, observation and validity are derived from
the immutable execution terms and candidate life; a caller price is never
accepted. The opaque official FX evidence is revalidated at the same frozen
instant and translated through `officialfx.Reserve` only inside `riskbucket`.

The resulting current policy version and digest bind the signed policy,
strategy execution terms, official FX evidence and frozen instant. This keeps
immutable journal policy rows collision-free without treating a caller string
as authority.

## Journal usage snapshot

The authority opens the existing journal with SQLite `mode=ro` and
`query_only(true)`, requires exact schema v26 and reads no legacy substitute.
For each of the five canonical dimensions it scans every matching committed
reservation across historical policy versions and performs bounded exact
decimal addition in Go. Released rows contribute their retained filled amount
and zero HELD; all other rows contribute their exact stored filled and HELD.
Any malformed amount, unknown state, overage/unknown-actual latch, broken
policy/snapshot join or schema drift makes the market unavailable.

The snapshot version/digest seals the exact ordered row identities and amounts,
signed policy digest and frozen time. Empty prior usage is an observed journal
fact and is represented as canonical zero; an unreadable or missing journal is
not.

## Paired failure isolation

KR and US load in separate bounded goroutines from the same frozen instant.
Missing or invalid KR policy/journal/strategy/FX authority makes only KR
unready; the exact valid US result remains usable, and vice versa. A panic is
converted to a market-local internal refusal. No combined market authority is
created.

## RED/GREEN matrix

| Case | KR | US | Required result |
| --- | --- | --- | --- |
| exact signed policy + sealed result + current journal + official FX | GREEN | GREEN | five ordered sealed entries per market |
| peer policy mode/digest/signature failure | refusal | GREEN (or reverse) | no peer cancellation |
| cross-market policy/result/FX substitution | refusal | refusal | no bundle |
| stale policy, candidate, entry contract or FX | refusal | refusal | no bundle |
| journal usage changes or carries a latch | new digest/refusal | new digest/refusal | never reuse stale scalar authority |
| attempted SQL mutation on authority handle | refused | refused | journal bytes unchanged |

Both markets must pass in one wave before this component can advance the
production assembly. Missing later lane-input/first-leg/Gateway completeness
still keeps both production workers dormant.
