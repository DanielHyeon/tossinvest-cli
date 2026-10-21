# a163 AST 열거 출력 (`python3 -I enumerate.py` 생성 — 손편집 금지)

## `internal/breakoutlane/rules.go` `validateQuote` (32–59)

- 분기 7 · 반환 8 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 33 | `if !quoteValid(q) \|\| entry == 0 {` |  |
| branch:if | B2 | 37 | `if v.SourceObservedAtMS > v.ReceivedAtMS \|\| v.ReceivedAtMS > evaluated {` | CLOCK |
| branch:if | B3 | 40 | `if evaluated-v.SourceObservedAtMS > c.value.MaxQuoteAgeMS {` | CLOCK |
| branch:if | B4 | 45 | `if ov {` |  |
| branch:if | B5 | 48 | `if spread > c.value.MaxSpreadPPM {` |  |
| branch:if | B6 | 52 | `if ov {` |  |
| branch:if | B7 | 55 | `if drift > c.value.MaxEntryDriftPPM {` |  |
| return | | 34 | `return RefusalEvidenceInvalid` |  |
| return | | 38 | `return RefusalQuoteStale` | CLOCK |
| return | | 41 | `return RefusalQuoteStale` | CLOCK |
| return | | 46 | `return RefusalSizingOverflow` |  |
| return | | 49 | `return RefusalSpreadTooWide` |  |
| return | | 53 | `return RefusalSizingOverflow` |  |
| return | | 56 | `return RefusalEntryDriftExceeded` |  |
| return | | 58 | `return RefusalNone` |  |

## `internal/app/engine/exitloop.go` `ExitObserver.checkOutage` (916–958)

- 분기 6 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 918 | `if since.IsZero() {` | CLOCK |
| branch:if | B2 | 921 | `if o.clk.Now().Sub(since) < o.outageAfter() {` | CLOCK |
| branch:if | B3 | 924 | `if o.outageRaised {` |  |
| branch:if | B4 | 942 | `if o.opts.Escalate == nil \|\| strings.TrimSpace(o.opts.AccountRef) == "" {` |  |
| branch:if | B5 | 947 | `if err != nil {` |  |
| branch:if | B6 | 948 | `if !errors.Is(err, journal.ErrModeAnnouncementFailed) {` |  |
| return | | 922 | `return` |  |
| return | | 925 | `return` |  |
| return | | 943 | `return` |  |
| return | | 951 | `return` |  |
| call | | 918 | `since.IsZero` | CLOCK |
| call | | 921 | `o.clk.Now` | CLOCK |
| call | | 921 | `o.outageAfter` | CLOCK |
| call | | 932 | `o.clk.Now` | CLOCK |
| call | | 938 | `o.clk.Now` | CLOCK |

## `internal/app/engine/exitloop.go` `ExitObserver.judge` (961–1004)

- 분기 9 · 반환 7 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 963 | `if !o.quoteUsable(quote) {` | CLOCK |
| branch:if | B2 | 970 | `if o.noteHeldProposal(ctx, m) {` |  |
| branch:if | B3 | 973 | `if m.identityErr != nil {` |  |
| branch:if | B4 | 977 | `if m.reJudge {` |  |
| branch:if | B5 | 987 | `if err := o.opts.Journal.StampExitSnapshotQuarantineSelector(ctx,` |  |
| branch:if | B6 | 993 | `if err != nil {` |  |
| branch:switch | B7 | 998 | `switch m.state.PolicyKind {` |  |
| branch:case | B8 | 999 | `case journal.ExitPolicyLadder:` |  |
| branch:case | B9 | 1001 | `default:` |  |
| return | | 964 | `return nil` |  |
| return | | 971 | `return nil` |  |
| return | | 975 | `return nil` |  |
| return | | 989 | `return err` |  |
| return | | 995 | `return nil` |  |
| return | | 1000 | `return o.judgeLadder(ctx, m, quote, observation, breakEven, cycle)` |  |
| return | | 1002 | `return o.judgeRatchet(ctx, m, quote, observation, breakEven, cycle)` |  |
| call | | 963 | `o.quoteUsable` | CLOCK |

