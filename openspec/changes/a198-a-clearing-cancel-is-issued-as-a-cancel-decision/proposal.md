# a198 — 길 치우기의 취소가 실제로 나간다

> **상태: 1단계(계약·증거), High-risk(주문 취소·손절 경로).** Story `STORY-TOS-a198`(FEAT-TOS-009). 작성: Manager(2026-11-23).
> 발견: a197 1단계의 probe(`openspec/changes/a197-…/analysis/probe/`, 브랜치 `feat/a197-external-opposite-buy-cancel`). 분기 주장의 AST 는 그 조사 산출물 인용이며, 이 change 안 사본과 현재 HEAD 재확인은 tasks 1.3 이다 — 그 전까지 「조사 인용」이다.

## Why

보호 청산(손절)은 제출 전에 그 종목의 엔진 귀속 미체결 주문을 취소한다(정본 exit-policy 「보호 청산의 길 치우기는 엔진 귀속 주문만 다룬다」, Scenario 「엔진이 낸 반대 방향 주문이 손절을 막고 있다 → 치우기가 그 주문을 취소하고, 취소가 확정된 뒤 보호 청산이 제출된다」). **운영 배선에서 그 취소는 나갈 수 없다.**

| 사실 | 근거 |
|---|---|
| 길 치우기는 취소마다 `Issuer.IssueReduction` 으로 결정을 발행하고 그 결정을 `Submit.Cancel` 에 싣는다 | `ExitObserver.clearTheSymbol`(exitloop.go:1683-1716, 직접 읽음) |
| 운영에서 그 `Issuer` 는 `RiskGuardian` 이다 | `Context.ExitObserver` 의 `c.Guardian.(ReductionIssuer)`(exitwiring.go:326·339, 직접 읽음) |
| `RiskGuardian.IssueReduction` 은 방향을 매도로, 종류를 주문 생성(`KindPlace`)으로 고정한 결정을 기록한다 | riskguardian.go:587-633(직접 읽음 — `Kind: journal.KindPlace` :614, 매도가 아니면 거절 :592). 주문 생성 결정에는 멱등 키가 붙는다(`decision.go:373-375`, 조사 인용) |
| 그 결정으로 게이트웨이 취소를 부르면 원장이 거절한다 — 브로커 취소 0 건 | probe 실측(HEAD `064f59fd9`): `state=` 빈 값, `reason=invalid_mutation_request`, `journal: invalid request: CANCEL carries no idempotency key — only order creation does`. 매수·매도 주문 모두. 키가 요청으로 옮겨지는 자리 `gateway.go:1081`(조사 인용), 거절 자리 `internal/journal/durability.go:267-269`(직접 읽음) |
| 결정 행에는 종류가 따로 남지 않는다 — 종류는 키 유무로만 남고, 방향은 취소할 주문과 맞아야 한다. 매수 주문은 키가 없어도 매도 결정으로는 취소되지 않는다(매도 주문은 키만 없으면 방향 비교를 통과한다) | `journal.Decision` 에 종류 필드 없음, 키는 `DecisionRequest.build` 가 주문 생성에만 파생(AST: analysis/cited-ast/DecisionRequest_build.ast.json B22). 매수 주문 + 키 없는 매도 결정 → `checkPreimage` 방향 비교에서 거절(AST: analysis/cited-ast/checkPreimage.ast.json B5, 실측 analysis/probe/probe-ext-output.txt `guardian_intent_hash_mismatch`). `checkIdempotencyKey` 는 키가 **있을 때만** 막는 둘째 벽(AST: analysis/cited-ast/checkIdempotencyKey.ast.json B2) |
| 같은 취소를 수동 flatten 과 같은 모양의 결정(종류 `CANCEL`, 방향 = 주문 방향)으로 내면 접수된다 | probe 대조군: `state=CONFIRMED`, 브로커 취소 1. flatten 의 발행: `flatten.go:360`·`:659-667` → `execgw.Issuer.IssueReduction`(`issue.go:86-103`, 조사 인용) |
| 실패한 자리에는 기록이 없고, 뒤따르는 경보는 원인을 말하지 않는다 | `clearTheSymbol` 의 `if err != nil \|\| out.State != journal.StateConfirmed { res.cleared = false; continue }`(exitloop.go:1717-1720, 직접 읽음). 그 결과로 `record` 가 지연 타이머(`noteDelay` :1402 — 30 초 뒤 critical)와 연속 실패 계수(`noteClearFailure` :1403 — 3 회 연속 critical)를 돌리지만 문장에 거절 원인이 없다(조사 인용) |
| 운영에서 이 결함이 손절을 막았다는 증거는 없다 — 잠복 결함이다 | Manager 실측 2026-11-23(운영 장부, 종류·상태·건수·시각만): `mutation_attempts` 의 종류는 `PLACE` 뿐(CONFIRMED 111 · FAILED_CONFIRMED 12,948 · UNRESOLVED_IN_DOUBT 6) — 거절된 취소는 attempt 를 남기지 않는다. `alert_outbox` 의 `exit.liquidation_delayed` 는 1 건(2026-08-07, a094 이전)뿐이고 엔진 로그에 그 이벤트는 0 줄이다. 즉 a094 이후 엔진 귀속 미체결 위에서 손절이 발의된 일이 관측되지 않았다 |
| 기존 시험은 이 결합을 지나지 않는다 | 엔진 하네스는 진짜 `RiskGuardian` 을 쓰지만 제출자가 가짜다 — `fakeSubmitter.Cancel`(`exitloop_test.go:110-120`)은 결정을 검사하지 않고 `CONFIRMED` 를 돌려주고, `a094Submitter.Cancel`(`a094_exit_clear_test.go:100-107`)은 취소 attempt 를 원장에 직접 써 넣는다. 게이트웨이 시험은 결정을 `execgw.Issuer` 로 만든다(조사 인용) |

