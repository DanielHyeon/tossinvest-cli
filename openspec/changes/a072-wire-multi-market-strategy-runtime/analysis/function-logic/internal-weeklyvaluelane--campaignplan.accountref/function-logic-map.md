# Function Logic Map: `CampaignPlan.AccountRef`

- Source: `internal/weeklyvaluelane/plan.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receiver `CampaignPlan` | immutable value copied by value | sealed plan constructor | returns stored account reference; performs no validation or fallback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branch-free happy path | none | stored `p.accountRef` | plan identity and strategy-flow tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct field return | no error, timeout, retry, or I/O | `ast.json` shows no calls |

## State mutations and fallbacks

- No mutation, side effect, fallback, allocation, or external binding.
- The accessor does not infer an account from market, campaign, or runtime defaults.

## Safety conclusion

- Safe edit boundary: read-only exposure of the private, constructor-bound account reference.
- High-risk impact: low locally; downstream execution lineage depends on exact account identity.
