# Function Logic Map: `TestWireBodyFitsTheOpenAPIRequestSchema`

- Source: `tools/a087-market-sell-probe/contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| stored OpenAPI snapshot | verified public document | `docs/migration/openapi.latest.json` | test fails on schema or wording drift |
| assembled a087 order | fixed KR numeric market sell probe | existing executable policy | test fails on schema mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | request variant/ref/body construction validity | none | fatal test failure | this test |
| B7-B16 | required fields, enums, patterns, lengths | none | error collection | this test |
| B17-B20 | quantity and MARKET field restrictions | none | error collection | this test |
| B21 | public symbol description wording | none | error on drift | this test |
| B22-B23 | idempotency wording and confirmation value | none | error on drift | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openapiSpec` | read stored public snapshot | local file only | AST B1 |
| `assembleOrder` | serialize fixed order probe | no live request | AST B5 |
| `strings.Contains` | compare public wording | local assertion | AST B20-B22 |

## State mutations and fallbacks

- Test-only snapshot comparison. The a087 production numeric-only symbol guard, execution confirmation and mutation fixtures are unchanged.

## Safety conclusion

- Safe edit boundary: the expected published description substring only.
- High-risk impact: no production behavior edit; full a087 package regression retained.
