# Function Logic Map: `TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls`

Source `internal/console/optimization_ui_contract_test.go`; fresh structure `ast.json`; scan `risk-pattern-report.md`.

## Inputs and invariants

Reads rendered synthetic dashboard DOM. New computation section is verified against exact form action, field and control allowlist before removal. The unchanged assertions still inspect all remaining policy controls.

## Branches and early returns

| Branch | Condition | Behavior |
|---|---|---|
| B1 | if line 154: `if !ok {` | Existing assertion remains; failure reports through testing.T. |
| B2 | if line 163: `if err != nil {` | Existing assertion remains; failure reports through testing.T. |
| B3 | if line 172: `if err != nil {` | Existing assertion remains; failure reports through testing.T. |
| B4 | range line 175: `for _, marker := range []string{"설정 오류 · 읽기 전용", "owner descriptor를 바로잡기 전", "읽기 전용 · owner descriptor 확인 필요"} {` | Existing assertion remains; failure reports through testing.T. |
| B5 | if line 176: `if !strings.Contains(page, marker) {` | Existing assertion remains; failure reports through testing.T. |
| B6 | if line 180: `if got := len(nodesByAttribute(doc, "fieldset", "aria-readonly", "true")); got != 1 {` | Existing assertion remains; failure reports through testing.T. |
| B7 | range line 183: `for _, element := range []string{"form", "input", "button", "select", "textarea"} {` | Existing assertion remains; failure reports through testing.T. |
| B8 | if line 184: `if got := len(nodesByName(doc, element)); got != 0 {` | Existing assertion remains; failure reports through testing.T. |

## Calls and live bindings

Uses test-only fake commander, synthetic httptest harness and HTML parser; no live services.

## State mutations and fallbacks

Only test-local DOM and failure records change; no operational settings are touched.

## Safety conclusion

Scope excludes only a section that has passed strict compute-only verification. Forged setting field, apply button and extra form mutations must fail.
