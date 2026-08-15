# Function Logic Map: `TestProductionAPIExportsNoAuthorityMintingFunction`

- Source: `internal/protectionlifecycle/external_api_test.go` (11-30)
- AST evidence: `ast.json` — AST branches 6.

## Inputs and invariants

This external-package guard reads every non-test Go file in `internal/protectionlifecycle`, parses its package scope, and forbids exported production functions. It protects the A100 boundary: lifecycle code may not mint authority that runtime/wiring must hold.

## Branches and early returns

| Branch | Result | Existing / planned test |
|---|---|---|
| B1 | directory read failure is fatal | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B2 | each directory entry is inspected | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B3 | directories, non-Go, and tests are skipped | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B4 | parse failure is fatal | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B5 | package scope objects are inspected | `TestProductionAPIExportsNoAuthorityMintingFunction` |
| B6 | exported production function is fatal | `TestProductionAPIExportsNoAuthorityMintingFunction` |

## Calls and live bindings

Uses `os.ReadDir`, Go parser/token APIs, and package-scope objects. It has no runtime authority and does not execute production code.

## State mutations and fallbacks

Test-only read path; no fallback. A false negative would allow a lifecycle API to bypass the manager/runtime authority boundary.

## Safety conclusion

Keep this guard intact. A100 must place worker ownership/authority in the application runtime and journal seams, not expose an authority-minting lifecycle function.
