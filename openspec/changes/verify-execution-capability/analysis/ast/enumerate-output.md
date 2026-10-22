# verify-execution-capability AST 열거 출력 (`python3 -I enumerate.py` 생성 — 손편집 금지)

## `internal/brokerstate/derive.go` `Derive` (346–510)

- 산출물 `brokerstate--derive.ast.json` · 분기 50 · 반환 35 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:switch | B1 | 354 | `switch {` |  |
| branch:case | B2 | 355 | `case isNotFinite(v.Quantity) \|\| isNotFinite(v.FilledQuantity):` |  |
| branch:case | B3 | 358 | `case v.Quantity <= 0:` |  |
| branch:case | B4 | 361 | `case v.FilledQuantity < 0 && !qtyEqual(v.FilledQuantity, 0):` |  |
| branch:case | B5 | 364 | `case qtyGreater(v.FilledQuantity, v.Quantity):` |  |
| branch:if | B6 | 378 | `if f.filledAll {` |  |
| branch:switch | B7 | 383 | `switch status {` |  |
| branch:case | B8 | 384 | `case "":` |  |
| branch:case | B9 | 387 | `case rawStatusPending: // 6-9` | MARK |
| branch:switch | B10 | 388 | `switch {` |  |
| branch:case | B11 | 389 | `case f.canceled:` |  |
| branch:case | B12 | 392 | `case f.successor != "":` |  |
| branch:case | B13 | 395 | `case !f.filledNone:` |  |
| branch:case | B14 | 398 | `default:` | MARK |
| branch:case | B15 | 402 | `case rawStatusPartialFilled: // 10-14` | MARK |
| branch:switch | B16 | 403 | `switch {` |  |
| branch:case | B17 | 404 | `case f.canceled:` |  |
| branch:case | B18 | 407 | `case f.successor != "":` |  |
| branch:case | B19 | 410 | `case f.filledNone:` |  |
| branch:case | B20 | 413 | `case f.filledAll:` |  |
| branch:case | B21 | 416 | `default:` | MARK |
| branch:case | B22 | 420 | `case rawStatusPendingCancel: // 15-18` | MARK |
| branch:case | B23 | 423 | `case rawStatusPendingReplace: // 15-18` | MARK |
| branch:case | B24 | 426 | `case rawStatusFilled: // 19-22` | MARK |
| branch:switch | B25 | 427 | `switch {` |  |
| branch:case | B26 | 428 | `case !f.filledAll:` |  |
| branch:case | B27 | 431 | `case f.successor != "":` |  |
| branch:case | B28 | 434 | `case f.canceled:` |  |
| branch:case | B29 | 437 | `default:` | MARK |
| branch:case | B30 | 441 | `case rawStatusCanceled: // 23-26` | MARK |
| branch:switch | B31 | 442 | `switch {` |  |
| branch:case | B32 | 443 | `case f.successor != "":` |  |
| branch:case | B33 | 447 | `case f.filledAll:` |  |
| branch:case | B34 | 450 | `case f.filledNone:` |  |
| branch:case | B35 | 452 | `default:` | MARK |
| branch:case | B36 | 456 | `case rawStatusRejected: // 27-30` | MARK |
| branch:switch | B37 | 457 | `switch {` |  |
| branch:case | B38 | 458 | `case f.successor != "":` |  |
| branch:case | B39 | 461 | `case f.filledAll:` |  |
| branch:case | B40 | 464 | `case f.filledNone:` |  |
| branch:case | B41 | 466 | `default:` | MARK |
| branch:case | B42 | 470 | `case rawStatusReplaced: // 31-33` | MARK |
| branch:switch | B43 | 471 | `switch {` |  |
| branch:case | B44 | 472 | `case f.filledAll:` |  |
| branch:case | B45 | 475 | `case f.successor == "":` |  |
| branch:case | B46 | 483 | `default:` | MARK |
| branch:case | B47 | 487 | `case rawStatusCancelRejected: // 34` | MARK |
| branch:case | B48 | 497 | `case rawStatusReplaceRejected: // 35` | MARK |
| branch:case | B49 | 503 | `case rawGroupOpen, rawGroupClosed: // 36` | MARK |
| branch:case | B50 | 506 | `default: // 37` | MARK |
| return | | 356 | `return d.unknown(ReasonNonFiniteQuantity,` | MARK |
| return | | 359 | `return d.unknown(ReasonQuantityNotPositive,` | MARK |
| return | | 362 | `return d.unknown(ReasonFilledNegative,` | MARK |
| return | | 365 | `return d.unknown(ReasonFilledExceedsQuantity,` | MARK |
| return | | 385 | `return d.unknown(ReasonMissingStatus, "broker returned no status")` | MARK |
| return | | 390 | `return d.unknown(ReasonOpenWithCancelTimestamp,` | MARK |
| return | | 393 | `return d.unknown(ReasonOpenWithSuccessor,` | MARK |
| return | | 396 | `return d.unknown(ReasonPendingWithFill, fmt.Sprintf(` | MARK |
| return | | 399 | `return d.settle(StateOpenUnfilled)` |  |
| return | | 405 | `return d.unknown(ReasonOpenWithCancelTimestamp,` | MARK |
| return | | 408 | `return d.unknown(ReasonOpenWithSuccessor,` | MARK |
| return | | 411 | `return d.unknown(ReasonPartialFilledWithoutFill,` | MARK |
| return | | 414 | `return d.unknown(ReasonPartialFilledWithFullFill, fmt.Sprintf(` | MARK |
| return | | 417 | `return d.settle(StateOpenPartiallyFilled)` |  |
| return | | 421 | `return d.pendingMutation(f, "PENDING_CANCEL", StateCancelPending)` |  |
| return | | 424 | `return d.pendingMutation(f, "PENDING_REPLACE", StateReplacePending)` |  |
| return | | 429 | `return d.unknown(ReasonFilledWithRemainder, fmt.Sprintf(` | MARK |
| return | | 432 | `return d.unknown(ReasonReplacedWithFullFill,` | MARK |
| return | | 435 | `return d.unknown(ReasonCancelledWithFullFill,` | MARK |
| return | | 438 | `return d.settle(StateFilled)` |  |
| return | | 444 | `return d.unknown(ReasonCancelledWithSuccessor,` | MARK |
| return | | 448 | `return d.unknown(ReasonCancelledWithFullFill,` | MARK |
| return | | 451 | `return d.settle(StateCancelled)` |  |
| return | | 453 | `return d.settle(StateCancelledPartiallyFilled)` |  |
| return | | 459 | `return d.unknown(ReasonRejectedWithSuccessor,` | MARK |
| return | | 462 | `return d.unknown(ReasonRejectedWithFullFill,` | MARK |
| return | | 465 | `return d.settle(StateRejected)` |  |
| return | | 467 | `return d.settle(StateRejectedPartiallyFilled)` |  |
| return | | 473 | `return d.unknown(ReasonReplacedWithFullFill,` | MARK |
| return | | 480 | `return d.unknown(ReasonReplacedWithoutSuccessor, fmt.Sprintf(` | MARK |
| return | | 484 | `return d.settle(StateReplaced)` |  |
| return | | 493 | `return d.unknown(ReasonCancelRejectedRecord,` | MARK |
| return | | 499 | `return d.unknown(ReasonReplaceRejectedRecord,` | MARK |
| return | | 504 | `return d.deriveGroupLabel(status, f)` |  |
| return | | 507 | `return d.unknown(ReasonUnknownStatus,` | MARK |
| call | | 356 | `d.unknown` | MARK |
| call | | 359 | `d.unknown` | MARK |
| call | | 362 | `d.unknown` | MARK |
| call | | 365 | `d.unknown` | MARK |
| call | | 385 | `d.unknown` | MARK |
| call | | 390 | `d.unknown` | MARK |
| call | | 393 | `d.unknown` | MARK |
| call | | 396 | `d.unknown` | MARK |
| call | | 405 | `d.unknown` | MARK |
| call | | 408 | `d.unknown` | MARK |
| call | | 411 | `d.unknown` | MARK |
| call | | 414 | `d.unknown` | MARK |
| call | | 429 | `d.unknown` | MARK |
| call | | 432 | `d.unknown` | MARK |
| call | | 435 | `d.unknown` | MARK |
| call | | 444 | `d.unknown` | MARK |
| call | | 448 | `d.unknown` | MARK |
| call | | 459 | `d.unknown` | MARK |
| call | | 462 | `d.unknown` | MARK |
| call | | 473 | `d.unknown` | MARK |
| call | | 480 | `d.unknown` | MARK |
| call | | 493 | `d.unknown` | MARK |
| call | | 499 | `d.unknown` | MARK |
| call | | 507 | `d.unknown` | MARK |

