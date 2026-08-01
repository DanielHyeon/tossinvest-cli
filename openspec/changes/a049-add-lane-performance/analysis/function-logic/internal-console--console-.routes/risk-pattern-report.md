# Risk Pattern Report: `internal/console/console.go`

| Rule | Location | Message |
|---|---|---|
| go-panic | `internal/console/console.go:931` | panic can bypass normal error and shutdown handling; map the recovery boundary. |

> Findings are review candidates, not automatic defect verdicts.

The single panic finding is outside `Console.routes` (the mapped function ends at line 776). It is pre-existing template parsing startup behavior and is not reached or changed by the `/performance-history` route registration. Classification: not applicable to this function edit.
