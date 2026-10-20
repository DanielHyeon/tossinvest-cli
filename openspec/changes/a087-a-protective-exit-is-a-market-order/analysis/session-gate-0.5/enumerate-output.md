### `ExitObserver.checkOutage` (internal/app/engine/exitloop.go:916-958)

분기 6 · 반환 4 · 호출 21

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :918 | `if since.IsZero() {` | CLOCK |
| 분기 B2 (if) | :921 | `if o.clk.Now().Sub(since) < o.outageAfter() {` | CLOCK |
| 분기 B3 (if) | :924 | `if o.outageRaised {` | - |
| 분기 B4 (if) | :942 | `if o.opts.Escalate == nil \|\| strings.TrimSpace(o.opts.AccountRef) == "" {` | - |
| 분기 B5 (if) | :947 | `if err != nil {` | - |
| 분기 B6 (if) | :948 | `if !errors.Is(err, journal.ErrModeAnnouncementFailed) {` | - |
| 호출 | :918 | `since.IsZero` | CLOCK |
| 호출 | :921 | `o.clk.Now` | CLOCK |
| 호출 | :921 | `o.outageAfter` | CLOCK |
| 호출 | :932 | `o.clk.Now` | CLOCK |
| 호출 | :938 | `o.clk.Now` | CLOCK |

### `ExitObserver.judge` (internal/app/engine/exitloop.go:961-1004)

분기 9 · 반환 7 · 호출 8

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :963 | `if !o.quoteUsable(quote) {` | CLOCK |
| 분기 B2 (if) | :970 | `if o.noteHeldProposal(ctx, m) {` | - |
| 분기 B3 (if) | :973 | `if m.identityErr != nil {` | - |
| 분기 B4 (if) | :977 | `if m.reJudge {` | - |
| 분기 B5 (if) | :987 | `if err := o.opts.Journal.StampExitSnapshotQuarantineSelector(ctx,` | - |
| 분기 B6 (if) | :993 | `if err != nil {` | - |
| 분기 B7 (switch) | :998 | `switch m.state.PolicyKind {` | - |
| 분기 B8 (case) | :999 | `case journal.ExitPolicyLadder:` | - |
| 분기 B9 (case) | :1001 | `default:` | - |
| 호출 | :963 | `o.quoteUsable` | CLOCK |

### `ExitObserver.judgeLadder` (internal/app/engine/exitloop.go:1079-1146)

분기 8 · 반환 7 · 호출 19

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :1082 | `if err != nil {` | - |
| 분기 B2 (if) | :1086 | `if err := o.checkLadderPolicyStillFits(m, ladder); err != nil {` | - |
| 분기 B3 (if) | :1091 | `if err != nil {` | - |
| 분기 B4 (if) | :1096 | `if m.state.PendingLevel != "" {` | - |
| 분기 B5 (if) | :1097 | `if idx, err := exitpolicy.RungIndex(m.state.PendingLevel); err == nil {` | - |
| 분기 B6 (if) | :1125 | `if err != nil {` | - |
| 분기 B7 (if) | :1138 | `if !o.quoteUsable(quote) {` | CLOCK |
| 분기 B8 (if) | :1141 | `if !o.requiresFullJudgement(m, snapshot) {` | - |
| 호출 | :1138 | `o.quoteUsable` | CLOCK |

### `ExitObserver.judgeRatchet` (internal/app/engine/exitloop.go:1025-1077)

분기 6 · 반환 6 · 호출 19

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :1028 | `if err != nil \|\| !samePolicyIdentity(m.policyIdentity, activeIdentity) {` | - |
| 분기 B2 (if) | :1029 | `if err == nil {` | - |
| 분기 B3 (if) | :1039 | `if err != nil {` | - |
| 분기 B4 (if) | :1059 | `if err != nil {` | - |
| 분기 B5 (if) | :1067 | `if !o.quoteUsable(quote) {` | CLOCK |
| 분기 B6 (if) | :1070 | `if !o.requiresFullJudgement(m, snapshot) {` | - |
| 호출 | :1067 | `o.quoteUsable` | CLOCK |

### `ExitObserver.observe` (internal/app/engine/exitloop.go:842-913)

