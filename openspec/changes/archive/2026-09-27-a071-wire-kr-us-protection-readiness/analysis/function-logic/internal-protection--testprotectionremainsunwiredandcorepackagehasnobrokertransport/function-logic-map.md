# Function Logic Map: `TestProtectionRemainsUnwiredAndCorePackageHasNoBrokerTransport`

- Source: `internal/protection/dormant_test.go` (14-88)
- Revision: current — HEAD `648df8ef`; source_sha256 `0f715a4e134e7477c6ca042152def330ad9255a8e2b74570a38847bb4ad10640`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 20 · returns 4 · calls 41
- Exact AST return positions: 48:5, 51:5, 55:5, 67:4
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs | typed values | TestProtectionRemainsUnwiredAndCorePackageHasNoBrokerTransport | fail closed or test failure |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 15:2 | `if execgw.ProfileProtection != execgw.ProtectionUnwired {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | if at 22:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B3 | range at 25:2 | `for _, entry := range entries {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B4 | if at 26:3 | `if entry.IsDir() \|\| !strings.HasSuffix(entry.Name(), ".go") \|\| strings.HasSuffix(entry.Name(), "_test.g...` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B5 | if at 31:3 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B6 | range at 34:3 | `for _, imp := range file.Imports {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B7 | if at 36:4 | `if name == "net/http" \|\| strings.Contains(name, "/internal/official") \|\| strings.Contains(name, "/inter...` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B8 | range at 45:2 | `for _, dir := range []string{filepath.Join(root, "cmd"), filepath.Join(root, "internal", "app")} {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B9 | if at 47:4 | `if walkErr != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B10 | if at 50:4 | `if d.IsDir() \|\| !strings.HasSuffix(path, ".go") \|\| strings.HasSuffix(path, "_test.go") {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B11 | if at 54:4 | `if parseErr != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B12 | range at 57:4 | `for _, imp := range file.Imports {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B13 | if at 59:5 | `if name == "github.com/JungHoonGhae/tossinvest-cli/internal/protection" {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B14 | if at 62:6 | `if !allowed[filepath.ToSlash(rel)] {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B15 | if at 69:3 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B16 | if at 74:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B17 | range at 78:2 | `for _, required := range []string{"protectionreadiness.NewProductionProvider", "protection.NewPairedReadine...` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B18 | if at 79:3 | `if !strings.Contains(text, required) {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B19 | range at 83:2 | `for _, forbidden := range []string{"protection.NewSupervisor", "protectionofficial.New", "protection.db", "...` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B20 | if at 84:3 | `if strings.Contains(text, forbidden) {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | mapped AST control flow | bounded to function | typed return | affected regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| mapped dependencies | preserve function contract | caller handles error | CodeGraph + AST |

## State mutations and fallbacks

- No authority broadening; current behavior is covered by focused tests.

## Safety conclusion

- Safe edit boundary: TestProtectionRemainsUnwiredAndCorePackageHasNoBrokerTransport only.
- High-risk impact: reviewed and regression-tested.
