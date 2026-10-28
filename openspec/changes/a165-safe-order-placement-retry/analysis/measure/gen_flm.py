#!/usr/bin/env python3
"""a165 0.4 FLM/BTM 생성기 — ast.json + 소스 줄 + 패키지 coverprofile 로 두 Markdown 을 씀.

분기의 조건 칸은 AST 가 준 줄의 원문이고, 실행 칸은 coverprofile 에서 그 분기 줄에서 시작하는 블록의
실행 수다(손으로 고르지 않음). 함수별 역할·불변식 산문만 NOTES 에 손으로 적음.

사용: python3 gen_flm.py <change-dir> <go-coverprofile>...   (저장소 루트에서 실행)
"""
import json
import sys
from pathlib import Path

NOTES = {
    "Gateway.submit": dict(
        role="게이트웨이 순서(결정 읽기 → Prepare → 거절 → DISPATCH_STARTED → 직전 재검증 → 브로커 1회 → 정착)의 본체. "
             "a165 는 dispatch 클로저를 고치지 않고, B49 갈래가 `DispatchVerified` 에 넘기는 함수만 바깥 래퍼로 감싼다(design D5, 사람 결정 P2-2).",
        inputs=[("plan mutationPlan", "place/cancel/amend 계획", "Gateway.place·Cancel·Amend", "prepareRequest 오류면 기록 없이 반환"),
                ("ref GuardianDecision", "결정 id", "journal", "읽기 실패는 refusePrepared 로 기록"),
                ("plan.strategy", "nil 또는 paired KR/US first-leg 능력", "PlaceClaimedStrategy", "nil 이면 일반 경로")],
        calls="`plan.call` 은 클로저 `call` 안의 한 자리(708)에서만 불리고, `call` 은 B44(`withStrategyEntryGateAuthority` 의 fn, 712) "
              "또는 B45(717) 중 정확히 한 갈래에서 한 번 불린다. 함수 전체에 for/range/select/go 가 0 이다 — 즉 지금은 "
              "attempt 당 브로커 호출이 정확히 1회다. 결과는 725 의 `classifyMutation(tracker.State(), result, callErr)` 한 자리에서 분류된다.",
        mutations="journal Prepare(RECORDED) · DispatchVerified/DispatchStrategyVerified(DISPATCH_STARTED → 정착) · 심볼 claim/release. "
                  "재시도가 생기면 같은 attempt·같은 nonce·같은 clientOrderId 로 브로커 호출이 늘 뿐 journal 행은 늘지 않는다(design D1).",
        safety="a165 편집 경계(3판 보완, P2-2): dispatch 클로저(B29–B48, 636–726)·클로저 밖 거절 순서(B6–B28)·정착(B53–B57)은 바꾸지 않는다. 편집은 B49 갈래 729 의 인자 식 하나(`dispatch` → `g.withCancelRefusal(plan, dispatch)`)뿐이고, 정책 nil 이거나 취소가 아니면 그 식의 값이 원 `dispatch` 다. 3판의 `sendOnce` 추출은 폐기. High-risk: yes."),
    "Gateway.place": dict(
        role="place 계획을 만들고 `plan.call` 클로저(394–397)를 정의한다. 클로저는 호출마다 같은 `keyed(key)` 로 PlaceIntent 를 만들므로 "
             "같은 key 면 같은 본문이다.",
        inputs=[("req.Intent", "PlaceIntent", "호출자", "submit 이 검증"), ("strategy", "nil 또는 능력", "PlaceClaimedStrategy", "—")],
        calls="`g.trading.Place` 는 클로저 안에서만 불리고 클로저는 submit 이 부른다.",
        mutations="상태 없음 — 계획 구성만.",
        safety="a165 는 이 함수를 바꾸지 않을 계획(재시도 판정은 submit 쪽). High-risk: yes(주문 경로)."),
    "Gateway.Cancel": dict(
        role="cancel 계획. `plan.call` 은 `g.trading.Cancel` 이고 trading.Service.Cancel 이 GET 사전 확인(GetOrderAvailableActions) 뒤 POST 한다.",
        inputs=[("req.Intent.OrderID", "브로커 주문 id", "호출자", "submit 이 검증")],
        calls="`g.trading.Cancel` 1회(클로저).",
        mutations="상태 없음 — 계획 구성만.",
        safety="편집 계획 없음(409 재시도 판정은 submit 쪽). High-risk: yes."),
    "Gateway.Amend": dict(
        role="amend 계획. 생산 호출자는 CodeGraph 에서 0(ops `modifyHandler` 와 시험만).",
        inputs=[("req.Intent", "AmendIntent", "호출자", "submit 이 검증")],
        calls="`g.trading.Amend` 1회(클로저).",
        mutations="상태 없음.",
        safety="편집 계획 없음. High-risk: yes."),
    "classifyMutation": dict(
        role="브로커 호출 1회의 결과를 journal 분류(ACKED/NOT_SENT/REJECTED/AMBIGUOUS)로 바꾼다. a165 의 갈래표(미접수 확정 / 결과 불명 / 확정 거절)는 "
             "이 함수의 반환 지점 위에 얹는다(design D2) — 이 함수 자체는 바꾸지 않는다.",
        inputs=[("send journal.SendState", "NotStarted/Partial/Complete", "WithSendTracker", "—"),
                ("err", "official 오류", "Client.send", "nil 이면 B1 ACKED")],
        calls="429 는 B8(statusOf→429)→`ClassifyHTTPMutation` 의 default(AMBIGUOUS). 5xx 도 B8(500)→AMBIGUOUS. 상태 없는 전송 오류는 "
              "98 의 `ClassifyHTTPMutation(send,0,err)` — send==NotStarted 면 NOT_SENT, 아니면 AMBIGUOUS. 409 `*APIError` 는 B3 code 판정 없음 → B8 → AMBIGUOUS.",
        mutations="없음(순수).",
        safety="편집 계획 없음. 갈래 판정의 입력으로만 씀."),
    "statusOf": dict(
        role="official 오류에서 HTTP status 를 복원. 429→`ErrRateLimited`(B4), 5xx→`ErrServer`(B5), 401/403→B3.",
        inputs=[("err", "official 오류", "Client.send", "모르면 (0,false)")],
        calls="errors.As/Is 만.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "Client.send": dict(
        role="모든 verb 의 인증 요청 흐름. B4 의 for 루프가 **기존 401 재전송**이다(사람 결정 3 이 「확실히 미접수」로 인정한 자리).",
        inputs=[("makeReq", "토큰을 받아 요청을 만드는 클로저", "postWithHeaders 등", "오류면 그대로 반환"),
                ("ctx", "WithSendTracker 가 붙인 httptrace 포함", "submit", "—")],
        calls="B4 조건 `attempt < 2 && code == http.StatusUnauthorized` — 401 일 때만 돈다. 루프 안 `c.tm.refresh` 후 `makeReq(tok)` 로 **다시 만든** 요청을 "
              "`doRequest` 한다. B8 `!adopted` 면 break — 그래서 401 재전송은 최대 2회(첫 refresh 가 채택 토큰이면 2회, 아니면 1회). "
              "postWithHeaders 는 본문을 `json.Marshal` 로 루프 **밖에서 한 번** 인코딩하고 makeReq 가 그 바이트로 `bytes.NewReader` 를 만들므로 재전송 본문(같은 clientOrderId)은 바이트 동일하다(소스 client.go 427–448).",
        mutations="토큰 캐시 갱신(refresh).",
        safety="편집 계획 없음 — a165 는 시험으로 근거만 고정(사람 결정 3). 401 을 받은 POST 가 실제 미접수라는 실측은 없다."),
    "Gateway.withStrategyEntryGateAuthority": dict(
        role="strategy first-leg 의 최종 진입 게이트. `g.entry.mu` 를 잡은 채(63 defer Unlock) `fn()` 을 80 에서 **한 번** 부른다 — 브로커 호출 동안 EntryGate 잠금이 잡혀 있다.",
        inputs=[("expected", "관찰된 게이트 권한", "ObserveStrategyEntryGate", "다르면 거절"), ("fn", "submit 의 call 클로저", "submit", "nil 이면 거절")],
        calls="fn 은 반환문(80)에서 1회. 루프 없음.",
        mutations="잠금 획득/해제만.",
        safety="편집 계획 없음. 재시도의 대기를 이 잠금 안에 두면 EntryGate 를 쓰는 다른 경로(Block 등)가 최대 대기만큼 멈춘다 — design D1 은 재전송마다 이 함수를 새로 통과하고 대기는 잠금 밖에 둔다."),
    "ClassifyHTTPMutation": dict(
        role="journal 의 비관적 상태 분류. 429·5xx·그 밖 상태는 default(B7) AMBIGUOUS, 400/401/403/404/405/415/422 는 B5 REJECTED, 전송 오류는 B2(NotStarted→NOT_SENT) 아니면 AMBIGUOUS.",
        inputs=[("send", "SendState", "tracker", "—"), ("statusCode", "HTTP", "statusOf", "0 이면 B6"), ("err", "전송 오류", "—", "—")],
        calls="isDefinitiveRejection 만.",
        mutations="없음.",
        safety="편집 계획 없음."),

    "classifyStatus": dict(
        role="비 2xx status 를 official 오류로. B5(56) `code >= 500` 은 **sentinel `ErrServer` 그 자체**(57)를 돌려준다 — a165 (b) 의 「5xx 를 실제로 받음」은 이 동일성(`err == official.ErrServer`)으로 판정한다(1.3 A2).",
        inputs=[("code", "HTTP status", "doRequest", "—"), ("body", "응답 본문", "doRequest", "401/403 은 본문 대신 코드만 실음")],
        calls="`reIPWord.Match`, `fmt.Errorf`(401/403), `&APIError`(default).",
        mutations="없음.",
        safety="편집 계획 없음 — 동일성 반환을 시험으로 고정(orders_raw_test.go:201 이 이미 `err == ErrServer` 를 단언)."),
    "Client.doRequest": dict(
        role="요청 1회. B1(194) 전송 오류 → status 0 + `ErrTransport` 감쌈(196). B2(201) 본문 읽기 실패 → status 를 함께 반환하지만 `send` 가 버린다(1.3 A2) — 그래서 a165 는 「응답 첫 바이트를 받았는가」를 httptrace `GotFirstResponseByte` 로 따로 잰다.",
        inputs=[("req", "요청", "send", "—")],
        calls="`c.hc.Do`, `io.ReadAll`, `observeAttempt`, `c.rates.record`.",
        mutations="rate budget 기록.",
        safety="편집 계획 없음."),
    "unwrapAndDecode": dict(
        role="2xx 본문의 envelope 해석. 실패는 B2(218)·B3(221)·B4(224) 모두 **감싼** `ErrServer`(219·222·225) — 2xx 로 접수된 응답이다. a165 (b) 에서 제외(동일성 판정이 거름).",
        inputs=[("body", "2xx 본문", "send", "—"), ("out", "nil 가능", "호출자", "B1 nil 이면 버림")],
        calls="`json.Unmarshal`, `fmt.Errorf`.",
        mutations="out 채움.",
        safety="편집 계획 없음."),
    "tokenManager.refresh": dict(
        role="401 뒤 토큰. B1(119) 캐시 토큰이 유효하고 거절된 것과 다르면 채택(adopted=true, 121), 아니면 교환(125, adopted=false). `Client.send` B8 이 adopted=false 면 break — 401 재전송 횟수의 근거.",
        inputs=[("refused", "거절된 토큰", "send", "—")],
        calls="`m.loadCache`, `isStillValid`, `m.exchange`.",
        mutations="토큰 캐시.",
        safety="편집 계획 없음 — D9 시험의 근거."),
    "refusalCodes": dict(
        role="오류 본문의 최상위 `code` 와 `error.code` 를 읽는다. 취소 409 `already-processing` 판정은 이 한 곳을 쓴다(1.3 A9) — 두 자리 모순이면 재시도·확정 분류 없음.",
        inputs=[("body", "APIError.Body", "classifyStatus default", "JSON 아님이면 ok=false(B1)")],
        calls="`json.Unmarshal`, `readCode`, `isJSONNull`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "parseOrderFacts": dict(
        role="브로커 주문 JSON → 매처 입력. quantity 결손은 `parseOptionalDecimal` 이 0 으로(B3), 그러면 `matcher.matches` B3 가 불일치로 판정 — a165 조회는 0 을 「모름 = 일치」로 다룬다(1.3 A8).",
        inputs=[("raw", "주문 JSON", "OrdersPageRaw", "B1·B2 오류")],
        calls="`json.Unmarshal`, `parseOptionalDecimal`, `time.Parse`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "newMatcher": dict(
        role="intent·attempt 로 매처. B3(673) 디스패치 시각 ±SubmitWindow(기본 5분) 창. a165 조회는 첫 전송 시각을 stamp 로 준다.",
        inputs=[("intent", "journal.Intent", "Prepare", "수량 파싱 실패 B1"), ("rec", "AttemptRecord", "—", "시각 없으면 창 미사용")],
        calls="`strconv.ParseFloat`, `parseOptionalDecimal`, `time.Parse`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "matcher.matches": dict(
        role="지문 비교. 모르는 값은 일치 쪽(B1·B2·B4·B5 의 빈 값 통과) — 단 수량(B3)은 결손 0 이 불일치가 된다.",
        inputs=[("o", "orderFacts", "parseOrderFacts", "—")],
        calls="`qtyClose`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "Gateway.roundTripFor": dict(
        role="ACK 뒤 정착 전 확인 훅. B1(72) place 가 아니거나 `g.orders` 가 nil 이면 nil. a165 A1 수리는 이 훅을 감싸 재전송 ACK 에 목록 대조를 더한다.",
        inputs=[("plan", "mutationPlan", "submit", "—")],
        calls="`g.confirmCreatedOrder`.",
        mutations="없음.",
        safety="편집 계획: 감싸는 쪽(submit)에서 — 이 함수는 그대로."),
    "Attempt.DispatchVerified": dict(
        role="dispatch 1회 → 정착. ACK 이면 B11 MarkAcked(181) 뒤 B12·B13 verify(184–185) 실패 시 MarkInDoubt(191) — A1 의 목록 대조가 들어갈 자리. 클래스 REJECTED 는 B6 FAILED_CONFIRMED.",
        inputs=[("dispatch", "DispatchFunc", "submit", "nil 이면 B1"), ("verify", "ExistenceCheck", "roundTripFor", "nil 이면 생략")],
        calls="`MarkDispatchStarted`, `Settle`, `MarkAcked`, `MarkInDoubt`.",
        mutations="attempt 상태 전이(journal).",
        safety="편집 계획 없음."),
    "Service.Cancel": dict(
        role="취소. B3(218) `GetOrderAvailableActions` GET 사전 확인 뒤 POST. 재전송마다 GET 이 한 번 더 나간다(rate 예산).",
        inputs=[("intent", "CancelIntent", "Gateway.Cancel", "guard 실패 B1")],
        calls="`s.guard`, `GetOrderAvailableActions`, `CancelPendingOrder`, `withLineage`.",
        mutations="lineage 기록(성공 시).",
        safety="편집 계획 없음."),
    "Saga.settleStep": dict(
        role="CLI flatten 의 취소 한 단계. B9(395) 분류가 InDoubt 면 B10(400) `resolveStep` 을 **동기로** 돈다(해소 최소 45s) — 409 를 확정 미수락으로 바꾸면 이 동기 해소를 타지 않는다.",
        inputs=[("step", "FlattenStep", "CancelAll", "—")],
        calls="`decisionFor`, `Gateway.Cancel`, `classifyCancel`, `resolveStep`, `UpdateFlattenStep`.",
        mutations="flatten step 상태.",
        safety="편집 계획 없음 — 배선(cmd/tossctl/flatten.go)만."),
    "classifyCancel": dict(
        role="게이트웨이 결과 → flatten step. B3(458) FAILED_CONFIRMED → FlattenStepFailed, B5(464) IN_DOUBT → InDoubt.",
        inputs=[("out", "Outcome", "Gateway.Cancel", "—")],
        calls="없음(`err.Error`).",
        mutations="없음.",
        safety="편집 계획 없음."),
    "ExitObserver.clearTheSymbol": dict(
        role="엔진 청산 전 미체결 정리. B11(1697) 취소가 오류이거나 CONFIRMED 가 아니면 `cleared=false` 로 다음 주문으로 continue — 같은 주기 안 대기 없음. B3(1639) 미정착 mutation 이 있으면 그 종목은 이번 주기 치우지 않음. 409 를 FAILED_CONFIRMED 로 분류하면 그 attempt 는 미정착이 아니므로 다음 주기에 다시 취소된다(사용자 결정 1 D).",
        inputs=[("m", "managed", "ObserveOnce", "—"), ("withPending", "bool", "호출자", "—")],
        calls="`LiveOrdersForSymbol`, `UnsettledOnSymbol`, `ConfirmedCancelOf`, `IssueReduction`, `Submit.Cancel`.",
        mutations="발의 해제(release).",
        safety="편집 계획 없음."),

    "New": dict(
        role="게이트웨이 생성자. a165 는 `Options.CancelRefusal` 을 `g.cancelRefusal` 로 옮기는 대입 한 줄만 더한다(분기 추가 없음).",
        inputs=[("opts", "Options", "배선", "B2–B4 필수 필드 없으면 오류")],
        calls="`strings.TrimSpace`, `clock.System`.",
        mutations="Gateway 구성.",
        safety="High-risk 경로의 구성만 — 대입 1줄."),
    "buildGateway": dict(
        role="엔진 게이트웨이 배선(execgw.New 호출 311). a165 는 `CancelRefusal: execgw.EngineCancelRefusal()` 한 필드를 더한다 — 취소 409 확정 미수락, 같은 주기 재시도 0.",
        inputs=[("in", "gatewayInputs", "엔진 기동", "B1–B6 오류 반환")],
        calls="`execgw.New` 등.",
        mutations="엔진 배선.",
        safety="필드 1개 추가. 되돌리기 = 그 줄 삭제."),
    "buildFlattenWiring": dict(
        role="CLI `tossctl flatten` 배선(execgw.New 233, `flatten.Saga` 247 — 생산 유일한 Saga 생성 자리). a165 는 `CancelRefusal: execgw.FlattenCancelRefusal()` 를 더한다 — 취소 409 재시도(누적 5s·2회), 소진 시 오늘 경로.",
        inputs=[("root, opts", "CLI 옵션", "cobra", "B1·B2·B5·B7·B8 오류 반환")],
        calls="`execgw.New`, `journal.Open` 등.",
        mutations="CLI 배선.",
        safety="필드 1개 추가."),
    "Saga.resolveStep": dict(
        role="flatten 의 IN_DOUBT 취소를 **동기로** 해소. B3 switch: CONFIRMED(434) / FAILED_CONFIRMED(437) / 그 밖(443). a165 R4 가 소진되면 결과가 오늘처럼 IN_DOUBT 여서 이 함수로 온다(1.3 재리뷰 F1 수리).",
        inputs=[("step", "FlattenStep", "settleStep", "B1 Resolver 없음")],
        calls="`s.Resolver.Resolve`.",
        mutations="없음(값 반환).",
        safety="편집 계획 없음."),
    "Saga.planLiquidation": dict(
        role="청산 매도 계획. B5(307) `UnsettledCancelSymbols` 가 준 held 종목은 계획에서 뺀다 — 취소가 정착 안 된 종목의 oversell 방지 가드. FlattenStepFailed 는 held 가 아니라서 매도가 계획된다(1.3 재리뷰 F1 의 사슬).",
        inputs=[("snapshot", "reconcile.Snapshot", "Saga.Run", "B1 오류")],
        calls="`UnsettledCancelSymbols`, `priceAndSize`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "Journal.UnsettledCancelSymbols": dict(
        role="flatten 취소 step 중 Pending·InDoubt·Held 인 종목만 센다(쿼리 조건). Failed·Done 은 정착으로 본다.",
        inputs=[("sagaID", "string", "planLiquidation", "B1 오류")],
        calls="`db.QueryContext`, `rows.Scan`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "classifyRefusalCode": dict(
        role="확정 거절 code 목록(`opposite-pending-order-exists`)만 확정으로. `already-processing` 은 목록에 없어 B6(74) 판정 없음 → 종전 경로(409 AMBIGUOUS). B3(63) 두 자리 모순이면 모호 강제. 주문 종류를 가리지 않으므로 a165 는 409 AP 를 이 목록에 **넣지 않고** 취소에만 별도 판정을 둔다.",
        inputs=[("err", "official 오류", "classifyMutation", "APIError 아니면 B1")],
        calls="`refusalCodes`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "Client.PlaceOrder": dict(
        role="POST /api/v1/orders. `buildOrderCreate`(B1) → `acctHeaders`(B2) → `postWithHeaders`(B3, 401 재전송은 그 안의 send) → echo 검사(B4). R5 시험의 대상.",
        inputs=[("intent", "PlaceIntent", "officialBroker", "B1·B2 오류")],
        calls="`buildOrderCreate`, `c.acctHeaders`, `c.postWithHeaders`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "Client.CancelOrder": dict(
        role="POST /api/v1/orders/{id}/cancel. 409 는 `postWithHeaders`→`send`→`classifyStatus` default 의 `*APIError{409, Body}` 로 올라온다(B2).",
        inputs=[("orderID", "string", "officialBroker", "B1 계좌 헤더 오류")],
        calls="`c.acctHeaders`, `c.postWithHeaders`.",
        mutations="없음.",
        safety="편집 계획 없음."),
    "Client.postWithHeaders": dict(
        role="본문을 `json.Marshal` 로 **한 번**(B1, 431) 인코딩하고 makeReq 가 그 바이트로 reader 를 만든다 — `send` 의 401 재전송 본문이 바이트 동일한 근거(R5).",
        inputs=[("body", "any", "PlaceOrder 등", "B1 마샬 오류")],
        calls="`json.Marshal`, `http.NewRequestWithContext`, `c.send`.",
        mutations="없음.",
        safety="편집 계획 없음."),
}


