# Function Logic Map: `runHTTPAPI`

- Source: `cmd/tossctl/httpapi.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The private API is authenticated and read-only for strategy runtime and performance. Projection connectivity, journal schema, TLS and bind policy fail independently without opening an account mutation path.

## Branches and early returns

B1–B21 cover input validation, journal/performance/optimization capability construction, strategy projection RPC dialing, TLS serving and bounded shutdown. Missing optional reads are retained as typed errors in the reader.

## Calls and live bindings

The function injects `strategyprojection.Reader`, immutable performance reads and `performancejournal.New(journalReader)` into `httpAPIReader`; router methods remain GET/SSE only for these resources.

## State mutations and fallbacks

It owns process-local listeners and the existing optimization control store. Runtime/performance failures never default to current/zero values and never start the engine.

## Safety conclusion

No broker, Gateway, lane activation, protection weakening or LIVE approval method is introduced by the new bindings.
