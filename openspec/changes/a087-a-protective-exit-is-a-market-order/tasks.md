# a087 tasks

> **High-risk.** 손절 주문의 **유형**을 바꾼다. 구현 착수 전 Pre-Edit 선언과
> proposal-freeze 재리뷰(적대적 Eng 관점 포함)가 필요하다.
> 초안(호가 그리드)의 리뷰 결과는 `review.md`에 보존한다 — 이 tasks는 그 리뷰가
> 도달한 결론을 실행하는 것이지 우회하는 것이 아니다.
> **2026-10-10 개정**: 2차 proposal-freeze 재리뷰 13건 처분(0.6) — §1 에 셋째 관문(1.5)·종단 시험(1.6), §3 에 3.6~3.8,
> §4 에 4.5·4.6, §5 에 실행 수단과 결과 → 처분 표.
> **2026-10-10 3차 개정**: 3차 재리뷰 처분(0.8) + 사용자 결정 2차 — 1.5.x 를 **엔진 인스턴스 한정 개방**으로 다시 썼고
> (1.5.5 = upstream 동일성 고정 시험), §5 「수용」 기준에 주문 상태 확인·누락 결과 3행·실행 전제를 더했다.
> **2026-10-28 7판**: 사용자 결정 Q1~Q4 — US 정규장 LIMIT(2.6.3), 구간 입력 = 브로커 달력 + 고정 시각(2.6.1·2.6.4), §5.2 not-applicable(5.2·5.3·2.6.6 닫음),
> 착수 조건 ④ 「US 미개방」 확인, 0.10 리뷰 PASS.
> **2026-10-28 0.10 독립 리뷰(GPT-5.6 Terra) FAIL → 보완**: P0-1 엔진 전용 배선(1.5.x — 옵션은 `NewContext` 에서만) · P1-2 KR 정수 수량(1.1a·2.6.x) 닫음.
> P1-1(KRX 단독 판정 입력 없음)은 **사람 결정 대기**(`issues.md` D-1) — 결정 전 §1~§3 착수 금지.
> **2026-10-28 8판 (사람 결정 「3은 3」 — `issues.md` D-1 (다))**: 보호 매도 = **그날 하한가 지정가**(design D11). §1 은 「관문 무변경 고정」으로, §2 는
> 하한가 판 · 갱신기 · 가격 결정 · 거부 되먹임으로, §3 의 3.1~3.4 는 가격 출처 기록으로 다시 썼다. 7판 §1·§2·3.1~3.4 원문은 각 절 끝 「7판 원문 (8판에서 대체됨)」에
> 체크 상자 없이 남겼다. 0.10 은 **8판 개정 — 재리뷰 대기**, 재리뷰 전 §1~§3 착수 금지.
> **2026-10-29 9판 (Manager 결정 — 8판 독립 재리뷰 FAIL 처분)**: design D12. tasks 1.0(a165 재대조 = 구현 착수 조건) · 2.4(토큰 갱신기) · 2.5(비동기 되먹임) · 2.8(엔진 전용 읽기) ·
> 2.9(재시작 재구성) · 3.1(골든) · 5.4(착지 전 필수 gate). 0.10 = **9판 — 재리뷰 대기**.
> **2026-10-29 10판 (Manager 결정 — 9판 재리뷰 FAIL 처분, KISS)**: design D13. 하한가는 그날 그 종목의 첫 보호 매도 시도에만(2.5 · 2.9), 429 정지 입력 = `doRequest` 의 HTTP 429 시각
> 하나(2.4 · 2.10), release 실패 종단 시험(2.5). 0.10 = **10판 — 재리뷰 대기**.
> **2026-10-29 11판 (Manager 결정 — 10판 재리뷰 FAIL 처분)**: design D14. 재구성 action = `isProtective` 와 같은 Go 상수(2.9 구조 · 행동 시험), 시작 정숙 5분 · 401 시 wire 3(2.4),
> 예산 기준 계좌 공유 3 req/s. 0.10 = **11판 — 재리뷰 대기**.
> **2026-10-29 12판 (Manager 결정 — 11판 재리뷰 FAIL 처분)**: design D15. 하한가 GET 은 a166 계좌 공유 admission 이 주입될 때만(2.4 — 엔진 배선은 그 전까지 nil, 생산 변화 0),
> 2.10 not-applicable(429 기록은 a166), 구조 시험은 평가된 `Action` 값 집합(2.9). 0.10 = **12판 — 재리뷰 대기**.

## 0. 게이트 선행

- [x] 0.1 `capture_change_base.py --change a087-a-protective-exit-is-a-market-order`로 base commit 재고정 (디렉터리명이 바뀌었다) — `5491451b`(ec29dc72 → 102d4e99, 2026-09-30, 승인 기록 review.md)
- [x] 0.2 `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive` — 2026-10-10 rc=0
      ("is valid", 문서 수정 없음)
- [x] 0.3 **proposal-freeze 재리뷰** 실행 후 `review.md`에 2차 절 추가 (적대적 Eng 필수)
      — 2026-10-10 실행(Opus 독립 리뷰, `[subagent-only]`): **판정 = 수정 필요, freeze 불가.**
      P0 1·P1 4·P2 5·P3 3. P0-1(`trading/service.go` 셋째 관문 — tasks 대로면 보호 청산 100%
      로컬 거부)·P1-1(5.1 이 출하 도구로 실행 불가) 은 Manager 가 HEAD 실물 대조로 확인.
      P0-1·P1-1~4 를 문서로 닫고 Eng 재리뷰 재실행 전 §1~§3 착수 금지. P0-1 처분
      (internal/trading 수정 = 사람 CLI upstream 동작 동반 변경)은 **사람 결정**.
      → 이후 정정: 사용자 결정 2차(2026-10-10)로 ③ 개방은 **엔진 인스턴스 한정**, 사람 CLI 동반 변경은 철회(1.5).
- [x] 0.4 `make sdd-sync` 후 `sellIntent`·`checkOrderShape`·`isProtective`·`buildOrderCreate`의
      definition/callers/impact 확인 — 2026-10-10 HEAD `6e844e11`, codegraph 1.6.0. `make sdd-sync` rc=2(2회):
      CodeGraph sync 성공·fingerprint 기록, CodeGraphContext `update` 300초 타임아웃(advisory 미갱신).
      산출물 `analysis/gate-0.4-codegraph.md`
- [x] 0.5 **엔진 세션 게이트 유무 확인 (design D7 — 재리뷰 P1-4)** — 청산 판정 상류(`ObserveOnce`·`judge`·`record`)를
      CodeGraph 로 따라가 세션·시각 입력이 있는지 열거하고 `analysis/` 에 남긴다. 결과와 §5.2 를 입력으로 D7 교집합의
      처분을 사람이 정한다. **§1~§3 착수 조건**. read-only 작업이고 그 답이 spec Requirement 1 의 SHALL NOT 을 무효화할 수
      있으므로 **0.9 freeze 판정 전에 수행하는 것을 권장**한다(3차 재리뷰 A-P3-4 — 차단 아님, Manager 판단)
      — 2026-10-20 HEAD `00fed684`, logic-map AST 12개 + codegraph 1.6.0: **세션 입력 0 · 시각 입력 40(전부 경과 시간)** →
      엔진에 세션 게이트 없음(D7 「교집합 공집합」 분기 불성립). `analysis/gate-0.5-session-inputs.md` · `analysis/session-gate-0.5/`.
      **D7 처분은 미정 — §5.2 결과와 함께 사람 결정 대기**
      → **D7 처분 결정(사용자 2026-10-22)**: 시장가를 받지 않는 구간은 오늘처럼 지정가. spec Requirement 1 개정, design D7 「처분」, 구현 task 2.6
- [x] 0.6 **2차 재리뷰 문서 수리** — P0-1·P1-1~4·P2-1~5·P3-1~3 처분(2026-10-10, 처분 대조표 `review.md` 끝).
      사용자 결정 2건(2026-10-10) 반영. production 편집 0
- [x] 0.7 **Eng 재리뷰 1회** (재리뷰 「freeze 재개 조건」) — 0.6 개정본 대상. 2026-10-10 실행(3차, `e93cd6a4`):
      **판정 = freeze 재개 불가**(문서 P1 2·P2 5·P3 4 / 도구 P1 1·P2 1·P3 6). 사용자 결정 2차(`77a8e119`) — I-R1 = 엔진 인스턴스 한정
- [x] 0.8 **3차 재리뷰 수리** — 3차 발견 전수(P1 3·P2 6·P3 10) 처분(처분 대조표 3차분 `review.md` 끝). 도구 변이 하네스는
      사본 실행으로 수리(`45167aa7`, 변이 13/13 CAUGHT). production 편집 0
- [x] 0.9 **Manager 대조 + ③ 인스턴스 옵션 설계 문단 독립 확인 1회** (3차 「freeze 재개 조건」 4 — 전면 Eng 재리뷰 불요).
      통과 전 §1~§3 착수 금지
      — 2026-10-10 Manager 종결: ① 대조 — 3차 발견 19건 처분을 대조표(3차분, review.md:577)와
      실물로 교차(「9분」·「a088」 잔존분은 전부 정정 기록 문맥 = 올바름; §5 수용 조건 200+상태,
      1.5.5 upstream 동일성 시험, mutants.sh 사본 실행 git worktree 확인), 배터리 재현(vet rc=0 ·
      gofmt 빈 출력 · `-race` PASS). ② 독립 확인 — D3 ③ 하중 주장 둘을 Manager 가 이 세션에서
      직접 검증: `engine.go:401` 자체 Service 인스턴스 + 빌더 관례(`WithConditional`) 실물,
      `placeIntentSupported` 비시험 호출자 = `PreviewPlace`·`Place` 둘(0.4 CodeGraph 증거).
      ③ 보류 판단 둘 처분 — I-R3: D6 은 「미채택·원안 평가 미수행」 유지(평가 없는 문언 기각은
      부정확; 실측 422 거부 시 재검토가 §5 표에 이미 있음). I-R8: 0.5 는 §1~§3 착수 전 조건
      유지(freeze 는 문서 상태 판정, §1~§3 은 어차피 §5 실측 대기).
      → **proposal-freeze 재개 선언.** 남은 §1~§3 착수 조건 = §5 실측(결과→처분 표) + 0.5.
