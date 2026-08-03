# Function Logic Map: `TestAuthorityOriginRejectsConfiguredTransport`

- Source: `internal/official/client_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| default and explicitly configured clients | only untouched default client authoritative | official.New option provenance | test fails if any configured client mints token |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | untouched default client is not authoritative or shares global transport | none | `t.Fatal` | authority-origin test |
| B2 | iterate configured client cases | none | continue matrix | authority-origin test |
| B3 | a configured case mints authority | none | `t.Fatalf` | authority-origin test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Client.AuthorityOrigin | verify opaque provenance result | pure | AST |

## State mutations and fallbacks

- Covers same-string base override and `http.DefaultClient` override, not only visibly different endpoints.

## Safety conclusion

- Safe edit boundary: authority is constructor provenance, not equality of user-supplied values.
- High-risk impact: yes, negative test for arbitrary FX minting.
