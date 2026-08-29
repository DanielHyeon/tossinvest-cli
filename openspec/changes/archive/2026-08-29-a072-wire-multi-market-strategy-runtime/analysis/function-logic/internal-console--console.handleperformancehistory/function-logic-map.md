# Function Logic Map: `Console.handlePerformanceHistory`

- Source: `internal/console/performance_history.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Only GET/HEAD is allowed. Dashboard and account-bound attribution are read through immutable interfaces; missing attribution must not erase valid aggregates.

## Branches and early returns

B1–B5 cover method refusal, unwired reader, dashboard failure, dashboard success and independent attribution failure/success.

## Calls and live bindings

Calls fixed `Dashboard` and `readPerformanceAttributions`; rendering receives plain values only.

## State mutations and fallbacks

No form or mutation control exists. Attribution errors are displayed separately and never converted to zero.

## Safety conclusion

The page explains KR/US campaign lineage without acquiring a writer.
