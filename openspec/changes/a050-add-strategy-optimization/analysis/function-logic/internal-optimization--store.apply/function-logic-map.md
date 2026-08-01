# Function Logic Map: `Store.apply`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| capability, persisted candidate, current snapshot/registry, confirmation and clock | capability hashes to one immutable candidate; persisted metadata must exactly rederive; current control pointer/snapshot are digest-bound even on replay | SQLite candidate/application/control/snapshot rows and current registry | typed fail-closed error; no partial write |
| durable commit | control CAS binds old/new pointer digests; snapshot, full event-digest audit rows and application commit atomically | one SQLite transaction | any CAS/row-ID/insert/commit error rolls back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B40 | candidate/MAC/time/evidence/registry validation and authenticated current snapshot (including replay) | no durable mutation before validation | typed fail-closed error or verified replay | hardening and pointer tests |
| B41-B44 | digest-bound control CAS and immutable snapshot insert | transaction-local | conflict/error | concurrency and pointer tests |
| B45-B49 | explicit audit row IDs/digests, application insert and commit | append-only transaction | error rolls back; commit returns result | audit/tamper/lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| candidate query + validation helper | reads immutable command evidence | parse/validation failure fails closed | AST B1-B13 |
| `currentSnapshot`, `Registry.Field` | rederive values from current authority | database/registry errors abort transaction | CodeGraph callers/impact |
| digest-bound control CAS | prevents version or snapshot-digest pointer rollback/race | exact old version+digest predicate | pointer/concurrency tests |
| `insertSnapshot`, `digestAuditEvent`, audit/application inserts | durable append-only result | explicit row IDs and every digest in one transaction | audit/lifecycle tests |

## State mutations and fallbacks

- Replay first authenticates the current control pointer/snapshot, then returns only the persisted immutable result; it never bypasses current-state integrity.
- Fresh applies rederive category/source/reason, unique changes, before value, descriptor timing/safety, risk/restart/effective-entry flags, and the created-at-derived risk-wait/TTL schedule before CAS. No live, lane, gate, journal, or order mutation occurs.

## Safety conclusion

- Safe edit boundary: optimization candidate-to-snapshot transaction and audit actor only.
- High-risk impact: no trading authority; durable audit integrity is safety-sensitive and tested fail-closed.
