# Function Logic Map: `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation`

- Source: `cmd/tossctl/soak_test.go` (471-496); SHA-256 `5821be913c3823ef7d8d5ef08ceec63fd190530ddb20b595406a344dc3e8b059` (post-edit). Pre-edit AST: `../../scope-absorb/pre-edit-ast/cmd-tossctl--testsoakattestrecordsissuedstatusbesidetheresolvedprofileattestation.json` (471-493, same 5 branches).
- AST evidence: `ast.json`. Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| isolated profile with a qualifying soak record **and (D6 edit) a supervised KR verification record** | synthetic fixture | `seedQualifyingRecord`, `seedSupervisedRecord` | `t.Fatalf`/`t.Errorf` |

The pre-edit fixture had no supervised record, so the issued file lacked the live-only endpoints and the assertion `len(status.ReasonCodes) == 0` pinned the exact gap D6 closes. The edit seeds the supervised record so this test keeps meaning "fully covered issuance records no reasons"; the uncovered case is `TestA063RenewalWithoutSupervisedEvidenceRecordsTheUncoveredScope`.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `soak attest --record-renewal-status` errors | none | Fatal | self |
| B2 | `LoadRenewalStatus` errors | none | Fatal | self |
| B3 | `attest.Load` errors | none | Fatal | self |
| B4 | status not issued / expiry mismatch / any reason code | none | Errorf | self |
| B5 | a status file exists beside another attestation path | none | Error | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `testenv.Isolate`, `seedQualifyingRecord`, `seedSupervisedRecord`, `runCLI`, `soak.LoadRenewalStatus`, `attest.Load` | build isolated fixture and observe outputs | synthetic only, no network | CodeGraph + AST |

## State mutations and fallbacks

- Writes only under the isolated temp profile.

## Safety conclusion

- Safe edit boundary: one fixture line added; assertions unchanged.
- High-risk impact: no.
