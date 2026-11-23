# a198 Design

## 1. 결정

| # | 결정 | 근거 |
|---|---|---|
| D1 | 길 치우기의 취소 결정 = 종류 `CANCEL`, 등급 위험 감소, 방향 = **취소할 주문의 방향**, 최대 수량 = 그 주문의 수량, 사유 문장은 종전 것(「cancelling the working … order … before liquidating position …」). 수동 flatten 의 취소 결정과 같은 필드 구성 | probe 대조군이 접수된 모양이다. 방향을 매도로 고정한 종전 결정은 매수 주문의 취소에 대해 내용도 틀렸다(「매도 감소」가 아니라 「그 주문의 철회」) |
| D2 | **Manager 결정 2026-11-23 — (가-B)**: 발행 자리는 관측기 안이다. `clearTheSymbol` 이 관측기가 이미 가진 값(`opts.Journal` · 관측기 시계 `o.clk` · `opts.AccountRef`)으로 `execgw.Issuer` 를 만들어 flatten 과 **같은 함수** `Issuer.IssueReduction`(종류 `CANCEL`)을 부른다. TTL·ID 생성기는 비워 둬 Issuer 기본값(`execgw.DefaultDecisionTTL` 60s · `execgw.randomID` 128비트)을 쓴다. 새 옵션 칸·새 인터페이스·`Context.ExitObserver` 편집·시험 하네스 변경 없음. 기존 `ReductionIssuer`(제출용)는 그대로 | 고칠 기존 함수가 가장 적고(analysis/edit-surface.md §3), 결정 기록이 flatten 과 같은 함수에서 나오며, a087 번들이 고정한 exitwiring.go 를 건드리지 않는다. 선행 확인(analysis/claim-check.md 「D2 선행 확인」): 운영에서 관측기 시계와 게이트웨이 시계는 같은 `clock.System()` 값 하나다. 취소 결정에 운영 모드·킬 스위치 검사가 걸리지 않음은 AST 로 확인(claim-check (a)) |
| D3 | **삭제 — Manager 결정 2026-11-23.** (가-B) 에서는 조립이 공급할 발행기가 없다(이미 필수인 Journal·AccountRef 에서 유도). 새 기동 실패 조건도 생기지 않는다 | r1 P2-5 의 「조립 실패가 엔진 런타임 전체를 막는다」는 이 삭제로 소멸 |
| D4 | **Manager 결정 2026-11-23(r2-1 뒤 재결정) — 같은 싱크에 새 함수가 직접 쓴다.** `clearTheSymbol` 의 취소 결정 발행 실패(B9)와 취소 제출의 비확정 결과(B11 — 오류 또는 `CONFIRMED` 아님)마다 새 파일의 새 함수(`noteClearCancelFailure`)가 `opts.ExitOrderLog`(`*obs.Logger`, exitloop.go:223 — 운영 배선 cmd/tossctl/engine.go:680)에 **직접** 한 줄을 쓴다. 로거 nil 이면 아무것도 안 함. a129 의 대기열·중복 억제·줄 타입(`noteExitOrder`·`endExitOrderCycle`·`writeExitOrder`·`exitOrderLine.args`)은 쓰지 않고 무편집. 줄 = 이벤트(a129 의 submit 과 같은 기준: 발행 실패·브로커 전 거절 `obs.EventOrderNotDispatched`, `FAILED_CONFIRMED` `obs.EventOrderRejected`, IN_DOUBT `obs.EventOrderInDoubt` — 새 obs 타입 없음) · 종류 취소 · 포지션 id · 종목 · 시장 · 대상 주문 번호 · 결과 상태 · 사유 **코드**(`exitOrderToken` 거름). 발행 실패는 상태 없이 고정 사유 토큰. Detail · 원문 오류 · 계좌 · 수량 · 가격 금지. 중복 억제는 새 함수 자신의 메모리 표시(key = 포지션·대상 주문 번호·결과 상태·사유 코드, 관측기 구조체의 새 맵 필드 — 새 함수 안에서 nil 이면 만듦): 같은 실패가 주기마다 반복되면 한 번, 사유·상태가 바뀌면 다시, 그 주문의 취소가 접수 확정되면 그 주문의 표시를 지움. `res.cleared = false` 판정과 그 뒤 흐름은 한 글자도 바꾸지 않는다 | `warnEpisode`(운영 무기록)·`opts.Log` 연결(계좌 노출)·a129 경로(줄 타입의 SELL·LIMIT 고정 출력, 래치 서명에 주문 번호 없음 — r2-1) 모두 기각. 편집하는 기존 함수는 `clearTheSymbol` 하나 |
| D5 | 손절·익절 **제출** 경로(`submit` 의 `IssueReduction`)는 무편집 | 그 경로는 운영에서 동작한다(PLACE CONFIRMED 111 건) |
| D6 | 토글 없음 | 정본이 이미 요구하는 동작의 결함 수리다. 꺼 둘 「종전 동작」이 「손절 무기한 보류」다 |
| D7 | **Manager 결정 2026-11-24 — 이미 끝난 주문을 이유로 한 취소 거절은 확정 거절.** 취소 요청의 응답 본문 `code` 가 `already-filled`·`already-canceled`·`already-rejected`·`already-modified` 이면 그 CANCEL attempt 는 IN_DOUBT 가 아니라 확정 거절(취소 미성립)로 종결한다. 판정은 정본 「브로커가 이름을 준 거절은 code가 분류한다」와 같은 규칙 — 최상위 `code` 와 `error.code` 두 자리만, 대소문자 무시 전체 일치, 두 자리 값이 다르면 모호 강제, status·message·부분문자열로 걸지 않음. **CANCEL 종류에만**(PLACE·AMEND 응답의 같은 문자열은 이 분류 밖). HTTP status 목록에 409 를 넣지 않는다(`409 request-in-progress`·code 없는 409 는 계속 모호). 자리(증거 analysis/d7-d8-evidence.md §2): CANCEL 만 가를 수 있는 기존 자리는 `Gateway.withCancelRefusal` 하나 — 그 함수 1 + 새 판정 함수(새 파일, `refusalCodes` 판독 재사용). a165 의 `already-processing` 판정과 code 가 서로소라 재시도 창에 들어가지 않는다 | 리뷰 r1 P0-1 실측: 계약상 이 code 들은 409 로 오고(openapi.latest.json:6565-6612, 자리는 `error.code`), 409 는 확정 status 목록(journal/dispatch.go:351) 밖이라 CANCEL attempt 가 IN_DOUBT → 같은 종목 미종결 가드가 손절·flatten 을 막고 CANCEL 은 세션 중 해소되지 않아 재시작까지 막힘(편집 전에는 취소가 원장에서 거절돼 나가지 않았으므로 종결 기록 뒤 손절이 나갔음 — 이 change 가 만든 회귀). 「그 주문은 더 이상 취소할 수 있는 상태가 아니다」 = 취소 미성립은 확정이고, 주문의 운명은 체결 감지가 원장에 적는다 → 치움 미완료로 종결 기록 대기(일시 보류). 실물 응답은 미측정(계약 문서만) |
| D4a | **Manager 결정 2026-11-24(구현 리뷰 r1 P2) — D4 보강.** 수량·가격 해석 실패 갈래(B9)에도 실패 줄을 한 번 쓴다(고정 사유 토큰). 중복 억제 표시는 로그 쓰기 **뒤에** 세운다(쓰기가 패닉해 삼켜지면 표시가 남지 않아 다음 주기에 다시 쓴다) | spec delta 셋째 Requirement 「발행이나 제출이 실패하면 기록」 문언과 B9 무기록의 틈 · 표시를 먼저 세우면 쓰기 패닉 뒤 그 실패가 다시 기록되지 않음 |
| D8 | **Manager 결정 2026-11-24 — 엔진의 취소가 이미 접수 확정된 매수는 다시 취소하지 않고 치워진 것으로 본다.** `clearTheSymbol` 이 매도에 쓰는 기존 판독 `ConfirmedCancelOf` 를 매수에도 적용해, 접수 확정된 엔진 취소가 있는 매수는 취소를 내지 않고 `res.cleared` 에 영향을 주지 않는다(매도는 종전대로 종결 증거 대기). 편집 함수는 `clearTheSymbol`(이미 편집 대상) | 정본 「매수 주문의 취소는 취소 접수로 치운다」의 귀결 — 접수된 순간 치워졌고 그 성질은 다음 주기에도 같다. 리뷰 P2·P3 모양(접수된 매수의 재취소 → 409 already-canceled)을 원천에서 없앤다. `ConfirmedCancelOf` 의 거래일 결속(전일 주문)은 기존 잔여 그대로(그때는 재취소 → D7 확정 거절 → 일시 보류) |

