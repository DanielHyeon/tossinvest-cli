# Function Logic Map: `Console.handleProtectionPreview`
- Source: `internal/console/optimization.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Accepts only an opaque server-issued row action token after shared POST/CSRF gating; symbols, quantities, triggers, and reasons are not accepted.
## Branches and early returns
- B1 rejects unwired command seam; B2 rejects empty token; B3 rejects preview failure; B4 rejects non-weakening or capability-less results.
## Calls and live bindings
- Calls `ProtectionCommander.Preview` and renders the fixed preview model on success.
## State mutations and fallbacks
- No mutation occurs; every invalid condition returns a fail-closed refusal.
## Safety conclusion
- Preview is server-defined and cannot broaden scope or activate LIVE execution.
