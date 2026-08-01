# Function Logic Map: `Console.handleProtectionApply`
- Source: `internal/console/optimization.go`; evidence: `ast.json`, `risk-pattern-report.md`.
## Inputs and invariants
- Accepts only an opaque preview capability plus a fixed checkbox; shared middleware enforces POST and CSRF before this handler.
## Branches and early returns
- B1 rejects unwired seam; B2 rejects missing capability; B3-B8 map early, stale, unchecked, and other failures to explicit non-success responses.
## Calls and live bindings
- Calls only `ProtectionCommander.Apply`; the seam owns one-shot scope and the 3-second timing check.
## State mutations and fallbacks
- Successful application redirects with status; failures preserve existing protection and render no arbitrary input field.
## Safety conclusion
- Weakening remains capability-bound, checkbox-confirmed, delayed, and default unavailable when unwired.