즉 **엔진이 낸 미체결 주문(예: 체결되지 않고 남은 익절 지정가, 앞선 손절 주문)이 있는 종목에서 손절이 발의되면 길 치우기가 끝나지 못해 손절이 무기한 보류되고, 뒤따르는 경보는 원인을 말하지 않는다.** 정본이 요구하는 동작을 코드가 한 번도 수행한 적이 없다 — 지금까지는 그 상황이 운영에서 관측되지 않았을 뿐이다.

## What Changes

1. **취소 결정의 발행**: 길 치우기가 엔진 귀속 주문을 취소할 때 싣는 결정을 취소 종류(`CANCEL`)·위험 감소 등급·그 주문의 방향과 수량으로 발행한다 — 수동 flatten 과 **같은 발행 함수**(`execgw.Issuer.IssueReduction`)를 관측기가 이미 가진 원장·시계·계좌로 불러 발행한다(design D2, Manager 결정 2026-11-23). 손절·익절 **제출**의 결정 발행(`IssueReduction`)은 바꾸지 않는다.
2. **실패를 기록한다**: 취소 결정 발행이나 취소 제출이 실패하면(오류·비확정) 그 사실을 운영 로그 싱크(`ExitOrderLog`)에 계좌 정보 없이 한 줄로 남긴다(종류 취소·결과 상태·사유 코드, 같은 실패 반복은 한 번 — design D4). 판정(치움 미완료 → 보호 청산 미제출)은 종전 그대로다.
3. **결합을 지나는 시험**: 실제 게이트웨이와 실제 원장(임시 경로), 가짜 브로커로 길 치우기 → 취소 → 손절 제출을 끝까지 잇는 시험을 둔다. 운영 조립(`Context.ExitObserver`)으로 만든 관측기가 그 시험 가운데 하나를 구동한다.
4. **(보강 2026-11-24, 구현 리뷰 r1 P0-1) 이미 끝난 주문의 취소 거절은 확정 거절**: 엔진 게이트웨이의 취소가 409 `already-filled`·`already-canceled`·`already-rejected` 로 거절되면 그 CANCEL attempt 를 IN_DOUBT 가 아니라 확정 거절(기존 사유 코드 `dispatch_rejected_by_broker`)로 종결한다(code 필드 판정, 엔진 취소 정책에만 — `already-modified`·CLI flatten·발주·정정은 무변, design D7). 정본 「취소의 일시적 거절 처리」 의 Scenario 「다른 사유의 409」 를 이 예외에 맞춰 MODIFIED 로 고친다. 없으면 그 종목의 손절이 재시작까지 막힌다(이 change 가 취소를 처음 내보내면서 생긴 회귀, 리뷰 탐침 실측).
5. **(보강) 접수 확정된 매수는 다시 취소하지 않는다**: 엔진의 취소가 이미 접수 확정된 매수는 치워진 것으로 본다(design D8).

## 이 change 가 하지 않는 것

