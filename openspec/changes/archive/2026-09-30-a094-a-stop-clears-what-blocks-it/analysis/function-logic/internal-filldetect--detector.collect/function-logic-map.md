# Function Logic Map: `Detector.collect`

- Source: `internal/filldetect/detect.go` (`363`–`458`)
- Qualified: `Detector.collect`
- AST evidence: `ast.json` (`source_sha256` 5441296826821097…)
- Risk scan: `risk-pattern-report.md`
- 분기 19 · return 12 · 호출 43

**역할.** 브로커의 미체결 목록과 추적 주문을 읽어 스냅샷으로 만든다. **B1 `:365` 직전의 `ScanOrders`가 a094 3판이 재사용할 읽기다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.Orders` | `execgw.OrderPager` | `engine.go:420` `OfficialOrders{Client: ectx.Official}` | **a094 R2가 새로 넣으려던 것과 같은 원천** |
| `cfg.MaxPages` | 페이지 상한 | `filldetect.Config` | `ScanOrders`의 완주 조건 |
| ``OrderQuery{Status: statusOpen}`` | 계정 전체 OPEN | `detect.go:364` | **종목 필터 없음 — 계정 전체다** |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy,filldetect}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:365` `if err != nil {` | `d.Tracked.TrackedOrders`, `fmt.Errorf`, `len` | :366 | 예 |
| B2 | if | `:371` `if err != nil {` | `d.Tracked.SelectedAccountRef`, `fmt.Errorf`, `strings.TrimSpace` | :372 | 아니오 |
| B3 | if | `:375` `if accountRef == "" {` | `errors.New`, `len`, `make` | :376 | 아니오 |
| B4 | range | `:382` `for _, order := range tracked {` | `trackedOrderKey` | — | 예 |
| B5 | if | `:384` `if keyErr != nil {` | — | :385 | 아니오 |
| B6 | if | `:387` `if _, duplicate := trackedByKey[key]; duplicate {` | `append`, `d.brokerVisibleFallback`, `fmt.Errorf`, `len`, `make` | :388 | 예 |
| B7 | range | `:399` `for _, raw := range raws {` | `parseSnapshot` | — | 예 |
| B8 | if | `:401` `if err != nil {` | `fmt.Errorf`, `snapshotOrderKey` | :404 | 예 |
| B9 | if | `:409` `if complete {` | — | — | 예 |
| B10 | else | `:414` `} else if len(trackedByID[snap.OrderID]) > 0 {` | — | — | 예 |
| B11 | if | `:410` `if order, locallyTracked := trackedByKey[key]; locallyTracked {` | — | — | 예 |
| B12 | if | `:414` `} else if len(trackedByID[snap.OrderID]) > 0 {` | `append`, `brokerstate.Derive`, `fmt.Errorf`, `len`, `viewOf` | :415 | 예 |
| B13 | range | `:423` `for _, key := range trackedKeys {` | — | — | 예 |
| B14 | if | `:424` `if seen[key] {` | `d.Order.OrderRaw` | — | 예 |
| B15 | if | `:429` `if err != nil {` | `fmt.Errorf`, `parseSnapshot` | :430 | 아니오 |
| B16 | if | `:434` `if err != nil {` | `fmt.Errorf` | :435 | 아니오 |
| B17 | if | `:440` `if strings.TrimSpace(snap.OrderID) == "" {` | `snapshotOrderKey`, `strings.TrimSpace` | — | 예 |
| B18 | if | `:445` `if !complete {` | `fmt.Errorf` | :446 | 아니오 |
| B19 | if | `:449` `if actual != key {` | `append`, `brokerstate.Derive`, `fmt.Errorf`, `viewOf` | :450, :457 | 예 |

## Calls and live bindings

**`execgw.ScanOrders(ctx, d.Orders, OrderQuery{Status: statusOpen}, cfg.MaxPages)`(`:364`)** · `d.Tracked.TrackedOrders` · `parseSnapshot` · `d.brokerVisibleFallback`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 읽고 스냅샷을 만든다. 원장 쓰기는 호출자가 한다.

## Safety conclusion

- **Safe edit boundary**: **a094 3판은 이 함수를 바꾸지 않고 그 산출을 읽는다.** `Detector`는 프로덕션에서 무조건 돌고(`cmd/tossctl/engine.go:391-396`) 주기는 3초(`detect.go:128`)다. 따라서 손절 경로에 동기 브로커 왕복을 넣는 대신 **최대 3초 된 스냅샷**을 읽는다 — §0.3 지연이 2초에서 ~0이 되고 §0.4 호출이 늘지 않는다. 대가는 **계정 전체 목록이므로 종목 필터가 클라이언트측**이 되는 것과, 스냅샷이 상한보다 오래되면 `clear=false`로 떨어뜨려야 한다는 것이다.
- **High-risk impact**: no — 이 함수 자체는 읽기다. 위험은 그 산출을 손절 판단에 쓰는 쪽에 있다.
