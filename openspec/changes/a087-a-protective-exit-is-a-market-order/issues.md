# a087 issues

## I-P1 — Phase 1(D2a) 의 대상 모집단이 생산에서 0 이다 (blocking · 2026-09-30 · Teammate → Manager)

**분류**: WORKFLOW 예외 경로 ① blocking(스펙 전제와 코드 사실의 모순). production 편집 전 중단.

### 주장

design D2 는 "관측가도 기준선도 없으면 청산을 거부한다. 시세를 못 읽은 순간 손절이 막힌다"고 쓰고, D2a 는 그 거부
(`sellIntent` B2, `exitloop.go:1576-1579`) 앞에 KR 하한가 단을 둔다. **현재 HEAD 에서 B2 는 — 그 앞의 B1(기준선 폴백)부터 —
생산 경로로 도달할 수 없다.** 따라서 D2a 를 구현하면 생산 동작 변화는 0 이고, 새 단과 `PriceLimits` 배선은 도달 불가 코드가 된다.

### 증거 (tasks P1.2 ① 모집단 열거)

1. **호출 사슬 (CodeGraph 1.6.0)** — `sellIntent` 호출자 = `submit` 1(`:1382`), `submit` 비시험 호출자 = `record` 1(`:1301`),
   `record` 호출자 = `judgeRatchet`·`judgeLadder`. 메서드 값 참조 0(`grep '\.sellIntent\b|\.submit\b'`, 비시험).
2. **값의 출처** — `submit` 은 `observed` 를 재대입 없이 넘기고, `record` 는 `judgement.ObservedPrice = snapshot.ObservedPrice`
   (`:1190`). 스냅숏의 값은 `EvaluateRatchetSnapshot`(`snapshot.go:227`) · `EvaluateLadderSnapshot`(`:155`)이 채운다.
3. **AST (tools/logic-map, `analysis/p1-population/*.ast.json`)** — 두 스냅숏 함수는 `EvaluateRatchet`(`:189`) ·
   `EvaluateLadder`(`:120`) 의 오류를 그대로 반환(`:191` · `:122`)한 뒤에만 스냅숏을 만든다. 두 평가기는 성공 반환 전에
   `positive("observed price", …)`(`ratchet.go:347` · `ladder.go:333`)를 무조건 지나며, 그 앞의 반환은 전부 오류 반환이다
   (ratchet 341·345 / ladder 309·314·319·322·326). `positive` 는 파싱 실패와 `Sign() <= 0` 을 거부(`ratchet.go:592-601`).
4. **커버리지 실측** — `go test ./internal/app/engine/ -count=1 -coverprofile` (157 s, 70.7%): B1 본문 `1573.17-1575.3` = 0,
   B2 본문 `1576.17-1579.3` = 0 (`analysis/p1-population/sellintent-coverage.txt`). 거부문을 단언하는 시험은 저장소에 0.

### "가격이 없어 손절이 막히는" 실제 자리

가격이 없으면 거부가 `sellIntent` 에서 일어나는 것이 아니라 **판정 자체가 일어나지 않는다** — `observe` 가 `Last <= 0` 을 거르고
(`:765`), `ObserveOnce` 가 응답에서 빠진 종목·만료된 시세를 무음 `continue` 한다. 이것은 a090(미관측 계수) 의 범위이고, 하한가 단은
그 자리를 돕지 못한다(이탈 여부를 판정할 가격이 없으므로 보호 제안 자체가 없다).

### Manager 결정 요청 (선택지)

1. **Phase 1 을 not-applicable 로 닫는다** — 모집단 0 을 근거로 P1.2~P1.4 를 `not-applicable`, D2a 를 "측정 결과 불요" 로 정정.
   B1·B2 는 오늘처럼 방어 코드로 남는다(도달 불가지만 fail-closed). 가장 작다(YAGNI).
2. **방어 심층으로 구현한다** — 미래 호출자가 빈 관측가를 넘길 때만 효력. 비용: 도달 불가 분기에 High-risk 코드 + 새 GET 배선,
   변이 시험은 직접 호출로만 가능(생산 경로 지도자 없음 — [[tests-that-always-drive-leave-production-untested]]).
3. **Phase 1 을 재조준한다** — 실제 가격 부재 경로(판정 불가)로. 다만 그것은 a090 과 표면이 겹친다.

Teammate 권고: 1. B2 의 도달 불가를 시험으로 못 박는 것(평가기가 빈/0 관측가를 거부 — `ratchet_test.go:383` 은 `-1` 만,
ladder 쪽 빈 값 단언은 없음)은 선택 사항으로 함께 제안한다.

