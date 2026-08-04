# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| resolved account/config paths | exact official account and canonical config directory | existing startup assembly | fail closed and close already-owned resources |
| protection readiness | paired read-only provider from pinned manifest; lifecycle assemblies always UNWIRED | buildGateway | never blocks safety runtime for a protection DB |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing order-path/audit/account/journal/gateway/guardian/interlock error | bounded journal cleanup only | startup refusal | engine regressions |
| B2 | successful assembly | context owns journal and safety runtime only | context | production assembly test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `buildGateway` | assemble normal execution stack and read-only protection refusal | no protection mutation path | CodeGraph + AST |

## State mutations and fallbacks

- No lane, gate, autostart or LIVE setting is changed. No protection supervisor exists in Context.

## Safety conclusion

- Safe edit boundary: pass canonical path/pin into read-only provider and preserve journal cleanup order.
- High-risk impact: yes — production engine assembly.
