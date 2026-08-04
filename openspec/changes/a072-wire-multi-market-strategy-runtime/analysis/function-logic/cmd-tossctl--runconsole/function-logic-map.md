# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The authenticated console binds only read capabilities for strategy runtime and performance. Missing journal, performance, RPC, TLS, or optional operator data remains typed unavailable; none of those branches may create order or activation authority.

## Branches and early returns

B1–B41 cover path resolution, optional read-capability construction, strict TLS/network validation, server construction and shutdown. The new strategy RPC dial and performance reader follow existing optional read-only fallback branches.

## Calls and live bindings

`strategyprojectionrpc.Dial` reads the engine-owned Unix projection. `openConsolePerformanceCapabilities` opens immutable `performance.db`. `console.New` receives interfaces with no broker or operating mutation methods.

## State mutations and fallbacks

Only console-local lifecycle resources and the pre-existing optimization command store are opened. Strategy/performance read failures produce unavailable UI state; no zero or peer-market fallback is synthesized.

## Safety conclusion

The edit adds read surfaces only. Live-order, lane activation and protection mutation capability do not cross this function.
