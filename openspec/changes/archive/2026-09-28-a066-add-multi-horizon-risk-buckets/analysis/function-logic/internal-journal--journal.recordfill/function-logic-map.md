# Function Logic Map: `Journal.RecordFill`

- Source: `internal/journal/fills.go`
- AST evidence: `ast.json` (extracted Wave 2A 2026-09-25 at HEAD `648df8ef`; 313–558, 38 branches, 26 returns;
  byte-identical to the archived a072 bundle `imported--internal-journal-fills--journal.recordfill/ast.json`)
- Risk scan: `risk-pattern-report.md`

## Why this bundle is new in Wave 2A

a066 changed this existing function in `c60fee07` (Wave 1C, sidecar call) and `4a364caf` (Wave 1D, ambiguous
ownership latch). The Wave 1C review recorded it as "the only existing fill-path body edit", yet no bundle was
written for it on 2026-08-04. The Wave 2A attribution (functions that existed at base `23794f86` and whose body
an a066 commit changed) found the gap; this bundle closes it at the HEAD revision, which also contains
`8022f578`'s B33/B34 and earlier lines from other changes.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `obs.OrderID` | non-blank; stored verbatim | broker observation | `ErrInvalidRequest`, nothing written |
| canonical scope | account, market, trading day, symbol, side complete together, or fully legacy-unscoped | `fillSnapshotScopeOf(obs)` | `ErrInvalidRequest`, nothing written |
| filled quantity | finite decimal | broker observation | `ErrInvalidRequest`, nothing written |
| ownership | unique confirmed order-creating attempt strictly earlier than the evidence | journal attempts (`confirmedFillOwners`, `resolveFillOrigin`) | broker-only order stays an observation; ambiguity latches risk entry (a066) |
| transaction | one `BEGIN IMMEDIATE` for snapshot, events, reservation release, a066 sidecar and apply hooks | journal | any storage error rolls every write back |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage at `648df8ef` |
|---|---|---|---|---|
| B1 | if at 314:2 | `if strings.TrimSpace(obs.OrderID) == "" {`; then `return FillResult{}, fmt.Errorf("%w: a fill snapshot needs an order id", ErrInvalidRequest)` (line last changed by `15b25b40`) | not a066 | covered |
| B2 | if at 318:2 | `if !scope.complete() && !scope.legacyUnscoped() {`; then `return FillResult{}, fmt.Errorf(` (line last changed by `e3f8b391`) | not a066 | covered |
| B3 | if at 324:2 | `if err != nil {`; then `return FillResult{}, fmt.Errorf("%w: filled quantity %q is not a decimal",` (line last changed by `15b25b40`) | not a066 | covered |
| B4 | if at 328:2 | `if math.IsNaN(filled) \|\| math.IsInf(filled, 0) {`; then `return FillResult{}, fmt.Errorf("%w: filled quantity %q is not finite",` (line last changed by `15b25b40`) | not a066 | NOT covered |
| B5 | if at 340:2 | `if err != nil {`; then `return res, fmt.Errorf("journal: starting the fill transaction for %s: %w", orderID, err)` (line last changed by `15b25b40`) | not a066 | NOT covered |
| B6 | switch at 346:2 | `switch {`; then `case errors.Is(err, ErrFillNotFound):` (line last changed by `15b25b40`) | not a066 | covered |
| B7 | case at 347:2 | `case err == nil:`; then `case errors.Is(err, ErrFillNotFound):` (line last changed by `15b25b40`) | not a066 | covered |
| B8 | case at 348:2 | `case errors.Is(err, ErrFillNotFound):`; then `prev = FillSnapshotRecord{}` (line last changed by `15b25b40`) | not a066 | covered |
| B9 | case at 350:2 | `default:`; then `return res, err` (line last changed by `15b25b40`) | not a066 | NOT covered |
| B10 | if at 356:2 | `if ownerErr != nil {`; then `return res, ownerErr` (line last changed by `c93f5f4a`) | not a066 | NOT covered |
| B11 | if at 364:2 | `if hadPrev && scope.complete() {`; then `if owners == 1 && prev.CommittedAt <= ownershipAt {` (line last changed by `c93f5f4a`) | not a066 | covered |
| B12 | if at 368:3 | `if owners == 1 && prev.CommittedAt <= ownershipAt {`; then `prev = FillSnapshotRecord{}` (line last changed by `c93f5f4a`) | not a066 | covered |
| B13 | if at 374:2 | `if err != nil {`; then `return res, fmt.Errorf("journal: ordering fill evidence for %s: %w", orderID, err)` (line last changed by `c93f5f4a`) | not a066 | NOT covered |
| B14 | if at 380:2 | `if hadPrev {`; then `prevFilled, _ = strconv.ParseFloat(strings.TrimSpace(orZero(prev.FilledQuantity)), 64)` (line last changed by `15b25b40`) | not a066 | covered |
| B15 | if at 385:2 | `if refusal := classifyFillRefusal(obs, prev, hadPrev, filled, prevFilled); refusal != nil {`; then `res.FailClosed = true` (line last changed by `15b25b40`) | not a066 | covered |
| B16 | if at 392:3 | `if err := markFillRefused(ctx, tx, scope, obs, refusal, now, hadPrev); err != nil {`; then `return res, err` (line last changed by `77d6dc60`) | not a066 | NOT covered |
| B17 | if at 401:3 | `if err != nil {`; then `return res, err` (line last changed by `8c173ede`) | not a066 | NOT covered |
| B18 | if at 405:3 | `if err := tx.Commit(); err != nil {`; then `return res, fmt.Errorf("journal: committing the refused snapshot of %s: %w", orderID, err)` (line last changed by `15b25b40`) | not a066 | NOT covered |
| B19 | if at 413:2 | `if nearlyZero(delta, filled) {`; then `delta = 0` (line last changed by `15b25b40`) | not a066 | covered |
| B20 | if at 420:2 | `if hadPrev && delta == 0 && sameSnapshot(obs, prev) {`; then `if err := tx.Commit(); err != nil {` (line last changed by `15b25b40`) | not a066 | covered |
| B21 | if at 421:3 | `if err := tx.Commit(); err != nil {`; then `return res, fmt.Errorf("journal: closing the no-op snapshot of %s: %w", orderID, err)` (line last changed by `15b25b40`) | not a066 | NOT covered |
| B22 | if at 437:2 | `if err := upsertFillSnapshot(ctx, tx, scope, obs, now); err != nil {`; then `return res, fmt.Errorf("journal: recording the fill snapshot of %s: %w", orderID, err)` (line last changed by `77d6dc60`) | not a066 | NOT covered |
| B23 | if at 441:2 | `if correction {`; then `if err := recordExecutionCorrection(ctx, tx, ExecutionCorrection{` (line last changed by `298c9443`) | not a066 | covered |
| B24 | if at 442:3 | `if err := recordExecutionCorrection(ctx, tx, ExecutionCorrection{`; then `return res, err` (line last changed by `298c9443`) | not a066 | NOT covered |
| B25 | if at 464:2 | `if delta > 0 {`; then `if _, err := tx.ExecContext(ctx,` (line last changed by `15b25b40`) | not a066 | covered |
| B26 | if at 465:3 | `if _, err := tx.ExecContext(ctx,`; then `return res, fmt.Errorf("journal: appending the fill of %s: %w", orderID, err)` (line last changed by `15b25b40`) | not a066 | NOT covered |
| B27 | if at 504:2 | `if err != nil {`; then `return res, err` (line last changed by `f7bb9e03`) | not a066 | NOT covered |
| B28 | if at 512:2 | `if obs.Terminal && locallyOwned {`; then `released, err := releaseReservationsForOrder(ctx, tx, orderID, ownedOrigin.IntentID,` (line last changed by `f7bb9e03`) | not a066 | covered |
| B29 | if at 516:3 | `if err != nil {`; then `return res, err` (line last changed by `8c173ede`) | not a066 | NOT covered |
| B30 | if at 526:2 | `if locallyOwned {`; then `if err := j.applyRiskBucketFillInTx(ctx, tx, applied); err != nil {` (line last changed by `c60fee07`) | a066: a066 risk-bucket sidecar runs only for a locally owned fill, inside the same BEGIN IMMEDIATE fill transaction as the snapshot, Position and hooks (Wave 1C) | covered |
| B31 | else at 536:9 | `} else if err := latchRiskBucketFillFailureForScope(ctx, tx, applied,`; then `"confirmed fill ownership is ambiguous for a registered risk order"); err != nil {` (line last changed by `4a364caf`) | a066: ownership is ambiguous: every registered, unreleased risk owner in scope gets a REPLAY_MISMATCH scope latch, unknown-actual flags and a FILL_UNACCOUNTED event; the broker fill itself is kept (Wave 1C/1D) | covered |
| B32 | if at 527:3 | `if err := j.applyRiskBucketFillInTx(ctx, tx, applied); err != nil {`; then `return res, fmt.Errorf("journal: applying risk buckets for fill %s: %w", orderID, err)` (line last changed by `c60fee07`) | a066: sidecar storage error aborts the whole fill transaction; semantic gaps latch entry inside the callee and return nil, so they never take this branch | NOT covered |
| B33 | if at 530:3 | `if err := j.releaseTerminalRiskBucketOrderInTx(ctx, tx, applied); err != nil {`; then `return res, fmt.Errorf("journal: releasing terminal risk buckets for fill %s: %w", orderID, err)` (line last changed by `8022f578`) | not a066 | NOT covered |
| B34 | if at 533:3 | `if err := applyWeeklyReservationLifecycleInTx(ctx, tx, applied); err != nil {`; then `return res, fmt.Errorf("journal: applying weekly reservation lifecycle for fill %s: %w", orderID, err)` (line last changed by `8022f578`) | not a066 | NOT covered |
| B35 | if at 536:9 | `} else if err := latchRiskBucketFillFailureForScope(ctx, tx, applied,`; then `return res, fmt.Errorf("journal: latching ambiguous risk ownership for fill %s: %w", orderID, err)` (line last changed by `4a364caf`) | a066: latch write error aborts the whole fill transaction | NOT covered |
| B36 | if at 547:2 | `if locallyOwned {`; then `if err := j.runApplyHooks(ctx, tx, applied); err != nil {` (line last changed by `f7bb9e03`) | not a066 | covered |
| B37 | if at 548:3 | `if err := j.runApplyHooks(ctx, tx, applied); err != nil {`; then `return res, err` (line last changed by `f7bb9e03`) | not a066 | covered |
| B38 | if at 553:2 | `if err := tx.Commit(); err != nil {`; then `return res, fmt.Errorf("journal: committing the fill snapshot of %s: %w", orderID, err)` (line last changed by `15b25b40`) | not a066 | NOT covered |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `lookupFillSnapshotScoped`, `confirmedFillOwners` | previous cumulative snapshot and ownership boundary | storage error returns before any write | CodeGraph callees + source |
| `classifyFillRefusal`, `markFillRefused`, `alertsForOrder` | caller/derived fail-closed verdicts are durable but do not advance the snapshot | refusal commits alone; holds stay | source |
| `upsertFillSnapshot`, `recordExecutionCorrection`, `fill_events` insert | advance the cumulative snapshot and append the delta | error aborts the transaction | source |
| `resolveFillOrigin`, `releaseReservationsForOrder` | ownership resolution and terminal hold release | error aborts the transaction | source |
| `applyRiskBucketFillInTx` (a066) | HELD→FILLED transfer, actual evidence, overage/unknown latches for registered exposure-raising orders | storage error aborts; semantic gaps latch entry and return nil; SELL/risk-reducing is a no-op (source comment at the call site) | Wave 1C tests; CodeGraph callers of `ApplyFill` |
| `releaseTerminalRiskBucketOrderInTx`, `applyWeeklyReservationLifecycleInTx` (`8022f578`) | terminal risk-order release and weekly reservation lifecycle | storage error aborts | source |
| `latchRiskBucketFillFailureForScope` (a066) | ambiguous ownership: `REPLAY_MISMATCH` scope latch, unknown-actual flags and a `FILL_UNACCOUNTED` event for every unreleased owner in scope | storage error aborts; never rejects the fill | Wave 1D tests |
| `runApplyHooks` | Position / campaign / a066 owner bind / exit projection in the same transaction | hook error aborts everything | `internal-journal--journal-.runapplyhooks` bundle |

The table is a hand summary; CodeGraph reports 34 callees (`codegraph callees RecordFill --json`).

## State mutations and fallbacks

- Writes: `fill_snapshots` upsert, `fill_events` append, refusal records, reservation release rows, a066
  risk-bucket fill/latch rows, weekly reservation lifecycle rows and whatever the apply hooks write — all in one
  transaction.
- No fallback ever drops or truncates an authoritative broker fill for a risk reason: a066 semantic failures
  become durable entry latches (`UNKNOWN_ACTUAL_RISK`, `RISK_OVERAGE`, or a `REPLAY_MISMATCH`
  scope latch with a `FILL_UNACCOUNTED` event), not errors.

## Safety conclusion

- High-risk impact: yes — fill apply path (safety invariant 5).
- a066's edits are limited to the sidecar block (B30–B32, B35). Fill detection itself is never delayed by a
  bucket evaluation: the sidecar is synchronous journal work inside the existing transaction with no network,
  FX or lock wait.
- Measured residual risk: 19 of 38 branch bodies are not executed by any `./internal/journal` or
  `./internal/riskbucket` test, including the a066 error exits B32 and B35 (storage failure inside the sidecar and
  the ambiguity latch). Their contract (abort the whole transaction) is structural, not tested.