## `internal/app/engine/exitloop.go` `ExitObserver.judgeLadder` (1079–1146)

- 분기 8 · 반환 7 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 1082 | `if err != nil {` |  |
| branch:if | B2 | 1086 | `if err := o.checkLadderPolicyStillFits(m, ladder); err != nil {` |  |
| branch:if | B3 | 1091 | `if err != nil {` |  |
| branch:if | B4 | 1096 | `if m.state.PendingLevel != "" {` |  |
| branch:if | B5 | 1097 | `if idx, err := exitpolicy.RungIndex(m.state.PendingLevel); err == nil {` |  |
| branch:if | B6 | 1125 | `if err != nil {` |  |
| branch:if | B7 | 1138 | `if !o.quoteUsable(quote) {` | CLOCK |
| branch:if | B8 | 1141 | `if !o.requiresFullJudgement(m, snapshot) {` |  |
| return | | 1084 | `return nil` |  |
| return | | 1088 | `return nil` |  |
| return | | 1093 | `return nil` |  |
| return | | 1127 | `return nil` |  |
| return | | 1139 | `return nil` |  |
| return | | 1142 | `return o.refreshObservation(ctx, m, snapshot, exitpolicy.NewLadderRecoveryPolicy(evaluation), quote)` |  |
| return | | 1144 | `return o.record(ctx, m, snapshot, exitpolicy.NewLadderRecoveryPolicy(evaluation),` |  |
| call | | 1138 | `o.quoteUsable` | CLOCK |

## `internal/app/engine/exitloop.go` `ExitObserver.judgeRatchet` (1025–1077)

- 분기 6 · 반환 6 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 1028 | `if err != nil \|\| !samePolicyIdentity(m.policyIdentity, activeIdentity) {` |  |
| branch:if | B2 | 1029 | `if err == nil {` |  |
| branch:if | B3 | 1039 | `if err != nil {` |  |
| branch:if | B4 | 1059 | `if err != nil {` |  |
| branch:if | B5 | 1067 | `if !o.quoteUsable(quote) {` | CLOCK |
| branch:if | B6 | 1070 | `if !o.requiresFullJudgement(m, snapshot) {` |  |
| return | | 1036 | `return nil` |  |
| return | | 1041 | `return nil` |  |
| return | | 1061 | `return nil` |  |
| return | | 1068 | `return nil` |  |
| return | | 1071 | `return o.refreshObservation(ctx, m, snapshot, exitpolicy.NewRatchetRecoveryPolicy(evaluation), quote)` |  |
| return | | 1074 | `return o.record(ctx, m, snapshot,` |  |
| call | | 1067 | `o.quoteUsable` | CLOCK |

## `internal/app/engine/exitloop.go` `ExitObserver.observe` (842–913)

- 분기 15 · 반환 6 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 849 | `if qerr != nil {` |  |
| branch:range | B2 | 858 | `for _, symbol := range symbols {` |  |
| branch:range | B3 | 862 | `for _, q := range quotes {` |  |
| branch:if | B4 | 864 | `if _, needed := wanted[symbol]; !needed \|\| q.Last <= 0 \|\| math.IsNaN(q.Last) \|\| math.IsInf(q.Last, 0) {` |  |
| branch:if | B5 | 868 | `if q.FetchedAt.IsZero() {` | CLOCK |
| branch:else | B6 | 871 | `} else if effectiveAt.After(validatedAt) \|\|` | CLOCK |
| branch:if | B7 | 871 | `} else if effectiveAt.After(validatedAt) \|\|` | CLOCK |
| branch:if | B8 | 878 | `if len(next) == 0 {` |  |
| branch:if | B9 | 884 | `if err != nil {` |  |
| branch:if | B10 | 887 | `if needsFallback {` |  |
| branch:if | B11 | 893 | `if err != nil {` |  |
| branch:for | B12 | 896 | `for {` |  |
| branch:if | B13 | 898 | `if current >= maximum \|\| o.observationSequence.CompareAndSwap(current, maximum) {` |  |
| branch:range | B14 | 904 | `for symbol, quote := range out {` |  |
| branch:if | B15 | 905 | `if quote.Source != "" {` |  |
| return | | 850 | `return qerr` |  |
| return | | 879 | `return execgw.NewEvidenceInvalidError(fmt.Sprintf("no valid quote evidence for %d managed symbol(s)", len(symbols)))` |  |
| return | | 882 | `return nil` |  |
| return | | 885 | `return nil, err` |  |
| return | | 894 | `return nil, err` |  |
| return | | 912 | `return out, nil` |  |
| call | | 855 | `clock.LeaseAnchor` | CLOCK |
| call | | 867 | `q.FetchedAt.UTC` | CLOCK |
| call | | 868 | `q.FetchedAt.IsZero` | CLOCK |
| call | | 871 | `effectiveAt.After` | CLOCK |
| call | | 872 | `validatedAt.Sub` | CLOCK |

