# Function Logic Map: `SaveCredentials`

- Source: `internal/official/credentials.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| target path | resolved credential path | caller/config paths | any filesystem error is returned |
| credential value | JSON-marshalable fixed struct | submitted or CLI credentials | values are never included in errors |
| final file | regular file, mode 0600 | atomic temporary-file replacement | verification failure is returned |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | parent creation fails | none | error |
| B2 | JSON marshal fails | none | error |
| B3 | temporary file creation fails | none | error |
| B4 | deferred cleanup sees open file | close/remove temporary | function result unchanged |
| B5 | chmod fails | cleanup temporary | error |
| B6 | write fails | cleanup temporary | error |
| B7 | file fsync fails | cleanup temporary | error |
| B8 | close fails | cleanup temporary | error |
| B9 | atomic rename fails | cleanup temporary | error |
| B10 | final lstat fails | marker-owning caller remains fail closed | error |
| B11 | final target is not regular 0600 | marker-owning caller remains fail closed | fixed error |
| B12 | parent directory open fails | target replaced, caller remains fail closed | error |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.CreateTemp` / `Chmod` | create a new owner-only generation | errors clean temporary and return | CodeGraph + AST |
| `Write` / `Sync` / `Close` | durably finish file before publication | any error prevents rename | CodeGraph + AST |
| `os.Rename` | atomically replace permissive or symlink targets | failure leaves old target and returns | CodeGraph + AST |
| `os.Lstat` | verify regular 0600 result | mismatch fails closed | CodeGraph + AST |
| parent `Sync` | durably publish directory entry | failure returned to marker-owning caller | CodeGraph + AST |

## State mutations and fallbacks

- The old target is untouched until a complete 0600 temporary generation is
  fsynced and closed.
- Rename publishes the completed file atomically; final mode/type is verified.
- Any post-publication error is still returned, so console onboarding retains
  its pending marker and starts no soak.

## Safety conclusion

- Safe edit boundary: replace `os.WriteFile` with an atomic owner-only
  temporary-file publication while preserving the JSON schema and API.
- High-risk impact: yes — this function persists API secrets for every caller.
