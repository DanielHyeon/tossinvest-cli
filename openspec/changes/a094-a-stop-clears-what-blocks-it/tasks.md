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
- [x] 0.5f **proposal-freeze 리뷰 4라운드**(codex 교차 모델) — **REJECT**(P0 1 · P1 5 · P2 2), `review.md` 「4라운드」 ·
      `analysis/freeze-review/codex-r4-output.md`(`f2decd0a`)
- [x] 0.5h **사건 두 행의 현재 상태 재측정**(문서 리뷰 P1, Manager 조건부 승인) — 운영 원장 읽기 전용(`mode=ro`, `query_only`),
      브로커 호출 0. 결과: 두 행 모두 **2026-08-08 에 이미 `UNRESOLVED_IN_DOUBT`**, 발의 무장 유지. 모호 전이는 재분류 조건 2~7 을 채우나
      조건 1 을 못 채운다 → 사건의 해동은 Q4-1 운영자 도구 경로. design D−2.2 에 기록
- [ ] 0.5g **4판 편집 전 산출물 재작성**(F10) — `record` 번들의 FLM 「Branches」 논증과 BTM 미진입 요약을 현재 AST(16분기)
      기준으로 다시 쓰고, 4판이 편집하는 기존 함수(`ResolveExitProposal` 호출 형태 · 기동 이음매)의 FLM 을 편집 전에 갖춘다.
      **5판 현황**: `record` FLM 「Branches」 표를 16분기로 재번호(BTM 은 이미 재번호) — 남은 것은 편집 전 FLM 셋(4.0a · 4.0b · N1 의 3.0a)
- [x] 0.5i **5판 반영** — `design.md` D−3(N1 Q4-4 번복 · N4 ACKED 기동 정산 · 재분류 이연 · N3 3상 분류기 · N7 정정 · N2 잔여 · N8 · a089
      전제), spec delta 2(재분류 요구·시나리오 삭제, 3상·ACKED 정산·park 위 발의 보존 추가), tasks §2·§3·§4·§4bis·§6 개정. **판정 아님**
- [x] 0.5j **proposal-freeze 리뷰 5라운드**(codex) — **REJECT**(P0 2 · P1 5), `review.md` 「5라운드」 · `codex-r5-output.md`.
- [x] 0.5k **6판 반영** — `design.md` D−4(R5-1 기동 정산 철회·알림만 · R5-7 취소 ACK≠치움(형태 B) · R5-2 park 알림 판정 경로로 ·
      R5-3 해동 명령 1급 요구 · R5-4 enqueue-only · R5-5 셋째 기전 확인·제안 · R5-6 §0.4 재계수 0), spec delta 2, `record` FLM 정정. **판정 아님**
- [x] 0.5l **proposal-freeze 리뷰 6라운드**(codex) — **REJECT**(P0 0 · P1 5 · P2 1 · P3 1), `review.md` 「6라운드」 · `codex-r6-output.md`
- [x] 0.5m **7판 반영** — `design.md` D−5(R6-1 공유 확정 판정 강화 · R6-2 에피소드 key · R6-3 적재 실패 진입 잠금 · R6-4 무기한 대기 정직 명시 ·
      R6-5 인증 요청 포함 계수 · R6-6 잔여 · R6-7 정리), spec delta 2. **판정 아님**
- [x] 0.5o **8판 반영(a092 22라운드 교차 3건)** — D−6.1 단일 입구 사용(D−5.3 직접 잠금 철회) · D−6.2 세대 읽기 시점(오류 반환 직후) · D−6.3 에피소드⊥재알림 창
- [x] 0.5n **proposal-freeze 리뷰 7라운드**(codex) — **REJECT**(P0 0 · P1 4 · P2 3 · P3 1), `review.md` 「7라운드」 · `codex-r7-output.md`
- [x] 0.5p **9판 반영**(Manager 2026-09-29) — D−7(R7-1 연속 id · R7-3 종결 증거 해소 명령 · R7-4 a092 재무장 대상 정합 · R7-7 근거 정정 · R7-5/6/8). **판정 아님**
- [x] 0.5q **proposal-freeze 리뷰 8라운드(좁은 확인)**(codex) — **REJECT**(P0 0 · P1 4 · P2 1), `review.md` 「8라운드」 · `codex-r8-output.md`
- [x] 0.5r **10판 반영**(Manager 2026-09-29, (A)) — D−8(운영자 종결 단언 철회 · Q9-1 무효화 · §4 논증 · 감지 복구 절차·경보 본문 · R8-4/R8-5). **판정 아님**
- [ ] 0.5s **proposal-freeze 리뷰 9라운드(좁은 확인 — 철회 절 + §4 논증 한정)**(codex) — 대기열 — **a089 처분(사용자 답) 뒤에만**(D−3.9). 그 전에는 freeze 하지 않는다.
      전제 충족: a089 아카이브(`64a1b2b3`, 2026-09-28) — codex 슬롯 대기열(a092 r21 → a095 r3 → a094 r5)

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
- [ ] 2.5e **RED (5판 N3 — 모호 강제)** — 최상위 `code` 와 `error.code` 가 둘 다 있고 값이 다르면 `DispatchAmbiguous` 를 **즉시** 반환하고
      뒤의 분류(`ClassifyBrokerRefusal` · 상태 코드)를 타지 않는다(D−3.5)
