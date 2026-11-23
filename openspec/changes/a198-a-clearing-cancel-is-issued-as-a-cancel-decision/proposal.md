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
| 키만 없애서는 통과하지 않는다 — 결정의 방향·종류가 취소할 주문과 맞아야 한다 | 매수 주문의 취소는 `checkPreimage` 의 방향 비교(`guardian.go:333`), 모든 취소는 `checkIdempotencyKey`(`:398`)가 뒤에서 또 막는다(조사 인용) |
| 같은 취소를 수동 flatten 과 같은 모양의 결정(종류 `CANCEL`, 방향 = 주문 방향)으로 내면 접수된다 | probe 대조군: `state=CONFIRMED`, 브로커 취소 1. flatten 의 발행: `flatten.go:360`·`:659-667` → `execgw.Issuer.IssueReduction`(`issue.go:86-103`, 조사 인용) |
| 실패한 자리에는 기록이 없고, 뒤따르는 경보는 원인을 말하지 않는다 | `clearTheSymbol` 의 `if err != nil \|\| out.State != journal.StateConfirmed { res.cleared = false; continue }`(exitloop.go:1717-1720, 직접 읽음). 그 결과로 `record` 가 지연 타이머(`noteDelay` :1402 — 30 초 뒤 critical)와 연속 실패 계수(`noteClearFailure` :1403 — 3 회 연속 critical)를 돌리지만 문장에 거절 원인이 없다(조사 인용) |
| 운영에서 이 결함이 손절을 막았다는 증거는 없다 — 잠복 결함이다 | Manager 실측 2026-11-23(운영 장부, 종류·상태·건수·시각만): `mutation_attempts` 의 종류는 `PLACE` 뿐(CONFIRMED 111 · FAILED_CONFIRMED 12,948 · UNRESOLVED_IN_DOUBT 6) — 거절된 취소는 attempt 를 남기지 않는다. `alert_outbox` 의 `exit.liquidation_delayed` 는 1 건(2026-08-07, a094 이전)뿐이고 엔진 로그에 그 이벤트는 0 줄이다. 즉 a094 이후 엔진 귀속 미체결 위에서 손절이 발의된 일이 관측되지 않았다 |
| 기존 시험은 이 결합을 지나지 않는다 | 엔진 하네스는 진짜 `RiskGuardian` 을 쓰지만 제출자가 가짜다 — `fakeSubmitter.Cancel`(`exitloop_test.go:110-120`)은 결정을 검사하지 않고 `CONFIRMED` 를 돌려주고, `a094Submitter.Cancel`(`a094_exit_clear_test.go:100-107`)은 취소 attempt 를 원장에 직접 써 넣는다. 게이트웨이 시험은 결정을 `execgw.Issuer` 로 만든다(조사 인용) |

즉 **엔진이 낸 미체결 주문(예: 체결되지 않고 남은 익절 지정가, 앞선 손절 주문)이 있는 종목에서 손절이 발의되면 길 치우기가 끝나지 못해 손절이 무기한 보류되고, 뒤따르는 경보는 원인을 말하지 않는다.** 정본이 요구하는 동작을 코드가 한 번도 수행한 적이 없다 — 지금까지는 그 상황이 운영에서 관측되지 않았을 뿐이다.

## What Changes

1. **취소 결정의 발행**: 길 치우기가 엔진 귀속 주문을 취소할 때 싣는 결정을 취소 종류(`CANCEL`)·위험 감소 등급·그 주문의 방향과 수량으로 발행한다 — 수동 flatten 이 쓰는 것과 같은 모양. 손절·익절 **제출**의 결정 발행(`IssueReduction`)은 바꾸지 않는다.
2. **실패를 기록한다**: 취소 결정 발행이나 취소 제출이 실패하면(오류·비확정) 그 사실을 포지션 단위 에피소드로 남긴다. 판정(치움 미완료 → 보호 청산 미제출)은 종전 그대로다.
3. **결합을 지나는 시험**: 실제 게이트웨이와 실제 원장(임시 경로), 가짜 브로커로 길 치우기 → 취소 → 손절 제출을 끝까지 잇는 시험을 둔다. 운영과 같은 조립(`Context.ExitObserver`)이 취소 결정 발행기를 실제로 공급하는지도 시험한다.

## 이 change 가 하지 않는 것

- 엔진에 귀속되지 않는(사람의) 주문의 취소 — a197(보류 중, 이 change 가 선행).
- 세션 중 `CANCEL` 의 IN_DOUBT 해소(a193 design §7-3 후속).
- 길 치우기의 판정 규칙(매도 취소의 종결 증거 대기, 미종결 attempt 가드, 연속 실패 트리거)·재제출 시점·알림 key 의 변경.
- 운영에서 과거에 이 실패가 실제로 손절을 막았는지의 소급 조사(장부에 흔적이 없다 — 위 표).

## Capabilities

### Modified Capabilities

- `exit-policy`: 길 치우기의 취소가 주문 경로가 받아들이는 취소 결정으로 발행되어야 하고, 그 결합이 실제 주문 경로를 지나는 시험으로 고정되어야 하며, 발행·제출 실패가 조용히 넘어가서는 안 된다는 요구 추가.

## Impact

- 기존 함수 본문 편집(gate ⑤ FLM 대상): `ExitObserver.clearTheSymbol`(취소 결정 발행 호출과 실패 기록), `Context.ExitObserver`(발행기 공급 — 조사상 a166 census 밖). 정확한 목록과 census·구조 고정 적중은 tasks 1.4 에서 센다. `RiskGuardian.IssueReduction` 본문은 고치지 않는다(새 발행은 새 함수).
- 새 코드: 취소 결정 발행 함수(자리는 design D2), 시험 파일.
- 브로커 호출: 길 치우기가 원래 내려던 취소가 이제 실제로 나간다 — 엔진 귀속 미체결 주문당 취소 1. 새 호출 종류가 아니라 정본이 이미 계상한 호출이며, retry matrix 의 취소 항목과 대조한다(tasks 1.5).
- 원장: 취소 attempt(`kind=CANCEL`)가 처음으로 기록되기 시작한다. 스키마 변경 없음 — 수동 flatten 이 이미 쓰는 형식.
- 설정·토글 없음. 정본이 이미 요구하는 동작의 수리이므로 토글 뒤에 두지 않는다.
- 롤백 = 이전 이미지(취소가 다시 거절되는 종전 상태). 이 change 가 쓴 취소 attempt 는 옛 이미지가 읽는 기존 형식이다.
- 진행 중 change 와의 겹침: a087(같은 파일, `landed-commit.txt` 없음)·a167(등록 단계, `record`·`submit` 예정) — tasks 1.4.
