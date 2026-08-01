# Function Logic Map: `Console.handlePositionManagement`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| position states | engine-owned typed states | read-only RPC list | render disconnected/load error without mutation |
| row actions | fixed registered policies and typed lifecycle eligibility | server registry/state | omit ineligible lifecycle actions |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | commander absent or list fails | render status/error only | return | console tests |
| B3-B4 | each managed state | add fixed policy actions | continue | management page tests |
| B5-B6 | fixed registered policies | add no-input override choices | continue | policy choice tests |
| B7 | managed external-lifecycle state | add RELEASE danger action | continue | provenance release test |
| B8 | released external-lifecycle state | add READOPT danger action | continue | eligibility tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RegisteredCommonPolicies` | produce fixed StockOS-like choices without arbitrary input | registry only | AST |
| `ExternalLifecycleEligible` | gate release/readopt lifecycle authority | typed provenance+eligibility | AST |

## State mutations and fallbacks

- Policy changes remain fixed buttons; provenance and lifecycle eligibility are visible in each row.

## Safety conclusion

- Safe edit boundary: engine-entry rows may change policy but never receive release; lifecycle actions require typed external eligibility.
- High-risk impact: yes