- [ ] 2.5f **RED (5판 N3 — 계약 반례)** — **422** + 두 자리가 다른 code → 확정 거절이 **아니다**(모호). 4판의 「분류하지 않음 → 종전 경로」 는
      이 입력을 상태 코드 분기의 422 확정 거절(`journal/dispatch.go:323-327`·`:351`)로 떨어뜨렸다 — 그 반례를 계약 시험으로 고정한다
- [ ] 2.5g **RED (5판 N3)** — 세 결과를 표로 고정한다: 확정 거절(목록 안 · 모순 없음) · 모호 강제(두 자리 모순) · 판정 없음(JSON 아님 · code 없음 ·
      목록 밖 → 종전 경로). 판정 없음 입력이 409 면 종전대로 모호, 422 면 종전대로 확정 거절임을 함께 단언한다
- [ ] 2.6 **RED** — `isDefinitiveRejection`의 상태 목록 **무변화**를 표로 고정
- [ ] 2.7 **RED** — 기존 세 code(`trade_auth_required`·`fx_consent`·`funding_required`)의
      분류 **무변화**
- [ ] 2.8 **GREEN** — `ReasonOppositePendingOrder` 추가 + `AllReasonCodes()` 등록 +
      **`classifyRefusalCode` 신설 — 반환은 3상**(D−3.5; `(ReasonCode, bool)` 은 5판에서 폐기).
      `code`와 `error.code`만 읽고, `classifyRefusalBody`보다 **먼저** 부른다.
      **기존 세 항목의 substring 매칭은 건드리지 않는다**(D0)
- [ ] 2.9 `submit`이 `default:` 갈래(`exitloop.go:1410-1416`, base 번호 B10 `:1304`)로 가는 것을 확인 — `alertProposalRefused` + 레벨 재무장
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
- [ ] 3.0a **FLM(편집 전, 5판 N1)** — `clearTheSymbol` 번들을 현재 소스로 재생성하고 해제 자리(`:1492-1495`)의 분기를 AST 로 열거한 뒤 편집한다.
      발의 intent 의 attempt 상태를 읽는 원장 질의가 없으면 새 함수로 둔다(새 파일)
- [ ] 3.N1 **RED (5판 N1 핵심)** — 무장된 익절 발의의 attempt 가 `UNRESOLVED_IN_DOUBT` 인 종목에서 손절 조건이 서면 청소는 발의를
      **비우지 않고**(`ProposalCancelled` 0건) `clear=false` 이며, 보호 청산 제출 0건이다
- [ ] 3.N1a **RED (5판 N1)** — 같은 상황에서 critical 이 **한 번** 나가고 그 이벤트가 park 된 attempt id·종목·운영자 해소 필요를 싣는다.
      key 는 포지션 단위다(두 포지션이면 두 key)
- [ ] 3.N1b **RED (5판 N1)** — 발의 attempt 가 `RECORDED`·`DISPATCH_STARTED`·`ACKED`·`IN_DOUBT` 여도 해제하지 않는다(표). 그 경우 park 원인
      critical 은 나가지 않고 기존 타이머·D−2.7 트리거가 그대로다
- [ ] 3.N1c **RED (5판 N1, 무변화)** — 발의 attempt 가 모두 `CONFIRMED` 면 오늘대로(청소 목록 → 취소 확정 → 해제). 모두 비수용 종결이면 4.7 의
      판정 함수로 해제
- [ ] 3.N1d **RED (5판 N1 해동)** — park 된 attempt 가 운영자 해소(`OperatorResolve` → 판정 함수, 4.3e) 또는 종결 증거로 비수용 종결되면 다음
      관측에서 발의가 풀리고 손절이 제안된다. 그 전에는 몇 주기가 지나도 풀리지 않는다
- [ ] 3.R2 **RED (6판 R5-2 — 도달성)** — 무장 발의가 **손절 자신**(`STOP_LOSS_LADDER`·RATCHET 의 `BASELINE_BREACH`)이고 그 attempt 가 park 인
      포지션에서 손절 조건이 다시 성립: 평가는 `SuppressedPending`(`ladder.go:441-443`·`ratchet.go:422-426`) 그대로, park 원인 critical 이
      포지션 key 로 **1회** 적재된다(사건 두 행의 모양). 무장 발의가 익절인 경우도 같은 critical 이 청소 자격과 무관하게 난다
- [ ] 3.R2a **RED (6판)** — park 원인 판정은 평가·억제·청소의 결과를 바꾸지 않는다(판정 전후 스냅숏·원장 행 동일)
- [ ] 3.R7 **RED (6판 R5-7 핵심)** — 무장 익절 매도를 청소가 취소해 취소가 CONFIRMED 인데 그 주문의 종결 스냅숏이 없는 주기: 발의 해제 0,
      보호 청산 제출 0, **같은 주문 재취소 0**. 종결 스냅숏을 기록한 다음 주기: 발의 해제 → 손절 제출