## `internal/brokerstate/derive.go` `Derived.deriveGroupLabel` (537–575)

- 산출물 `brokerstate--derived.derivegrouplabel.ast.json` · 분기 15 · 반환 12 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 538 | `if label == rawGroupOpen {` | MARK |
| branch:switch | B2 | 539 | `switch {` |  |
| branch:case | B3 | 540 | `case f.canceled:` |  |
| branch:case | B4 | 543 | `case f.successor != "":` |  |
| branch:case | B5 | 546 | `case f.filledNone:` |  |
| branch:case | B6 | 548 | `default:` | MARK |
| branch:switch | B7 | 552 | `switch {` |  |
| branch:case | B8 | 553 | `case f.successor != "" && f.filledAll:` |  |
| branch:case | B9 | 556 | `case f.successor != "":` |  |
| branch:case | B10 | 558 | `case f.canceled && f.filledAll:` |  |
| branch:case | B11 | 561 | `case f.canceled && f.filledNone:` |  |
| branch:case | B12 | 563 | `case f.canceled:` |  |
| branch:case | B13 | 565 | `case f.filledAll:` |  |
| branch:case | B14 | 567 | `case f.filledNone:` |  |
| branch:case | B15 | 570 | `default:` | MARK |
| return | | 541 | `return d.unknown(ReasonOpenWithCancelTimestamp,` | MARK |
| return | | 544 | `return d.unknown(ReasonOpenWithSuccessor,` | MARK |
| return | | 547 | `return d.settle(StateOpenUnfilled)` |  |
| return | | 549 | `return d.settle(StateOpenPartiallyFilled)` |  |
| return | | 554 | `return d.unknown(ReasonReplacedWithFullFill,` | MARK |
| return | | 557 | `return d.settle(StateReplaced)` |  |
| return | | 559 | `return d.unknown(ReasonCancelledWithFullFill,` | MARK |
| return | | 562 | `return d.settle(StateCancelled)` |  |
| return | | 564 | `return d.settle(StateCancelledPartiallyFilled)` |  |
| return | | 566 | `return d.settle(StateFilled)` |  |
| return | | 568 | `return d.unknown(ReasonClosedWithoutFillOrCancel,` | MARK |
| return | | 571 | `return d.unknown(ReasonClosedWithUnexplainedRemainder,` | MARK |
| call | | 541 | `d.unknown` | MARK |
| call | | 544 | `d.unknown` | MARK |
| call | | 554 | `d.unknown` | MARK |
| call | | 559 | `d.unknown` | MARK |
| call | | 568 | `d.unknown` | MARK |
| call | | 571 | `d.unknown` | MARK |

