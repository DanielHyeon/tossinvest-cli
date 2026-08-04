# Status — a066-add-multi-horizon-risk-buckets

- Updated: 2026-08-04
- Overall: IN PROGRESS
- Current wave: Wave 1E v24 owner-lifecycle hardening GREEN and independently CLEAN; official holdings mint pending
- Runtime authority: dormant q_final Guardian/Gateway seam only; no sealed strategyflow/engine/broker/toggle activation

## Read-only KR/US snapshot-authority checkpoint

- A pure `riskbucket` service contract now validates one exact sealed bundle containing the five
  horizon/market/strategy/sector/symbol snapshots, their policy provenance and immutable matching
  journal references for either KR/KRW or US/USD.
- Scope, policy and snapshot freshness, currency pairing, missing/duplicate dimensions, monetary or
  reference tampering, and KR/US cross-reuse all fail closed before a bundle is returned.
- The source interface and material constructors are package-private, returned entries are value
  copies, and the bundle has a canonical SHA-256 seal. Caller-provided authority strings cannot mint
  a production source.
- The original zero-argument constructor still fails closed, while the new package-owned production
  loader consumes exact signed per-market policy, sealed strategy/FX authority and schema-v26 journal
  usage. It has no writer, signer, activation or execution capability and does not enable live entry.

## Sealed KR/US FX-authority checkpoint

- `officialfx.Evidence` remains opaque from the official read through q_final precheck and final
  issuance. Guardian time revalidates it and `riskbucket.BindFXAuthority` alone derives the exact
  arithmetic DTO; caller-provided public FX fields are never monetary authority.
- The official client configuration is sealed before `New` returns. Retained option closures cannot
  change its endpoint, transport or account selection, and official-origin validation plus the FX GET
  execute inside one read-lock boundary. Configured, custom and non-comparable transports fail closed.
- Same-currency identity and cross-currency haircut policy are private sealed capabilities with bounded
  identities and freshness. No production snapshot/policy loader exists, so both KR and US production
  FX minting remain unavailable and q_final entry remains closed.

## Completed in Wave 1B checkpoint

- Released schema 21→22 remains byte-for-byte immutable for policy/snapshot provenance, final
  admission decisions, owners, reservations and its original order/fill evidence. Wave 1C adds a
  separate atomic 22→23 transition: released order/fill/allocation tables are retained under `_v22`
  legacy names and new scoped tables are created without auto-promoting legacy rows.
- Journal-owned admission recalculation and one atomic transaction for `q_final`, the pre-existing
  HELD Guardian reservation reference, one owner and all five HELD bucket reservations.
- Database-enforced one-owner-per-account/market/symbol arbitration across concurrent journal
  processes, exact digest idempotence and stable mismatch on divergent transaction replay.
- Read/replay projection of owner and HELD/FILLED usage with a persisted state digest. Missing
  legacy state returns `ErrRiskBucketStateUnknown`; drift returns `ErrRiskBucketReplayMismatch`
  and is never silently repaired or deleted.
- Independent source-review hardening rejects immutable key/digest collisions, requires exact
  prospective identity for owner reuse, and gives same-owner scale-in a monotonic sequence plus a
  canonical aggregate replay digest independent of timestamp text ordering.
- Admission receipts contain exactly five unique reservation IDs; scale-in is restricted to the
  owner's exact existing five bucket keys and policy versions.
- Owner market/symbol must equal the corresponding bucket values, and snapshot references must
  exactly match the sealed policy/snapshot evidence consumed by admission; tampering produces zero
  writes and a stable snapshot mismatch.
- Idempotence hashes the canonical ordered full consumed bucket bindings, so equal caps cannot hide
  changed limit/FILLED/HELD values or different authority evidence on retry.

## Completed in Wave 1C checkpoint

- A confirmed official BUY order can be registered against its exact risk decision, owner and five
  reservations. Broker order IDs are scoped by account and market, so equal KR/US IDs cannot cross
  release or actual-evidence boundaries.
- `Journal.RecordFill` applies the a066 sidecar inside the same `BEGIN IMMEDIATE` transaction as its
  authoritative fill snapshot, Position and existing campaign/exit hooks. A later hook failure rolls
  back both projections; semantic a066 drift preserves the fill and Position and latches new entry.
