# a094 · tasks

- **Change**: `a094-a-stop-clears-what-blocks-it`
- **위험 등급**: **High-risk** — 손절 주문의 분류와 충돌 해소. §0.3 손절 즉시성 적용.
- **base-commit**: `ec29dc72c0fd589daa2069ccf26bad26baeb2a04`

## 0. 게이트 선행

- [x] 0.1 `base-commit.txt` 고정
- [x] 0.2 `openspec validate a094-a-stop-clears-what-blocks-it --strict` — **통과**
- [x] 0.3 **AST 산출물이 문서보다 먼저** — 함수 9개, 분기 79개
- [x] 0.4 `check_analysis.py --change a094-…` — **통과**(`evidence complete`).
      두 Markdown은 **측정으로** 채웠다: 조건은 소스 원문, 창의 호출·return은
      `ast.json` 좌표, 진입 여부는 `go test -covermode=set` 프로파일이다
- [x] 0.5 **proposal-freeze 리뷰 1라운드**(적대적 Eng) → `review.md` §1. **FAIL** —
      차단 8건. 보이스 셋 전부 FAIL이나 **증거 사슬은 깨끗했다**(좌표 60/60 · 분기쌍
      38/38 · 조건 158/158 · 창 79/79 · sha256 9/9 · 커버리지 독립 재현 일치).
      거부된 것은 처방이다. **교차 모델 미충족** — 사용자 지시로 Claude 보이스 셋
      (`클로드로 돌리세요`, 2026-08-06)
- [x] 0.5a **2판 반영** → `review.md` §2. R1 필드 파싱 · R2 재작성 · R3 진입점 교체 ·
      R4 design 반영 · spec 정본 보존 프로그램 확인. **판정 아님**
- [x] 0.5b **proposal-freeze 리뷰 2라운드**(gstack plan-eng-review) → `review.md` §2.
      **FAIL** — 차단 8건. **P0: 잠금을 잘못 지목하고 있었다**(§2.1).
      **교차 모델 미충족** — Codex 사용량 한도(2026-08-08 12:36 복구), Claude 서브에이전트 대체
- [x] 0.5c **3판 반영** — AST 6개 추가(101분기) → design D−1·D1 소급·D2 축소·D3 교체 ·
      spec delta 2개 재작성 · tasks §3·§4·§4bis
- [x] 0.5d **proposal-freeze 리뷰 3라운드**. **교차 모델을 여기서 지킨다** —
      a092 여섯 + a094 두 라운드가 미충족이다 → codex(gpt-6-astra, session `01a0e2d0-a6e1-7322-a506-c5dbde92c8e9`)
      **REJECT**(P0 3 · P1 5 · P2 2), `review.md` 「3라운드」 · `analysis/freeze-review/codex-r3-output.md`. **교차 모델 충족**
- [x] 0.5e **4판 반영** — `design.md` D−2(F1 park 해제 철회 · F2 원 발주 응답 한정 · F3 엔진 귀속 축소 + 외부 취소는 사용자
      결정 대기 · F4 기대 intent·기동 따라잡기 · F5 스냅숏 계약은 선행 조건으로 · F6 기동 순서 재핀 · F7 두 결말 · F8
      PENDING_CANCEL 정의 · F9 AC1 범위 · F10 낡은 논증), spec delta 2 · tasks §3·§4·§4bis 개정. **판정 아님**
- [ ] 0.5f **proposal-freeze 리뷰 4라운드**(codex 교차 모델) — Manager 대기열 순서대로
- [ ] 0.5g **4판 편집 전 산출물 재작성**(F10) — `record` 번들의 FLM 「Branches」 논증과 BTM 미진입 요약을 현재 AST(16분기)
      기준으로 다시 쓰고, 4판이 편집하는 기존 함수(`ResolveExitProposal` 호출 형태 · 기동 이음매)의 FLM 을 편집 전에 갖춘다

## 1. 산출물 (완료 — 문서보다 먼저)

> **3판이 6개 101분기를 더했다.** 합계 **함수 15개 · 분기 180개**.
> 새 6개는 `EvaluateLadder`(32) · `EvaluateRatchet`(22) · `armExitProposalTx`(4) ·
> `ResolveExitProposal`(14) · `RecoverPending`(10) · `Detector.collect`(19)이며,
> **D−1이 지목한 진짜 잠금과 D3의 해동 경로가 전부 그 안에 있다.**