## `internal/app/engine/exitloop.go` `ExitObserver.ObserveOnce` (457–526)

- 분기 8 · 반환 5 · go 0 · defer 2

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 465 | `if o.opts.SLO != nil && o.opts.SLO.FillDetectionBehind() {` |  |
| branch:if | B2 | 475 | `if err != nil {` |  |
| branch:if | B3 | 479 | `if len(states) == 0 {` |  |
| branch:if | B4 | 492 | `if err != nil {` |  |
| branch:range | B5 | 501 | `for _, state := range states {` |  |
| branch:if | B6 | 503 | `if !ok {` |  |
| branch:if | B7 | 511 | `if !o.quoteUsable(quote) {` | CLOCK |
| branch:if | B8 | 519 | `if err := o.judge(ctx, state, quote, observation, &cycle); err != nil && cycle.Err == nil {` |  |
| return | | 471 | `return cycle` |  |
| return | | 477 | `return cycle` |  |
| return | | 489 | `return cycle` |  |
| return | | 495 | `return cycle` |  |
| return | | 525 | `return cycle` |  |
| call | | 463 | `o.clk.Now` | CLOCK |
| call | | 487 | `o.clk.Now` | CLOCK |
| call | | 497 | `o.clk.Now` | CLOCK |
| call | | 511 | `o.quoteUsable` | CLOCK |

## `internal/app/engine/exitloop.go` `ExitObserver.quoteUsable` (1179–1188)

- 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 1184 | `if now.UTC().Before(quote.FetchedAt) {` | CLOCK |
| return | | 1185 | `return false` |  |
| return | | 1187 | `return clock.LeaseElapsed(o.clk, quote.leaseAnchor) <= execgw.QueryPriceEvidenceDuration` | CLOCK |
| call | | 1183 | `o.clk.Now` | CLOCK |
| call | | 1184 | `Before` | CLOCK |
| call | | 1187 | `clock.LeaseElapsed` | CLOCK |

## `internal/app/engine/exitloop.go` `ExitObserver.record` (1288–1425)

- 분기 18 · 반환 6 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 1291 | `if !o.quoteUsable(quote) {` | CLOCK |
| branch:if | B2 | 1312 | `if judgement.ObservationSource == "" {` |  |
| branch:if | B3 | 1313 | `if quote.FetchedAt.IsZero() {` | CLOCK |
| branch:else | B4 | 1315 | `} else {` |  |
| branch:if | B5 | 1336 | `if orderable && m.floorExpectedState != nil {` |  |
| branch:else | B6 | 1341 | `} else if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` |  |
| branch:if | B7 | 1341 | `} else if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` |  |
| branch:if | B8 | 1342 | `if m.reJudge && !isProtective(proposal) {` |  |
| branch:else | B9 | 1364 | `} else {` |  |
| branch:if | B10 | 1366 | `if err != nil {` |  |
| branch:if | B11 | 1369 | `if !cleared.cleared {` |  |
| branch:else | B12 | 1376 | `} else {` |  |
| branch:if | B13 | 1384 | `if orderable {` |  |
| branch:if | B14 | 1386 | `if intentID == "" {` |  |
| branch:if | B15 | 1398 | `if err != nil {` |  |
| branch:if | B16 | 1399 | `if errors.Is(err, journal.ErrProposalPending) {` |  |
| branch:if | B17 | 1405 | `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` |  |
| branch:if | B18 | 1418 | `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` |  |
| return | | 1292 | `return nil` |  |
| return | | 1367 | `return err` |  |
| return | | 1403 | `return nil` |  |
| return | | 1416 | `return fmt.Errorf("engine: recording the exit judgement of %s: %w", m.position.ID, err)` |  |
| return | | 1419 | `return nil` |  |
| return | | 1423 | `return o.submit(ctx, m, proposal, quantity, intentID, judgement.ObservedPrice,` |  |
| call | | 1291 | `o.quoteUsable` | CLOCK |
| call | | 1313 | `quote.FetchedAt.IsZero` | CLOCK |

