# Function Logic Map: `verifyStrategyRiskBinding`

Source `internal/journal/strategy_lineage.go`; base 28a800f5. AST `ast.json`; syntax `ast-grep.txt`; risk scan `risk-pattern-report.md`.

## Inputs and invariants

Exact authority and legacy compatibility from approved design. Preserve protection and reservations.

## Branches and early returns

| Branch | Line | Condition | Required regression |
|---|---|---|---|
| B1 | 748 | `if journalHash := HashPreimage(decision.RiskPreimage); journalHash != decision.RiskHash {` | Preserve existing branch; test new contract separately. |
| B2 | 752 | `if err != nil {` | Preserve existing branch; test new contract separately. |
| B3 | 756 | `if !ok {` | Preserve existing branch; test new contract separately. |
| B4 | 760 | `if err != nil {` | Preserve existing branch; test new contract separately. |
| B5 | 763 | `if versioned {` | Preserve existing branch; test new contract separately. |
| B6 | 764 | `if schema != strategyflowRiskBindingSchemaVersionV2 && schema != strategyflowRiskBindingSchemaVersion {` | Preserve existing branch; test new contract separately. |
| B7 | 767 | `if riskRecord.AccountRef != decision.AccountRef {` | Preserve existing branch; test new contract separately. |

## Calls and live bindings

AST calls: HashPreimage, errors.New, ParsePreimage, fmt.Errorf, strategyDecisionPayloadSchema, verifyStrategyflowRiskBinding, verifyLegacyStrategyRiskBinding. Source contains full expressions and error propagation; config binding is documented in discovery.

## State mutations and fallbacks

Journal transaction commits or rolls back atomically; do not infer missing v4 as legacy. Preserve existing protection.

## Safety conclusion

High-risk yes. Test mappings below are planned, not claims of executed coverage.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