- [x] 1.1 `journal.isDefinitiveRejection` (분기 3 · return 2)
- [x] 1.2 `journal.ClassifyHTTPMutation` (분기 7 · return 6)
- [x] 1.3 `execgw.classifyMutation` (분기 7 · return 5)
- [x] 1.4 `execgw.checkSymbolFree` (분기 9 · return 8)
- [x] 1.5 `engine.submit` (분기 11 · return 9)
- [x] 1.6 `engine.record` (분기 14)
- [x] 1.7 `engine.clearTheSymbol` (분기 9 · return 4)
- [x] 1.8 `journal.LiveOrdersForSymbol` (분기 7 · return 6)
- [x] 1.9 `reconcile.Run` (분기 12 · return 8)
- [x] 1.10 **Branch Test Map** — 위 9개. **분기 79개 중 진입 49 · 미진입 27 · 자체 블록 없음 3**
      (`go test ./internal/{journal,execgw,app/engine,reconcile}/... -count=1 -covermode=set`)

      | 함수 | 분기 | 미진입 |
      | --- | --- | --- |
      | `journal.ClassifyHTTPMutation` | 7 | **0** |
      | `journal.isDefinitiveRejection` | 3 | 0 (블록없음 1) |
      | `execgw.classifyMutation` | 7 | 2 |
      | `execgw.checkSymbolFree` | 9 | 4 |
      | `engine.submit` | 11 | 5 |
      | `engine.record` | 14 | 2 |
      | `engine.clearTheSymbol` | 9 | **5** |
      | `journal.LiveOrdersForSymbol` | 7 | 2 |
      | `reconcile.Run` | 12 | **7** |

      **가장 무거운 실측**: `clearTheSymbol` **B3** `:1343`
      (`if !buy && !withPending`)이 **미진입**이다 — 기존 시험이 매수 건너뛰기 술어를
      한 번도 밟지 않는다. a094가 바로 그 술어 위에 R2를 얹으므로 3.1이 이것을 먼저 덮는다.
      `reconcile.Run`의 미진입 7개는 `B1,B2,B4,B5,B8,B9,B11`이다 — 해소 경로(B5~B9)에
      드는 것은 **3개뿐**이고 나머지 넷은 그 밖이다(1라운드 정정)
- [ ] 1.11 `classifyRefusalBody`의 **소비자 조사** — 새 reason code가 닿는 자리
      (`AllReasonCodes()` 고정 테스트 · 콘솔 필터 · 원장 질의 · Phase 2 ledger)

## 2. R1 — code가 분류한다 (D1)

- [ ] 2.0 **Pre-Edit 선언** — `internal/execgw/failclosed.go`, `internal/execgw/reason.go`
- [ ] 2.1 **RED** — 409 + `code=opposite-pending-order-exists` → `DispatchRejected`,
      attempt **종결**, `PendingAttempts`에서 제외
- [ ] 2.2 **RED** — **422** + 같은 code → 같은 결과 (계약대로 왔을 때도 같아야 한다)
- [ ] 2.3 **RED** — 409 + `code=request-in-progress` → **종전대로 Ambiguous**.
      **이 케이스가 R1의 안전 경계다** — 깨지면 살아 있는 주문을 은퇴시킨다
- [ ] 2.4 **RED** — code 없는 409 → 종전대로 Ambiguous
- [ ] 2.5 **RED** — message에만 그 문구가 있고 code는 다름 → **분류되지 않는다**
      (D0: message로 걸지 않는다)
- [ ] 2.5a **RED** — 본문의 code가 **`error` 아래**에 있어도 잡힌다
      (프로덕션 3건의 실물 모양: `{"error":{"requestId":…,"code":"opposite-pending-order-exists",…}}`)
- [ ] 2.5b **RED** — 본문의 code가 **최상위**에 있어도 잡힌다
      (`testdata/interactive_auth_challenge.json`의 모양). **두 자리를 다 읽는다**