## `internal/app/engine/exitloop.go` `ExitObserver.refreshObservation` (1158–1177)

- 분기 3 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 1161 | `if !o.quoteUsable(quote) {` | CLOCK |
| branch:if | B2 | 1170 | `if errors.Is(err, journal.ErrExitObservationConflict) \|\| errors.Is(err, journal.ErrExitObservationStale) {` | CLOCK |
| branch:if | B3 | 1173 | `if err != nil {` |  |
| return | | 1162 | `return nil` |  |
| return | | 1171 | `return nil` |  |
| return | | 1174 | `return fmt.Errorf("engine: refreshing exit observation of %s: %w", m.position.ID, err)` |  |
| return | | 1176 | `return nil` |  |
| call | | 1161 | `o.quoteUsable` | CLOCK |

## `internal/app/engine/exit_unobserved.go` `ExitObserver.settleUnobserved` (139–204)

- 분기 12 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 144 | `if tally != nil && tally.owner != cycle {` |  |
| branch:if | B2 | 147 | `if tally == nil \|\| len(tally.held) == 0 {` |  |
| branch:range | B3 | 153 | `for id := range o.unobserved {` |  |
| branch:if | B4 | 154 | `if _, ok := tally.held[id]; !ok {` |  |
| branch:range | B5 | 160 | `for id := range tally.held {` |  |
| branch:range | B6 | 164 | `for _, id := range ids {` |  |
| branch:if | B7 | 167 | `if rec == nil {` |  |
| branch:if | B8 | 173 | `if judged, ok := tally.judged[id]; ok {` |  |
| branch:if | B9 | 174 | `if rec.streak != nil {` |  |
| branch:if | B10 | 187 | `if cause == "" {` |  |
| branch:if | B11 | 190 | `if rec.streak == nil {` |  |
| branch:if | B12 | 199 | `if elapsed < o.outageAfter() {` | CLOCK |
| return | | 150 | `return` |  |
| call | | 176 | `clock.LeaseElapsed` | CLOCK |
| call | | 176 | `clock.LeaseElapsed` | CLOCK |
| call | | 198 | `clock.LeaseElapsed` | CLOCK |
| call | | 199 | `o.outageAfter` | CLOCK |

## `internal/app/engine/position_policy_command.go` `PositionPolicyCommandService.prepare` (303–365)

