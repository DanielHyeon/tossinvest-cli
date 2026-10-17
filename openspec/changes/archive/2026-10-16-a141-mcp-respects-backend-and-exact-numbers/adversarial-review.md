# a141 independent implementation adversarial review

Reviewer: `/root/a141_adversarial_reviewer` (independent Codex context; not Terra implementer). Date: 2026-10-15. Baseline: `6765858af1beb472268cb699efa1964addb84d60`. Review target: working-tree implementation over baseline, initially on planning commit `ba7901cc`. This is the dedicated implementation adversarial review, **not** the subsequent additional gstack review.

## Interim verdict: PENDING

Terra is still changing tests, final maps and receipts. No final approval or successful test claim is made. Reviewer read the production diff, new tests, existing CLI eligibility tests, MCP dependency/call path and pre-edit AST evidence. Reviewer ran no tests and changed no production code, configuration, credentials, runtime or live state. Final verdict requires final source hashes and Terra receipts.

## Findings to close

1. **P1 — actual MCP wiring regression not exercised** (Manager raised; corroborated). Initial `cmd/tossctl/mcp_test.go:TestMCPOfficialEligibilityMatchesCLI` only calls the predicate helper. Replacing constructor use with `if creds != nil` would leave this test green. Need actual offline command/server wiring coverage, negative official direct-read and trading reachability, positive WTS reachability, and a constructor-bypass mutation that the test catches. Disabled config with explicit official/openapi override must be included. No live requests; fake transports/cache or existing seams only.
2. **P2 — float64 upper boundary converted before validation** (Manager raised; corroborated). Initial `omittedCount` returns `int64(n), ... n <= float64(maxInt)`. On 64-bit, `float64(maxInt)` rounds to 2^63 and permits an out-of-range conversion. Need reject nonfinite/fractional/out-of-range values before conversion, and a boundary regression. A negative converted result must never enter cumulative count.
3. **P2 — exact-number test does not assert JSON number type** (reviewer raised). Initial `TestToolResultKeepsExactNumbersAfterTrimming` uses `strings.Contains` then ordinary `json.Unmarshal`. Numeric strings would pass the substring test, contrary to the frozen spec. Decode with UseNumber, assert retained `id`, `negative` and `decimal` values are json.Number and equal exact original tokens, including nested retention. Test must still assert truncation/valid JSON.
4. **P2 — repeated-trim count conservation unverified** (reviewer raised). Initial `TestTrimAcceptsExactAndMalformedOmissionMarkers` tests one pass using a json.Number marker. It does not pass generated int marker through a second trim. Require repeated trimming and `remaining real items + cumulative omitted == original count`; retain pre-existing decoded marker coverage.

All findings were sent to Manager and Terra. Initial production diff otherwise uses the approved eligibility guard before official-client/service construction, keeps the existing lineage/service setup, and replaces float decoding only in oversized results. No additional demonstrated runtime defect beyond the numeric boundary was found in this interim inspection. That statement is scoped to reviewed interim files; pending test/maps work is not assumed complete.

## Final verification checklist (not execution claims)

- Re-read final diff vs immutable baseline; ensure only approved code surfaces changed.
- Verify F1–F4 are closed by actual code/test evidence; inspect constructor bypass mutation result.
- Check final AST source hashes/logic maps align with edited functions and branch tests.
- Read targeted test receipts and scope; full gate/additional gstack remain separate responsibilities.
- Record final reviewed source hashes and verdict here; Manager independently decides completion.

## Re-review 2 — remaining blockers

Read the updated tests and mutation receipt after Terra's focused GREEN report. F3 is closed: retained fields are decoded with UseNumber and asserted as json.Number with exact token equality. F4 is closed for the demonstrated repeated-trim regression: a generated int marker is passed into a second trim and cumulative count 3 is asserted from an original four-item list; the json.Number prior-marker case remains.

F1 remains open. The actual command test covers effective WTS only; disabled config with an explicit official override is tested only through a helper. It checks generic isError rather than the official dependency-denial message, and positive WTS operation behavior is not exercised. More seriously, the command test has no explicit offline transport boundary. `newMCPCmd` calls updatecheck at startup, and a constructor-bypass mutant can send the supplied `accounts` request through an actual official client before response assertions run. Synthetic credentials do not prevent network access. `official.New` builds its own HTTP transport, so replacing only http.DefaultTransport is not sufficient protection for that path.