## `internal/brokerstate/derive.go` `Derived.pendingMutation` (515–529)

- 산출물 `brokerstate--derived.pendingmutation.ast.json` · 분기 5 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:switch | B1 | 516 | `switch {` |  |
| branch:case | B2 | 517 | `case f.canceled:` |  |
| branch:case | B3 | 520 | `case f.successor != "":` |  |
| branch:case | B4 | 523 | `case f.filledAll:` |  |
| branch:case | B5 | 526 | `default:` | MARK |
| return | | 518 | `return d.unknown(ReasonOpenWithCancelTimestamp,` | MARK |
| return | | 521 | `return d.unknown(ReasonOpenWithSuccessor,` | MARK |
| return | | 524 | `return d.unknown(ReasonPendingMutationWithFullFill, fmt.Sprintf(` | MARK |
| return | | 527 | `return d.settle(state)` |  |
| call | | 518 | `d.unknown` | MARK |
| call | | 521 | `d.unknown` | MARK |
| call | | 524 | `d.unknown` | MARK |

## `internal/brokerstate/derive.go` `Derived.settle` (577–586)

- 산출물 `brokerstate--derived.settle.ast.json` · 분기 1 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 582 | `if state != StateReplaced {` |  |
| return | | 585 | `return d` |  |

## `internal/brokerstate/derive.go` `Derived.unknown` (588–595)

