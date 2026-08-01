# Function Logic Map: `typeCheckPureApprovedCandidateBoundary`

- Source: `internal/candidate/approved_boundary_typecheck_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | repository AST/types or sealed candidate test fixtures, as declared in the signature | current source and persisted a047 base | violation/error/test failure; no approval is minted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 219: `if packageRel == "internal/strategy" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `range` at source line 220: `for _, file := range files {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 222: `if spec, ok := node.(*ast.TypeSpec); ok && spec.Name.Name == "ApprovedSnapshot" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `range` at source line 230: `for _, file := range files {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `range` at source line 231: `for _, spec := range file.Imports {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `if` at source line 233: `if path != candidatePath {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `if` at source line 255: `if checkErr != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B8 | exact AST `range` at source line 256: `for _, typeErr := range typeErrors {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B9 | exact AST `if` at source line 259: `if len(typeErrors) == 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B10 | exact AST `range` at source line 264: `for _, file := range files {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B11 | exact AST `type-switch` at source line 267: `switch value := node.(type) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B12 | exact AST `case` at source line 268: `case *ast.GenDecl:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B13 | exact AST `if` at source line 269: `if value.Tok == token.VAR {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B14 | exact AST `if` at source line 270: `if _, topLevel := parents[value].(*ast.File); topLevel {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B15 | exact AST `range` at source line 271: `for _, spec := range value.Specs {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B16 | exact AST `range` at source line 273: `for _, name := range declaration.Names {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B17 | exact AST `case` at source line 279: `case *ast.TypeSpec:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B18 | exact AST `if` at source line 280: `if object := info.Defs[value.Name]; object != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B19 | exact AST `case` at source line 284: `case *ast.FuncDecl:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B20 | exact AST `if` at source line 285: `if sealedSnapshotBoundary && !allowedApprovedSnapshotDeclaration(value) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B21 | exact AST `if` at source line 288: `if value.Recv != nil && !approvedSnapshotMethod(value) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B22 | exact AST `if` at source line 291: `if value.Name.Name == "init" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B23 | exact AST `if` at source line 294: `if function, ok := info.Defs[value.Name].(*types.Func); ok {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B24 | exact AST `if` at source line 295: `if signature, ok := function.Type().(*types.Signature); ok {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B25 | exact AST `if` at source line 296: `if receiver := signature.Recv(); receiver != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B26 | exact AST `for` at source line 300: `for index := 0; index < signature.Params().Len(); index++ {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B27 | exact AST `for` at source line 305: `for index := 0; index < signature.Results().Len(); index++ {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B28 | exact AST `if` at source line 308: `if name == "" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B29 | exact AST `if` at source line 314: `if parameters := signature.TypeParams(); parameters != nil && parameters.Len() != 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B30 | exact AST `case` at source line 319: `case *ast.ValueSpec:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B31 | exact AST `range` at source line 320: `for _, name := range value.Names {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B32 | exact AST `if` at source line 321: `if object := info.Defs[name]; object != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B33 | exact AST `case` at source line 326: `case *ast.AssignStmt:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B34 | exact AST `range` at source line 327: `for _, left := range value.Lhs {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B35 | exact AST `if` at source line 328: `if kind := forbiddenAssignmentKind(left); kind != "" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B36 | exact AST `range` at source line 332: `for index, right := range value.Rhs {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B37 | exact AST `case` at source line 336: `case *ast.IncDecStmt:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B38 | exact AST `if` at source line 337: `if kind := forbiddenAssignmentKind(value.X); kind != "" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B39 | exact AST `case` at source line 340: `case *ast.SendStmt:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B40 | exact AST `case` at source line 342: `case *ast.GoStmt:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B41 | exact AST `case` at source line 344: `case *ast.DeferStmt:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B42 | exact AST `case` at source line 346: `case *ast.FuncLit:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B43 | exact AST `case` at source line 348: `case *ast.TypeAssertExpr:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B44 | exact AST `case` at source line 351: `case *ast.CallExpr:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B45 | exact AST `if` at source line 352: `if allowedApprovedAccessorCall(value, info, candidatePath) \|\| allowedPureBuiltinCall(value, info) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B46 | exact AST `case` at source line 356: `case *ast.SelectorExpr:` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B47 | exact AST `if` at source line 358: `if selection == nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B48 | exact AST `if` at source line 361: `if _, method := selection.Obj().(*types.Func); !method {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B49 | exact AST `if` at source line 364: `if call, direct := parents[value].(*ast.CallExpr); direct && call.Fun == value {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| parser/type/guard helpers named in `ast.json` | walks imports, declarations, types, assignments, calls, and control-flow capabilities at the pure approval boundary | no network, timeout, retry, or fallback; parse/type errors fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Mutations are limited to test-local finding/path/type maps and synthetic fixtures; no production candidate, threshold, order, or account state is changed.

## Safety conclusion

- Safe edit boundary: `typeCheckPureApprovedCandidateBoundary` walks imports, declarations, types, assignments, calls, and control-flow capabilities at the pure approval boundary and returns findings or test assertions without granting authority.
- High-risk impact: yes — static guard logic protects the candidate-to-strategy authority boundary.