- [ ] 0.10 **requirement 변경 리뷰 (2026-10-22 D7 처분) — 12판 — 재리뷰 대기 (2026-10-29)** — spec Requirement 1 에 「시장가를 받지 않는 구간은 오늘처럼 지정가」 예외를 넣은
      것은 Requirement 수준 수정이다(WORKFLOW 「리뷰 게이트」). 수정분(spec Requirement 1 · design D7 「처분」 · tasks 2.4~2.6)에 적대적 Eng
      관점 포함 리뷰 1회를 받아 `review.md` 에 남긴다. Manager 확인 항목 하나를 함께 판정한다 — **접수 미측정 구간과 구간 판정 불가를 LIMIT
      으로 두는 것**은 사용자 문언(「시장가가 안 되는 시간대」)이 아니라 design D7 의 불변식 4 유도다. **§1~§3 착수 조건**
      — 2026-10-28: 사용자 결정 Q1~Q4(「권장대로」) 반영 뒤 리뷰(`review.md` 「0.10 requirement 변경 리뷰 — D7 7판」). **판정 PASS** — P0 0 · P1 2(합집합 달력이
      KRX 조기 폐장을 숨김 · 즉시성 제약 부재 — 둘 다 이 개정에서 닫음) · P2 4(닫음 1, **Manager 처분 대기 3**: P2-2 「고정 시각」 해석 · P2-3 gate ⑤ 창 크기 ·
      P2-4 종가단일가 LIMIT) · P3 6. Q1(위 유도) **수락**. 근거 AST `analysis/evidence/`(HEAD `0c9bfd272`, 20개, stale 0). 작성자와 같은 세션의 리뷰라 독립 리뷰 아님 —
      착수 전 Manager 대조 1회
      — **2026-10-28 독립 리뷰(GPT-5.6 Terra, `review.md` 「0.10 독립 리뷰」) FAIL: P0 1 · P1 2 → 보완 완료, 사람 결정 대기.** P0-1(`NewOrderPath` 공용 빌더 —
      비시험 호출 자리 2: `engine.go:438` · `flatten.go:200`)과 P1-2(KR 정수 수량)는 문서로 닫았다(design D3 「0.10 독립 리뷰 반영」, `review.md` 처분 절).
      P1-1 — KRX 단독 개장 시각 · 시장 전체 정지 입력이 openapi 에 없고 종목별 정지 · VI 는 GET 0 아래 「지금」을 못 말한다 → KR 도 LIMIT 이면 Phase 2 의 실제 변경 0 —
      은 **사람 결정 대기**(`issues.md` D-1, 선택지 (가)(나)(다), 권장 (다)). 결정 뒤 그 판본으로 재리뷰(독립)를 받아야 이 줄을 닫는다
      — **2026-10-28 사람 결정 (다)(「3은 3」) → 8판 개정 — 재리뷰 대기.** design D11(주문 형태 · 하한가 출처 · 판 · 갱신기 · fail-safe · 되먹임 · 구간 위험 ·
      불변식 3/4/6 · 정수 수량 불요 · a165 재대조) · spec delta 3 Requirement 교체 · tasks §1/§2/§3 재편 · issues D-1 결정 기록 + D-2 사람 확인 4건. 근거 AST 36개
      (`analysis/evidence/`, HEAD `1f4145c77`, 새 11개, `enumerate.py` stale 0). P0-1(공용 빌더)·P1-1(KRX 단독 입력)·P1-2(정수 수량)는 유형이 LIMIT 이라 문제 자체가 소멸
      (design D11 (e)(f) · D7 머리말) — **독립 재리뷰가 그 소멸 주장과 D11 을 판정해야 이 줄을 닫는다.** 재리뷰 전 §1~§3 착수 금지
      — **2026-10-29 8판 독립 재리뷰(`review.md` 「8판 독립 재리뷰」, `5e691ac4e`) FAIL: P0 3 · P1 3 · P2 1 → 9판(Manager 결정).** design D12(D12.1 timestamp · raw decimal ·
      D12.2 예산 직렬화 가) · D12.3 비동기 되먹임 · 재시작 · D12.4 5.4 gate · D12.5 a165 · D12.6 골든 · D12.7 불변식 4 표) · spec Req 1 개정 · tasks 1.0 · 2.4 · 2.5 · 2.7 · 2.8 · 2.9 · 3.1 · 5.4 ·
      issues D-2 갱신. 근거 AST 42개(새 6, stale 0). **9판 — 재리뷰 대기.** 재리뷰 전 §1~§3 착수 금지
      — **2026-10-29 9판 재리뷰(`review.md` 「9판 재리뷰」, `525b4fb6c`) FAIL: P0 1 · P1 3 · P2 1 → 10판(Manager 결정, KISS).** design D13(D13.1 첫 시도 한정 · D13.2 `doRequest` 429 기록 ·
      D13.3 timestamp 5.4 보류 · D13.4 정수 정본 비교 · D13.5 불변식 4 표 + release 실패 칸) · spec Req 1 개정 · tasks 2.2 · 2.4 · 2.5 · 2.7 · 2.9 · 2.10 · issues D-2 ① ② · ⑤ 대체.
      근거 AST 43개(새 1, stale 0). **10판 — 재리뷰 대기.** 재리뷰 전 §1~§3 착수 금지
      — **2026-10-29 10판 재리뷰(`review.md` 「10판 재리뷰」, `4242ce3ed`) FAIL: P0 1 · P1 1 → 11판(Manager 결정).** design D14(D14.1 보호 action 상수 유도 · `BASELINE_BREACH_EXIT` 오기 정정 ·
      D14.2 401 시 토큰당 wire 3 · 엔진 시작 뒤 5분 GET 0 · D14.3 예산 기준 계좌 공유 3 req/s · D14.4 불변식 4 표) · tasks 2.4 · 2.7 · 2.9. **11판 — 재리뷰 대기.** 재리뷰 전 §1~§3 착수 금지
      — **2026-10-29 11판 재리뷰(`review.md` 「11판 재리뷰」, `bb0080ecd`) FAIL: P0 1 · P1 1 → 12판(Manager 결정).** design D15(D15.1 a166 admission 주입 시에만 하한가 GET — 그 전 nil · 생산 변화 0 ·
      `doRequest` 429 기록은 a166 으로 · D15.2 `Action` 값 집합 비교 · D15.3 불변식 4 표) · spec Req 1 · tasks 2.0 · 2.1 · 2.4 · 2.7 · 2.9 · 2.10(n/a) · issues D-2 ⑥. **12판 — 재리뷰 대기.** 재리뷰 전 §1~§3 착수 금지

## P1. Phase 1 — 가격 사다리 KR 하한가 단 (선행 · 실측 불요 · design D2a)

> **종결 (2026-09-30, Manager 판정 선택지 1): 불구현 — 전제 반증.** P1.2 ① 열거에서 대상 모집단이 생산에서 0
> (B1·B2 도달 불가, 엔진 커버리지 0)으로 측정됐다. 영수증: `issues.md` I-P1 · design D2a 「반증」 ·
> `analysis/function-logic/internal-app-engine--exitobserver.sellintent/` · `analysis/p1-population/`.
> Phase 1 의 대체 산출물은 P1.7 핀 시험 하나다. production 편집 0.

- [x] P1.1 **Pre-Edit 선언 + FLM** — FLM/BTM 작성(`b10cefaf`). Pre-Edit 선언: not-applicable — production 편집 없음
- [x] P1.2 **RED — 모집단 열거 포함** — ① 열거 완료: 모집단 0(`issues.md` I-P1, `b10cefaf`). ②~④ not-applicable —
      세울 단이 없음
- [x] P1.3 **GREEN** — not-applicable (D2a 불구현, `PriceLimits` 배선 없음 · 새 브로커 호출 0)
- [x] P1.4 원장 기록 — not-applicable (하한가 제출 경로 없음)
- [x] P1.5 변이 + 리뷰 — P1.7 핀 시험에 대해 수행: 무변이 대조군 GREEN 후 변이 4/4 CAUGHT(`7bd8f197` 메시지).
      Manager 독립 대조로 모집단 판정 확인(2026-09-30)
- [x] P1.6 D2b 제거 예약 — 제거할 단이 없음으로 정정(design D2b 취소선 주석)
- [x] P1.7 **평가기 관측가 가드 핀 시험** — `internal/exitpolicy/a087_observed_price_pin_test.go`(`7bd8f197`):
      빈·공백·"0"·파싱 불가 관측가 → `EvaluateRatchet`·`EvaluateLadder` 모두 "observed price" 거부. B1·B2 가 닫혀 있는
      등식을 고정. FLM: not-applicable — 시험 전용(비례 원칙)

> **아래 §1~§3(Phase 2)의 착수 조건 (2026-10-10 개정)**: ① §5 실측 기록과 그 결과가 아래 §5 표의 「§1~§3 착수 가능」
> 행일 것(§0.7 사람 승인), ② 0.5 확인 + D7 처분(**처분 결정됨 2026-10-22** — 남은 것은 0.10 리뷰), ③ 0.9 통과(0.7 은 재개 불가로 끝났다), ④ §5 표 아래 「US」 행 사람 확인.
> Phase 1 은 그 실측을 기다리지 않았다(종결).
>
> **상태 (2026-10-28, 사용자 결정 「권장대로」)**: ① **충족** — 5.1 수용(2026-10-19 M50, FILLED), 5.2 not-applicable(Q4). ② 0.5 완료 ·
> D7 처분 7판(Q1~Q3) · 0.10 리뷰 결과는 0.10 줄. ③ **충족**(0.9, 2026-10-10). ④ **확인 처리 — 「US 미개방」**(Q2: US 보호 청산은 실측 전까지
> LIMIT 이라 「US 를 미측정으로 착지」할 MARKET 칸이 없다).
>
> **8판 (2026-10-28)**: ① 5.1(MARKET 접수)은 8판의 전제가 아니다 — 8판은 MARKET 을 내지 않는다. 8판의 대응 실측은 **5.4(하한가 지정가 1주)** 이고,
> 착수 조건으로 둘지 착지 전 조건으로 둘지는 **사람 결정**(`issues.md` D-2 ②). ② 0.10 = 8판 재리뷰. ③ 그대로. ④ 그대로(US 는 오늘 LIMIT).

## 1. 실행 관문 — 8판: 무변경 고정 (Phase 2)

> **8판**: 보호 매도가 LIMIT 이므로 ②`checkOrderShape` · ③`placeIntentSupported` · `trading.Service` · `NewOrderPath` 는 **편집하지 않는다**(design D11 (e)).
> 남는 일은 「바뀌지 않았다」를 시험으로 고정하는 것뿐이다. FLM: not-applicable — 기존 함수 본문 편집 0(시험 전용, 비례 원칙).

- [ ] 1.0 **구현 착수 조건 — a165 재대조 (9판 — design D12.5)** — a165 archive · push 가 확인된 sha 에서 `clearTheSymbol` · `checkSymbolFree` · `cancel_refusal*.go` 와 a087 이
      인용 · 편집하는 `exitloop.go` 함수(`submit` · `Run` · `ObserveOnce` 인용 좌표)의 AST(`analysis/evidence/` 재생성, `enumerate.py` stale 0) · FLM · 공유 시험 더블을 다시 대조하고
      design D11 (g) · D12 좌표를 고친다. T2 를 고르면 `Gateway.submit` 도. **이 줄 전 §1~§3 착수 금지**
