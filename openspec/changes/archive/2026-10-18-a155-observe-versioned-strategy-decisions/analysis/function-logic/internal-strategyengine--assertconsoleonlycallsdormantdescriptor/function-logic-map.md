# Function Logic Map: `assertConsoleOnlyCallsDormantDescriptor`

Source `internal/strategyengine/wiring_test.go`; base `ast.json` pins source SHA. Existing unchanged descriptor inspector; checker conservatively requires evidence for its adjacent file change.

## Inputs and invariants

This test parses repository Go source and reports violations; it does not perform runtime trading, database, network or settings mutations. Existing descriptor restriction and original allowlist remain. New leaf must only use allowed pure RR identifiers; production import of the observation package is limited to the offline CLI or self-package.

## Branches and early returns

| ID | AST branch | Behavior |
|---|---|---|
| B1 | if, line 66 | Non-call AST nodes are ignored. |
| B2 | if, line 70 | Non-selector calls are ignored. |
| B3 | if, line 74 | Calls not using strategyengine identifier are ignored. |
| B4 | if, line 77 | Non-descriptor call reports a test error. |

## Calls and live bindings

WalkDir, parser and AST inspection operate on files; helper failures produce test errors. No live bindings. CodeGraph and pre-edit AST are in `analysis/code-context`.

## State mutations and fallbacks

Only local test bookkeeping and error reporting.

## Safety conclusion

Gate2 requested this bundle for both functions; the descriptor helper body itself has no diff from base and its AST is pinned to the base revision. The main test AST is pinned to current source. Base AST source SHA is `6298b6` prefix; line references here are base lines 62–82.
