# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `runEngineRun(params=2, results=1)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 180 | `if ctx == nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B2 | `if` at line 186 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B3 | `if` at line 192 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B4 | `if` at line 202 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B5 | `if` at line 203 | `if clauses := engine.UnmetInterlockClauses(err); clauses != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B6 | `range` at line 205 | `for _, clause := range clauses {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B7 | `if` at line 214 | `if !ectx.Automation.Verified {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B8 | `if` at line 224 | `if lockPath, verr := engineVerifyLockPath(root); verr == nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B9 | `if` at line 225 | `if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.StaleAfter); fresh {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B10 | `if` at line 236 | `if merr != nil {` | local/read-model state only; see AST assignments | continues through the function contract | go test ./cmd/tossctl |
| B11 | `else` at line 240 | `} else {` | local/read-model state only; see AST assignments | continues through the function contract | go test ./cmd/tossctl |
| B12 | `if` at line 247 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B13 | `if` at line 251 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B14 | `if` at line 255 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |
| B15 | `if` at line 260 | `if err != nil {` | local/read-model state only; see AST assignments | early return/error path nearby | go test ./cmd/tossctl |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.Context` | execute the explicit dependency at line 179 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `context.Background` | execute the explicit dependency at line 181 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `cmd.OutOrStdout` | execute the explicit dependency at line 183 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `cmd.ErrOrStderr` | execute the explicit dependency at line 183 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `engineJournalDir` | execute the explicit dependency at line 185 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `enginelock.Acquire` | execute the explicit dependency at line 191 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `lock.Release` | execute the explicit dependency at line 195 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `fmt.Fprintf` | execute the explicit dependency at line 196 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `lock.Path` | execute the explicit dependency at line 196 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `clock.System` | execute the explicit dependency at line 199 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `obs.NewLogger` | execute the explicit dependency at line 200 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `engineAssemble` | execute the explicit dependency at line 201 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `engine.UnmetInterlockClauses` | execute the explicit dependency at line 203 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `call` | execute the explicit dependency at line 212 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `ectx.Close` | execute the explicit dependency at line 212 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `engineVerifyLockPath` | execute the explicit dependency at line 224 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `runlock.Fresh` | execute the explicit dependency at line 225 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `clk.Now` | execute the explicit dependency at line 225 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `Format` | execute the explicit dependency at line 227 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `at.UTC` | execute the explicit dependency at line 227 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `enginelock.MarkerPath` | execute the explicit dependency at line 233 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `enginelock.Hold` | execute the explicit dependency at line 234 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `releaseMarker` | execute the explicit dependency at line 235 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `engineRuntimeFactory` | execute the explicit dependency at line 246 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `engine.NewPositionPolicyCommandService` | execute the explicit dependency at line 250 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `engine.StartPositionPolicyCommandServer` | execute the explicit dependency at line 254 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `policyControl.Close` | execute the explicit dependency at line 258 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `engine.StartPositionPolicyRuntimeServer` | execute the explicit dependency at line 259 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `policyRuntime.Close` | execute the explicit dependency at line 263 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `ectx.Automation.MaskedAccount` | execute the explicit dependency at line 265 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.Join` | execute the explicit dependency at line 265 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `rt.LoopNames` | execute the explicit dependency at line 265 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `context.WithCancel` | execute the explicit dependency at line 268 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `cancel` | execute the explicit dependency at line 269 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `watchStopSignals` | execute the explicit dependency at line 270 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `stopWatching` | execute the explicit dependency at line 271 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `rt.Run` | execute the explicit dependency at line 273 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 20 assignment(s), 11 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
