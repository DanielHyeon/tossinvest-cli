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

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.6.1 post-edit) |
|---|---|---|---|---|
| B1 | range at 473:2 | `for _, row := range rows {`; then `rowFilled, filledOK := new(big.Int).SetString(row.FilledMinor, 10)` (line last changed by `8022f578`) | not a066 | covered |
| B2 | if at 476:3 | `if !filledOK \|\| !heldOK \|\| rowFilled.Sign() < 0 \|\| rowHeld.Sign() < 0 \|\| rowFilled.BitLen() > 256 \|\| rowHeld.BitLen() > 256 \|\|`; then `return JournalBucketUsage{}, errors.New("risk bucket: invalid or latched journal usage")` (line last changed by `8022f578`) | not a066 | covered |
| B3 | if at 486:3 | `if filled.BitLen() > 256 \|\| held.BitLen() > 256 {`; then `return JournalBucketUsage{}, errors.New("risk bucket: journal usage overflow")` (line last changed by `8022f578`) | not a066 | NOT covered |

5.6.1 post-edit (HEAD `b8211926`): 3 → 3; returns `JournalBucketUsage`; latched rows set `Latched` instead of failing (the production caller refuses them with the same text). Pre-edit table: `analysis/pre-edit/5.6.1/internal-riskbucket--aggregateproductionriskusage.md`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| big.Int arithmetic | exact held/filled sums | overflow → error | AST |

## State mutations and fallbacks

- Pure.

## Safety conclusion

- Safe edit boundary (5.6.1 F1): return a `JournalBucketUsage` and report latched rows as `Latched` instead of an error; the sole production caller turns `Latched` into the same error text, so production behaviour is unchanged. Validation of every other row property stays an error.
- High-risk impact: yes — q_final sizing/admission authority.
