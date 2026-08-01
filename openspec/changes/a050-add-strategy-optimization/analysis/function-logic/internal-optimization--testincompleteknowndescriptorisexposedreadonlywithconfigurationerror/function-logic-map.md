# Function Logic Map: `TestIncompleteKnownDescriptorIsExposedReadOnlyWithConfigurationError`

- Source: `internal/optimization/registry_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| known malformed descriptor | stable key/owner but missing description/apply timing | fixture | remains visible only as read-only, no options, explicit configuration error |
| identity-less descriptor | blank stable key | fixture | entire registry construction fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | safe known malformed field is rejected wholesale | test failure only | fatal | this test |
| B2 | projection is absent, writable, has no error, or retains options | test failure only | fatal | this test |
| B3 | blank identity is accepted | test failure only | fatal | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `optimization.BuildRegistry`, `Registry.Field` | verify safe visibility and identity fail-closed behavior | no retry | assertions |

## State mutations and fallbacks

- Test-only descriptor copies; the only permitted fallback is a non-actionable read-only projection for a still-identifiable field.

## Safety conclusion

- Safe edit boundary: malformed metadata display without writable controls.
- High-risk impact: no direct side effect; guards fail-closed configuration behavior.
