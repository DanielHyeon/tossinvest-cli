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
             "a165 는 이 함수의 dispatch 클로저 안에 재전송 루프를 넣는다(design D1).",
        inputs=[("plan mutationPlan", "place/cancel/amend 계획", "Gateway.place·Cancel·Amend", "prepareRequest 오류면 기록 없이 반환"),
                ("ref GuardianDecision", "결정 id", "journal", "읽기 실패는 refusePrepared 로 기록"),
                ("plan.strategy", "nil 또는 paired KR/US first-leg 능력", "PlaceClaimedStrategy", "nil 이면 일반 경로")],
        calls="`plan.call` 은 클로저 `call` 안의 한 자리(708)에서만 불리고, `call` 은 B44(`withStrategyEntryGateAuthority` 의 fn, 712) "
              "또는 B45(717) 중 정확히 한 갈래에서 한 번 불린다. 함수 전체에 for/range/select/go 가 0 이다 — 즉 지금은 "
              "attempt 당 브로커 호출이 정확히 1회다. 결과는 725 의 `classifyMutation(tracker.State(), result, callErr)` 한 자리에서 분류된다.",
        mutations="journal Prepare(RECORDED) · DispatchVerified/DispatchStrategyVerified(DISPATCH_STARTED → 정착) · 심볼 claim/release. "
                  "재시도가 생기면 같은 attempt·같은 nonce·같은 clientOrderId 로 브로커 호출이 늘 뿐 journal 행은 늘지 않는다(design D1).",
        safety="편집 경계는 dispatch 클로저(636–726) 안. 클로저 밖 거절 순서(B6–B28)와 정착(B49–B57)은 바꾸지 않는다. High-risk: yes."),
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
            ran = ("측정 불가(본문 블록 없음)" if cnt is None else
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
                f"- 실행 칸 측정: 편집 전 기준선 `go test -count=1 -coverprofile` (execgw·official 패키지 전체, journal 은 `-run 'Classify|SendTracker|Dispatch'`) — "
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