def load_cov(paths):
    blocks = []
    for p in paths:
        for line in Path(p).read_text().splitlines()[1:]:
            loc, _, count = line.rpartition(" ")
            loc, _, _n = loc.rpartition(" ")
            fname, _, span = loc.rpartition(":")
            s, e = span.split(",")
            sl, sc = map(int, s.split("."))
            blocks.append((fname, sl, sc, int(count)))
    return blocks


def body_count(blocks, relfile, line, col):
    if not any(f.endswith("/" + relfile) for (f, _sl, _sc, _c) in blocks):
        return "nofile"
    hits = [c for (f, sl, sc, c) in blocks if f.endswith("/" + relfile) and sl == line and sc > col]
    if not hits:
        return None
    return max(hits)


def esc(s):
    return s.replace("|", "\\|").replace("`", "'")


def main():
    change = Path(sys.argv[1])
    blocks = load_cov(sys.argv[2:])
    for ast_path in sorted((change / "analysis" / "function-logic").glob("*/ast.json")):
        a = json.loads(ast_path.read_text())
        q = f"{a['receiver']}.{a['function']}" if a.get("receiver") else a["function"]
        n = NOTES[q]
        rel = a["file"]
        src = Path(rel).read_text().splitlines()
        start, end = a["start"]["line"], a["end"]["line"]
        branches = a.get("branches") or []
        returns = a.get("returns") or []
        calls = a.get("calls") or []
        rows, trows = [], []
        for b in branches:
            ln, col = b["at"]["line"], b["at"]["column"]
            cond = esc(src[ln - 1].strip())
            cnt = body_count(blocks, rel, ln, col)
            ran = ("미측정 — 이 패키지 coverprofile 없음" if cnt == "nofile" else
                   "측정 불가(본문 블록 없음)" if cnt is None else
                   (f"실행됨 — 패키지 스위트 count={cnt}" if cnt > 0 else "미실행 — 패키지 스위트 count=0"))
            rows.append(f"| {b['id']} | `{b['kind']} at {ln}:{col}` | {cond} | 아래 「State mutations」 | 소스 줄 {ln} 이하 | {ran} |")
            trows.append(f"| {b['id']} | `{b['kind']} at {ln}:{col}` | {cond} | {ran} | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |")
        if not branches:
            rows.append(f"| B1 | `branchless happy path at {start}:1` | 분기 없음 — 단일 경로 | 아래 | 반환 {', '.join(str(r['at']['line']) for r in returns)} | 패키지 스위트 |")
            trows.append(f"| B1 | `branchless happy path at {start}:1` | 분기 없음 — 단일 경로 | 패키지 스위트 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |")
        names = sorted({c.get("text", "") for c in calls if c.get("text")})
        ret = ", ".join(f"{r['at']['line']}:{r['at']['column']}" for r in returns)
        inputs = "\n".join(f"| {esc(i[0])} | {esc(i[1])} | {esc(i[2])} | {esc(i[3])} |" for i in n["inputs"])
        head = (f"- Source: `{rel}` ({start}-{end})\n- Qualified function: `{q}`\n"
                f"- Revision: `current` (a165 base `34e12967e`, `source_sha256` {a['source_sha256'][:16]}…)\n"
                f"- AST evidence: `ast.json` — AST branches {len(branches)}, 반환 {len(returns)}, 호출 {len(calls)}\n"
                f"- Risk scan: `risk-pattern-report.md`\n"
                f"- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official·trading·flatten 패키지 전체, journal `-run 'Classify|SendTracker|Dispatch'`, app/engine `-run 'Clear|Cancel'` — journal·engine 은 부분 실행이라 count=0 이 「미실행」을 뜻하지 않을 수 있음, cmd/tossctl 은 미측정) — "
                f"분기 줄에서 시작하는 본문 블록의 실행 수. 시험 이름별 귀속은 GREEN 재추출에서 함.\n")
        flm = f"""# Function Logic Map: `{q}`

{head}
## 역할 (a165)

{n['role']}

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
{inputs}

## Branches and early returns

AST 열거 순서 그대로. Condition 칸은 AST 좌표 줄의 원문.

Exact AST return positions: {ret}

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
{chr(10).join(rows)}

## Calls and live bindings

{n['calls']}

AST 호출 이름 전체: {', '.join('`'+x+'`' for x in names)}.

## State mutations and fallbacks

- {n['mutations']}

## Safety conclusion

- {n['safety']}
"""
        btm = f"""# Branch Test Map: `{q}`

- Source: `{rel}` ({start}-{end})
- AST branches {len(branches)} — 실행 칸은 편집 전 패키지 스위트 coverprofile 측정(시험 이름 귀속은 GREEN 재추출 때)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
{chr(10).join(trows)}
"""
        (ast_path.parent / "function-logic-map.md").write_text(flm)
        (ast_path.parent / "branch-test-map.md").write_text(btm)
        print("wrote", ast_path.parent.name)


if __name__ == "__main__":
    main()
