# KR/US production router authority — pre-edit specification

## Decision

KR and US route authority is loaded in one paired component wave. Each market has an independent
signed lane manifest and an independent read-only journal reconstruction. No combined KR/US route
capability is minted, and neither market's success or operational stability is a prerequisite for
the peer.

This slice constructs only a sealed `strategyrouter.RouteRequest`. It creates no lane request,
q_candidate, q_final, reservation, lease, Gateway call, broker request, activation, signer or writer.

## Closed production inputs

The only manifest filenames are:

- `strategy-lane-authority-KR.json`
- `strategy-lane-authority-US.json`

Each path is resolved below one absolute configured directory. The regular file must not be a
symlink, must be owned by the running effective UID and must have exact mode `0400`. The schema-v26
journal must be an absolute owner-only regular file with exact mode `0600`; it is opened with SQLite
`mode=ro`, `PRAGMA query_only=true`, bounded busy timeout and an exact `user_version=26` check.

Each manifest is externally SHA-256 pinned and Ed25519 signed. TossOS contains no manifest writer,
signing key, signing endpoint or digest auto-discovery fallback. Strict canonical JSON rejects
duplicate/unknown keys, trailing data and non-canonical representations.

## Signed body

The signed body binds:

- schema/domain/signature algorithm/key ID/generation;
- account and market, plus a canonical ordered set of symbol scopes;
- each symbol scope's exact numeric position generation and owner snapshot expected revision;
- market-record revision, desired/effective `ON`, runtime `UNOBSERVED`;
- exact scheduler activation digest and expiry;
- calendar generation/digest, timezone and `REGULAR` session;
- config version, approving actor, observed-at and fresh-until; and
- exactly three market-local candidates per symbol scope: continuation, reversal and weekly-value.

Scopes are strictly ordered by `(symbol,position_generation)`, unique and bounded. This lets one
market manifest authorize every explicitly listed approved symbol without deriving a path from symbol
text or silently selecting only the first discovery result. An approved symbol absent from the signed
scope set is refused.

Each candidate binds canonical horizon, lane ID/version, signed score, eligibility, desired/effective
state, evidence digest and config digest. Each scope must equal the three descriptors registered for
the manifest market; duplicate scopes/candidates, foreign-market lanes, unknown lanes and missing
lanes refuse the market.
Router ambiguity remains fail closed: equal top eligible scores return the existing typed
`AMBIGUOUS` refusal.

The manifest activation/calendar fields must exactly repeat the separately verified scheduler
authority held privately by the engine. A signed lane manifest alone cannot bypass scheduler OFF,
calendar mismatch or expiry.

## Journal owner reconstruction

`OwnerSnapshot` is never reconstructed from caller values or manifest owner claims. The loader reads
the full owner history for `(account,market,symbol)` from `risk_bucket_owners`, ordered by acquisition
identity, and hashes the exact rows into the snapshot digest. Revision is the monotonic row-history
count plus one and must equal the signed expected revision.

For no active owner, the snapshot contains no owners and the router may consider the signed candidate
set. For one active owner, the loader joins the exact `position_campaigns` row and requires:

- matching account/market/symbol/lane/campaign/prospective token;
- non-null canonical actual position generation equal to the route key generation;
- campaign state `ACTIVE`, `entry_blocked=0`, no risk-overage/unknown latch;
- lane ID/version registered for the same market, with its canonical horizon.

The loader then creates one active `Owner` with desired/effective `ON`. A prospective owner without
actual generation, multiple active owners, terminal/blocked campaign, latch, join mismatch, corrupt
history or cross-market identity refuses routing. It does not invent an owner revision or reuse the
prospective token as numeric generation.

## Market record reconstruction

The package-private `newMarketRecord` consumes only the verified signed body. It requires exact ON/ON,
canonical market timezone, `REGULAR` session, nonzero revision, activation expiry after update and
current frozen observation strictly inside both manifest freshness and activation lifetime.

The package-private `newOwnerSnapshot` and `newMarketRecord` remain unexported. The only exported
production entry point is a read-only loader whose result is already sealed by `strategyrouter`.
Existing external API guards continue to forbid `NewOwnerSnapshot` and `NewMarketRecord`.

## Paired loader and failure isolation

The engine starts KR and US loads concurrently from one frozen clock and separate SQLite read-only
connections. Within each market, all approved symbols share one manifest read and one `query_only`
SQLite transaction so owner snapshots cannot be assembled across different journal revisions. A panic,
timeout, bad signature/mode/owner, corrupt journal row or route refusal is
classified only for that market. The peer result is retained. Public engine snapshots contain only
market, readiness/refusal, selected horizon/lane/version and manifest/owner digests; opaque
`RouteRequest` values remain private.

## Required RED/GREEN matrix

The same suite covers KR and US:

1. exact three-candidate scopes and empty owner history route every signed approved symbol independently;
2. one valid active owner routes before candidates and preserves campaign identity;
3. digest/signature/key/mode/owner/symlink/canonical-JSON failures are market-local;
4. scheduler activation/calendar mismatch performs no route acceptance;
5. owner revision/history digest drift and A→B→A history do not revive old authority;
6. prospective-without-actual, blocked/terminal/latching/multiple/cross-market owner refuses;
7. tie, all OFF and unsupported descriptor preserve existing typed router refusals; and
8. loader performs zero journal writes, zero activation changes and zero broker calls.

KR-only GREEN or US-only GREEN is not completion. Both markets and the paired fault-isolation test
must pass in the same implementation checkpoint.
