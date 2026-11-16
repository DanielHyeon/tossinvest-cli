# a191 design — 목록을 떠난 추적 주문의 종결 확인

> 1단계 설계. 분기·early return·side effect 를 근거로 쓰는 문장은 `tools/logic-map` AST 열거에서 옮겼다(좌표 `B<n>(kind@line)`).
> AST 가 열거하지 못하는 것(SQL 문자열·복합 리터럴의 필드 값·주석)은 **「손으로 읽음」**으로 표시했다. AST 원문은
> `analysis/function-logic/internal-filldetect--detector.collect/ast.json`(편집 대상)과 `analysis/cited-ast/*.ast.json`(인용만), 요약은
> `analysis/cited-ast/README.md`. 기준 HEAD `a392308e8`.

## 1. 결함의 기전 (AST)

1. `Detector.collect`(detect.go 363~458)는 OPEN 목록을 끝까지 걷고(`execgw.ScanOrders` call@364, 실패 B1 if@365 → return@366), 추적 집합을
   읽고(call@370, B2 if@371), 목록 행마다 정규 키로 `seen` 을 표시한다(B7 range@399 … B11 if@410).
2. 추적 키 순회 B13(range@423): B14(if@424) `seen[key]` 면 continue. 아니면 `d.Order.OrderRaw`(call@428) — 실패 B15(if@429) → return@430
   (주기 전체 실패). `parseSnapshot`(@433) — 실패 B16(if@434) → return@435. 빈 id 보충 B17(if@440). 정규 범위 불완전 B18(if@445) → return@446,
   범위 불일치 B19(if@449) → return@450. 그다음 `brokerstate.Derive`(call@454) → append(@455). **이 구간의 호출은 이것이 전부이며 OPEN 부재를
   상태 판정에 쓰는 분기는 없다.**
3. `Derive`(derive.go 346~510): `PENDING` B9(case@387) → B10(switch@388): 취소 증거 B11(@389)·후속 B12(@392)·체결>0 B13(@395)이 아니면 B14
   (default@398) `settle(OPEN_UNFILLED)` return@399. M32 의 단건 응답(`PENDING`, `canceledAt` 없음, 체결 0)은 정확히 이 길로 간다.
4. `pollLocked` B5(range@316)가 스냅숏마다 `Ledger.Apply`(call@317). 같은 스냅숏이면 `RecordFill` B20(if@420 `sameSnapshot`) → return@422/425
   (no-op). 그래서 매 주기 아무것도 쓰이지 않는다.
5. `Run` 은 B1(for@518)·B2(if@519 ctx)·B3(if@523 sleep)뿐이고 `PollOnce`(call@522)의 결과를 읽지 않는다(로그 호출 0) — 조용함의 기전.

## 2. 가설 검토 — 채택하되 네 곳을 고친다

Manager 가설: 「추적 주문이 OPEN 에 없고 단건이 살아 있다면 CLOSED 목록에서 id 를 바이트 정확히 찾아 그 행으로 상태를 정한다. 없으면 사유
코드와 경보(fail-closed, 종결 기록 안 씀).」

| 가설의 부분 | 판정 | 이유 |
|---|---|---|
| CLOSED 목록에서 id 를 바이트 정확히 찾는다 | 채택 | 의존이 이미 있다(D2). 선례 `Resolver.scanBoth` B3(range@419)·B6(if@429, `seen[facts.OrderID]` 바이트 키) |
| 「그 행으로 상태를 정한다」 | **수정 — 상태만** | 행의 체결 수량을 그대로 쓰면 종결 목록 혼자서 체결을 만들거나 지운다(§4 Q4). 단건과 수량이 같을 때만 받는다(D3) |
| 「없으면 … 경보」 | **수정 — 지속될 때만** | 방금 접수돼 OPEN 읽기 뒤에 확정된 주문은 정상적으로 OPEN 에도 CLOSED 에도 없다(§3 N1). 첫 관측에서 경보하면 정상 입력을 거부한다 |
| 「fail-closed」 | **수정 — 「오늘과 같음」** | 못 찾음을 refusal(`FailClosed`)로 쓰면 `pollLocked` B7(if@324) → `blockSymbol`(call@326) → `Gate.Block`(blockSymbol call@682)로 운영자 해제 전용 계좌 래치가 새로 생긴다. 못 찾음은 오늘의 단건 스냅숏을 그대로 쓰는 것으로 충분히 보수적이다(종결을 안 씀 = 추적 유지 = 대사 차단 유지) |
| (빠진 것) 조회 빈도 | **추가** | 못 찾음이 지속되면 매 주기 조회가 된다. 주문별 최소 간격(D5)과 retry matrix 계상(D10) |
| (빠진 것) 조회 실패 | **추가** | CLOSED 읽기 실패로 주기를 실패시키면 OPEN·단건·잔고 반영까지 멈추고 outage 가 쌓인다(`pollLocked` B2 if@294 → `outage.failure`). 실패는 그 주문만 「오늘과 같음」(D7) |

더 작은 정석을 찾아봤다(ponytail·search-first): ① 단건 대신 CLOSED 만 읽기 — 체결 반영의 신선도·비용이 나빠짐(CLOSED 는 페이지, 단건은
1요청). ② 「OPEN 에 없음 + 단건 PENDING ⇒ 취소」로 추론 — N1(방금 접수) 경합을 취소로 오판한다. 기각. ③ 대사 쪽에서 missing-order 를 풀기 —
체결 원장의 주인은 감지기이고 대사는 원장을 쓰지 않는다; 판정을 둘로 만든다. 기각. ④ 이미 있는 것: `ScanOrders`(페이지 상한·커서 반복),
`OrderPager`(CLOSED 질의 가능), `parseSnapshot`, `snapshotOrderKey`, `Derive` — 전부 재사용. 새 의존·표준 밖 라이브러리 0.

## 3. 설계

### D1 발동 조건 (M2 = 제외, Manager 2026-11-16)

`collect` B14 가 거짓(OPEN 에서 그 정규 키를 못 봄)이고, 단건 조회가 B15~B19 의 기존 가드를 **모두 통과한 뒤**, 단건 `Derive` 결과가
**`!Terminal && !FailClosed`** 일 때만 후보가 된다. 위치는 call@454 와 append@455 사이 — 기존 가드 뒤라 그 가드들을 가리지 않는다.
UNKNOWN(FailClosed) 단건은 **조회하지 않는다**: CLOSED 의 `REPLACED` 가 전부 미실측인데 그 위에서 기존 fail-closed 래치(UNKNOWN → refusal →
`blockSymbol`)를 없애는 변경이 되고, 오늘 결함(깨끗한 `OPEN_UNFILLED`)에는 필요 없다. 정정 옛 id(`open_with_successor`)는 후속(a167 과 함께).