- [ ] 3.R7a **RED (6판 R5-7)** — 발의 intent 의 CONFIRMED 주문이 소유 모호로 `LiveOrdersForSymbol` 에서 빠지고 종결 증거가 없으면 발의 해제 0
      (그 intent 의 주문 번호로 종결을 본다 — `fills.go:1866-1872` 필터에 기대지 않는다)
- [ ] 3.R7b **RED (6판 R5-7, 무변화)** — **매수**(진입) 주문 취소는 종전대로 CONFIRMED 로 치운다 — 진입 매수 + 손절 경로의 제출 시점 무변화
- [ ] 3.R7c **RED (6판)** — 종결 증거를 기다리는 취소는 D−2.7 계수에서 빠지지 않는다(3.E5 의 제외는 기록·전송·인수 단계만)
- [ ] 3.R5 **RED (6판 R5-5 — 셋째 기전)** — 무장 발의 없음 + 같은 종목에 **다른 intent** 의 IN_DOUBT attempt: 청소 `clear=false`, 무장 0,
      `ProposalCancelled` 0, `clearDelay` 호출 0 → 기존 30초 지연 경보가 한계에서 1회. 여러 주기 반복해도 `PROPOSAL_CANCELLED` 0
- [ ] 3.R5a **RED (6판, 구조 — Q6-2 확정)** — 청소의 미종결 판정과 `checkSymbolFree` 가 **같은 게이트웨이 메서드**를 부른다(AST 구조 단언 + 그 메서드의
      대상 판정을 바꾼 변이가 두 경로 시험을 모두 깨뜨린다)
- [ ] 3.R4 **RED (6판 R5-4 · 9판 R7-8 — enqueue-only)** — 이 change 의 새 critical(3.E4 · 3.R2 · 4.3d · 4.N4)은 **a092 단일 입구(`RecordAlert`, 창 0)** 로
      기록되고 직접 `Journal.EnqueueAlert` 호출은 0(a092 census 핀), 관측 루프의 `ExitAlerter.Notify` 호출도 0. 전송자를 막아 둔 fixture 에서 다른 포지션의 손절 제출 시점 무변화
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
- [ ] 3.E4 **RED (4판 D−2.7)** — 같은 포지션의 청소가 **연속 N회**(`N = obs.DefaultCriticalAttempts` — a124 관례, 시험은 상수를 인용) `clear=false` 로 끝나면
      `EventExitLiquidationDelayed`(이미 critical, `internal/obs/event.go:336`)를 **새 트리거로, 타이머와 다른 알림 key 로 한 번** 낸다.
      `delayAlerted` 를 건드리지 않는다. **기존 30초 타이머**(`noteDelay` `:1252`·`:1675-`, 한계 `:116`, key `type|positionID` `:1688`)의
      시작·해제·한계·중복 방지는 **무변화** — 두 경보가 모두 나가는 것을 단언한다. 그 뒤에도 자동 제출하지 않는다(§6)
- [ ] 3.E5 **RED (4판 D−2.7)** — 그 주기에 치우지 못한 주문이 **모두** 엔진의 `CANCEL` attempt(같은 `target_order_id`)가
      `RECORDED`·`DISPATCH_STARTED`·`ACKED` 인 주문이면 새 계수를 늘리지 않는다. **`IN_DOUBT` 취소는 센다.** 하나라도 다른 이유면 센다.
      **기존 타이머에는 이 제외가 적용되지 않는다** — 둘을 따로 단언한다

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

- [ ] 4.0 **Pre-Edit 선언** — `Journal.ResolveExitProposal`(호출 형태 변경) · 기동 복구 이음매(`engineRecoverySequence`) ·
      기동 ACKED 확정·알림 자리(D−4.2 — 바이트 일치 확정만 상태 변경) · `Gateway.confirmCreatedOrder` 의 종목 비교(D−5.1 강화) · 운영자 해동 명령(새 파일) · `Gateway.checkSymbolFree`(판정 함수 추출, D−4.7)
- [ ] 4.0a **FLM(편집 전)** — `ResolveExitProposal` 은 이미 번들이 있다(refresh 됨). `engineRecoverySequence`·`recoverThenReady`
      (`cmd/tossctl/engine.go:604-606`·`:677`)의 AST·FLM·BTM 을 **편집 전에** 만든다
- [ ] 4.0b **FLM(편집 전, 6판)** — `Recovery.Run`(ACKED 알림을 그 안에 둘 경우)과 `Gateway.checkSymbolFree`(판정 추출) 번들을 현재 소스로
      재생성한다. `ExitObserver.judge`/`record`(park 알림 자리) 번들도 편집 전에 갖춘다
- [ ] 4.N4 **RED (6판 D−4.2, Q6-1)** — ACKED PLACE(기록 번호 있음)를 남긴 채 재시작: 가짜 주문 읽기가 **같은 번호 바이트 일치 + 같은 종목**을
      돌려주면 `CONFIRMED`, 알림 0, 읽기 호출 **정확히 1**