## 2. 이 수리가 처음으로 실행 가능하게 만드는 경로 — 달라지는 운영 동작

오늘은 취소가 원장에서 거절되어 그 뒤의 코드가 운영에서 한 번도 돌지 않았다. 이 change 뒤 처음 도는 것:

1. 엔진 귀속 **매수** 미체결이 있는 종목의 손절: 취소 접수 → 그 주기에 치움 완료 → 손절 제출.
2. 엔진 귀속 **매도** 미체결(무장된 발의의 주문 — 익절 지정가나 앞선 손절)이 있는 종목의 손절: 취소 접수 → 「종결 증거 대기」(정본 — 매도의 취소 접수는 치움의 증거가 아니다) → 체결 감지가 그 주문의 종결을 기록한 뒤 손절 제출.
3. 취소가 모호(IN_DOUBT)로 남는 경우: 같은 종목의 미종결 attempt 가드가 손절을 보류한다(정본 — 미종결 attempt 가 있으면 치우기는 완료가 아니다). `CANCEL` 은 세션 중 해소되지 않으므로 재시작 복구까지 보류가 이어질 수 있다 — **기존 정본 규칙의 귀결이고 이 change 가 만든 것이 아니지만, 이 change 가 그 경로를 처음 열므로 공시한다.** 기존 지연 경보와 연속 실패 트리거가 그것을 알린다(tasks 3.x 에서 그 경보가 실제로 나가는지 시험).

