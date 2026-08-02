# Function Logic Map: `httpAPIReader.readPositions`

- Source: `cmd/tossctl/httpapi_reader.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `httpAPIReader.readPositions(params=1, results=2)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 92 | `if r == nil \|\| r.holdings == nil \|\| r.accountRef == nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B2 | `if` at line 96 | `if err != nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B3 | `if` at line 100 | `if err != nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B4 | `if` at line 107 | `if journalReadable {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B5 | `if` at line 109 | `if err != nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B6 | `if` at line 113 | `if policyErr != nil {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |
| B7 | `range` at line 116 | `for _, state := range policyStates {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B8 | `range` at line 121 | `for _, row := range journalRows {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B9 | `range` at line 129 | `for _, broker := range brokerRows {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B10 | `if` at line 133 | `if stored, ok := byKey[key]; ok {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B11 | `else` at line 141 | `} else {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B12 | `if` at line 137 | `if released {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B13 | `if` at line 143 | `if !journalReadable {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B14 | `range` at line 151 | `for _, stored := range journalRows {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B15 | `if` at line 152 | `if _, ok := seen[positionKey(stored.Position.Market, stored.Position.Symbol)]; ok {` | local/projection assignment | continues through contract | go test ./cmd/tossctl |
| B16 | `if` at line 162 | `if released {` | local/projection assignment | early return/error nearby | go test ./cmd/tossctl |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.New` | explicit dependency at line 93 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.accountRef` | explicit dependency at line 95 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.holdings.Holdings` | explicit dependency at line 99 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.journal.LivePositionExits` | explicit dependency at line 108 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.journal.PositionPolicies` | explicit dependency at line 112 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `strings.TrimSpace` | explicit dependency at line 117 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `make` | explicit dependency at line 120 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `len` | explicit dependency at line 120 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionKey` | explicit dependency at line 122 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.readManagementRuntime` | explicit dependency at line 124 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `r.clockNow` | explicit dependency at line 126 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionMarket` | explicit dependency at line 130 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `positionFromBroker` | explicit dependency at line 132 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `attest.Mask` | explicit dependency at line 132 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `applyStoredPosition` | explicit dependency at line 134 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `lifecycleFlags` | explicit dependency at line 136 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `applyReleasedExitTruth` | explicit dependency at line 138 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `applyManagementProjection` | explicit dependency at line 140 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `httpapi.ExitLineFrom` | explicit dependency at line 146 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `operatorview.BuildExitLine` | explicit dependency at line 146 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `append` | explicit dependency at line 149 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `strings.ToUpper` | explicit dependency at line 156 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `stored.Position.ExitEligible` | explicit dependency at line 158 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |
| `pointerTo` | explicit dependency at line 168 | result/error follows the AST-recorded branch; no hidden retry is assumed | current AST + focused tests |

## State mutations and fallbacks

- AST records 29 assignment(s), 6 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
