# Function Logic Map: `protectionPreviewPage.Refresh`
- Source: `internal/console/optimization.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Preview pages contain one-shot capabilities and must never auto-refresh.
## Branches and early returns
- Branchless constant false return prevents replay-prone refresh behavior.
## Calls and live bindings
- No callees or live bindings.
## State mutations and fallbacks
- No mutation or fallback.
## Safety conclusion
- The capability page is stable until deliberate user action and shared expiry checks.
