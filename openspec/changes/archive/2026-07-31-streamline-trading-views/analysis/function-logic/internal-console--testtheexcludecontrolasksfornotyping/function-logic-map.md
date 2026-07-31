# Function Logic Map: `TestTheExcludeControlAsksForNoTyping`

- Source: `internal/console/settings_exclude_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rendered positions page | authenticated fake settings/journal | CSP-safe control contract | one of four assertion branches records failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | switch over forbidden/required control shapes | none | first matching case reports error | this test |
| B2 | `confirm(` remains | test error | switch exits | CSP regression |
| B3 | explicit exclusion button absent | test error | switch exits | action regression |
| B4 | `prompt(` remains | test error | switch exits | typing regression |
| B5 | text input appears | test error | switch exits | one-click regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| settingsHarness/seedJournal | isolated fixture page | temp files and fakes only | test setup |
| strings.Contains | inspect rendered bytes | pure scan | AST B2-B5 |

## State mutations and fallbacks

- Test fixture state only. No POST is issued and no production setting is changed.

## Safety conclusion

- Safe edit boundary: replaces obsolete `confirm()` expectation with the deployed CSP contract.
- High-risk impact: no; test-only render assertion.