- [ ] 4.N4a **RED (6판, 안전)** — 같은 fixture 에서 읽기가 ① 실패 ② 다른 번호(대소문자·공백만 다른 번호 포함) ③ 다른 종목 ④ 해석 불가 → attempt ACKED
      그대로, 발의 무장 그대로, 명명 critical 1회(attempt key), `Recovery.Run` 은 `ErrRecoveryIncomplete` 없이 끝남. 해소기(`Resolver.Resolve`)·목록
      조회 호출 0
- [ ] 4.N4b **RED (6판)** — CANCEL/AMEND ACKED 와 기록 번호 없는 PLACE ACKED: 읽기 호출 0, 상태 무변경, 명명 critical 1회
- [ ] 4.N4c **RED (6판, 구조)** — 기동 확정의 번호·종목 판정과 `confirmCreatedOrder` 가 **같은 함수**를 부른다(AST 구조 단언 + 그 함수의 번호 비교를
      바꾼 변이가 발주 직후 확인 시험과 4.N4 를 모두 깨뜨린다)
- [ ] 4.N4d **RED (6판)** — `ResolveConfirmed` 원장 쓰기 실패 fixture: ACKED 그대로, 명명 critical, 복구 성공(루프 시작)
- [ ] 4.N4e **RED (7판 D−5.1 — 종목 공백)** — 발주 직후 확인과 기동 확정 **둘 다**: 번호 바이트 일치 + 응답 종목이 ① 필드 없음 ② `null` ③ `""` ④ 공백만 →
      확인 실패(발주 직후는 IN_DOUBT `ack_round_trip_unconfirmed`, 기동은 ACKED 잔존 + 알림). 한 함수의 종목 비교를 되돌린 변이가 두 경로 시험을 모두 깬다.
      **확인 읽기를 실제로 타는 픽스처 중 종목 없는 것을 먼저 센다**(후보 `wts_isolation_test.go:96`) — 그 목록만 고친다
- [ ] 4.N4f **RED (7판 D−5.2 · 9판 D−7.1 — 에피소드 key)** — ① 같은 attempt 의 기동 ACKED 알림을 전달→재시작→재적재: 새 행 0, 재전송 0 ② 다른 attempt: 새 행
      ③ park 원인: 같은 attempt 재시작 → 같은 행, 다른 attempt park → 새 행 ④ 청소 연속 실패: key 는 **연속 id**(`opts.NewID`) — 연속 해제 뒤 새 연속은 새 행,
      **같은 벽시계 시각에 시작한 두 연속도 다른 행**, 재시작 뒤 이어지는 연속은 새 에피소드(한 번 더 알림 — 명시된 보수 중복)
- [ ] 4.N4g **RED (8판 D−6.1 — 적재 실패)** — 새 기록자들이 **알림기 단일 입구**로 기록한다(직접 `EnqueueAlert` 호출 0 — a092 census 핀에 걸리지 않음).
      입구의 기록 실패 fixture(확정 전이 실패와 겹친 경우 포함): 진입 잠금은 입구의 생산자 래치가 한다, **관측 루프 계속·청산 무영향**.
      해제 세대 읽기는 입구의 몫(기록 오류 반환 직후) — a094 경로에서 적재 중 해제가 끼어도 잠기는지 단언(D−6.2)
- [ ] 4.N4h **구현 하드 조건(8판 개정)** — a092 `RecordAlert` 입구가 main 에 착지한 뒤에만 4.N4g 와 새 기록자 구현을 시작한다(설계 freeze 는 독립)
- [ ] 3.R7d **RED (7판 D−5.4 — 무기한 대기 · 10판: 사람 경로는 체결 감지 수리, D−8.2)** — 체결 감지가 종결 스냅숏을 **계속** 기록하지 않는 fixture(30초 · D−2.7 한계를 넘어): 발의 해제 0,
      제출 0, 지연 경보 1회, D−2.7 경보 1회 — 시간 경과로 풀리는 경로 0
- [ ] 3.R7e **RED (10판 D−8.3 — 경보 본문)** — 형태 B 로 보류 중인 포지션의 지연 경보와 D−2.7 경보 본문에 체결 감지 확인·복구 절차 안내 문장이 있다. 등급·key·빈도는
      무변화(형태 B 가 아닌 지연의 경보 본문은 무변화)
- [ ] 3.R7f **(10판 D−8.3 — 문서)** — `docs/operations.md` 「체결 감지가 멈췄을 때」 절(확인 · 흔한 원인과 조치 · 수리 뒤 확인 · 하지 말 것 — 원장 수기 기록 금지, 엔진 정지 금지)
- [ ] 4.N4x **후속 기록** — 나머지 ACKED 정산의 선행 조건: matcher 주문 번호 판별자(`indoubt.go:638-650`) + 덮어쓰기 좌표 `indoubt.go:307` 의
      반례(오답 단일 일치·복수 일치). 이 change 에서 구현하지 않는다
- [ ] 4.T **RED (6판 D−4.5 — 해동 명령)** — `tossctl` 운영자 명령: annotation `mutating: true` · 운영자·승인 참조·note·목표 필수 ·
      audit 가 먼저(audit 실패 fixture → 원장 무변화) · park 가 아닌 attempt 는 거절 · 비수용 종결 뒤 같은 명령 안에서 발의 해제(4.3e 의
      판정 함수 그 자체) · 엔진 lock 을 잡지 않는다 · 콘솔 라우트 0