- [ ] 1.1 **관문 무변경 고정 시험** (7판 1.3 승계) — 엔진 인스턴스(`NewContext` 로 만든 경로)와 기본 인스턴스 각각에서 비분수 `sell+market` · `buy+market` 이 오늘처럼
      거부됨. 진입 시장가 거부가 살아 있음(riskcalc 규칙 — design D3 「왜 진입은 계속 막는가」)
- [ ] 1.2 **하한가 판 생성 자리 구조 시험** (7판 1.5.5 (c) 의 8판 판본 — 0.10 독립 리뷰 P0-1 규칙) — 판 생성 호출이 비시험 코드에서 `NewContext` 본문 하나뿐이고
      `NewOrderPath` · `buildFlattenWiring` · `app.go` · `mcp.go` 에는 0 임을 함수 단위 AST 로 센다. 변이: 「`NewOrderPath` 에서 판 생성」이 잡혀야 한다
- [ ] 1.6 **종단 회귀 시험** — 엔진 보호 청산 하한가 LIMIT 이 `sellIntent` · ② · ③ 을 **모두** 지나 전송 계층(시험 더블)에 닿고 wire body 의 `price` 가 판의 하한가다.
      창 밖 · 판 없음 행은 오늘 LIMIT 골든과 같은 wire body

### 7판 원문 (8판에서 대체됨 — 관문을 열지 않는다)


> **개정 (2026-10-10, 재리뷰 P0-1 · 사용자 결정)**: 3판의 §1 은 ②(`checkOrderShape`)만 다뤘다. ③(`placeIntentSupported`)을
> 열지 않으면 보호 청산이 100% 로컬 거부된다. 1.5 가 ③ 이다. ② 도 High-risk 기존 함수 내부 편집인데 3판에 Pre-Edit·FLM
> task 가 없었으므로 1.0·1.0.1 을 더한다(§2 와 같은 형식).

- 1.0 **Pre-Edit 선언** — `execgw.checkOrderShape` (WORKFLOW §Pre-Edit 형식)
- 1.0.1 **Function Logic Map** — `internal/execgw/failclosed.go` / `checkOrderShape`
      (`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`). HEAD 에서 생성.
      design D3 의 「새 분기 위치(통화·수량 검사 뒤)」·「정수 수량 검사 자리」를 이 열거로 확정
- 1.1 **RED** — `checkOrderShape` 표 테스트(②만 단독으로 — ③을 거치지 않는 직접 호출): `sell+market`(통과)·
      `buy+market`(거부)·`sell+market`에 가격 있음(거부)·`sell+market` + KR 비 KRW 통화(거부 유지)·`sell+market` + 수량 0(거부
      유지)·US fractional 기존 분기 무변화·`limit` 전 분기 무변화
- 1.1a **비분수 MARKET 정수 수량** (재리뷰 P2-2 → **0.10 독립 리뷰 P1-2 로 정해짐**: 「브로커 400 위임」 철회) — ② `sell+market`(비분수) 갈래에
      정수 backstop(execgw 정수 술어 하나 — `canonicalProtectionQuantity` 를 내보내 판정과 공유, design D3 「0.10 독립 리뷰 반영」). RED 행: KR `sell+market` 수량
      `1.5` · `1e-9` · `2^53+1` · `NaN` · `+Inf` → 거부(사유 문구는 「판정·관문 불일치」로 읽히게), `1` → 통과. B10 예외 안에 두고 조기 통과 금지 유지.
      **backstop 은 생산에서 도달 불가여야 한다** — 도달하면 `submit` B13(`:1582`) 해제 → 재발의 루프(2.6.3 종단 행이 잡는다)
- 1.2 **GREEN** — [`internal/execgw/failclosed.go`](../../../internal/execgw/failclosed.go) `checkOrderShape`
      의 `orderType != "limit"` 거부에 축소 시장가 예외 추가(조기 통과 반환 금지 — design D3)
- 1.3 진입 경로가 여전히 시장가를 거부함을 **별도 테스트로 고정**(②·③ 각각) — riskcalc의 규칙은
      살아 있고 이 change가 그것을 건드리지 않았다는 증거
- 1.4 `ReasonUnsupportedOrderType` 메시지 문구 갱신 (지금 "only limit orders (and US
      fractional market orders)"라고 단언한다)
- 1.5 **셋째 관문 — `internal/trading` `placeIntentSupported`, 엔진 인스턴스 한정** (사용자 결정 2차 2026-10-10:
      sell+market·가격 없음 허용은 엔진이 만드는 `trading.Service` 에만 생성자 옵션으로. CLI·ops·MCP 기본 경로는 upstream 바이트
      동일 — 1차 결정의 「사람 CLI 동작 변경」은 철회. design D3 ③)
  - 1.5.0 **Pre-Edit 선언** — `trading.placeIntentSupported`·`Service` 생성(옵션 추가)·`engine.go` **`NewContext`**(옵션 부착 — 0.10 독립 리뷰 P0-1:
        `NewOrderPath` 는 flatten-all 도 쓰는 공용 빌더라 편집하지 않는다)
        (WORKFLOW §Pre-Edit 형식). upstream 상속 테스트 영향 = **기본 동작 무변화**(술어 시그니처에 인자 하나가 더해지는 것뿐 —
        upstream 시험의 기대값은 그대로여야 한다). 회귀 방지 = 1.5.5. 불변식 3 = 엔진 꺼짐 ⇒ 엔진 인스턴스 없음 ⇒ 흔적 0
        (`issues.md` I-R1 종결 인용). 엔진 인스턴스를 쓰는 사람 조작 표면이 1.5.1 에서 발견되면 그 사실을 여기 적는다
  - 1.5.1 **Function Logic Map** — `internal/trading/service.go` / `placeIntentSupported` (+ `PreviewPlace`·`Place` 의 술어 호출 자리)
        (`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`). HEAD 에서 생성.
        risk-pattern-report 에 비시험 호출자 둘(`PreviewPlace`·`Place`)과 `trading.NewService` 생성 자리 셋 — 엔진
        (`engine.go` `OrderPath.Trading` → `execgw` 게이트웨이·엔진 `tradingService`), CLI 앱(`app.go` → `tossctl order place`·`tossctl ops`),
        MCP 서버(`mcp.go` → ops `place_order`, MCP 카탈로그 노출) — 를 CodeGraph 로 적는다. design D3 ③ 의 「통화 검사가 비지정가 거부
        뒤」(조기 통과 금지 근거)를 이 열거로 확정
  - 1.5.2 **RED** — `placeIntentSupported`(③만 단독으로 — ② 를 거치지 않는 직접 호출) 표, **옵션 켬 / 옵션 끔** 두 열:
        옵션 켬 — `sell+market`·가격 없음(지원)·`buy+market`(비지원)·`sell+market`·가격 있음(비지원)·**`sell+market` + KR 비 KRW
        통화(비지원)**·**US `sell+market` + 통화 KRW/USD 외(비지원)**·fractional·limit 무변화. 옵션 끔 — 비분수 market 전부 비지원
        (오늘과 같음). 엔진 인스턴스의 `PreviewPlace(sell+market)` `LiveReady` 가 `Place` 의 지원 판정과 일치. **조기 통과 금지**
        (design D3 ③ — P2-2 를 ③ 에 적용): 통화 행이 GREEN 이 조기 `return true` 로 구현되면 실패해야 한다. `Place` 가 `buy+market` 에
        `ErrPlaceUnsupported` 를 계속 반환함을 고정
  - 1.5.3 **GREEN** — 생성자 옵션(모양은 design D3 ③ 제안 — 기존 빌더 관례) 추가, 술어가 그 값을 인자로 받아 비분수 비지정가
        거부에 `sell+market`·가격 없음 예외를 둔다(조기 통과 반환 금지 — 뒤의 통화 검사 생존). 예외에 KR 정수 수량 backstop(0.10 P1-2 — ③ 은 execgw 를
        import 하지 않으므로 자기 검사, 1.5.2 에 행 추가). **`NewContext` B1(`:439`) 뒤에서 `path.Trading` 에만 켠다** — `NewOrderPath`(`:401`)는 바이트 그대로(0.10 P0-1)
  - 1.5.4 주석·문구 갱신 — `checkOrderShape` 의 「internal/trading is not ours to change — design D1」 주석을 「기본 동작은
        upstream 그대로, 엔진 인스턴스만 옵션」으로. `PreviewPlace` 경고 문구는 **기본 인스턴스에서 바이트 동일**; 엔진 인스턴스의
        문구 처리(그대로 / 한 줄 덧붙임)만 정한다. ops `place_order` Summary(「limit or US fractional market」)는 사실 그대로라 무변경
  - 1.5.5 **upstream 동일성 고정 시험** (3차 재리뷰 A-P1-1 — 사람·에이전트 경로 무변화) — 실주문 없음, 시험 더블:
        (a) 옵션 없는 기본 `trading.NewService` 가 비분수 market 매도를 `ErrPlaceUnsupported` 로 계속 거부하고 `PreviewPlace` 의
        `LiveReady=false`·경고 문구가 바이트 동일, (b) ops `place_order` 핸들러가 기본 인스턴스로 market 매도를 계속 거부(MCP 노출
        표면과 `tossctl ops` 가 같은 핸들러), (c) **구조 시험** — 옵션 호출 자리가 저장소 비시험 코드에서 **`NewContext` 본문 하나뿐**이고
        `NewOrderPath` · `buildFlattenWiring` · `app.go` · `mcp.go` 에는 0(함수 단위 AST 로 센다 — 0.10 독립 리뷰 P0-1). (c) 가 없으면 누가 CLI 생성에 옵션을 붙여도
        (a)(b) 는 시험이 만든 인스턴스를 재므로 초록이다. (d) **flatten 경로 행동 시험** — `engine.NewOrderPath`(임시 config dir · 시험 자격 증명, 네트워크 없음)가
        돌려준 `Trading` 과 `buildFlattenWiring` 이 게이트웨이에 넘기는 인스턴스의 `PreviewPlace(KR sell+market).LiveReady == false`. (e) `NewContext` 엔진의 같은 의도는
        `LiveReady == true`. **변이**: 「`NewOrderPath` 반환식에 옵션」·「`buildFlattenWiring` 에서 옵션 켜기」(flatten 경로에 옵션 — (c)(d) 가 잡음) ·
        「`NewContext` 에서 옵션 제거」((e)·1.6 이 잡음)
- 1.6 **종단 회귀 시험 (P0-1)** — 엔진 보호 청산 MARKET 이 ①`sellIntent`·②·③을 **모두** 지나 전송 계층(시험 더블)에
      닿는다. 이 시험이 없으면 P0-1 이 재발해도 단위 시험은 전부 초록이다

## 2. 보호 청산의 가격 — 8판: 그날 하한가 (High-risk 본체 · Phase 2)

