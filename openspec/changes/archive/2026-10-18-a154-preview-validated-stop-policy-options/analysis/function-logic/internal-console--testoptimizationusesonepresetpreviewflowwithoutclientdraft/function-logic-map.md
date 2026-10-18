# Function Logic Map: `TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft`

Source `internal/console/optimization_ui_contract_test.go`; fresh structure `ast.json`; scan `risk-pattern-report.md`.

## Inputs and invariants

Reads rendered synthetic dashboard DOM. New computation section is verified against exact form action, field and control allowlist before removal. The unchanged assertions still inspect all remaining policy controls.

## Branches and early returns

| Branch | Condition | Behavior |
|---|---|---|
| B1 | if line 120: `if err != nil {` | Existing assertion remains; failure reports through testing.T. |
| B2 | if line 123: `if got := len(nodesByName(doc, "form")); got != 3 {` | Existing assertion remains; failure reports through testing.T. |
| B3 | if line 126: `if got := len(nodesByName(doc, "button")); got != 3 {` | Existing assertion remains; failure reports through testing.T. |
| B4 | range line 129: `for _, forbidden := range []string{"select", "textarea", "script"} {` | Existing assertion remains; failure reports through testing.T. |
| B5 | if line 130: `if got := len(nodesByName(doc, forbidden)); got != 0 {` | Existing assertion remains; failure reports through testing.T. |
| B6 | range line 134: `for _, marker := range []string{` | Existing assertion remains; failure reports through testing.T. |
| B7 | if line 140: `if !strings.Contains(page, marker) {` | Existing assertion remains; failure reports through testing.T. |
| B8 | range line 144: `for _, inventedState := range []string{"localStorage", "sessionStorage", "data-dirty", "미저장 변경", "초기화"} {` | Existing assertion remains; failure reports through testing.T. |
| B9 | if line 145: `if strings.Contains(page, inventedState) {` | Existing assertion remains; failure reports through testing.T. |

## Calls and live bindings

Uses test-only fake commander, synthetic httptest harness and HTML parser; no live services.

## State mutations and fallbacks

Only test-local DOM and failure records change; no operational settings are touched.

## Safety conclusion

Scope excludes only a section that has passed strict compute-only verification. Forged setting field, apply button and extra form mutations must fail.
