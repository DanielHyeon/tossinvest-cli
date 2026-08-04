# Function Logic Map: `TestOpenAPIStrategyRuntimeIsStrictPairedReadOnlyProjection`

- Source: `internal/httpapi/model_openapi_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The OpenAPI document must describe exact KR and US keys, strict schemas and GET/SSE-only routes.

## Branches and early returns

B1–B14 cover file/decode failures, schema lookup assertions, required fields, market const identity, enum bounds and mutation-route absence.

## Calls and live bindings

Reads only the repository OpenAPI artifact.

## State mutations and fallbacks

No mutations; every missing or swapped contract fails the test.

## Safety conclusion

Generated clients cannot accept cross-market identities or invent mutation endpoints.