- 산출물 `brokerstate--derived.unknown.ast.json` · 분기 0 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| return | | 594 | `return d` |  |

## `internal/app/engine/gateway.go` `buildGateway` (245–370)

- 산출물 `engine--buildgateway.ast.json` · 분기 6 · 반환 8 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 250 | `if err := checkProjectionWired(in.journal); err != nil {` |  |
| branch:if | B2 | 272 | `if err := tracker.Restore(ctx); err != nil {` |  |
| branch:if | B3 | 280 | `if err := restoreAlertEntryLatch(ctx, in.journal, entry); err != nil {` |  |
| branch:if | B4 | 284 | `if err := bindOperatingModeProjection(ctx, in.journal, entry, in.accountRef, in.logger); err != nil {` |  |
| branch:if | B5 | 308 | `if err != nil {` |  |
| branch:if | B6 | 332 | `if err != nil {` |  |
| return | | 251 | `return engineWiring{}, err` |  |
| return | | 273 | `return engineWiring{}, fmt.Errorf("engine: restoring the RECONCILE projection: %w", err)` |  |
| return | | 281 | `return engineWiring{}, fmt.Errorf("engine: restoring the alert entry latch: %w", err)` |  |
| return | | 285 | `return engineWiring{}, fmt.Errorf("engine: binding the operating-mode projection: %w", err)` |  |
| return | | 288 | `return tracker.Refresh(context.Background())` |  |
| return | | 309 | `return engineWiring{}, fmt.Errorf("engine: constructing the paired read-only protection readiness adapter: %w", err)` |  |
| return | | 333 | `return engineWiring{}, fmt.Errorf("engine: constructing the execution gateway: %w", err)` |  |
| return | | 355 | `return engineWiring{` |  |

## `internal/execgw/replay.go` `Gateway.ReplayInDoubt` (228–374)