4. 취소 1건 = 사전 단건 조회 1 + 취소 요청 1, 둘 다 관측 루프 안에서 동기(Manager 결정 2026-11-23 — 정본의 「브로커 조회를 더하지 않는다」는 길 치우기가 **대상을 찾으려고** 브로커를 읽지 않는다는 뜻이고, 사전 단건 조회는 주문 경로의 기존 계약이다; retry matrix 행 37). 사전 조회 시한 2s(`precheckTimeout`, internal/app/engine/precheck.go:58), 취소 요청은 HTTP 클라이언트 시한 15s(`defaultTimeout`, internal/official/client.go:22·137). 대상 주문 수 × (2s + 15s) 는 **HTTP 시간 예산**일 뿐 최악 지연 상한이 아니다 — 토큰 획득(internal/official/client.go:336 `tm.token(ctx)` → token.go:63 context 로 끊기지 않는 `mu.Lock()`, 다른 루프의 갱신이 잠금을 쥐면 2s 시한이 지나도 대기)·401 재전송(최대 2회)·a166 계좌 예산 admission 대기·원장 처리는 계산 밖이다. **전체 최악 지연은 미확정·미측정**(r2 P1-2 정정).
5. 매수 취소 접수 뒤 같은 주기의 손절이 브로커에서 반대 미체결(`opposite-pending-order-exists`)로 거절되는 경우(브로커가 취소를 늦게 반영). **정정(2026-11-24, 구현 적대 리뷰 r1 P0-1)**: 2026-11-23 처분은 「다음 주기의 재취소가 거절되면(확정 거절) 종결 기록까지 일시 보류」를 전제했으나 틀렸다 — 계약상 그 재취소는 **409 `already-canceled`** 로 오고, 409 는 확정 status 목록 밖이라 CANCEL attempt 가 IN_DOUBT 가 되어 그 종목의 모든 변경이 **재시작까지** 막혔다(리뷰 탐침 P2 실측, 편집 전 base 에서는 통과). D8 뒤: 접수 확정된 매수는 다시 취소하지 않고, 다음 주기에 손절을 다시 낸다 — 브로커가 취소를 반영할 때까지 손절이 주기마다 `opposite-pending-order-exists`(확정 거절, 종목을 잠그지 않음)로 거절될 수 있다. 취소 접수 뒤 그 매수가 체결에 이기면: 손절 접수 **전**이면 보유가 늘고 같은 주기 손절은 그 이전 수량이라 과소 보호가 남을 수 있고, 손절 접수 **뒤**(CLOSING)면 프로젝션은 `ENTRY_WHILE_CLOSING` 거절로 RECONCILE 로 간다(보강 리뷰 r1 P2-4 정정 — 회복은 대사 경로). 어느 쪽도 초과 매도 방향은 아니다(미측정, d7-d8-evidence.md §6).
6. 취소 대상이 이미 체결·취소·거부·정정됐으면(리뷰 탐침 P1·P3) 계약상 409 `already-*` — D7 뒤 확정 거절 → 치움 미완료 → 그 주문의 종결 기록이 들어온 뒤 주기에 손절.

