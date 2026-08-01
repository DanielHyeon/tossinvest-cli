# Function Logic Map: `computeTradeOutcome`

- Source: `internal/journal/trade_outcomes.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `positionID` | exact closed projected position | `positions.id` / `trade_outcomes.position_id` | missing or unpriceable evidence returns `(zero,false)` |
| ledger rows | one position, frozen exit state, locally claimed entry/exit fills | journal foreign keys and monotone `fill_events.id` rule | any ambiguity/missing fact returns no measurement, never a guessed value |
| cost model | configured, validated market rates | `internal/costs.Model` | invalid market/rate returns no outcome |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at line 165: `if !model.Configured() {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B2 | AST `if` at line 180: `if err != nil {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B3 | AST `if` at line 183: `if !rows.Next() {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B4 | AST `if` at line 187: `if err := rows.Scan(&account, &market, &symbol, &decisionID, &adoptionID,` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B5 | AST `if` at line 193: `if decisionID == "" && adoptionID == "" {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B6 | AST `if` at line 201: `if !ok {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B7 | AST `if` at line 206: `if adoptionID != "" {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B8 | AST `else` at line 208: `} else {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B9 | AST `if` at line 211: `if !ok \|\| buy.quantity.Sign() <= 0 {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B10 | AST `if` at line 217: `if err != nil {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B11 | AST `if` at line 221: `if err != nil {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |
| B12 | AST `if` at line 234: `if denominator.Sign() != 0 {` | read-only derivation; no write occurs until caller freezes the result | explicit success/error/continue path; no invented fallback | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestABackfillRecoversTheGapAndRefusesToRewriteIt`, negative outcome fixtures |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `frozenExitFacts` | reads immutable risk/policy stage | any failure is unmeasured | current HEAD + tests |
| `roundTripLegs` / `adoptedRoundTripLegs` | exact ledger fill attribution | no symbol/time nearest-neighbour fallback | current HEAD + provenance tests |
| `Model.EstimateTradeCost` | compute each leg from same model used for net PnL | both legs required; no retry/default | `internal/costs` tests |

## State mutations and fallbacks

- Function is pure with respect to journal state; it performs bounded reads only.
- Total cost is summed from the exact two cost values already deducted from PnL, preventing a second cost formula.

## Safety conclusion

- Safe edit boundary: return one additional immutable scalar derived from already-authoritative inputs.
- High-risk impact: yes — journal analytics feeds operator performance; fail-closed tests are mandatory.
