# Function Logic Map: `LaneJSONFields`

- Source: `internal/strategyprojection/lanes.go`
- Source SHA-256: `9b6b15b0ebe8e74fcd582a8a6d6abed626706542ba9706e5ba551a3c302315a4`
- Signature: `LaneJSONFields(params=0, results=1)`
- Source range: `197:1`–`201:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 이름 목록은 실제 직렬화(`TestLaneAndCoordinatorJSONNamesAreTheContract`)와 OpenAPI(`TestOpenAPIDocumentsTheLaneAndCoordinatorChildrenByTheirExactNames`)가 잰다.

## Branches and early returns

- Exact AST return nodes: `198:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 읽기 전용 목록.