2·3 의 기존 규칙들이 실제 게이트웨이 위에서 문서대로 도는지는 이 change 의 결합 시험이 처음 잰다. 어긋남이 나오면 고치지 말고 Manager 에게 보고한다(범위 판단).

## 3. 거절·침묵하는 정상 입력

새 거절은 더하지 않는다. 종전에 조용히 실패하던 취소가 나가고, 실패하면 운영 로그에 기록된다. D3 삭제로 새 기동 실패 조건도 없다.

D7 이 **모호 → 확정 거절**로 바꾸는 응답(이것만): CANCEL attempt 의 응답 오류가 공식 클라이언트 오류이고 본문이 JSON 객체이며, 최상위 `code` 또는 `error.code` 의 값이(대소문자 무시 전체 일치) `already-filled`·`already-canceled`·`already-rejected`·`already-modified` 이고, 두 자리가 함께 있으면 값이 같은 경우. 계약의 실제 모양은 `{"error":{"requestId":…,"code":"already-filled","message":…}}`(openapi.latest.json:6575-6580).

바꾸지 않는 것: `409 request-in-progress`(발주 전용 계약 — 취소에 오면 그대로 모호) · code 없는 409 · 두 자리 값이 다른 409(모호 강제) · `already-processing`(a165 경로 그대로 — 엔진은 `broker_rejected` 확정 거절, flatten 은 재시도 창) · `already-filled-v2` 처럼 접미가 붙은 값 · message 에만 그 문자열이 있는 응답 · PLACE·AMEND 응답의 같은 문자열 · 정책이 없는 게이트웨이(운영 0, 시험만 — d7-d8-evidence.md §2).

## 4. 시험 전략 (RED 먼저)

새 시험 파일 `internal/app/engine/a198_clearing_cancel_test.go`. **가짜 제출자를 쓰지 않는다** — 실제 `execgw.Gateway` + 실 원장(임시 경로) + 가짜 브로커 + 운영과 같은 Guardian(`RiskGuardian`, 제출용). 기존 실제-게이트웨이 하네스(`buildE2EStack`, exit_e2e_test.go:267-277)를 재사용할 수 있다. 실패 기록 단언은 `ExitOrderLog` 에 넣은 싱크의 줄로 한다. 픽스처: 계좌·주문번호·수량·가격은 합성, 종목 코드는 저장소 관례(`005930`, 101 개 시험 파일이 쓰는 값 — 구현 리뷰 r1 처분).

