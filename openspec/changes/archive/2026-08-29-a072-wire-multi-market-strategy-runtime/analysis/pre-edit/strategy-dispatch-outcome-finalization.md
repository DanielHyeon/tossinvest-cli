# Strategy dispatch composite settlement pre-edit

## Scope

- Production owner: `internal/journal/strategy_dispatch_runtime.go`
- Tests: strategy dispatch outcome/composite-settlement journal tests
- Documentation: this pre-edit, the affected function/branch maps, and a072 strategy-runtime spec/tasks
- The released `Attempt.DispatchVerified` contract remains unchanged. A strategy-only variant is additive.
- No Gateway, broker adapter, activation, market-authority, lease-mint, or automatic recovery authority is added.
- `RecoverClaimedStrategyDispatchLease` remains dormant; post-crash resolution requires a future sealed official attestation and can never resend.

## Defect and durable evidence

The first outcome implementation accepted caller-provided class, broker order ID, evidence digest, and
observation time after the core `mutation_attempt` had already settled. That is not an authority boundary:
a caller can strengthen its own result and the core confirmation commits before the real risk order mapping,
campaign watermark, strategy lineage, and lease transfer.

That split creates a concrete fill race. Once the core attempt is `CONFIRMED`, `RecordFill` can identify the
order as locally owned. If the fill commits before the strategy sidecars are linked, risk-bucket and campaign
apply paths see no mapping and return a no-op. A byte-identical cumulative re-observation takes the snapshot
early-return path, so the missed accounting is permanent. If the fill arrives while the core attempt is only
`ACKED`, all apply hooks may be skipped. A crash in either gap leaves no safe caller-authorized recovery.

The durable authority already available to the strategy path is:

- the exact core mutation attempt and its intent, decision ID, client-order identity, account and lifecycle;
- the exact strategy first-leg binding and planned lineage attempt;
- the exact lease owner/fence/revision and durable `SUBMITTING` transport marker;
- the exact aggregate and five monetary reservation rows;
- the broker order ID recorded byte-for-byte on the ACKED mutation attempt; and
- any exact scoped fill snapshot committed before composite settlement obtains the journal writer lock.

## Intended contract

Add an `Attempt` strategy-only verified-dispatch variant. Its last pre-send journal transaction atomically:

1. reloads and cross-binds the attempt, intent, decision, strategy first-leg binding, planned strategy lineage,
   lease, account, market, symbol, quantity, and client-order identity;
2. consumes the decision nonce and changes `RECORDED -> DISPATCH_STARTED`; and
3. changes the exact `CLAIMED + RESERVED` lease to `SUBMITTING + RESERVED`.

The broker is then invoked exactly once. An ACK with a non-blank opaque broker ID is committed byte-exact as
`ACKED`, followed by the existing official round-trip callback. The after-send settlement phase uses a bounded
context detached from caller cancellation.

For the normal live strategy path, one `BEGIN IMMEDIATE` settlement transaction must commit all of the
following or none:

- the core attempt's final/uncertain transition and immutable transition evidence;
- the lease outcome derived from the durable attempt state and transition, never from a caller outcome enum,
  broker ID, digest, or time;
- for `CONFIRMED`, the exact risk-bucket order plus five order-reservation mappings, first-leg campaign order
  watermark and campaign/leg transition, and strategy execution lineage;
- for `CONFIRMED`, backfill of any durable exact scoped fill snapshot that arrived before mappings existed,
  using the same authoritative apply hooks and idempotent sidecar watermarks; and
- only after all confirmed links/backfill succeed, `SUBMITTED + TRANSFERRED`.

There is no public post-outcome finalizer. The strategy-only dispatch variant is the sole normal production
path, so no caller can first terminalize an ordinary core attempt and then ask the journal to strengthen a lease.
Future crash recovery remains a separate sealed official-attestation boundary and is dormant in this build.

Derived mapping:

- durable `NOT_DISPATCHED` or `FAILED_CONFIRMED`, with no broker order identity: `REFUSED + RELEASED`;
- durable `CONFIRMED`, with a non-blank byte-exact broker order identity and all links/backfill: `SUBMITTED + TRANSFERRED`;
- every unknown, in-doubt, malformed, or unprovable state: `AMBIGUOUS + HELD`.

No unknown state releases capacity. Only definitive not-sent/rejected states release it. Trimming is permitted
only to test broker-ID emptiness; storage, joins, digests, mappings, and lineage preserve the raw value.

## Atomicity and race invariants

- KR/KRW and US/USD use one market-parameterized implementation and the same paired test matrix.
- The current central owner and the lease's original owner/fence/revision are proven inside both pre-send and
  settlement transactions.
- Exact attempt-to-intent-to-decision-to-strategy-binding-to-lease identity is reloaded, never supplied by the
  caller. Cross-attempt, cross-account, cross-market, cross-symbol, or cross-client-order requests fail closed.
- `CONFIRMED` cannot publish unless one aggregate hold, five distinct full monetary holds, one risk order,
  exactly five order mappings, one campaign watermark, and exact strategy lineage are present in the same tx.
- A fill that commits first is visible to the settlement transaction and is backfilled. A fill that starts later
  sees the committed mappings. SQLite's single `BEGIN IMMEDIATE` writer orders these cases without a gap.
- Backfill uses the durable scoped snapshot and sidecar watermarks; replay never double-applies cumulative fill.
- Any late mapping, lineage, backfill, lease, or immutable outcome failure rolls back the complete settlement.
- After-send settlement ignores caller cancellation but has a finite deadline. It never retries broker transport.
- Duplicate/replayed settlement verifies the exact terminal composite and returns the existing result; a mismatch
  is consumed/replay-mismatch and never rewrites authority.

## Branch test map

| Branch | Expected state | Test evidence |
|---|---|---|
| Paired KR/US confirmed composite | core `CONFIRMED`, lease `SUBMITTED + TRANSFERRED`, 5 risk mappings, campaign watermark, strategy lineage | paired composite table |
| Fill between ACK/core and links | exact cumulative fill is applied once to risk, campaign and projection during composite tx | paired fill-race test |
| Fill after composite | normal `RecordFill` applies once through installed mappings | paired fill accounting test |
| Forged caller class/ID/digest/time | fields do not exist on request and cannot strengthen durable state | compile/API shape plus durable classification tests |
| Cross-attempt/cross-market binding | error, core/lease/holds/links unchanged | paired cross-binding table |
| Definitive not-sent/rejected | core terminal, lease `REFUSED + RELEASED`, exact aggregate+five released | paired refusal table |
| Unknown/in-doubt/malformed durable state | lease `AMBIGUOUS + HELD`, no release, no mappings | paired ambiguity table |
| Opaque broker ID | leading/trailing bytes survive attempt, risk order, campaign watermark, lineage and outcome | paired opaque-ID test |
| Late mapping/backfill/outcome failure | complete rollback to pre-settlement durable state | injected-trigger table |
| Duplicate settlement | exact terminal composite is idempotent; mismatched attempt is refused | replay tests |
| Caller cancellation after broker send | bounded detached settlement still records durable outcome | cancellation test |

## Safety impact

The additive strategy path can invoke the broker only through the caller-supplied existing dispatch callback and
does so exactly once after the durable pre-send fence. It adds no activation, LIVE approval, lease mint, resend,
or recovery capability. Ambiguity holds capacity, and recovery remains sealed/dormant.
