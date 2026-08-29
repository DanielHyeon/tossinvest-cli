# KR/US production lane proposal authority — pre-edit specification

## Decision

KR and US advance in the same component wave. For each market the engine consumes one externally
digest-pinned, Ed25519-signed proposal manifest, replays the referenced immutable snapshots from the
existing append-only `evidence.db`, and evaluates only the exact lane selected by the sealed production
router. A valid result is a `strategyflow.ValidProposal()` q_candidate; it is not q_final, a Guardian
decision, a journal reservation, a dispatch lease or an order.

The production evaluator does not copy StockOS's Python runtime or call WTS. StockOS-compatible source
evidence crosses the boundary only through the a064 point-in-time envelope and snapshot contract. The
lane arithmetic, 8:4:2 continuation, reversal-confirmed 2:4:8, weekly value rules, non-retreating stop and
execution-term validation remain the existing TossOS pure evaluators.

## Read-only evidence authority

`evidence.db` is an absolute owner-only regular file with exact mode `0600`, not a symlink. Production
opens it with SQLite `mode=ro`, `query_only=true`, a bounded busy timeout and exact schema version 1. It
must not create a directory, database, WAL, migration or snapshot. Only an already sealed
`(snapshot_id,snapshot_digest,market)` may be replayed. Full header and payload integrity, both point-in-
time cutoffs and the deterministic snapshot digest are revalidated by `DormantSnapshotReadPort`.

Each proposal scope binds one snapshot reference. The replayed snapshot must exactly match account-
independent market/symbol, evaluation instant and the lane's required evidence kinds. Missing,
unavailable, unverified, stale, conflicting, foreign-market or duplicate required evidence refuses that
symbol. Required metric fields come from the canonical snapshot payload; signed caller scalars cannot
replace them. The lane evidence lineage digest is the immutable snapshot digest.

## Signed proposal manifest

The only filenames are `strategy-proposal-input-KR.json` and `strategy-proposal-input-US.json`. Each is
an owner-only regular `0400` file, canonical JSON, externally SHA-256 pinned and Ed25519 signed. TossOS
contains no writer, signer, private key, approval endpoint or digest discovery fallback.

The signed body binds schema/domain/key/generation, account/market, router manifest digest, scheduler
activation/calendar/config facts, observed/fresh interval, actor/revocation state, evidence DB identity,
and a strictly ordered bounded symbol scope set. Every scope binds:

- symbol, numeric position generation, candidate life ID and the exact route owner/market revisions;
- selected horizon/lane/version plus router evidence/config digest;
- deterministic campaign ID and first-leg ordinal/current filled quantity;
- immutable account-base risk budget, per-share risk, planned quantity and policy digest;
- snapshot ID/digest and required evidence kinds;
- exact stop candidate, saved effective stop and entry/target/cost/RR preimages; and
- the lane-specific threshold/config version and, for weekly value, official market-week identity and
  durable reservation identity.

The manifest is checked against the separately held approved candidate, route request/decision,
scheduler activation/calendar and official FX evidence. It cannot switch the selected lane, invent an
owner, weaken a stop, enlarge a position, select a later evidence snapshot or extend freshness.

## Six paired adapters

All six bindings are delivered together:

1. KR continuation replays verified KRX/source-neutral flow evidence and proposes the current 8:4:2 leg.
2. US continuation replays verified participation/price-volume evidence and proposes the paired 8:4:2 leg.
3. KR reversal replays absorption metrics; leg 3 additionally requires sweep→break→reclaim evidence.
4. US reversal replays dislocation metrics; leg 3 has the same structural confirmation contract.
5. KR weekly value replays point-in-time OpenDART disclosure/model evidence.
6. US weekly value replays point-in-time SEC EDGAR disclosure/model evidence.

Official FX remains opaque. Each lane adapter validates the exact `(quote,account)` pair and freshness at
the frozen evaluation instant, then mints only its package-private frozen copy. Same-currency KR uses the
already verified identity evidence; US uses the same official quote-to-account evidence later consumed by
the five-bucket loader. No lane reads an environment variable, source API, broker or journal.

## Weekly durable uniqueness

The in-memory weekly reservation state is not production durability. Before weekly q_candidate can be
accepted, the read-only proposal path must reconstruct an existing journal-owned reservation keyed by
`(campaign_id,market,stable_market_week_identity)` with exact reservation ID, ordinal, calendar lineage and
active status. The q_final/first-leg atomic transaction must consume or bind that same key so restart,
calendar generation A→B and campaign replay cannot mint a second weekly slot. A signed manifest alone is
not a reservation.

If the required journal schema support is absent, weekly KR and weekly US both remain typed-refused; the
short lanes and the peer market are not blocked. Schema evolution must be additive, preserve v26 rows and
make an older reader fail closed.

## Pairing, isolation and completion

KR and US proposal collection starts concurrently at the same frozen observation. Panic, timeout, bad
file/signature, snapshot replay refusal, missing lane evidence, stale FX, router drift or lane refusal is
contained to the affected market/symbol. Public snapshots expose only counts, typed reasons and digests;
approved candidates, route requests, evidence payloads, FX and proposal results remain private.

This checkpoint is complete only when continuation, reversal and weekly adapters for both KR and US,
paired concurrency/failure-isolation, read-only evidence guards and weekly durable uniqueness pass in the
same release candidate. A KR-only or US-only pass is incomplete.

No completion in this slice enables a lane, autostart, automation, LIVE approval or broker transport.
