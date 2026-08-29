# Function Logic Map: `TestTheEngineSpellsExactlyOneBuy`

- Source: `internal/app/engine/entryreach_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| engine production sources | every non-test `.go` file | repository source tree | test fails on read or unexpected BUY site |
| allowed sites | strategy claimed-lease adapter and unreachable tracer | reviewed architecture | exact ordered equality required |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate production engine files | source reads | scan all files | this structural test |
| B2 | source read fails | none | fatal test | filesystem error path |
| B3 | iterate source lines | none | inspect every line | this structural test |
| B4 | line is a comment | none | skip | comment false-positive control |
| B5 | BUY literal is found | append location | continue scan | exact location assertion |
| B6 | found count differs from reviewed count | none | fatal test | extra/missing BUY regression |
| B7 | compare ordered reviewed files | none | inspect both | exact-path assertion |
| B8 | a BUY location differs | none | test error | unexpected site regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `goSources` | enumerate production Go files only | test helper; fail closed | AST |
| `os.ReadFile` | inspect source spelling | no retry | AST |
| `mustRel` | stable diagnostic path | deterministic | AST |

## State mutations and fallbacks

- Test-only source reads; no runtime or broker mutation.

## Safety conclusion

- Safe edit boundary: structural allowlist for BUY construction sites.
- High-risk impact: yes; any unreviewed BUY path must make the suite fail.