- [ ] 4.Ta **RED (6판)** — 해동 명령의 두 쓰기 사이 충돌 fixture → 다음 기동 따라잡기(4.3b)가 발의를 푼다
- ~~4.Tb~~ **10판에서 철회**(D−8.1 — 운영자 종결 단언 없음)
- ~~4.Tc · 4.Td~~ **10판에서 철회** — Q9-1(additive 단언 테이블) 승인 무효화(D−8.1 근거 1~4). 이 change 는 스키마를 바꾸지 않는다
- [ ] 4.1 **RED (핵심)** — attempt가 `FAILED_CONFIRMED`로 종결하면 그 intent를 가리키는 발의가 해제되고, **쓸 수 있는 시세가
      있는** 다음 관측에서 손절이 다시 제안된다. `EvaluateLadder` **B26** `:441`이 더는 억제하지 않는 것을 확인
- [ ] 4.1a **RED** — **RATCHET에서도 같다.** `EvaluateRatchet` **B17** `:423`
- [ ] 4.2 **RED (안전)** — attempt가 `CONFIRMED`면 발의를 **해제하지 않는다**
- [ ] 4.3 **RED (안전, 4판)** — `NOT_DISPATCHED` 는 해제 대상이다. **`UNRESOLVED_IN_DOUBT` 는 해제하지 않는다** — park 된 attempt 의
      발의는 무장된 채 남고, 원 매도가 살아 있다고 가정한 반례(park + 살아 있는 SELL)에서 두 번째 매도가 무장되지 않음을 단언한다
- [ ] 4.3a **RED (4판 F4)** — 해제는 attempt 의 `intent_id` = 현재 `pending_intent_id` 일 때만 비운다. 다르면 무변화
      (늦게 도착한 해제 · 재무장 뒤의 옛 해제)
- [ ] 4.3b **RED (4판 F4)** — 기동 따라잡기: `pending_intent_id` 가 찬 발의마다 그 intent 의 attempt 가 **모두**
      `FAILED_CONFIRMED`·`NOT_DISPATCHED` 이거나 **하나도 없으면** 해제. 종결과 해제 사이 충돌 · 무장 뒤 `Prepare` 전 충돌 fixture 가
      다음 기동에서 풀린다. attempt 하나라도 `CONFIRMED`·park·미종결이면 해제하지 않는다. 멱등
- [ ] 4.3c **RED (4판, 문서 리뷰 P0)** — 세션 중 `submit` 은 `out.State ∈ {NOT_DISPATCHED, FAILED_CONFIRMED}` 이거나
      `out.AttemptID == ""` 일 때만 해제한다. dispatch 뒤 `MarkAcked`(또는 `Settle`·`MarkInDoubt`) 쓰기를 실패시킨 fixture
      (`gateway.go:747` — `State==""`, `AttemptID!=""`)에서 발의가 **무장된 채** 남음을 단언한다
- [ ] 4.3d **RED (4판, 문서 리뷰 P1 · Q4-5)** — 기동 단계(따라잡기 — 5판에서 재분류는 이연)의 한 행 실패는 그 행을 바꾸지 않고 다음 행으로 가며 Recover 가
      nil 을 돌려준다(루프 시작 — `runtime.go:289-295`). 그 실패는 **포지션 단위 key 의 critical** 로 발송된다
- [ ] 4.3e **(Q4-1 → 6판 4.T 로 구현 범위 편입)** — 운영자 해소 도구가 `OperatorResolve` 뒤 **같은 해제 판정 함수**를 그 자리에서 부른다는 계약을
      4.7 의 판정 함수 시그니처에 반영한다(판정 함수 하나 — 세션 중 `submit` · 기동 따라잡기 · 운영자 도구가 공유). ~~도구 구현은 범위 밖~~ — **6판: 4.T 로 구현 범위**
- [ ] 4.3f **정지 조건(Q4-6)** — 구현이 「rate-limit 으로 park 된 attempt」 를 타입으로 가려야 하게 되면(원장 reason 수준에서 429 가
      `dispatch_outcome_unknown` 과 구별되지 않는다 — `classify.go:114-117`) 새 reason code 추가를 Manager 에게 올리고 멈춘다
- [ ] 4.4 **RED (4판 F6 — 재핀, 5판 개정)** — 기동 순서: ① ~~재분류~~(5판 이연, D−3.4), ② 따라잡기(4.3b)가 `Recovery.Run` 뒤·`ready` 앞,
      ③ `Recovery.Run` 본문(재시작 규칙 → **기동 ACKED 확정·알림(D−4.2 — 7판)** → 재생 → 해소)과 인터록 의미는 그 외 **무변화**. 좌표 단언이 아니라 순서 단언이다(`engineRecoverySequence` `:604-606`,
      `recoverThenReady` `:677`). ②의 자리는 **`engineRecoverySequence` 클로저 안 `r.Run` 뒤** — 루프는 Recover 반환 뒤에만 시작한다