- 엔진에 귀속되지 않는(사람의) 주문의 취소 — a197(보류 중, 이 change 가 선행).
- 세션 중 `CANCEL` 의 IN_DOUBT 해소(a193 design §7-3 후속)와 ACKED 로 남은 `CANCEL` 의 복구. 공시: 취소가 ACKED 또는 IN_DOUBT 로 남으면 그 종목의 손절은 보류되고, ACKED 는 연속 실패 계수에서도 빠진다(대상 전부가 그 진행 중 취소일 때 — design §6). 지연 경보(30 초)는 나간다. 지금 빌드에는 ACKED 로 남은 취소를 닫는 운영자 경로가 없다 — 후속 change 『CANCEL attempt 의 복구 계약(IN_DOUBT·ACKED)』의 대상이고, a197 재개의 선행이기도 하다.
- 전일 주문의 거래일 결속(design §6). 매수 취소 접수 뒤 재취소는 D8 이 없앤다(design §2-5).
- 길 치우기의 판정 규칙(매도 취소의 종결 증거 대기, 미종결 attempt 가드, 연속 실패 트리거)·재제출 시점·알림 key 의 변경.
- 운영에서 과거에 이 실패가 실제로 손절을 막았는지의 소급 조사(장부에 흔적이 없다 — 위 표).

## Capabilities

### Modified Capabilities

- `exit-policy`: 길 치우기의 취소가 주문 경로가 받아들이는 취소 결정으로 발행되어야 하고, 그 결합이 실제 주문 경로를 지나는 시험으로 고정되어야 하며, 발행·제출 실패가 조용히 넘어가서는 안 된다는 요구 추가. 엔진의 취소가 접수 확정된 매수는 다시 취소하지 않는다는 요구 추가(D8).
- `order-execution`: 엔진 취소가 이미 끝난 주문을 이유로 거절되면(409 `already-filled`·`already-canceled`·`already-rejected`) 확정 거절로 분류한다는 요구 추가(D7), 「취소의 일시적 거절 처리」 의 Scenario 「다른 사유의 409」 를 그 예외에 맞춰 수정(MODIFIED — analysis/modified-delta/ 의 생성기·대조 diff).

## Impact

- 기존 함수 본문 편집(gate ⑤ FLM 대상, analysis/edit-surface.md): `ExitObserver.clearTheSymbol` 하나(취소 결정 발행 호출과 실패 기록 호출). `Context.ExitObserver`·`RiskGuardian.IssueReduction`·a129 함수·시험 하네스는 무편집. a166 census·a163 렌더 해시 적중 0.
- 새 코드: 실패 기록 함수(새 파일, `ExitOrderLog` 에 직접 한 줄 — design D4)와 그 중복 억제 맵(관측기 구조체 새 필드), 시험 파일.
- 보강(D7·D8, analysis/d7-d8-evidence.md §7): 기존 함수 편집 +1 `Gateway.withCancelRefusal`(internal/execgw — CANCEL 만 가를 수 있는 유일한 기존 자리) + 새 판정 함수(새 파일). `clearTheSymbol` 은 D8·D4a 로 추가 편집. 사유 코드는 기존 `dispatch_rejected_by_broker` 재사용 — `AllReasonCodes`·골든 무편집(Manager 결정 2026-11-24). `Gateway.submit`·`classifyMutation`·journal·flatten 패키지 무편집. D7 은 엔진 취소 정책에만 걸려 CLI flatten 의 동작은 바뀌지 않는다(Manager 결정 2026-11-24).
- 브로커 호출: 길 치우기가 원래 내려던 취소가 이제 실제로 나간다 — 취소 1건 = 사전 단건 조회 1 + 취소 요청 1, 둘 다 관측 루프 안에서 동기(retry matrix 행 31·33·37, analysis/rate-limit.md). 대상 주문 수 × (사전 조회 시한 2s + 취소 HTTP 시한 15s) 는 HTTP 시간 예산이며, 토큰 잠금 대기·401 재전송·계좌 예산 admission 을 더한 전체 최악 지연은 미확정 — design §2-4. 새 호출 종류가 아니다.
- 원장: 취소 attempt(`kind=CANCEL`)가 처음으로 기록되기 시작한다. 스키마 변경 없음 — 수동 flatten 이 이미 쓰는 형식.
- 설정·토글 없음. 정본이 이미 요구하는 동작의 수리이므로 토글 뒤에 두지 않는다.
- 롤백 = 이전 이미지(취소가 다시 거절되는 종전 상태). 이 change 가 쓴 취소 attempt 는 옛 이미지가 읽는 기존 형식이다.
- 진행 중 change 와의 겹침: a087(같은 파일 exitloop.go·exit_order_log.go, `landed-commit.txt` 없음 — a198 은 기다리지 않음, design §6)·a167(등록 단계, `record`·`submit` 예정, 함수 겹침 0).