- 분기 14 · 반환 10 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 307 | `if req.Action != positionpolicy.ActionRelease && req.Action != positionpolicy.ActionReadopt {` |  |
| branch:if | B2 | 311 | `if err != nil {` |  |
| branch:if | B3 | 314 | `if !state.ExternalLifecycleEligible() {` |  |
| branch:if | B4 | 317 | `if req.Action == positionpolicy.ActionRelease {` |  |
| branch:if | B5 | 320 | `if s.prices == nil {` |  |
| branch:if | B6 | 330 | `if s.retrier != nil {` |  |
| branch:else | B7 | 332 | `} else {` |  |
| branch:if | B8 | 335 | `if err != nil {` |  |
| branch:range | B9 | 339 | `for _, quote := range quotes {` |  |
| branch:if | B10 | 340 | `if strings.EqualFold(strings.TrimSpace(quote.Symbol), strings.TrimSpace(state.Symbol)) && quote.Last > 0 {` |  |
| branch:if | B11 | 345 | `if observed == "" {` |  |
| branch:if | B12 | 349 | `if err := exitpolicy.ValidateStopPct(stopPct); err != nil {` |  |
| branch:if | B13 | 353 | `if err != nil {` |  |
| branch:if | B14 | 357 | `if policyID == "" {` |  |
| return | | 308 | `return req, nil` |  |
| return | | 312 | `return positionpolicy.Request{}, err` |  |
| return | | 315 | `return positionpolicy.Request{}, positionpolicy.ErrIneligible` |  |
| return | | 318 | `return req, nil` |  |
| return | | 321 | `return positionpolicy.Request{}, errors.New("engine: authoritative price reader is unavailable")` |  |
| return | | 328 | `return readErr` |  |
| return | | 336 | `return positionpolicy.Request{}, fmt.Errorf("engine: observing re-adopt t0 for %s: %w", state.Symbol, err)` |  |
| return | | 346 | `return positionpolicy.Request{}, fmt.Errorf("engine: no positive authoritative price for %s", state.Symbol)` |  |
| return | | 354 | `return positionpolicy.Request{}, fmt.Errorf("engine: deriving re-adopt stop: %w", err)` |  |
| return | | 364 | `return req, nil` |  |
| call | | 305 | `s.clk.Now` | CLOCK |
| call | | 323 | `s.clk.Now` | CLOCK |

## `internal/app/engine/adoption.go` `ReconcileDriver.observeCandidates` (271–342)

- 분기 11 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 275 | `for _, c := range candidates {` |  |
| branch:if | B2 | 278 | `if symbol == "" \|\| market == "" {` |  |
| branch:if | B3 | 282 | `if markets == nil {` |  |
| branch:if | B4 | 307 | `if d.opts.Retrier != nil {` |  |
| branch:else | B5 | 309 | `} else {` |  |
| branch:if | B6 | 312 | `if err != nil {` |  |
| branch:range | B7 | 318 | `for _, q := range quotes {` |  |
| branch:if | B8 | 319 | `if q.Last <= 0 {` |  |
| branch:if | B9 | 326 | `if len(markets) != 1 {` |  |
| branch:range | B10 | 332 | `for candidateMarket := range markets {` |  |
| branch:if | B11 | 336 | `if !ok \|\| !strings.EqualFold(strings.TrimSpace(q.Currency), expectedCurrency) {` |  |
| return | | 304 | `return err` |  |
| return | | 313 | `return nil, time.Time{}, fmt.Errorf(` | CLOCK |
| return | | 341 | `return out, readAt, nil` |  |
| call | | 298 | `d.clk.Now` | CLOCK |

## `internal/app/engine/exitloop.go` `stableObservationID` (1202–1226)

- 분기 4 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 1205 | `if err != nil {` |  |
| branch:if | B2 | 1209 | `if strings.HasPrefix(quote.Source, "cycle:") \|\| quote.FetchedAt.IsZero() {` | CLOCK |
| branch:if | B3 | 1212 | `if sequence == "" {` |  |
| branch:range | B4 | 1217 | `for _, field := range []string{` |  |
| return | | 1206 | `return "", err` |  |
| return | | 1225 | `return "obs_" + fmt.Sprintf("%x", h.Sum(nil)), nil` |  |
| call | | 1208 | `quote.FetchedAt.UTC` | CLOCK |
| call | | 1209 | `quote.FetchedAt.IsZero` | CLOCK |

## `internal/app/engine/tracer.go` `Tracer.freshenGateInputs` (353–403)

