# Function Logic Map: `Journal.RecordQFinalDecisionAndReserve`

- Source: `internal/journal/risk_bucket_issuance.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `j` / `j.db` | non-nil open journal | journal constructor | explicit `journal required`; no write |
| `request.Issue` | canonical exposure-raising `RiskIntent`, one aggregate reservation | `IssueRequest.build` and canonical preimage parser | original validation error; no transaction |
| q_final policy binding | policy contains exactly one a066 marker with the admission transaction id | `QFinalPolicyVersion` / `splitQFinalPolicyVersion` | snapshot mismatch; no transaction |
| calculated admission | accepted five-bucket calculation and exact decision/reservation/quantity binding | `riskbucket.CalculateAdmission`, `validateRiskBucketAdmission` | typed refusal or snapshot mismatch; no write |
| request digests | canonical full admission and issue preimages | `riskBucketAdmissionDigest`, `qFinalIssueDigest` | canonicalization error; no transaction |
| durable replay state | either no transaction id rows, or one byte-exact complete prior issue | `recoverQFinalIssueReplayTx` | exact replay returns original receipt; divergence/partial state fails closed |
| aggregate journal state | current observed reservation version and limits | `reservePrecheck`, `reserveRows` | rollback on stale/conflict/insert error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | journal/db nil | none | `journal required` | nil receiver/db test |
| B2 | `Issue.build` rejects canonical decision or reserve | none | original validation error | IssueRequest validation suite |
| B3-B5 | wrong safety/preimage kind, parse failure or non-`RiskIntent` preimage | none | canonical RiskIntent error | non-risk and malformed preimage cases |
| B6 | policy marker absent, malformed or bound to another transaction | none | `ErrRiskBucketSnapshotMismatch` | q_final policy binding cases |
| B7 | pure admission refuses | none | original typed risk refusal | rejected bucket calculation case |
| B8-B10 | decision/reservation/quantity differs from calculated q_final | none | `ErrRiskBucketSnapshotMismatch` | missing/multiple reservation and quantity mismatch cases |
| B11-B13 | admission validation or digest construction fails | none | original error | cross-scope/tampered snapshot and canonical digest cases |
| B14 | DB transaction cannot begin | none | wrapped begin error | closed DB/cancelled context case |
| B15-B17 | replay lookup errors, exact replay, or replay commit fails | read transaction; exact replay commits no new row | mismatch/error, or original receipt | identical/divergent/partial replay plus commit-failure injection |
| B18-B21 | reserve precheck, decision insert, aggregate reserve, or five-bucket/owner insert fails | all prior writes are inside the same transaction | original error with deferred rollback | stale version, identity collision, owner conflict and per-boundary rollback |
| B22 | fresh transaction commit fails | rollback/SQLite atomicity | wrapped commit error | commit-failure injection |
| success | all validations and writes succeed | decision + aggregate + q_final/owner/five buckets commit | receipt | KR and US paired atomic success |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `IssueRequest.build` / `ParsePreimage` | canonicalize the legacy decision/reservation and recover exact RiskIntent | synchronous, no retry, no side effect | AST B2-B5; decision tests |
| `riskbucket.CalculateAdmission` | recompute q_final from sealed bucket evidence rather than accept caller q_final | pure typed refusal | AST B7; riskbucket tests |
| `validateRiskBucketAdmission` | bind scope, owner, snapshots and exact five dimensions | synchronous fail closed | AST B11; admission tests |
| `riskBucketAdmissionDigest` / `qFinalIssueDigest` | seal the full replay preimage | pure canonical digest; any error aborts before DB | AST B12-B13 |
| `recoverQFinalIssueReplayTx` | distinguish fresh issue from exact/divergent/partial replay | same DB transaction; no repair | AST B15-B17; issuance replay tests |
| `reservePrecheck` / `insertDecisionRow` / `reserveRows` | enforce aggregate reservation CAS and insert decision/HELD rows | same transaction; any error rolls back | AST B18-B20 |
| `commitFreshRiskBucketAdmissionTx` | insert q_final, risk owner, five monetary HELD reservations and event | same transaction; conflict rolls everything back | AST B21; owner-conflict test |
| `tx.Commit` | publish either exact replay or fresh issue atomically | no automatic retry | AST B17/B22 |

## State mutations and fallbacks

- All mutations occur after `BeginTx`; the deferred rollback is the only failure fallback.
- Fresh success inserts a decision, one aggregate HELD reservation, one q_final decision, one owner,
  exactly five monetary HELD reservations and one admission event.
- Exact replay performs validation and commits the read transaction without inserting or repairing.
- Divergent or structurally partial replay never falls back to fresh issuance.
- The function performs no network, Gateway, broker, activation or operating-setting call.

## Safety conclusion

- Safe edit boundary: extract/reuse pre-transaction preparation and in-transaction fresh/replay helpers
  only if the standalone API keeps every branch and byte-exact replay behavior above. The new first-leg
  API must add its rows inside the same transaction rather than sequence public APIs.
- High-risk impact: yes — this is exposure-raising Guardian/risk authority. Paired KR/US RED tests,
  statement-boundary rollback, replay mismatch and post-edit AST/branch comparison are mandatory.
