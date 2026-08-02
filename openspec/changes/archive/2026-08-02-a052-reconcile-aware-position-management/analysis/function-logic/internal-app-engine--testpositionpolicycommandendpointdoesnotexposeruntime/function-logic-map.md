# Function Logic Map: `TestPositionPolicyCommandEndpointDoesNotExposeRuntime`

- Source: `internal/app/engine/position_policy_transport_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function contract | `TestPositionPolicyCommandEndpointDoesNotExposeRuntime(params=1, results=0)` | current Go signature and callers | errors/unknown values propagate without inventing effective state |
| Runtime or persisted state | values supplied by the owning engine/journal/read boundary | current HEAD plus approved a052 spec | unavailable or ambiguous facts remain unknown/deferred |
| Safety boundary | read-only projection, or pre-journal fail-closed validation | a052 design and TossOS safety invariants | no live order, reconciliation resolution, or operating-toggle mutation |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 163 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyCommandEndpointDoesNotExposeRuntime |
| B2 | `if` at line 168 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyCommandEndpointDoesNotExposeRuntime |
| B3 | `if` at line 172 | `if err := json.Unmarshal(body, &descriptor); err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyCommandEndpointDoesNotExposeRuntime |
| B4 | `if` at line 176 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyCommandEndpointDoesNotExposeRuntime |
| B5 | `if` at line 181 | `if err != nil {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyCommandEndpointDoesNotExposeRuntime |
| B6 | `if` at line 185 | `if response.StatusCode != http.StatusNotFound \|\| commands.runtimeCalls != 0 {` | local/read-model state only; see AST assignments | continues through the function contract | TestPositionPolicyCommandEndpointDoesNotExposeRuntime |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `privateEngineTestDir` | execute the explicit dependency at line 160 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `StartPositionPolicyCommandServer` | execute the explicit dependency at line 162 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatal` | execute the explicit dependency at line 164 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `server.Close` | execute the explicit dependency at line 166 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `os.ReadFile` | execute the explicit dependency at line 167 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `positionpolicyrpc.DescriptorPath` | execute the explicit dependency at line 167 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `json.Unmarshal` | execute the explicit dependency at line 172 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `http.NewRequest` | execute the explicit dependency at line 175 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `req.Header.Set` | execute the explicit dependency at line 179 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `http.DefaultClient.Do` | execute the explicit dependency at line 180 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `response.Body.Close` | execute the explicit dependency at line 184 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `t.Fatalf` | execute the explicit dependency at line 186 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- AST records 7 assignment(s), 0 return statement(s), and 0 goroutine launch(es).
- Fallback is fail-closed: missing, mismatched, unavailable, or ambiguous operational truth is not promoted to managed/effective/actionable state.
- The a052 path adds no reconciliation-resolution command and does not authorize a live trade.

## Safety conclusion

- Safe edit boundary: Keep the change inside transport/read-model/rendering behavior; effective runtime facts must never be inferred from desired configuration.
- High-risk impact: no direct order side effect; nevertheless unknown data and account identity remain fail-closed/masked.
