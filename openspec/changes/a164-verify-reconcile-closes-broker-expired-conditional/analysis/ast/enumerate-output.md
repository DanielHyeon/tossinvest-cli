# a164 AST 열거 출력 (`python3 -I enumerate.py` 생성 — 손편집 금지)

## `internal/official/conditional_reads.go` `Client.ConditionalOrder` (214–220)

- 산출물 `official--client.conditionalorder.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 216 | `if err := c.getAcct(ctx, "/api/v1/conditional-orders/"+url.PathEscape(id), nil, &raw); err != nil {` |  |
| return | | 217 | `return domain.ConditionalOrder{}, err` |  |
| return | | 219 | `return adaptConditionalOrder(raw), nil` |  |

## `internal/official/protection_reads.go` `Client.ConditionalOrderRaw` (13–19)

- 산출물 `official--client.conditionalorderraw.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 15 | `if err := c.getAcct(ctx, "/api/v1/conditional-orders/"+url.PathEscape(id), nil, &o); err != nil {` |  |
| return | | 16 | `return RawConditionalOrder{}, err` |  |
| return | | 18 | `return adaptProtectionConditional(o), nil` |  |

## `internal/official/reconcile_reads.go` `Client.ReconcileConditionalOrdersPage` (192–201)

- 산출물 `official--client.reconcileconditionalorderspage.ast.json` · 분기 2 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 193 | `if strings.TrimSpace(status) == "" {` |  |
| branch:if | B2 | 197 | `if err := c.getAcct(ctx, "/api/v1/conditional-orders", reconcileListQuery(status, symbol, cursor, limit), &raw); err != nil {` |  |
| return | | 194 | `return ReconcileConditionalPage{}, fmt.Errorf("%w: the reconcile read names its group", ErrOrderStatusRequired)` | MARK |
| return | | 198 | `return ReconcileConditionalPage{}, err` |  |
| return | | 200 | `return DecodeReconcileConditionalPage(raw)` |  |

## `internal/official/reconcile_reads.go` `DecodeReconcileConditionalPage` (134–151)

- 산출물 `official--decodereconcileconditionalpage.ast.json` · 분기 3 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 136 | `if err != nil {` |  |
| branch:range | B2 | 140 | `for i, raw := range env.rows {` |  |
| branch:if | B3 | 142 | `if err := json.Unmarshal(raw, &o); err != nil {` |  |
| return | | 137 | `return ReconcileConditionalPage{}, err` |  |
| return | | 143 | `return ReconcileConditionalPage{}, reconcileSchema("conditionalOrders[%d] does not decode: %v", i, err)` |  |
| return | | 150 | `return page, nil` |  |

## `cmd/tossctl/verify_reconcile.go` `newVerifyReconcileCmd` (42–70)

- 산출물 `tossctl--newverifyreconcilecmd.ast.json` · 분기 0 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| return | | 63 | `return runVerifyReconcile(cmd, root, opts)` |  |
| return | | 69 | `return cmd` |  |

## `cmd/tossctl/verify_reconcile.go` `newVerifyReconcileReader` (167–194)

- 산출물 `tossctl--newverifyreconcilereader.ast.json` · 분기 5 · 반환 5 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 170 | `if err != nil {` |  |
| branch:range | B2 | 175 | `for _, a := range accounts {` |  |
| branch:if | B3 | 176 | `if strings.TrimSpace(a.DisplayName) == "" {` |  |
| branch:if | B4 | 181 | `if len(accounts) != 1 {` |  |
| branch:if | B5 | 186 | `if err != nil \|\| seq <= 0 {` |  |
| return | | 171 | `return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseAccountsRead,` | MARK |
| return | | 177 | `return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseAccountMalformedRow,` | MARK |
| return | | 182 | `return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseAccountCount,` | MARK |
| return | | 187 | `return nil, verifylive.ReconcileAccount{}, reconcileRefusal(verifylive.RefuseAccountSeqZero,` | MARK |
| return | | 192 | `return &reconcileReads{client: bound, ref: ref, seq: seq},` |  |

## `cmd/tossctl/verify_reconcile.go` `runVerifyReconcile` (73–117)

- 산출물 `tossctl--runverifyreconcile.ast.json` · 분기 6 · 반환 6 · go 0 · defer 2

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 75 | `if ctx == nil {` |  |
| branch:if | B2 | 81 | `if err != nil {` |  |
| branch:if | B3 | 85 | `if err != nil {` |  |
| branch:if | B4 | 90 | `if err != nil {` |  |
| branch:if | B5 | 96 | `if err != nil {` |  |
| branch:if | B6 | 111 | `if err != nil {` |  |
| return | | 82 | `return err` |  |
| return | | 86 | `return reconcileRefusal(verifylive.RefuseExecutionLock, err.Error())` | MARK |
| return | | 91 | `return reconcileRefusal(verifylive.RefuseRateBudget, err.Error())` | MARK |
| return | | 97 | `return err` |  |
| return | | 112 | `return err` |  |
| return | | 116 | `return nil` |  |

## `cmd/tossctl/verify.go` `validateM0TriggerMode` (406–432)

- 산출물 `tossctl--validatem0triggermode.ast.json` · 분기 9 · 반환 7 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 407 | `if !opts.includeTrigger {` |  |
| branch:if | B2 | 410 | `if !opts.confirmEach \|\| !opts.resume \|\| opts.includeTTLEdge \|\| len(opts.redo) != 1 \|\|` |  |
| branch:if | B3 | 415 | `if checkpoint, ok, err := verifylive.M0Unsettled(prior); err != nil {` | MARK |
| branch:else | B4 | 417 | `} else if ok && checkpoint.Kind != "pending-create" {` |  |
| branch:if | B5 | 417 | `} else if ok && checkpoint.Kind != "pending-create" {` |  |
| branch:else | B6 | 419 | `} else if !ok {` |  |
| branch:if | B7 | 419 | `} else if !ok {` |  |
| branch:if | B8 | 424 | `if err := verifylive.M0ExactPrerequisites(prior); err != nil {` | MARK |
| branch:if | B9 | 428 | `if len(verifylive.Outstanding(prior)) != 0 {` | MARK |
| return | | 408 | `return nil` |  |
| return | | 413 | `return fmt.Errorf("verify: --include-trigger requires --confirm-each --resume --redo conditional-trigger only and --trigger-receipt-dir")` |  |
| return | | 416 | `return err` |  |
| return | | 418 | `return fmt.Errorf("verify: M0 causal receipt HOLD: unsettled %s owner requires manual reconciliation before broker construction", checkpoint.Kind)` | MARK |
| return | | 425 | `return err` |  |
| return | | 429 | `return fmt.Errorf("verify: M0 trigger receipt HOLD: prior verification record has outstanding artifacts; inspect with `tossctl verify status` and reconcile manually before any cleanup")` | MARK |
| return | | 431 | `return nil` |  |
| call | | 415 | `verifylive.M0Unsettled` | MARK |
| call | | 424 | `verifylive.M0ExactPrerequisites` | MARK |
| call | | 428 | `verifylive.Outstanding` | MARK |

## `internal/verifylive/reconcile_record.go` `Artifact.MarshalJSON` (109–126)

- 산출물 `verifylive--artifact.marshaljson.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 111 | `if !a.ReconciledAbsent {` | MARK |
| return | | 112 | `return json.Marshal(plainArtifact(a))` |  |
| return | | 114 | `return json.Marshal(struct {` |  |

## `internal/verifylive/record.go` `Artifact.terminal` (736–736)

- 산출물 `verifylive--artifact.terminal.ast.json` · 분기 0 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| return | | 736 | `func (a Artifact) terminal() bool { return a.Cancelled \|\| a.Filled \|\| a.ReconciledAbsent }` | MARK |

## `internal/verifylive/record.go` `ChainOf` (679–689)

- 산출물 `verifylive--chainof.ast.json` · 분기 3 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 681 | `for _, e := range entries {` |  |
| branch:range | B2 | 682 | `for _, a := range e.Artifacts {` |  |
| branch:if | B3 | 683 | `if a.Kind == kind && a.ID == id && a.ChainID != "" {` | MARK |
| return | | 688 | `return chain` |  |

## `internal/verifylive/reconcile_check.go` `checkReconcileAbsence` (371–396)

- 산출물 `verifylive--checkreconcileabsence.ast.json` · 분기 7 · 반환 6 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 372 | `if len(r.condOpen) > 0 {` |  |
| branch:if | B2 | 376 | `if len(r.plainOpen) > 0 {` |  |
| branch:range | B3 | 380 | `for _, row := range r.closed {` |  |
| branch:if | B4 | 381 | `if row.ID != target.ID {` |  |
| branch:if | B5 | 384 | `if row.Status == "EXPIRED" {` | MARK |
| branch:range | B6 | 390 | `for _, row := range r.closed {` |  |
| branch:if | B7 | 391 | `if row.TriggeredOrderID != "" \|\| row.Status == "COMPLETED" {` | MARK |
| return | | 373 | `return refuse(RefuseOpenConditional,` | MARK |
| return | | 377 | `return refuse(RefuseOpenPlainOrder,` | MARK |
| return | | 385 | `return refuse(RefuseClosedTargetExpired, "the artifact is listed CLOSED as EXPIRED: "+reconcileQ6Message+` | MARK |
| return | | 388 | `return refuse(RefuseClosedTargetPresent, "the artifact's identifier is listed in the CLOSED group — it is not absent")` | MARK |
| return | | 392 | `return refuse(RefuseClosedFiredTrace, "the symbol's CLOSED group holds a fired conditional — a successor may have fired")` | MARK |
| return | | 395 | `return nil` |  |

## `internal/verifylive/reconcile_check.go` `checkReconcileClosedStatus` (399–407)

- 산출물 `verifylive--checkreconcileclosedstatus.ast.json` · 분기 2 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 400 | `for _, row := range r.closed {` |  |
| branch:if | B2 | 401 | `if row.Status == "" \|\| reconcileOpenFamily[row.Status] \|\| !reconcileClosedStatusAllowlist[row.Status] {` | MARK |
| return | | 402 | `return refuse(RefuseClosedStatus,` | MARK |
| return | | 406 | `return nil` |  |

## `internal/verifylive/reconcile_check.go` `checkReconcileRetention` (116–133)

- 산출물 `verifylive--checkreconcileretention.ast.json` · 분기 4 · 반환 5 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 118 | `if m == nil \|\| m.Bound <= 0 {` |  |
| branch:if | B2 | 121 | `if m.Eviction != evictionDuration {` |  |
| branch:if | B3 | 125 | `if target.CreatedAt.IsZero() {` |  |
| branch:if | B4 | 129 | `if age < 0 \|\| age > m.Bound {` |  |
| return | | 119 | `return refuse(RefuseRetentionUnmeasured, "the fired-conditional retention of the CLOSED list has not been measured (Q1)")` | MARK |
| return | | 122 | `return refuse(RefuseRetentionEviction,` | MARK |
| return | | 126 | `return refuse(RefuseCreatedAtZero, "the outstanding line carries no creation time, so its age cannot be measured")` | MARK |
| return | | 130 | `return refuse(RefuseArtifactTooOld, fmt.Sprintf("artifact age %s is outside the measured retention bound %s", age, m.Bound))` | MARK |
| return | | 132 | `return nil` |  |

## `internal/verifylive/reconcile_check.go` `checkReconcileRows` (334–367)

- 산출물 `verifylive--checkreconcilerows.ast.json` · 분기 12 · 반환 7 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 336 | `for _, row := range conditional {` |  |
| branch:if | B2 | 337 | `if row.HasSecond {` | MARK |
| branch:range | B3 | 341 | `for _, row := range conditional {` |  |
| branch:if | B4 | 342 | `if row.ID == "" \|\| row.Symbol == "" \|\| row.Market == "" {` |  |
| branch:range | B5 | 346 | `for _, row := range r.plainOpen {` |  |
| branch:if | B6 | 347 | `if row.ID == "" \|\| row.Symbol == "" {` |  |
| branch:range | B7 | 351 | `for _, row := range conditional {` |  |
| branch:if | B8 | 352 | `if row.Symbol != symbol {` |  |
| branch:range | B9 | 356 | `for _, row := range r.plainOpen {` |  |
| branch:if | B10 | 357 | `if row.Symbol != symbol {` |  |
| branch:range | B11 | 361 | `for _, row := range conditional {` |  |
| branch:if | B12 | 362 | `if row.Market != market {` |  |
| return | | 338 | `return refuse(RefuseOCO, "a row carries a second (OCO) leg — OCO conditionals are not reconciled through this path")` | MARK |
| return | | 343 | `return refuse(RefuseRowIncomplete, "a conditional row lacks its id, symbol or market")` | MARK |
| return | | 348 | `return refuse(RefuseRowIncomplete, "a plain order row lacks its id or symbol")` | MARK |
| return | | 353 | `return refuse(RefuseRowSymbol, "a conditional row carries another symbol — the symbol filter was not honoured")` | MARK |
| return | | 358 | `return refuse(RefuseRowSymbol, "a plain order row carries another symbol — the symbol filter was not honoured")` | MARK |
| return | | 363 | `return refuse(RefuseRowMarket, "a conditional row reports another market")` | MARK |
| return | | 366 | `return nil` |  |

## `internal/verifylive/reconcile_check.go` `checkReconcileWindow` (136–143)

- 산출물 `verifylive--checkreconcilewindow.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 138 | `if elapsed < 0 \|\| elapsed > reconcileFreshnessBound {` | MARK |
| return | | 139 | `return refuse(RefuseFreshnessExceeded,` | MARK |
| return | | 142 | `return checkReconcileRetention(target, now)` | MARK |
| call | | 142 | `checkReconcileRetention` | MARK |

## `internal/verifylive/cleanup.go` `cleanupFrom` (123–136)

- 산출물 `verifylive--cleanupfrom.ast.json` · 분기 3 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 125 | `for _, l := range outstandingLines(entries) {` |  |
| branch:if | B2 | 127 | `if gate == "" {` |  |
| branch:if | B3 | 131 | `if settled(gate) && heldAfter(entries, gate, l.at) {` |  |
| return | | 135 | `return out` |  |

## `internal/verifylive/reconcile_record.go` `encodeReconcileLine` (67–89)

- 산출물 `verifylive--encodereconcileline.ast.json` · 분기 2 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 69 | `if err != nil {` |  |
| branch:if | B2 | 85 | `if err != nil {` |  |
| return | | 70 | `return nil, err` |  |
| return | | 86 | `return nil, fmt.Errorf("verify reconcile: encoding the reconcile line: %w", err)` |  |
| return | | 88 | `return append(b, '\n'), nil` |  |

## `internal/verifylive/m0_recovery.go` `M0ExactPrerequisites` (98–108)

- 산출물 `verifylive--m0exactprerequisites.ast.json` · 분기 3 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 99 | `for _, step := range Steps() {` |  |
| branch:if | B2 | 100 | `if step.ID == StepConditionalTrigger \|\| step.ID == StepIdempotencyTTLEdge {` |  |
| branch:if | B3 | 103 | `if !Passed(entries, step.ID) {` | MARK |
| return | | 104 | `return fmt.Errorf("verify: M0 trigger mode requires prior PASS for %s before conditional-trigger redo", step.ID)` |  |
| return | | 107 | `return nil` |  |
| call | | 103 | `Passed` | MARK |

## `internal/verifylive/runner.go` `New` (320–425)

- 산출물 `verifylive--new.ast.json` · 분기 26 · 반환 14 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 321 | `if o.Broker == nil {` |  |
| branch:if | B2 | 324 | `if o.Recorder == nil {` |  |
| branch:if | B3 | 327 | `if o.Confirm == nil {` |  |
| branch:if | B4 | 330 | `if o.ConfirmBatch == nil {` |  |
| branch:if | B5 | 334 | `if strings.TrimSpace(o.AccountRef) == "" {` |  |
| branch:if | B6 | 338 | `if err != nil {` |  |
| branch:if | B7 | 341 | `if o.IncludeTrigger {` |  |
| branch:if | B8 | 342 | `if !o.ConfirmEach \|\| !o.Resume \|\| o.IncludeTTLEdge \|\| o.Receipt == nil \|\| !o.Receipt.usable() \|\|` |  |
| branch:if | B9 | 346 | `if len(Outstanding(o.Prior)) != 0 {` | MARK |
| branch:if | B10 | 349 | `if checkpoint, ok, checkpointErr := M0Unsettled(o.Prior); checkpointErr != nil {` | MARK |
| branch:else | B11 | 351 | `} else if ok && checkpoint.Kind != "pending-create" {` |  |
| branch:if | B12 | 351 | `} else if ok && checkpoint.Kind != "pending-create" {` |  |
| branch:else | B13 | 353 | `} else if !ok {` |  |
| branch:if | B14 | 353 | `} else if !ok {` |  |
| branch:if | B15 | 358 | `if err := M0ExactPrerequisites(o.Prior); err != nil {` | MARK |
| branch:if | B16 | 362 | `if _, ok := official.M0ReadSourceFor(o.Broker); !ok {` |  |
| branch:if | B17 | 392 | `if r.approvalChannel == "" {` |  |
| branch:if | B18 | 395 | `if r.out == nil {` |  |
| branch:if | B19 | 398 | `if r.now == nil {` |  |
| branch:if | B20 | 401 | `if r.sleep == nil {` |  |
| branch:if | B21 | 404 | `if r.maxSellQuantity <= 0 {` |  |
| branch:if | B22 | 407 | `if r.ttlWait <= 0 {` |  |
| branch:if | B23 | 410 | `if r.triggerWindow <= 0 {` |  |
| branch:if | B24 | 413 | `if r.process.InstanceID == "" {` |  |
| branch:range | B25 | 416 | `for _, id := range o.Redo {` |  |
| branch:if | B26 | 420 | `if r.m0Receipt != nil {` |  |
| return | | 322 | `return nil, errors.New("verifylive: a broker is required")` |  |
| return | | 325 | `return nil, errors.New("verifylive: a recorder is required — the evidence is the point of the exercise")` |  |
| return | | 328 | `return nil, errors.New("verifylive: a confirmer is required; there is no unconfirmed mode")` |  |
| return | | 331 | `return nil, errors.New("verifylive: a batch confirmer is required; the run-wide approval is the default " +` |  |
| return | | 335 | `return nil, errors.New("verifylive: the account could not be identified, so nothing can be attested about it")` |  |
| return | | 339 | `return nil, err` |  |
| return | | 344 | `return nil, errors.New("verifylive: --include-trigger requires M0 receipt, --confirm-each, --resume, and --redo conditional-trigger only")` |  |
| return | | 347 | `return nil, errors.New("verifylive: M0 trigger receipt holds before cleanup because the prior record has outstanding artifacts")` |  |
| return | | 350 | `return nil, checkpointErr` |  |
| return | | 352 | `return nil, fmt.Errorf("verifylive: M0 causal receipt HOLD: unsettled %s owner requires manual reconciliation", checkpoint.Kind)` | MARK |
| return | | 359 | `return nil, err` |  |
| return | | 363 | `return nil, errors.New("verifylive: M0 causal receipt requires an official-client raw read source")` |  |
| return | | 399 | `r.now = func() time.Time { return time.Now().UTC() }` |  |
| return | | 424 | `return r, nil` |  |
| call | | 346 | `Outstanding` | MARK |
| call | | 349 | `M0Unsettled` | MARK |
| call | | 358 | `M0ExactPrerequisites` | MARK |

## `internal/verifylive/record.go` `outstandingLines` (702–727)

- 산출물 `verifylive--outstandinglines.ast.json` · 분기 6 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 705 | `for i, e := range entries {` |  |
| branch:range | B2 | 706 | `for _, a := range e.Artifacts {` |  |
| branch:if | B3 | 708 | `if _, seen := latest[key]; !seen {` |  |
| branch:if | B4 | 712 | `if seen && prev.terminal() && !a.terminal() {` | MARK |
| branch:range | B5 | 721 | `for _, key := range order {` |  |
| branch:if | B6 | 722 | `if l := latest[key]; !l.terminal() {` | MARK |
| return | | 726 | `return out` |  |
| call | | 712 | `prev.terminal` | MARK |
| call | | 712 | `a.terminal` | MARK |
| call | | 722 | `l.terminal` | MARK |

## `internal/verifylive/cleanup.go` `PendingCleanup` (119–121)

- 산출물 `verifylive--pendingcleanup.ast.json` · 분기 0 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| return | | 120 | `return withoutM0ManualReconcile(entries, cleanupFrom(entries, func(id StepID) bool { return Settled(entries, id) }))` |  |
| return | | 120 | `return withoutM0ManualReconcile(entries, cleanupFrom(entries, func(id StepID) bool { return Settled(entries, id) }))` |  |

## `internal/verifylive/reconcile_check.go` `readReconcileSet` (184–197)

- 산출물 `verifylive--readreconcileset.ast.json` · 분기 3 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 187 | `if out.condOpen, err = readReconcileConditionalGroup(ctx, r, "OPEN", ReconcileGroupConditionalOpen, symbol); err != nil {` |  |
| branch:if | B2 | 190 | `if out.plainOpen, err = readReconcilePlainGroup(ctx, r, symbol); err != nil {` |  |
| branch:if | B3 | 193 | `if out.closed, err = readReconcileConditionalGroup(ctx, r, "CLOSED", ReconcileGroupConditionalClosed, symbol); err != nil {` |  |
| return | | 188 | `return reconcileRead{}, err` |  |
| return | | 191 | `return reconcileRead{}, err` |  |
| return | | 194 | `return reconcileRead{}, err` |  |
| return | | 196 | `return out, nil` |  |

## `internal/verifylive/reconcile.go` `Reconcile` (168–316)

- 산출물 `verifylive--reconcile.ast.json` · 분기 31 · 반환 30 · go 0 · defer 1

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 170 | `if now == nil {` |  |
| branch:if | B2 | 173 | `if p.Reader == nil {` |  |
| branch:if | B3 | 176 | `if p.Approve == nil {` |  |
| branch:if | B4 | 180 | `if out == nil {` |  |
| branch:if | B5 | 186 | `if err != nil {` |  |
| branch:if | B6 | 193 | `if err != nil {` |  |
| branch:if | B7 | 199 | `if err != nil {` |  |
| branch:if | B8 | 205 | `if err := checkReconcileRetention(target, now()); err != nil {` | MARK |
| branch:if | B9 | 208 | `if reconcileFreshnessBound <= 0 {` | MARK |
| branch:if | B10 | 217 | `if err := approval.WriteText(out); err != nil {` |  |
| branch:if | B11 | 220 | `if err := p.Approve(ctx, approval); err != nil {` |  |
| branch:if | B12 | 226 | `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` |  |
| branch:if | B13 | 233 | `if err != nil {` |  |
| branch:if | B14 | 237 | `if err != nil {` |  |
| branch:if | B15 | 242 | `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` |  |
| branch:if | B16 | 245 | `if err := checkReconcileWindow(target, start, now()); err != nil {` |  |
| branch:if | B17 | 251 | `if !sameReconcileMultiset(basisRows, second.basis()) {` |  |
| branch:if | B18 | 257 | `if err := checkReconcileRows(first, symbol, p.Market); err != nil {` |  |
| branch:if | B19 | 260 | `if err := checkReconcileRows(second, symbol, p.Market); err != nil {` |  |
| branch:if | B20 | 263 | `if err := checkReconcileAbsence(first, target); err != nil {` |  |
| branch:if | B21 | 266 | `if err := checkReconcileClosedStatus(first); err != nil {` | MARK |
| branch:if | B22 | 273 | `if err != nil {` |  |
| branch:if | B23 | 274 | `if errors.Is(err, errRecordLocked) {` | MARK |
| branch:if | B24 | 281 | `if err != nil {` |  |
| branch:if | B25 | 284 | `if _, err := decodeRecordStrict(again); err != nil {` |  |
| branch:if | B26 | 287 | `if sha256.Sum256(again) != fingerprint {` |  |
| branch:if | B27 | 290 | `if len(again) > 0 && again[len(again)-1] != '\n' {` |  |
| branch:if | B28 | 301 | `if err != nil {` |  |
| branch:if | B29 | 304 | `if err := checkReconcileWindow(target, start, at); err != nil {` |  |
| branch:if | B30 | 309 | `if err := lock.appendLine(line, int64(len(again))); err != nil {` |  |
| branch:if | B31 | 310 | `if errors.Is(err, errRecordSizeMoved) {` | MARK |
| return | | 174 | `return ReconcileResult{}, refuse(RefuseReadError, "reconcile: no read-only official reader was supplied")` | MARK |
| return | | 177 | `return ReconcileResult{}, refuse(RefuseNotApproved, "reconcile: no human approval step was supplied")` | MARK |
| return | | 187 | `return ReconcileResult{}, err` |  |
| return | | 194 | `return ReconcileResult{}, err` |  |
| return | | 200 | `return ReconcileResult{}, err` |  |
| return | | 206 | `return ReconcileResult{}, err` |  |
| return | | 209 | `return ReconcileResult{}, refuse(RefuseFreshnessUnfixed, "the Q3 freshness bound is not fixed")` | MARK |
| return | | 218 | `return ReconcileResult{}, refuse(RefuseApprovalDisplay, "the approval could not be shown to the operator: "+err.Error())` | MARK |
| return | | 221 | `return ReconcileResult{}, refuse(RefuseNotApproved, "operator approval was not given: "+err.Error())` | MARK |
| return | | 227 | `return ReconcileResult{}, err` |  |
| return | | 234 | `return ReconcileResult{}, err` |  |
| return | | 238 | `return ReconcileResult{}, err` |  |
| return | | 243 | `return ReconcileResult{}, err` |  |
| return | | 246 | `return ReconcileResult{}, err` |  |
| return | | 252 | `return ReconcileResult{}, refuse(RefuseReadsDiffer, "the two reads disagree as a multiset of (group, id, status, triggeredOrderId)")` | MARK |
| return | | 258 | `return ReconcileResult{}, err` |  |
| return | | 261 | `return ReconcileResult{}, err` |  |
| return | | 264 | `return ReconcileResult{}, err` |  |
| return | | 267 | `return ReconcileResult{}, err` |  |
| return | | 275 | `return ReconcileResult{}, refuse(RefuseRecordLocked, "another reconciliation holds this record file (the same file, possibly under another path)")` | MARK |
| return | | 277 | `return ReconcileResult{}, refuse(RefuseRecordUnreadable, "the record cannot be opened for the append: "+err.Error())` | MARK |
| return | | 282 | `return ReconcileResult{}, refuse(RefuseRecordUnreadable, err.Error())` | MARK |
| return | | 285 | `return ReconcileResult{}, err` |  |
| return | | 288 | `return ReconcileResult{}, refuse(RefuseRecordChanged, "the record changed between candidate selection and the append")` | MARK |
| return | | 291 | `return ReconcileResult{}, refuse(RefuseRecordTailNoNewline, "the record does not end with a newline")` | MARK |
| return | | 302 | `return ReconcileResult{}, err` |  |
| return | | 305 | `return ReconcileResult{}, err` |  |
| return | | 311 | `return ReconcileResult{}, refuse(RefuseRecordChanged, "the record grew between the locked read and the append — another writer interleaved")` | MARK |
| return | | 313 | `return ReconcileResult{}, fmt.Errorf("verify reconcile: %w", err)` |  |
| return | | 315 | `return ReconcileResult{Artifact: reconciled, Basis: basis}, nil` |  |
| call | | 205 | `checkReconcileRetention` | MARK |
| call | | 266 | `checkReconcileClosedStatus` | MARK |
| call | | 274 | `errors.Is` | MARK |
| call | | 310 | `errors.Is` | MARK |

## `internal/verifylive/projection_reconciled.go` `ReconciledArtifacts` (8–31)

- 산출물 `verifylive--reconciledartifacts.ast.json` · 분기 6 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 11 | `for _, e := range entries {` |  |
| branch:range | B2 | 12 | `for _, a := range e.Artifacts {` |  |
| branch:if | B3 | 15 | `if !seen {` |  |
| branch:if | B4 | 18 | `if seen && prev.terminal() && !a.terminal() {` | MARK |
| branch:range | B5 | 25 | `for _, key := range order {` |  |
| branch:if | B6 | 26 | `if a := latest[key]; a.ReconciledAbsent {` | MARK |
| return | | 30 | `return out` |  |
| call | | 18 | `prev.terminal` | MARK |
| call | | 18 | `a.terminal` | MARK |

## `internal/verifylive/mutate.go` `Runner.cancelConditional` (617–640)

- 산출물 `verifylive--runner.cancelconditional.ast.json` · 분기 2 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 624 | `if err := r.gate(sr, request{` |  |
| branch:if | B2 | 634 | `if err != nil {` |  |
| return | | 628 | `return err` |  |
| return | | 635 | `return err` |  |
| return | | 639 | `return nil` |  |

## `internal/verifylive/steps.go` `Runner.readConditional` (975–1055)

- 산출물 `verifylive--runner.readconditional.ast.json` · 분기 15 · 반환 17 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 976 | `if r.m0ReadSource != nil {` |  |
| branch:if | B2 | 982 | `if r.m0ReceiptErr != nil {` |  |
| branch:if | B3 | 985 | `if err != nil {` |  |
| branch:if | B4 | 986 | `if r.m0CriticalWindow {` |  |
| branch:if | B5 | 991 | `if r.m0ReceiptErr != nil {` |  |
| branch:if | B6 | 1001 | `if r.m0ReceiptUsable() {` |  |
| branch:if | B7 | 1003 | `if !ok {` |  |
| branch:if | B8 | 1012 | `if err != nil {` |  |
| branch:if | B9 | 1013 | `if r.m0CriticalWindow {` |  |
| branch:if | B10 | 1018 | `if r.m0ReceiptErr != nil {` |  |
| branch:if | B11 | 1019 | `if r.m0CriticalWindow {` |  |
| branch:if | B12 | 1034 | `if err != nil {` |  |
| branch:if | B13 | 1035 | `if r.m0CriticalWindow {` |  |
| branch:if | B14 | 1040 | `if r.m0ReceiptErr != nil {` |  |
| branch:if | B15 | 1041 | `if r.m0CriticalWindow {` |  |
| return | | 980 | `return value, callErr` |  |
| return | | 981 | `}, func(value official.RawConditionalOrder) any { return Digest(value) })` |  |
| return | | 983 | `return conditionalView{}, fmt.Errorf("verify: M0 causal receipt HOLD: %w", r.m0ReceiptErr)` | MARK |
| return | | 989 | `return conditionalView{}, err` |  |
| return | | 992 | `return conditionalView{}, fmt.Errorf("verify: M0 causal receipt HOLD: %w", r.m0ReceiptErr)` | MARK |
| return | | 994 | `return conditionalView{ID: raw.ID, ClientOrderID: raw.ClientOrderID, Symbol: raw.Symbol, Market: raw.Market,` |  |
| return | | 1004 | `return conditionalView{}, fmt.Errorf("verify: M0 causal receipt HOLD: broker lacks raw conditional reader")` | MARK |
| return | | 1009 | `return reader.ConditionalOrderRaw(ctx, id)` |  |
| return | | 1011 | `func(v official.RawConditionalOrder) any { return Digest(v) })` |  |
| return | | 1016 | `return conditionalView{}, err` |  |
| return | | 1022 | `return conditionalView{}, fmt.Errorf("verify: M0 causal receipt HOLD: %w", r.m0ReceiptErr)` | MARK |
| return | | 1024 | `return conditionalView{ID: raw.ID, ClientOrderID: raw.ClientOrderID, Symbol: raw.Symbol, Market: raw.Market,` |  |
| return | | 1032 | `func(ctx context.Context) (domain.ConditionalOrder, error) { return r.broker.ConditionalOrder(ctx, id) },` |  |
| return | | 1033 | `func(c domain.ConditionalOrder) any { return c.Status })` | MARK |
| return | | 1038 | `return conditionalView{}, err` |  |
| return | | 1044 | `return conditionalView{}, fmt.Errorf("verify: M0 causal receipt HOLD: %w", r.m0ReceiptErr)` | MARK |
| return | | 1046 | `return conditionalView{` |  |

## `internal/verifylive/cleanup.go` `Runner.runCleanup` (230–254)

- 산출물 `verifylive--runner.runcleanup.ast.json` · 분기 8 · 반환 0 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 232 | `for _, a := range targets {` |  |
| branch:switch | B2 | 234 | `switch a.Kind {` |  |
| branch:case | B3 | 235 | `case KindOrder:` |  |
| branch:case | B4 | 237 | `case KindConditional:` |  |
| branch:case | B5 | 239 | `default:` |  |
| branch:if | B6 | 242 | `if err == nil {` |  |
| branch:if | B7 | 245 | `if first == nil {` |  |
| branch:if | B8 | 248 | `if errors.Is(err, ErrOutsidePlan) \|\| errors.Is(err, context.Canceled) \|\|` | MARK |
| call | | 248 | `errors.Is` | MARK |
| call | | 248 | `errors.Is` | MARK |
| call | | 249 | `errors.Is` | MARK |

## `internal/verifylive/reconcile_check.go` `selectReconcileCandidate` (34–54)

- 산출물 `verifylive--selectreconcilecandidate.ast.json` · 분기 7 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 36 | `for _, a := range PendingCleanup(entries) {` |  |
| branch:if | B2 | 37 | `if a.Kind == KindConditional {` |  |
| branch:if | B3 | 42 | `if err != nil {` |  |
| branch:switch | B4 | 45 | `switch len(candidates) {` |  |
| branch:case | B5 | 46 | `case 0:` |  |
| branch:case | B6 | 48 | `case 1:` |  |
| branch:case | B7 | 50 | `default:` |  |
| return | | 43 | `return Artifact{}, err` |  |
| return | | 47 | `return Artifact{}, refuse(RefuseNoCandidate, "no released outstanding conditional artifact to reconcile")` | MARK |
| return | | 49 | `return candidates[0], nil` |  |
| return | | 51 | `return Artifact{}, refuse(RefuseMultipleCandidates,` | MARK |
| call | | 41 | `withoutM0Unsettled` | MARK |

합계: AST 29 · 분기 169 · 반환 138 · STALE 0