- [ ] 2.5c **RED** — code 값 비교는 **대소문자 무시 + 전체 일치**다.
      `opposite-pending-order-exists-v2` 같은 값은 **잡히지 않는다**(substring 아님)
- [ ] 2.5d **RED** — JSON이 아닌 본문 · code 필드 부재 · 빈 code → **분류하지 않음**
      (fail-closed: 모르면 종전 경로)
- [ ] 2.6 **RED** — `isDefinitiveRejection`의 상태 목록 **무변화**를 표로 고정
- [ ] 2.7 **RED** — 기존 세 code(`trade_auth_required`·`fx_consent`·`funding_required`)의
      분류 **무변화**
- [ ] 2.8 **GREEN** — `ReasonOppositePendingOrder` 추가 + `AllReasonCodes()` 등록 +
      **`classifyRefusalCode(body string) (ReasonCode, bool)` 신설**.
      `code`와 `error.code`만 읽고, `classifyRefusalBody`보다 **먼저** 부른다.
      **기존 세 항목의 substring 매칭은 건드리지 않는다**(D0)
- [ ] 2.9 `submit`이 B10 `:1304`로 가는 것을 확인 — `alertProposalRefused` + 레벨 재무장
- [ ] 2.10 **golden 갱신** — `internal/execgw/testdata/reason_codes.golden`에 새 code 한 줄.
      `TOSSOS_UPDATE_GOLDEN=1 go test -run TestWriteReasonCodeGolden`으로 만든다.
      **손으로 고치지 않는다**
- [ ] 2.11 **RED (재생 경계)** — 재생 응답에는 이 code 분류가 적용되지 않는다.
      오늘 `classifyReplay`가 `classifyMutation`과 코드를 공유하지 않아 **우연히** 안전하나,
      재생 attestation이 켜지는 날 이 code는 반대 방향으로 작동한다.
      **구조로 고정하고 spec에 SHALL NOT으로 적는다**

## 3. R2 — 청소는 엔진 귀속 주문만 다룬다 (D2 → **4판 D−2.4 로 축소**)

> **4판의 변경(3라운드 F3·F5)**: 청소 목록을 브로커 미체결로 넓히지 **않는다.** 대상은 오늘의
> `Journal.LiveOrdersForSymbol`(엔진 귀속 주문)이다. 사람이 넣은 외부 주문의 취소는 **사용자 결정 대기**(3.X).
> 3판의 3.A(스냅샷 주입)·3.B1·3.B2·3.B4·3.B5·3.1·3.4·3.10·3.E1·3.E3 은 3.X 로 옮긴다.

- [ ] 3.0 **Pre-Edit 선언** — `internal/app/engine/exitloop.go` `clearTheSymbol`
- [ ] 3.2 **RED** — 취소 확정 후에만 보호 청산이 제출된다 (`clearTheSymbol` `:1485-1486`·`:1489-1490` 유지)
- [ ] 3.3 **RED** — 취소가 확정되지 않으면 **제출하지 않는다**
- [ ] 3.5 **RED** — 이 경로에서 나가는 mutation은 **취소뿐** — 신규·정정 0건
- [ ] 3.6 **RED** — 저널에 intent가 있는 주문의 lineage 처리 **무변화**
- [ ] 3.8 **RED** — 익절 경로에서도 같게 동작한다
- [ ] 3.9 **RED (4판)** — 엔진에 귀속되지 않는 미체결 주문(원장에 없는 주문)은 **취소 대상이 아니다** — 오늘 동작 고정
- [ ] 3.B3 **RED** — 원장 주문의 빈 `price` 는 **0으로** 읽고 치우기 실패로 판정하지 않는다(`floatOf` `:1779-1785`, a087 대비)
- [ ] 3.7 **RED (§0.3 회귀)** — 취소·해석 실패는 `clearTheSymbol` **내부에서 흡수**되어 `clear=false` 가 된다
      (`:1465`·`:1471`·`:1485-1486`). 원장 목록 읽기 실패는 오늘처럼 오류로 반환된다(`:1443-1445`) — 무변화 고정
