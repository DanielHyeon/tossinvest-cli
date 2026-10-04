# Function Logic Map: `SelectedScopeJSONFields`

- Source: `internal/strategyprojection/lanes.go`
- Source SHA-256: `fb4b35f5447efe634788c0e4ba572808ddc5a703777739b3a9e143c9e7c22bd7`
- Signature: `SelectedScopeJSONFields(params=0, results=1)`
- Source range: `193:1`–`195:2` (base)
- Pinned revision: `base` — this function's body is **unchanged**. The frozen base `1e25b3a31b7109cf7688768000914ce04360a5a9` is pinned because the diff only touches its neighbourhood: LaneRuntimes · LaneShadowOutcomes were inserted right after it, and a pure insertion after a function's last line counts as intersecting that function.
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 목록 불변 — `TestLaneAndCoordinatorJSONNamesAreTheContract`.

## Branches and early returns

- Exact AST return nodes: `194:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 읽기 전용 목록.
