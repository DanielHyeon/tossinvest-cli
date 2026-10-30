# Function Logic Map: `probe.cycle`

Source: `tools/a128-jev-shadow-probe/collect.go`. Current AST and source digest: `ast.json`. Pre-edit snapshot: `../../pre-edit-ast/tools-a128-jev-shadow-probe--probe.cycle.json`.

## Inputs and invariants

market, session, timestamp, configured profile and source. Collects once; stores source state; judges one or two profiles independently; screens per profile; appends judgments/gaps and shared pending labels. Public-only data and isolated experimental output remain invariants.

## Branches and early returns

| Branch | Source condition |
|---|---|
| B1 | if at 349:3: `if p.cfg.InputProfile != "" && p.cfg.InputProfile != "current" {` |
| B2 | if at 357:2: `if err != nil {` |
| B3 | if at 363:2: `if err != nil {` |
| B4 | if at 367:2: `if len(received) > rankingFetch {` |
| B5 | if at 370:2: `if len(received) == 0 {` |
| B6 | if at 381:2: `if inputs.fxPair != "" {` |
| B7 | if at 384:2: `if len(candidates) == 0 {` |
| B8 | range at 391:2: `for _, item := range items {` |
| B9 | if at 398:2: `if stocks, err := withRateLimitRetry(ctx, p, groupStock, func(ctx context.Context) ([]domain.Quote, error) {` |
| B10 | else at 402:9: `} else {` |
| B11 | range at 403:3: `for _, quote := range stocks {` |
| B12 | if at 412:2: `if err != nil {` |
| B13 | range at 416:2: `for _, quote := range prices {` |
| B14 | range at 436:2: `for _, item := range items {` |
| B15 | if at 438:3: `if base <= 0 {` |
| B16 | if at 439:4: `if err := gap("prices", item.Symbol, errors.New("no positive last price for the symbol")); err != nil {` |
| B17 | if at 446:3: `if quote, found := meta[item.Symbol]; found {` |
| B18 | else at 448:10: `} else if len(cycleGaps) == 0 {` |
| B19 | if at 448:10: `} else if len(cycleGaps) == 0 {` |
| B20 | if at 455:3: `if err != nil {` |
| B21 | else at 457:10: `} else {` |
| B22 | if at 467:3: `if newsRead {` |
| B23 | if at 469:4: `if metaPtr != nil {` |
| B24 | if at 478:3: `if err != nil {` |
| B25 | if at 482:3: `if p.cfg.InputProfile == "initial" {` |
| B26 | if at 485:3: `if p.cfg.InputProfile == "compare" {` |
| B27 | range at 488:3: `for _, profile := range profiles {` |
| B28 | if at 490:4: `if profile == "initial" {` |
| B29 | if at 493:5: `if err != nil {` |
| B30 | if at 498:4: `if judgeErr != nil {` |
| B31 | if at 499:5: `if p.cfg.InputProfile == "" \|\| p.cfg.InputProfile == "current" {` |
| B32 | else at 503:12: `} else if err := p.ledger.gap(gapRow{At: stamp(at), Market: market, Stage: "judgment", Symbol: item.Symbol, Reason: reasonOf(judgeErr), InputProfile: profile, Cohort: p.cfg.Cohort, SourceStateDigest: digest}); err != nil {` |
| B33 | if at 500:6: `if err := gap("judgment", item.Symbol, judgeErr); err != nil {` |
| B34 | if at 503:12: `} else if err := p.ledger.gap(gapRow{At: stamp(at), Market: market, Stage: "judgment", Symbol: item.Symbol, Reason: reasonOf(judgeErr), InputProfile: profile, Cohort: p.cfg.Cohort, SourceStateDigest: digest}); err != nil {` |
| B35 | if at 510:4: `if !haveJ1 \|\| !haveJ2 {` |
| B36 | range at 527:2: `for _, profile := range []string{"current", "initial"} {` |
| B37 | range at 530:3: `for _, entry := range judged {` |
| B38 | if at 531:4: `if entry.profile == profile {` |
| B39 | if at 536:3: `if len(entries) == 0 {` |
| B40 | range at 542:3: `for _, screen := range screens {` |
| B41 | if at 543:4: `if screen.RedFlagExcluded {` |
| B42 | if at 546:4: `if screen.Finalist {` |
| B43 | range at 550:3: `for i, entry := range entries {` |
| B44 | if at 552:4: `if item == nil {` |
| B45 | range at 557:4: `for _, q := range p.questions {` |
| B46 | if at 559:5: `if p.cfg.InputProfile != "" && p.cfg.InputProfile != "current" {` |
| B47 | if at 563:5: `if p.cfg.InputProfile != "" && p.cfg.InputProfile != "current" {` |
| B48 | if at 565:6: `if profile == "initial" {` |
| B49 | if at 569:5: `if err := p.ledger.judgment(row); err != nil {` |
| B50 | if at 575:3: `if profile == "current" || p.cfg.InputProfile == "initial" {` |
| B51 | range at 579:2: `for _, symbol := range pendingOrder {` |
| B52 | if at 586:2: `if p.cfg.Interval > 0 && collect > p.cfg.Interval/2 {` |

## Calls and live bindings

The complete current call enumeration is in `ast.json`. The current source is authoritative; observed call context includes the local ledger, profile helper, and existing probe/report operations appropriate to this function. CodeGraph source context located probe.cycle and assembleState; precise isolated checkout graph freshness was unavailable.

## State mutations and fallbacks

Collects once; stores source state; judges one or two profiles independently; screens per profile; appends judgments/gaps and shared pending labels. Failures remain explicit; no production order or runtime configuration changes are introduced.

## Safety conclusion

Shadow measurement only. Existing broker/source calls are read-only; validation uses fake transports. Full production high-risk mutation review is not applicable.