- 산출물 `execgw--gateway.replayindoubt.ast.json` · 분기 25 · 반환 19 · go 0 · defer 1

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 233 | `if err != nil {` |  |
| branch:if | B2 | 238 | `if err != nil {` |  |
| branch:if | B3 | 245 | `if rec.State.IsTerminal() {` |  |
| branch:if | B4 | 251 | `if g.replay == nil {` | MARK |
| branch:if | B5 | 259 | `if g.attested == nil \|\| !g.attested(ctx) {` | MARK |
| branch:if | B6 | 268 | `if !g.claimSymbol(symbolKey) {` |  |
| branch:if | B7 | 275 | `if err != nil {` |  |
| branch:for | B8 | 280 | `for {` |  |
| branch:if | B9 | 285 | `if err != nil {` |  |
| branch:if | B10 | 290 | `if rejected := g.replayGuards(rec, cfg); rejected != nil {` | MARK |
| branch:if | B11 | 296 | `if wait := replayWait(rec, cfg, g.clk.Now()); wait > 0 {` | MARK |
| branch:if | B12 | 297 | `if waits >= cfg.MaxWaits {` |  |
| branch:if | B13 | 302 | `if err := g.clk.Sleep(ctx, wait); err != nil {` |  |
| branch:switch | B14 | 309 | `switch {` |  |
| branch:case | B15 | 310 | `case errors.Is(err, journal.ErrReplayCapReached):` | MARK |
| branch:case | B16 | 314 | `case errors.Is(err, journal.ErrReplayNotInDoubt):` | MARK |
| branch:case | B17 | 316 | `case err != nil:` |  |
| branch:switch | B18 | 326 | `switch verdict.kind {` |  |
| branch:case | B19 | 327 | `case replayRecovered:` | MARK |
| branch:if | B20 | 331 | `if err := attempt.ResolveConfirmed(ctx, verdict.orderID,` |  |
| branch:case | B21 | 341 | `case replayInProgress:` | MARK |
| branch:if | B22 | 346 | `if rerr != nil {` |  |
| branch:case | B23 | 354 | `case replayKeyConflict:` | MARK |
| branch:if | B24 | 357 | `if err := attempt.ResolveUnresolved(ctx, journal.ReasonReplayKeyConflict, verdict.detail); err != nil {` | MARK |
| branch:case | B25 | 366 | `default:` | MARK |
| return | | 234 | `return out, err` |  |
| return | | 239 | `return out, err` |  |
| return | | 249 | `return out, nil` |  |
| return | | 252 | `return declineReplay(out, ReasonReplayIneligible,` | MARK |
| return | | 260 | `return declineReplay(out, ReasonReplayNotAttested,` | MARK |
| return | | 269 | `return declineReplay(out, ReasonSymbolInFlight,` | MARK |
| return | | 276 | `return out, err` |  |
| return | | 286 | `return out, err` |  |
| return | | 291 | `return declineReplay(out, rejected.Reason, rejected.Detail), nil` | MARK |
| return | | 298 | `return declineReplay(out, ReasonReplayExhausted, fmt.Sprintf(` | MARK |
| return | | 303 | `return out, err` |  |
| return | | 312 | `return declineReplay(out, ReasonReplayExhausted, fmt.Sprintf(` | MARK |
| return | | 315 | `return declineReplay(out, ReasonReplayIneligible, err.Error()), nil` | MARK |
| return | | 317 | `return out, err` |  |
| return | | 333 | `return out, err` |  |
| return | | 339 | `return out, nil` |  |
| return | | 347 | `return out, rerr` |  |
| return | | 358 | `return out, err` |  |
| return | | 364 | `return out, nil` |  |
| call | | 229 | `g.replayCfg.withDefaults` | MARK |
| call | | 252 | `declineReplay` | MARK |
| call | | 259 | `g.attested` | MARK |
| call | | 260 | `declineReplay` | MARK |
| call | | 269 | `declineReplay` | MARK |
| call | | 290 | `g.replayGuards` | MARK |
| call | | 291 | `declineReplay` | MARK |
| call | | 296 | `replayWait` | MARK |
| call | | 298 | `declineReplay` | MARK |
| call | | 308 | `g.journal.MarkReplayStarted` | MARK |
| call | | 312 | `declineReplay` | MARK |
| call | | 315 | `declineReplay` | MARK |
| call | | 323 | `g.replay.ReplayPlace` | MARK |
| call | | 324 | `classifyReplay` | MARK |
| call | | 345 | `g.journal.RefundReplay` | MARK |

## `internal/execgw/replay.go` `ReplayConfig.withDefaults` (130–144)

