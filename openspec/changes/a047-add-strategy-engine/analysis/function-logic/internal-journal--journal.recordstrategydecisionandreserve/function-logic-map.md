# Function Logic Map: `Journal.RecordStrategyDecisionAndReserve`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| issue request | canonical EXPOSURE_RAISING RiskIntent | Guardian | reject before transaction |
| strategy plan | canonical full DecisionRecord payload, identity, denormalized decision/attempt/manifest bindings | Guardian strategy issuer | reject before transaction |
| snapshot/version | fresh and current | broker collector + reservation ledger | rollback/recollection |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Journal or database is nil | none | journal-required error | direct test missing |
| B2 | `request.Issue.build()` fails | none | build error | invalid request/callee tests |
| B3 | built decision class/preimage kind is not canonical exposure-raising RiskIntent | none | canonical-RiskIntent error | risk-binding refusal tests |
| B4 | `plan.CreatedAt.IsZero()` | local plan time becomes journal UTC now | continue | production issuance success |
| B5 | supplied creation time is nonzero | local plan time becomes UTC | continue | direct test missing |
| B6 | complete lineage, manifest, attempt or revision binding fails | none | complete-exact-binding error | exhaustive binding tests |
| B7 | `verifyStrategyRiskBinding` fails | none | exact risk/payload binding error | exhaustive projection/identity/payload tests |
| B8 | transaction begin fails | none | database error | injected begin failure missing |
| B9 | reservation precheck fails | transaction rollback | reservation error | reservation refusal suite |
| B10 | decision insert fails | transaction rollback | insert error | isolated direct test missing |
| B11 | reservation row insertion fails | transaction rollback | reserve error | rollback/callee tests |
| B12 | exact strategy decision insert fails | transaction rollback | collision/insert error | collision helper and projection tests |
| B13 | exact strategy attempt insert fails | transaction rollback | collision/insert error | production rollback trigger test |
| B14 | exact `DISPATCH_START` insert fails | transaction rollback | collision/insert error | exact execution helper tests; direct function isolation missing |
| B15 | commit fails | transaction rollback attempt | wrapped commit error | injected commit failure missing |
| Scenario | all exact inserts and commit succeed | one commit across decision, reservation, decision lineage, attempt and start event | issue result + complete receipt | production atomic success/restart tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `IssueRequest.build` / `ParsePreimage` | canonical risk authority | fail closed | AST |
| strict DecisionRecord decode + canonical remarshal | bind every current/future record field into identity while rejecting unknown/trailing/noncanonical JSON | fail closed before SQL | payload/lineage tests |
| reservation precheck/rows | take exact aggregate hold | transaction rollback | existing reservation suite |
| exact strategy insert helpers | idempotency/collision enforcement | no divergent replay | strategy lineage tests |

## State mutations and fallbacks

- One SQL transaction owns decision, reservations, strategy decision, strategy attempt and DISPATCH_START.
- No gateway/network call occurs inside the transaction.

## Safety conclusion

- Safe edit boundary: append-only v14 strategy tables plus existing atomic reservation path.
- High-risk impact: yes, this is the exposure-raising durability point.
