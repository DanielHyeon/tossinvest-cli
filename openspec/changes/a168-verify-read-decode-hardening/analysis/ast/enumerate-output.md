# a168 AST 열거 출력 (`python3 -I enumerate.py` 생성 — 손편집 금지)

## `internal/official/client.go` `Client.doRequest` (203–222)

- 산출물 `official--client.dorequest.ast.json` · 분기 3 · 반환 4 · go 0 · defer 1

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 206 | `if err != nil {` |  |
| branch:if | B2 | 208 | `if c.budget != nil {` |  |
| branch:if | B3 | 216 | `if err != nil {` |  |
| return | | 209 | `return 0, nil, fmt.Errorf("%w: %w", ErrTransport, err)` |  |
| return | | 211 | `return 0, nil, fmt.Errorf("%w: %s", ErrTransport, err)` |  |
| return | | 218 | `return resp.StatusCode, nil, fmt.Errorf("%w: reading body: %s", ErrTransport, err)` | MARK |
| return | | 221 | `return resp.StatusCode, body, nil` | MARK |
| call | | 207 | `observeAttempt` | MARK |
| call | | 217 | `observeAttempt` | MARK |
| call | | 220 | `observeAttempt` | MARK |

## `internal/official/client.go` `Client.get` (249–251)

- 산출물 `official--client.get.ast.json` · 분기 0 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| return | | 250 | `return c.getWithHeaders(ctx, path, q, nil, out)` |  |

## `internal/official/client.go` `Client.getAcct` (304–311)

