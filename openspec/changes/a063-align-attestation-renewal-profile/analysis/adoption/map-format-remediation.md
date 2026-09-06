# Function Logic Map format remediation

The final checker requires every current E-based Function Logic Map to expose these exact headings:

- `## Inputs and invariants`
- `## Branches and early returns`
- `## Calls and live bindings`
- `## State mutations and fallbacks`
- `## Safety conclusion`

This remediation reorganizes the pre-existing source-derived prose from the nine current E-based maps into that required structure. It retains their source snapshot identities, AST B-id tables, focused-test limits, returned-error behavior, local-state boundaries, and no-live-order/no-gate conclusions. The three maps whose E-base test functions are deleted state under the required headings that they are immutable historical E evidence only and make no current runtime or execution claim.

`cmd-tossctl--newsoakattestcmd/branch-test-map.md` also now names B1 as the constructor happy path, bounded strictly to the existing registration/read-only test; it does not claim flag-binding or `RunE` coverage.