1. **RED(결함 재현)**: 엔진 귀속 매수 미체결 1 + 보호 손절 발의 → 현재 HEAD 에서 브로커 취소 0·손절 미제출임을 단언 실패로 남긴다(`analysis/tdd/red.txt` — 원장 거절 문구 포함).
2. 엔진 귀속 매수 미체결 → 브로커 취소 1, 원장의 attempt 가 `kind=CANCEL`·`CONFIRMED`·대상 주문 번호를 갖고, 그 결정 행에 멱등 키가 없으며 preimage 의 방향이 주문 방향(매수) — 결정의 종류는 원장에 없다(키 유무로만 남음), 같은 주기에 손절 PLACE 제출·접수.
3. 엔진 귀속 매도 미체결(무장된 발의의 주문) → 브로커 취소 1(결정 방향 매도), 그 주기 손절 미제출(종결 증거 대기), 종결 체결 스냅숏 기록 뒤 주기에 손절 제출. 접수 확정된 취소를 다시 내지 않음(브로커 취소 여전히 1).
4. 취소 제출이 브로커에서 거절 → 손절 미제출, 그 로거에 줄 정확히 1(종류 취소·결과 상태·사유 코드·대상 주문 번호), 다음 주기의 같은 거절 반복에도 여전히 1, 사유가 바뀌면 2, 줄에 계좌·수량·가격·Detail 원문 0.
5. 취소 결정 발행 실패(원장 쓰기 오류 주입) → 손절 미제출, 그 로거에 줄 정확히 1(고정 사유 토큰, 상태 없음), 반복에도 1.
6. 취소가 5xx 로 IN_DOUBT → 다음 주기 손절 미제출(미종결 가드), 기존 지연 경보가 한계 뒤에 나감(가짜 시계)과 연속 실패 경보가 각각 나감. 경보는 관측 주기가 다시 `noteDelay` 에 닿아야 나가므로(첫 호출은 시계만 시작 — exitloop.go:1925) **전체 관측 주기를 반복해** 알림 기록으로 단언한다(`noteDelay` 직접 호출 금지). 변형: 취소 attempt 가 **ACKED 로 고착**(접수 뒤 최종 쓰기 전 중단을 원장으로 재현) → 손절 미제출, 연속 실패 계수 제외, **지연 경보(30 초)가 실제로 나감**을 단언 — 안 나가면 고치지 말고 Manager 에게 보고(범위 판단이 바뀐다).
7. 조립: 운영 조립(`Context.ExitObserver`)으로 만든 관측기가 실제 게이트웨이로 2 를 구동한다(운영 조립이 지도자인 실행). 그 관측기의 `ExitOrderLog` 가 nil 이 아님도 단언한다.
8. 손절·익절 제출의 결정은 종전과 같은 발행(`IssueReduction`, 종류 `PLACE`)임을 원장으로 단언(회귀).
9. 기존 a094 길 치우기 시험·a129 줄 시험 전부 무편집 통과((가-B)는 새 인터페이스가 없어 하네스 변경 0, a129 경로 무편집).

**보강(2026-11-24, D7·D8 · 구현 리뷰 r1)** — 실제 게이트웨이, 가짜 브로커가 409 `already-*` 를 계약 모양(`error.code`)으로 답함:

10. (리뷰 탐침 P1) 손절 발의 때 그 매수가 브로커에서 이미 체결 → 첫 취소가 409 `already-filled` → CANCEL attempt 가 `FAILED_CONFIRMED`(IN_DOUBT 아님), 그 종목의 미종결 attempt 0, 손절은 그 주기 미제출 → 그 매수의 종결 체결 기록 뒤 주기에 손절 제출. 편집 전(base)과 같은 결말을 단언.
11. (P2) 매수 취소 접수 → 그 주기 손절이 409 `opposite-pending-order-exists` 로 거절 → 다음 주기 **취소 요청 0**(D8) · 손절 재제출.
12. (P3) 매수 둘 — 하나 400 거절·하나 접수 → 다음 주기 접수된 쪽 재취소 0(D8), 거절된 쪽만 재시도.
13. 분류 단위 시험(execgw 내부): 최상위 `code` · `error.code` 각각 · 대소문자(`ALREADY-FILLED`) → 확정 거절; 두 자리 값 불일치 → 모호; `already-filled-v2`·message 에만 있음 → 모호; 같은 본문이 PLACE·AMEND 결과면 → 종전 결과 그대로; `409 request-in-progress`·code 없는 409 → 모호; `already-processing` → a165 경로 그대로(엔진 `broker_rejected`, flatten 재시도).
14. D8 단독: 접수 확정 취소가 있는 매수 → 다음 주기 취소 요청 0, `cleared` 유지. 매도는 종전대로 대기(시험 3).
15. (리뷰 P1-2) 취소 결정만 실패시키는 주입(손절 자신의 결정은 기록됨) → `cleared` 거짓 유지를 단언(변이 「발행 실패에도 cleared 유지」를 죽임).
16. (리뷰 P2) 수량 해석 실패 갈래 → 실패 줄 1 · 손절 미제출; 접수 확정 뒤 표시 정리(같은 주문이 나중에 다시 실패하면 줄 다시 1); 결정의 최대 수량이 정확히 주문 수량(원장 preimage `max_quantity`); withPending=false 매도 경로 1 개(실제 게이트웨이).
17. (보강 리뷰 r1 P2-5 — 결정에 묶이지 않는 공백만) 네 `already-*` code 각각(10 은 `already-filled` 하나) · 읽을 수 없는 code 값 · 같은 본문이 AMEND 결과면 미적용 · 정책 nil 게이트웨이 · D8 에서 다른 계좌·시장·종목·거래일의 취소 증거는 「접수 확정」으로 읽지 않음과 `ConfirmedCancelOf` 조회 오류 · D4a 의 **가격** 해석 실패 · 첫 로그 쓰기 패닉 뒤 다음 주기 재기록(「표시를 쓰기 앞에」 변이의 관측점). `already-modified`·재전송 묶음(`already-processing → already-canceled`)·늦은 매수 체결 시험은 Manager 결정 뒤(보강 리뷰 r1 P1-1·P2-2·P2-4).

변이(전부 CAUGHT, 변이마다 닿음 확인, 사본 하네스·무변이 대조군 GREEN·시작 sha 단언): 취소에 종전 `IssueReduction` 결정을 싣기(= 결함 복원 — 2·3·7 이 죽어야 함) · 결정 방향을 매도로 고정(2 가 죽어야 함) · 결정에 멱등 키가 붙는 발행으로 되돌리기 · 실패 기록 삭제(4·5) · 실패 기록의 중복 억제 삭제(매 주기 한 줄 — 4) · 실패 줄에 Detail 원문 싣기(4) · 매도 취소 접수를 치움으로 판정(3 이 죽어야 함 — 기존 규칙이 실제 경로에서 지켜지는지; 1 판에서 A198 만으로는 생존·A094 로 CAUGHT — 16 의 withPending=false 시험이 실제 경로로 죽여야 함). 보강: D7 을 PLACE 에도 적용 · 부분문자열 일치로 완화 · status 409 전체를 확정 거절로 · D7 삭제(10 이 죽어야 함) · D8 삭제(11·12) · **D8 을 매도에도 「치움」으로 확대(초과 매도 방향 — 3 이 반드시 죽어야 함)** · 발행 실패에도 cleared 유지(15) · 최대 수량 ×2(16) · 표시를 쓰기 앞에(쓰기 패닉 시 재기록 안 됨 — 16).

선택 시험: `./internal/app/engine -run 'A198|A094|ExitClear|ExitObserver'`, `./internal/execgw -run 'Cancel|RiskGuardian|Issue'`, `./internal/journal -run 'Prepare|Decision'`. `-race` 는 `-run A198` 만. 전체 스위트는 종결 게이트에서만.

## 5. 운영 (사람 승인)

