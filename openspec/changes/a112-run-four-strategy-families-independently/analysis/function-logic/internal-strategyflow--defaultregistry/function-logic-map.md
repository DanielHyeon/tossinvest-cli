# Function Logic Map: `defaultRegistry`

- Source: `internal/strategyflow/adapters.go` (13-24)
- Function: `defaultRegistry` in package `strategyflow`
- Signature: `defaultRegistry(params=0, results=1)`
- File SHA-256: `0f6b4e682e89e6d24c4c3686a5a1ad5ea1f0825e904236ea892b5905029065b6`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Builds the evaluation registry: one binding per frozen descriptor. L3 grew it from six to eight by adding `evaluateBreakoutKR`/`evaluateBreakoutUS` against `pairedDescriptors[6]`/`[7]`. Branchless — the list is the whole logic.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. package suite: `go test -tags tossos_testseams -covermode=count ./internal/strategyflow/`; engine suite: `go test -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body executed **8x** under the package suite.

Exact AST return positions: 14:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | body | 13:1 | branchless: the whole body is one path |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newRegistry` | 14:9 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.
- Constructs and returns a fresh registry; mutates nothing reachable from outside.

## Safety conclusion

- The count is load-bearing: an index that does not exist would not compile, and `ValidateDescriptors` refuses any set that is not exactly the frozen eight.
