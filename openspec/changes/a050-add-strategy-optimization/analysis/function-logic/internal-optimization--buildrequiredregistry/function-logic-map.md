# Function Logic Map: `BuildRequiredRegistry`

- Source: `internal/optimization/registry.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| required manifest | non-empty, unique trimmed key, non-empty trimmed owner, fixed category | release composition in `CoreRegistry` | `ErrInvalidRegistry`; no partial registry |
| provider bindings | exact-one-owner validated descriptor set | `BuildRegistry` | propagated fail-closed error |
| coverage | registry field count and every key/owner/category exactly match manifest | a050 frozen release manifest | omission, surprise field, owner drift, or category drift rejected |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | required manifest is empty | none | invalid registry | coverage test `empty manifest` case |
| B2 | iterate required entries | local map only | validate all requirements | valid/invalid manifest cases |
| B3 | trimmed key or owner is blank | none | invalid registry | blank key/owner cases |
| B4 | required key is duplicated | none | invalid registry | duplicate requirement case |
| B5 | ordinary registry construction fails | none | propagate error | provider validation suite |
| B6 | field count differs from manifest count | none | invalid registry | missing and unexpected cases |
| B7 | iterate exact manifest keys | none | compare every field | valid/drift cases |
| B8 | key absent, category differs, or provenance owner differs | none | invalid registry | missing/wrong owner/wrong category cases |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `BuildRegistry` | performs provider/category/key/provenance/descriptor validation first | one in-memory call, error propagated, no retry | registry validation tests |
| `strings.TrimSpace` | canonicalizes manifest identity | pure | blank/duplicate manifest tests |

## State mutations and fallbacks

- Only local requirement and registry maps are built. No persisted, trading, journal, lane, gate, or LIVE state changes.
- There is no fallback that accepts missing or extra writable fields; all manifest drift fails closed.

## Safety conclusion

- Safe edit boundary: exact release-time writable-field coverage only.
- High-risk impact: yes; this prevents silent widening/narrowing of the setting control surface.