> 설계 정본 design D11. 편집하는 기존 함수(FLM 대상)는 `ExitObserver.submit` · `exitOrderLineFor` · `exitOrderLine.args` · `NewContext` · `Context.ExitObserver` ·
> `engineRuntime` + **9판 `ExitObserver.Run`** + ~~10판 `official.Client.doRequest`~~(12판에서 뺌)(design D11 「착수 시 FLM 대상」 · D12). `sellIntent` · `noteHeldProposal` 은 본문 무변경. 새 함수(판 · 갱신기 · `exitLimitPrice`)는 새 파일 `internal/app/engine/exit_lower_limit.go`.

- [ ] 2.0 **Pre-Edit 선언** — 위 기존 함수 7(9판: `Run` 추가 · 10판의 `official.Client.doRequest` 는 12판에서 뺌 — D15.1) + 새 파일 셋(엔진 판 · 엔진 전용 읽기 · 원장 읽기)(WORKFLOW §Pre-Edit 형식). 불변식 3: 판은 `NewContext` 에서만 · 갱신기는 보조 실행자 · 엔진 꺼짐 ⇒ 흔적 0.
      불변식 4: 제출 경로 GET 0 · 결정 실패 = 오늘 LIMIT · 되먹임. 불변식 6: 손절 수준 · 수량 무변경, 가격만 체결 쪽
- [ ] 2.1 **Function Logic Map** — 위 기존 함수 7 의 번들(12판: `doRequest` 제외)(`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`)을 착수 HEAD 에서 만든다.
      `analysis/evidence/` 의 HEAD `1f4145c77` AST 는 0.10 증거이지 번들이 아니다. risk-pattern-report 에 `isProtective` 의 세 소비자(`record` · `submit` a091 floor ·
      a087 가격 결정)와 `exitOrderLineFor` 의 사다리 사본(design D11 (b) 「기록」)을 적는다
- [ ] 2.2 **RED — 가격 결정 표** (`exitLimitPrice` 단독, 시계 · 판 시험 더블, 벽시계 · 실제 브로커 금지): design D11 「8판 검증」 첫 항목 전 행 — 창 경계 09:00:00 · 15:29:59 → 하한가,
      08:59:59 · 15:30:00 · 주말 → 오늘 사다리 · 익절 → 관측가 · US → 관측가 · 판 nil / 그날 항목 없음 / 전날 항목 / 하한가 0 / 하한가 > 관측가 / 관측가 파싱 불가 / 그날 거부됨 /
      시장 문자열 판정 불가 → 오늘 사다리 · 관측가 빔 → 기준선 · 둘 다 빔 → `""`. **폴백 행의 `PlaceIntent` 는 편집 전 HEAD 의 `sellIntent` 출력 literal 골든과 같다**
      (골든은 이 RED 커밋에 먼저 커밋 — 바뀐 함수끼리 비교 금지, 0.10 독립 리뷰 P2-2 승계)
      **9판 추가 행**: 판 항목 「금지」 → 오늘 사다리 · 판 값이 정수 decimal 문자열 그대로 `sellIntent` 로 감(B3·B4 통과)
      **10판 추가 행(D13.4)**: 하한가 정본 `^[1-9][0-9]*$` · 정수 파싱, 관측가 decimal 값 비교 — `"50400"` vs 관측가 `"50400"`/`"50401"` → 하한가, `"50399"` → 관측가 · 하한가 `"050400"` · `"50400.0"` · `" 50400"` ·
      `"5.04e4"` · `""` → 관측가 · 관측가 파싱 불가 → 관측가 · 판 항목 「금지」 → 관측가
- [ ] 2.3 **GREEN — 판 · 가격 결정** — 새 파일에 불변 스냅숏 판(원자 포인터, 쓰기 잠금) · `exitLimitPrice` 총함수(error 없음 · 패닉 없음 · GET 0). `submit` 머리에서 한 번
      결정하고 그 값을 `sellIntent`(`observed` 자리)와 `exitOrderLineFor` 로 넘긴다. `NewContext` 반환 리터럴에 판 한 줄, `Context.ExitObserver` 가 덮음
- [ ] 2.4 **RED/GREEN — 갱신기 (9판 D12.2 가) 토큰 · 10판 D13.2 429 기록)** — 보조 실행자(선례 `Context.ManualBuyRefresher`)가 **관측 주기 끝 토큰**(`ExitObserver.Run` 이
      `ObserveOnce` 반환 뒤 · `Sleep` 앞에서 용량 1 채널에 막히지 않게 보냄 — `Run` 편집은 이 한 줄)을 받을 때만 GET 1건(시한 min(주기/2, 2초)). 창 안 · KR · 그날 항목 없는 종목만.
      **429 정지 = 입력 하나**: `official.Client` 의 「마지막 HTTP 429 시각」(헤더 무관, `doRequest` 가 기록 — 2.10)이 5분 안이면 GET 0. 입력 없으면 취소 대기. `engineRuntime` `Auxiliary` 등록.
      **시험**: 가짜 transport 로 관측 `/prices` 진행 중 하한가 요청 동시 수 = 0 · 한 주기 `MARKET_DATA` 요청 수 ≤ 오늘(판 없는 같은 시나리오) + 1 · **헤더 없는 429 → 재시도 성공(`cycle.Err == nil`) → 5분 GET 0** ·
      헤더 있는 429 · 갱신기 자신의 429 → 5분 GET 0 · 토큰 없으면 GET 0 · 결함이 `exit` 루프를 멈추지 않음
      **11판 (design D14.2)**: **엔진 시작 뒤 5분 GET 0**(시작 정숙 — 갱신기 `Run` 시작 시각 기준). 시험: 4분 59초 토큰 GET 0 · 5분 뒤 토큰 GET 1 · 헤더 없는 429 → 재시도 성공 → 재시작 → 5분 GET 0 ·
      401 → 401 → 200 에서 토큰 하나의 wire GET = 3(표준 `Client.send` 재전송 — 새 경로 없음) · 다음 토큰까지 추가 0. 수치 기준은 계좌 공유 3 req/s(D14.3)
      **12판 (design D15.1)**: 갱신기는 계좌 공유 예산 **admission 함수가 주입될 때만** 돈다 — nil 이면 GET 0. 엔진 배선은 a166 착지 전까지 **nil**(생산 동작 변화 0 · 활성화 조건).
      **시험 셋**: ① admission nil → 하한가 GET 0 · 판 비어 있음 · 창 안 보호 매도 = 오늘 LIMIT 골든 ② admission 거절 → 그 요청 GET 0 ③ 401 → 401 → 200 의 wire 요청 3건이 **각각** admission 을 탐
      (호출 계수 3 · 둘째 거절이면 셋째 wire 0 — 시험 더블 관문). 위 「429 정지 = 입력 하나(`doRequest` 기록)」는 12판에서 대체됨 — 429 뒤 정지는 a166 admission 의 거절로 표현
- [x] 2.10 ~~**RED/GREEN — `official.Client.doRequest` 의 HTTP 429 기록 (10판 — design D13.2)** — `c.hc.Do` 성공 뒤 `resp.StatusCode == 429` 이면 헤더와 무관하게 원자 값에 시각 저장, 읽기 접근자 하나
      (새 파일). 반환값 · 오류 · `rates.record` 무변경(기존 ratebudget · client 시험 green). **착수 전 겹침 대조**: a166(삽입 후보 `doRequest` · `send`) · a100(`client.dorequest` FLM 번들) ·
      a165(`client.send` 번들 — `doRequest` 호출자) 의 그 시점 상태를 1.0 과 같은 방식으로 대조. a166 이 먼저 계좌 공유 429 기록을 착지하면 이 task 는 그것을 읽는 쪽으로 줄인다~~
      — **not-applicable (12판 — design D15.1)**: `doRequest` 429 원자 기록은 a166 이 계좌 공유 예산과 함께 가져간다. a087 의 `doRequest` 편집 0 — a100 · a166 겹침 소멸