분기 15 · 반환 6 · 호출 28

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :849 | `if qerr != nil {` | - |
| 분기 B2 (range) | :858 | `for _, symbol := range symbols {` | - |
| 분기 B3 (range) | :862 | `for _, q := range quotes {` | - |
| 분기 B4 (if) | :864 | `if _, needed := wanted[symbol]; !needed \|\| q.Last <= 0 \|\| math.IsNaN(q.Last) \|\| math.IsInf(q.Last, 0) {` | - |
| 분기 B5 (if) | :868 | `if q.FetchedAt.IsZero() {` | CLOCK |
| 분기 B6 (else) | :871 | `} else if effectiveAt.After(validatedAt) \|\|` | CLOCK |
| 분기 B7 (if) | :871 | `} else if effectiveAt.After(validatedAt) \|\|` | CLOCK |
| 분기 B8 (if) | :878 | `if len(next) == 0 {` | - |
| 분기 B9 (if) | :884 | `if err != nil {` | - |
| 분기 B10 (if) | :887 | `if needsFallback {` | - |
| 분기 B11 (if) | :893 | `if err != nil {` | - |
| 분기 B12 (for) | :896 | `for {` | - |
| 분기 B13 (if) | :898 | `if current >= maximum \|\| o.observationSequence.CompareAndSwap(current, maximum) {` | - |
| 분기 B14 (range) | :904 | `for symbol, quote := range out {` | - |
| 분기 B15 (if) | :905 | `if quote.Source != "" {` | - |
| 호출 | :855 | `clock.LeaseAnchor` | CLOCK |
| 호출 | :867 | `q.FetchedAt.UTC` | CLOCK |
| 호출 | :868 | `q.FetchedAt.IsZero` | CLOCK |
| 호출 | :871 | `effectiveAt.After` | CLOCK |
| 호출 | :872 | `validatedAt.Sub` | CLOCK |

### `ExitObserver.ObserveOnce` (internal/app/engine/exitloop.go:457-526)

분기 8 · 반환 5 · 호출 23

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :465 | `if o.opts.SLO != nil && o.opts.SLO.FillDetectionBehind() {` | CLOCK |
| 분기 B2 (if) | :475 | `if err != nil {` | - |
| 분기 B3 (if) | :479 | `if len(states) == 0 {` | - |
| 분기 B4 (if) | :492 | `if err != nil {` | - |
| 분기 B5 (range) | :501 | `for _, state := range states {` | - |
| 분기 B6 (if) | :503 | `if !ok {` | - |
| 분기 B7 (if) | :511 | `if !o.quoteUsable(quote) {` | CLOCK |
| 분기 B8 (if) | :519 | `if err := o.judge(ctx, state, quote, observation, &cycle); err != nil && cycle.Err == nil {` | - |
| 호출 | :463 | `o.clk.Now` | CLOCK |
| 호출 | :465 | `o.opts.SLO.FillDetectionBehind` | CLOCK |
| 호출 | :487 | `o.clk.Now` | CLOCK |
| 호출 | :497 | `o.clk.Now` | CLOCK |
| 호출 | :511 | `o.quoteUsable` | CLOCK |

### `ExitObserver.quoteUsable` (internal/app/engine/exitloop.go:1179-1188)

분기 1 · 반환 2 · 호출 4

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :1184 | `if now.UTC().Before(quote.FetchedAt) {` | CLOCK |
| 호출 | :1183 | `o.clk.Now` | CLOCK |
| 호출 | :1184 | `Before` | CLOCK |
| 호출 | :1187 | `clock.LeaseElapsed` | CLOCK |

### `ExitObserver.record` (internal/app/engine/exitloop.go:1288-1425)

분기 18 · 반환 6 · 호출 24

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :1291 | `if !o.quoteUsable(quote) {` | CLOCK |
| 분기 B2 (if) | :1312 | `if judgement.ObservationSource == "" {` | - |
| 분기 B3 (if) | :1313 | `if quote.FetchedAt.IsZero() {` | CLOCK |
| 분기 B4 (else) | :1315 | `} else {` | - |
| 분기 B5 (if) | :1336 | `if orderable && m.floorExpectedState != nil {` | - |
| 분기 B6 (else) | :1341 | `} else if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | - |
| 분기 B7 (if) | :1341 | `} else if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | - |
| 분기 B8 (if) | :1342 | `if m.reJudge && !isProtective(proposal) {` | - |
| 분기 B9 (else) | :1364 | `} else {` | - |
| 분기 B10 (if) | :1366 | `if err != nil {` | - |
| 분기 B11 (if) | :1369 | `if !cleared.cleared {` | - |
| 분기 B12 (else) | :1376 | `} else {` | - |
| 분기 B13 (if) | :1384 | `if orderable {` | - |
| 분기 B14 (if) | :1386 | `if intentID == "" {` | - |
| 분기 B15 (if) | :1398 | `if err != nil {` | - |
| 분기 B16 (if) | :1399 | `if errors.Is(err, journal.ErrProposalPending) {` | - |
| 분기 B17 (if) | :1405 | `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | - |
| 분기 B18 (if) | :1418 | `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | - |
| 호출 | :1291 | `o.quoteUsable` | CLOCK |
| 호출 | :1313 | `quote.FetchedAt.IsZero` | CLOCK |

