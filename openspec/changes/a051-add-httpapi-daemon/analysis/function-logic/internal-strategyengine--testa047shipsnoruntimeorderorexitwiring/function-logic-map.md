# Function Logic Map: `TestA047ShipsNoRuntimeOrderOrExitWiring`

- Source: `internal/strategyengine/wiring_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| repository Go imports and dormant-contract allowlist | strategyengine may be imported only by reviewed dormant/read adapters | a047 wiring guard | unreviewed runtime wiring fails the test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | walk/parser error | none | test fatal | this guard itself |
| B2 | file is directory, vendor, test, or strategyengine itself | none | skip | this guard itself |
| B3 | strategyengine importer is not allowlisted | reports runtime wiring | test error | RED when `internal/httpapi/read.go` first added |
| B4 | console or HTTP API importer calls anything but dormant descriptor | AST selector inspection | test error | focused wiring test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `filepath.WalkDir`, `parser.ParseFile`, `assertConsoleOnlyCallsDormantDescriptor` | enumerate imports and restrict selector calls | parse/walk errors are fatal | AST + focused strategyengine test |

## State mutations and fallbacks

- Test-only allowlist state; no runtime mutation.
- a051 permits `internal/httpapi/read.go` only because it projects `DormantRuntimeDescriptor` and has no lane/order authority.

## Safety conclusion

- Safe edit boundary: add the HTTP read model and apply the same descriptor-only selector assertion as console.
- High-risk impact: yes; a047 must remain dormant, so any other selector fails closed.
