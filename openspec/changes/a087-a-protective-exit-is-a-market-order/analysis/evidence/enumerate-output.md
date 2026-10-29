## Client.GetPriceLimits — internal/client/marketdata.go:91-115 (fresh, sha 57bf3ef2c147)
    B1 if      :93    if err != nil {
     R return  :94    return domain.PriceLimits{}, err
    B2 if      :96    if deriveSecurityType(productCode) != "kr-s" {
     R return  :97    return domain.PriceLimits{}, fmt.Errorf("상/하한가는 국내(KRX) 종목만 제공됩니다 (미국장은 일일 가격제한 제도가 없음): %s", symbol)
    B3 if      :103   if err := c.getJSON(ctx, endpoint, &envelope); err != nil {
     R return  :104   return domain.PriceLimits{}, err
     R return  :107   return domain.PriceLimits{
  branches=3 returns=4

## Market.InRegularSession — internal/clock/market.go:146-152 (fresh, sha 85264b6e9b4e)
    B1 if      :148   if err != nil {
     R return  :149   return false, err
     R return  :151   return s.Contains(t), nil
  branches=1 returns=2

## Market.Location — internal/clock/market.go:79-92 (fresh, sha 85264b6e9b4e)
    B1 if      :80    if _, ok := tzName[m]; !ok {
     R return  :81    return nil, fmt.Errorf("%w: %q", ErrUnknownMarket, string(m))
    B2 if      :84    if locErr != nil {
     R return  :85    return nil, locErr
    B3 if      :88    if !ok || loc == nil {
     R return  :89    return nil, fmt.Errorf("%w: %q", ErrUnknownMarket, string(m))
     R return  :91    return loc, nil
  branches=3 returns=4

## Market.RegularSession — internal/clock/market.go:122-143 (fresh, sha 85264b6e9b4e)
    B1 if      :124   if err != nil {
     R return  :125   return Session{}, err
    B2 if      :128   if !ok {
     R return  :129   return Session{}, fmt.Errorf("%w: %q", ErrUnknownMarket, string(m))
     R return  :136   return Session{
  branches=2 returns=3

## Session.Contains — internal/clock/market.go:113-118 (fresh, sha 85264b6e9b4e)
    B1 if      :114   if s.Weekend || s.Open.IsZero() || s.Close.IsZero() {
     R return  :115   return false
     R return  :117   return !t.Before(s.Open) && t.Before(s.Close)
  branches=1 returns=2

## Context.ExitObserver — internal/app/engine/exitwiring.go:319-364 (fresh, sha 6020a40b2030)
    B1 if      :320   if c == nil {
     R return  :321   return nil, ErrExitObserverUnavailable
    B2 if      :323   if !c.Automation.Verified {
     R return  :324   return nil, fmt.Errorf("%w: the automation gate is not verified", ErrExitObserverUnavailable)
    B3 if      :327   if !ok {
     R return  :328   return nil, fmt.Errorf("%w: the injected Guardian does not issue reductions",
    B4 if      :343   if opts.Names == nil {
    B5 if      :348   if c.Notifier != nil {
    B6 if      :360   if opts.Floor == nil {
     R return  :363   return NewExitObserver(opts)
  branches=6 returns=4

## Context.ManualBuyRefresher — internal/app/engine/manual_buy_refresh.go:62-81 (fresh, sha 4d6a361efa0c)
    B1 if      :64    if c != nil {
    B2 if      :65    if guardian, ok := c.Guardian.(*execgw.RiskGuardian); ok {
    B3 if      :68    if c.Official != nil {
    B4 if      :71    if c.Journal != nil {
    B5 if      :76    if r.clock == nil {
     R return  :80    return AuxiliaryExecutor{Name: manualBuyRefreshName, Run: r.Run, StopEvent: obs.EventManualBuyRefreshFailed}
  branches=5 returns=1

## ExitObserver.noteHeldProposal — internal/app/engine/exit_held_proposal.go:126-173 (fresh, sha bc7a564eb8b1)
    B1 if      :127   if !m.state.Pending() {
     R return  :128   return false
    B2 if      :130   if strings.TrimSpace(m.state.PendingIntentID) == "" {
     R return  :133   return false
    B3 if      :136   if err != nil {
     R return  :138   return false
    B4 if      :140   if facts.Verdict == journal.ExitIntentUnaccepted {
    B5 if      :142   if err != nil {
     R return  :145   return released
    B6 range   :148   for _, rec := range facts.Parked {
    B7 range   :154   for _, order := range facts.Awaiting {
    B8 if      :155   if !strings.EqualFold(strings.TrimSpace(order.Side), "SELL") || strings.TrimSpace(order.OrderID) == "" {
    B9 if      :162   if err != nil {
     R return  :164   return false
   B10 range   :166   for _, order := range live {
   B11 if      :167   if !strings.EqualFold(strings.TrimSpace(order.Side), "SELL") || seen[order.OrderID] {
     R return  :172   return false
  branches=11 returns=6

## ExitObserver.observe — internal/app/engine/exitloop.go:860-931 (fresh, sha 4d92acb6e7e1)
    B1 if      :867   if qerr != nil {
     R return  :868   return qerr
    B2 range   :876   for _, symbol := range symbols {
    B3 range   :880   for _, q := range quotes {
    B4 if      :882   if _, needed := wanted[symbol]; !needed || q.Last <= 0 || math.IsNaN(q.Last) || math.IsInf(q.Last, 0) {
    B5 if      :886   if q.FetchedAt.IsZero() {
    B6 else    :889   } else if effectiveAt.After(validatedAt) ||
    B7 if      :889   } else if effectiveAt.After(validatedAt) ||
    B8 if      :896   if len(next) == 0 {
     R return  :897   return execgw.NewEvidenceInvalidError(fmt.Sprintf("no valid quote evidence for %d managed symbol(s)", len(symbols)))
     R return  :900   return nil
    B9 if      :902   if err != nil {
     R return  :903   return nil, err
   B10 if      :905   if needsFallback {
   B11 if      :911   if err != nil {
     R return  :912   return nil, err
   B12 for     :914   for {
   B13 if      :916   if current >= maximum || o.observationSequence.CompareAndSwap(current, maximum) {
   B14 range   :922   for symbol, quote := range out {
   B15 if      :923   if quote.Source != "" {
     R return  :930   return out, nil
  branches=15 returns=6

## ExitObserver.ObserveOnce — internal/app/engine/exitloop.go:469-541 (fresh, sha 4d92acb6e7e1)
    B1 if      :477   if o.opts.SLO != nil && o.opts.SLO.FillDetectionBehind() {
     R return  :483   return cycle
    B2 if      :487   if err != nil {
     R return  :489   return cycle
    B3 if      :491   if len(states) == 0 {
     R return  :501   return cycle
    B4 if      :504   if err != nil {
     R return  :507   return cycle
    B5 range   :513   for _, state := range states {
    B6 if      :515   if !ok {
    B7 if      :523   if !o.quoteUsable(quote) {
    B8 if      :531   if err := o.judge(ctx, state, quote, observation, &cycle); err != nil && cycle.Err == nil {
     R return  :540   return cycle
  branches=8 returns=5

## ExitObserver.release — internal/app/engine/exitloop.go:1603-1608 (fresh, sha 4d92acb6e7e1)
    B1 if      :1604  if _, _, err := o.opts.Journal.ReleaseUnacceptedExitProposal(ctx, m.position.ID, intentID, how); err != nil {
     R return  :1605  return fmt.Errorf("engine: releasing the exit proposal of %s: %w", m.position.ID, err)
     R return  :1607  return nil
  branches=1 returns=2

## ExitObserver.Run — internal/app/engine/exitloop.go:407-417 (fresh, sha 4d92acb6e7e1)
    B1 for     :408   for {
    B2 if      :409   if err := ctx.Err(); err != nil {
     R return  :410   return err
    B3 if      :413   if err := o.clk.Sleep(ctx, o.Interval()); err != nil {
     R return  :414   return err
  branches=3 returns=2

## ExitObserver.sellIntent — internal/app/engine/exitloop.go:1806-1832 (fresh, sha 4d92acb6e7e1)
    B1 if      :1808  if price == "" {
    B2 if      :1811  if price == "" {
     R return  :1812  return orderintent.PlaceIntent{}, fmt.Errorf(
    B3 if      :1816  if err != nil {
     R return  :1817  return orderintent.PlaceIntent{}, err
    B4 if      :1820  if err != nil {
     R return  :1821  return orderintent.PlaceIntent{}, err
     R return  :1823  return orderintent.PlaceIntent{
  branches=4 returns=4

## ExitObserver.submit — internal/app/engine/exitloop.go:1483-1596 (fresh, sha 4d92acb6e7e1)
    B1 if      :1487  if err != nil {
     R return  :1491  return err
    B2 if      :1494  if isZeroQuantity(submitQuantity) {
     R return  :1501  return releaseErr
    B3 if      :1516  if err != nil {
     R return  :1520  return releaseErr
    B4 if      :1527  if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {
     R return  :1529  return fmt.Errorf("engine: attaching the exit intent of %s: %w", m.position.ID, err)
    B5 if      :1533  if err != nil {
     R return  :1537  return releaseErr
    B6 switch  :1547  switch {
    B7 case    :1548  case out.State == journal.StateConfirmed:
     R return  :1556  return nil
    B8 case    :1557  case out.State == journal.StateInDoubt || out.State == journal.StateUnresolvedInDoubt:
    B9 if      :1562  if out.State == journal.StateUnresolvedInDoubt {
     R return  :1566  return nil
   B10 case    :1567  case out.Reason == execgw.ReasonSymbolInFlight:
     R return  :1571  return releaseErr
   B11 case    :1572  case out.AttemptID != "" && out.State != journal.StateNotDispatched && out.State != journal.StateFailedConfirmed:
   B12 if      :1576  if err == nil {
     R return  :1580  return fmt.Errorf("engine: the exit order of %s did not settle as unaccepted; the proposal stays armed: %w",
   B13 case    :1582  default:
   B14 if      :1584  if detail == "" && err != nil {
   B15 if      :1589  if out.State == journal.StateFailedConfirmed {
     R return  :1594  return releaseErr
  branches=15 returns=10

## exitOrderLine.args — internal/app/engine/exit_order_log.go:140-169 (fresh, sha 4d4ef9eb1630)
    B1 if      :161   if l.fill {
     R return  :168   return args
  branches=1 returns=1

## exitOrderLineFor — internal/app/engine/exit_order_log.go:102-114 (fresh, sha 4d4ef9eb1630)
    B1 if      :104   if price == "" {
     R return  :107   return exitOrderLine{
  branches=1 returns=1

## floatOf — internal/app/engine/exitloop.go:2014-2020 (fresh, sha 4d92acb6e7e1)
    B1 if      :2016  if err != nil {
     R return  :2017  return 0, fmt.Errorf("engine: %s %q is not a number: %w", field, s, err)
     R return  :2019  return v, nil
  branches=1 returns=2

## isProtective — internal/app/engine/exitloop.go:1463-1465 (fresh, sha 4d92acb6e7e1)
     R return  :1464  return p.Action == exitpolicy.ActionBaselineBreach || p.Action == exitpolicy.ActionLadderStop
  branches=0 returns=1

## manualBuyRefresh.Run — internal/app/engine/manual_buy_refresh.go:84-99 (fresh, sha 4d6a361efa0c)
    B1 if      :85    if r.board == nil || r.reads == nil || r.orders == nil {
     R return  :87    return ctx.Err()
    B2 for     :91    for {
    B3 select  :93    select {
     R return  :95    return ctx.Err()
  branches=3 returns=2

## NewContext — internal/app/engine/engine.go:437-607 (fresh, sha 60e3a8d45487)
    B1 if      :439   if err != nil {
     R return  :440   return nil, err
    B2 if      :446   if err != nil {
     R return  :447   return nil, err
    B3 if      :450   if clk == nil {
    B4 if      :464   if opts.Publisher != nil {
    B5 if      :472   if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications,
     R return  :477   return nil, fmt.Errorf("engine: recording the automation-gate audit trail: %w", err)
    B6 if      :487   if err != nil {
     R return  :488   return nil, refuseStartup(auditLog, gate, fmt.Errorf("%w: %w", ErrAccountUnresolved, err))
    B7 if      :494   if err != nil {
     R return  :495   return nil, refuseStartup(auditLog, gate, err)
    B8 if      :503   if err := bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)); err != nil {
     R return  :505   return nil, refuseStartup(auditLog, gate, err)
    B9 if      :522   if err != nil {
     R return  :524   return nil, refuseStartup(auditLog, gate, err)
   B10 if      :537   if gate.Enabled && guardian == nil && !opts.disableProductionGuardian {
   B11 if      :539   if factory == nil {
   B12 if      :545   if err != nil {
     R return  :547   return nil, refuseStartup(auditLog, gate,
   B13 if      :565   if err != nil {
     R return  :567   return nil, err
   B14 if      :569   if !automation.Verified {
   B15 if      :574   if err != nil {
     R return  :576   return nil, fmt.Errorf("engine: initialize strategy runtime projection: %w", err)
     R return  :578   return &Context{
  branches=15 returns=11

## NewOrderPath — internal/app/engine/engine.go:359-406 (fresh, sha 60e3a8d45487)
    B1 if      :361   if getenv == nil {
    B2 if      :366   if err != nil {
     R return  :367   return nil, err
    B3 if      :371   if err != nil {
     R return  :372   return nil, err
    B4 if      :376   if err != nil {
     R return  :377   return nil, err
    B5 if      :380   if err != nil {
     R return  :381   return nil, fmt.Errorf("loading official credentials: %w", err)
    B6 if      :386   if creds == nil || creds.APIKey == "" || creds.SecretKey == "" {
     R return  :387   return nil, ErrOfficialCredentialsRequired
     R return  :396   return &OrderPath{
  branches=6 returns=6

## strategyScheduleAuthorityLoader.prepare — internal/app/engine/strategy_schedule_authority.go:186-233 (fresh, sha f81f9798d4e7)
    B1 if      :192   if name == "" || !filepath.IsAbs(loader.configDir) {
     R return  :194   return value
    B2 if      :197   if err != nil {
     R return  :199   return value
    B3 if      :202   if !desired.Enabled || !desired.AutoStart {
     R return  :203   return value
    B4 if      :206   if pinErr == nil {
    B5 if      :210   if loader.calendar == nil {
     R return  :212   return value
    B6 if      :216   if err != nil {
     R return  :218   return value
    B7 if      :221   if err != nil {
     R return  :223   return value
    B8 if      :226   if err != nil {
     R return  :228   return value
     R return  :232   return value
  branches=8 returns=8

## canonicalProtectionQuantity — internal/execgw/protection.go:112-123 (fresh, sha 71e4923e1301)
    B1 if      :115   if canonical == "" || strings.ContainsAny(canonical, ".eE-+") {
     R return  :116   return 0, false
    B2 if      :119   if err != nil || quantity == 0 || quantity > maximumExactlyRepresentableInteger || float64(quantity) != value {
     R return  :120   return 0, false
     R return  :122   return quantity, true
  branches=2 returns=3

## checkOrderShape — internal/execgw/failclosed.go:53-98 (fresh, sha fbe97b79db74)
    B1 if      :59    if market != "kr" && market != "us" {
     R return  :60    return reject(ReasonUnsupportedOrderType,
    B2 if      :63    if side != "buy" && side != "sell" {
     R return  :64    return reject(ReasonInvalidRequest, "side %q is neither buy nor sell", intent.Side)
    B3 if      :66    if currency != "KRW" && currency != "USD" {
     R return  :67    return reject(ReasonUnsupportedOrderType, "currency mode %q is not supported", intent.CurrencyMode)
    B4 if      :70    if intent.Fractional {
    B5 switch  :71    switch {
    B6 case    :72    case market != "us":
     R return  :73    return reject(ReasonUnsupportedOrderType, "fractional orders exist only on the US market")
    B7 case    :74    case orderType != "market":
     R return  :75    return reject(ReasonUnsupportedOrderType, "a fractional order is always a market order")
    B8 case    :76    case side == "buy" && intent.Amount <= 0:
     R return  :77    return reject(ReasonInvalidRequest, "a fractional buy needs a positive amount")
    B9 case    :78    case side == "sell" && intent.Quantity <= 0:
     R return  :79    return reject(ReasonInvalidRequest, "a fractional sell needs a positive quantity")
     R return  :81    return nil
   B10 if      :84    if orderType != "limit" {
     R return  :85    return reject(ReasonUnsupportedOrderType,
   B11 if      :88    if market == "kr" && currency != "KRW" {
     R return  :89    return reject(ReasonUnsupportedOrderType, "a KR order is priced in KRW, not %s", currency)
   B12 if      :91    if intent.Quantity <= 0 {
     R return  :92    return reject(ReasonInvalidRequest, "quantity %s is not positive", decimalString(intent.Quantity))
   B13 if      :94    if intent.Price <= 0 {
     R return  :95    return reject(ReasonInvalidRequest, "a limit order needs a positive price")
     R return  :97    return nil
  branches=13 returns=13

## Preflight.CheckPlace — internal/execgw/failclosed.go:40-45 (fresh, sha fbe97b79db74)
    B1 if      :41    if rejected := checkOrderShape(intent); rejected != nil {
     R return  :42    return rejected
     R return  :44    return p.checkFunding(ctx, intent)
  branches=1 returns=2

## Retrier.Query — internal/execgw/retry.go:335-389 (fresh, sha 1036f9ee8394)
    B1 if      :337   if policy.MaxAttempts <= 0 {
    B2 if      :341   if clk == nil {
    B3 for     :347   for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
    B4 switch  :350   switch ClassifyQueryError(err) {
    B5 case    :351   case ClassOK:
    B6 if      :352   if r.Gate != nil {
     R return  :355   return nil
    B7 case    :357   case ClassAuthFatal:
    B8 if      :361   if r.Gate != nil {
     R return  :365   return errors.Join(err, r.escalateCredentialFailure(ctx))
    B9 case    :367   case ClassPermanent, ClassCanceled:
     R return  :368   return err
   B10 case    :370   case ClassRateLimited:
   B11 if      :372   if !ok {
     R return  :373   return fmt.Errorf("%w (Retry-After beyond the %s cap)", err, policy.MaxRetryAfter)
   B12 if      :375   if !r.sleepWithin(ctx, clk, wait, deadline) {
     R return  :376   return err
   B13 case    :379   default: // ClassTransient
   B14 if      :380   if attempt == policy.MaxAttempts {
     R return  :381   return err
   B15 if      :383   if !r.sleepWithin(ctx, clk, policy.BackoffFor(attempt, policy.BaseBackoff), deadline) {
     R return  :384   return err
     R return  :388   return lastErr
  branches=15 returns=8

## Saga.aggressivePrice — internal/flatten/liquidate.go:372-398 (fresh, sha ef9ad2603edf)
    B1 if      :373   if s.Prices != nil {
    B2 if      :374   if limits, err := s.Prices.PriceLimits(ctx, symbol); err == nil && limits.LowerLimit > 0 {
     R return  :377   return limits.LowerLimit, "exchange lower limit", nil
    B3 if      :381   if err != nil {
     R return  :382   return 0, "", fmt.Errorf("no price for %s, so no limit can be set: %w", symbol, err)
    B4 range   :384   for _, q := range quotes {
    B5 if      :385   if !strings.EqualFold(strings.TrimSpace(q.Symbol), symbol) || q.Last <= 0 {
    B6 if      :389   if discount <= 0 {
     R return  :394   return roundDownToTick(q.Last*(1-discount), market), "last trade less the aggressive discount", nil
     R return  :397   return 0, "", fmt.Errorf("no usable price for %s; a liquidation order cannot be priced", symbol)
  branches=6 returns=4

## Attempt.transition — internal/journal/durability.go:548-674 (fresh, sha 29ec7a1849fa)
    B1 if      :554   if err != nil {
     R return  :555   return fmt.Errorf("journal: starting the transition transaction for %s: %w", a.id, err)
    B2 if      :560   if err := tx.QueryRowContext(ctx,
    B3 if      :562   if errors.Is(err, sql.ErrNoRows) {
     R return  :563   return fmt.Errorf("%w: %s", ErrAttemptNotFound, a.id)
     R return  :565   return fmt.Errorf("journal: reading attempt %s: %w", a.id, err)
    B4 if      :567   if err := checkTransitionAllowed(a.id, current, to, o.from); err != nil {
     R return  :568   return err
    B5 if      :570   if o.consumeNonce {
    B6 if      :571   if err := consumeDecisionNonce(ctx, tx, a.id, now); err != nil {
     R return  :572   return err
    B7 if      :575   if to == StateConfirmed && (a.kind == KindPlace || a.kind == KindAmend) {
    B8 if      :577   if orderID == "" {
    B9 if      :578   if err := tx.QueryRowContext(ctx,
     R return  :580   return fmt.Errorf("journal: reading confirmed order id for %s: %w", a.id, err)
   B10 if      :584   if err := tx.QueryRowContext(ctx, `
     R return  :600   return fmt.Errorf("journal: reading prior evidence for confirmed order %s: %w", orderID, err)
   B11 if      :604   if orderErr != nil {
     R return  :605   return fmt.Errorf("journal: ordering confirmed ownership of %s: %w", orderID, orderErr)
   B12 if      :611   if o.setDispatchStarted {
   B13 if      :615   if o.setSettled {
   B14 if      :619   if o.brokerOrderID != "" {
   B15 if      :623   if o.reasonCode != "" {
   B16 if      :627   if o.detail != "" {
   B17 if      :635   if err != nil {
     R return  :636   return fmt.Errorf("journal: updating attempt %s: %w", a.id, err)
   B18 if      :639   if err != nil {
     R return  :640   return fmt.Errorf("journal: updating attempt %s: %w", a.id, err)
   B19 if      :642   if affected != 1 {
     R return  :646   return &StateError{AttemptID: a.id, Want: o.from, Got: current, To: to}
   B20 if      :649   if _, err := tx.ExecContext(ctx,
     R return  :653   return fmt.Errorf("journal: recording the %s->%s transition for %s: %w", current, to, a.id, err)
   B21 if      :662   if releasesReservations(to) {
   B22 if      :663   if _, err := releaseReservationsForAttempt(ctx, tx, a.id, ReleaseReasonBrokerTerminal,
     R return  :665   return err
   B23 if      :668   if err := tx.Commit(); err != nil {
     R return  :669   return fmt.Errorf("journal: committing the %s->%s transition for %s: %w", current, to, a.id, err)
     R return  :673   return nil
  branches=23 returns=15

## adaptPriceLimits — internal/official/price_limits_reads.go:55-62 (fresh, sha c7331f5b21e8)
     R return  :56    return domain.PriceLimits{
  branches=0 returns=1

## buildOrderCreate — internal/official/orders_write.go:100-153 (fresh, sha d89e0bdbffdc)
    B1 if      :103   if intent.Fractional && intent.Side == "buy" {
    B2 if      :105   if intent.Amount <= 0 {
     R return  :106   return nil, fmt.Errorf("fractional buy requires a positive amount, got %v", intent.Amount)
     R return  :108   return orderCreateV1{
    B3 if      :126   if intent.Fractional && intent.Side == "sell" {
    B4 if      :128   if intent.Quantity <= 0 {
     R return  :129   return nil, fmt.Errorf("fractional sell requires a positive quantity, got %v", intent.Quantity)
     R return  :133   return v0, nil
    B5 if      :136   if intent.Quantity <= 0 {
     R return  :137   return nil, fmt.Errorf("order requires a positive quantity, got %v", intent.Quantity)
    B6 if      :144   if orderType == "LIMIT" {
    B7 if      :145   if intent.Price <= 0 {
     R return  :146   return nil, fmt.Errorf("limit order requires a positive price, got %v", intent.Price)
     R return  :152   return v0, nil
  branches=7 returns=7

## Client.doRequest — internal/official/client.go:191-207 (fresh, sha 6d145916a679)
    B1 if      :194   if err != nil {
     R return  :196   return 0, nil, fmt.Errorf("%w: %s", ErrTransport, err)
    B2 if      :201   if err != nil {
     R return  :203   return resp.StatusCode, nil, fmt.Errorf("%w: reading body: %s", ErrTransport, err)
     R return  :206   return resp.StatusCode, body, nil
  branches=2 returns=3

## Client.PriceLimits — internal/official/price_limits_reads.go:31-39 (fresh, sha c7331f5b21e8)
    B1 if      :35    if err := c.get(ctx, "/api/v1/price-limits", q, &raw); err != nil {
     R return  :36    return domain.PriceLimits{}, err
     R return  :38    return adaptPriceLimits(symbol, raw), nil
  branches=1 returns=2

## Client.TypedMarketCalendar — internal/official/typed_calendar_reads.go:37-54 (fresh, sha 5f97bf011626)
    B1 if      :39    if country != "KR" && country != "US" {
     R return  :40    return MarketCalendarResponse{}, fmt.Errorf("market calendar country must be KR or US, got %q", country)
    B2 if      :42    if err := validateMarketCalendarDate(date); err != nil {
     R return  :43    return MarketCalendarResponse{}, err
    B3 if      :46    if date != "" {
    B4 if      :50    if err := c.get(ctx, "/api/v1/market-calendar/"+country, query, &out); err != nil {
     R return  :51    return MarketCalendarResponse{}, err
     R return  :53    return out, nil
  branches=4 returns=4

## parseDecimal — internal/official/market_reads.go:15-24 (fresh, sha a2c2f01619e0)
    B1 if      :16    if s == "" {
     R return  :17    return 0
    B2 if      :20    if err != nil {
     R return  :21    return 0
     R return  :23    return f
  branches=2 returns=3

## adaptCalendarDay — internal/scheduler/calendar.go:103-134 (fresh, sha 004b8011b160)
    B1 if      :104   if day.Date == "" {
     R return  :105   return CalendarDay{}, nil
    B2 if      :107   if _, err := time.Parse("2006-01-02", day.Date); err != nil {
     R return  :108   return CalendarDay{}, fmt.Errorf("invalid date %q: %w", day.Date, err)
    B3 switch  :111   switch market {
    B4 case    :112   case marketclock.MarketKR:
    B5 if      :113   if day.Integrated != nil {
    B6 case    :116   case marketclock.MarketUS:
    B7 case    :118   default:
     R return  :119   return CalendarDay{}, fmt.Errorf("unsupported market %q", market)
    B8 if      :122   if session == nil {
     R return  :123   return out, nil
    B9 if      :125   if session.StartTime.IsZero() || session.EndTime.IsZero() || !session.StartTime.Before(session.EndTime) {
     R return  :126   return CalendarDay{}, fmt.Errorf("invalid regular session")
   B10 if      :128   if session.StartTime.In(loc).Format("2006-01-02") != day.Date || session.EndTime.In(loc).Format("2006-01-02") != day.Date {
     R return  :129   return CalendarDay{}, fmt.Errorf("regular session is outside exchange-local date %s", day.Date)
     R return  :133   return out, nil
  branches=10 returns=7

## AdaptOfficialCalendar — internal/scheduler/calendar.go:56-101 (fresh, sha 004b8011b160)
    B1 if      :57    if fetchedAt.IsZero() {
     R return  :58    return CalendarSnapshot{}, fmt.Errorf("calendar fetched-at is required")
    B2 if      :61    if err != nil {
     R return  :62    return CalendarSnapshot{}, err
    B3 if      :65    if err != nil {
     R return  :66    return CalendarSnapshot{}, fmt.Errorf("previous business day: %w", err)
    B4 if      :69    if err != nil {
     R return  :70    return CalendarSnapshot{}, fmt.Errorf("today: %w", err)
    B5 if      :73    if err != nil {
     R return  :74    return CalendarSnapshot{}, fmt.Errorf("next business day: %w", err)
    B6 if      :79    if previousErr != nil || todayErr != nil || nextErr != nil || !previousDate.Before(todayDate) || !todayDate.Before(nextDate) {
     R return  :80    return CalendarSnapshot{}, errors.New("calendar requires previous < today < next business-day evidence")
    B7 if      :82    if previous.Regular == nil || next.Regular == nil {
     R return  :83    return CalendarSnapshot{}, errors.New("previous and next business days require regular-session evidence")
    B8 if      :93    if err != nil {
     R return  :94    return CalendarSnapshot{}, fmt.Errorf("canonical calendar: %w", err)
     R return  :97    return CalendarSnapshot{
  branches=8 returns=9

## CalendarSnapshot.ValidityAt — internal/scheduler/calendar.go:138-149 (fresh, sha 004b8011b160)
    B1 if      :139   if c.FetchedAt.IsZero() || now.Before(c.FetchedAt) {
     R return  :140   return CalendarClockSkew
    B2 if      :142   if now.Sub(c.FetchedAt) >= calendarMaxAge {
     R return  :143   return CalendarStale
    B3 if      :145   if c.Today.Regular != nil && !now.Before(c.Today.Regular.Open) && now.Before(c.Today.Regular.Close) && !c.FetchedAt.Before(c.Today.Regular.Open) {
     R return  :146   return CalendarRefreshTooLate
     R return  :148   return CalendarValid
  branches=3 returns=4

## buildFlattenWiring — cmd/tossctl/flatten.go:199-272 (fresh, sha 3a2b20f4c145)
    B1 if      :201   if err != nil {
     R return  :202   return nil, err
    B2 if      :206   if err != nil {
     R return  :207   return nil, fmt.Errorf("flatten-all: the account could not be identified: %w", err)
    B3 range   :210   for _, a := range accounts {
    B4 if      :211   if strings.TrimSpace(a.DisplayName) != "" {
    B5 if      :216   if accountRef == "" {
     R return  :217   return nil, errors.New("flatten-all: the broker returned no account number")
    B6 if      :221   if root.configDir != "" {
    B7 if      :227   if err != nil {
     R return  :228   return nil, err
    B8 if      :241   if err != nil {
     R return  :243   return nil, err
     R return  :265   return &flattenWiring{
  branches=8 returns=6

## engineRuntime — cmd/tossctl/engine.go:655-752 (fresh, sha 753594c5667b)
    B1 if      :658   if err != nil {
     R return  :659   return nil, err
    B2 if      :667   if err != nil {
     R return  :668   return nil, err
    B3 if      :684   if err != nil {
     R return  :685   return nil, err
    B4 if      :689   if err != nil {
     R return  :690   return nil, err
    B5 if      :693   if err != nil {
     R return  :694   return nil, err
    B6 if      :705   if err != nil {
     R return  :706   return nil, err
     R return  :709   return engine.NewRuntime(engine.RuntimeOptions{
  branches=6 returns=7

## placeIntentSupported — internal/trading/service.go:269-288 (fresh, sha 631b2f02437e)
    B1 if      :270   if intent.Market != "us" && intent.Market != "kr" {
     R return  :271   return false
    B2 if      :273   if intent.Fractional {
     R return  :275   return intent.Market == "us" && intent.OrderType == "market" &&
    B3 if      :281   if intent.OrderType != "limit" {
     R return  :282   return false
    B4 if      :284   if intent.Market == "kr" {
     R return  :285   return intent.CurrencyMode == "KRW"
     R return  :287   return intent.CurrencyMode == "KRW" || intent.CurrencyMode == "USD"
  branches=4 returns=5

## Service.guard — internal/trading/service.go:237-251 (fresh, sha 631b2f02437e)
    B1 if      :238   if err := s.requireActionEnabled(action); err != nil {
     R return  :239   return err
    B2 if      :241   if !opts.Execute {
     R return  :242   return fmt.Errorf("%w; rerun with --execute after reviewing `tossctl order preview`", ErrExecuteRequired)
    B3 if      :244   if !s.policy.AllowLiveOrderActions {
     R return  :245   return ErrLiveActionsDisabled
    B4 if      :247   if subtle.ConstantTimeCompare([]byte(opts.Confirm), []byte(preview.ConfirmToken)) != 1 {
     R return  :248   return ErrConfirmMismatch
     R return  :250   return nil
  branches=4 returns=5

## Service.Place — internal/trading/service.go:186-208 (fresh, sha 631b2f02437e)
    B1 if      :188   if !placeIntentSupported(intent) {
     R return  :189   return MutationResult{}, ErrPlaceUnsupported
    B2 if      :192   if intent.Side == "sell" && !s.policy.Sell {
     R return  :193   return MutationResult{}, &DisabledActionError{Action: "sell"}
    B3 if      :196   if intent.Fractional && !s.policy.Fractional {
     R return  :197   return MutationResult{}, &DisabledActionError{Action: "fractional"}
    B4 if      :200   if err := s.guard(ctx, ActionPlace, s.PreviewPlace(intent), opts); err != nil {
     R return  :201   return MutationResult{}, err
    B5 if      :204   if s.broker == nil {
     R return  :205   return MutationResult{}, ErrLiveMutationPending
     R return  :207   return s.withLineage(s.broker.PlacePendingOrder(ctx, intent))
  branches=5 returns=6

## Service.PreviewPlace — internal/trading/service.go:109-143 (fresh, sha 631b2f02437e)
    B1 if      :116   if !s.policy.Place {
    B2 if      :119   if intent.Side == "sell" && !s.policy.Sell {
    B3 if      :122   if intent.Fractional && !s.policy.Fractional {
    B4 if      :125   if !s.policy.AllowLiveOrderActions {
    B5 if      :129   if intent.Side == "sell" {
    B6 if      :132   if intent.Fractional {
     R return  :135   return Preview{
  branches=6 returns=1

stale=0