- [ ] 3.E2 **RED (§0.4)** — 이 경로에서 나가는 **새 브로커 조회가 0건**임을 고정
- [ ] 3.E4 **RED (4판 D−2.7)** — 같은 포지션의 청소가 **연속 3회** `clear=false` 로 끝나면 `EventExitLiquidationDelayed`
      (이미 critical, `internal/obs/event.go:336`)를 **새 트리거로 한 번** 낸다. **기존 30초 타이머**(`noteDelay` `:1252`·
      `:1675-`, 한계 `:116`)의 시작·해제·한계·중복 방지는 **무변화**. 그 뒤에도 자동 제출하지 않는다(§6)
- [ ] 3.E5 **RED (4판 D−2.7)** — 그 주기에 치우지 못한 주문이 **모두** 엔진의 미종결 `CANCEL` attempt(같은 `target_order_id`,
      `PendingAttempts` 에 있음)의 대상이면 새 계수를 늘리지 않는다. 하나라도 다른 이유면 센다. **기존 타이머에는 이 제외가
      적용되지 않는다** — 둘을 따로 단언한다

### 3.D 자기 방향 부재 확인 — **철회** (2라운드 §2.2)

- [x] 3.D0 철회 결정을 `review.md` §2.2와 `design.md` D2에 기록
- [ ] 3.D1 **RED (회귀)** — `withPending=false`인 주기에 자기 방향 매도가 있어도 **보호 청산은 제출된다**(오늘 동작 유지)
- [ ] 3.D2 **RED** — `armExitProposalTx` `:666`이 미결 발의 위의 두 번째를 거부한다(초과 매도의 1차 방벽 — 4판 D−2.2 가
      그것을 해제가 치우면 안 되는 이유로 쓴다)

### 3.X 사용자 결정 대기 — 엔진 밖 주문의 취소 (4판 D−2.4)

확대는 사용자 몫이다. 결정이 나기 전에는 아래를 **하지 않는다.** 결정이 "넓힌다" 면 그 change 가 D−2.4 의 선행 조건
넷(귀속 규칙 · 다른 포지션/사람 보호 매도 제외 · 공표된 OPEN 스냅숏 계약 · 감사)을 **모두** 가져간다.

- [ ] 3.X1 사용자 결정 기록(넓힌다/넓히지 않는다) — 3판 3.A·3.B1·3.B2·3.B4·3.B5·3.1·3.4·3.10·3.E1·3.E3 의 처분이 이것을 따른다

## 4. R3 — 비수용으로 종결된 attempt 가 발의를 푼다 (D3 → **4판 D−2.2·D−2.5**)

> **4판의 변경(3라운드 F1·F4)**: park(`UNRESOLVED_IN_DOUBT`)에서는 **해제하지 않는다.** 해제는 입증된 비수용
> (`FAILED_CONFIRMED`·`NOT_DISPATCHED`)에서만, **기대 intent 대조**와 함께. 두 쓰기 사이의 충돌은 **기동 따라잡기**가 닫는다.

- [ ] 4.0 **Pre-Edit 선언** — `Journal.ResolveExitProposal`(호출 형태 변경) · 기동 복구 이음매(`engineRecoverySequence`)
- [ ] 4.0a **FLM(편집 전)** — `ResolveExitProposal` 은 이미 번들이 있다(refresh 됨). `engineRecoverySequence`·`recoverThenReady`
      (`cmd/tossctl/engine.go:604-606`·`:677`)의 AST·FLM·BTM 을 **편집 전에** 만든다
- [ ] 4.1 **RED (핵심)** — attempt가 `FAILED_CONFIRMED`로 종결하면 그 intent를 가리키는 발의가 해제되고, **쓸 수 있는 시세가
      있는** 다음 관측에서 손절이 다시 제안된다. `EvaluateLadder` **B26** `:441`이 더는 억제하지 않는 것을 확인
- [ ] 4.1a **RED** — **RATCHET에서도 같다.** `EvaluateRatchet` **B17** `:423`
- [ ] 4.2 **RED (안전)** — attempt가 `CONFIRMED`면 발의를 **해제하지 않는다**
- [ ] 4.3 **RED (안전, 4판)** — `NOT_DISPATCHED` 는 해제 대상이다. **`UNRESOLVED_IN_DOUBT` 는 해제하지 않는다** — park 된 attempt 의
      발의는 무장된 채 남고, 원 매도가 살아 있다고 가정한 반례(park + 살아 있는 SELL)에서 두 번째 매도가 무장되지 않음을 단언한다
