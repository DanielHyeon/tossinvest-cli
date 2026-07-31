# Function Logic Map: `Console.restartSoakLocked`

- Source: `internal/console/restart.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The caller holds `openAPIMu`. `RestartSoak` is nil or a narrow process seam; its
operator note may be blank.

## Branches and early returns

| Branch | Condition | Side effect | Required test |
|---|---|---|---|
| B1 | seam nil | redirect only | unwired test |
| B2 | restart error | failure redirect | failed restart test |
| B3 | blank success note | fixed note | restart result tests |

## Calls and live bindings

`RestartSoak` is called once only after ready preflight or successful save.
`redirectDashboard` preserves PRG.

## State mutations and fallbacks

No credential or account mutation occurs in this package.

## Safety conclusion

- Safe edit boundary: extracted original restart body, called under the
  credential-generation mutex.
- High-risk impact: process start only; no LIVE account operation.
