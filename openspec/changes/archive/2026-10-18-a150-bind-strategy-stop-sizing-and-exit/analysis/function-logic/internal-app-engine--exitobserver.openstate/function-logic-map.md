# Function Logic Map: `openState`

Base 28a800f5; source `internal/app/engine/exitloop.go`; authoritative branch and call extraction: `ast.json`; syntax cross-check: `ast-grep.txt`; scan: `risk-pattern-report.md`.

## Inputs and invariants

Inputs and durable identities are read from the function signature and callers. The approved design retains legacy v2/v3, authoritative risk and exit protection, atomic reservation, fail-closed v4, and no new network lock. Refer to `analysis/discovery-plan.md` and design decisions for exact binding.

## Branches and early returns

| Branch | Source line | Existing condition / branch | Required regression |
|---|---|---|---|
| B1 | 711 | `if p.Adopted() {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B2 | 715 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B3 | 720 | `if err != nil {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B4 | 725 | `if !ok {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B5 | 732 | `if o.opts.CommonPolicy != "" {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B6 | 742 | `switch {` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B7 | 743 | `case errors.Is(err, journal.ErrExitStateExists):` | Preserve existing branch; new contract refusal and failure paths require added tests. |
| B8 | 746 | `case err != nil:` | Preserve existing branch; new contract refusal and failure paths require added tests. |

Return lines: 712, 716, 721, 726, 745, 747, 755. Full expressions and effects are in source and `ast-grep.txt`; line excerpts do not replace source review.

## Calls and live bindings

AST calls: p.Adopted, o.openAdoptedState, o.opts.Journal.LookupDecision, fmt.Errorf, journal.ParsePreimage, o.opts.Journal.OpenExitState, errors.Is, o.log. Journal transaction failures return errors with rollback; projection and digest functions do not perform transport. Existing configuration binding and writer/replay scope are documented in discovery.

## State mutations and fallbacks

Inspect source and AST assignments for current mutations. No legacy fallback is permitted for marker 4. Existing protective exits and in-flight reservations remain authoritative; see approved design for transactional ordering and restrictions.

## Safety conclusion

High-risk impact: yes. Pre-edit inventory only; branch test outcomes are pending and cannot be inferred from these maps.

Source/function binding: `internal/app/engine/exitloop.go` / `ExitObserver.openState`.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
