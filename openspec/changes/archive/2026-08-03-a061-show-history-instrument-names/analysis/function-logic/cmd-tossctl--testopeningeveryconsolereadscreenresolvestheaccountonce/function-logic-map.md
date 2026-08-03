# Function Logic Map: `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`

- Source: `cmd/tossctl/console_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fake broker factory | counts account-resolved client builds | test-local replacement of `verifyBrokerFactory` | restored by cleanup |
| runtime seam list | every account-scoped read screen | same constructors as `runConsole` | any error fails the test |
| static seam table | constructor name plus shared-resolver argument index | explicit allowlist in this test | omitted/new seam fails wiring assertion |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B12 | serial and concurrent runtime reads | count factory calls and collect errors | any screen error/count != 1 fails | this test |
| B13-B21 | parse production wiring and enumerate factory build sites | no production mutation | unargued/stale site fails | this test |
| B22-B30 | locate one shared resolver and each named seam argument | no production mutation | missing/different holder fails | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newConsoleBroker` and seam constructors | reproduce production dependency wiring | no live credentials/network; factory is replaced | AST + fake server |
| Go parser/AST inspection | bind runtime claim to `runConsole` source | parse failure is fatal | AST branches |

## State mutations and fallbacks

- Mutates only test-local counters, fakes, and the temporary global factory restored by cleanup.

## Safety conclusion

- Safe edit boundary: add the instrument-name constructor to the exhaustive shared-resolver seam table and preserve all account resolution assertions.
- High-risk impact: test-only, but guards the production auth/rate-budget boundary.