- [ ] 4.3a **RED (4판 F4)** — 해제는 attempt 의 `intent_id` = 현재 `pending_intent_id` 일 때만 비운다. 다르면 무변화
      (늦게 도착한 해제 · 재무장 뒤의 옛 해제)
- [ ] 4.3b **RED (4판 F4)** — 기동 따라잡기: `pending_intent_id` 가 찬 발의마다 그 intent 의 **마지막** attempt 가
      `FAILED_CONFIRMED`·`NOT_DISPATCHED` 면 해제. 종결과 해제 사이 충돌 fixture 가 다음 기동에서 풀린다. 멱등
- [ ] 4.4 **RED (4판 F6 — 재핀)** — 기동 순서: ① 재분류(§4bis)가 `Recovery.Run` 의 해소(`internal/reconcile/recovery.go` —
      `Resolver.Resolve`)보다 **먼저**, ② 따라잡기(4.3b)가 `Recovery.Run` 뒤·`ready` 앞, ③ `Recovery.Run` 본문(재시작 규칙 →
      재생 → 해소)과 인터록 의미 **무변화**. 좌표 단언이 아니라 순서 단언이다(`engineRecoverySequence` `:604-606`,
      `recoverThenReady` `:677`)
- [ ] 4.4a **RED (핵심)** — 이 경로는 `Journal.RecoverPending`을 **세션 중에 부르지 않는다**(`journal/recovery.go:95-109`)
- [ ] 4.5 **RED (§6)** — 발의 해제가 **손절 가격을 바꾸지 않는다.** `entry_price`·`initial_stop`·`baseline_price` 쓰기 0건
- [ ] 4.5a **RED** — `pending_level`이 음수면 rung 되돌림 없이 해제된다(`RungIndex` 거부). **080220이 그 경우다**
- [ ] 4.6 **RED** — 해제는 **멱등**이다. `ResolveExitProposal` B8
- [ ] 4.7 **GREEN** — 후처리 하나 + 기동 따라잡기. **새 루프도 주기도 브로커 조회도 없다**

## 4bis. R1 소급 — 원 발주 응답으로 식별되는 IN_DOUBT 만 다시 읽는다 (D1 → **4판 D−2.3**)

> **4판의 변경(3라운드 F2)**: 재분류 대상은 `attempt_transitions` 로 **원 발주 응답임이 양성 식별되는** attempt 뿐이다.
> 스키마 추가 없음. 추가가 필요해지면 멈추고 보고한다.

- [ ] 4b.0 **Pre-Edit 선언** — 기동 경로
- [ ] 4b.1 **RED** — 기동 시 1회, D−2.3 의 여섯 조건(PLACE · `broker_order_id=''` · ACKED 이력 없음 · 모호 전이 하나가
      `DISPATCH_STARTED`→`IN_DOUBT`/`dispatch_outcome_unknown` · detail 이 상태 코드 분기 모양이고 두 상태가 같음 · code 가
      목록에 있음)을 모두 만족하는 attempt 가 `FAILED_CONFIRMED`로 재분류된다
- [ ] 4b.2 **RED (안전)** — **code로만 판단한다.** detail 의 `official: API error <n>: ` 뒤 본문만 JSON 으로 읽고 엔진 산문을
      매칭하지 않는다
- [ ] 4b.2a **RED (안전, 4판)** — `MarkAcked` 뒤 readback 실패로 모호가 된 attempt(`ACKED → IN_DOUBT`,
      `ack_round_trip_unconfirmed`, `broker_order_id` 있음)는 그 detail 에 확정 거절 code 가 있어도 **재분류하지 않는다**
- [ ] 4b.2b **RED (안전, 4판)** — 전송 실패 분기(`transport failed with the request …`) · 상태 없음 분기 · 출처 불명 기록은 대상이 아니다
- [ ] 4b.2c **RED (안전, 4판)** — 최상위 `code` 와 `error.code` 가 다르면 재분류하지 않는다
- [ ] 4b.2d **확인(FLM)** — 재생 경로(`internal/execgw/replay.go`)가 원장 전이를 어떻게 쓰는지 편집 전 FLM 으로 열거하고, 조건 3
      ("모호 전이 하나")이 재생 뒤 모양을 배제하는지 단언으로 고정