### 처분 (2026-09-30, Manager)

**선택지 1 채택** — Phase 1 불구현 종결, D2a 전제 정정. Manager 가 모집단 열거를 독립 대조해 확인했다(positive 가드 두 자리,
snapshot 평가-성공-후-생성 두 자리, `:765` 무음 skip, `sellIntent` 생산 호출 1곳). 근거: 도달 불가 High-risk 코드 금지 ·
선택지 2 는 미검증 가상 경로에 주문을 내주므로 보수 방향이 아님(B2 거부가 이미 fail-closed) · 선택지 3 은 a090 표면(중복 금지).
대체 산출물: 평가기 핀 시험 `internal/exitpolicy/a087_observed_price_pin_test.go`(`7bd8f197`, 변이 4/4 CAUGHT).
a087 은 Phase 2 사람 게이트(§0.7) 대기로 전환.

---

## 2차 proposal-freeze 재리뷰 처분에서 남긴 미해결 (2026-10-10)

`review.md` 「2차 proposal-freeze 재리뷰」 처분 중 문서 수정·구현 task 어느 쪽으로도 닫히지 않은 것. 처분 대조표는 `review.md` 끝.

### I-R1 — 셋째 관문 개방과 불변식 3(토글 OFF = upstream 동작) (**종결** · 2026-10-10 사용자 결정 2차)

1차 결정(2026-10-10)은 `internal/trading` `placeIntentSupported` 를 sell+market·가격 없음에 **무조건** 열려 했다. 3차 Eng
재리뷰(A-P1-1)가 판정했다: 이 술어를 지나는 인스턴스는 엔진·CLI 앱·**MCP 서버** 셋이고 ops `place_order` 는 MCP 카탈로그로
노출되는 에이전트 표면이라, 무조건 개방은 모든 토글이 OFF 여도 upstream 의 `ErrPlaceUnsupported` 를 브로커 전송으로 바꾼다 —
불변식 3 과 양립 불가. 1차 결정 기록도 범위를 「사람 CLI」로만 적어 실제 범위(에이전트 MCP)보다 좁았다.

**처분 (사용자 결정 2차, `review.md` 「사용자 결정 (2026-10-10, 2차)」)**: 개방은 엔진이 만드는 `trading.Service` 인스턴스에만
생성자 옵션으로 선다. CLI·ops·MCP 기본 생성 경로는 upstream 과 바이트 동일. 엔진이 꺼져 있으면 엔진 인스턴스가 없으므로
흔적 0 — 불변식 3 보존. 1차 결정의 사람 CLI upstream 동작 변경 수용은 철회. upstream 동일성은 tasks 1.5.5 가 시험(구조 시험
포함)으로 고정하고, 1.5.0 Pre-Edit 가 이 종결을 인용한다. 설계는 design D3 ③.

### I-R2 — `order-type-not-allowed` 422 분류 (미해결 · 무소유)

openapi `POST /api/v1/orders` 422 예시에 `order-type-not-allowed`(「현재 사용할 수 없는 호가 유형」)가 있다. 저장소 Go 코드
참조 0, `refusal_code.go` 목록에도 없다(재리뷰 P1-4). a087 이후 세션 밖 MARKET 보호 청산이 이 코드를 받으면 분류되지 않은
거부가 된다. 분류 필요 여부는 §5.2 결과 뒤 판정. 400/422 분류 전반(1차 리뷰 I2)과 함께 무소유.

### I-R3 — MARKET vs 하한가 지정가의 bps 재비교 (미해결 · 사람 결정)

2차 리뷰 「3차 교정」 4 는 두 방식을 청산 슬리피지 bps 데이터로 결정하라고 했다. 사용자 결정(2026-10-10)이 시장가 방향을
택했으므로 이 change 는 비교를 **미채택** 기록(design D6)으로 남기고 산출(tasks 3.8 — 기존 칸 유도, schema 없음)만 둔다. 착지 뒤
bps 데이터로 재비교할지와 그 임계는 정하지 않았다.

**3차 재리뷰 A-P2-3 정정**: 4판은 이 처분을 「사용자 기각」이라 적었으나 사용자 결정 기록에 원안은 나오지 않는다 — 「시장가 방향
사용자 결정에 따른 미채택, 원안 자체 평가는 미수행」이 정확한 기록이다. 재리뷰는 D6 표를 사용자에게 보이고 문언 기각을 받으라고
권고했다. 2차 결정으로 ③ 개방이 엔진 인스턴스 옵션 하나로 줄어 원안의 비교 우위(「③ 변경 불요」)도 작아졌다. 문언 기각을 받을지는
**사람 판단 — 0.9 Manager 대조 때 함께 묻는다**. 받지 않아도 §5.1 「거부」 행이 원안을 재검토 후보로 되살린다.