### D2 질의

`execgw.ScanOrders(ctx, d.Orders, execgw.OrderQuery{Status: "CLOSED", Symbol: <추적 종목>, From: <거래일-1>, To: <거래일+1>, Limit: 100}, 3)`.

- 의존은 이미 있다: `Detector.Orders` 는 `execgw.OrderPager`(detect.go:153), 운영 배선은 `execgw.OfficialOrders`(engine.go:769·772)이고 그
  어댑터가 `Status/Symbol/From/To/Limit` 을 전부 넘긴다(orders_source.go:47~63 → orders_raw.go:62~90, 손으로 읽음 — 분기 없는 필드 복사).
- 페이지 상한·커서 반복 방어는 `ScanOrders` 재사용: B3(for@743)·B4(if@744 상한 → `ErrTooManyPages` return@745)·B5(if@749 오류 → return@750)·
  B6(if@755 끝)·B7(if@758 반복 커서 → `ErrCursorLoop` return@759).
- 거래일 ±1일: `from`/`to` 는 KST 주문 생성일(openapi), 추적 주문의 `TradingDay` 는 시장 현지 거래일(`snapshotTradingDay` → `market.TradingDay`,
  payload.go:144~163 손으로 읽음). US 세션은 KST 로 다음 날에 걸치므로 ±1일이면 양쪽을 다 덮는다. 거래일 문자열이 날짜로 안 읽히면 조회하지
  않고 사유 `closed_lookup_window_invalid`.
- 페이지 상한 3: 종목·3일 창 안의 종결 주문이 300건을 넘는 것은 비정상이다. 넘으면 `ErrTooManyPages` → 사유 `closed_list_truncated`.

### D3 받아들이는 경우 — 하나뿐

종결 목록의 각 행을 `parseSnapshot` 으로 읽고 `AccountRef` 를 선택 계좌로 채운 뒤(단건 경로 detect.go:443 과 같음):

