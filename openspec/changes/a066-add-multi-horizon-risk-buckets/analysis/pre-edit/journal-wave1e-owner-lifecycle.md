# Wave 1E pre-edit evidence — authoritative owner bind and clean release

- Date: 2026-08-04
- Scope: `internal/journal`, minimum `internal/riskbucket`, and a066 evidence only.
- Excluded: execgw, engine, protectionreadiness, performance, broker transport and runtime toggles.
- Initial schema hypothesis: released v23 appeared to have sufficient lifecycle evidence. Independent
  review disproved this because reconcile evidence had no structured market/quantity/provenance seal and
  released retries had no immutable receipt. The implemented correction is additive v24; v23 SQL remains
  byte-for-byte unchanged.

## Authority map

| Fact | Journal authority | Fail-closed rule |
|---|---|---|
| prospective owner | active `risk_bucket_owners` exact account/market/symbol/token/lane/campaign | missing or duplicate exact scope refuses |
| actual generation | `position_campaigns.actual_position_generation`, matching active claim, exact latest `positions.instance_seq`, projection version and entry decision | caller never supplies generation; any mismatch refuses |
| CLOSED/zero | exact generation `positions.state`, decimal `quantity`, `closed_at` | state other than CLOSED, nonzero/invalid quantity or missing close time refuses |
| broker-zero reconciliation | released symbol `reconcile_states` whose original cause is `QUANTITY_MISMATCH` and automatic release is `RECHECK_MATCHED` or `ADJUSTMENT_APPLIED`, exact Position is CLOSED/zero, no active account/symbol reconcile state, causally after close and all relevant facts | missing, operator-only, ambiguous market scope, unknown cause or older evidence refuses |
| entry/HELD clean | legacy `risk_reservations`, a066 reservation HELD totals, registered order remaining and BUY mutation/snapshot lifecycle | any HELD or unresolved/non-terminal exposure mutation refuses |
| protection clean | exact account/market/symbol/actual-generation `protection_sagas` and `protection_mutation_attempts` | only TRIGGERED/CLOSED saga plus CLOSED attempts are clean |
| sell/reduce-only clean | exact-scope SELL intents, mutation attempts and terminal scoped fill snapshots; campaign claim/state | missing attempt, pending/in-doubt/unresolved or live broker order refuses |
| fill/latch clean | a066 fills+actual evidence, scoped fail-closed observations, `FILL_UNACCOUNTED`, owner/reservation/scope latches | any unresolved actual or latch refuses |

## Function Logic Map: new lifecycle functions

The bind/release derivation functions are new leaf functions, so existing-body AST mapping is
`not-applicable`. Their executable Branch Test Map is fixed below. The only existing body edit is
`(*Journal).runApplyHooks`; its generated AST/Function Logic Map is under
`analysis/function-logic/internal-journal--journal-.runapplyhooks/`.

## Branch Test Map

| Branch | Required RED/GREEN evidence |
|---|---|
| first fill bind | Project+Campaign persist the successor Position generation, then the same fill tx binds exactly one a066 owner |
| bind retry/restart | same actual generation is idempotent; another generation or stale campaign/Position authority writes nothing |
| CLOSED but dirty | every individual HELD, reconcile, protection, sell mutation/claim, unresolved fill/actual/latch blocker returns a typed field and zero writes |
| clean release | exact bound owner releases once; retry/restart reports already released without deleting claims/history |
| race | two journal instances produce one release and one idempotent retry |
| KR/US isolation | evidence for one market cannot bind or release the other market owner |
| terminal but dirty | terminal entry order or CLOSED Position alone cannot bypass remaining protection/sell/fill evidence |
| late fill/reopen | released owner is never resurrected; authoritative late fill remains observable and blocks unsafe new exposure through existing reconciliation/latch paths |
| exit bypass | release evaluation is a local bounded journal transaction and is never called by stop, emergency-exit, reconciliation, fill-detection or SELL paths |

## Safety invariant

Caller-supplied booleans, generation strings, claim enums or release attestations are never journal
authority. Binding and release derive all facts under one SQLite transaction. A dirty semantic fact
may block future exposure but must not reject an authoritative fill/Position or mutate/delete the
claim that caused the refusal.

## RED → GREEN checkpoint

- RED: focused journal tests failed to compile on absent `bindRiskBucketOwnerActual`,
  `releaseRiskBucketOwner`, typed lifecycle errors and results.
- GREEN: KR/US bind, same-fill hook ordering, explicit no-write refusal, full clean release,
  terminal-but-dirty protection/sell/actual blockers, reconciliation freshness/scope isolation,
  retry/restart/race and late-fill/reopen isolation pass.
- Schema correction: v24 adds structured official-zero observations, reconcile observation references
  plus exact market scope, and immutable release receipts. Released v23 remains unchanged and its legacy
  reconcile rows remain observation-unknown, so they cannot be promoted into release authority. The scalar
  observation plan was removed; no production official-capability mint exists in this checkpoint.
