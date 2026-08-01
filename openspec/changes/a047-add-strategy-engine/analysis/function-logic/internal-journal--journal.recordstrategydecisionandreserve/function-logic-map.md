# Function Logic Map: `Journal.RecordStrategyDecisionAndReserve`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| issue request | canonical EXPOSURE_RAISING RiskIntent | Guardian | reject before transaction |
| strategy plan | complete exact decision/attempt/manifest bindings | Guardian strategy issuer | reject before transaction |
| snapshot/version | fresh and current | broker collector + reservation ledger | rollback/recollection |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | journal/request/lineage invalid | none | error | binding tests |
| B2 | RiskIntent hash/typed canonical fields disagree | none | exact binding error | divergent replay tests |
| B3 | reservation precheck/insert fails | transaction rollback | issuance error | rollback test |
| B4 | decision/lineage/attempt/start collision | transaction rollback | typed collision | collision tests |
| Success | all exact inserts succeed | one commit across five authorities | issue result + complete receipt | atomic/restart tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `IssueRequest.build` / `ParsePreimage` | canonical risk authority | fail closed | AST |
| reservation precheck/rows | take exact aggregate hold | transaction rollback | existing reservation suite |
| exact strategy insert helpers | idempotency/collision enforcement | no divergent replay | strategy lineage tests |

## State mutations and fallbacks

- One SQL transaction owns decision, reservations, strategy decision, strategy attempt and DISPATCH_START.
- No gateway/network call occurs inside the transaction.

## Safety conclusion

- Safe edit boundary: append-only v14 strategy tables plus existing atomic reservation path.
- High-risk impact: yes, this is the exposure-raising durability point.