1. 주문번호가 추적 id 와 **바이트 동일**한 행만 후보다. id 를 못 읽는 행이 하나라도 있으면 결론 불가(`closed_list_row_unreadable`).
2. 같은 id 후보가 둘 이상이면, **raw 행 바이트가 전부 같을 때만** 하나로 센다(페이지 경계 중복). 정규 키가 다르거나, 키는 같아도 status·수량 등
   payload 가 다르면 `closed_list_ambiguous`(리뷰 1 #3 수정 — 앞 판본은 「키가 다를 때만」이라 「정확히 하나」와 어긋났음).
3. 후보가 0 이면 `closed_list_not_found`.
4. 후보 키 ≠ 추적 키면 `closed_list_scope_mismatch`.
5. `Derive(viewOf(행, 추적 lineage))` 가 `Terminal && !FailClosed` 가 아니면 `closed_list_not_clean_terminal`.
6. 행의 `Quantity`·`FilledQuantity` 가 단건 스냅숏과 같지 않으면(`qtyEqual`, brokerstate 와 같은 상대 허용 1e-9) `closed_list_quantity_disagrees`.
7. 위를 모두 통과하면 단건 스냅숏에서 **상태 필드만** 행의 것으로 바꾼다: `RawStatus`·`Canceled`·`CanceledAt`·`Derived`. 수량·평균가·체결금액·
   `execution.filledAt`(`BrokerVisibleAt`·`HasBrokerTime`)은 **단건 것을 유지**한다(리뷰 1 #4 수정 — 앞 판본은 행 스냅숏을 통째로 써서, 원장
   직전값보다 체결이 늘어난 주기에는 `RecordFill` B25(if@464)가 **CLOSED 의 가격으로** delta 를 기록할 수 있었다; delta 는 단건이 아니라 원장
   직전 스냅숏과 비교된다 — fills.go:379~382·412~415). 이후는 기존 `pollLocked` → `Apply` → `RecordFill` 그대로다.
8. 주의(사실): `ScanOrders` 는 `HasNext=true` 인데 `NextCursor` 가 빈 값이면 **오류 없이 끝으로 친다**(B6 if@755 `!result.HasNext || next == ""`
   → return@756). 즉 「완주」는 브로커가 커서를 빼먹지 않는다는 가정 위에 있다(OPEN 걷기도 같음). 이것을 거절로 바꾸는 것은 새 가드라
   Manager 결정 M7 로 넘긴다.
9. **(M6 = 요구, Manager 2026-11-16)** 행에 `execution.filledQuantity` 가 **존재하고 십진수로 읽혀야** 받는다(값 0 허용). 없거나 null·빈 값이면
   `closed_list_execution_missing`(오늘과 같음). 이유: `parseSnapshot` 은 `execution` 이 없거나 null 이면 체결 0 으로 읽고(payload.go:87 의 nil 검사,
   `parseDecimal` 178~184 — 손으로 읽음), 그러면 6 의 수량 일치가 **표본 없이 저절로 참**이 된다. 이 판정은 수용 조건 안의 판정이지 새 래치가
   아니다. Q9 조회가 「체결 없이 취소된 행은 execution 이 null」로 나오면 뒤집힐 수 있어 **한 함수(`closedRowCarriesFill`)·한 시험·한 변이(m22)**로
   격리한다. 순서: 4(범위) 뒤, 5(파생) 앞. 체결 수량이 숫자가 아니거나 문자열이 아니어서 `parseSnapshot` 자체가 실패한 경우도 같은 사유로 가른다(코드 리뷰 1 #3).

### D4 받아들이지 않을 때 — 오늘과 같다

D3 의 1~6 어느 것이든, 조회 자체를 안 하거나(D5 간격) 실패한 경우(D7) 모두: **단건 스냅숏을 그대로 쓴다**(오늘의 행동). 종결 기록 없음, 새
refusal 없음, 새 게이트 래치 없음. 사유 코드만 남는다(`Cycle.Unresolved`, D6).

| 사유 코드 | 뜻 |
|---|---|
| `closed_list_not_found` | 창 안의 CLOSED 에 그 id 가 없음 |
| `closed_list_read_failed` | 페이지 읽기 실패(429 포함) |
| `closed_list_truncated` | 페이지 상한(`ErrTooManyPages`) 또는 커서 반복(`ErrCursorLoop`) |
| `closed_list_row_unreadable` | 창 안의 행 하나를 id 까지 못 읽음 |
| `closed_list_ambiguous` | 같은 id 에 raw 바이트가 다른 행이 둘 이상(범위가 다르거나, 같아도 payload 가 다름) |
| `closed_list_scope_mismatch` | 유일한 후보의 범위가 추적 범위와 다름 |
| `closed_list_not_clean_terminal` | 행 파생이 비종결이거나 UNKNOWN |
| `closed_list_quantity_disagrees` | 주문·체결 수량이 단건과 다름 |
| `closed_lookup_window_invalid` | 추적 거래일이 날짜로 안 읽힘 |
| `closed_list_execution_missing` | 행에 `execution.filledQuantity` 가 없거나 읽히지 않음(M6) |
| `closed_lookup_deferred` | 최소 간격 안이거나, 주기당 1건 상한으로 이번 주기에는 안 물음(M8) |

### D5 조회 간격·주기당 1건·시한 (M3 = 30s, M8 = (가) 축소판, Manager 2026-11-16)

- 상태: `Detector` 의 비공개 `leftOpen map[canonicalOrderKey]*leftOpenTrack{firstSeen, lastLookup, reported}`. poll 뮤텍스 아래에서만 쓴다
  (`collect` 호출자 1 = `pollLocked`, `PollOnce` 가 `d.poll` 을 잡음). 시각 판정(간격·보고 시한)은 전부 **주입 `Clock`**(`d.clk().Now()`)으로 한다.
- 후보 등록: `collect` 의 첫 삽입 줄이 그 주기의 후보(키·추적 주문·스냅숏 위치)를 그 주기 시작 시각(`started`)과 함께 적는다. 처음 보는 키는
  `firstSeen = now`.
- **주기당 조회 1건**: 끝 삽입 줄이 그 주기 후보 가운데 「간격이 지난 것」(`lastLookup` 이 비었거나 `now − lastLookup ≥ 30s`)만 고르고,
  한 번도 안 물은 것 → `lastLookup` 이 가장 오래된 것 순(같으면 추적 순서)으로 **하나만** 묻는다. 나머지 후보는 `closed_lookup_deferred` —
  이때 `lastLookup` 은 **갱신하지 않고**, 정리에서 **지우지도 않는다**. 묻은 키는 결과와 상관없이 `lastLookup = now`(실패 포함 — 같은 주기
  재시도 없음).
- **조회 시한 2초**: 조회 **전체**(페이지 합계, a166 예산 대기 포함)를 `context.WithTimeout(ctx, 2s)` 로 감싼다. 이 시한은 매달린 전송을 끊는
  용도로만 쓰고 간격 판정에는 쓰지 않는다. 시한 초과는 `closed_list_read_failed`.
- 정리: 끝 삽입 줄이 이번 주기 후보가 아닌 키를 지운다(OPEN 재등장·종결·추적 종료). `collect` 가 중간에 오류로 끝난 주기는 정리하지 않고
  그 주기의 후보 기록은 다음 주기에 `started` 가 달라 버려진다.
- 재시작하면 상태가 비어 후보 N 개가 **N 주기에 걸쳐 하나씩** 물린다(주기당 1건).
- 수용해도 상태를 바로 지우지 않는다 — 그 스냅숏이 적용되기 전에 주기가 실패하면 다음 주기에도 후보로 남는데 그때 30s 간격을 지켜야
  한다. 종결이 원장에 적용되면 다음 주기에 후보가 아니어서 정리된다(코드 리뷰 1 #2).
- `noteLeftOpen` 은 다른 주기의 남은 후보를 버린다(settle 전 실패한 주기 후보가 쌓이지 않게, 코드 리뷰 1 #1).

### D6 경보 — 지속될 때만 (M1 = A, Manager 2026-11-16)

같은 키가 `firstSeen` 으로부터 30s 이상(주입 Clock) 해결되지 않으면 **한 번** 알린다. 해결(OPEN 재등장·종결·추적 종료)되면 상태를 지우고,
다시 생기면 다시 알린다.

- 통로: `Detector.Alerts`(선택 필드, `Notify(ctx, obs.Event) error`). 운영 배선은 `ectx.Notifier`. nil `*obs.Notifier` 를 인터페이스에 그대로 넣으면
  nil 아닌 인터페이스가 되어 패닉하므로(attestation_watch.go:192 선례 주석) cmd/tossctl 의 **새 파일 도우미**가 nil 을 거른다 — `engineFillDetector`
  는 필드 1줄만 늘어난다(분기 추가 0).
- 이벤트: `fill.order_unresolved`, severity normal. a129 선례(exit_order_log.go:39~46)대로 `internal/obs` 를 고치지 않고 filldetect 지역
  `obs.EventType` 리터럴(등급표에 없으므로 normal).
- 싣는 것: `intent_id`(내부 식별자 — a129 화이트리스트에 있는 필드 종류, exit_order_log.go:153), `market`, `reason`(사유 코드),
  `unresolved_seconds`(지속 시간). **싣지 않는 것**: 브로커 주문번호, 계좌, 종목, 수량, 가격. 제목·본문은 고정 문장.
- 내부 식별자: `filldetect.TrackedOrder` 는 intent id 를 들고 있지 않다. 원장 `journal.TrackedFillOrder.IntentID` 는 있다(fills.go:297). 그래서
  `TrackedOrder` 에 `IntentID` 필드를 더하고 `JournalTracked.TrackedOrders` 가 그것을 옮긴다(기존 함수 본문 1줄 — gate ⑤ 번들 대상).
- 알림 호출도 같은 주기 안에서 일어나므로 **조회와 별개의** 새 2s 시한 context 로 감싼다(조회 시한이 이미 지난 context 를 재사용하지 않음).
  `Notifier.Notify` 는 먼저 동기 로그를 쓰고(Logger 는 ctx 대신 `context.Background()` — log.go:194~203) normal 경로의 `Publisher.Publish` 를 동기
  호출함(notifier.go:149~192). 로그 쓰기가 막히면 시한으로 끊기지 않는다(요구 변경 리뷰 #2 — 확인함, → M12). normal 등급이라 전송 실패에도
  nil 이고 전달은 보장하지 않는다 — spec R4 를 「기록·전송 1회 시도」로 고침(리뷰 #4).

### D7 조회 실패는 주기를 실패시키지 않는다

`collect` 의 기존 오류 반환(B1·B2·B3·B5·B6·B8·B10·B15·B16·B18·B19)은 그대로다. 새 조회의 실패만 그 주문의 「오늘과 같음」+ 사유로 끝난다.
근거: 실패해도 그림은 오늘보다 덜 완전하지 않다(오늘은 이 조회 자체가 없다). 주기를 실패시키면 `pollLocked` B2(if@294) → `outage.failure`
(call@295) → return@297 로 다른 주문의 체결 반영과 freshness 도장(`stampFreshness` call@348)까지 멈춘다. 같은 주기 재시도는 하지 않는다
(a166 「엔진 폴링 같은 조회 즉시 재시도 금지」 — `lastLookup` 을 실패에도 기록).

### D8 upstream·토글 (질문 7)

`internal/filldetect` 는 TossOS 가 만든 패키지다(`git log --diff-filter=A`: `eb58ba489` 2026-07-26 harden-execution-base 3.1). 감지기는 엔진
런타임의 감독 루프 하나로만 돈다(engine.go:739~743 `Name: "filldetect"`, 생성은 `engineFillDetector` 하나 — CodeGraph 호출자 `engineRuntime` 1 +
시험 2). 엔진을 안 켜면(토글 OFF) 이 코드는 돌지 않으므로 「OFF = upstream 동일」(불변식 3)에 닿지 않는다. 새 토글은 두지 않는다(YAGNI):
OFF 로 둘 이유가 「종결을 기록하지 않는 오늘의 결함으로 돌아가기」뿐이고, 받아들이는 경우가 D3 하나로 좁아 끌 대상이 위험한 행동이 아니다.

### D9 진행 중 change 와의 관계 (질문 6)

| change | 같은 가정 | 충돌 | 양보 |
|---|---|---|---|
| a167 (구현 0/19) | 「비종결 = 추적 집합(`terminal = 0`)」(proposal §1) | a191 전에는 KR 에서 취소된 보호 매도가 영원히 비종결이라 a167 의 「비종결 N 초」가 **이미 끝난 주문**에 대해 재가격을 발동한다. a167 §4 는 「재제출은 옛 주문이 종결로 확인된 뒤, status 하나로 판정 안 함」을 자기 판정으로 계획한다 | **a167 이 양보**: 종결 확인을 새로 만들지 말고 a191 의 종결 스냅숏(같은 원장)을 쓴다 — 판정이 둘이면 서로의 시험을 통과시킨다. a191 은 a167 의 선행 |
| a100 (40/115) | 「fill detector 코드는 바꾸지 않고」(design 3c), D8 이 `pollLocked` AST(`internal-filldetect--detector.polllocked/`)와 `detect.go:316-342`·`:349`·`:523` 을 인용 | a191 은 `pollLocked` 를 안 바꾼다(그 번들의 `source_sha256` 은 파일 단위라 **바뀐다**). `collect` 에 2줄이 들어가면 `Run`(:523)·`ensure`(:563) 등 뒤쪽 줄 번호가 밀린다 | **a100 이 재인용**(줄 번호·sha). 설계 충돌은 없음 — a100 의 보호 child 가 목록을 떠나 단건이 낡은 상태를 돌려줄 때 a191 이 같은 규칙으로 닫는다 |
| a087 (48/57) | 「체결가는 filldetect 가 체결 조회로 확정」(design :905), 핀 시험 `TestA087MarketSellRemainderIsTrackedUntilTheBrokerEndsIt` | 없음. 그 시험의 단건은 `CANCELED`(종결)라 D1 이 발동하지 않는다. a191 은 체결 수량·가격을 CLOSED 혼자서 정하지 않는다(D3-6) | — |
| a166 (20/26) | 계좌 공유 3 req/s, 보호 요청 우선, 같은 조회 즉시 재시도 금지 | 없음. 새 호출은 Normal 우선순위(감지기 요청은 우선순위를 안 붙임 — exitloop 만 `Critical` 을 붙임, exitloop.go:872·1393·1554 손으로 읽음)이고 재시도 0 | — |

### D10 요청 예산 (질문 1, §0.4)

| 상황 | 추가 요청 |
|---|---|
| 정상(목록을 떠난 주문은 체결로 떠나 단건이 `FILLED`) | 0 — D1 이 발동하지 않음 |
| 방금 접수돼 OPEN 읽기 뒤에 확정(N1) | 그 주문당 1(대개 1페이지), 다음 주기 OPEN 에 보이면 끝 |
| 이번 결함 상태 | 첫 주기 1회로 종결 → 추적 종료 → 이후 0. 오늘은 그 주문당 단건 1요청/3s(0.33 req/s)가 영원히 나간다 — a191 이 그것도 끝낸다 |
| 못 찾음이 지속 | 주문당 조회 1회/30s, 조회당 페이지 ≤3 → 주문당 최대 ≈0.1 req/s. **M8 로 주기(3s)당 조회 최대 1건**이라 미해결 주문 수와 무관하게 상한은 3페이지/3s(≈1 req/s, 대개 1페이지 → ≈0.33 req/s). 이 수치는 **인증 재시도 없는 정상 경로의 목록 요청량**이다 — 공식 클라이언트의 401 refresh·재전송과 토큰 교환은 따로 센다(client.go:348~375, 요구 변경 리뷰 #6). 재시작은 후보를 한 주기에 하나씩 묻는다 |

retry matrix 정본 표(`openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md`)와 `retry-matrix-slo-2026-10-22.md` §2 에 행을
더하는 것은 구현 커밋과 같은 커밋(tasks 4.4).

### D11 손절 즉시성 (불변식 4)

새 호출은 `filldetect` 루프 안에서만 일어난다. 청산은 별도 감독 루프 `exit`(engine.go:731~737)이고 그 경로에 호출이 더해지지 않는다.
공유 자원은 공식 클라이언트와(a166 활성 시) 계좌 예산뿐이며, 감지기 요청은 Normal 이라 보호 요청(Critical)보다 대기열에서 앞서지 않는다.


**간접 지연은 있다(리뷰 1 #5).** 조회는 `collect` 안, 즉 그 주기의 **어떤 스냅숏도 적용되기 전에** 돈다(`pollLocked` call@293 → B5 range@316).
공식 클라이언트 요청 시한은 15s(client.go:22·137, 손으로 읽음)다. **M8 처분**: 조회 전체 2s 시한 + 주기당 1건. **단 2s 는 취소 요청이지 반환 상한이
아니다(요구 변경 리뷰 #1 — 확인함)**: 토큰 취득은 ctx 를 안 보는 `m.mu.Lock()` 뒤에서 교환까지 함(token.go:62~80), a166 예산 전송은 응답 없는
실패 뒤 `WithoutCancel` 로 최대 2s 더 정리함(transport_linux.go:405~407). 이 성질은 오늘의 OPEN 걷기·단건 조회에도 똑같이 있다. 그래서
같은 주기 다른 주문의 반영 지연은 「정상 경로 2s + 스케줄 오차」이고, 토큰 경합·예산 정리가 겹치면 그보다 길 수 있다(→ M11). 진행 중 CLOSED 요청이 a166 공유 wire 를 그 2s 안에서 점유하는 것은 남는다(Critical 은
다음 소유자 선택에서만 우선). **M11 = 받아들임(Manager 2026-11-17)**: 2s 는 취소 요청이며 반환 상한이 아니다 — goroutine 분리는 하지 않는다. tasks 4.6 은 「정상 경로에서 2s」만 단언한다. **M12 = 그대로**: 경보 호출은 동기(다른 엔진 루프의 Notifier 호출과 같은 성질). **M13 = 현행 유지**: 「안 물은 것 먼저」 — 굶으려면 새 후보가 매 주기 계속 들어와야 한다. 반대로 a191 은 청산을 **푼다**:
보호 매도가 종결되면 `ApplyExitFill` B11(if@1009)이 `ResolvePending`(call@1011)을 불러 다음 관측이 남은 포지션을 다시 판정한다.

### D12 파일 배치

- 새 파일 `internal/filldetect/closed_lookup.go`: 후보 등록·선택·질의·행 선택·수용 판정·간격·지속 보고. 새 시험 파일
  `internal/filldetect/a191_closed_lookup_test.go`(상태별 페이지와 받은 `OrderQuery` 를 기록하는 새 가짜 pager — 기존 `fakePager` 는 안 고침).
- 새 파일 `cmd/tossctl/a191_fill_alerts.go`: nil `*obs.Notifier` 를 거르는 도우미.
- 기존 함수 편집(gate ⑤ 번들 대상 3):
  - `Detector.collect` — call@454 와 append@455 사이 1줄(후보 등록), return@457 앞 1줄(조회·교체·정리·보고). 분기 추가 0.
  - `JournalTracked.TrackedOrders`(ledger.go) — `IntentID` 옮기는 필드 1줄.
  - `engineFillDetector`(cmd/tossctl/engine.go) — `Alerts:` 필드 1줄.
- 타입 필드 추가(함수 아님): `Detector.Alerts`·`Detector.leftOpen`·`Detector.leftOpenCycle`, `TrackedOrder.IntentID`, `Cycle.ClosedReads`·`Cycle.Unresolved`.
- 바꾸지 않는 것: `pollLocked`, `Run`, `parseSnapshot`, `Derive`, journal, reconcile, `ScanOrders`.
- 결정 반영 외 잔여 위험(M7·M9·M10): M7 불채택 — `ScanOrders` 의 빈 커서 종료는 감지기 전체(OPEN 걷기 포함)가 이미 기대는 가정이다. 수용하는 위험은 둘이다: 잘못된 종결
  (「절단 + 뒤 페이지의 상충 중복 행」이 겹칠 때)과, 뒤 페이지에 있던 **유일한 정상 행을 못 봐** `not_found` 로 남는 정상 종결 누락·지속 미해결
  (요구 변경 리뷰 #7). M9 — freshness 도장 유지(잔여 위험 수용). M10 — 잘못된 종결의 안전망 없음(재등장 경보는 후속).

## 4. 질문 1~9 답

**Q1 (의존·예산·페이지 방어)** — 예. `Detector.Orders`(execgw.OrderPager)로 CLOSED 질의 가능(D2), 페이지 상한·커서 반복은 `ScanOrders` 재사용
(AST B4·B7). 추가 요청은 D10 표. 기존 감지기 주기는 OPEN 1요청 + 목록 이탈 비종결 주문당 단건 1요청 + 잔고 sweep 이다(`collect` call@364·@428,
`pollLocked` call@300·@308).

**Q2 (CLOSED 의 시간 범위)** — openapi: `from`/`to` 미지정이면 「전체 기간」, 지정하면 `orderedAt`(KST) 날짜 범위. 영수증: M10 이 날짜 필터 없이
CLOSED 26페이지·2,320건을 걸었고 「계정 CLOSED 이력 ≥2,500건」(R10). 그러니 며칠 묵은 추적 주문도 **찾는 것이 정상**이다. 보존 한도(몇 달 뒤
사라지는지)는 [미측정]. 그래서 창을 그 주문의 거래일 ±1일로 좁힌다(D2) — 오래된 주문도 그 날짜 창 안에서 찾는다.

**Q3 (취소·정정 옛 id 의 CLOSED status)**

| 조합 | 영수증 | 상태 |
|---|---|---|
| KR 취소 주문이 CLOSED 에서 `CANCELED` | M31: CLOSED+OPEN 197건 중 `CANCELED` 119(계좌 이력 집계, **특정 주문과 단건 응답을 짝지은 관측 아님**) | 집계로만 있음 — 「단건 PENDING 인 바로 그 주문이 CLOSED 에 CANCELED 로 있다」는 **미실측** |
| CLOSED `CANCELED` 행의 `canceledAt` 유무 | 없음 | **미실측** — 설계는 의존하지 않음(`Derive` B30 CANCELED → B34 체결 0 → CANCELLED 는 취소 증거를 안 봄) |
| KR 정정 옛 id 의 CLOSED status | M31 은 `REPLACED` 0건(관측된 값은 셋뿐), M32 는 단건만 `PENDING` | **미실측** |
| US 정정 옛 id | M28: 단건 `PENDING_REPLACE` | CLOSED 쪽 **미실측** |
| 부분체결 뒤 취소의 CLOSED status | openapi: `PARTIAL_FILLED` 가 두 그룹 모두에 속함 | **미실측** |

미실측 조합은 전부 D3 의 「깨끗한 종결 + 수량 일치」를 못 넘으면 「오늘과 같음」이다. **그 보수가 거부하게 될 정상 입력**(먼저 전부 열거):

| # | 정상 입력 | 처리 | 비용 |
|---|---|---|---|
| N1 | 방금 접수돼 OPEN 읽기(call@364) 뒤, 추적 읽기(call@370) 전에 확정된 주문(또는 OPEN 반영 지연 [미측정]) | 단건 PENDING → CLOSED 못 찾음 → 오늘과 같음. 다음 주기 OPEN 에 보이면 상태 삭제 | 1요청. 경보 없음(지속 아님) |
| N2 | OPEN 목록이 잘림 — openapi·M7 은 「OPEN 은 cursor/limit 무시, 전량」이라 일어나면 안 됨 | 살아 있는 주문 → 못 찾음 → 오늘과 같음 → 지속되면 경보(거짓 경보, 상태 손상 0) | 30s 당 1요청 |
| N3 | 429·전송 실패로 CLOSED 를 못 읽은 주기 | 오늘과 같음 + `closed_list_read_failed`, 주기 성공, 간격 뒤 재조회 | 0(실패한 1요청) |
| N4 | 정정 확정 직후 lineage 가 아직 안 적힌 순간에 CLOSED 가 `REPLACED` | `Derive` B42 → B45(case@475 후속 없음) UNKNOWN → 오늘과 같음. 다음 주기 lineage 가 있으면 단건(`PENDING`/`PENDING_REPLACE` + 후속)이 UNKNOWN(`open_with_successor`, B12 case@392 · `pendingMutation`)이 되고, **M2 = 제외**라 조회하지 않으므로 오늘과 같은 즉시 refusal·계좌 래치(`pollLocked` B7 → `blockSymbol`)가 난다 — 기존 fail-closed 그대로(리뷰 1 #9). 정정 옛 id 는 후속 | 1요청 |
| N5 | CLOSED 반영 지연(취소됐는데 아직 CLOSED 에 없음) [미측정] | 못 찾음 → 오늘과 같음 → 나타나면 받음 | 30s 당 1요청 |
| N6 | 부분체결 뒤 취소인데 단건의 체결 수량이 낡음 | 수량 불일치 → 오늘과 같음 → 지속 경보 → 운영자 | 운영자 개입 1회 |
| N7 | OPEN 에 없고 CLOSED 가 `PARTIAL_FILLED` | `Derive` OPEN_PARTIALLY_FILLED(비종결) → 오늘과 같음 | 30s 당 1요청 |
| N8 | 며칠 묵은 추적 주문 | 그 거래일 ±1일 창에서 찾음 | 1요청 |
| N9 | 그 종목 3일 창의 종결 주문 300건 초과 | `closed_list_truncated` → 오늘과 같음 → 지속 경보 | 3요청/30s |
| N10 | 창 안에 id 를 못 읽는 행(브로커 형식 변화) | `closed_list_row_unreadable` → 오늘과 같음 | — |

**반대로 거부하지 않을 때 통과하는 것**(D3 를 다 넘는 경우): (a) 브로커가 살아 있는 주문을 CLOSED 에 깨끗한 종결로 실음. (b) 같은 정규 범위 안 id
재사용 — 기존 모델 밖. (c) 단건과 CLOSED 가 **둘 다** 체결 0 으로 낡았는데 실제로는 부분체결, 또는 행에 `execution` 이 없어 0 으로 읽힘(D3-9).
결과는 추적 종료·예약 해제(`RecordFill` B28 → call@513)다. **보유 대사가 이것을 언제나 잡는다는 앞 판본의 주장은 틀렸다(리뷰 1 #8):** 그 심볼에
로컬 포지션이 있고 계좌 보유가 달라질 때만 `Comparer.Compare` B13(case@426) 수량 불일치가 막는다. 살아 있는 미체결 주문은 보유를 바꾸지 않고
(브로커 OPEN 행은 짝이 없으면 외부 주문 — B21 range@468 · B22 if@469 → `ExternalOrd` append@470, 막지 않음), 로컬 포지션 없는 매수 체결은 외부 포지션(B11 case@410, 막지 않음 —
`BlocksEntry` 는 외부를 안 셈)이다. 그래서 (a)·(b)·(c) 는 **오늘 단건이 같은 말을 할 때와 같은 계열**의 잔여 위험이지만 안전망이 없는 경우가 있다.
막는 추가 가드(종결 뒤 검증 유지·브로커 식별 버전 등)는 Manager 결정 M6·M10.

**Q4 (CLOSED 는 FILLED/부분, 단건은 PENDING)** — D3-6 이 막는다: 수량이 다르면 받지 않으므로 그 차이는 **어느 경로로도 체결로 기록되지 않는다**
(오늘과 같음 + 경보). 만약 받았다면: `RecordFill` B25(if@464 delta>0) → `fill_events` 삽입(call@465), B36(if@550 locallyOwned) →
`runApplyHooks`(call@551). 포지션이 이미 닫혀 있으면 매도 delta 는 투영 거절(`reconcileCauseFor` 가 `RefusalOversell` 을 `QUANTITY_MISMATCH` 로,
position_projection.go:428~433 **손으로 읽음 — AST 안 만듦**)로 RECONCILE 에 들어간다. 그래서 체결 수량·체결가를 CLOSED 행 혼자서 믿지 않는다.
수량이 같을 때도 평균가·금액·체결 시각은 **단건 것을 쓴다**(D3-7, 리뷰 1 #4 로 수정). 받아들인 주기의 delta 는 원장 직전 스냅숏 대비이므로
(fills.go:379~382·412~415) 0 이 아닐 수 있다 — 그 delta·가격은 **단건이 같은 주기에 혼자 말했을 값과 같다**. 「CLOSED 가 이긴다」 대안은
**Manager 결정 M4**.

**Q5 (종결 스냅숏 뒤의 사슬)**

- (a) `TrackedFillOrders` 에서 빠지나 — **예**, 조건부. 종결 기록은 `RecordFill` → `upsertFillSnapshot`(call@437)이 scoped 행을 쓰고 같은 범위의
  legacy 행을 미러한다(`mirrorScopedFillSnapshot` 의 `WHERE` 가 범위 일치 — fills.go:778~805, **SQL, 손으로 읽음**). 추적 SQL 은 `f.terminal = 0`
  만 돌려주고(fills.go:1604·1627) 셋째 갈래는 「소유 뒤 커밋된 스냅숏이 하나도 없음」(1641~1651)이다 — **SQL, 손으로 읽음**. 예외: 그 id 에
  **빈 범위 legacy 행**(v15 시대, 계좌·거래일·방향이 빈 행)이 있으면 미러가 그 행을 못 고치고 추적이 남을 수 있다(1608~1611). **Manager 장부 확인 2026-11-16: 해당 없음**(scoped·legacy 행 모두 범위가 채워져 있음).
- (b) `LocalStateFromJournal` 이 열린 주문으로 안 세나 — **(a) 와 같은 조건부 예**. `OpenOrders` 는 `TrackedFillOrders`(call@201)의 결과만으로 B4(range@212)에서 채우므로 (a) 의 legacy 예외가 그대로 넘어온다(리뷰 2 #2).
- (c) 심볼 차단이 스스로 풀리나 — **아니오**. missing-order 블록은 `blocksFor` B2(range@1418)가 만들고 그 필드 값은 `Cause: QUANTITY_MISMATCH`,
  `Release: ReleaseOnAdjustedReconcile`(복합 리터럴 1419~1430, **값은 손으로 읽음**). 차이 없는 비교에서 `Observe` B2(if@637) → B6(range@653) →
  B7(if@654) 통과 → B8(if@662 `!creditUsableBy(t.adjusted[symbol], …)`) 이면 `AwaitingAdjustment`(append@667)로 남는다. 조정 크레딧은 수렴이
  그 심볼에 조정을 썼을 때만 생긴다(`AdjustmentApplied`, 손으로 읽음). 수렴 `ConvergeQuantities` 는 `diff.Quantities` 가 비면 바로 돌아간다
  (converge.go:142~145, 손으로 읽음 — 리뷰 1 #10 이 짚음) → missing-order 만 있는 심볼은 크레딧을 못 얻는다. 해제 길은 운영자 해제뿐이다.
- (d) 계좌 영구 차단을 `reconcile-resolve` 로 풀 수 있게 되나 — **조건부 예**: (a) 의 legacy 예외가 없고 다른 차이가 없을 때 실행 가능해진다(리뷰 1 #10 — 빈 범위 legacy 행이 남으면 여전히 B21 에서 거절). 명령은 B21(if@198 `diff.BlocksEntry()`)에서 막혔는데
  `BlocksEntry`(return@313)는 `Quantities`·`MissingOrders` 길이만 본다(식 자체는 손으로 읽음). 그 주문이 빠지면 다른 차이가 없는 한 통과하고,
  B22(if@201) `validate` 클로저(`buildReconcileResolveRuntime` B12 if@304 블록 0 → 거절, B13 range@307 · B14 if@308 원인 ≠ QUANTITY_MISMATCH → 거절)
  에서 missing-order 블록(원인 QUANTITY_MISMATCH)과 영구 블록(원인 QUANTITY_MISMATCH, `Observe` B18 if@727 의 리터럴 731행 — 값은 손으로 읽음)은
  통과한다. 그 뒤 `Tracker.Resolve`(call@204)는 AST 안 만듦. 엔진 정지 + 공식 스냅숏 3회 안정이 전제(명령 설명).
- 부수 효과(종결 기록이 하는 일, `RecordFill` AST): 원장 직전값이 단건과 같으면(이번 결함 — 매 주기 같은 스냅숏이 no-op 이었음) B25(delta>0) 거짓 → `fill_events` 0, 포지션 수량 변화 0. 직전값보다 늘었으면 단건이 혼자 말했을 delta·가격 그대로 기록된다(리뷰 1 #4 로 수정 — 앞 판본은 「수량 같음이므로 delta 0」이라 틀렸음). B28(if@512
  `obs.Terminal && locallyOwned`) → 예약 해제(`releaseReservationsForOrder` call@513). B30(if@526) → 위험 버킷(call@527·530·533 — 매도는
  sidecar no-op 이라는 주석, 손으로 읽음). B36(if@550) → `runApplyHooks`(call@551) → `ApplyExitFill` B11(if@1009) → `ResolvePending`(call@1011)·
  `ExitEventProposalFilled` 기록. 그 뒤 `JournalLedger.Apply` B7(if@84) → 대사 권위 `Refresh`(call@85). **전부 오늘 단건이 `CANCELED` 를 돌려줄 때
  일어나는 일과 같다** — 새 종류의 부수 효과 없음. 위험 버킷은 delta=0 이면 종결 해제(call@530)만, 단건이 체결 증가를 말한 주기(delta>0)의 등록 BUY 면 체결 전이(call@527 → risk_bucket_fill.go:208~243, 리뷰 2 #6 인용 — 손으로 확인 안 함)도 함께 일어난다 — 어느 쪽이든 단건이 혼자 말했을 때와 같다.

**Q6** — D9 표.

**Q7** — D8.

**Q8 (`reconcile.clean`)** — 범위 밖, proposal 후속. 근거: `ReconcileDriver.logCycle`(reconcileloop.go:535~556)은 `cycle.Err` 가 없으면 무조건
`EventReconcileClean` 을 낸다 — 손으로 읽음(이 change 는 그 함수를 근거로 판정하지 않으므로 AST 안 만듦).

**Q9 (사람 승인 읽기 전용 조회 1회)** — 가장 큰 미실측 가정은 「단건이 `PENDING` 인 **바로 그** KR 주문이 CLOSED 에 깨끗한 종결로, 단건과 같은
수량으로 실려 있다」다(Q3 1행). 그것을 없애는 조회:

- **정확한 조회(새 코드 필요)**: `GET /api/v1/orders?status=CLOSED&symbol=<그 종목>&from=<거래일-1>&to=<거래일+1>&limit=100` 1회 +
  `GET /api/v1/orders/{그 id}` 1회(짝 대조). 출하 CLI 에는 이 필터 조합으로 공식 일반 주문 목록을 읽는 명령이 **없다**(`orders list`·`order show` 는
  WTS, `order conditional list` 는 조건주문 — order.go:91·466, orders.go:18 손으로 읽음; `cmd/tossctl` 비시험 파일 중 `OrdersRaw(`·
  `OrdersPageRaw(` 를 부르는 곳은 grep 으로 `soak.go`·`console.go` 둘뿐이고 둘 다 종목·날짜 필터를 받는 명령이 아님). 그래서 이 형태는 일회용 읽기 전용 도구(a087 probe 선례)를
  만들어야 하고 그 자체가 승인 대상이다. 실행 시 `--config-dir ~/.config/tossctl`(엔진과 같은 토큰 캐시 — 토큰 전쟁 기억).
- **새 코드 없는 대안**: 운영 콘솔의 주문 화면을 한 번 연다. 그 화면은 OPEN 전량 1요청 + CLOSED 최신 100건 1요청(종목·날짜 필터 없음) + 조건주문
  1요청을 보낸다(console.go:1059~1086 손으로 읽음) — 읽기 전용, 주문·토글 버튼은 누르지 않는다. 한계: 그 주문이 최신 100건 밖이면 결론 불가
  (CLOSED 의 정렬 순서 [미측정]), `canceledAt` 은 화면 열로 안 보임(파생 표시에만 씀).
- **볼 필드**: 그 id 와 바이트 동일한 행의 개수, `status`, `canceledAt` 유무(값 아님), `quantity`, `execution.filledQuantity`, `orderedAt` 의 날짜,
  같은 순간 단건 응답의 `status`·`execution.filledQuantity`, 그 id 가 OPEN 에 있는지.
- **가릴 필드**: 계좌번호·계좌 순번, 가격·평균가·금액, 다른 주문의 id·종목·수량, 요청 id, 토큰. 보고에는 「일치 행 수 / status / 수량 일치 여부」만.

## 5. 시험·변이 계획

픽스처는 합성 값만(예: 종목 `SYN001`/`SYNX`, 주문번호 `o-a191-…`, 수량 1~10, 가격 없음 또는 `1`). 핵심 축은 **「OPEN 에 없음 + 단건 PENDING」**
이다 — 기존 시험은 단건이 종결을 돌려주는 경우만 다룬다(`TestPollReadsTrackedOrdersThatLeftTheOpenList`, a087 핀). 기존 `fakePager` 는
`q.Status` 를 무시하므로(detect_test.go:55~67 — 커서로만 고름) 새 시험 파일에 상태별 페이지와 받은 `OrderQuery` 를 기록하는 새 가짜를 둔다.

RED 이름은 tasks 3.1~3.13. base 에서 실패해야 하는 것: 3.1·3.2(종결이 안 쓰임), 3.3(`Cycle.Unresolved` 필드 없음 — 컴파일 실패도 RED 로 셈),
3.6·3.10·3.11. 3.4·3.5·3.7·3.8·3.9 는 base 에서 **통과**한다(조회가 없으니 오늘과 같음) — 그것들은 GREEN 뒤 변이로 가치를 증명한다.

| 변이 | 무엇을 바꾸나 | 잡아야 할 시험 |
|---|---|---|
| m1 | 발동 안 함 | 3.1 |
| m2 | 단건 종결이어도 조회 | 3.9 `TerminalSingleRead…`(CLOSED 호출 0) |
| m3 | OPEN 에서 본 주문도 조회 | 3.9 `StillInTheOpenList…` |
| m4 | `!FailClosed` 조건 삭제 | 3.8 `ReplacedWithoutLineage…` |
| m5 | `Terminal` 조건 삭제 | 3.8 `PartialFilledInTheClosedGroup…` |
| m6 | 체결 수량 일치 삭제 | 3.4 |
| m7 | 주문 수량 일치 삭제 | 3.4(수량만 다른 경우) |
| m8 | 범위 일치 삭제 | 3.7 `AnotherScope…` |
| m9 | 중복 후보 중 첫 행 채택 | 3.7 `Duplicate…` |
| m10 | 조회 실패 시 주기 오류 반환 | 3.5 |
| m11 | 간격 제거 | 3.6 |
| m12 | 창 `From` 을 거래일 그대로(−1 없음) | 3.10(US) |
| m13 | 종목 필터 누락 | 3.10 |
| m14 | 보고 시한 0(즉시 보고) | 3.11 |
| m15 | id 비교를 Trim·대소문자 무시로 | 3.7 `ByteForByte` |
| m16 | 수량 불일치 시 행의 체결 수량 채택 | 3.4(fill_events 0) |
| m17 | 새 호출을 B15~B19 앞으로 이동 | 3.9 `ExistingTrackedReadGuards…`(범위 불일치 단건에서 CLOSED 호출 0) |
| m18 | 정리 호출 삭제(상태가 안 지워짐) | 3.11 `JustPlaced…`(재발 시 다시 즉시 조회·보고) |
| m19 | 수용 시 행의 평균가·금액·체결 시각을 씀(D3-7 위반) | 3.4b `TestA191AcceptedClosedRowKeepsTheSingleReadExecution`(원장 0 → 단건 `PARTIAL_FILLED`/4 → CLOSED `CANCELED`/4: delta 4 가 **단건 가격**으로 기록) |
| m20 | 중복을 정규 키로만 판정(같은 키·다른 payload 를 하나로) | 3.7b `TestA191SameScopeRowsWithDifferentPayloadsAreAmbiguous`(`CANCELED` 대 `PARTIAL_FILLED`, 페이지 순서 뒤집어 두 번) |
| m21 | 발동 조건에서 `!FailClosed` 삭제(UNKNOWN 단건도 조회 — M2 위반) | `TestA191UnknownSingleReadIsNeverLookedUp`(CLOSED 호출 0, 기존 refusal 그대로) |
| m22 | `closedRowCarriesFill` 이 늘 참(M6 위반) | `TestA191ClosedRowWithoutExecutionIsNotAccepted` |
| m23 | 주기당 1건 상한 제거 | `TestA191OnlyOneClosedLookupPerCycle`(후보 3 → 3 주기에 하나씩, 안 물은 것 먼저) |
| m24 | 조회 시한 제거 | `TestA191SlowClosedLookupIsCutAtTheDeadline`(멈춘 pager → 2s 근처에서 끝, 같은 주기 다른 주문은 적용됨) |
| m25 | 경보에 브로커 주문번호·종목을 실음 | `TestA191UnresolvedAlertCarriesOnlyTheInternalIdentifier`(필드 키 집합 = intent_id·market·reason·unresolved_seconds + 값에 주문번호 없음) |
| m26 | deferred 후보의 `lastLookup` 갱신 | `TestA191OnlyOneClosedLookupPerCycle`(미뤄진 것이 다음 주기에 바로 물림) |

가드 위치: 새 판정은 전부 기존 가드(B15~B19) **뒤**, 기존 `Ledger.Apply`·`RecordFill` 판정(refusal·감소·역순) **앞**이다. 새 경로는 스냅숏을
바꿀 뿐 `RecordFill` 의 `classifyFillRefusal`(B1 if@668 caller FailClosed, B5 if@682 감소, B6 if@693 역순)을 건너뛰지 않으므로 뒤의 가드가
여전히 산다 — 3.4 가 「수량 불일치를 받아들이는 변이(m16)」에서도 감소 가드에 가려지지 않는지 보려면 단건 체결 0 → CLOSED 체결 >0(증가
방향)으로 픽스처를 짠다(감소 방향이면 m16 을 `classifyFillRefusal` B5 가 대신 막아 변이가 살아남는다).
