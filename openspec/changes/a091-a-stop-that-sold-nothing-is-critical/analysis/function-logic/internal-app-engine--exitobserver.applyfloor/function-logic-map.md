# Function Logic Map: `ExitObserver.applyFloor`

- Source: `internal/app/engine/exitloop.go` (`1617`–`1661`)
- Qualified: `ExitObserver.applyFloor`
- AST evidence: `ast.json` (`source_sha256` 0f943813a3efa423…) — base `b30318d6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 6 · 반환 7

**편집.** **a091 편집 대상**(tasks 3.x) — 보호 여부 인자 하나 추가 · 0주 두 경로(B2 · 끝)의 보고 종류 · 문구. 반환값 `(수량, capped, err)` 무변경.

**역할.** RECONCILE 확정 하한으로 청산 수량을 자른다. 0주는 두 자리에서만 나온다: B2(하한 계산 실패 → 리터럴 `"0"`) · 끝(`floor.Quantity` 가 원안보다 작고 그것이 `"0"`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `quantity` | 양의 정수 정규형(`big.Int.String`) | `record` → `snapshot.ProjectedQuantity` ← `ProjectWholeShares`(`snapshot.go:93`) | 0 이면 `orderable=false` 라 이 함수에 오지 않는다(아래 M1) |
| `floor.Quantity` | 음 아닌 십진 정규형, **0 가능** | `riskcalc.ConfirmedFloorQuantity` — `MaxDecimal("0", …)` → `CanonicalDecimal`(`decimal.go:92-101`) 또는 `zeroFloor` 리터럴 `"0"`(`confirmed_floor.go:236-243`) | 0 이면 끝 경로가 `"0"` 반환 |
| `o.opts.Floor` | nil 허용 | 주입(`exitSideFloor`) | nil 이면 무캡(B1) |
| **보호/익절 맥락** | — | **인자에 없다** | ⚠ 이 함수는 제안이 손절인지 모른다 — a091 이 `submit` 에서 `isProtective(proposal)` 를 넘긴다(D2) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 base `b30318d6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`). engine 패키지 실행은 `-trimpath` 로 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다 — 커버리지 프로파일은 그대로 쓰인다(두 시험은 이 함수들과 무관한 AST 핀).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1618` `if o.opts.Floor == nil {` | 예 |
| B2 | if | `:1622` `if err != nil {` | 예 |
| B3 | if | `:1630` `if !applies {` | 예 |
| B4 | if | `:1634` `if err != nil {` | 아니오 |
| B5 | if | `:1637` `if cmp >= 0 {` | 예 |
| B6 | if | `:1641` `if err != nil {` | 아니오 |

Exact AST return positions: `1619:3`, `1628:3`, `1631:3`, `1635:3`, `1638:3`, `1642:3`, `1660:2`

## Calls and live bindings

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `o.opts.Floor.ConfirmedFloor` | `:1621` | RECONCILE 확정 하한 | 생산 = `reconcileFloor`(exit 사본, `exitSideFloor` — `exit_record_only.go:26`). RECONCILE 블록이 그 종목을 덮지 않으면 브로커 0회(`exitwiring.go:196-200`). 덮으면 `Retrier.Query` **2회**(Holdings · SellableQuantity, `:204` · `:227`) + 원장 읽기 1(`localOpenSells` → `LiveOrdersForSymbol`). 정책 `DefaultRetryPolicy`(`retry.go:129-137`, 배선 `exitwiring.go:48`): 최대 **3**시도, 대기 400ms → 800ms(±25%, 상한 3s), 예산 **8s** 는 **대기만** 자른다(`sleepWithin` `:444-452` — 시도 자체는 자르지 않음). 한 시도 = 공식 클라이언트 `send`(`client.go:320-360`): 토큰(캐시 유효면 0, 아니면 교환 1) + GET 1, 401 이면 refresh 최대 **2**회(각 교환 ≤1 + GET 1). HTTP 하나의 상한은 `http.Client.Timeout` **15s**(`client.go:20` · `:131`). 토큰 관리자 잠금(`tm.mu`) 대기는 **기한 없음**(다른 보유자의 교환 1회 ≤15s 동안). 한 Query 최악 ≈ 재시도 시작 창 8s + 한 시도 ≤(15+15)+2×(15+15)=90s ≈ **98s**, 두 Query ≈ **196s**(401 연속 · 교환 반복의 병적 경우). 401/403 은 재시도 없이 게이트 래치 + 모드 강화(통지는 기록 전용 — a092). **a091 은 이 호출을 바꾸지 않는다**(기존 비용 — §0.3 의 새 항 아님) |
| `o.logErr` | `:1626` | B2 의 유일한 기록(오류 객체) | 구조화 로그 한 줄(`exitloop.go:1834-1839`), 반환 없음. **종류가 `EventExitProposalCapped`** — H2 가 바꾸는 자리 |
| `riskcalc.CompareDecimal` | `:1633` | 하한 vs 원안 | 순수 계산, 오류 → B4 |
| `fmt.Errorf` | `:1635` | B4 오류 감싸기 | 순수 |
| `riskcalc.SubDecimal` | `:1640` | 잔여 | 순수 계산, 오류 → B6 |
| `fmt.Errorf` | `:1642` | B6 오류 감싸기 | 순수 |
| `o.alert` | `:1644` | 캡 알림(부분 · 0주 공통 — 현행) | `o.opts.Alerts` = `obs.RecordOnly{N, Relay}`(생산 배선 `exitwiring.go:348-349`). 일반 등급 → `NormalRelay.Offer`(`normal_relay.go:43-58` — 비차단 `select`, 버퍼가 차면 버림을 로그로 기록). critical → `n.mu` 아래 `Journal.RecordAlert`(`record_only.go:136-137` — SQLite `BEGIN IMMEDIATE`, `busy_timeout` **5s**(`journal.go:36`) · `synchronous=FULL` fsync, 원격 전송 0). `n.mu` 대기는 **기한 없음**(보유자는 로컬 원장 연산만 — a092 정본 「등급화된 알림」 잠금 문단). 재알림 창 `DefaultRemindAfter` **1h**(`notifier.go:59`). 기록 실패 → 그 자리에서 게이트 래치 + 승격 시도, 오류 반환 → `o.alert` 가 `logErr` 한 줄로 삼킴(`exitloop.go:1818-1819`). **현행 종류 `EventExitProposalCapped` = normal** → Relay 경로(outbox 행 0) |
| `string` | `:1646` | event key 조립 `type\|position` | 순수 |
| `o.label` | `:1647` | 종목 표시명 | 메모리 조회 |
| `fmt.Sprintf` | `:1648` | 본문 | 순수 |

## State mutations and fallbacks

상태 변경 없음. 반환값과 로그 · 알림이 전부. B2 의 fail-closed(0 으로 봄) 방향은 옳다 — 문제는 보고의 등급 · 종류 · 문구다.

## Safety conclusion

- 제출 수량 계산(B1~B6 · 끝의 반환)은 a091 이 건드리지 않는다(§0.3 · §0.9). 편집은 B2 의 `logErr` 종류 · B2 에 알림 추가 · 끝의 알림 종류/문구 분기뿐. High-risk: yes(손절이 0주로 깎이는 유일한 자리).
