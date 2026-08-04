# Function Logic Map: `testManifestClaim.verifyActivation`

- Source: `internal/scheduler/desired_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The test verifier records invocation and returns controlled evidence or error.

## Branches and early returns

B1–B4 cover call counting, injected error, default generation normalization and configured expiry/default expiry.

## Calls and live bindings

No production external call exists.

## State mutations and fallbacks

Only test fixture counters change.

## Safety conclusion

Test-only verifier exercises the production fail-closed evidence contract.
