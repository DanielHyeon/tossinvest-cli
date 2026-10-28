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

## isProtective — internal/app/engine/exitloop.go:1463-1465 (fresh, sha 4d92acb6e7e1)
     R return  :1464  return p.Action == exitpolicy.ActionBaselineBreach || p.Action == exitpolicy.ActionLadderStop
  branches=0 returns=1

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
