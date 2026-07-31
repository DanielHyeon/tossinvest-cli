# Function Logic Map: `TestDesignatingASymbolFromThePositionsScreen`

- Source: `internal/console/settings_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fake adoption settings and broker/journal fixture | isolated unmanaged symbol `000660` | handler/template contract | fatal/error when action, idempotency, or status is wrong |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | explicit include form/button absent | test fatal | no later POST | rendered action guard |
| B2 | two identical POST attempts | fake settings saves | loop completes twice | idempotency exercise |
| B3 | saved list/count is not one symbol/two writes | test error | continue | idempotency assertion |
| B4 | post-save page lacks reservation note | test error | return | state feedback assertion |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| settingsHarness/seedJournal | isolated console fixture | temp journal and fake settings only | setup |
| dashboardHarness.post | exercise existing CSRF-protected include route | httptest only; no real listener | B2 |
| fakeSettings.saved | inspect captured config block | mutex-protected fake; no retry | B3 |

## State mutations and fallbacks

- Mutates only the fake settings seam in a test temp directory. Does not start engine/soak or place an order.

## Safety conclusion

- Safe edit boundary: changes the rendered control assertion from checkbox to explicit CSP-safe button.
- High-risk impact: no; endpoint behavior and safety assertions are unchanged.
