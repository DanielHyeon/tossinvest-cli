# Function Logic Map: `ExitObserver.applyFloor`

- Source: `internal/app/engine/exitloop.go` (`1624`–`1676`)
- Qualified: `ExitObserver.applyFloor`
- AST evidence: `ast.json` (`source_sha256` 53e631730f185be7…) — 편집 뒤 `3ec1efd2` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 7 · 반환 8

**편집.** **a091 편집(구현 로트)** — 보호 여부 인자 `protective` 추가 · B2 는 `logErr` 대신 `reportZeroFloor`(원인 분류 · 게이트 · 가린 로그 · critical 이면 알림) · 끝의 0주를 B7 로 갈라 `reportZeroFloor`(부분 캡 알림은 그대로). 반환값 `(수량, capped, err)` 무변경 — 변이 M19 · M20 이 잡음.

**역할.** RECONCILE 확정 하한으로 청산 수량을 자른다. 0주는 두 자리에서만 나온다: B2(하한 계산 실패 → 리터럴 `"0"`) · B7(`floor.Quantity` 가 원안보다 작고 0).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `quantity` | 양의 정수 정규형(`big.Int.String`) | `record` → `snapshot.ProjectedQuantity` ← `ProjectWholeShares`(`snapshot.go:93`) | 0 이면 `orderable=false` 라 이 함수에 오지 않는다(design D6) |
| `floor.Quantity` | 음 아닌 십진 정규형, **0 가능** | `riskcalc.ConfirmedFloorQuantity` — `MaxDecimal` → `CanonicalDecimal` 또는 `zeroFloor` 리터럴 `"0"` | 0 이면 B7 |
| `floor.Bound` | 한정 항 | 같음 | Holdings 이면 ③(보유 0) — 옛 종류 |
| `protective` | `isProtective(proposal)` | `submit`(`:1402`) | 보고 등급만 가름 — 반환값과 무관 |
| `o.opts.NotificationsEnabled` | 로드된 설정 | 생산 배선이 덮음(`exitwiring.go:355`) | 거짓이면 옛 종류 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `3ec1efd2` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1625` `if o.opts.Floor == nil {` | 예 |
| B2 | if | `:1629` `if err != nil {` | 예 |
| B3 | if | `:1638` `if !applies {` | 예 |
| B4 | if | `:1642` `if err != nil {` | 아니오 |
| B5 | if | `:1645` `if cmp >= 0 {` | 예 |
| B6 | if | `:1649` `if err != nil {` | 아니오 |
| B7 | if | `:1652` `if isZeroQuantity(floor.Quantity) {` | 예 |

Exact AST return positions: `1626:3`, `1636:3`, `1639:3`, `1643:3`, `1646:3`, `1650:3`, `1657:3`, `1675:2`

## Calls and live bindings

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `o.opts.Floor.ConfirmedFloor` | `:1628` | RECONCILE 확정 하한 | 생산 = `reconcileFloor`(exit 사본, `exitSideFloor` — `exit_record_only.go:26`). RECONCILE 블록이 그 종목을 덮지 않으면 브로커 0회(`exitwiring.go:196-200`). 덮으면 `Retrier.Query` **2회**(Holdings `exitwiring.go:207` · SellableQuantity `:231` — 둘째는 첫째가 성공할 때만) + 원장 읽기 1(`localOpenSells` → `LiveOrdersForSymbol`). 정책 `DefaultRetryPolicy`(`retry.go:129-137`, 배선 `exitwiring.go:48`): 최대 **3**시도, 대기 400ms → 800ms(±25%, 상한 3s), 예산 **8s** 는 **대기만** 자른다(`sleepWithin` `:444-452` — 시도 자체는 자르지 않음). 한 시도 = 공식 클라이언트 `send`(`client.go:320-360`): 아래 요청 열. HTTP 요청 하나의 시한은 `http.Client.Timeout` **15s**(`client.go:20` · `:131`). **상한은 없다(4판 정정 — 2라운드 R2-11)**: 토큰 관리자 잠금 `tm.mu` 은 캐시 파일 읽기 · 교환 · 저장을 쥐고(`token.go:61-78` · `:109-125` · `:172-223`) 기한이 없다. 실현 가능한 요청 열(한 시도): 토큰 캐시 유효 → GET 1 / 캐시 무효 → 교환 1 + GET 1 / GET 이 401 → refresh 1(채택이면 교환 0, 아니면 1) + GET 1, 채택한 토큰도 401 이면 refresh 한 번 더(이번엔 교환) + GET 1(`client.go:344-360` — 둘째 refresh 는 첫째가 채택일 때만). HTTP 만 세면 한 시도 ≤ 교환 2 + GET 3 = 5 요청 × 15s, 한 Query ≤ 3 시도 — **가정이 붙은 HTTP 추정이지 벽시계 상한이 아니다**. 401/403 은 재시도 없이 게이트 래치 + 모드 강화(통지는 기록 전용 — a092). **a091 은 이 호출을 바꾸지 않는다**(기존 비용 — §0.3 의 새 항 아님) |
| `o.reportZeroFloor` | `:1633` | B2 보고(원인 ①④) | 로그 한 줄(계좌 없음 · 가린 오류) + critical 이면 `Alerts.Notify(WithoutCancel)` — `o.opts.Alerts` = `obs.RecordOnly{N, Relay}`(생산 배선 `exitwiring.go:348-349`). 일반 등급 → `NormalRelay.Offer`(`normal_relay.go:43-58` — 비차단 `select`, 버퍼가 차면 버림을 로그로 기록). critical → `n.mu` 아래 `Journal.RecordAlert`(`record_only.go:136-137` — SQLite `BEGIN IMMEDIATE`, `busy_timeout` **5s**(`journal.go:36`) · `synchronous=FULL` fsync, 원격 전송 0). `n.mu` 대기는 **기한 없음**(보유자는 로컬 원장 연산만 — a092 정본 「등급화된 알림」 잠금 문단). 재알림 창 `DefaultRemindAfter` **1h**(`notifier.go:59`). 기록 실패 → 그 자리에서 게이트 래치 + 승격 시도, 오류 반환 → `o.alert` 가 `logErr` 한 줄로 삼킴(`exitloop.go:1818-1819`) |
| `classifyZero` | `:1634` | 원인 분류 | 순수(`ctx.Err()` · 오류 나무 · 한정 항) |
| `riskcalc.CompareDecimal` | `:1641` | 하한 vs 원안 | 순수, 오류 → B4 |
| `fmt.Errorf` | `:1643` | B4 오류 감싸기 | 순수 |
| `riskcalc.SubDecimal` | `:1648` | 잔여 | 순수, 오류 → B6 |
| `fmt.Errorf` | `:1650` | B6 오류 감싸기 | 순수 |
| `isZeroQuantity` | `:1652` | B7 0주 판정 | 순수(수치 비교) |
| `o.reportZeroFloor` | `:1654` | B7 보고(원인 ②③) | 알림 하나(종류는 게이트 · 원인이 고름) — `o.opts.Alerts` = `obs.RecordOnly{N, Relay}`(생산 배선 `exitwiring.go:348-349`). 일반 등급 → `NormalRelay.Offer`(`normal_relay.go:43-58` — 비차단 `select`, 버퍼가 차면 버림을 로그로 기록). critical → `n.mu` 아래 `Journal.RecordAlert`(`record_only.go:136-137` — SQLite `BEGIN IMMEDIATE`, `busy_timeout` **5s**(`journal.go:36`) · `synchronous=FULL` fsync, 원격 전송 0). `n.mu` 대기는 **기한 없음**(보유자는 로컬 원장 연산만 — a092 정본 「등급화된 알림」 잠금 문단). 재알림 창 `DefaultRemindAfter` **1h**(`notifier.go:59`). 기록 실패 → 그 자리에서 게이트 래치 + 승격 시도, 오류 반환 → `o.alert` 가 `logErr` 한 줄로 삼킴(`exitloop.go:1818-1819`) |
| `classifyZero` | `:1655` | 원인 분류 | 순수 |
| `o.alert` | `:1659` | 부분 캡 알림(무변경) | `o.opts.Alerts` = `obs.RecordOnly{N, Relay}`(생산 배선 `exitwiring.go:348-349`). 일반 등급 → `NormalRelay.Offer`(`normal_relay.go:43-58` — 비차단 `select`, 버퍼가 차면 버림을 로그로 기록). critical → `n.mu` 아래 `Journal.RecordAlert`(`record_only.go:136-137` — SQLite `BEGIN IMMEDIATE`, `busy_timeout` **5s**(`journal.go:36`) · `synchronous=FULL` fsync, 원격 전송 0). `n.mu` 대기는 **기한 없음**(보유자는 로컬 원장 연산만 — a092 정본 「등급화된 알림」 잠금 문단). 재알림 창 `DefaultRemindAfter` **1h**(`notifier.go:59`). 기록 실패 → 그 자리에서 게이트 래치 + 승격 시도, 오류 반환 → `o.alert` 가 `logErr` 한 줄로 삼킴(`exitloop.go:1818-1819`). 종류 `EventExitProposalCapped` = normal |
| `string` | `:1661` | event key | 순수 |
| `o.label` | `:1662` | 종목 표시명 | 메모리 조회 |
| `fmt.Sprintf` | `:1663` | 부분 캡 본문(무변경) | 순수 |

## State mutations and fallbacks

상태 변경 없음. 반환값과 로그 · 알림이 전부. 새 알림 기록은 `context.WithoutCancel` — 종료 중이면 그 기록만큼(기한 없이) 종료가 늦는다(design D5 「종료 중 보고 대기」).

## Safety conclusion

- 제출 수량 계산(B1~B7 의 반환)은 편집 전과 같다(§0.3 · §0.9 — 변이 M19 · M20 CAUGHT). 편집은 보고뿐. High-risk: yes(**확정 하한이** 손절을 0주로 깎는 유일한 자리).
