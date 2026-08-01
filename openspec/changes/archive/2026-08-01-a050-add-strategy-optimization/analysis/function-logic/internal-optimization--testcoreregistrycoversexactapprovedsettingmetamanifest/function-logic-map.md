# Function Logic Map: `TestCoreRegistryCoversExactApprovedSettingmetaManifest`

- Source: `internal/optimization/registry_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| core registry | exactly one `exit.common-policy` field owned by a041 in exit protection | a050 approved settingmeta manifest | construction or tuple drift fails test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `CoreRegistry` returns error | test failure only | fatal | this test |
| B2 | count/key/owner/category tuple differs | test failure only | fatal | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `optimization.CoreRegistry`, `Registry.All` | construct and inspect the frozen manifest | one call, no retry | assertions |

## State mutations and fallbacks

- Test-only in-memory registry; no fallback or external mutation.

## Safety conclusion

- Safe edit boundary: approved core manifest regression coverage.
- High-risk impact: no direct side effect; protects writable authority composition.
