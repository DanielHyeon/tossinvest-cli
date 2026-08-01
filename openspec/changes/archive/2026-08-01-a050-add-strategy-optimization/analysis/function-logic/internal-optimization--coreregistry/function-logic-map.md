# Function Logic Map: `CoreRegistry`

- Source: `internal/optimization/providers.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| approved core descriptor manifest | exact a041 `exit.common-policy` owner/category/key tuple | owner package descriptor plus a050 manifest | construction fails closed on omission, duplication, or drift |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | current implementation has no branch; descriptor composition and manifest validation are delegated | no external mutation | registry or validation error | `TestCoreRegistryCoversExactApprovedDescriptorManifest` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.CommonPolicyFieldDescriptor` | obtains owner-authored finite option metadata | pure, no retry; invalid descriptor fails in registry build | current AST + descriptor tests |
| `BuildRegistry` / required coverage validation | exact-one-owner composition and completeness check | fail closed, no fallback descriptor invention | registry tests |

## State mutations and fallbacks

- Constructs only immutable in-memory metadata. It never reads or mutates broker, order, journal, lane, gate, LIVE, or position state.
- Missing future owner adapters remain explicitly read-only in their owner views; the manifest covers only transport-neutral `settingmeta` providers present at this HEAD.

## Safety conclusion

- Safe edit boundary: a050 core descriptor composition and exact approved manifest only.
- High-risk impact: yes; an omitted or drifted writable field must disable construction rather than silently disappear.
