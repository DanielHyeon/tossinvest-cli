# Function Logic Map: `TestTheStatusColumnHeaderSaysAdoption`

- Source: `internal/console/portfolio_label_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Rendered `/positions` page | Authenticated settings harness with seeded journal | `settingsHarness`, positions template | Test fails if compact headers or adoption status disappear |
| Unmanaged symbol | `000660` fixture remains outside adoption | Seed journal + fake settings | `rowFor` fails if the row is absent |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Iterate required compact headers | none | Continue | This test, header loop |
| B2 | A required compact header is absent | `testing.T` failure only | Continue collecting failures | This test, header assertion |
| B3 | Dedicated `관리 편입` header remains | `testing.T` failure only | Continue | This test, legacy-column assertion |
| B4 | Instrument row no longer carries `관리 외(미편입)` | `testing.T` failure only | Continue | This test, unmanaged-row assertion |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `settingsHarness` / `seedJournal` / `authenticate` | Create a deterministic authenticated console fixture | Test helper fails immediately on setup error | AST + source |
| `page` | Render the positions page without network access | Test helper fails immediately on HTTP/render error | AST + source |
| `rowFor` | Isolate the requested holding row | Test helper fails when symbol is missing | Source |

## State mutations and fallbacks

- Only test-local fixture state is created. There are no production mutations,
  broker calls, live orders, configuration writes, or retries.

## Safety conclusion

- Safe edit boundary: assertions describing the compact table contract only.
- High-risk impact: no; test-only contract migration.
