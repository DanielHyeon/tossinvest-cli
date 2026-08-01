# Function Logic Map: `missingEvidence`

- Source: `internal/optimizationevidence/provider.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| dashboard state counts | non-negative a049 derived counts | `performance.DashboardView.States` | zero complete or any incomplete state adds an explicit reason |
| aggregates | complete lineage, current semantics, at least default minimum samples | a049 aggregate contract | malformed/incomplete rows add reasons; they are never silently skipped |
| metrics | exactly one complete non-blank summary for every required key | frozen `requiredMetricKeys` | missing, duplicate, incomplete, undersampled, or blank value adds the metric key |
| output | unique, stable lexical order | local set plus `sort.Strings` | deterministic list suitable for audit/digest display |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no complete rows | add `complete-lineage` | continue collecting all reasons | lineage test |
| B2 | complete count below minimum, insufficient rows present, or aggregates empty | add `minimum-sample` | continue | minimum sample matrix |
| B3 | link-missing rows present | add `link_missing` | continue | link-missing test |
| B4 | not-measured rows present | add `not_measured` | continue | not-measured test |
| B5 | iterate every aggregate | local set only | inspect all rows | multi-aggregate test |
| B6 | any market/lane/policy lineage component is blank | add `complete-lineage` | continue | incomplete lineage matrix |
| B7 | aggregate status is not complete or samples are below minimum | add `minimum-sample` | continue | aggregate status/sample test |
| B8 | semantics version differs | add `semantics-version` | continue | semantics drift test |
| B9 | iterate aggregate metrics | build key map and duplicate set | continue | complete metric matrix |
| B10 | metric key already appeared | mark duplicate | continue | duplicate metric test |
| B11 | iterate every frozen required metric key | none | inspect coverage | required metric test |
| B12 | required metric missing, duplicated, incomplete, undersampled, or blank | add `required-metric:<key>` | continue | required metric defect matrix |
| B13 | iterate unique missing reasons | append to output | sorted output | deterministic-order test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | rejects blank lineage and metric values | pure | malformed lineage/metric tests |
| `performance.DefaultMinimumSample` | shares the a049 conservative sample threshold | compile-time binding; no caller override | minimum sample tests |
| `sort.Strings` | makes missing reasons deterministic | pure | deterministic output assertion |

## State mutations and fallbacks

- Mutations are confined to local maps/slices. Duplicate reasons collapse but distinct defects remain visible.
- No defect is repaired, defaulted, or promoted to complete; callers receive every detected reason in stable order.

## Safety conclusion

- Safe edit boundary: conservative validation of derived performance evidence only.
- High-risk impact: yes; false completeness could enable a recommendation, so unknown or malformed inputs are insufficient.
