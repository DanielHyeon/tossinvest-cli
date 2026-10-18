# Function Logic Map: `TestA047ShipsNoRuntimeOrderOrExitWiring`

Source `internal/strategyengine/wiring_test.go`; current `ast.json` pins source SHA. Repository-wide test-only import guard; exact offline leaf restriction is new, all previous dormant protections remain.

## Inputs and invariants

This test parses repository Go source and reports violations; it does not perform runtime trading, database, network or settings mutations. Existing descriptor restriction and original allowlist remain. New leaf must only use allowed pure RR identifiers; production import of the observation package is limited to the offline CLI or self-package.

## Branches and early returns

| ID | AST branch | Behavior |
|---|---|---|
| B1 | if, line 27 | Walk error propagates. |
| B2 | if, line 30 | Directory traversed without parsing. |
| B3 | if, line 31 | Skip .git and vendor trees. |
| B4 | if, line 36 | Skip non-Go and test files. |
| B5 | if, line 41 | Parse error propagates. |
| B6 | range, line 44 | Inspect every production import. |
| B7 | if, line 46 | Invalid Go import literal propagates. |
| B8 | if, line 49 | Engine imports outside engine package enter dormant guard. |
| B9 | if, line 51 | Unknown engine importer fails test. |
| B10 | if, line 54 | Exact observation leaf enters restricted selector guard. |
| B11 | if, line 55 | Forbidden or unnamed engine references fail test. |
| B12 | if, line 59 | Existing console and HTTP importers retain descriptor-only check. |
| B13 | if, line 64 | Runtime observation import outside offline CLI fails test. |
| B14 | if, line 69 | Walk/parse error fails test. |

## Calls and live bindings

WalkDir, parser and AST inspection operate on files; helper failures produce test errors. No live bindings. CodeGraph and pre-edit AST are in `analysis/code-context`.

## State mutations and fallbacks

Only local test bookkeeping and error reporting.

## Safety conclusion

Gate2 requested this bundle for both functions; the descriptor helper body itself has no diff from base and its AST is pinned to the base revision. The main test AST is pinned to current source.
