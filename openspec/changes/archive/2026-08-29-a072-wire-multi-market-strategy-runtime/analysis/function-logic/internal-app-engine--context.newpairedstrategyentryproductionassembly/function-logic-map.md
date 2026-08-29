# Function Logic Map: `Context.NewPairedStrategyEntryProductionAssembly`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receiver `c` | non-nil fully assembled engine context | `NewContext` production assembly | nil refuses without filesystem or broker activity |
| `ctx` | live request context | `engineRuntime` | cancellation is contained in market-local read classifications; constructor itself stays fail-closed |
| `clk` | non-nil engine clock | command runtime | zero/unavailable observation yields paired schedule/candidate refusal and dormant workers |
| `c.Journal.Path()` | exact journal path selected by the engine profile | opened journal | candidate DB is resolved beside this path, preserving default and `--config-dir` profiles without a second path authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | receiver is nil | none | `strategy production context is unavailable` | existing nil-context test |
| B2 | schedule collection completes with any market-local result | owner-only desired/activation files and official calendars are read; no writers | continue with the private paired schedule authority | paired schedule tests |
| B3 | candidate collection runs from that same frozen schedule instant | scheduler-ready markets read exact threshold/evidence/approval files and a read-only candidate DB handle; peer is independent | preserve market-local candidate readiness and immutable snapshots | paired candidate authority tests |
| B4 | route collection runs from the same candidate/schedule frozen instant | KR and US signed route manifests plus read-only schema-v26 owner reconstruction start concurrently; symbol refusal stays local | retain sealed route authority privately for the current assembly step and expose scalar paired route observations | paired route authority tests |
| B5 | paired FX collection runs from the same candidate/schedule frozen instant | candidate-ready KR re-verifies account identity and candidate-ready US reads signed-policy official FX independently | preserve market-local opaque FX authority and scalar observations | paired FX authority tests |
| B5a | paired lane-proposal collection runs from the exact route/schedule/FX authority | signed proposal manifests, immutable evidence.db snapshots and weekly journal v27 reservations are read only; scope failure stays market-local | preserve sealed q_candidate privately and scalar proposal observations | paired proposal authority tests |
| B6 | risk collection is reached without a production lane result | no risk policy or journal read; both markets classify `LANE_NOT_READY` independently | continue with scalar paired risk observations | paired risk loader skip-read test |
| B7 | a later complete KR/US lane result reaches the risk loader | owner-only signed policy and schema-v26 journal are read-only; sealed strategy and opaque FX are revalidated | exact market-local five-bucket bundle or typed market-local refusal | paired risk authority tests |
| B7a | proposal reaches paired account loader | owner-only signed official account snapshots are read concurrently | market-local account authority or typed refusal | paired account authority tests |
| B7b | production Guardian is exact `RiskGuardian` | construct package-private account/risk/FX→atomic first-leg bridge | no call, write, lease, or broker request during assembly | paired first-leg adapter tests |
| B8 | dormant supervisor construction refuses | no broker/journal mutation | return zero assembly and error | existing production assembly validation tests |
| B9 | construction succeeds | no entry trigger is issued and workers stay ineffective while lane-input/admission/Gateway cycle is incomplete | return supervisor plus scalar schedule/candidate/route/FX/risk observations | public capability and dormant-worker tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newStrategyScheduleAuthorityLoader(...).collect` | verify paired desired/calendar/signed activation at one frozen clock | market-local refusal; no writer | schedule loader tests + AST |
| `newStrategyCandidateAuthorityLoader(...).collect` | verify paired threshold/evidence/human approval and sanitize discovery verdicts | market-local refusal/panic containment; read-only DB | paired candidate loader tests + AST |
| `newStrategyRouteAuthorityLoader(...).collect` | verify signed per-market lane scopes and reconstruct the exact current owner set for every approved symbol | KR/US start together; manifest/journal/route refusal is market-local; no writer or broker | paired route loader tests + AST |
| `newStrategyFXAuthorityLoader(...).collect` | mint KR identity and US official quote-to-base evidence with the existing signed-policy service | candidate-not-ready skips the market read; peer failure does not cancel | paired FX loader tests + AST |
| `newStrategyProposalAuthorityLoader(...).collect` | evaluate the exact routed continuation/reversal/weekly adapter against immutable evidence and durable weekly uniqueness | KR/US start together; malformed/stale/tampered scope is omitted without contaminating its peer | paired proposal loader tests + AST |
| `newStrategyRiskAuthorityLoader(...).collect` | consume sealed lane results, opaque FX, signed five-bucket policies and read-only current journal usage | lane-not-ready skips all files; policy/journal failure remains market-local | paired risk loader tests + AST |
| `newStrategyAccountAuthorityLoader(...).collect` | consume signed, digest-pinned official account/exposure snapshots | KR/US concurrent and market-local failure | paired account loader tests + AST |
| `newProductionStrategyFirstLegAuthorityLoader` | bind private proposal, weekly reservation, account, FX, buckets and current campaign CAS | creates no decision until a cycle calls Guardian | first-leg adapter contract |
| `NewPairedStrategyEntryProductionAssembly(snapshot)` | build exact KR/US supervisor with no trigger capability | error aborts assembly | existing supervisor tests |

## State mutations and fallbacks

- No live order, journal, approval, desired-state or operating-toggle mutation occurs.
- Raw activation, threshold sets, evidence, candidates, sealed routes, FX, q_candidate proposals, five-bucket and account authorities remain package-private.
- The public result grows only by scalar market-keyed observations.

## Safety conclusion

- Safe edit boundary: collect paired account authority and construct an untriggered first-leg bridge; do not change worker effectiveness until the dispatch lease and Gateway cycle exists.
- High-risk impact: yes. Accidentally using readiness as activation or exposing an opaque authority would create an unauthorized entry path.
