# Wave 1C pre-edit evidence — authoritative fill sidecar

- Date: 2026-08-04
- Scope: `internal/journal` risk-bucket order/fill accounting and the minimum pure
  `internal/riskbucket` support only.
- Excluded: execgw, engine, protection, console, httpapi, broker and operating toggles.

## CodeGraph hard evidence

The index reports 1,495 files, 26,204 nodes and 84,811 edges, with only the two
new a066 evidence-accessor files pending indexing.

- `Journal.RecordFill` is `internal/journal/fills.go:313-539`. CodeGraph reports the
  production caller `internal/filldetect.JournalLedger.Apply`, many regression callers,
  20 direct callees and a depth-two impact set of 231 nodes. It is the authoritative
  `BEGIN IMMEDIATE` transaction containing snapshot/event persistence, terminal legacy
  reservation release, position/campaign/exit hooks and commit.
- `Journal.runApplyHooks` is `internal/journal/apply_hook.go:263-292`; CodeGraph reports
  a two-node impact set. It orders Project, Campaign, Exit inside the caller's transaction.
- `ProjectPosition` has a five-node impact set and writes Position through `ApplyTx`.
- `ApplyPositionCampaignFill` is `internal/journal/position_campaign.go:944-1161`, has a
  two-node indexed impact set and advances replacement/predecessor watermarks inside the
  same fill transaction.

## Function Logic Map — `Journal.RecordFill`

| Stage | Condition | Durable effect | Return/next |
|---|---|---|---|
| validate | blank order, partial canonical scope, invalid/non-finite quantity | none | error |
| begin/read | valid observation | `BEGIN IMMEDIATE`; read scoped prior snapshot and confirmed ownership | continue |
| ownership reset | prior external snapshot predates unique local ownership | in-memory baseline reset | continue |
| refusal | caller/broker-state/cumulative inconsistency | append refusal + reservation alerts; snapshot/Position unchanged | commit success, fail-closed result |
| exact replay | prior snapshot byte-identical and delta zero | none | commit no-op |
| correction | cumulative unchanged, price/amount changed | advance snapshot + append correction | continue |
| positive delta | cumulative increased | advance snapshot + append fill event | continue |
| origin | external order or ambiguous confirmed owner | external is observation-only; ambiguity errors/latches per existing origin contract | continue/error |
| terminal local | derived terminal local order | release legacy Guardian reservation | continue |
| atomic apply | locally owned | Project → Campaign → Exit through same `*sql.Tx` | hook error rolls all prior effects back |
| commit | every prior step succeeds | snapshot/event/reservation/Position/campaign/exit commit together | changed result |

Wave 1C will add one call at the atomic apply point, after ownership resolution and before
commit. It will not alter validation, refusal, external-order, terminal-release, hook order or
risk-reducing semantics.

## Branch Test Map for the new sidecar call

| Branch | Required evidence |
|---|---|
| no registered a066 risk order | existing RecordFill behavior unchanged |
| registered BUY, partial delta | proportional HELD transfer in all five buckets in the same commit |
| exact re-observation | no second fill/allocation/event or monetary movement |
| successor fill | child uses its own remaining reservation; predecessor watermark retained |
| predecessor late fill after release/replacement | fill persists; deficient HELD produces durable all-bucket/owner overage latch |
| actual evidence known | every bucket records `filled=max(transfer,actual)` with price/fee/FX provenance |
| actual evidence absent/invalid | fill and Position persist; every bucket and owner latches `UNKNOWN_ACTUAL_RISK` |
| corrupt/orphan/snapshot drift | authoritative fill is not dropped; scope is latched and replay returns a stable mismatch |
| semantic retry mismatch | original fill remains; scope is latched; no partial monetary movement |
| injected failure before outer commit | fill snapshot, Position and risk-bucket movements all roll back |
| SELL/risk-reducing fill | sidecar performs no entry accounting and cannot block the fill |
| cancel/expiry before fill | HELD releases exactly once; later fill is retained and overage-latched |

## Safety invariant

An a066 calculation or evidence problem is data to persist and a reason to block later exposure,
never a reason to reject an authoritative broker fill or roll back Position. Only a storage/commit
failure may fail the enclosing journal transaction, matching the existing atomic projection contract.

## RED/GREEN and review disposition

- RED: the Wave 1C journal tests initially failed to compile because order registration, actual-fill
  completion, release and tx-scoped accounting APIs did not exist.
- GREEN: focused tests now cover every Branch Test Map row, including a failure injected after the
  a066 sidecar but before the outer commit.
- Review HIGH — SQLite `INTEGER` aggregation could truncate 256-bit minor values: resolved by exact
  bounded Go arithmetic and a value-above-`uint64` regression.
- Review HIGH — a broker order ID could collide across KR/US: resolved by exact account/market/order,
  owner and decision scoping, with a same-ID cross-market regression.
- Review HIGH — a multi-decision owner could be ambiguously aggregated: resolved fail-closed for this
  wave; order registration refuses it and an active order prevents later scale-in admission.
- Review HIGH — releasing a REPLACED predecessor could double-release HELD already transferred to the
  successor: resolved by refusing predecessor release and retaining successor-owned HELD.
- Review CRITICAL — JSON cannot marshal `map[BucketKey]string`, so ignored marshal errors collapsed
  replay digests to SHA-256(empty): resolved with canonical required-dimension order/fill DTOs and
  propagated encoding errors. Divergent quantity/reservation retries now fail with zero writes.
- Review HIGH — caller-provided policy and currency strings could masquerade as authority: resolved by
  deriving the five-policy canonical identity/record/policy digest and the single quote/base pair from
  persisted policy rows. US is pinned to USD/KRW in regression coverage.
- Review HIGH — reuse of a REPLACED/RELEASED predecessor and mismatched release-reason retries could
  move HELD twice: resolved by exact ACTIVE predecessor validation, one-row compare-and-transition and
  reason-bound idempotence.
- Review HIGH — ambiguous/corrupt risk rows could return an error before the non-drop path: semantic
  ambiguity/reconstruction now latches all applicable owner/reservation rows and appends
  `FILL_UNACCOUNTED`, while storage/transport failures retain normal transaction rollback.
- Post-review CRITICAL — editing released v22 DDL stranded existing `user_version=22` journals. The
  released SQL is restored exactly; v23 atomically renames the three incompatible tables to immutable
  `_v22` evidence, creates only the new scoped companions, never auto-promotes legacy rows, and relies
  on the normal pre-migration backup plus transactional DDL rollback. Preservation, injected failure
  rollback and older-build refusal are executable tests.
- Post-review authority boundary — actual evidence completion and release are package-private with zero
  production callers until official sealed evidence and confirmed cancel/expiry/broker-zero lifecycle
  adapters exist. Tests retain same-package seams without widening runtime authority.