- [ ] 4.4a **RED (핵심)** — 이 경로는 `Journal.RecoverPending`을 **세션 중에 부르지 않는다**(`journal/recovery.go:95-109`)
- [ ] 4.5 **RED (§6)** — 발의 해제가 **손절 가격을 바꾸지 않는다.** `entry_price`·`initial_stop`·`baseline_price` 쓰기 0건
- [ ] 4.5a **RED** — `pending_level`이 음수면 rung 되돌림 없이 해제된다(`RungIndex` 거부). **080220이 그 경우다**
- [ ] 4.6 **RED** — 해제는 **멱등**이다. `ResolveExitProposal` B8
- [ ] 4.7 **GREEN** — 후처리 하나 + 기동 따라잡기. **새 루프도 주기도 없고, 브로커 조회는 기동 ACKED 확정 읽기(D−4.2 · D−5.5)뿐이다**(9판 R7-8)

## 4bis. R1 소급 — **5판에서 이연(선택 후속, 이 change 범위 밖)** (D1 → 4판 D−2.3 → **5판 D−3.4**)

> **5판(D−3.4)**: 사건 두 행은 이미 park 이고 재분류 조건 1 을 못 채운다(0.5h) — 이득 없이 전략 PLACE 행 변이 위험만 남는다(YAGNI).
> 아래 항목은 **이 change 에서 구현하지 않는다.** 후속 change 가 다시 열 때의 기록이며, 그때 조건 4 는 D−3.6(N7 정정)대로 **구조로**
> 강화하고 감싼 오류 픽스처(앞·뒤 문맥 덧붙임 · 표식 둘)가 거절됨을 시험에 넣는다. 사건 경로는 Q4-1 운영자 도구로 일원화한다.
>
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
- [ ] 4b.2d **RED (4판, 문서 리뷰 P1)** — 한 번이라도 재생된 attempt(`replay_count > 0` 또는 `last_replay_at` 있음 — `RefundReplay`
      뒤에도 시각은 남는다)는 원 409 본문이 조건을 모두 채워도 **재분류하지 않는다**
- [ ] 4b.2e **RED (4판, 문서 리뷰 P1)** — code 는 `official: API error <n>: ` 표식이 **정확히 한 번** 있을 때 그 뒤 JSON 에서만 읽는다.
      표식 0개·2개 이상·JSON 아님 → 그대로 둔다. 엔진 산문(`HTTP <n> does not prove …`)을 매칭하지 않는다
- [ ] 4b.3 **RED (안전)** — 확정 거절 code가 **없는** IN_DOUBT는 건드리지 않는다. `request-in-progress`를 포함
- [ ] 4b.4 **RED (안전)** — 재분류는 **attempt 상태만** 바꾼다. 발의 해제는 §4가 한다
- [ ] 4b.5 **RED** — **기동 시 1회.** 주기적으로 돌지 않는다
- [ ] 4b.7 **RED (4판)** — 매수 PLACE attempt 도 같은 조건이면 재분류되고, `FAILED_CONFIRMED` 전이가 그 결정의 예약을 푼다
      (`durability.go:653-663`) — 부수 효과를 단언한다
- [ ] 4b.8 **GREEN** — 재분류용 reason code 하나를 더하고 `testdata/reason_codes.golden`·`AllReasonCodes` 를 생성기로 갱신한다
- [ ] 4b.6 ~~실측 재생~~ — **5판에서 §6.1a 로 옮겼다**(재분류 없이, park 된 두 행 그대로의 결말)

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

- [ ] 6.1 **(5판 N5 — 기대 결말 고정)** 2026-08-06~07의 세 건(`6GKYatiUehps5SQX`·`7d3we7ZD3dtxWTMO`·`7k5oRgmEHnoU5Vfi`)을
      fixture로 재생한다. 기대 결말은 **5판 의미론**이다: 새 409(원 발주 응답·목록 안 code)는 R1 으로 `FAILED_CONFIRMED` →
      판정 함수가 발의를 풀고 → 다음 관측에서 손절이 다시 발의되며 → 반대 매수가 엔진 밖 주문이라 청소가 치우지 않아(D−2.4)
      **같은 409 로 다시 거절되고 보고된다.** 손절이 나가는 결말은 이 change 의 기대가 아니다(3.X 사용자 결정)
- [ ] 6.1a **(5판 N5 — park 된 두 행)** 원장의 두 행(`034e5b79…`·`8f68e7c3…`, 0.5h 에서 `UNRESOLVED_IN_DOUBT`)을 fixture로 재생하면
      **발의는 무장된 채 남고**(재분류 이연 · park 해제 없음) 보호 청산 제출 0건이며, park 원인 critical(**3.R2** — 무장 발의가 손절이라
      청소가 아니라 판정 경로에서)이 나간다. 해동 명령(4.T)으로 비수용 종결을 넣은 뒤에야 6.1 과 같은 결말로 간다
