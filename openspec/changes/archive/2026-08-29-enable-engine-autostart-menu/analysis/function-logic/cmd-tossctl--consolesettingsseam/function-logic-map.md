# Function Logic Map: `consoleSettingsSeam`

- Source: `cmd/tossctl/console.go`
- Qualified function: `consoleSettingsSeam`
- Revision: `base` (function is removed/replaced in the current revision)
- AST evidence: `ast.json` (`e6159d29ec6c306f4a9a942c50e634235683d592e393c0f37ccb6eac6bb92a81`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| base function inputs | base signature and pre-change contracts | persisted base commit plus `ast.json` | replacement must retain compatibility or fail closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | base `if` at `cmd/tossctl/console.go:352` | base behavior is preserved or deliberately replaced by the adjacent current function | existing base return/error contract | base-revision compatibility and affected-package regression tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| base AST call list | define the pre-change behavior that the replacement must preserve | no new fallback or capability may appear during replacement | base commit + `ast.json` |

## State mutations and fallbacks

- This evidence binds the removed/replaced base function; current behavior is covered by its adjacent new-function tests and maps.
- Compatibility remains fail-closed and does not alter LIVE, trading, or order authority.

## Safety conclusion

- Safe edit boundary: explicit replacement with tested compatibility.
- High-risk impact: no independent authority expansion; base evidence remains mandatory.