- 산출물 `execgw--replayconfig.withdefaults.ast.json` · 분기 4 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 131 | `if c.Margin <= 0 {` | MARK |
| branch:if | B2 | 134 | `if c.MaxReplays <= 0 {` | MARK |
| branch:if | B3 | 137 | `if c.MinInterval <= 0 {` |  |
| branch:if | B4 | 140 | `if c.MaxWaits <= 0 {` |  |
| return | | 143 | `return c` |  |

## `internal/filldetect/detect.go` `Detector.blockSymbol` (669–683)

- 산출물 `filldetect--detector.blocksymbol.ast.json` · 분기 3 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 670 | `if d.Gate == nil {` |  |
| branch:if | B2 | 674 | `if reason == "" {` |  |
| branch:if | B3 | 678 | `if detail == "" {` |  |
| return | | 671 | `return` |  |

## `internal/filldetect/detect.go` `Detector.pollLocked` (283–357)

- 산출물 `filldetect--detector.polllocked.ast.json` · 분기 10 · 반환 6 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 285 | `if err := d.validate(); err != nil {` |  |
| branch:if | B2 | 294 | `if err != nil {` |  |
| branch:if | B3 | 301 | `if err != nil {` |  |
| branch:if | B4 | 309 | `if err != nil {` |  |
| branch:range | B5 | 316 | `for _, snap := range snaps {` |  |
| branch:if | B6 | 318 | `if err != nil {` |  |
| branch:if | B7 | 324 | `if applied.FailClosed {` | MARK |
| branch:if | B8 | 329 | `if applied.Delta > 0 {` |  |
| branch:if | B9 | 332 | `if committed.IsZero() {` |  |
| branch:if | B10 | 336 | `if latency < 0 {` |  |
| return | | 286 | `return Cycle{}, err` |  |
| return | | 297 | `return cycle, err` |  |
| return | | 304 | `return cycle, fmt.Errorf("filldetect: sweeping the account: %w", err)` |  |
| return | | 312 | `return cycle, err` |  |
| return | | 321 | `return cycle, fmt.Errorf("filldetect: committing the %s snapshot: %w", snap.OrderID, err)` |  |
| return | | 356 | `return cycle, nil` |  |

## `internal/filldetect/ledger.go` `JournalLedger.Apply` (36–110)

- 산출물 `filldetect--journalledger.apply.ast.json` · 분기 11 · 반환 7 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 37 | `if l.Journal == nil {` |  |
| branch:if | B2 | 42 | `if selectedAccount == "" {` |  |
| branch:if | B3 | 45 | `if observedAccount == "" {` |  |
| branch:if | B4 | 48 | `if observedAccount != selectedAccount {` |  |
| branch:if | B5 | 76 | `if snap.HasBrokerTime {` |  |
| branch:if | B6 | 81 | `if err != nil {` |  |
| branch:if | B7 | 84 | `if l.Refresh != nil {` |  |
| branch:if | B8 | 85 | `if err := l.Refresh(ctx); err != nil {` |  |
| branch:if | B9 | 97 | `if res.FailClosed {` | MARK |
| branch:if | B10 | 102 | `if res.Detail != "" {` |  |
| branch:if | B11 | 106 | `if ts, perr := time.Parse(time.RFC3339, res.CommittedAt); perr == nil {` |  |
| return | | 38 | `return Applied{}, fmt.Errorf("filldetect: the ledger has no journal")` |  |
| return | | 43 | `return Applied{}, fmt.Errorf("filldetect: the ledger has no selected account")` |  |
| return | | 46 | `return Applied{}, fmt.Errorf("filldetect: snapshot %q has no account scope", snap.OrderID)` |  |
| return | | 49 | `return Applied{}, fmt.Errorf(` |  |
| return | | 82 | `return Applied{}, err` |  |
| return | | 86 | `return Applied{}, fmt.Errorf("filldetect: refreshing RECONCILE authority after the fill: %w", err)` |  |
| return | | 109 | `return applied, nil` |  |

합계: AST 11 · 분기 130 · 반환 95 · STALE 0