- 산출물 `official--client.getacct.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 306 | `if err != nil {` |  |
| return | | 307 | `return err` |  |
| return | | 310 | `return c.getWithHeaders(ctx, path, q, extra, out)` |  |

## `internal/official/client.go` `Client.getWithHeaders` (407–426)

- 산출물 `official--client.getwithheaders.ast.json` · 분기 3 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 409 | `if len(q) > 0 {` |  |
| branch:if | B2 | 415 | `if err != nil {` |  |
| branch:range | B3 | 419 | `for k, v := range extraHeaders {` |  |
| return | | 416 | `return nil, fmt.Errorf("%w: %s", ErrTransport, err)` |  |
| return | | 422 | `return req, nil` |  |
| return | | 425 | `return c.send(ctx, makeReq, out)` |  |
| call | | 410 | `q.Encode` | MARK |

## `internal/official/reconcile_detail_reads.go` `Client.ReconcileConditionalOrder` (176–188)

- 산출물 `official--client.reconcileconditionalorder.ast.json` · 분기 3 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 177 | `if strings.TrimSpace(id) == "" {` |  |
| branch:if | B2 | 181 | `if err := c.getAcct(ctx, "/api/v1/conditional-orders/"+url.PathEscape(id), nil, &raw); err != nil {` | MARK |
| branch:if | B3 | 182 | `if requestID, silent := reconcileDetailSilent(err); silent {` |  |
| return | | 178 | `return ReconcileConditionalDetail{}, reconcileSchema("the by-id read names no identifier")` | MARK |
| return | | 183 | `return ReconcileConditionalDetail{Silent: true, RequestID: requestID}, nil` | MARK |
| return | | 185 | `return ReconcileConditionalDetail{}, err` | MARK |
| return | | 187 | `return DecodeReconcileConditionalDetail(raw)` | MARK |
| call | | 178 | `reconcileSchema` | MARK |
| call | | 181 | `c.getAcct` | MARK |
| call | | 187 | `DecodeReconcileConditionalDetail` | MARK |

## `internal/official/reconcile_reads.go` `Client.ReconcileConditionalOrdersPage` (194–203)

- 산출물 `official--client.reconcileconditionalorderspage.ast.json` · 분기 2 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 195 | `if strings.TrimSpace(status) == "" {` |  |
| branch:if | B2 | 199 | `if err := c.getAcct(ctx, "/api/v1/conditional-orders", reconcileListQuery(status, symbol, cursor, limit), &raw); err != nil {` | MARK |
| return | | 196 | `return ReconcileConditionalPage{}, fmt.Errorf("%w: the reconcile read names its group", ErrOrderStatusRequired)` | MARK |
| return | | 200 | `return ReconcileConditionalPage{}, err` | MARK |
| return | | 202 | `return DecodeReconcileConditionalPage(raw)` | MARK |
| call | | 199 | `c.getAcct` | MARK |
| call | | 202 | `DecodeReconcileConditionalPage` | MARK |

## `internal/official/reconcile_reads.go` `Client.ReconcileInstrument` (215–228)

- 산출물 `official--client.reconcileinstrument.ast.json` · 분기 2 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 221 | `if err := c.get(ctx, "/api/v1/stocks", q, &raw); err != nil {` |  |
| branch:if | B2 | 224 | `if len(raw) != 1 {` |  |
| return | | 222 | `return "", err` |  |
| return | | 225 | `return "", fmt.Errorf("%w: the instrument read returned %d items, want exactly one", ErrReconcileSchema, len(raw))` | MARK |
| return | | 227 | `return raw[0].Symbol, nil` |  |

## `internal/official/reconcile_reads.go` `Client.ReconcileOpenOrdersPage` (206–212)

- 산출물 `official--client.reconcileopenorderspage.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 208 | `if err := c.getAcct(ctx, "/api/v1/orders", reconcileListQuery("OPEN", symbol, cursor, limit), &raw); err != nil {` | MARK |
| return | | 209 | `return ReconcileOrderPage{}, err` | MARK |
| return | | 211 | `return DecodeReconcileOrderPage(raw)` | MARK |
| call | | 208 | `c.getAcct` | MARK |
| call | | 211 | `DecodeReconcileOrderPage` | MARK |

## `internal/official/client.go` `Client.send` (335–381)

- 산출물 `official--client.send.ast.json` · 분기 9 · 반환 8 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 337 | `if err != nil {` |  |
| branch:if | B2 | 341 | `if err != nil {` |  |
| branch:if | B3 | 345 | `if err != nil {` |  |
| branch:for | B4 | 359 | `for attempt := 0; attempt < 2 && code == http.StatusUnauthorized; attempt++ {` | MARK |
| branch:if | B5 | 362 | `if err != nil {` |  |
| branch:if | B6 | 366 | `if err != nil {` |  |
| branch:if | B7 | 370 | `if err != nil {` |  |
| branch:if | B8 | 373 | `if !adopted {` |  |
| branch:if | B9 | 377 | `if code < 200 \|\| code >= 300 {` | MARK |
| return | | 338 | `return err` |  |
| return | | 342 | `return err` |  |
| return | | 346 | `return err` |  |
| return | | 363 | `return err` |  |
| return | | 367 | `return err` |  |
| return | | 371 | `return err` |  |
| return | | 378 | `return classifyStatus(code, body)` | MARK |
| return | | 380 | `return unwrapAndDecode(body, out)` | MARK |
| call | | 378 | `classifyStatus` | MARK |
| call | | 380 | `unwrapAndDecode` | MARK |

## `internal/official/strict_quote_reads.go` `Client.strictQuoteFetch` (209–226)

- 산출물 `official--client.strictquotefetch.ast.json` · 분기 3 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 214 | `if outer != nil {` |  |
| branch:if | B2 | 218 | `if err := c.get(traced, path, query, nil); err != nil {` |  |
| branch:if | B3 | 222 | `if err != nil {` |  |
| return | | 219 | `return AttemptTrace{}, err` |  |
| return | | 223 | `return AttemptTrace{}, strictQuoteAdopt(StrictQuoteReasonNoAttempt, err)` |  |
| return | | 225 | `return used, nil` |  |
| call | | 212 | `WithAttemptObserver` | MARK |

## `internal/official/reconcile_detail_reads.go` `DecodeReconcileConditionalDetail` (50–92)

- 산출물 `official--decodereconcileconditionaldetail.ast.json` · 분기 13 · 반환 9 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 53 | `if err != nil {` |  |
| branch:range | B2 | 57 | `for key, dst := range map[string]*string{"conditionalOrderId": &d.ID, "status": &d.Status, "symbol": &d.Symbol, "market": &d.Market} {` | MARK |
| branch:if | B3 | 58 | `if *dst, err = reconcileRequiredString(fields, key); err != nil {` |  |
| branch:if | B4 | 62 | `if d.Type, err = reconcileOptionalString(fields, "type"); err != nil {` |  |
| branch:if | B5 | 66 | `if !ok {` |  |
| branch:if | B6 | 70 | `if err != nil {` |  |
| branch:if | B7 | 73 | `if d.FirstStatus, err = reconcileRequiredString(first, "status"); err != nil {` | MARK |
| branch:if | B8 | 76 | `if d.TriggeredOrderID, err = reconcileOptionalString(first, "triggeredOrderId"); err != nil {` |  |
| branch:if | B9 | 80 | `if rawSecond, ok := fields["second"]; ok {` |  |
| branch:switch | B10 | 81 | `switch t := bytes.TrimSpace(rawSecond); {` |  |
| branch:case | B11 | 82 | `case string(t) == "null":` |  |
| branch:case | B12 | 83 | `case len(t) > 0 && t[0] == '{':` |  |
| branch:case | B13 | 85 | `default:` |  |
| return | | 54 | `return ReconcileConditionalDetail{}, err` | MARK |
| return | | 59 | `return ReconcileConditionalDetail{}, err` | MARK |
| return | | 63 | `return ReconcileConditionalDetail{}, err` | MARK |
| return | | 67 | `return ReconcileConditionalDetail{}, reconcileSchema("first is missing")` | MARK |
| return | | 71 | `return ReconcileConditionalDetail{}, err` | MARK |
| return | | 74 | `return ReconcileConditionalDetail{}, err` | MARK |
| return | | 77 | `return ReconcileConditionalDetail{}, err` | MARK |
| return | | 86 | `return ReconcileConditionalDetail{}, reconcileSchema("second is neither null nor an object")` | MARK |
| return | | 91 | `return d, nil` |  |
| call | | 67 | `reconcileSchema` | MARK |
| call | | 86 | `reconcileSchema` | MARK |
| call | | 90 | `hex.EncodeToString` | MARK |

## `internal/official/reconcile_reads.go` `DecodeReconcileConditionalPage` (136–153)

- 산출물 `official--decodereconcileconditionalpage.ast.json` · 분기 3 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 138 | `if err != nil {` |  |
| branch:range | B2 | 142 | `for i, raw := range env.rows {` |  |
| branch:if | B3 | 144 | `if err := json.Unmarshal(raw, &o); err != nil {` | MARK |
| return | | 139 | `return ReconcileConditionalPage{}, err` | MARK |
| return | | 145 | `return ReconcileConditionalPage{}, reconcileSchema("conditionalOrders[%d] does not decode: %v", i, err)` | MARK |
| return | | 152 | `return page, nil` |  |
| call | | 137 | `decodeReconcileEnvelope` | MARK |
| call | | 144 | `json.Unmarshal` | MARK |
| call | | 145 | `reconcileSchema` | MARK |

## `internal/official/reconcile_reads.go` `decodeReconcileEnvelope` (78–132)

- 산출물 `official--decodereconcileenvelope.ast.json` · 분기 17 · 반환 13 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 80 | `if len(trimmed) == 0 \|\| trimmed[0] != '{' {` |  |
| branch:if | B2 | 84 | `if err := json.Unmarshal(trimmed, &fields); err != nil {` | MARK |
| branch:if | B3 | 88 | `if !ok {` |  |
| branch:if | B4 | 91 | `if t := bytes.TrimSpace(rawRows); len(t) == 0 \|\| t[0] != '[' {` |  |
| branch:if | B5 | 95 | `if err := json.Unmarshal(rawRows, &rows); err != nil {` | MARK |
| branch:range | B6 | 98 | `for i, row := range rows {` |  |
| branch:if | B7 | 99 | `if t := bytes.TrimSpace(row); len(t) == 0 \|\| t[0] != '{' {` |  |
| branch:if | B8 | 104 | `if !ok {` |  |
| branch:switch | B9 | 108 | `switch string(bytes.TrimSpace(rawHasNext)) {` |  |
| branch:case | B10 | 109 | `case "true":` |  |
| branch:case | B11 | 111 | `case "false":` |  |
| branch:case | B12 | 112 | `default:` |  |
| branch:if | B13 | 116 | `if !ok {` |  |
| branch:if | B14 | 120 | `if t := bytes.TrimSpace(rawCursor); string(t) != "null" {` |  |
| branch:if | B15 | 121 | `if len(t) == 0 \|\| t[0] != '"' {` |  |
| branch:if | B16 | 124 | `if err := json.Unmarshal(t, &cursor); err != nil {` | MARK |
| branch:if | B17 | 128 | `if hasNext && strings.TrimSpace(cursor) == "" {` |  |
| return | | 81 | `return reconcilePageEnvelope{}, reconcileSchema("result is not an object (%.20s)", trimmed)` | MARK |
| return | | 85 | `return reconcilePageEnvelope{}, reconcileSchema("result does not decode: %v", err)` | MARK |
| return | | 89 | `return reconcilePageEnvelope{}, reconcileSchema("%s is missing", collection)` | MARK |
| return | | 92 | `return reconcilePageEnvelope{}, reconcileSchema("%s is not an array", collection)` | MARK |
| return | | 96 | `return reconcilePageEnvelope{}, reconcileSchema("%s does not decode: %v", collection, err)` | MARK |
| return | | 100 | `return reconcilePageEnvelope{}, reconcileSchema("%s[%d] is not an object", collection, i)` | MARK |
| return | | 105 | `return reconcilePageEnvelope{}, reconcileSchema("hasNext is missing")` | MARK |
| return | | 113 | `return reconcilePageEnvelope{}, reconcileSchema("hasNext is not a boolean (%.20s)", rawHasNext)` | MARK |
| return | | 117 | `return reconcilePageEnvelope{}, reconcileSchema("nextCursor is missing")` | MARK |
| return | | 122 | `return reconcilePageEnvelope{}, reconcileSchema("nextCursor is neither null nor a string")` | MARK |
| return | | 125 | `return reconcilePageEnvelope{}, reconcileSchema("nextCursor does not decode: %v", err)` | MARK |
| return | | 129 | `return reconcilePageEnvelope{}, reconcileSchema("hasNext is true but nextCursor is empty")` | MARK |
| return | | 131 | `return reconcilePageEnvelope{rows: rows, nextCursor: cursor, hasNext: hasNext}, nil` |  |
| call | | 81 | `reconcileSchema` | MARK |
| call | | 84 | `json.Unmarshal` | MARK |
| call | | 85 | `reconcileSchema` | MARK |
| call | | 89 | `reconcileSchema` | MARK |
| call | | 92 | `reconcileSchema` | MARK |
| call | | 95 | `json.Unmarshal` | MARK |
| call | | 96 | `reconcileSchema` | MARK |
| call | | 100 | `reconcileSchema` | MARK |
| call | | 105 | `reconcileSchema` | MARK |
| call | | 113 | `reconcileSchema` | MARK |
| call | | 117 | `reconcileSchema` | MARK |
| call | | 122 | `reconcileSchema` | MARK |
| call | | 124 | `json.Unmarshal` | MARK |
| call | | 125 | `reconcileSchema` | MARK |
| call | | 129 | `reconcileSchema` | MARK |

## `internal/official/reconcile_reads.go` `DecodeReconcileOrderPage` (156–174)

- 산출물 `official--decodereconcileorderpage.ast.json` · 분기 3 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 158 | `if err != nil {` |  |
| branch:range | B2 | 162 | `for i, raw := range env.rows {` |  |
| branch:if | B3 | 168 | `if err := json.Unmarshal(raw, &o); err != nil {` | MARK |
| return | | 159 | `return ReconcileOrderPage{}, err` | MARK |
| return | | 169 | `return ReconcileOrderPage{}, reconcileSchema("orders[%d] does not decode: %v", i, err)` | MARK |
| return | | 173 | `return page, nil` |  |
| call | | 157 | `decodeReconcileEnvelope` | MARK |
| call | | 168 | `json.Unmarshal` | MARK |
| call | | 169 | `reconcileSchema` | MARK |

## `internal/official/reconcile_detail_reads.go` `reconcileObject` (96–127)

- 산출물 `official--reconcileobject.ast.json` · 분기 8 · 반환 8 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 97 | `if len(raw) == 0 \|\| raw[0] != '{' {` |  |
| branch:if | B2 | 101 | `if _, err := dec.Token(); err != nil { // 여는 '{'` | MARK |
| branch:for | B3 | 105 | `for dec.More() {` | MARK |
| branch:if | B4 | 107 | `if err != nil {` |  |
| branch:if | B5 | 112 | `if err := dec.Decode(&value); err != nil {` | MARK |
| branch:if | B6 | 115 | `if _, dup := fields[key]; dup {` | MARK |
| branch:if | B7 | 120 | `if _, err := dec.Token(); err != nil { // 닫는 '}'` | MARK |
| branch:if | B8 | 123 | `if dec.More() {` | MARK |
| return | | 98 | `return nil, reconcileSchema("%s is not an object (%.20s)", name, raw)` | MARK |
| return | | 102 | `return nil, reconcileSchema("%s does not decode: %v", name, err)` | MARK |
| return | | 108 | `return nil, reconcileSchema("%s does not decode: %v", name, err)` | MARK |
| return | | 113 | `return nil, reconcileSchema("%s.%s does not decode: %v", name, key, err)` | MARK |
| return | | 116 | `return nil, reconcileSchema("%s repeats the key %q — a later value would hide an earlier one", name, key)` | MARK |
| return | | 121 | `return nil, reconcileSchema("%s does not decode: %v", name, err)` | MARK |
| return | | 124 | `return nil, reconcileSchema("%s has trailing data", name)` | MARK |
| return | | 126 | `return fields, nil` |  |
| call | | 98 | `reconcileSchema` | MARK |
| call | | 100 | `json.NewDecoder` | MARK |
| call | | 102 | `reconcileSchema` | MARK |
| call | | 108 | `reconcileSchema` | MARK |
| call | | 112 | `dec.Decode` | MARK |
| call | | 113 | `reconcileSchema` | MARK |
| call | | 116 | `reconcileSchema` | MARK |
| call | | 121 | `reconcileSchema` | MARK |
| call | | 124 | `reconcileSchema` | MARK |

## `internal/official/strict_minute_candles.go` `strictMinuteCheckJSON` (489–501)

- 산출물 `official--strictminutecheckjson.ast.json` · 분기 3 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 491 | `if err := strictMinuteWalk(decoder, 0); err != nil {` | MARK |
| branch:if | B2 | 494 | `if _, err := decoder.Token(); err != io.EOF {` | MARK |
| branch:if | B3 | 495 | `if err == nil {` |  |
| return | | 492 | `return err` |  |
| return | | 496 | `return errors.New("trailing JSON value")` | MARK |
| return | | 498 | `return err` |  |
| return | | 500 | `return nil` |  |
| call | | 490 | `json.NewDecoder` | MARK |
| call | | 494 | `decoder.Token` | MARK |

## `internal/official/strict_minute_candles.go` `strictMinuteFinalAttempt` (247–259)

- 산출물 `official--strictminutefinalattempt.ast.json` · 분기 2 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 248 | `if len(attempts) == 0 {` | MARK |
| branch:if | B2 | 253 | `if last.Err != nil \|\| last.StatusCode < 200 \|\| last.StatusCode >= 300 {` | MARK |
| return | | 250 | `return AttemptTrace{}, strictMinuteRefuse(StrictReasonNoAttempt, "no attempt was observed")` | MARK |
| return | | 255 | `return AttemptTrace{}, strictMinuteRefuse(StrictReasonNoAttempt,` | MARK |
| return | | 258 | `return last, nil` |  |
| call | | 250 | `strictMinuteRefuse` | MARK |
| call | | 255 | `strictMinuteRefuse` | MARK |

## `internal/official/strict_minute_candles.go` `strictMinuteWalk` (503–551)

- 산출물 `official--strictminutewalk.ast.json` · 분기 15 · 반환 11 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 505 | `if err != nil {` |  |
| branch:if | B2 | 509 | `if !ok {` |  |
| branch:if | B3 | 513 | `if depth > strictMinuteMaxDepth {` |  |
| branch:switch | B4 | 516 | `switch delimiter {` |  |
| branch:case | B5 | 517 | `case '{':` |  |
| branch:for | B6 | 519 | `for decoder.More() {` | MARK |
| branch:if | B7 | 521 | `if err != nil {` |  |
| branch:if | B8 | 525 | `if !ok {` |  |
| branch:if | B9 | 528 | `if _, exists := seen[key]; exists {` |  |
| branch:if | B10 | 532 | `if err := strictMinuteWalk(decoder, depth+1); err != nil {` | MARK |
| branch:case | B11 | 536 | `case '[':` |  |
| branch:for | B12 | 537 | `for decoder.More() {` | MARK |
| branch:if | B13 | 538 | `if err := strictMinuteWalk(decoder, depth+1); err != nil {` | MARK |
| branch:case | B14 | 542 | `default:` |  |
| branch:if | B15 | 547 | `if _, err := decoder.Token(); err != nil {` | MARK |
| return | | 506 | `return err` |  |
| return | | 510 | `return nil` |  |
| return | | 514 | `return errors.New("JSON nesting is deeper than " + strconv.Itoa(strictMinuteMaxDepth))` |  |
| return | | 522 | `return err` |  |
| return | | 526 | `return errors.New("object key is not a string")` |  |
| return | | 529 | `return fmt.Errorf("duplicate object key %q", key)` | MARK |
| return | | 533 | `return err` |  |
| return | | 539 | `return err` |  |
| return | | 545 | `return errors.New("unexpected JSON delimiter")` |  |
| return | | 548 | `return err` |  |
| return | | 550 | `return nil` |  |
| call | | 504 | `decoder.Token` | MARK |
| call | | 519 | `decoder.More` | MARK |
| call | | 520 | `decoder.Token` | MARK |
| call | | 537 | `decoder.More` | MARK |
| call | | 547 | `decoder.Token` | MARK |

## `internal/official/client.go` `unwrapAndDecode` (228–243)

- 산출물 `official--unwrapanddecode.ast.json` · 분기 4 · 반환 5 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 229 | `if out == nil {` |  |
| branch:if | B2 | 233 | `if err := json.Unmarshal(body, &env); err != nil {` | MARK |
| branch:if | B3 | 236 | `if env.Result == nil {` |  |
| branch:if | B4 | 239 | `if err := json.Unmarshal(env.Result, out); err != nil {` | MARK |
| return | | 230 | `return nil` |  |
| return | | 234 | `return fmt.Errorf("%w: decoding envelope: %s", ErrServer, err)` | MARK |
| return | | 237 | `return fmt.Errorf("%w: response has no 'result' key", ErrServer)` | MARK |
| return | | 240 | `return fmt.Errorf("%w: decoding result payload: %s", ErrServer, err)` | MARK |
| return | | 242 | `return nil` |  |
| call | | 233 | `json.Unmarshal` | MARK |
| call | | 239 | `json.Unmarshal` | MARK |

## `cmd/tossctl/verify_reconcile.go` `reconcileReads.ensureIdentity` (237–251)

- 산출물 `tossctl--reconcilereads.ensureidentity.ast.json` · 분기 3 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 238 | `if r.verified {` |  |
| branch:if | B2 | 242 | `if err != nil {` |  |
| branch:if | B3 | 245 | `if len(accounts) != 1 \|\| strings.TrimSpace(accounts[0].DisplayName) != r.ref \|\|` |  |
| return | | 239 | `return nil` |  |
| return | | 243 | `return fmt.Errorf("re-reading the account list before the reads: %w", err)` |  |
| return | | 247 | `return errReconcileIdentityChanged` | MARK |
| return | | 250 | `return nil` |  |

## `cmd/tossctl/verify_reconcile_expired.go` `reconcileReads.ReconcileConditionalOrder` (14–19)

- 산출물 `tossctl--reconcilereads.reconcileconditionalorder.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 15 | `if err := r.ensureIdentity(ctx); err != nil {` |  |
| return | | 16 | `return official.ReconcileConditionalDetail{}, err` | MARK |
| return | | 18 | `return r.client.ReconcileConditionalOrder(ctx, id)` | MARK |
| call | | 18 | `r.client.ReconcileConditionalOrder` | MARK |

## `cmd/tossctl/verify_reconcile.go` `reconcileReads.ReconcileConditionalOrdersPage` (253–258)

- 산출물 `tossctl--reconcilereads.reconcileconditionalorderspage.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 254 | `if err := r.ensureIdentity(ctx); err != nil {` |  |
| return | | 255 | `return official.ReconcileConditionalPage{}, err` | MARK |
| return | | 257 | `return r.client.ReconcileConditionalOrdersPage(ctx, status, symbol, cursor, limit)` | MARK |
| call | | 257 | `r.client.ReconcileConditionalOrdersPage` | MARK |

## `cmd/tossctl/verify_reconcile.go` `reconcileReads.ReconcileInstrument` (267–269)

- 산출물 `tossctl--reconcilereads.reconcileinstrument.ast.json` · 분기 0 · 반환 1 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| return | | 268 | `return r.client.ReconcileInstrument(ctx, symbol)` | MARK |
| call | | 268 | `r.client.ReconcileInstrument` | MARK |

## `cmd/tossctl/verify_reconcile.go` `reconcileReads.ReconcileOpenOrdersPage` (260–265)

- 산출물 `tossctl--reconcilereads.reconcileopenorderspage.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 261 | `if err := r.ensureIdentity(ctx); err != nil {` |  |
| return | | 262 | `return official.ReconcileOrderPage{}, err` | MARK |
| return | | 264 | `return r.client.ReconcileOpenOrdersPage(ctx, symbol, cursor, limit)` | MARK |
| call | | 264 | `r.client.ReconcileOpenOrdersPage` | MARK |

## `internal/verifylive/reconcile_check.go` `checkReconcileAbsence` (384–399)

- 산출물 `verifylive--checkreconcileabsence.ast.json` · 분기 4 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 385 | `if err := checkReconcileNoLiveOrder(r); err != nil {` | MARK |
| branch:range | B2 | 388 | `for _, row := range r.closed {` |  |
| branch:if | B3 | 389 | `if row.ID != target.ID {` |  |
| branch:if | B4 | 392 | `if row.Status == "EXPIRED" {` | MARK |
| return | | 386 | `return err` |  |
| return | | 393 | `return refuse(RefuseClosedTargetExpired, "the artifact is listed CLOSED as EXPIRED: "+reconcileQ6Message+` | MARK |
| return | | 396 | `return refuse(RefuseClosedTargetPresent, "the artifact's identifier is listed in the CLOSED group — it is not absent")` | MARK |
| return | | 398 | `return checkReconcileNoFiredTrace(r)` | MARK |
| call | | 385 | `checkReconcileNoLiveOrder` | MARK |
| call | | 398 | `checkReconcileNoFiredTrace` | MARK |

## `internal/verifylive/reconcile_check.go` `checkReconcileClosedStatus` (427–435)

- 산출물 `verifylive--checkreconcileclosedstatus.ast.json` · 분기 2 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 428 | `for _, row := range r.closed {` |  |
| branch:if | B2 | 429 | `if row.Status == "" \|\| reconcileOpenFamily[row.Status] \|\| !reconcileClosedStatusAllowlist[row.Status] {` | MARK |
| return | | 430 | `return refuse(RefuseClosedStatus,` | MARK |
| return | | 434 | `return nil` |  |

## `internal/verifylive/reconcile_expired.go` `checkReconcileExpiry` (79–126)

- 산출물 `verifylive--checkreconcileexpiry.ast.json` · 분기 14 · 반환 10 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 80 | `if err := checkReconcileNoLiveOrder(first); err != nil {` | MARK |
| branch:if | B2 | 83 | `if err := checkReconcileNoFiredTrace(first); err != nil {` | MARK |
| branch:if | B3 | 88 | `if byID[0].Silent \|\| byID[1].Silent {` |  |
| branch:if | B4 | 89 | `if len(closed) > 0 {` |  |
| branch:range | B5 | 97 | `for i, d := range byID {` |  |
| branch:if | B6 | 98 | `if d.ID != target.ID \|\| d.Symbol != target.Symbol \|\| d.Market != market \|\| d.Type != "SINGLE" {` |  |
| branch:if | B7 | 102 | `if err := checkReconcileExpiryStatement(fmt.Sprintf("by-id read %d", i+1), d.Status, d.FirstStatus, d.TriggeredOrderID, d.HasSecond); err != nil {` | MARK |
| branch:range | B8 | 106 | `for _, row := range closed {` |  |
| branch:if | B9 | 107 | `if err := checkReconcileExpiryStatement("the CLOSED row", row.Status, row.FirstStatus, row.TriggeredOrderID, row.HasSecond); err != nil {` | MARK |
| branch:if | B10 | 112 | `if chain := ChainOf(entries, target.Kind, target.ID); chain != "" {` |  |
| branch:range | B11 | 113 | `for _, a := range Outstanding(entries) {` |  |
| branch:if | B12 | 114 | `if a.ChainID == chain && (a.Kind != target.Kind \|\| a.ID != target.ID) {` |  |
| branch:range | B13 | 120 | `for _, row := range first.closed {` |  |
| branch:if | B14 | 121 | `if row.Status == "" \|\| reconcileOpenFamily[row.Status] \|\| !reconcileExpiryClosedVocabulary[row.Status] {` | MARK |
| return | | 81 | `return err` |  |
| return | | 84 | `return err` |  |
| return | | 90 | `return refuse(RefuseExpirySourcesDisagree, fmt.Sprintf(` | MARK |
| return | | 93 | `return refuse(RefuseExpiryEvidenceAbsent, fmt.Sprintf(` | MARK |
| return | | 99 | `return refuse(RefuseExpiryTargetStatus, fmt.Sprintf(` | MARK |
| return | | 103 | `return err` |  |
| return | | 108 | `return err` |  |
| return | | 115 | `return refuse(RefuseExpiryChainOpen, "another artifact of the same chain is still outstanding — the target's expiry says nothing about it")` | MARK |
| return | | 122 | `return refuse(RefuseClosedStatus, fmt.Sprintf("CLOSED row status %q is outside the expiry-mode vocabulary", row.Status))` | MARK |
| return | | 125 | `return nil` |  |
| call | | 80 | `checkReconcileNoLiveOrder` | MARK |
| call | | 83 | `checkReconcileNoFiredTrace` | MARK |
| call | | 102 | `checkReconcileExpiryStatement` | MARK |
| call | | 107 | `checkReconcileExpiryStatement` | MARK |

## `internal/verifylive/reconcile_check.go` `checkReconcileInstrument` (158–167)

- 산출물 `verifylive--checkreconcileinstrument.ast.json` · 분기 2 · 반환 3 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 160 | `if err != nil {` |  |
| branch:if | B2 | 163 | `if echo != symbol {` |  |
| return | | 161 | `return refuse(RefuseInstrumentControl, "the instrument read for the artifact's symbol failed: "+err.Error())` | MARK |
| return | | 164 | `return refuse(RefuseInstrumentControl, "the instrument read echoed a different symbol")` | MARK |
| return | | 166 | `return nil` |  |
| call | | 159 | `r.ReconcileInstrument` | MARK |

## `internal/verifylive/reconcile_check.go` `checkReconcileNoFiredTrace` (417–424)

- 산출물 `verifylive--checkreconcilenofiredtrace.ast.json` · 분기 2 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:range | B1 | 418 | `for _, row := range r.closed {` |  |
| branch:if | B2 | 419 | `if row.TriggeredOrderID != "" \|\| row.Status == "COMPLETED" {` | MARK |
| return | | 420 | `return refuse(RefuseClosedFiredTrace, "the symbol's CLOSED group holds a fired conditional — a successor may have fired")` | MARK |
| return | | 423 | `return nil` |  |

## `internal/verifylive/reconcile_expired.go` `readReconcileByID` (56–62)

- 산출물 `verifylive--readreconcilebyid.ast.json` · 분기 1 · 반환 2 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 58 | `if err != nil {` |  |
| return | | 59 | `return official.ReconcileConditionalDetail{}, refuse(RefuseExpiryReadError, "the by-id read failed: "+err.Error())` | MARK |
| return | | 61 | `return d, nil` |  |
| call | | 57 | `r.ReconcileConditionalOrder` | MARK |

## `internal/verifylive/reconcile_check.go` `readReconcileConditionalGroup` (255–281)

- 산출물 `verifylive--readreconcileconditionalgroup.ast.json` · 분기 7 · 반환 6 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:for | B1 | 258 | `for page := 0; page < maxFixturePages; page++ {` |  |
| branch:if | B2 | 259 | `if err := p.before(); err != nil {` |  |
| branch:if | B3 | 263 | `if err != nil {` |  |
| branch:range | B4 | 266 | `for _, row := range got.Rows {` |  |
| branch:if | B5 | 267 | `if err := p.row(row.ID); err != nil {` |  |
| branch:if | B6 | 273 | `if err != nil {` |  |
| branch:if | B7 | 276 | `if done {` |  |
| return | | 260 | `return nil, err` |  |
| return | | 264 | `return nil, refuse(RefuseReadError, group+": "+err.Error())` | MARK |
| return | | 268 | `return nil, err` |  |
| return | | 274 | `return nil, err` |  |
| return | | 277 | `return rows, nil` |  |
| return | | 280 | `return nil, refuse(RefusePageCap, fmt.Sprintf("%s: more than %d pages", group, maxFixturePages))` | MARK |
| call | | 256 | `newReconcilePager` | MARK |
| call | | 262 | `r.ReconcileConditionalOrdersPage` | MARK |

## `internal/verifylive/reconcile_check.go` `readReconcilePlainGroup` (283–309)

- 산출물 `verifylive--readreconcileplaingroup.ast.json` · 분기 7 · 반환 6 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:for | B1 | 286 | `for page := 0; page < maxFixturePages; page++ {` |  |
| branch:if | B2 | 287 | `if err := p.before(); err != nil {` |  |
| branch:if | B3 | 291 | `if err != nil {` |  |
| branch:range | B4 | 294 | `for _, row := range got.Rows {` |  |
| branch:if | B5 | 295 | `if err := p.row(row.ID); err != nil {` |  |
| branch:if | B6 | 301 | `if err != nil {` |  |
| branch:if | B7 | 304 | `if done {` |  |
| return | | 288 | `return nil, err` |  |
| return | | 292 | `return nil, refuse(RefuseReadError, ReconcileGroupPlainOpen+": "+err.Error())` | MARK |
| return | | 296 | `return nil, err` |  |
| return | | 302 | `return nil, err` |  |
| return | | 305 | `return rows, nil` |  |
| return | | 308 | `return nil, refuse(RefusePageCap, fmt.Sprintf("%s: more than %d pages", ReconcileGroupPlainOpen, maxFixturePages))` | MARK |
| call | | 284 | `newReconcilePager` | MARK |
| call | | 290 | `r.ReconcileOpenOrdersPage` | MARK |

## `internal/verifylive/reconcile_check.go` `readReconcileSet` (195–208)

- 산출물 `verifylive--readreconcileset.ast.json` · 분기 3 · 반환 4 · go 0 · defer 0

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 198 | `if out.condOpen, err = readReconcileConditionalGroup(ctx, r, "OPEN", ReconcileGroupConditionalOpen, symbol); err != nil {` | MARK |
| branch:if | B2 | 201 | `if out.plainOpen, err = readReconcilePlainGroup(ctx, r, symbol); err != nil {` | MARK |
| branch:if | B3 | 204 | `if out.closed, err = readReconcileConditionalGroup(ctx, r, "CLOSED", ReconcileGroupConditionalClosed, symbol); err != nil {` | MARK |
| return | | 199 | `return reconcileRead{}, err` |  |
| return | | 202 | `return reconcileRead{}, err` |  |
| return | | 205 | `return reconcileRead{}, err` |  |
| return | | 207 | `return out, nil` |  |
| call | | 198 | `readReconcileConditionalGroup` | MARK |
| call | | 201 | `readReconcilePlainGroup` | MARK |
| call | | 204 | `readReconcileConditionalGroup` | MARK |

## `internal/verifylive/reconcile.go` `Reconcile` (184–371)

- 산출물 `verifylive--reconcile.ast.json` · 분기 46 · 반환 35 · go 0 · defer 1

| 종류 | id | 줄 | 원문 | 표시 |
| --- | --- | --- | --- | --- |
| branch:if | B1 | 186 | `if now == nil {` |  |
| branch:if | B2 | 189 | `if p.Reader == nil {` |  |
| branch:if | B3 | 192 | `if p.Approve == nil {` |  |
| branch:if | B4 | 196 | `if out == nil {` |  |
| branch:if | B5 | 202 | `if err != nil {` |  |
| branch:if | B6 | 209 | `if err != nil {` |  |
| branch:if | B7 | 215 | `if err != nil {` |  |
| branch:if | B8 | 221 | `if p.Expired {` |  |
| branch:else | B9 | 225 | `} else if err := checkReconcileRetention(target, now()); err != nil {` | MARK |
| branch:if | B10 | 222 | `if err := checkReconcileExpiryVocabulary(); err != nil {` | MARK |
| branch:if | B11 | 225 | `} else if err := checkReconcileRetention(target, now()); err != nil {` | MARK |
| branch:if | B12 | 228 | `if reconcileFreshnessBound <= 0 {` |  |
| branch:if | B13 | 237 | `if p.Expired {` |  |
| branch:if | B14 | 240 | `if err := approval.WriteText(out); err != nil {` |  |
| branch:if | B15 | 243 | `if err := p.Approve(ctx, approval); err != nil {` |  |
| branch:if | B16 | 249 | `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` | MARK |
| branch:if | B17 | 257 | `if p.Expired {` |  |
| branch:if | B18 | 258 | `if byID[0], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {` | MARK |
| branch:if | B19 | 263 | `if err != nil {` |  |
| branch:if | B20 | 267 | `if err != nil {` |  |
| branch:if | B21 | 270 | `if p.Expired {` |  |
| branch:if | B22 | 271 | `if byID[1], err = readReconcileByID(ctx, p.Reader, target.ID); err != nil {` | MARK |
| branch:if | B23 | 277 | `if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {` | MARK |
| branch:if | B24 | 280 | `if err := checkReconcileWindow(target, start, now(), p.Expired); err != nil {` | MARK |
| branch:if | B25 | 286 | `if !sameReconcileMultiset(basisRows, second.basis()) {` | MARK |
| branch:if | B26 | 289 | `if p.Expired {` |  |
| branch:if | B27 | 290 | `if err := checkReconcileByIDPair(byID); err != nil {` | MARK |
| branch:if | B28 | 297 | `if err := checkReconcileRows(first, symbol, p.Market); err != nil {` | MARK |
| branch:if | B29 | 300 | `if err := checkReconcileRows(second, symbol, p.Market); err != nil {` | MARK |
| branch:if | B30 | 303 | `if p.Expired {` |  |
| branch:else | B31 | 308 | `} else {` |  |
| branch:if | B32 | 305 | `if err := checkReconcileExpiry(first, second, byID, target, entries, p.Market); err != nil {` | MARK |
| branch:if | B33 | 309 | `if err := checkReconcileAbsence(first, target); err != nil {` | MARK |
| branch:if | B34 | 312 | `if err := checkReconcileClosedStatus(first); err != nil {` | MARK |
| branch:if | B35 | 320 | `if err != nil {` |  |
| branch:if | B36 | 321 | `if errors.Is(err, errRecordLocked) {` |  |
| branch:if | B37 | 328 | `if err != nil {` |  |
| branch:if | B38 | 331 | `if _, err := decodeRecordStrict(again); err != nil {` | MARK |
| branch:if | B39 | 334 | `if sha256.Sum256(again) != fingerprint {` |  |
| branch:if | B40 | 337 | `if len(again) > 0 && again[len(again)-1] != '\n' {` |  |
| branch:if | B41 | 347 | `if p.Expired {` |  |
| branch:else | B42 | 352 | `} else {` |  |
| branch:if | B43 | 356 | `if err != nil {` |  |
| branch:if | B44 | 359 | `if err := checkReconcileWindow(target, start, at, p.Expired); err != nil {` | MARK |
| branch:if | B45 | 364 | `if err := lock.appendLine(line, int64(len(again))); err != nil {` |  |
| branch:if | B46 | 365 | `if errors.Is(err, errRecordSizeMoved) {` |  |
| return | | 190 | `return ReconcileResult{}, refuse(RefuseReadError, "reconcile: no read-only official reader was supplied")` | MARK |
| return | | 193 | `return ReconcileResult{}, refuse(RefuseNotApproved, "reconcile: no human approval step was supplied")` | MARK |
| return | | 203 | `return ReconcileResult{}, err` | MARK |
| return | | 210 | `return ReconcileResult{}, err` | MARK |
| return | | 216 | `return ReconcileResult{}, err` | MARK |
| return | | 223 | `return ReconcileResult{}, err` | MARK |
| return | | 226 | `return ReconcileResult{}, err` | MARK |
| return | | 229 | `return ReconcileResult{}, refuse(RefuseFreshnessUnfixed, "the Q3 freshness bound is not fixed")` | MARK |
| return | | 241 | `return ReconcileResult{}, refuse(RefuseApprovalDisplay, "the approval could not be shown to the operator: "+err.Error())` | MARK |
| return | | 244 | `return ReconcileResult{}, refuse(RefuseNotApproved, "operator approval was not given: "+err.Error())` | MARK |
| return | | 250 | `return ReconcileResult{}, err` | MARK |
| return | | 259 | `return ReconcileResult{}, err` | MARK |
| return | | 264 | `return ReconcileResult{}, err` | MARK |
| return | | 268 | `return ReconcileResult{}, err` | MARK |
| return | | 272 | `return ReconcileResult{}, err` | MARK |
| return | | 278 | `return ReconcileResult{}, err` | MARK |
| return | | 281 | `return ReconcileResult{}, err` | MARK |
| return | | 287 | `return ReconcileResult{}, refuse(RefuseReadsDiffer, "the two reads disagree as a multiset of (group, id, status, triggeredOrderId)")` | MARK |
| return | | 291 | `return ReconcileResult{}, err` | MARK |
| return | | 298 | `return ReconcileResult{}, err` | MARK |
| return | | 301 | `return ReconcileResult{}, err` | MARK |
| return | | 306 | `return ReconcileResult{}, err` | MARK |
| return | | 310 | `return ReconcileResult{}, err` | MARK |
| return | | 313 | `return ReconcileResult{}, err` | MARK |
| return | | 322 | `return ReconcileResult{}, refuse(RefuseRecordLocked, "another reconciliation holds this record file (the same file, possibly under another path)")` | MARK |
| return | | 324 | `return ReconcileResult{}, refuse(RefuseRecordUnreadable, "the record cannot be opened for the append: "+err.Error())` | MARK |
| return | | 329 | `return ReconcileResult{}, refuse(RefuseRecordUnreadable, err.Error())` | MARK |
| return | | 332 | `return ReconcileResult{}, err` | MARK |
| return | | 335 | `return ReconcileResult{}, refuse(RefuseRecordChanged, "the record changed between candidate selection and the append")` | MARK |
| return | | 338 | `return ReconcileResult{}, refuse(RefuseRecordTailNoNewline, "the record does not end with a newline")` | MARK |
| return | | 357 | `return ReconcileResult{}, err` | MARK |
| return | | 360 | `return ReconcileResult{}, err` | MARK |
| return | | 366 | `return ReconcileResult{}, refuse(RefuseRecordChanged, "the record grew between the locked read and the append — another writer interleaved")` | MARK |
| return | | 368 | `return ReconcileResult{}, fmt.Errorf("verify reconcile: %w", err)` | MARK |
| return | | 370 | `return ReconcileResult{Artifact: reconciled, Basis: basis}, nil` | MARK |
| call | | 208 | `selectReconcileCandidate` | MARK |
| call | | 214 | `bindReconcileAccount` | MARK |
| call | | 222 | `checkReconcileExpiryVocabulary` | MARK |
| call | | 225 | `checkReconcileRetention` | MARK |
| call | | 249 | `checkReconcileInstrument` | MARK |
| call | | 258 | `readReconcileByID` | MARK |
| call | | 262 | `readReconcileSet` | MARK |
| call | | 266 | `readReconcileSet` | MARK |
| call | | 271 | `readReconcileByID` | MARK |
| call | | 277 | `checkReconcileInstrument` | MARK |
| call | | 280 | `checkReconcileWindow` | MARK |
| call | | 286 | `sameReconcileMultiset` | MARK |
| call | | 290 | `checkReconcileByIDPair` | MARK |
| call | | 297 | `checkReconcileRows` | MARK |
| call | | 300 | `checkReconcileRows` | MARK |
| call | | 305 | `checkReconcileExpiry` | MARK |
| call | | 309 | `checkReconcileAbsence` | MARK |
| call | | 312 | `checkReconcileClosedStatus` | MARK |
| call | | 331 | `decodeRecordStrict` | MARK |
| call | | 342 | `ReconcileBasisDigest` | MARK |
| call | | 349 | `ReconcileExpiredBasisDigest` | MARK |
| call | | 351 | `encodeReconcileExpiredLine` | MARK |
| call | | 354 | `encodeReconcileLine` | MARK |
| call | | 359 | `checkReconcileWindow` | MARK |

합계: AST 34 · 분기 189 · 반환 177 · STALE 0