- [ ] 6.2 **(5판 N5)** **272210의 라이브락**(`STOP_LOSS_LADDER → PROPOSAL_CANCELLED` 1931건 · 약 2h54m · inter-arrival 중앙값 5.0초)을
      재생한다. 4판 fixture 는 4판이 내지 못하는 결말(손절 제출)을 요구했다 — 5판의 기대는 **반복이 멈추는 것**이지 손절 제출이 아니다:
      발의 attempt 가 모호·park 면 청소가 발의를 비우지 않고(D−3.2) 원인 알림이 한 번 나가며, 비수용 종결이면 판정 함수가 한 번 풀고
      409 재거절 보고로 간다. **단언은 결말 불문 「같은 포지션의 `PROPOSAL_CANCELLED` 반복 기록 0」** 이다. 재생이 그 반복의 기전을
      이 두 갈래 밖에서 보이면 기대를 고치지 말고 멈추고 보고한다.
      **6판**: 셋째 갈래(다른 intent 의 미종결 attempt 위 반복, D−4.7)를 코드로 확인했다 — 3.R5 가 그것을 막으므로 단언은 세 갈래 모두에서
      성립해야 한다. 재생 fixture 가 272210 의 반복을 어느 갈래로 재현하는지 **기록**한다(인과 확정)
- [ ] 6.3 결과를 `issues.md`에 기록. **a087·a089·a091·a092와의 상호작용**을 명시한다

## 7. 게이트

- [ ] 7.0 **`openspec validate --strict`의 한계를 명시한다** — MODIFIED는 요구 블록을
      **통째로 치환**하므로(`specs-apply.js:207-236`), 본문·시나리오를 빠뜨린 delta도
      `--strict`를 통과한다. 1판이 실제로 그렇게 정본을 지웠다.
      **따라서 검사는 「기존 요구 본문과 시나리오가 delta 안에 그대로 있는지」를
      문자열 대조로 따로 한다.** validate 통과는 그 증거가 아니다
- [ ] 7.1 `go test ./... -count=1 -race` 회귀 0
- [ ] 7.2 **§0.3 확인** — 4판 R2 는 원장만 읽으므로(`LiveOrdersForSymbol`) **새 브로커 왕복이 0건**이다(3.E2). 5판 N1 의 attempt 상태 읽기도
      원장이다. **6판: 관측 루프·손절 경로의 새 브로커 호출 0**(D−4.8). 새 호출은 기동의 ACKED PLACE 읽기뿐이다(`ready` 앞). R5-7 의 대가는 **무장 익절 매도 위의 손절이 종결 증거가 올 때까지
      늦는 것**이며 **시간 상한이 없다**(7판 D−5.4 — 체결 감지가 멈추면 무기한, 그때는 지연·연속 실패 경보가 드러낸다). 매수 취소 경로는 무변화다.
      **2판의 「상한 2초」 약속은 폐기됐음을 적는다**. 3판 3.E1·3.E3(스냅샷 신선도·시각 차)은 3.X 로 옮겼다
- [ ] 7.3 **§0.4 확인(7판 개정)** — 이 change 가 더하는 HTTP 요청은 **기동의 ACKED PLACE 행마다 주문 GET ≤3 + 토큰 POST ≤2**, 프로세스당 첫 토큰
      ≤1(캐시가 비었을 때)뿐임을 D−5.5 표와 FLM calls 표로 보인다. 전부 `roundTripTimeout` 3초 안. 관측 루프·손절 경로 0
- [ ] 7.4 **토글 OFF 동등성** — 이 change는 토글을 도입하지 않는다.
      도입하지 않았음을 명시한다(`not-applicable` 아님 — 해당 없음이 아니라 무도입)
- [ ] 7.5 FLM·AST **재생성** (구현 후) + `check_analysis.py` 통과
- [ ] 7.6 `make sdd-sync` → `make sdd-check`
- [ ] 7.7 **격리 worktree에서** `make gate CHANGE=a094-a-stop-clears-what-blocks-it`
- [ ] 7.8 **독립 리뷰**(구현과 분리된 컨텍스트). **교차 모델을 지킨다**
- [ ] 7.9 PM 동기화 → `openspec archive`

## 8. 배포와 운영 — 사람이 승인한다

- [ ] 8.1 배포 전 `main`과 **SchemaVersion 대조** (낮으면 엔진이 조용히 죽는다)
- ~~8.1a~~ **10판에서 철회**(스키마 무변경). 롤백 원칙은 8.2 로 옮겼다(R8-4)
- [ ] 8.2 **엔진 재시작은 사람이 직접 승인한다.** **롤백 원칙(10판 R8-4)**: 단순 백업 복원은 백업 뒤 mutation 0 이 검증된 창에서만, 그 밖은 현 저널(DB/WAL/SHM)을
      보존하고 통제된 복구·대사 뒤 재개(`internal/journal/backup.go:22-35`). 재시작 자체가 recovery를 돌려
      현재 얼어붙은 attempt를 park시키므로, 그 시점에 무엇이 일어나는지 미리 적어 둔다. **6판**: 첫 재시작에서 ACKED PLACE 행마다 주문 읽기 1회
      (바이트 일치 + 종목 일치면 CONFIRMED), 나머지 ACKED 행마다 attempt 당 critical 1건 — 배포 전 운영 원장의 ACKED 행 수를 종류별로 읽기 전용으로 세어
      요청 수(행당 ≤5, D−5.5)·알림 수·확정 읽기가 더하는 시간 상한(행 수 × 3초)을 미리 적는다
