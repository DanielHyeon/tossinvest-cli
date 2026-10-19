# Function Logic Map: `TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions`

Source `internal/console/optimization_ui_contract_test.go`; fresh structure `ast.json`; scan `risk-pattern-report.md`.

## Inputs and invariants

Reads rendered synthetic dashboard DOM. New computation section is verified against exact form action, field and control allowlist before removal. The unchanged assertions still inspect all remaining policy controls.

## Branches and early returns

| Branch | Condition | Behavior |
|---|---|---|
| B1 | if line 21: `if err != nil {` | Existing assertion remains; failure reports through testing.T. |
| B2 | range line 25: `for _, field := range commander.view.Registry.Fields(strategyopt.CategoryExitProtection) {` | Existing assertion remains; failure reports through testing.T. |
| B3 | range line 26: `for _, option := range field.Descriptor.Options {` | Existing assertion remains; failure reports through testing.T. |
| B4 | if line 31: `if node.Type != html.ElementNode {` | Existing assertion remains; failure reports through testing.T. |
| B5 | switch line 34: `switch node.Data {` | Existing assertion remains; failure reports through testing.T. |
| B6 | case line 35: `case "textarea":` | Existing assertion remains; failure reports through testing.T. |
| B7 | case line 37: `case "input":` | Existing assertion remains; failure reports through testing.T. |
| B8 | if line 39: `if kind != "hidden" && kind != "checkbox" {` | Existing assertion remains; failure reports through testing.T. |
| B9 | if line 44: `if name == "option_id" && !allowedOptions[value] {` | Existing assertion remains; failure reports through testing.T. |
| B10 | if line 48: `if _, found := htmlAttribute(node, "contenteditable"); found {` | Existing assertion remains; failure reports through testing.T. |
| B11 | range line 52: `for _, forbidden := range []string{type="text", type="number", type="range", "typed confirmation", "symbol 입력", "reason 입력"} {` | Existing assertion remains; failure reports through testing.T. |
| B12 | if line 53: `if strings.Contains(strings.ToLower(page), strings.ToLower(forbidden)) {` | Existing assertion remains; failure reports through testing.T. |

## Calls and live bindings

Uses test-only fake commander, synthetic httptest harness and HTML parser; no live services.

## State mutations and fallbacks

Only test-local DOM and failure records change; no operational settings are touched.

## Safety conclusion

Scope excludes only a section that has passed strict compute-only verification. Forged setting field, apply button and extra form mutations must fail.