- Partial fills transfer proportional HELD in every bucket. Monotonic actual evidence completes
  usage as `filled=max(transfer,actual)`; missing actual price/fee/FX latches
  `UNKNOWN_ACTUAL_RISK` without rejecting the fill.
- Replacement ownership is handed to the successor reservation. A replaced predecessor cannot be
  released independently; predecessor-late or post-cancel/expiry fills remain authoritative and
  latch overage when released HELD is insufficient.
- Duplicate/retry, restart replay, orphan mapping, snapshot drift and crash atomicity have focused
  coverage. All monetary sums use bounded exact 256-bit integers rather than SQLite integer casts.
- SELL/risk-reducing fills bypass entry accounting. Single-decision owners are supported; owners with
  multiple final decisions fail closed at order registration until aggregate order binding lands.
- Order replay seals are canonical ordered DTOs, never JSON encodings of struct-keyed maps. The
  reservation-policy seal and quote/base currencies are derived from the five persisted authority
  policies; caller strings cannot become authority. Order quantity must equal the one exact confirmed
  intent quantity. US coverage pins USD/KRW while KR remains isolated.
- Ambiguous or corrupt sidecar reconstruction is classified separately from database transport
  failure. It preserves the broker fill and Position, latches every applicable reservation and owner,
  and records `FILL_UNACCOUNTED`; genuine storage failures still roll back the outer transaction.
- v22→v23 preservation, failed-migration rollback and older-v22-build refusal are step-pinned. The
  original v22 SQL matches commit `4aee6853`; a v22 binary sees v23 as `ErrSchemaTooNew`.
- Actual-evidence completion and cancel/expiry release remain package-private test seams with zero
  non-test callers. They cannot become production authority until official sealed fill evidence and
  journal-derived confirmed cancellation/expiry, broker-zero and clean lifecycle adapters land.

## Completed in Wave 1D checkpoint

- Same-owner scale-in may append a second exact decision while an earlier decision order remains
  active. Existing owner, bucket key/policy, state-digest and confirmed intent quantity/scope checks
  remain fail closed.
- Fill reconstruction is owner-wide across all decisions, orders, fills and allocations, while order
  watermarks are keyed by immutable `order_key`. A broker order ID collision across decisions is
  rejected before any second order is written.
- Aggregate HELD/FILLED/overage deltas are computed with bounded exact arithmetic and applied only to
  the target order's decision-specific reservation IDs. A late fill with insufficient target HELD
  cannot consume another decision's HELD; it preserves the fill and latches overage instead.
- UNKNOWN/overage latch state is propagated across every matching owner reservation. Actual evidence
  clears UNKNOWN only after every owner order fill is resolved, including after restart.
- Semantic order/ownership ambiguity preserves the authoritative fill. Risk-sidecar ambiguity also
  preserves the Position hook commit, writes no false allocation and records `FILL_UNACCOUNTED` while
  latching every owner decision in scope.
- No schema/version change, runtime toggle, Gateway, broker or live-order path was added in Wave 1D.

## Completed in Wave 1E checkpoint

- KR and US owners use the same journal-derived prospective→actual generation contract. Binding runs
  after Position and campaign projection in the authoritative fill transaction and is set-once.
- Missing/ambiguous/stale bind authority latches only new exposure and never turns a semantic a066
  gap into a dropped fill. Explicit lifecycle calls return typed blockers with zero writes.
- Release derives CLOSED/zero, exact latest generation and entry decision, released legacy and a066
  HELD state, terminal entry/SELL mutations, absent campaign claims, clean exact-generation protection,
  resolved actual fills/latches and an exact structured official broker-zero observation in one transaction.
- Additive v24 preserves v23 unchanged and records official source, canonical zero quantity, broker-as-of,
  capability/build/source versions and payload digest by exact account/market/symbol/actual generation.
  `ADJUSTMENT_APPLIED` also requires the exact zero adjustment and a later official zero recheck.
