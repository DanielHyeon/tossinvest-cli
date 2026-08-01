# Function Logic Map: `Console.routes`
- Source: `internal/console/console.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Routes are fixed server-owned paths; mutating routes remain POST-only and pass the shared CSRF gate.
- `/optimization/exit-policy` passes `4096` to `mutating`, so `MaxBytesReader` rejects oversized bodies before parsing, CSRF, or command dispatch.
## Branches and early returns
- B1/B2 conditionally add only wired monitoring and decision seams; the protection routes themselves remain registered but fail closed when unwired.
## Calls and live bindings
- Binds handlers through `http.ServeMux.HandleFunc`; static route tests enumerate every mutating seam.
- `c.mutating(c.handleOptimizationSave, 4096)` supplies POST, origin, size, parse, and CSRF gates in that order.
## State mutations and fallbacks
- Registration mutates only the in-memory mux; no trading state or runtime toggle is changed.
## Safety conclusion
- Safe boundary is route registration. Default runtime remains OFF/UNWIRED and handler authorization is tested independently.