배포 뒤 첫 길 치우기 취소가 일어나면 원장의 `kind=CANCEL` attempt 와 그 뒤 손절 접수 시각을 `review.md` 에 시각·상태만 기록한다. 발생을 인위로 만들지 않는다(LIVE 주문 금지).

## 6. 잔여·후속

- §2-3 · ACKED(Manager 결정 2026-11-23 — 범위 밖, 이름 붙인 잔여): 취소가 ACKED 또는 IN_DOUBT 로 남으면 그 종목의 손절은 보류되고, ACKED 는 연속 실패 계수에서도 빠진다(정확히는: 같은 종목의 미종결 attempt 가 치울 대상 **전부**의 진행 중 취소뿐일 때 — `onlyEngineCancelsInFlight`; 대상이 종결 기록으로 빠지거나 다른 대상이 남으면 다시 센다). 지연 경보(30 초)는 나간다. 지금 빌드에는 ACKED 로 남은 취소를 닫는 운영자 경로가 없다 — 후속 change 『CANCEL attempt 의 복구 계약(IN_DOUBT·ACKED)』의 대상이고, a197 재개의 선행이기도 하다. (근거: 기동 복구 internal/reconcile/acked_boot.go:51 「CANCEL/AMEND … stays ACKED」, 운영자 해소 internal/journal/resolution.go:143 은 UNRESOLVED_IN_DOUBT 에서만, 세션 해소는 PLACE 만.)
- 전일 주문의 거래일 결속(범위 밖 — **알려진 정본 미충족 잔여**: 정본의 「확정된 취소를 다시 내지 않는다」와 delta 의 「한 번 나간다」를 이 경우 충족하지 않는다): `ConfirmedCancelOf` 는 취소 intent 의 거래일 = 주문 거래일을 요구한다 — 전일 주문을 오늘 취소하면 대기가 성립하지 않아 확정된 취소가 주기마다 다시 나간다. 엔진 주문은 당일 주문이라 거래일을 넘겨 원장에 「살아 있는」 주문은 브로커에서 이미 만료됐을 가능성이 높고, 그 정리는 체결 감지(a191)의 영역이다. 고치지 않는다.
- 매수 취소 접수 뒤 재취소(§2-5): **D8 로 대체(2026-11-24)** — 재취소는 없어지고, 남는 것은 브로커가 취소를 반영할 때까지 손절의 반대 미체결 거절 반복과, 취소 접수 뒤 체결 경합의 과소 보호(미측정).
- D7 의 근거는 계약 문서뿐이다 — `already-*` 409 의 실물 응답은 저장소에 측정 기록 0(d7-d8-evidence.md §1).
- 확정 거절(400·404·422·D7 의 409)은 매 주기 재시도되어 그 주문의 종결 기록이 올 때까지 주기마다 결정 행·CANCEL attempt 행이 쌓인다(구현 리뷰 r1 — 기록만).
- 정정된 주문의 낡은 수량·가격: 엔진에 정정 경로가 없어 엔진 주문의 정정은 사람이 한 것뿐이다(구현 리뷰 r1 — 기록만; `already-modified` 의 처분은 Manager 결정 대기, d7-d8-evidence.md §5).
- a087 과의 순서: a198 은 기다리지 않는다. `clearTheSymbol`(exitloop.go)이 a087 의 gate ⑤ 창에 들어가는 것은 a087 종결 쪽의 기존 비용(착지 기록 또는 번들)이다. 참고: a087 번들의 `source_sha256` 은 파일 해시다 — exitloop.go 쪽은 이미 HEAD 와 다르고(과거 커밋 5c0aeb2b8 과 일치), exit_order_log.go 쪽(`exitOrderLine.args`·`exitOrderLineFor`)은 지금 HEAD 와 같고 재결정된 D4 로 무편집이라 충돌 없음. exitwiring.go 는 (가-B) 로 무편집.
- 과거 운영에서 이 결함이 실제 손절을 막았는지는 미확인이다.
- 「가짜가 생산 계약을 가린다」: 주문 경로의 다른 결합(정정 등)에도 같은 공백이 있는지는 보지 않았다.
