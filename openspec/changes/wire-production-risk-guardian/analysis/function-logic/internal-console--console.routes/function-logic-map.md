# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Console` | initialized session/CSRF and handlers | `New` | server cannot route otherwise |
| update install | POST-only mutating handler | system-update spec | session/CSRF refusal |
| engine/verify starts | authenticated mutating routes | existing route table | serialized with update commit |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path (AST is branchless) | register exact route table | return mux | static route tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `session0`, `mutating` | enforce authentication/method/CSRF | refusal before handler | CodeGraph + static tests |
| `startExclusive` | serialize two start routes with installation | blocks/refuses after commit | concurrency test |
| register screen helpers | append read-only routes | unchanged | CodeGraph + AST |

## State mutations and fallbacks

- Update installation receives both session and CSRF wrappers.
- Engine stop stays available; only work-start routes share update exclusion.

## Safety conclusion

- Safe edit boundary: add one mutating route and wrap the two start routes without changing handler order.
- High-risk impact: yes — route wrapper omission could expose self-replacement or race active work.
