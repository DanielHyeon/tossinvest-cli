# Branch Test Map: `reportCommand`

Relevant package test surfaces: TestReportCommandReadsARealLedger, TestCompareReportRejectsMismatchesAndDuplicates. These tests cover named scenarios, not every AST branch; rows below identify unproven individual branch coverage rather than asserting it. Existing socket-backed tests could not execute in this sandbox.

| Branch | Source branch | Individual branch coverage |
|---|---|---|
| B1 | if at 237:2: `if err := flags.Parse(args); err != nil {` | Not individually established |
| B2 | if at 242:3: `if f.Name == "ledger" {` | Not individually established |
| B3 | if at 246:2: `if !explicitLedger {` | Not individually established |
| B4 | if at 250:2: `if err != nil {` | Not individually established |
| B5 | if at 253:2: `if *compare {` | Not individually established |
| B6 | if at 260:2: `if err == nil && !explicitLedger {` | Not individually established |