- [ ] 2.5 **RED/GREEN — 첫 시도 한정 (10판 — design D13.1)** — 하한가는 그날 그 종목의 첫 보호 매도 시도에만. 술어 하나 = 「같은 `account_ref` · `market` · `trading_day` · symbol 에
      보호 청산 매도 intent 존재」. `submit` 이 보호 청산 `Place` 를 부르면 결과와 무관하게 메모리 「금지」. **종단 시험**: 첫 시도 하한가 CONFIRMED → 둘째 시도 관측가 LIMIT ·
      직접 거부 B13 → 다음 주기 관측가 LIMIT · `IN_DOUBT → StateFailedConfirmed → noteHeldProposal B4 해제 → 다음 주기 관측가 LIMIT` · 첫 시도 창 밖(오늘 LIMIT) → 창 안 재발의 관측가 LIMIT.
      **release 실패 종단 시험(D13.5 #3)**: 직접 거부 B13 에서 `ReleaseUnacceptedExitProposal` 오류 → 발의 무장 유지 → 다음 주기 `noteHeldProposal` B4→B5 재해제 → 관측가 LIMIT ·
      IN_DOUBT → 비수용 확정 해제(B4)에서 release 오류 → 같은 복구 → 관측가 LIMIT(두 경우 모두 하한가 재제출 0). `noteHeldProposal` · `release` 는 편집하지 않는다
- [ ] 2.8 **RED/GREEN — 엔진 전용 하한가 읽기 (9판 — design D12.1)** — 새 파일 `internal/official/a087_price_limits_raw.go`: `lowerLimitPrice *string` · `timestamp` 원문 보존.
      `Client.PriceLimits` · `adaptPriceLimits` · `domain.PriceLimits` · `tossctl quote limits` 출력 바이트 무변경(골든). 쓰는 조건 = timestamp KST 날짜 = 그날 거래일 **그리고** ≥ 09:00 KST.
      **사유별 행**: null · 파싱 불가 · 0 이하 · 정수 아님 · timestamp 결측 · 파싱 불가 · 다른 날 · 09:00 전 · 「오늘 수신한 전일 payload」(timestamp 가 전날) → 항목 없음 + 그 사유 줄 1회.
      판 값 → `sellIntent` → B5 미도달
- [ ] 2.9 **RED/GREEN — 재시작 재구성 (10판 — design D13.1)** — 새 파일의 읽기 전용 원장 함수 하나: 같은 `account_ref` · `market`(소문자) · `trading_day` · `UPPER(TRIM(symbol))` 의
      `intents` 중 `side='SELL'` 이고 `exit_events.proposed_intent_id` 로 가리켜진 판정의 `action` 이 `BASELINE_BREACH_EXIT` · `STOP_LOSS_LADDER` 인 행의 **존재 여부**. 가격 · 상태 · 출처는 읽지 않는다.
      갱신기는 항목을 내기 **전에** 이것으로 금지를 정한다. 원장 읽기 오류 · 파싱 불가 → 금지(fail-closed). **재시작 시험**: `IN_DOUBT → CONFIRMED → restart` → 관측가 LIMIT ·
      하한가 FAILED → 관측가 CONFIRMED → restart → 관측가 LIMIT · 다른 계좌 · 다른 시장의 같은 종목 intent 만 있음 → 하한가 · 익절 intent 만 있음 → 하한가 · 원장 오류 → 관측가 LIMIT.
      9판의 T2/T3 갈래는 10판에서 대체됨
      **11판 (design D14.1)**: action 은 리터럴이 아니라 `string(exitpolicy.ActionBaselineBreach)`(`BASELINE_BREACH_EXIT`) · `string(exitpolicy.ActionLadderStop)` — `isProtective` 와 같은 상수.
      **구조 시험**: `internal/exitpolicy` 의 `Action` 상수 전부를 AST 로 열거 → `isProtective` 참 집합 = 질의 action 집합(열거 0 · 참 집합 0 이면 실패 — 새 보호 action 이 생기면 실패).
      **행동 시험**: 기준선 이탈 하한가 시도(CONFIRMED · FAILED_CONFIRMED · IN_DOUBT→비수용 각각) → 재시작 → 그날 금지(관측가 LIMIT), 손절 사다리도 같은 세 행
      **12판 (design D15.2)**: 구조 시험은 상수 이름 수가 아니라 **평가된 값 집합** — `go/packages`/`go/types` 로 `internal/exitpolicy` 패키지 범위의 타입이 `exitpolicy.Action` 인
      `*types.Const` 전부를 모으고 `go/constant` 로 값을 평가한다(다른 파일 · `Action("…")` 형 선언 포함). 같은 값의 두 이름은 값 → 이름 목록으로 기록. 값 집합 · 참 집합이 비면 실패
- [ ] 2.6 **RED — 즉시성 · 경합** — 제출 경로 GET 0(호출 계수 더블) · 판 미준비에서 `submit` B5 미도달 · 갱신기 쓰기와 결정 읽기의 경합(`-race`, 반쪽 스냅숏 없음)
- [ ] 2.7 **변이** — 「창 조건 제거」 · 「창 끝 15:30 → 16:00 이상(시간외 누출)」 · 「익절에도 하한가」 · 「US 에도 하한가」 · 「날짜 대조 제거」 · 「하한가 ≤ 관측가 검사 제거」 ·
      「하한가 0 허용」 · 「되먹임 제거」 · 「되먹임을 B8(결과 미상)에도」 · 「결정에서 GET」 · 「결정 실패 → 오류 반환」 · 「`exitOrderLineFor` 가 사본 사다리 유지(줄 가격 ≠ 보낸 가격)」 ·
      「`NewOrderPath` 에서 판 생성」이 CAUGHT 인지 원장에 남긴다
      **9판 추가 변이**: 「timestamp 날짜 대조 제거」 · 「09:00 이전 허용」 · 「null 을 0 으로 합침(사유 하나)」 · 「토큰 없이 GET」 · 「주기당 2건 이상」 · 「429 정지 제거(세 입력 각각)」 ·
      「GET 시한 제거」 · 「되먹임을 B13 만」 · 「B8 에서 금지 안 함」 · 「CONFIRMED 뒤에도 금지」(이득 손실 쪽 — 잡히는지 기록) · 「재구성 생략」 · 「재구성 전에 항목 공개」
      **10판 추가 변이**: 「첫 시도 술어를 최신 상태로」 · 「계좌 조건 제거」 · 「시장 조건 제거」 · 「trading_day 조건 제거」 · 「익절 intent 도 셈」(이득 손실 쪽 — 기록) · 「원장 오류 → 허용」 ·
      「`Place` 결과가 CONFIRMED 면 금지 안 함」 · 「`doRequest` 가 헤더 있는 429 만 기록」 · 「429 기록 제거」 · 「정지 5분 → 0」 · 「하한가 문자열 비교」 · 「선행 0 허용」
      **11판 추가 변이**: 「질의 action 을 리터럴 `BASELINE_BREACH` 로」 · 「질의에서 손절 사다리 상수 제거」 · 「시작 정숙 구간 제거」 · 「정숙 5분 → 0」
      **12판 추가 변이**: 「admission nil 이어도 GET」 · 「admission 을 첫 요청에만」(401 재전송이 우회) · 「거절 뒤에도 GET」 · 「엔진 배선에 admission 주입」(a166 전 활성화 — 구조 시험이 잡아야 함) · 「구조 시험이 이름 수로 셈」

### 7판 원문 (8판에서 대체됨 — 시장가 유형 결정 · 달력 구간 입력)


- 2.0 **Pre-Edit 선언** — `ExitObserver.sellIntent` (WORKFLOW §Pre-Edit 형식)
- 2.1 **Function Logic Map** — `internal/app/engine/exitloop.go` / `ExitObserver.sellIntent`
      (`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`).
      `ast.json`의 `file`은 **저장소 상대 경로**. **HEAD 에서 재생성한다** — 기존
      `analysis/function-logic/internal-app-engine--exitobserver.sellintent/ast.json` 은 base `102d4e99`(start.line 1571)
      산출물이라 재사용 금지(재리뷰 P2-5; 본문 바이트 동일이라 분기 집합은 같을 것 — 검증 필요). risk-pattern-report 에
      `isProtective` 의 세 소비자(`record` 재판정 보류 예외·`submit` a091 floor·a087 유형 결정 — design D1)를 적는다
- 2.2 **RED** — `sellIntent` 분기: `BASELINE_BREACH_EXIT`→market·`STOP_LOSS_LADDER`→market·
      `LADDER_TAKE_PROFIT`→limit(가격은 종전대로 관측가)·시장가에는 가격이 실리지 않음.
      그리고 `isProtective` **술어 자체가 바뀌지 않음**을 고정(재리뷰 P3-2 — 세 행동 집합 그대로).
      **D7 구간 행 (2026-10-22)**: MARKET 구간에서 보호 2종 → market, LIMIT 구간(KRX 시간외단일가 · 접수 미측정 구간)과 구간 판정 불가에서
      보호 2종 → 오늘과 같은 limit(가격은 관측가 → 기준선 사다리 그대로)
- 2.3 **GREEN** — `sellIntent`가 `proposal`을 인자로 받아 `isProtective`로 유형을 정한다.
      현재 시그니처는 proposal을 안 받으므로 `submit`의 호출부도 함께 바꾼다
- 2.4 **MARKET 구간의** 보호 제안일 때 관측가·기준선을 **읽지 않는지** 확인 — 시장가 경로에서 가격이
      필요 없어졌으므로 "가격 없음" 거부(`has no price to submit a liquidation at`)가
      보호 청산을 막지 않아야 한다. **이것이 §0.3의 핵심 개선분이다**.
      (2026-10-22 D7 처분: LIMIT 구간에서는 이 문장이 무효 — 그 구간의 보호 청산은 오늘처럼 가격을 읽는다)
- 2.5 익절 경로와 **D7 LIMIT 구간의 보호 경로**는 가격 없음에 대해 종전 거부를 유지
- 2.6 **구간·접수 입력 배선 (design D7 「처분」 — 사용자 결정 2026-10-22)** — **착수 금지: `issues.md` D-1 사람 결정 대기**(0.10 독립 리뷰 P1-1 —
      KRX 단독 판정 입력 없음. (나)(다)를 고르면 2.6 은 다시 쓴다) — 엔진에는 오늘 세션 게이트가 없다(0.5: 청산 판정 상류
      AST 12개에서 세션 입력 0). 그래서 `sellIntent` 가 MARKET/LIMIT 을 가를 **구간·접수 입력**을 새로 받아야 한다.
  - 2.6.0 **Pre-Edit 선언** — 입력 출처 함수 · `sellIntent`(시그니처 변경) · `submit` 호출부 · 엔진 배선(WORKFLOW §Pre-Edit 형식).
        불변식 4: 이 입력이 거짓·결측이면 LIMIT(오늘 동작)으로 기운다
  - 2.6.1 **Function Logic Map — 입력 출처** — 출처는 확정됐다(사용자 결정 2026-10-28 Q3: 브로커 달력 `official.Client.TypedMarketCalendar` +
        고정 시각 `clock`, 달력 오류·결측이면 LIMIT — design D7 「구간 입력」). 0.10 리뷰용 HEAD `0c9bfd272` AST 가 `analysis/evidence/` 에 있다
        (`TypedMarketCalendar` · `scheduler.AdaptOfficialCalendar` · `adaptCalendarDay` · `CalendarSnapshot.ValidityAt` · `clock.Market.RegularSession` ·
        `Session.Contains` · `InRegularSession` · `Location` · 진입 쪽 `strategyScheduleAuthorityLoader.prepare`). 그 열거가 보인 것: `clock` 술어에는
        휴장·조기 폐장·개장 지연 분기가 없고, 달력 KR 정규장은 KRX·NXT 합집합이며 현재 Go 형은 KRX 기준 필드를 읽지 않는다. 착수 시 할 일: 본문을
        **고치는** 기존 함수(예: `official.MarketCalendarSession` 을 읽는 함수, 배선 함수 `Context.ExitObserver`)의 FLM 번들을 HEAD 에서 만들고,
        판정 함수는 **새 파일의 새 함수**로 둔다(비례 원칙 — gate ⑤ required 를 늘리지 않음). 출처가 D7 조건 ①~⑤ 를 못 가르면 멈추고 design D7 을 개정해 리뷰에 올린다
  - 2.6.2 **FLM 재생성** — 2.1 의 `sellIntent` FLM 을 시그니처 변경 뒤 분기 집합으로 같은 로트에서 다시 만든다(분기 번호는 위치다 —
        옛/새 `ast.json` 정렬)
  - 2.6.3 **RED** — 구간 표(design D7 7판) 전 행. 시계·달력은 시험 더블로 고정(벽시계·실제 브로커 금지):
        정상 영업일 KRX 정규장 접속매매(09:00:00 · 15:19:59) → market · 종가단일가(15:20:00 · 15:29:59) → limit(반열림) · 정상일 08:59:59 → limit ·
        KRX 시간외단일가 → limit ·
        KRX 장 마감 후/장 개시 전 → limit · **조기 폐장일**(합집합 끝 13:00, 또는 합집합은 정상인데 KRX 기준 종가단일가 시작 ≠ 15:20) 10:00·14:00 → limit ·
        **개장 지연일**(시작 10:00) 10:30 → limit · 휴장일(`integrated` null) → limit · 주말 → limit · **US 정규장 → limit** · US 프리/애프터 → limit ·
        달력 GET 오류 · 응답 없음(기동 직후) · `AdaptOfficialCalendar` 오류 · 읽은 지 6시간 이상 · 읽은 시각이 미래 · 오늘 날짜 불일치 · 시장 문자열 판정 불가 → limit.
        **limit 행은 `PlaceIntent` 전체가 오늘 `sellIntent` 출력과 같음**(유형만이 아니라 가격·수량·통화까지 — 골든). **골든은 변경 전 고정 fixture**
        (0.10 독립 리뷰 P2-2): 편집 전 HEAD 의 `sellIntent` 로 각 행의 `PlaceIntent` 를 literal 로 뽑아 RED 커밋에 먼저 커밋하고, GREEN 뒤에는 그 파일과만
        비교한다 — 바뀐 렌더러끼리(바뀐 함수를 다시 불러 기대값을 만드는 방식) 비교하지 않는다 ·
        **KR 정수 수량 행**(0.10 P1-2): 정상일 접속매매 + 수량 `1.5`·`1e-9` → **limit 제출**(B13 해제가 아님 · wire 에 소수 MARKET 0) · 수량 `3` → market ·
        **캐시 미준비 행**(기동 직후 스냅숏 nil · 갱신 실패) → limit, B5 미도달 · 스냅숏 교체와 판정의 경합(`-race`) · **판정 실패 행에서 청산이 거부되지 않음**
        (`submit` B5 거부 갈래에 닿지 않음) · **제출 경로의 달력 GET 0**(호출 계수 더블) · 정상일 정규장 중 장 개시 뒤 읽은 달력도 market(`ValidityAt` B3 미적용 고정)
  - 2.6.4 **GREEN** — 입력을 엔진에서 `sellIntent` 까지 배선. 구간 판정 실패는 LIMIT 갈래로(fail-to-today). 범위: KR 달력 응답을 메모리에 두는
        갱신(청산 관측 주기 밖 · 6시간보다 짧은 주기 · KR 만), `official.MarketCalendarSession` 에 `singlePriceAuctionStartTime` 추가(디코드 전용 — `scheduler.CalendarDay` 는 **무변경**: 그 JSON sha256 이 `Version` 이라 a112 권한·콘솔 값이 바뀐다, design D7 「스냅숏 형」), 판정 스냅숏 형·판정 함수(새 파일),
        `Context.ExitObserver` 배선. 엔진이 꺼지면 갱신·판정 0(불변식 3)
  - 2.6.5 **변이** — 「입력 결측 → market」·「시간외단일가 → market」·「구간 무시(항상 market)」·「US → market」·「조기 폐장 무시(합집합 창만 봄)」·
        「KRX 기준 필드 무시」·「고정 시각 조건 제거」·「종가단일가(15:20~15:30) → market」·「판정 실패 → 거부」·「제출 경로에서 달력 GET」·「`ValidityAt` 그대로 사용(B3)」·
        (0.10 독립 리뷰 P2-2) 「신선도 6시간 검사 제거」·「미래 `fetchedAt` 검사 제거」·「오늘 날짜 대조 제거」·「주말 검사 제거」·(P1-2) 「판정의 정수 조건 제거」이 CAUGHT 인지 원장에 남긴다
  - 2.6.6 ~~§5.2 결과 반영~~ — **not-applicable (2026-10-28 Q4)**: §5.2 를 측정하지 않으므로 반영할 결과가 없다. LIMIT 칸을 MARKET 으로 옮기는 것은
        새 실측 · spec 개정 · 리뷰 뒤에만 한다(이 change 밖 — spec Requirement 1 「옮기려면」 문장)

## 3. 원장·관측·화면 (Phase 2)

- [ ] 3.1 **RED/GREEN — a129 줄의 가격 · 출처** — `exitOrderLine` 에 `price_source`(`lower_limit` · `observed` · `baseline`), 줄의 `price` = 보낸 가격, `observed_price` ·
      `baseline` 은 그대로. 하한가 행 · 폴백 행 · 거부 행(B13 줄도 같은 출처)
      **9판 (design D12.6)**: a129 골든에 `price_source` 를 **필수 키**로(부분 비교 helper 가 빠뜨리지 못하게 키 존재를 따로 단언), 값 `lower_limit` · `observed` · `baseline` 각각.
      행: B1(하한 산술 실패) · B5(관측가 · 기준선 행) · B7 · B13 · B8 → 비수용 확정(비동기 결과 · 체결 따라가기 줄) · 폴백. 소비자 0 은 구현 시 재검색
- [ ] 3.2 **RED — 원장 가격** — 하한가 보호 청산의 `intents.price` = 하한가, 폴백은 오늘과 같은 값(schema 무변경)
- [x] 3.3 ~~원장 기록 경로 GREEN~~ — **not-applicable (8판)**: 원장 쓰기 경로는 `PlaceIntent.Price` 를 그대로 적는다 — 바꿀 코드가 없다(3.2 RED 가 고정)
- [x] 3.4 ~~화면 「시장가」 표시~~ — **not-applicable (8판)**: 가격 없는 청산 의도를 생산하지 않는다(design D4 머리말)

> **7판 원문 (8판에서 대체됨)**:
> - 3.1 **RED** — 시장가 청산의 `intents.price`가 비고 `order_type`이 market인지
> - 3.2 **RED** — 관측가·기준선이 제출가로 기록되지 **않는지** (주문된 적 없는 가격)
> - 3.3 **GREEN** — 원장 기록 경로
> - 3.4 **RED/GREEN** — 운영자 화면·알림이 가격 없는 청산을 "시장가"로 표시(결측 아님).
>       `operatorview`·콘솔 템플릿·`obs` 필드

- [x] 3.5 체결가 확정이 체결 조회 경로로 이뤄지는지 확인 — 이미 `filldetect`가 하는 일이면
      변경 없음을 테스트로 고정, 아니면 갭을 `issues.md`에 기록
      — 2026-10-22 HEAD `0686c0f0a`: **이미 하는 일, 갭 없음.** 체결가는 브로커 `execution.averageFilledPrice` 에서만 옴
      (`parseSnapshot` payload.go:91, 유형·intent 가격을 읽는 분기 0 — AST). 고정 `internal/filldetect/a087_market_exit_fill_pin_test.go`
      `TestA087MarketSellFillPriceIsTheBrokersAverage`(가격 NULL KR 시장가 매도의 스냅숏 평균 = 브로커 값, intent 무변경). 변이 M10 CAUGHT.
      `analysis/fill-price-3.5/`. production 편집 0
- [x] 3.6 **부분체결 잔량** (재리뷰 P2-3 · 1차 리뷰 A14 의 행방, design D8) — MARKET 보호 청산의 잔량이 남는지·`filldetect`
      가 그 잔량을 어떻게 닫는지 확인. 갭이면 `issues.md` 에 기록
      — 2026-10-22: 잔량이 남는지는 브로커 `[미측정]`. 체결 감지는 비종결 주문을 추적 집합(`terminal = 0`)에 두고 매 주기 읽다가 종결
      (`CANCELLED_PARTIALLY_FILLED` 등) 관측으로 닫음 → `ApplyExitFill` 이 발의를 풀고 다음 관측이 남은 포지션을 재판정. 고정
      `TestA087MarketSellRemainderIsTrackedUntilTheBrokerEndsIt`(변이 M11 CAUGHT). **갭**: 잔량이 비종결로 남는 동안 재발의 억제 ·
      지연 경보 없음 → `issues.md` I-R9. 「종결 → 재발의 수량 = 잔량」 엔진 종단 시험은 없음(미검증). `analysis/partial-fill-3.6/`
- [x] 3.7 **`intents.price` NULL 소비부 전수 열거** (재리뷰 P3-3 — 표본 셋 `fills.go` coalesce·`exit_held_proposal.go`
      `workingOrderPrice`·`indoubt.go` 가격 와일드카드/notional 0 은 통과, 전수 아님). **방법 (3차 재리뷰 A-P3-3)**: CodeGraph 는
      심볼 단위라 SQL 열의 읽기 자리를 열거하지 못한다 — ① 비시험 Go 소스의 SQL 문자열에서 `intents` 와 `price` 를 함께 읽는 자리를
      `rg` 로 모으고, ② 그 질의를 감싼 저널 접근자 함수의 호출자를 CodeGraph 로 따라가 ①∪② 를 `analysis/` 에 남긴다. 각 자리의 NULL
      처분을 시험 또는 not-applicable 사유로 적는다
      — 2026-10-22 HEAD `0686c0f0a`, codegraph 1.6.0: ① `census.go`(비시험 Go 969 파일) → SQL 3 자리 중 price 를 읽는 것 2(`intentSelect`
      → `scanIntent` NullString · `LiveOrdersForSymbol` coalesce; `confirmedFillOwners` 의 `*` 는 price 없는 CTE). ② CodeGraph ∪ `flow.go`
      → 접근자 호출 자리 12(CodeGraph 가 `exitwiring.go:266` `localOpenSells` 를 놓침 — 합집합에 넣음). NULL 이 닿는 Go 소비 4
      (`Prepare` 비교 · `workingOrderPrice` · `parseOptionalDecimal`→matcher 와일드카드 · `intentNotional`) 전부 시험, 나머지 not-applicable
      (price 미사용). 신규 시험 5(journal 2 · execgw 2 · engine 1) + 기존 1, 변이 M12~M16 CAUGHT. 표·AST: `analysis/intents-price-null/`.
      부수 기록 `issues.md` I-R10. production 편집 0
- [x] 3.8 **청산 슬리피지 산출** (재리뷰 P1-3, design D6 · 3차 재리뷰 A-P2-4 정정) — **기존 칸 유도, schema·production 코드 추가
      없음**(proposal Impact 「Schema: 없음」과 일치). 결정 시 관측가 `exit_events.observed_price`(같은 행 `proposed_intent_id`) 와 그
      intent 의 브로커 `order_id` 로 찾은 체결 평균가 `fill_snapshots.average_price` 로 bps 를 계산하는 **읽기 전용 저널 질의**를
      `analysis/` 에 둔다(착지 뒤 보호 청산마다 사람이 돌림). 「원장에 남긴다」는 철회 — 새 열을 쓰지 않는다. 착수 시 intent → `order_id`
      결속이 기존 칸으로 성립하는지 먼저 확인하고, 성립하지 않으면 **멈추고** proposal Impact 를 schema 변경으로 고쳐 재리뷰를 받는다.
      비례 원칙: 보호 청산 판정을 바꾸지 않는 관측 산출이라 High-risk 무거운 규율(FLM·변이 원장) 대상 아님(`slippagePct` 는 진입 전용 —
      무변경)
      — 2026-10-22: **결속 성립 — schema 변경 불요**(proposal Impact 유지). 발의 행이 `observed_price`·`action`·`proposed_intent_id` 를 한
      행에 씀(`exit_state.go` INSERT INTO exit_events) → `mutation_attempts`(PLACE·CONFIRMED·`broker_order_id`) → `scoped_fill_snapshots`
      (order_id + intent 전체 범위 — 재사용 번호 차단; 범위 같은 행이 scoped 에 없을 때만 `fill_snapshots`). 질의
      `analysis/exit-slippage-3.8/exit_slippage.sql`(adverse_bps, 스냅숏 없으면 NULL). 시험 `internal/journal/a087_exit_slippage_query_test.go`
      (마이그레이션이 만든 임시 저널, 실제 저널 미접촉) — 질의 변이 Q1~Q3 CAUGHT(`query-mutants-2026-10-22.log`). 한계: AMEND 후속 주문 미추적

## 4. 회귀 방지

- [x] 4.1 `flatten` 무변화 — 이미 거래소 하한가 1순위이고 이 change의 범위 밖임을 테스트로 고정.
      `flatten.sell`/`Liquidate` 는 `checkOrderShape` 상류다(`analysis/gate-0.4-codegraph.md`) — ② 변경이 flatten 의 LIMIT 을 바꾸지 않음
      — `internal/flatten/a087_flatten_limit_pin_test.go` `TestA087FlattenStillSellsAtAnAggressiveLimit`(하한가·체결가 할인 두 행, 실제 Gateway 경유).
      변이 2/2 CAUGHT(`analysis/pin-mutations/`)
- [x] 4.2 `verifylive` 무변화 — 검증 도구의 "체결되면 안 되는 지정가" 성질 유지
      — `internal/verifylive/a087_far_limit_only_pin_test.go` `TestA087VerifyPlacesOnlyFarLimitOrders`(전체 2회 실행의 생성 7·정정 1 전수:
      LIMIT·시장성 없는 가격). 변이 2/2 CAUGHT(`analysis/pin-mutations/`)
- [x] 4.3 조건주문 경로 무변화 (2c 범위)
      — `internal/trading/a087_conditional_path_unchanged_pin_test.go` `TestA087ConditionalMarketStopBypassesThePlacePredicateUnchanged`
      (기본 Service 에서 ③ 이 KR 시장가 매도 비지원인 채로 SINGLE+MARKET 매도 조건주문이 의도 그대로 브로커에 닿음). 변이 2/2 CAUGHT(`analysis/pin-mutations/`)
- [ ] 4.4 upstream 상속 테스트 650 green 유지
- [x] 4.5 **전송 표류 가드 KR market sell 골든 행** (재리뷰 P3-3) — `decision_test.go` 표류 가드 골든은 US market sell 만 있다.
      KR market sell 행을 더해 `buildOrderCreate` 와 사본 `PlaceWireBody`(`execgw/wirebody.go`)가 둘 다 가격·TIF 를 싣지 않음을 고정
      — 기존 파일 대신 새 파일 `internal/execgw/a087_kr_market_sell_wire_pin_test.go` `TestA087KRMarketSellWireBodyCarriesNoPriceOrTIF`
      (같은 골든을 `PlaceWireBody` 와 httptest 로 잡은 `official.PlaceOrder` 전송 바이트 양쪽에 대조, 가격이 남은 intent 행 포함). 변이 3/3 CAUGHT
- [x] 4.6 ~~**재생 경로** (재리뷰 P3-3) — `PlaceWireBody` 를 쓰는 재생 경로가 MARKET 보호 청산 intent 를 가격 없이 재구성하는지 시험~~
      — **not-applicable (8판)**: MARKET 보호 청산 intent 를 생산하지 않는다. 하한가 LIMIT intent 의 재생은 오늘의 LIMIT 재생과 같은 경로다(재생 자체는 2.7 결정대로 꺼져 있음 — a165 proposal)

## 5. 실측 (사용자 승인 항목 — §0.7, 자동 실행 금지)

> **2026-10-05 큐 편성(사용자 승인 — a100 review.md 「사용자 결정 — 2026-10-05」 3항):**
> 5.1·5.2 를 a100 M-A 의 KR 장중 사람 실측 세션과 같은 큐에 올린다 — 같은 세션에서 측정해도
> 되고 별도 세션이어도 된다. 주문별 사람 즉시 승인·자동 실행 금지는 그대로다. a100 은 이
> 실측을 기다리지 않는다(a100 D8 계약 1 (다) 채택 — Phase 2 는 a100 선행 조건이 아니다).
>
> **2026-10-10 보류(2차 재리뷰 P1-1) → 같은 날 실행 수단 확정(사용자 결정 2):** 5.1 은 출하 도구로 실행 불가였다 —
> `tossctl order place` 도 엔진과 같은 `placeIntentSupported`(`internal/trading/service.go`, 비분수 `OrderType != "limit"`
> → `ErrPlaceUnsupported`)에서 **브로커에 닿기 전에** 로컬 거부되므로, 그 결과를 기록하면 우리 관문을 브로커 응답으로
> 적는 거짓 음성이 된다. 조건주문(MARKET 허용)은 다른 endpoint 라 `/api/v1/orders` 를 재지 못한다.
>
> **실행 수단 = 별도 실측 도구 `tools/a087-market-sell-probe/`** (사용자 결정 2026-10-10 — `internal/official` 직접 호출,
> 토큰 캐시 공유, 새 파일만, production 변경 **전에** 측정; 병행 팀메이트 구현 — 도구의 명령·출력 형식은 그 도구의 문서가
> 정본이며 이 tasks 는 선취하지 않는다). 주문 생성은 mutating — **실행은 사람 터미널 + 주문별 승인**, 에이전트는 준비까지.
> 대상 수량은 010170 매수 산식에 이미 포함(5.1 매도 1주·5.2 경계 1주). M-A 동승 큐(R2)는 이 도구가 준비되고 0.7 Eng
> 재리뷰가 끝난 뒤 재개한다. a100 `ma-runbook.md` R2 에 실행 명령이 없던 공백(재리뷰 P1-1)은 이 도구 경로로 메운다
> (runbook 쪽 반영은 이 디렉터리 범위 밖 — `issues.md` I-R5).

> **실행 전제 (2026-10-10, 3차 재리뷰 B-P1-1·A-P2-5·B-P2-1·B-P3-3·B-P3-4)** — 사람이 5.1·5.2 를 치기 직전에 확인한다:
>
> 1. **도구 소스가 착지 커밋 그대로일 것**: `git status --porcelain -- tools/a087-market-sell-probe` 가 비어 있고,
>    `git log -1 --format=%H -- tools/a087-market-sell-probe` 가 리뷰를 통과한 착지 커밋이다. 변이 하네스(`mutants.sh`)는
>    `45167aa7` 부터 저장소 밖 사본에서만 변이하므로 공유 트리를 바꾸지 않지만, 병행 세션의 미커밋 편집은 이 검사로만 잡힌다.
> 2. **매도가능수량 사전 스냅숏 (👁 read-only)**: 5.1 종목의 매도가능수량과 **같은 종목의 상주 SELL 조건주문**(a100 M-A 당일 등록분,
>    R0-3 기본 retain 잔여분)을 조회해 1주 여유를 확인한다. 도구는 이 표면을 일부러 읽지 않는다(`static_test.go`
>    `TestNoProductionOrderSurfaceIsNamed`). 실측 당일 a100 runbook §2 종목 확정 때 함께 본다 — 겹치면 5.1 종목을 바꾸거나
>    결과를 아래 표 「판정 불가(수량)」로 읽는다. 계좌 선택은 `accounts[0]`(도구는 계좌를 영수증에 적지 않음 — B-P3-6)이라
>    이 스냅숏이 「그 계좌에 그 종목이 있다」의 유일한 확인이다.
> 3. **결과 `unknown`(또는 `sending`/`no-answer` 영수증이 남음) 뒤에는 미리보기를 다시 돌리기 전에** `tossctl orders list` /
>    `tossctl orders completed` 로 그 종목 주문을 먼저 확인한다 — 새 미리보기는 새 멱등 키라 둘째 실주문이 된다(B-P3-4, 도구 잠금
>    대신 절차 — `issues.md` I-R7).
> 4. 확인 토큰의 「5분」은 검증 시점 기준이다 — 헤더 획득(토큰 교환·계좌 조회)이 끼어 실제 전송은 창 끝을 수십 초 넘을 수 있다.
>    브로커 멱등 창 10분 안이라 이중 주문 위험은 없다(B-P3-3).
>
> **영수증 디렉터리는 커밋 의도다 (B-P2-1 처분 — Manager 지시 2026-10-10)**: 기본 `--out` =
> `analysis/market-sell-receipts/` 의 영수증 JSON 을 측정 영수증으로 커밋한다. 담기는 것 — orderId·원 응답 본문·종목·수량·시각·
> 보낸 헤더 **이름** — 중 orderId 는 계좌 개인정보가 아니다(허용). 토큰·계좌 번호·API 키는 도구가 쓰지 않는다(헤더 값 미기록,
> sentinel 5종 시험). 원 응답 본문의 `error.data` 에 수량 값이 실릴 수 있으나 계좌 식별 정보가 아니다.

- [x] 5.1 **KR MARKET 매도 1회.** 최소 수량(1주)·장중. 스키마는 지원하나 실접수 미측정.
      수단 `tools/a087-market-sell-probe/`. 성공·실패 모두 응답 코드·본문 code·주문 상태를 기록. 처분은 아래 표
      — 2026-10-15 시도 2회(15:13 · 15:16 KST, 010170 1주, 사람 터미널): 둘 다 **HTTP 401 `invalid-token`** → 처분 「판정 불가(인증)」, MARKET 반증 아님. 주문 미접수 예상(401). 영수증 `analysis/market-sell-receipts/a087ms-20261015T151103-0a2582f5.json` · `…T151445-9157ce87.json`. 토큰 캐시 mtime 15:26:58 — 같은 OpenAPI 키 공유 이웃 제품의 토큰 재발급 정황(기억 「공유 credential 토큰 전쟁」), 미검증. 재시도는 다음 장중
      — **2026-10-19 재시도 = 처분 「accepted」(FILLED).** 사람 터미널, 010170 1주, `--config-dir ~/.config/tossctl`(엔진과 같은 토큰 캐시).
      영수증 `analysis/market-sell-receipts/a087ms-20261019T145815-8bdcb862.json`: 보냄 14:59:19.102 KST → HTTP 200 · orderId 수신 · clientOrderId 메아리 일치(14:59:19.251).
      주문 상태(👁 read-only `tossctl --config-dir ~/.config/tossctl orders completed`): 010170 SELL 1주 **체결완료**, 브로커 접수 14:59:19.277 KST —
      그 화면은 OpenAPI orderId 를 싣지 않아 시각(+0.2s) · 종목 · 방향 · 수량으로 대조했고 같은 시각의 다른 010170 매도 1주는 없음. 체결 품질은 재지 않음(도구 범위 밖).
      10-15 의 401 원인 정정: 그날 명령에는 `--config-dir` 가 없어 도구가 기본 토큰 캐시(`~/.cache/tossctl`)로 새 토큰을 받는 구조였다 — 2026-10-19 같은 구조의 도구가
      엔진 토큰과 번갈아 무효화하는 것을 실측(인수인계 §4 정정). 이웃 제품 가설보다 이쪽이 유력(10-15 당시 재현은 안 함)
- [x] 5.2 **세션 경계.** 정규장 밖 KR MARKET 매도(1주)의 응답 코드 관측. **예상 응답 집합**(재리뷰 P1-4):
      `422 order-hours-closed` · **`422 order-type-not-allowed`**(openapi `POST /api/v1/orders` 422 예시 「현재 사용할 수 없는
      호가 유형」 — 저장소 Go 코드 참조 0, `refusal_code.go` 목록에도 없음) · 그 밖의 코드(있는 그대로 기록).
      시간외단일가에는 시장가가 없다
      — **not-applicable (사용자 결정 2026-10-28 Q4)**: 실측하지 않는다. 사유 — design D7 이 정규장 밖의 모름을 LIMIT 으로 고정했으므로 실측의 이득은
      「나중에 그 칸을 MARKET 으로 옮길 근거」뿐이고, 실측은 실주문이 다음 세션에 체결될 위험이 있다(아래 표 「세션 밖인데 200」 행). 실주문 0 · 5.2 몫
      (010170 경계 1주)은 쓰지 않는다. design D7 「§5.2 처분」
- [x] 5.3 `openspec/changes/verify-execution-capability/measurements.md`에 M계열로 기록
      (`docs/trading/measurements.md`는 **존재하지 않는다** — 초안 리뷰 M7)
      — 5.1 분만 기록(2026-10-20): 「9차」 절 **M50**(KR 정규장 MARKET 매도 1주 접수·`체결완료`) + 10-15 401 두 시도 처분.
      ~~5.2 미측정이라 미체크~~ — **2026-10-28 닫음**: 5.2 가 not-applicable(Q4)이라 기록할 M 행이 더 없다. 기록 대상은 M50 하나로 끝난다

- [ ] 5.4 **(9판) 착지(운영 배포) 전 필수 gate — KR 하한가 지정가 매도 1주 + timestamp 의미** — 사람 승인 · 사람 터미널 · 자동 실행 금지. 구현 로트가 끝난 뒤, 평일 KRX 정규장 접속매매 중:
      ① 👁 엔진 전용 읽기 형(2.8)으로 그날 하한가 · `timestamp` 원문을 읽고 `tossctl order place` 로 그 가격의 LIMIT SELL 1주(사람 CLI 지정가 경로는 upstream 그대로 — 플래그는 `--help` 정본).
      기록: 접수 응답 · 주문 상태 · 체결 평균가 · 같은 시각 매수 1호가(👁) · 관측가. ② 👁 **timestamp 가 무엇을 가리키는지** — 같은 종목을 장전(09:00 전)과 장중에 읽어 하한가 값 ·
      timestamp 가 각각 어떻게 바뀌는지(효력 날짜인가 응답 시각인가). **처분 — 사람 결정 후보(`issues.md` D-2 ②)**: 수용 → 착지 진행 · 거부 → 범위를 KR off 로 축소(보호 매도 전부 오늘 LIMIT) ·
      판정 불가 → 착지 보류. timestamp 가 응답 시각이면 D12.1 조건 ①② 는 신선도 증명이 아니므로 그 사실과 처분을 같은 gate 에서 사람이 정한다

### 실측 결과 → 처분 (재리뷰 P1-1·P1-4, 사용자 결정 2026-10-10)

| 측정 | 결과 | 처분 |
| --- | --- | --- |
| 5.1 | **수용** — HTTP **200**(openapi `POST /api/v1/orders` 성공 응답은 200 하나 — 201 없음) **그리고** 사람이 `tossctl orders list` / `tossctl orders completed` 로 영수증 orderId 의 상태를 읽어 `PENDING`·`PARTIAL_FILLED`·`FILLED` 중 하나 | Phase 2 전제 성립. 나머지 착수 조건(위 §1 머리 ②~④) 충족 시 **§1~§3 착수 가능**. 체결가·관측가는 기록만(bps 판정 근거로 쓰지 않음 — n=1, design D6). 도구의 `accepted` 는 200 + orderId 일 뿐이고 이 행이 아니다(3차 재리뷰 A-P1-2) |
| 5.1 | **200 뒤 `REJECTED`** — 접수 응답은 200 이었으나 조회 상태가 `REJECTED`(openapi 주문 상태 어휘 「브로커가 주문을 거부한 상태」, MARKET·REJECTED 예시 존재) | 거부 사유가 MARKET 유형 자체이면 아래 「거부」 행, 종목 상태·수량 등 유형 무관 사유이면 「판정 불가」 행으로 읽는다. 사유를 못 읽으면 「판정 불가」 |
| 5.1 | **거부** — `422 order-type-not-allowed` 등 MARKET 매도 자체를 거부하는 4xx | **Phase 2 전제 반증.** production 무접촉(§1~§3 착수 안 함). change 처분(불구현 아카이브 / design D6 하한가 지정가 원안 재검토 등)은 **사람 결정** |
| 5.1 | **판정 불가 — 종목 상태 사유** — VI 발동 중 단일가·투자경고·단일가 매매 종목·거래정지 등 **그 종목의 그 시각** 상태를 말하는 거부(4xx 또는 200 뒤 `REJECTED`) | MARKET 유형 일반의 반증이 아니다. 구분 기준: 오류 code·message 가 종목·시장 상태를 가리키면 이 행, 주문 유형을 가리키면(`order-type-not-allowed` 류) 「거부」 행. 애매하면 이 행. 다른 종목·시각으로 재측정할지는 사람 결정 |
| 5.1 | **판정 불가 — 매도가능수량 사유** — 같은 종목의 상주 SELL 조건주문(a100 M-A 당일 등록분·R0-3 retain 잔여분)이나 다른 미체결이 매도가능수량을 묶어 수량 사유로 거부 | MARKET 반증 아님. 위 「실행 전제」 2 의 사전 스냅숏으로 피해야 하는 행이다 — 실측 당일 a100 runbook §2 종목 확정 때 겹침을 확인하고(M-A retain 결정과 교차: 잔여 조건주문은 취소하지 않으므로 5.1 종목을 바꾼다), 그래도 나오면 기록 후 사람 결정 |
| 5.1 | **판정 불가 — 그 밖** — 인증(401·403)·한도(429)·장 시간 등 MARKET 유형과 무관한 사유의 4xx, 5xx·타임아웃·409·in-doubt(도구 `unknown`) | 수용도 반증도 아니다. 원인·주문 상태(브로커 조회로 확정)를 기록하고 재측정 여부는 사람 결정. §1~§3 착수 안 함. 도구의 `refused` 는 4xx 라는 뜻뿐이라 이 행일 수 있다(B-P3-2) |
| 5.2 | ~~세션 밖 — 거부 응답~~ — **5.2 not-applicable(2026-10-28 Q4) — 아래 5.2 두 행은 실행하지 않으므로 기록으로만 남긴다** · 세션 밖 — 거부 응답(`422 order-hours-closed`·`order-type-not-allowed`·그 밖) | **응답 코드 기록.** 처분은 사용자 결정 2026-10-22(design D7 「처분」)로 정해졌다 — 그 구간의 보호 청산은 LIMIT(오늘과 같음)으로 확정. MARKET 유형 거부(`order-type-not-allowed`)가 아니라 장 시간 거부(`order-hours-closed`)이면 그 구간은 LIMIT 접수 여부도 따로 봐야 하므로 기록 후 사람 결정 |
| 5.2 | **세션 밖인데 200 으로 접수됨** (장전 동시호가 대기·익일 이월 등) | 실주문이 다음 세션에 체결될 수 있다. 즉시 `tossctl orders list` 로 상태를 읽고 **취소 여부는 사람이 그 자리에서 결정**(취소는 mutating — 사람 실행). 체결되면 010170 매수 산식의 5.2 몫(경계 1주)을 소진한 것으로 기록하고 추가 측정 수량을 다시 잡지 않는다. D7 표의 해당 구간은 「MARKET 접수됨(체결 시점 미확정)」으로 기록 |
| US | 측정 없음 | 자동 처분 없음. §1~§3 착수 전 사람이 「US 를 미측정으로 착지」(proposal 「US 한계」)를 확인한다 → **2026-10-28 확인 처리(Q2): US 미개방** — US 보호 청산은 실측 전까지 LIMIT(design D7 7판). 미측정 착지하는 US MARKET 칸이 없다 |

상태 어휘는 openapi 주문 상태 enum 10개(`PENDING`·`PENDING_CANCEL`·`PENDING_REPLACE`·`PARTIAL_FILLED`·`FILLED`·`CANCELED`·`REJECTED`·
`CANCEL_REJECTED`·`REPLACE_REJECTED`·`REPLACED`)다. 위 표가 이름 붙이지 않은 상태(예: 사람이 취소하지 않았는데 `CANCELED` — 잔량 소멸 등)는
**판정 불가**로 읽고 그 상태와 체결 수량을 그대로 기록한다(수용으로 올려 적지 않는다).

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0
- [ ] 6.2 `make sdd-sync` 재실행 (마지막 파일 편집 후)
- [ ] 6.3 `make sdd-check`
- [ ] 6.4 **격리 worktree에서** `make gate CHANGE=a087-a-protective-exit-is-a-market-order`
- [ ] 6.5 독립 검증 (구현과 분리된 컨텍스트)
- [ ] 6.6 PM 동기화 → `openspec archive`

## 후속 change (이 change에서 하지 않는다)

> **정정 (2026-10-10, 재리뷰 P3-1)**: 3판까지 이 표는 a088·a089 를 소유자로 적었다. 「a088」은 `openspec/changes`·`archive`
> 어디에도 없고, a089 는 다른 내용(「나가지 못한 손절은 세어진다」)으로 2026-09-28 불구현 아카이브됐다. 아래는 **무소유** 장부다.

| 소유자 | 내용 | 근거 |
| --- | --- | --- |
| **무소유** (옛 표기 「a088」) | 호가 그리드 정본화 — StockOS 형태 이식 | 익절 지정가·`flatten` fallback이 필요로 한다. 초안 리뷰의 A1·A4·A5·A10·A11·A12 |
| **무소유** (옛 표기 「a089」) | 긴급 청산 재가격 에스컬레이션 + 게이트 우회 | 7분 42초 공백은 주문 가능한 제안이 없던 구간(proposal Why 정정). 비종결 **잔량**의 재가격·재주문만은 아래 a167 행으로 갔다 |
| **a167** `a167-a-protective-remainder-is-re-priced` (`STORY-TOS-a167`, proposal 수준) | 비종결 보호 청산 잔량의 자동 재가격·재주문 | `issues.md` I-R9 처분(사용자 결정 2026-10-22), design D10 |
| **무소유** | 1차 리뷰 A1~A14 중 위 두 행 밖의 것, I2(400 분류) | a094 R1 이 가져간 것은 `opposite-pending-order-exists` 하나(`refusal_code.go`) |
| **무소유** | `order-type-not-allowed` 422 분류 | `issues.md` I-R2 |
