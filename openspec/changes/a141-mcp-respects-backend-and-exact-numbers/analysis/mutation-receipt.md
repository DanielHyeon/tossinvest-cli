# a141 focused mutation receipt

Date: 2026-10-15. Every mutant ran in its own `/tmp/a141-mutant.*` copy of the
current isolated worktree. No source, configuration, session, credential, or
operational file in either repository checkout was changed by a mutant. This
receipt establishes test sensitivity only; its original runs did not include a
verified process-level outbound-network guard and must not be used as proof of
network non-use.

| Mutant | Focused offline test | Result |
|---|---|---|
| Remove `openAPI.Enabled` from MCP official eligibility | `TestMCPOfficialEligibilityMatchesCLI` | Killed: disabled OpenAPI exposed credentials |
| Remove the effective-WTS check from eligibility | `TestMCPCmdWTSDoesNotWireOfficialDependencies` | Killed: fixture MCP `auth_status` reported official connected |
| Remove `Decoder.UseNumber` before result shrinking | `TestToolResultKeepsExactNumbersAfterTrimming` | Killed: `9007199254740993`, its negative, and `0.1234567890123456789` rounded |

The command wiring test uses only a temporary config directory, synthetic
stored credentials and a synthetic WTS session. Its JSON-RPC input reaches EOF
after `auth_status`, an official read, and an order preview; WTS remains
connected while both official operations return tool errors. No live account
or order was used; separate guarded reruns are required before claiming an
offline transport receipt.

## Final rerun

Final runs used `/tmp/a141-final-mutant.jOJOz0`, copied from source snapshot
`24b2d3f69f90`. All mutants were killed. The WTS-gate mutant used the guarded
fresh-subprocess test, which injects the local deny proxy through upper- and
lower-case proxy variables and clears `NO_PROXY`/`no_proxy`. The enabled-gate
mutant is a pure eligibility function test and makes no transport call; the
number mutant is an in-process JSON test and makes no transport call.

Focused WTS availability receipt: `TestCallOperationWTSDispatches` uses an
`httptest` fixture for `stock_ranking` and asserts its local request path and
returned `005930` payload.