- 분기 10 · 반환 12 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 356 | `if strings.EqualFold(strings.TrimSpace(t.opts.Params.Market), "us") {` |  |
| branch:if | B2 | 360 | `if err := t.opts.Retrier.Query(ctx, execgw.QueryOpenOrders, func(ctx context.Context) error {` |  |
| branch:if | B3 | 366 | `if err := t.opts.Retrier.Query(ctx, execgw.QueryBuyingPower, func(ctx context.Context) error {` |  |
| branch:if | B4 | 372 | `if err := t.opts.Retrier.Query(ctx, execgw.QueryHoldings, func(ctx context.Context) error {` |  |
| branch:if | B5 | 381 | `if err := t.opts.Retrier.Query(ctx, execgw.QueryPrice, func(ctx context.Context) error {` |  |
| branch:if | B6 | 383 | `if err != nil {` |  |
| branch:range | B7 | 386 | `for _, q := range quotes {` |  |
| branch:if | B8 | 387 | `if strings.EqualFold(strings.TrimSpace(q.Symbol), symbol) && q.Last > 0 {` |  |
| branch:if | B9 | 395 | `if last == "" {` |  |
| branch:if | B10 | 398 | `if age := t.opts.Clock.Now().Sub(asked); age > t.opts.Params.Freshness {` | CLOCK |
| return | | 362 | `return err` |  |
| return | | 364 | `return fmt.Errorf("%w: the open-order list could not be read: %v", ErrTracerRefused, err)` |  |
| return | | 368 | `return err` |  |
| return | | 370 | `return fmt.Errorf("%w: the buying power could not be read: %v", ErrTracerRefused, err)` |  |
| return | | 374 | `return err` |  |
| return | | 376 | `return fmt.Errorf("%w: the holdings could not be read: %v", ErrTracerRefused, err)` |  |
| return | | 384 | `return err` |  |
| return | | 391 | `return nil` |  |
| return | | 393 | `return fmt.Errorf("%w: the price of %s could not be read: %v", ErrTracerRefused, symbol, err)` |  |
| return | | 396 | `return fmt.Errorf("%w: the broker returned no last trade for %s", ErrTracerRefused, symbol)` |  |
| return | | 399 | `return fmt.Errorf("%w: the price of %s took %s to arrive, past the freshness bound of %s",` |  |
| return | | 402 | `return nil` |  |
| call | | 379 | `t.opts.Clock.Now` | CLOCK |
| call | | 398 | `t.opts.Clock.Now` | CLOCK |

## `internal/execgw/retry.go` `EntryGate.RecordSuccess` (530–540)

- 분기 1 · 반환 0 · go 0 · defer 1

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 536 | `if required && (!exists \|\| now.Sub(prior) > threshold) {` | CLOCK |
| call | | 533 | `g.clk.Now` | CLOCK |
| call | | 536 | `now.Sub` | CLOCK |

## `internal/flatten/liquidate.go` `Saga.aggressivePrice` (372–398)

- 분기 6 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 373 | `if s.Prices != nil {` |  |
| branch:if | B2 | 374 | `if limits, err := s.Prices.PriceLimits(ctx, symbol); err == nil && limits.LowerLimit > 0 {` |  |
| branch:if | B3 | 381 | `if err != nil {` |  |
| branch:range | B4 | 384 | `for _, q := range quotes {` |  |
| branch:if | B5 | 385 | `if !strings.EqualFold(strings.TrimSpace(q.Symbol), symbol) \|\| q.Last <= 0 {` |  |
| branch:if | B6 | 389 | `if discount <= 0 {` |  |
| return | | 377 | `return limits.LowerLimit, "exchange lower limit", nil` |  |
| return | | 382 | `return 0, "", fmt.Errorf("no price for %s, so no limit can be set: %w", symbol, err)` |  |
| return | | 394 | `return roundDownToTick(q.Last*(1-discount), market), "last trade less the aggressive discount", nil` |  |
| return | | 397 | `return 0, "", fmt.Errorf("no usable price for %s; a liquidation order cannot be priced", symbol)` |  |

## `internal/hybrid/client.go` `Client.GetQuote` (117–130)

