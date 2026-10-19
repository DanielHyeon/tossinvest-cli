# Function Logic Map: `TestSchemaTablesAndColumns`

Source: `internal/journal/schema_test.go`. Base and current AST in `pre-edit-ast.json` and `ast.json`; current name `TestSchemaTablesAndColumns`.

## Inputs and invariants

Existing test-only schema, fixture or AST census contract remains strict. Historical migration rows remain unchanged; supported weekly admission uses a genuine verified v4 projection.

## Branches and early returns

All prior assertions, mutation checks and error branches are retained. The current AST enumerates them.

## Calls and live bindings

Tests use local journals, immutable test-seam fixtures and production validation.

## State mutations and fallbacks

Only test fixtures write temporary journal state. No live operation.

## Safety conclusion

Repair pins the additional a150 contract without weakening existing checks. The first aggregate gate failure and targeted passing regressions are recorded in implementation evidence; the repeat gate is pending.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
