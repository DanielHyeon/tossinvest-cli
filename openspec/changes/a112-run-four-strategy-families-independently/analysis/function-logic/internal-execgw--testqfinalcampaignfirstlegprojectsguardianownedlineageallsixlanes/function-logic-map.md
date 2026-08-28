# Function Logic Map: `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllSixLanes`

- Source: `internal/execgw/riskguardian_account_base_testseam_test.go` (98-180)
- Function: `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllSixLanes` in package `execgw_test`
- Signature: `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllSixLanes(params=1, results=0)`
- File SHA-256: `57f4899aaf068d61c4456252f9a5e2d2372d0c515b2c395914e6609e86f32739`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 14.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The frozen-base record of the Guardian-owned lineage projection matrix. Its `len(descriptors) != 6` guard is the same 6→8 breakage as the engine and journal cases (decision 50); renamed here to `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes`.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 100:2 | no coverage block maps to this position |
| B2 | range | 104:2 | no coverage block maps to this position |
| B3 | if | 120:4 | no coverage block maps to this position |
| B4 | if | 127:4 | no coverage block maps to this position |
| B5 | if | 140:4 | no coverage block maps to this position |
| B6 | if | 143:4 | no coverage block maps to this position |
| B7 | if | 147:4 | no coverage block maps to this position |
| B8 | if | 150:4 | no coverage block maps to this position |
| B9 | if | 155:4 | no coverage block maps to this position |
| B10 | if | 159:4 | no coverage block maps to this position |
| B11 | if | 163:4 | no coverage block maps to this position |
| B12 | if | 167:4 | no coverage block maps to this position |
| B13 | if | 171:4 | no coverage block maps to this position |
| B14 | if | 177:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `strategyflow.Descriptors` | 99:17 |
| `len` | 100:5 |
| `t.Fatalf` | 101:3 |
| `len` | 101:53 |
| `t.Run` | 106:3 |
| `string` | 107:14 |
| `newGuardian` | 110:11 |
| `fixedIDs` | 111:21 |
| `qFinalAccountBaseRequest` | 113:13 |
| `risk.TestOnlySealAccountBaseFX` | 119:15 |
| `guardianPolicy` | 119:68 |
| `t.Fatal` | 121:5 |
| `bindPairedAccountBaseFXForTest` | 123:4 |
| `pairedStrategyflowMinorPrices` | 124:42 |
| `strategyflow.AcceptedResultForJournalTest` | 125:19 |
| `t.Fatal` | 128:5 |
| `strings.Repeat` | 130:28 |
| `reserveWeeklyFirstLegForTest` | 138:21 |
| `rig.guardian.PrecheckQFinalCampaignFirstLeg` | 139:21 |
| `t.Fatalf` | 141:5 |
| `errors.Unwrap` | 141:71 |
| `precheck.QCandidate` | 143:7 |
| `precheck.QFinal` | 143:38 |
| `t.Fatalf` | 144:5 |
| `precheck.QCandidate` | 144:85 |
| `precheck.QFinal` | 144:108 |
| `rig.guardian.IssuePrecheckedQFinalCampaignFirstLeg` | 146:20 |
| `context.Background` | 146:71 |
| `t.Fatalf` | 148:5 |
| `t.Fatalf` | 152:5 |
| `rig.journal.LookupDecision` | 154:21 |
| `context.Background` | 154:48 |
| `t.Fatal` | 156:5 |
| `journal.QFinalPolicyVersion` | 158:23 |
| `rig.guardian.PolicyVersion` | 158:51 |
| `t.Fatal` | 160:5 |
| `journal.ParsePreimage` | 162:21 |
| `t.Fatal` | 164:5 |
| `Identity` | 167:65 |
| `result.ExecutionTerms.Policy` | 167:65 |

(6 further call sites omitted; `ast.json` carries all 46.)

## State mutations and fallbacks

- AST assignments: 28. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