### `ExitObserver.Run` (internal/app/engine/exitloop.go:395-405)

분기 3 · 반환 2 · 호출 5

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (for) | :396 | `for {` | - |
| 분기 B2 (if) | :397 | `if err := ctx.Err(); err != nil {` | - |
| 분기 B3 (if) | :401 | `if err := o.clk.Sleep(ctx, o.Interval()); err != nil {` | CLOCK |
| 호출 | :401 | `o.clk.Sleep` | CLOCK |
| 호출 | :401 | `o.Interval` | CLOCK |

### `ExitObserver.workingSet` (internal/app/engine/exitloop.go:550-710)

분기 31 · 반환 3 · 호출 38

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :552 | `if err != nil {` | - |
| 분기 B2 (if) | :556 | `if err != nil {` | - |
| 분기 B3 (range) | :560 | `for _, result := range stateResults {` | - |
| 분기 B4 (range) | :565 | `for _, p := range positions {` | - |
| 분기 B5 (if) | :566 | `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | - |
| 분기 B6 (if) | :569 | `if !p.ExitEligible() {` | - |
| 분기 B7 (if) | :583 | `if !p.Adopted() {` | - |
| 분기 B8 (if) | :586 | `if err == nil && !protection.Legacy && ok && result.Corruption == nil && !strategyProtectionMatchesExit(protection, p, state) {` | - |
| 분기 B9 (if) | :589 | `if err != nil {` | - |
| 분기 B10 (if) | :592 | `if campaign != "" {` | - |
| 분기 B11 (if) | :593 | `if latchErr := o.opts.Journal.ReconcileStrategyProtection(ctx, campaign, p); latchErr != nil {` | - |
| 분기 B12 (if) | :598 | `if cycle.Err == nil {` | - |
| 분기 B13 (if) | :602 | `if !ok && protection.Policy == nil {` | - |
| 분기 B14 (if) | :609 | `if !ok {` | - |
| 분기 B15 (if) | :611 | `if err != nil {` | - |
| 분기 B16 (if) | :612 | `if cycle.Err == nil {` | - |
| 분기 B17 (if) | :617 | `if opened.PositionID == "" {` | - |
| 분기 B18 (if) | :628 | `if result.Corruption != nil {` | - |
| 분기 B19 (if) | :631 | `if qerr != nil {` | - |
| 분기 B20 (if) | :632 | `if cycle.Err == nil {` | - |
| 분기 B21 (if) | :642 | `if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); qerr != nil {` | - |
| 분기 B22 (else) | :647 | `} else if active && !q.NeedsReJudgement() {` | - |
| 분기 B23 (if) | :643 | `if cycle.Err == nil {` | - |
| 분기 B24 (if) | :647 | `} else if active && !q.NeedsReJudgement() {` | - |
| 분기 B25 (else) | :651 | `} else if active {` | - |
| 분기 B26 (if) | :651 | `} else if active {` | - |
| 분기 B27 (if) | :675 | `if identityErr != nil {` | - |
| 분기 B28 (if) | :678 | `if qerr != nil {` | - |
| 분기 B29 (if) | :679 | `if cycle.Err == nil {` | - |
| 분기 B30 (if) | :693 | `if protection.Policy != nil && protection.PositionGeneration == p.InstanceSeq {` | - |
| 분기 B31 (if) | :694 | `if cmp, compareErr := riskcalc.CompareDecimal(protection.Risk.StopPrice, state.Baseline); compareErr == nil && cmp > 0 {` | - |

### `adaptPrices` (internal/official/market_reads.go:168-180)

분기 1 · 반환 1 · 호출 6

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (range) | :170 | `for _, p := range raw {` | - |
| 호출 | :175 | `time.Now` | CLOCK |

### `Client.Prices` (internal/official/market_reads.go:142-150)

분기 1 · 반환 2 · 호출 4

| 종류 | 좌표 | 원문 | 분류 |
|---|---|---|---|
| 분기 B1 (if) | :146 | `if err := c.get(ctx, "/api/v1/prices", q, &raw); err != nil {` | - |

합계: SESSION 0 · CLOCK 40 (AST 12개)