F2 remains open as written: changing `<= float64(maxInt)` to `<` rejects the bool validity at the 64-bit boundary, but the tuple still evaluates `int64(n)` before validation. No explicit float64 boundary/nonfinite regression is present in the inspected test file. Require validation before conversion and a focused regression receipt.

The mutation receipt is a summary of three killed mutants, not a final all-findings-closed record. In particular, the first mutation targets the helper, the second auth status, and the third exact-number decoding; none proves the missing actual disabled-override path or network isolation. Verdict remains **CHANGES REQUESTED / final approval pending**. No tests were run by this reviewer.

## Re-review 3 — PASS CODE, evidence/gates pending

Reviewed committed source/tests through `399ca8cd` (following `602751f2` and `ba755da3`) against immutable base `6765858af1beb472268cb699efa1964addb84d60`. **All four source/test findings are closed. PASS CODE for the next independent gstack review; this is not change completion or gate approval.**

- F1: actual `newMCPCmd` execution now covers effective WTS and disabled+explicit-openapi, plus enabled auto/openapi. WTS case checks auth snapshot and precise official-credential denial for `accounts` and `place_order`; it no longer accepts arbitrary errors. The parent test launches a fresh subprocess with HTTP_PROXY/HTTPS_PROXY/ALL_PROXY (upper/lowercase) pointed to a local deny server and NO_PROXY cleared. Both official's constructor-owned transport and updatecheck use environment proxies; fresh subprocess avoids previously cached proxy settings. CONNECT is rejected by the local server without forwarding, so current public-URL requests cannot reach the remote service through this test. The matrix child executes only under that guarded parent. Existing `internal/mcp/wts_operations_test.go:TestCallOperationWTSDispatches` sends stock_ranking through a fixture HTTP server and checks a returned symbol; combined with actual-command WTS auth construction this is sufficient positive WTS coverage for the unchanged dispatcher, subject to its executed receipt.
- F2: `omittedCount` now rejects negative/fractional/out-of-range/nonfinite float64 values before the int64 conversion. Test checks NaN, positive infinity and float64(maxInt); json.Number fractional/negative/overflow cases remain. No pre-validation conversion remains on that branch.
- F3: exact retained JSON fields are parsed with UseNumber and checked as json.Number, with exact positive/negative >2^53 integer and long decimal tokens.
- F4: generated int omission marker is processed in a second trim and expected cumulative count is checked, alongside pre-existing json.Number marker handling.

The reviewed Go diff is limited to `cmd/tossctl/mcp.go`, its new test file, `internal/mcp/server.go` and its existing test file. No journal, risk-control, shared-token, engine, operating-settings or deployment code changed. Existing service/lineage construction stays inside the newly restricted eligibility branch. No further actionable source defect identified in this scope.

Final reviewed SHA-256:

```text
cmd/tossctl/mcp.go          5ce62e06dcc392d1701a197bca0ece3119a0cbf2da057f0d891753582beaf3d0
cmd/tossctl/mcp_test.go     4599da0635ccf209cb4a6ea7b2199481034dcbb117e094aef3943a4f1dd73998
internal/mcp/server.go      a621b9b36b040e9afdf193be32281b9950417bff998155d9c2a23dc31a1f876f
internal/mcp/server_test.go 49e2a2b0a6aa60b654b7a1d1c85012509aa3d34a9a03113575715a8a24bb94ba
```

Current AST source hashes for newMCPCmd/mcpOfficialEligible and toolResult/trim/omittedCount match the production files above. Map-checker completion is still pending per Manager; matching hashes alone do not prove map/branch coverage validation.

Execution-evidence limits: existing `analysis/mutation-receipt.md` explicitly retracts offline proof for original mutant runs and requires guarded reruns. Reviewer has not received final guarded mutant logs or complete focused/full-suite receipts in this file set, and has run no tests. Terra must supply guarded constructor-bypass/disabled-policy/precision sensitivity evidence, the focused tests (including WTS dispatch), final map-check success and repository gates. `TestMCPCmdDisabledAndEnabledBackendWiring` returns outside its guarded child, so its name alone is not evidence: execution receipts must run the parent `TestMCPCmdWTSDoesNotWireOfficialDependencies` (which launches both tests) or the complete package suite. Additional gstack review, Manager verification and archive/PM synchronization remain mandatory.
