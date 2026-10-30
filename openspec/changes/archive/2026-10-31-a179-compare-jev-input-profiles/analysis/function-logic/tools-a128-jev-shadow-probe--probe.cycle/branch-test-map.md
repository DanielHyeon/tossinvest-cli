# Branch Test Map: `probe.cycle`

Relevant package test surfaces: TestCompareCycleSharesCollectionAndLabels, TestOneProfileFailureKeepsOtherResult, TestBothProfileFailuresKeepSourceAndAreCounted, TestDefaultAndExplicitCurrentProduceSameRows. These tests cover named scenarios, not every AST branch; rows below identify unproven individual branch coverage rather than asserting it. Existing socket-backed tests could not execute in this sandbox.

| Branch | Source branch | Individual branch coverage |
|---|---|---|
| B1 | if at 349:3: `if p.cfg.InputProfile != "" && p.cfg.InputProfile != "current" {` | Not individually established |
| B2 | if at 357:2: `if err != nil {` | Not individually established |
| B3 | if at 363:2: `if err != nil {` | Not individually established |
| B4 | if at 367:2: `if len(received) > rankingFetch {` | Not individually established |
| B5 | if at 370:2: `if len(received) == 0 {` | Not individually established |
| B6 | if at 381:2: `if inputs.fxPair != "" {` | Not individually established |
| B7 | if at 384:2: `if len(candidates) == 0 {` | Not individually established |
| B8 | range at 391:2: `for _, item := range items {` | Not individually established |
| B9 | if at 398:2: `if stocks, err := withRateLimitRetry(ctx, p, groupStock, func(ctx context.Context) ([]domain.Quote, error) {` | Not individually established |
| B10 | else at 402:9: `} else {` | Not individually established |
| B11 | range at 403:3: `for _, quote := range stocks {` | Not individually established |
| B12 | if at 412:2: `if err != nil {` | Not individually established |
| B13 | range at 416:2: `for _, quote := range prices {` | Not individually established |
| B14 | range at 436:2: `for _, item := range items {` | Not individually established |
| B15 | if at 438:3: `if base <= 0 {` | Not individually established |
| B16 | if at 439:4: `if err := gap("prices", item.Symbol, errors.New("no positive last price for the symbol")); err != nil {` | Not individually established |
| B17 | if at 446:3: `if quote, found := meta[item.Symbol]; found {` | Not individually established |
| B18 | else at 448:10: `} else if len(cycleGaps) == 0 {` | Not individually established |
| B19 | if at 448:10: `} else if len(cycleGaps) == 0 {` | Not individually established |
| B20 | if at 455:3: `if err != nil {` | Not individually established |
| B21 | else at 457:10: `} else {` | Not individually established |
| B22 | if at 467:3: `if newsRead {` | Not individually established |
| B23 | if at 469:4: `if metaPtr != nil {` | Not individually established |
| B24 | if at 478:3: `if err != nil {` | Not individually established |
| B25 | if at 482:3: `if p.cfg.InputProfile == "initial" {` | Not individually established |
| B26 | if at 485:3: `if p.cfg.InputProfile == "compare" {` | Not individually established |
| B27 | range at 488:3: `for _, profile := range profiles {` | Not individually established |
| B28 | if at 490:4: `if profile == "initial" {` | Not individually established |
| B29 | if at 493:5: `if err != nil {` | Not individually established |
| B30 | if at 498:4: `if judgeErr != nil {` | Not individually established |
| B31 | if at 499:5: `if p.cfg.InputProfile == "" \|\| p.cfg.InputProfile == "current" {` | Not individually established |
| B32 | else at 503:12: `} else if err := p.ledger.gap(gapRow{At: stamp(at), Market: market, Stage: "judgment", Symbol: item.Symbol, Reason: reasonOf(judgeErr), InputProfile: profile, Cohort: p.cfg.Cohort, SourceStateDigest: digest}); err != nil {` | Not individually established |
| B33 | if at 500:6: `if err := gap("judgment", item.Symbol, judgeErr); err != nil {` | Not individually established |
| B34 | if at 503:12: `} else if err := p.ledger.gap(gapRow{At: stamp(at), Market: market, Stage: "judgment", Symbol: item.Symbol, Reason: reasonOf(judgeErr), InputProfile: profile, Cohort: p.cfg.Cohort, SourceStateDigest: digest}); err != nil {` | Not individually established |
| B35 | if at 510:4: `if !haveJ1 \|\| !haveJ2 {` | Not individually established |
| B36 | range at 527:2: `for _, profile := range []string{"current", "initial"} {` | Not individually established |
| B37 | range at 530:3: `for _, entry := range judged {` | Not individually established |
| B38 | if at 531:4: `if entry.profile == profile {` | Not individually established |
| B39 | if at 536:3: `if len(entries) == 0 {` | Not individually established |
| B40 | range at 542:3: `for _, screen := range screens {` | Not individually established |
| B41 | if at 543:4: `if screen.RedFlagExcluded {` | Not individually established |
| B42 | if at 546:4: `if screen.Finalist {` | Not individually established |
| B43 | range at 550:3: `for i, entry := range entries {` | Not individually established |
| B44 | if at 552:4: `if item == nil {` | Not individually established |
| B45 | range at 557:4: `for _, q := range p.questions {` | Not individually established |
| B46 | if at 559:5: `if p.cfg.InputProfile != "" && p.cfg.InputProfile != "current" {` | Not individually established |
| B47 | if at 563:5: `if p.cfg.InputProfile != "" && p.cfg.InputProfile != "current" {` | Not individually established |
| B48 | if at 565:6: `if profile == "initial" {` | Not individually established |
| B49 | if at 569:5: `if err := p.ledger.judgment(row); err != nil {` | Not individually established |
| B50 | if at 575:3: `if profile == "current" || p.cfg.InputProfile == "initial" {` | Not individually established |
| B51 | range at 579:2: `for _, symbol := range pendingOrder {` | Not individually established |
| B52 | if at 586:2: `if p.cfg.Interval > 0 && collect > p.cfg.Interval/2 {` | Not individually established |
