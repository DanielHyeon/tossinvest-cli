# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go` (234-355)
- AST evidence: `ast.json` — AST branches 5.

## Inputs and invariants

Construction follows projection wiring, tracker restore/refresh, alert-entry latch restore, paired readiness, then execution gateway assembly. It does not own worker supervision.

## Branches and early returns

| Branch | Result |
|---|---|
| B1 | projection wiring failure returns |
| B2 | tracker restore failure returns |
| B3 | alert-entry latch restore failure returns |
| B4 | paired readiness failure returns |
| B5 | execution gateway creation failure returns |

## Calls and live bindings

Calls projection guard, restore/refresh, `restoreAlertEntryLatch`, readiness and `execgw.New` in that order.

## State mutations and fallbacks

Restore work precedes return; no degraded gateway is produced.

## Safety conclusion

Task 3.9's worker belongs in `engineRuntime` supervision, not this constructor; preserve all failure ordering.
