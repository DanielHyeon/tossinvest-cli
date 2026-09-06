# Post-edit AST evidence

Historical intermediate snapshot generated on 2026-09-05 with
`go run ./tools/logic-map/extract_go_ast.go`. The hashes and branch numbering below
predate the final implementation and must not be used as current-source evidence.
The bundles in `function-logic/` were subsequently regenerated; their presence
does not establish preservation of the earlier pre-edit snapshots. Current
validation and provenance limitations are recorded in `coverage-status.md`.

| Function | Source SHA-256 | AST branches | Relevant post-edit branch coverage |
|---|---|---:|---|
| `runSoakAttest` | `bfdb472915dc17d511330e58b283eacbcdbd5989eb0c460a0aba0fb856a9d811` | B1–B19 | `TestSoakAttestRecordsBoundedRefusalStatus`, `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation`, `TestSoakAttestRenewalStatusRejectsPathOverrides` |
| `newSoakAttestCmd` | `bfdb472915dc17d511330e58b283eacbcdbd5989eb0c460a0aba0fb856a9d811` | none | CLI tests above exercise registration through `runCLI` |
| `Summary.Evaluate` | `a7015e77a473d23ffaa07c24da3f1cdc8bb5a73585073c0541c21c9755ecb005` | B1 range | existing evaluator tests plus typed refusal CLI test |
| `BuildAttestation` | `a7015e77a473d23ffaa07c24da3f1cdc8bb5a73585073c0541c21c9755ecb005` | B1–B6 | `TestBuildAttestationRefusesAnIncompleteSoak` and bounded refusal CLI test |
| `Console.readAttestation` | `c2b4c295663502cc0d58e71251ef1630e2c3a0a01b7d034200856492ce17f2a3` | B1–B8 | `TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability`, `TestDashboardTreatsStaleOrMismatchedRenewalStatusAsUnknown` |

`runSoakAttest` has a new deferred write path: B1 enables recording; B2 rejects
record/out overrides; B3 handles profile-path resolution; B4–B7 cover the
deferred attestation/status handling. The existing display-loop branches remain
B16–B19; they are not repurposed as status-write evidence.

`BuildAttestation` now calls `evaluateIssues` once and returns `IncompleteError`
for B1. `Summary.Evaluate` maps that same private result back to the established
message list, so command recording does not re-evaluate criteria or parse error
text.
