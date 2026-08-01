# Function Logic Map: `BuildRegistry`

- Source: `internal/optimization/registry.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provider bindings | setting categories only, non-nil provider, one provider per trimmed owner | owner package provider | `ErrInvalidRegistry` with no partial registry |
| descriptors | valid finite `settingmeta` descriptors, provenance owner exact, key unique | owner descriptor | `ErrInvalidRegistry`; no invented default/option |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate bindings | in-memory maps only | continue or first validation error | registry matrix tests |
| B2 | nil provider | none | invalid registry | `TestRegistryRejectsNilAndEmptyProviders` |
| B3 | unknown/read-only category | none | invalid registry | `TestRegistryRejectsNilAndEmptyProviders` |
| B4 | blank owner | none | invalid registry | `TestRegistryRejectsNilAndEmptyProviders` |
| B5 | duplicate provider owner | none | invalid registry | `TestRegistryRequiresExactlyOneMatchingOwnerForEveryField` |
| B6 | provider descriptor read fails | none | wrapped invalid registry | provider error test |
| B7 | provider returns no descriptors | none | invalid registry | empty provider test |
| B8 | iterate descriptors | clones accepted metadata | continue or reject | registry matrix tests |
| B9 | descriptor has no stable key | none | invalid registry | missing identity test |
| B10 | descriptor provenance owner differs | none | invalid registry | liar provider test |
| B11 | duplicate descriptor key | none | invalid registry | duplicate key test |
| B12 | known descriptor is incomplete | stores only a cloned read-only field with no options and explicit configuration error | registry remains readable; preview cannot target field | `TestIncompleteKnownDescriptorIsExposedReadOnlyWithConfigurationError`, store zero-candidate test, console DOM test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Provider.Descriptors` | fetches owner-authored metadata | one call per owner, no retry; any error aborts | current AST + provider error test |
| `FieldDescriptor.Validate` | enforces finite control/default/provenance contract | a known owner/key validation error is exposed as read-only configuration error with options removed; missing identity/owner still aborts | settingmeta, registry, store, and console tests |
| `cloneDescriptor` | prevents caller mutation aliasing | pure | defensive-copy test |

## State mutations and fallbacks

- Only local immutable registry maps are built. Failure returns no partial registry and there is no fallback or synthesized owner metadata.

## Safety conclusion

- Safe edit boundary: owner/category/key uniqueness and descriptor validity.
- High-risk impact: yes; writable option authority derives from this registry, so ambiguous coverage fails closed.