### I-R4 — 경계 조건의 미측정 두 줄 (미해결 · 범위 밖)

design D8: (가) VI 단일가 구간에서 MARKET 접수·체결 거동, (나) 1억 이상 주문에서 MARKET 의 금액 평가 기준
(`ConfirmHighValueOrder: false` → `400 confirm-high-value-required` 는 유형 무관). 둘 다 §5 측정 항목이 아니다.

### I-R5 — a100 쪽 반영 요청 (Manager — 이 디렉터리 범위 밖)

1. a100 design D8 계약 1·2 와 a087 MARKET 의 상호작용(design D9) — a100 착수 시 재대조 필요.
2. a100 `ma-runbook.md` R2 의 a087 5.1·5.2 실행 명령 공백(재리뷰 P1-1) — 실행 수단은 `tools/a087-market-sell-probe/`.
   이번 개정은 a087 디렉터리만 바꿨으므로 a100 문서는 손대지 않았다.

## 3차 proposal-freeze 재리뷰 처분에서 남긴 미해결 (2026-10-10)

`review.md` 「3차 proposal-freeze 재리뷰」 중 문서·도구 수정으로 닫히지 않은 것. 처분 대조표(3차분)는 `review.md` 끝.

### I-R6 — 도구 응답 해석의 계약 고정 없음 (미해결 · 한계 기록 · 3차 B-P3-5)

`tools/a087-market-sell-probe` 의 `readAnswer`(`send.go`) 봉투 모양 `{"result":{orderId,clientOrderId}}`·`{"error":{code,message}}` 은
mock 에서만 잰다. `contract_test.go` 는 **요청**만 openapi 원문·생산 직렬화기와 대조한다. openapi 200·409 예시와 모양이 일치함은
재리뷰가 읽어서 확인했다. 응답 모양이 다르면 도구는 orderId 를 못 읽어 `unknown` 으로 기운다(fail-closed — 이중 주문 아님). 응답 계약
시험을 더할지는 §5 실행 전 선택 사항으로 남긴다.

### I-R7 — 키가 다르면 둘째 실주문을 막는 것이 없다 (처분: 도구 잠금 대신 절차 · 3차 B-P3-4)

결과 `unknown` 뒤 사람이 미리보기를 다시 돌리면 새 멱등 키 = 새 실주문이다. 5.2 가 의도된 둘째 주문이라 전면 잠금은 맞지 않다.
같은 `--out` 에 `sending`/`no-answer`/`unknown` 영수증이 있으면 실행을 거절하는 가드는 값싸지만, 새 가드는 새 변이·시험을 부르고
전체 실측이 주문 2건이라 이 change 는 **절차로 대신한다** — tasks §5 「실행 전제」 3(재미리보기 전 `tossctl orders list` /
`completed` 확인 필수). 도구 가드가 필요하다고 사람이 판단하면 그때 더한다.

### I-R8 — 0.5 를 freeze 판정 앞에 둘지 (미해결 · Manager 판단 · 3차 A-P3-4)

0.5(엔진 세션 게이트 유무)는 지금 할 수 있는 read-only 작업이고, 그 답이 spec Requirement 1 의 SHALL NOT 을 무효화할 수 있다.
재리뷰는 freeze 를 0.5 뒤로 두는 편이 재개정 비용이 작다고 적었다(차단 아님 — 이미 §1~§3 착수 조건). tasks 0.5 에 권장으로 적었고
순서 결정은 0.9 Manager 대조에서 한다.

## tasks 3.5~3.8 read-only 분석에서 남긴 미해결 (2026-10-22)

### I-R9 — 부분체결 잔량이 비종결로 남으면 보호가 그 주문에 묶인다 (**이연 → a167** · 갭 · tasks 3.6)

MARKET 보호 청산이 부분체결된 뒤 잔량이 브로커에 비종결(PARTIAL_FILLED)로 남는 동안: 발의는 무장 유지(`ApplyExitFill` 은 종결에서만
`ResolvePending`), 같은 보호 행동의 재발의는 `SuppressedPending`(ratchet.go:423 · ladder.go:441)으로 억제, 억제된 판정은 발의가 비어
청소·제출·`noteDelay` 지연 경보 어디에도 닿지 않고(exitloop.go:1314), 무장 발의 알림(`noteAwaitingClose`)은 엔진 취소가 확정된
매도만 본다. 즉 **시간 상한도 경보도 없다.** 잔량이 브로커에서 종결(취소·체결)되면 체결 감지가 그것을 관측해 발의를 풀고 다음
관측이 남은 포지션을 다시 판정한다(근거·시험: `analysis/partial-fill-3.6/README.md`).