- Scalar observation plans were removed. The recorder consumes only an opaque sealed capability, and this
  change intentionally ships no production mint or call site because no immutable official holdings response
  authority exists yet. Unsealed, arbitrary and post-seal-mutated caller data is rejected before a transaction.
- Release seals campaign/Position versions, observation ID/digest, predecessor sequence/state digest,
  immutable `OWNER_RELEASED` event and receipt. Retry recomputes all facts; missing/divergent receipt,
  event or post-release state drift is refused instead of returning early as already released.
- A released predecessor's full `RecordFill` remains authoritative, creates durable ORPHAN_FILL and
  market-scoped symbol RECONCILE evidence even before a new owner exists. The journal admission gate runs
  before both first-owner INSERT and reuse; a US late fill blocks US but not the same account/symbol in KR.
- v24 replaces the legacy symbol-wide active index with separate global-NULL and exact-market unique
  indexes plus overlap triggers. Reconcile entry/read/release carries a validated KR/US scope: global NULL
  blocks both entries, KR and US exact rows coexist, and a market release can clear neither its peer nor the
  global row. Reverse-order coverage starts with KR active, records a US late fill, and proves independent
  latches, admissions, replay and release.
- Mutation cleanliness is allowlist-based; unknown attempt kind/state blocks. No runtime toggle, Gateway,
  broker transport or live-order path was added.

## Completed in Wave 1A

- Exact account-base-minor reservation using worst executable price, official frozen fresh FX,
  haircut, minimum fee and ceil.
- Bounded exact cap search and five-dimensional `q_final` intersection that cannot increase
  `q_candidate` or the existing Guardian cap.
- Typed fail-closed refusals for missing/stale/unknown policy, dimension and arithmetic evidence.
- Pure fill accounting for proportional HELD transfer, greater actual exposure, late actual-evidence
  completion, duplicate/retry idempotence, crash-pure error rollback and overage/unknown entry latch.
- Pure one-symbol owner acquisition, same-owner reuse, prospective-to-actual set-once binding and
  conservative idempotent clean release.
- Review hardening: bounded stored-minor addition/recompute with crash-pure overflow refusal;
  immutable policy/snapshot provenance bound to exact bucket identity and snapshot amounts; fresh
  release attestation bound to owner/lane/campaign/prospective/actual generations; exact actual-FX
  quote/base pair binding; deterministic duplicate-owner reconstruction refusal; bucket-local latch
  enforcement in `EntryBlocked`.

## Verification checkpoint

