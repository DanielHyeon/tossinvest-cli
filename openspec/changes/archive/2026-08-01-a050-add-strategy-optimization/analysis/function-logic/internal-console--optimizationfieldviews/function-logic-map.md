# Function Logic Map: `optimizationFieldViews`

- Source: `internal/console/optimization_view.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `fields` | zero or more category-scoped `optimization.RegisteredField` values | server-composed optimization registry | empty input returns an empty view slice; no controls are invented |
| `fields[i].Descriptor` | registry-owned key/metadata/options; invalid descriptors arrive as `ControlReadOnly` with `ConfigurationError` | `optimization.BuildRegistry` | configuration error is copied verbatim to the UI projection; owner options remain unavailable |
| `snapshot.Desired` / `snapshot.Effective` | option IDs previously accepted by the registry, or absent/unknown legacy state | immutable optimization snapshot | absent values use descriptor state labels; unknown values render an explicit refusal string |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each `RegisteredField` in `fields` | appends one detached display projection; copies option slice | returns one UI row per registry field in input order | complete field renders label/default/desired/effective/provenance; incomplete field exposes configuration error and no mutation control |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `displayUnit` | distinguish absent unit from a numeric zero | pure formatting; blank becomes `해당 없음` | AST + `optimization_lifecycle_test.go` |
| `displaySettingState` / `displayOption` | resolve server option IDs to operator label and explicit state | pure formatting; unknown option is visibly rejected, never accepted | AST + `optimization_lifecycle_test.go` |
| `displayApplyTiming` / `displaySafety` | translate registry enums without client defaults | pure map lookup; invalid enum yields blank display and remains non-writable upstream | AST + registry validation |
| `append` copy of `d.Options` | keep the view detached from registry-owned option backing arrays | in-memory only; no IO/retry | AST |

## State mutations and fallbacks

- Allocates and mutates only the local `out` slice.
- Does not mutate the registry, snapshot, LIVE/lane/gate state, broker state, or current position snapshots.
- Does not fall back to arbitrary values or create writable controls; template eligibility is derived from the copied `ConfigurationError` and server control kind.

## Safety conclusion

- Safe edit boundary: add presentation-only fields copied from `RegisteredField`; do not broaden option lists or infer configuration validity in the browser projection.
- High-risk impact: no direct trading effect. This is a fail-closed UI projection on the high-risk settings surface; a configuration error must suppress preview controls.