- 브로커 거동 `[미측정]`: KR 시장가 매도의 잔량이 즉시 취소되는지 잔존하는지. §5.1 은 1주 전량 체결 한 건뿐이다.
- 재가격·에스컬레이션은 무소유 장부(tasks 「후속 change」 2행). 지연 경보를 이 경로에 더할지는 사람 결정 — 이 change 는 production
  편집 0 으로 기록만 한다. LIMIT 보호 청산(오늘)도 같은 구조라 a087 이 만든 갭은 아니고, MARKET 이 「잔량이 남지 않는다」를 보장하지
  않는다는 것(design D8 1행)이 이 기록의 요지다.

### I-R10 — 부재 확증의 매수가능금액 갈래가 매도에서 무엇을 거르는지 미검증 (**종결 — 필요 없음** · 사용자 결정 2026-10-22 · tasks 3.7)

`execgw.Resolver.absenceCorroborated` B8(`indoubt.go:492` `notional > 0 && bpDelta < 0 && …`)은 시장가 intent(가격 NULL → `intentNotional`
0)에서 꺼진다(시험 `TestA087AnEmptyIntentPriceHasNoNotional`). 부재 확증은 그 앞 B7(보유수량 불변)에 기댄다. 지정가 **매도**에서 B8 이
실제로 무엇을 거르는지는 매도 접수가 매수가능금액을 어떻게 움직이는지가 `[미측정]` 이라 판정할 수 없다 — 유형 무관 기존 동작이며 a087
의 처분 대상 아님. 기록만.

## 사용자 결정 (2026-10-22) — D7 · I-R9 · I-R10

Manager 가 번호 매긴 질문으로 묻고 사용자가 답했다. 답 원문을 그대로 옮긴다.

1. **D7** — 「시장가가 안 되는 시간대에 손절을 어떻게 낼지. 권장: 그 시간대에는 지금처럼 지정가」
   → 처분: 브로커가 시장가를 받지 않는 구간에서는 보호 청산을 오늘과 같은 지정가로 낸다. spec Requirement 1 을 개정했다(예외 구간을
   명시). 설계와 불변식 4 근거는 design D7 「처분」에 있다. 엔진에 세션 게이트가 없으므로(0.5) 구간·접수 입력 배선이 구현 task 2.6 이 됐다.
   **해석 한 줄 (Manager 확인 대상, tasks 0.10)**: 「접수가 실측되지 않은 정규장 밖 구간」과 「구간 판정 불가」도 LIMIT 으로 둔다. 이것은
   사용자 문언이 아니라 불변식 4 에서 유도한 것이다. 모르는 구간의 MARKET 은 오늘의 LIMIT 보다 약해질 수 있다.
2. **I-R9** — 「남은 잔량 경보. 권장: 가격을 자동으로 다시 정해 재주문」
   → 사용자는 경보만이 아니라 **자동 재가격·재주문**을 골랐다. 이것은 청산 경로에 새 주문 행동을 더하는 High-risk 변경이다. a087 범위 밖이므로
   새 change `a167-a-protective-remainder-is-re-priced`(`STORY-TOS-a167`)로 이연했다. 이 로트는 그 change 를 **proposal 수준으로만**
   정의했다(proposal · tasks 골격 · 초안 spec delta · PM 등록). 발동 시간 N 은 사람 결정으로 남겼다. design D10, tasks 「후속 change」 a167 행.
3. **I-R10** — 「매도 쪽 매수가능금액 확인: 필요없음」
   → I-R10 을 **필요 없음으로 종결**한다. 측정하지 않는다. 기록된 한계(지정가 매도에서 B8 이 무엇을 거르는지 `[미측정]`)는 사실 그대로 남는다.
   부재 확증은 B7(보유수량 불변)에 기댄다는 위 기록도 그대로다.

이 결정 반영은 문서만 바꿨다. production 코드 · 주문 · 계좌 조회는 0 이다.
Function Logic Map: not-applicable — 문서 개정만 했다. 새 분기 주장은 0.5 AST(`analysis/session-gate-0.5/`)와 3.6 AST(`analysis/partial-fill-3.6/ast/`)
산출물에서만 인용했다. 입력 출처 함수의 분기는 아직 주장하지 않았다(2.6.1 이 AST 로 먼저 연다).
