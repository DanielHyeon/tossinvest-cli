# Function Logic Map: `TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed`

Source `internal/console/optimization_ui_contract_test.go`; fresh structure `ast.json`; scan `risk-pattern-report.md`.

## Inputs and invariants

Reads rendered synthetic dashboard DOM. New computation section is verified against exact form action, field and control allowlist before removal. The unchanged assertions still inspect all remaining policy controls.

## Branches and early returns

| Branch | Condition | Behavior |
|---|---|---|
| B1 | range line 229: `for _, want := range []string{"lifecycle을 읽지 못했다", "모든 변경을 닫았다", "미검증/사용 불가"} {` | Existing assertion remains; failure reports through testing.T. |
| B2 | if line 230: `if !strings.Contains(page, want) {` | Existing assertion remains; failure reports through testing.T. |
| B3 | if line 234: `if strings.Contains(page, <form) \|\| strings.Contains(page, <button) {` | Existing assertion remains; failure reports through testing.T. |

## Calls and live bindings

Uses test-only fake commander, synthetic httptest harness and HTML parser; no live services.

## State mutations and fallbacks

Only test-local DOM and failure records change; no operational settings are touched.

## Safety conclusion

Scope excludes only a section that has passed strict compute-only verification. Forged setting field, apply button and extra form mutations must fail.
