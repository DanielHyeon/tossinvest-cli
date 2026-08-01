# Function Logic Map: `TestOpenReadOnlyQueriesExistingDatabaseAndCannotWrite`

- Source: `internal/performance/readonly_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| checkpoint fixture | exact current schema with one measured trade and no WAL/SHM after writer close | normal performance writer used only in test setup | fixture setup failure is fatal |
| read-only capability | can return dashboard; cannot execute DDL, acquire writer transaction, expose writer methods, mutate main file metadata, or create sidecars | `ReadOnly` contract | any authority/mutation causes test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | writer open/collect/close/stat fixture setup failures | test fixture only | fatal | this test |
| B5-B6 | inspect initial sidecars; any exists | filesystem read | fatal | this test |
| B7-B9 | reader/open-immutable/dashboard failures or wrong complete count | read-only calls only | fatal | this test |
| B10-B14 | DDL/write transaction/query-only/schema-table/close invariants fail | attempted writes must be rejected | fatal | this test |
| B15-B17 | enumerate exported methods and detect forbidden writer methods | reflection only | report error | this test |
| B18-B20 | reader close/stat/metadata identity invariant fails | lifecycle/filesystem read | fatal | this test |
| B21-B22 | inspect final sidecars; any exists | filesystem read | fatal | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Open`, `Collect`, writer `Close` | create one canonical checkpoint fixture | test setup only | fixture assertions |
| `OpenReadOnly`, `Dashboard` | prove derived read capability | one call/no retry | result assertion |
| raw `ExecContext` and PRAGMAs | prove SQLite itself refuses write locks/DDL and is query-only | success is a test failure | authority assertions |
| reflection and `os.Stat` | prove narrow method surface and no file/sidecar mutation | pure/read-only | assertions |

## State mutations and fallbacks

- Setup writes only a temporary fixture through the normal writer, then records its closed checkpoint identity.
- The tested read capability has no accepted mutation or stale fallback; forbidden operations are deliberately attempted and must fail.

## Safety conclusion

- Safe edit boundary: end-to-end capability-negative test for performance evidence reads.
- High-risk impact: no LIVE/trading effect; protects against accidental persistent write authority.
