# Function Logic Map: `strategyflowDecisionIdentity`

Base 28a800f5; source `internal/journal/strategyflow_projection.go`; authoritative branch and call extraction: `ast.json`; syntax cross-check: `ast-grep.txt`; scan: `risk-pattern-report.md`.

## Inputs and invariants

Inputs and durable identities are read from the function signature and callers. The approved design retains legacy v2/v3, authoritative risk and exit protection, atomic reservation, fail-closed v4, and no new network lock. Refer to `analysis/discovery-plan.md` and design decisions for exact binding.

## Branches and early returns

| Branch | Source line | Existing condition / branch | Required regression |
|---|---|---|---|
| B1 | 123 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B2 | 128 | `if record.SchemaVersion == strategyflowRiskBindingSchemaVersionV2 {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B3 | 130 | `} else if record.SchemaVersion != strategyflowRiskBindingSchemaVersion {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B4 | 130 | `} else if record.SchemaVersion != strategyflowRiskBindingSchemaVersion {` | Preserve existing branch; new contract refusal and failure paths require added tests. |

Return lines: 124, 131, 133. Full expressions and effects are in source and `ast-grep.txt`; line excerpts do not replace source review.

## Calls and live bindings

AST calls: json.Marshal, fmt.Errorf, sha256.Sum256, errors.New, hex.EncodeToString. Journal transaction failures return errors with rollback; projection and digest functions do not perform transport. Existing configuration binding and writer/replay scope are documented in discovery.

## State mutations and fallbacks

Inspect source and AST assignments for current mutations. No legacy fallback is permitted for marker 4. Existing protective exits and in-flight reservations remain authoritative; see approved design for transactional ordering and restrictions.

## Safety conclusion

High-risk impact: yes. Pre-edit inventory only; branch test outcomes are pending and cannot be inferred from these maps.

Source/function binding: `internal/journal/strategyflow_projection.go` / `strategyflowDecisionIdentity`.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