- [ ] 8.3 배포 후 **첫 409 사건의 실물 확인** — attempt가 종결하는지, 청소가 도는지
- [ ] 8.4 이 change는 **현재 열린 세 포지션을 소급 보호하지 않는다.**
      배포 전까지 475150·080220·272210은 사람이 처리한다

## 선후 관계

```text
a094 (이 change) ── 409 동결과 충돌 해소
   │
   ├─ a087 보호 청산은 시장가      **3판에서 제약 해소.** 빈 가격을 저널분에도 0으로 읽는다
   ├─ a089 나가지 못한 손절을 센다  **규범 충돌(4판 정정)** — a089 R2 는 같은 code 로 동작 분기 금지. D−2.8 → a089 아카이브(64a1b2b3)로 해소
   ├─ a091 한 주도 못 판 손절      독립. 알림 등급
   ├─ a092 알림이 손절을 잡지 않는다 **구현 하드 의존(8판 개정)** — 새 critical 은 a092 의 단일 입구로 기록한다(D−6.1). 구현은 입구 착지 뒤에만. **설계 freeze 는 독립**
   ├─ a124 계속 실패하는 전달자      (아카이브) enqueue 된 critical 의 전달 실패 시 진입 차단을 맡는다. N 상수 공유는 관례 일관성(D−3.8)
   └─ (후속, 미개설) ACKED 정산      선행 조건: 해소기 matcher 의 주문 번호 판별자(`indoubt.go:638-650`, 덮어쓰기 `:307`) — D−4.2
```

**4판 정정: a094 는 a089 의 처분에 의존한다**(design D−2.8 — 두 delta 를 함께 조정하기 전에는 어느 것도 freeze 하지 않는다).
**5판: 그 처분은 아직 없다 — freeze 의 미충족 전제이며 5라운드는 사용자 답 뒤에 돈다**(D−3.9). → **a089 아카이브(`64a1b2b3`, 2026-09-28)로 해소** — 5라운드는 codex 슬롯 순번.
a087·a091·a092 에는 **구현** 의존이 없다(6판 — 전달 지연 잔여는 enqueue-only 로 이 change 안에서 닫았다). 넷 중 어느 것도 a094를 대신하지 않는다.
a089가 세는 "나가지 못한 손절"에 이 사건이 포함되지만, a089는 기록만 하고 원인을
제거하지 않는다.

## 안전 불변식 확인

| 불변식 | 이 change에서 |
| --- | --- |
| §1 사람 승인 없는 LIVE 주문 side effect 금지 | 구현·테스트는 fixture. 배포·재시작은 8절에서 사람이 승인 |
| §2 `mutating: true` 자동 실행 금지 | 준수 |
| §3 토글 OFF는 upstream과 동일 | **토글을 도입하지 않는다** |
| §4 손절 즉시성을 약화·지연하지 않는다 | 3.7·3.E2·3.E4·3.E5·7.2. **4판은 새 동기 왕복이 0건**이고 기존 30초 지연 경보를 바꾸지 않는다. **4판은 park 의 영구 억제를 풀지 않는다**(D−2.2 — 이중 매도 위험이 더 크다) — 그 비용은 Q4-1 과 운영자 해소에 남긴다. **5판 N1 은 손절 즉시성을 양보한다 — park 된 발의 위에서만, 원인을 명명한 critical 과 함께만**(D−3.2). 모호·전송 중 상태는 오늘도 `checkSymbolFree` 가 막으므로 새 양보가 아니다. ~~5판 N4 는 재시작마다 얼어 있는 ACKED 발의를 없앤다~~ — **6판: 번호가 바이트 일치하는 ACKED PLACE 는 기동에서 확정한다. 나머지는 없애지 못하고 **attempt 당 한 번** 알린다**(D−4.2 · D−5.2 에피소드 key, 정산은 후속). **R5-7 은 무장 익절 매도 위의 손절을 종결 증거가 올 때까지 늦춘다 — 시간 상한 없음, 시간 경과로 풀지 않음**(7판 D−5.4, 살아 있을 수 있는 매도 위의 두 번째 매도를 막는 대가; 경보로 드러난다). **되살린 30초 경보의 동기 전송은 잔여**(D−5.6, a092 소유). **R5-5 수리는 지연 경보를 되살린다**(D−4.7) |
| §5 High-risk 경로 | 주문·손절·대사 전부 해당. Pre-Edit 선언은 2.0·3.0·4.0. 4bis 는 5판에서 이연. 6판 기동 ACKED 는 **바이트 일치 확정만** 원장을 바꾼다 — 첫 재시작은 사람 승인(8.2). **해동 명령(4.T)은 사람이 실행하는 mutating 명령**이다 |
| §6 보수 방향만 | R1~R3 전부 **손절이 더 잘 나가는** 방향. 사이징·손절가·레벨은 안 바꾼다. **뒤집지 않는 것 둘**: 「못 치우면 팔지 않는다」(B7, 3.E4)와 「미정산은 그 종목을 막는다」(D6) |
| §7 운영 토글 flip과 live 검증은 사람이 | 8절 |
| §8 시크릿·계좌 개인정보 저장 금지 | 원장 인용은 종목코드·수량·시각·requestId까지. 계좌번호·잔고 절대액 없음 |
