# Function Logic Map: `aggregateProductionRiskUsage`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- AST evidence: `ast.json` (5.6.1 pre-edit, extracted 2026-09-27 at HEAD `f2decd0a` before any 5.6.1 edit)
- Risk scan: `risk-pattern-report.md`
- Coverage: pre-edit source copy (`analysis/harness/test_in_copy.sh`), `go test -coverpkg ./internal/journal,./internal/riskbucket` over the journal package (untagged) and riskbucket (untagged + `tossos_testseams`); rows rendered by `analysis/harness/preedit_rows.py` from `git show f2decd0a:<file>`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rows | reservation rows of one bucket | readProductionRiskUsage | invalid/latched row → error |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.7 post-edit) |
|---|---|---|---|---|
| B1 | range at 479:2 | `for _, row := range rows {`; then `rowFilled, filledOK := new(big.Int).SetString(row.FilledMinor, 10)` (line last changed by `8022f578`) | not a066 | covered |
| B2 | if at 482:3 | `if !filledOK \|\| !heldOK \|\| rowFilled.Sign() < 0 \|\| rowHeld.Sign() < 0 \|\| rowFilled.BitLen() > 256 \|\| rowHeld.BitLen() > 256 \|\|`; then `return JournalBucketUsage{}, ErrJournalUsageInvalid` (line last changed by `8022f578`) | not a066 | covered |
| B3 | if at 494:3 | `if filled.BitLen() > 256 \|\| held.BitLen() > 256 {`; then `return JournalBucketUsage{}, fmt.Errorf("%w: journal usage overflow", ErrJournalUsageInvalid)` (line last changed by `8022f578`) | not a066 | NOT covered |

5.7 post-edit (HEAD `54e67495` + 5.7 working tree): 3 → 3; the invalid-row and overflow returns now carry the typed `ErrJournalUsageInvalid` (same message text).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| big.Int arithmetic | exact held/filled sums | overflow → error | AST |

## State mutations and fallbacks

- Pure.

## Safety conclusion

- Safe edit boundary (5.6.1 F1): return a `JournalBucketUsage` and report latched rows as `Latched` instead of an error; the sole production caller turns `Latched` into the same error text, so production behaviour is unchanged. Validation of every other row property stays an error.
- High-risk impact: yes — q_final sizing/admission authority.

## 6.5 fix lot (2026-09-28)

6.5 fix lot: `OverageLatched`/`UnknownLatched` are now set alongside `Latched` (assignments only, shape 3→3). Mutation V07 was CAUGHT.

Positions re-read from the post-edit `ast.json` (2026-09-29): B1 477→479, B2 480→482, B3 490→494 — the new struct field and the two loop lines of `28629ec6` moved them; conditions unchanged. Coverage cells are the historical measurements at the commits they name.