| Check | Result |
|---|---|
| Initial RED compile | PASS (missing implementation symbols observed) |
| `go test ./internal/riskbucket` | PASS |
| `go test -race ./internal/riskbucket` | PASS |
| `go vet ./internal/riskbucket` | PASS |
| Review-fix RED compile | PASS: missing provenance, attestation and currency-pair contracts observed |
| Repeated overflow/reconstruction/property tests, 25x | PASS |
| `FuzzReservationIsMonotone`, 3 seconds | PASS, 475,508 executions |
| `FuzzApplyFillRetryIsPure`, 3 seconds | PASS, 222,136 executions |
| Focused statement coverage | 77.6% |
| `git diff --check` | PASS |
| CodeGraph sync/status | PASS: 1,368 files / 23,745 nodes / 77,709 edges |
| CodeGraphContext advisory update | INCOMPLETE: stalled after DB load; terminated |
| Wave 1B focused journal tests | PASS |
| Wave 1B focused journal race | PASS |
| `go vet ./internal/journal` | PASS |
| Strict OpenSpec validation | PASS |
| Full `go test ./internal/journal` | INCOMPLETE: no-output timeout at 240 seconds; focused suites remain GREEN |
| Wave 1C focused journal tests | PASS (including canonical replay, replacement/release and ambiguous non-drop regressions) |
| Wave 1C focused journal race | PASS (54.071s) |
| Wave 1C existing fill/apply-hook regressions | PASS |
| Wave 1C journal+riskbucket vet | PASS |
| Wave 1C strict OpenSpec validation / diff check | PASS |
| v22 SQL versus released `4aee6853` | PASS: exact diff |
| v22→v23 row preservation / no auto-promotion | PASS |
| broken v23 rename+version atomic rollback | PASS |
| v22 build opening v23 | PASS: `ErrSchemaTooNew` |
| v22→v23 pre-migration backup | PASS: self-contained v22 copy with exact legacy rows |
| v23 hardened focused journal race | PASS (34.960s) |
| Wave 1D owner-wide two-decision/restart tests | PASS |
| Wave 1D ambiguity/non-drop and broker-ID collision tests | PASS |
| Wave 1D focused journal+riskbucket race | PASS |
| Wave 1D journal+riskbucket vet / diff check | PASS |
| Wave 1E owner lifecycle RED compile | PASS: missing bind/release symbols observed |
| Wave 1E focused KR/US bind/release/restart/race/late-fill tests | PASS |
| v23→v24 preservation / legacy reconcile remains observation-unknown | PASS |
| broken v24 tables+columns+version atomic rollback | PASS |
| v23 build opening v24 | PASS: `ErrSchemaTooNew` |
| Wave 1E structured zero/receipt/late-RecordFill focused tests | PASS |
| arbitrary/unsealed/mutated official-zero capability | PASS: zero observation writes |
| late fill before any reopened owner / first admission / KR-US isolation | PASS |
| Wave 1E focused `-race` plus `go vet ./internal/journal` | PASS |
| Wave 1E full `go test ./internal/journal -count=1` | PASS (159.365s) |
| Wave 1E scoped-reconcile full journal rerun | PASS (163.752s) |
| Wave 1E scoped-reconcile focused `-race` | PASS (20.447s) |
| Wave 1E scoped-reconcile vet / strict OpenSpec / diff check | PASS |
| Wave 1E independent final re-review | CLEAN: 0 Critical/Warning; focused x10 and race x2 PASS |
| q_final Guardian/Journal focused tests | PASS: KR cap/atomicity, owner rollback, stale/mutated evidence, US currency refusal |
| Full `go test ./internal/execgw -count=1` after q_final wiring | PASS (32.705s) |
| q_final focused execgw/journal `-race` | PASS (15.520s / 7.158s) |
| q_final `go vet ./internal/execgw ./internal/journal` | PASS |
| q_final strict OpenSpec validation / diff check | PASS |
| Read-only KR/US snapshot authority RED compile | PASS: missing authority service and sealed bundle symbols observed |
| Read-only KR/US snapshot authority focused/race/vet | PASS |
| Sealed FX authority affected-package tests/race/vet | PASS |
| FX authority independent adversarial re-review | CLEAN after opaque q_final and Option-replay/TOCTOU hardening |

## Pending integration

- Package-owned production snapshot loader and sealed a072 strategyflow-to-q_final bridge; the
  legacy Parker adapter remains KR-only.
- Package-owned official FX identity/haircut loaders; public zero values remain unavailable by design.
- Entry-loss-lock integration and broader zero exposure-raising broker request spies.
- KR/US concurrent runtime integration and independent lane-operation tests.
- Full repository validation, independent implementation review and `make gate`.

No existing runtime function, live order path or operating toggle is activated by Wave 1B.
# q_final Guardian/Gateway checkpoint (2026-08-04)

- Added a dormant market-generic `QFinalEntryIssuance` seam. Its precheck is mutation-free, overwrites caller-supplied Guardian cap and evaluation time, calculates exact `q_final`, and returns an opaque precheck.
- Added one journal transaction for the q_final GuardianDecision, aggregate HELD reservation, five monetary HELD reservations and exact lane/campaign owner. Owner, snapshot or bucket conflict rolls every new row back.
- Added immutable q_final policy marking plus read-only Gateway revalidation of exact decision quantity, active owner, aggregate hold, five exact dimensions and current risk state digest. The same check runs again immediately before broker transport.
- KR/KRW is GREEN. Market/currency pairs are exact. A KRW Guardian cannot safely evaluate raw US/USD existing caps, so US returns typed `CURRENCY_UNRESOLVED`, q_final 0 and zero collection/broker calls until a072 supplies the sealed converted-cap/official-FX bridge.
- The legacy `strategyengine.Decision` remains KR/KRW-only and was not weakened. No runtime lane, automation toggle, approval or LIVE order was activated.