- [ ] 4b.3 **RED (안전)** — 확정 거절 code가 **없는** IN_DOUBT는 건드리지 않는다. `request-in-progress`를 포함
- [ ] 4b.4 **RED (안전)** — 재분류는 **attempt 상태만** 바꾼다. 발의 해제는 §4가 한다
- [ ] 4b.5 **RED** — **기동 시 1회.** 주기적으로 돌지 않는다
- [ ] 4b.6 **실측 재생** — 원장의 두 행(`034e5b79…`·`8f68e7c3…`)을 fixture로, 재분류 → 발의 해제 → 다음 주기 발의 → 청소 →
      제출까지 관통시킨다. **4판에서 기대 결말은 「같은 409 로 다시 거절되고 보고된다」 이다** — 반대 매수가 엔진 밖 주문이라
      청소가 치우지 않는다(D−2.4). 손절이 나가는 결말은 3.X 의 사용자 결정에 딸린다

## 5. R4 — **철회** (1라운드 차단 2·3)

1판의 5.1~5.6(보호 청산에 baseline 싣기)은 **전부 철회한다.** 값 원천이 없고
(`ExitObserverOptions` 22필드에 0개), 억지로 채우면 미체결 매도의 부재를 거짓 확증해
**살아 있는 매도를 은퇴시킨다**(`absenceCorroborated`는 매수 예약 모델이다).
오늘의 「항상 park」가 안전측이며 그것을 제거하는 것은 §6 위반이다.

- [x] 5.1 철회 결정을 `review.md` 1.7과 `proposal.md` R4 절에 기록
- [x] 5.2 spec에 SHALL NOT으로 고정 — 매도용 부재 증거 모델 없이 기준선 공급 금지,
      미측정 필드를 0으로 채우기 금지
- [ ] 5.3 `issues.md`에 **후속 조건** 기록: 부재 확증의 증거 모델이 매도에 대해
      아무것도 증명하지 못한다는 사실과, 매도용 모델(체결 이벤트 부재 + 목록 완주의 결합)이
      별도 change의 선행 조건이라는 것

## 6. 실측 재생

- [ ] 6.1 2026-08-06~07의 세 건(`6GKYatiUehps5SQX`·`7d3we7ZD3dtxWTMO`·`7k5oRgmEHnoU5Vfi`)을
      fixture로 재생 — attempt가 종결하는지, 발의가 풀리는지, **4판의 기대 결말(엔진 밖 반대 매수는 치우지 않음 →
      같은 409 로 다시 거절되고 보고됨)** 이 나오는지. 손절 제출 결말은 3.X 의 사용자 결정 뒤
- [ ] 6.2 **272210의 라이브락**(`STOP_LOSS_LADDER → PROPOSAL_CANCELLED` 1931건 · 약 2h54m ·
      inter-arrival 중앙값 5.0초)이
      재생에서 멈추는지
- [ ] 6.3 결과를 `issues.md`에 기록. **a087·a089·a091·a092와의 상호작용**을 명시한다

## 7. 게이트

- [ ] 7.0 **`openspec validate --strict`의 한계를 명시한다** — MODIFIED는 요구 블록을
      **통째로 치환**하므로(`specs-apply.js:207-236`), 본문·시나리오를 빠뜨린 delta도
      `--strict`를 통과한다. 1판이 실제로 그렇게 정본을 지웠다.
      **따라서 검사는 「기존 요구 본문과 시나리오가 delta 안에 그대로 있는지」를
      문자열 대조로 따로 한다.** validate 통과는 그 증거가 아니다
- [ ] 7.1 `go test ./... -count=1 -race` 회귀 0
- [ ] 7.2 **§0.3 확인** — 3판은 스냅샷 읽기이므로 **새 브로커 왕복이 0건**이다(3.E2).
      제출 시각 차를 3.E3으로 측정하고, **2판의 「상한 2초」 약속은 폐기됐음을 적는다**
