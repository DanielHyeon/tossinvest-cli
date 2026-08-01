# Function Logic Map: `TestRequiredRegistryCoverageRejectsMissingUnexpectedAndDriftedFields`

- Source: `internal/optimization/registry_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| valid manifest/binding | exact a041 key, owner, exit category | frozen test fixture | must construct successfully |
| invalid coverage matrix | empty/blank/missing/extra/drifted/duplicate requirements | table fixture | every case returns typed invalid-registry error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | valid manifest construction fails | test failure only | fatal | this test's valid precondition |
| B2 | iterate invalid coverage cases | test subtests only | continue through matrix | this table-driven test |
| B3 | invalid case is accepted or wrong error returned | test failure only | fail subtest | this table-driven test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `optimization.BuildRequiredRegistry` | verify exact release manifest contract | one call per case; no retry | assertions |

## State mutations and fallbacks

- Test-only fixtures and assertions; no persistence or trading authority.

## Safety conclusion

- Safe edit boundary: exact writable-field coverage regression matrix.
- High-risk impact: no direct side effect; protects a high-risk configuration authority boundary.