- 분기 2 · 반환 5 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 121 | `if err != nil {` |  |
| branch:if | B2 | 124 | `if len(qs) == 0 {` |  |
| return | | 118 | `return route(c,` |  |
| return | | 122 | `return domain.Quote{}, err` |  |
| return | | 125 | `return domain.Quote{}, fmt.Errorf("hybrid: no quote found for %s", symbol)` |  |
| return | | 127 | `return qs[0], nil` |  |
| return | | 129 | `func() (domain.Quote, error) { return c.Client.GetQuote(ctx, symbol) })` |  |

## `internal/official/market_reads.go` `adaptOrderbook` (229–261)

- 분기 2 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 232 | `for _, a := range raw.Asks {` |  |
| branch:range | B2 | 243 | `for _, b := range raw.Bids {` |  |
| return | | 252 | `return domain.OrderBook{` |  |
| call | | 258 | `time.Now` | CLOCK |

## `internal/official/market_reads.go` `adaptPrices` (168–180)

- 분기 1 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 170 | `for _, p := range raw {` |  |
| return | | 179 | `return out` |  |
| call | | 175 | `time.Now` | CLOCK |

## `internal/official/market_reads.go` `Client.Prices` (142–150)

- 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 146 | `if err := c.get(ctx, "/api/v1/prices", q, &raw); err != nil {` |  |
| return | | 147 | `return nil, err` |  |
| return | | 149 | `return adaptPrices(raw), nil` |  |

## `internal/officialbars/quote.go` `PollQuoteL1` (98–188)

- 분기 17 · 반환 13 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 100 | `if reader == nil \|\| store == nil {` |  |
| branch:if | B2 | 104 | `if err != nil {` |  |
| branch:if | B3 | 108 | `if err := checkSymbol(in.Symbol); err != nil {` |  |
| branch:if | B4 | 114 | `if err != nil {` |  |
| branch:if | B5 | 118 | `if err != nil {` |  |
| branch:if | B6 | 123 | `if err := checkQuoteIdentity(code, in.Symbol, top.Market, top.Symbol, "top of book"); err != nil {` |  |
| branch:if | B7 | 126 | `if err := checkQuoteIdentity(code, in.Symbol, last.Market, last.Symbol, "last price"); err != nil {` |  |
| branch:if | B8 | 129 | `if top.Currency != last.Currency {` |  |
| branch:range | B9 | 133 | `for _, instant := range []struct {` |  |
| branch:if | B10 | 142 | `if instant.value.IsZero() {` | CLOCK |
| branch:if | B11 | 151 | `if err != nil {` |  |
| branch:if | B12 | 163 | `if err != nil {` |  |
| branch:switch | B13 | 175 | `switch {` |  |
| branch:case | B14 | 176 | `case errors.Is(err, strategyevidence.ErrRevisionConflict):` |  |
| branch:case | B15 | 179 | `case err != nil:` |  |
| branch:case | B16 | 182 | `case appended.Idempotent:` |  |
| branch:case | B17 | 184 | `default:` |  |
| return | | 101 | `return result, refuse(RefusalDependencyMissing, "a reader and a store are required", nil)` |  |
| return | | 105 | `return result, refuse(RefusalMarketInvalid, "market "+strconv.Quote(string(in.Market))+" is not kr or us", nil)` |  |
| return | | 109 | `return result, err` |  |
| return | | 115 | `return result, refuse(RefusalQuoteReaderError, "reading the top of book", err)` |  |
| return | | 119 | `return result, refuse(RefusalQuoteReaderError, "reading the last price", err)` |  |
| return | | 124 | `return result, err` |  |
| return | | 127 | `return result, err` |  |
| return | | 130 | `return result, refuse(RefusalQuoteCurrency,` |  |
| return | | 143 | `return result, refuse(RefusalQuoteInstant, instant.name+" is missing", nil)` |  |
| return | | 152 | `return result, err` |  |
| return | | 164 | `return result, refuse(RefusalQuoteEnvelope, "the quote is not evidence", err)` |  |
| return | | 180 | `return result, refuse(RefusalQuoteStoreError,` |  |
| return | | 187 | `return result, nil` |  |
| call | | 142 | `instant.value.IsZero` | CLOCK |

합계: AST 22 · 분기 162 · 반환 109