- [ ] 7.3 **§0.4 확인** — **이 change가 더하는 브로커 호출은 0건**임을 FLM calls 표로 보인다.
      R2는 detector 스냅샷을 읽고, R3의 고리는 원장 쓰기이며, R1 소급은 원장 읽기다
- [ ] 7.4 **토글 OFF 동등성** — 이 change는 토글을 도입하지 않는다.
      도입하지 않았음을 명시한다(`not-applicable` 아님 — 해당 없음이 아니라 무도입)
- [ ] 7.5 FLM·AST **재생성** (구현 후) + `check_analysis.py` 통과
- [ ] 7.6 `make sdd-sync` → `make sdd-check`
- [ ] 7.7 **격리 worktree에서** `make gate CHANGE=a094-a-stop-clears-what-blocks-it`
- [ ] 7.8 **독립 리뷰**(구현과 분리된 컨텍스트). **교차 모델을 지킨다**
- [ ] 7.9 PM 동기화 → `openspec archive`

## 8. 배포와 운영 — 사람이 승인한다

- [ ] 8.1 배포 전 `main`과 **SchemaVersion 대조** (낮으면 엔진이 조용히 죽는다)
- [ ] 8.2 **엔진 재시작은 사람이 직접 승인한다.** 재시작 자체가 recovery를 돌려
      현재 얼어붙은 attempt를 park시키므로, 그 시점에 무엇이 일어나는지 미리 적어 둔다
- [ ] 8.3 배포 후 **첫 409 사건의 실물 확인** — attempt가 종결하는지, 청소가 도는지
- [ ] 8.4 이 change는 **현재 열린 세 포지션을 소급 보호하지 않는다.**
      배포 전까지 475150·080220·272210은 사람이 처리한다

## 선후 관계

```text
a094 (이 change) ── 409 동결과 충돌 해소
   │
   ├─ a087 보호 청산은 시장가      **3판에서 제약 해소.** 빈 가격을 저널분에도 0으로 읽는다
   ├─ a089 나가지 못한 손절을 센다  **규범 충돌(4판 정정)** — a089 R2 는 같은 code 로 동작 분기 금지. D−2.8
   ├─ a091 한 주도 못 판 손절      독립. 알림 등급
   └─ a092 알림이 손절을 잡지 않는다 독립. 알림 체류
```

**a094는 넷 중 어느 것에도 의존하지 않고, 넷 중 어느 것도 a094를 대신하지 않는다.**
a089가 세는 "나가지 못한 손절"에 이 사건이 포함되지만, a089는 기록만 하고 원인을
제거하지 않는다.

## 안전 불변식 확인

| 불변식 | 이 change에서 |
| --- | --- |
| §1 사람 승인 없는 LIVE 주문 side effect 금지 | 구현·테스트는 fixture. 배포·재시작은 8절에서 사람이 승인 |
| §2 `mutating: true` 자동 실행 금지 | 준수 |
| §3 토글 OFF는 upstream과 동일 | **토글을 도입하지 않는다** |
| §4 손절 즉시성을 약화·지연하지 않는다 | 3.7·3.E1·3.E2·3.E3·7.2. **3판은 새 동기 왕복이 0건**이다. 그리고 **§4의 더 큰 위반은 영구 억제였다** — D3의 고리가 그것을 푼다 |
| §5 High-risk 경로 | 주문·손절·대사 전부 해당. Pre-Edit 선언은 2.0·3.0·4.0·4b.0. **4bis는 저장된 원장 행의 상태를 바꾸므로 사람 승인 아래에서만 배포한다**(8절) |
| §6 보수 방향만 | R1~R3 전부 **손절이 더 잘 나가는** 방향. 사이징·손절가·레벨은 안 바꾼다. **뒤집지 않는 것 둘**: 「못 치우면 팔지 않는다」(B7, 3.E4)와 「미정산은 그 종목을 막는다」(D6) |
| §7 운영 토글 flip과 live 검증은 사람이 | 8절 |
| §8 시크릿·계좌 개인정보 저장 금지 | 원장 인용은 종목코드·수량·시각·requestId까지. 계좌번호·잔고 절대액 없음 |
