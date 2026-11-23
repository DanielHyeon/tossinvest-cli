# a198 보강 증거 — D7(이미 끝난 주문의 취소 거절)·D8(접수 확정 취소가 있는 매수) (2026-11-24)

기준: 브랜치 HEAD `72be1707e`(base `e1eaf720c` + a198 구현). AST 는 `analysis/cited-ast/`(tools/logic-map, 같은 작업 트리) — **단 `cited-ast/ExitObserver_clearTheSymbol.ast.json` 은 1단계에서 뽑은 편집 전(base) 판이다**(source_sha256 `bc81b5fd…`). 이 문서의 `clearTheSymbol` 분기 번호(B5·B9·B10 등)와 줄(L1669 …)은 편집 뒤 판 `analysis/function-logic/internal-app-engine--exitobserver.clearthesymbol/ast.json`(sha `6ffb1ec1…`)을 따른다(보강 리뷰 r1 P2-6 정정). 리뷰 탐침 원문은 `analysis/review/impl-r1-probe_test.go.txt`(구현 적대 리뷰 r1, 이 change 의 시험 하네스를 확장한 버리는 파일).

## 1. 계약 원문 — 취소 409

`docs/migration/openapi.latest.json:6496` `/api/v1/orders/{orderId}/cancel` POST 의 응답은 200·400·401·404·409·422·429·500. 409 = 「취소 불가 상태」(`:6565`), 스키마 `ErrorResponse`(required `error`, 그 아래 `ApiError`) — **code 자리는 `error.code` 뿐**(최상위 `code` 는 계약 예시에 없음). 예시 다섯:

| 줄 | code | 설명(summary / message) |
|---|---|---|
| 6578 | `already-filled` | 이미 체결된 주문 / 「이미 체결된 주문입니다.」 |
| 6588 | `already-canceled` | 이미 취소된 주문 / 「이미 취소된 주문입니다.」 |
| 6598 | `already-modified` | 이미 정정된 주문 / 「이미 정정된 주문입니다.」 |
| 6608 | `already-rejected` | 거부된 주문 취소 시도 / 「거부된 주문은 정정/취소할 수 없습니다.」 |
| 6618 | `already-processing` | 주문 처리 중 / 「잠시 후 다시 시도해 주세요.」 + `data.retryAfterSeconds` — a165 의 대상 |

- 같은 다섯 code 가 `/api/v1/orders/{orderId}/modify` 409 에도 있다. `/api/v1/orders` POST 409 는 `request-in-progress`·`opposite-pending-order-exists` 뿐(`:5739`) — **`request-in-progress` 는 취소 응답 계약에 없다**(발주 전용).
- 실물: 저장소 전수 grep(`already-filled|already-canceled|already-rejected|already-modified`, .go·.json·.md·.txt, openapi 제외) **0 건** — fixture·측정 기록·운영 관측 없음. **계약 문서만, 실물 미측정.** (`already-processing` 은 실측 있음 — verifylive/us_market_test.go:310 「Measured on 2026-07-27」.)

## 2. 분류 사슬과 종류(kind)별로 가를 수 있는 자리

| 함수(AST) | 종류 인자 | 409 already-* 의 오늘 결과 |
|---|---|---|
| `classifyMutation`(classify.go:21-99, 분기 10) — 호출 자리는 `Gateway.submit` 의 dispatch 클로저 1곳(gateway.go:729) | **없음**(send·result·err) | B3 `classifyRefusalCode` → 목록 밖이라 판정 없음 → B6 `ClassifyBrokerRefusal` 해당 없음 → B8 `statusOf` 409 → `journal.ClassifyHTTPMutation` |
| `classifyRefusalCode`(refusal_code.go:53-78, 분기 6) · `definitiveRefusalCodes`(:46, 값 `opposite-pending-order-exists` 하나, 읽는 자리 :73 하나) | 없음 | B6 목록 밖 → `refusalCodeNone` |
| `journal.ClassifyHTTPMutation`(dispatch.go:301) · `isDefinitiveRejection`(:351 `400, 401, 403, 404, 405, 415, 422`) | 없음 | 409 는 목록 밖 → default **DispatchAmbiguous**(`dispatch_outcome_unknown`) → attempt **IN_DOUBT** |
| `Gateway.withCancelRefusal`(cancel_refusal.go:44-106, 분기 13) — submit 이 dispatch 를 감싸는 자리(gateway.go:733) | **있음** — B1 `g.cancelRefusal == nil \|\| plan.kind != journal.KindCancel` 이면 원 dispatch 그대로 | 첫 결과가 모호이고 `alreadyProcessingRefusal`(:109, B4 code == `already-processing` 만) 일 때만 손댐 → already-* 는 B2 에서 원 결과(IN_DOUBT) 그대로 반환 |

- 결론: **CANCEL 만 가를 수 있는 기존 자리는 `withCancelRefusal` 하나**다. 운영의 게이트웨이 두 개는 모두 정책을 단다 — 엔진 `EngineCancelRefusal()`(internal/app/engine/gateway.go:329, `Definitive: true`), flatten `FlattenCancelRefusal()`(cmd/tossctl/flatten.go:251, 재시도 2·5s·8s). `execgw.New` 의 운영 호출자는 이 둘뿐(grep).
- D7 의 최소 편집안(가): `withCancelRefusal` 1 함수 + 새 판정 함수(새 파일, `alreadyProcessingRefusal` 과 같은 `refusalCodes` 판독·같은 모순 규칙). 첫 결과가 모호이고 409 의 code 가 already-filled·canceled·rejected·modified 중 하나면 확정 거절로 바꿔 반환. 편집 기존 함수 1.
- 대안(나): `classifyMutation`·`classifyRefusalCode` 에 종류 인자 → 기존 함수 2 + `Gateway.submit`(분기 57) 호출 자리 1줄 = 기존 함수 3, submit FLM 번들 필요. (가)가 최소.
- (가)의 정책 결합: B1 때문에 정책이 nil 인 게이트웨이(시험 다수 — `TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome` cancel_refusal_test.go:189 이 「nil 정책 취소는 결과 불변」을 고정)에는 D7 이 걸리지 않는다. 운영 영향 0(운영 게이트웨이 둘 다 정책 있음). D7 을 정책과 무관하게 걸려면 B1 의 구조를 바꿔야 하고 그 시험의 뜻(「정책 없음 = 손대지 않음」)과 부딪친다 — **Manager 결정 필요**(아래 질문 1).

## 3. 사유 코드·골든

- `AllReasonCodes`(failclosed.go:254)는 33 값, 동결 골든 `internal/execgw/testdata/reason_codes.golden`(33 줄, failclosed_test.go·gen_golden_test.go·a098_sender_down_reason_test.go 가 읽음). 선례: a094 가 `ReasonOppositePendingOrder = "opposite_pending_order_exists"`(refusal_code.go:29)를 더하며 `AllReasonCodes` 에 등록하고 골든에 줄을 더했다.
- 재사용 가능한 기존 값: `ReasonBrokerRejected = "broker_rejected"`(reason.go:114 — a165 엔진 `Definitive` 가 already-processing 을 확정 거절로 바꿀 때 쓰는 값, cancel_refusal.go:58), `journal.ReasonDispatchRejected = "dispatch_rejected_by_broker"`(lifecycle.go:94 — status 400 등의 확정 거절 값, a198 실패 줄에서 실측).
- 새 값을 만들면(예: 「취소 대상이 이미 끝남」) `AllReasonCodes` 편집(기존 함수 +1, FLM 대상)과 골든 한 줄 + 재생성이 필요. 재사용하면 0. **Manager 결정 필요**(질문 2).
- `analysis/goldens/` 동결 골든: 저장소에 있는 것은 아카이브 a112 의 것(openspec/changes/archive/2026-10-04-a112-…/analysis/goldens/) 뿐 — 이 사유 코드와 무관(grep `broker_rejected|opposite_pending` 0).

## 4. a165 「취소 409 재시도 창」과의 순서

- `withCancelRefusal` 흐름(AST): B1(정책 없음·비취소 → 그대로) → 첫 dispatch → B2(`already-processing` 이고 모호가 아니면 그대로 반환) → B3 엔진(`Definitive`) 이면 `broker_rejected` 로 확정 거절 → B4 정책 값 0 이면 그대로 → B5 재시도 루프(flatten: 최대 2 회, 누적 5s, 8s 창) → 다음 결과가 `already-processing` 모호가 아니면(B12) 그 결과를 반환(꼬리표 `outcome=retried`).
- already-* 는 `already-processing` 과 code 가 달라 **재시도 창에 들어가지 않는다**(두 판정은 서로소). D7 이 서야 할 자리는 둘: ① 첫 결과(B2 앞 — 엔진·flatten 공통), ② flatten 재시도의 다음 결과(B12 에서 반환되는 `next`) — 409 already-processing 뒤 재전송이 409 already-canceled 를 받으면, **우리 첫 취소가 처리 중이던 바로 그 취소가 성립했을 수 있다**(「취소 성립」을 「취소 미성립」 거절로 읽는 모양). 그 경우 확정 거절은 「치움 미완료 → 종결 기록 대기」로 보수 방향이지만 flatten 단계 결과의 뜻은 틀린다.

## 5. 다른 취소 호출자

| 경로 | 오늘(already-* → IN_DOUBT) | D7(가) 뒤 |
|---|---|---|
| flatten `Saga.settleStep`(flatten.go:301) → `classifyCancel`(:454, B5 IN_DOUBT) → `resolveStep`(:423) → `Resolver.Resolve` → `resolveCancel`(amend_indoubt.go:44) 이 그 주문을 읽어: 취소됨 → CONFIRMED(`Done`), 체결 → FAILED_CONFIRMED(`Pending`, 「다음 실행이 다시 취소」), 거부 → FAILED_CONFIRMED, **정정됨(REPLACED) → park(UNRESOLVED_IN_DOUBT)**(B19, 「API 가 말해 주지 않음」) | 정책이 flatten 에도 있으므로 D7 이 걸림 → FAILED_CONFIRMED → `classifyCancel` B3 → **`FlattenStepFailed`**(report.Failed++, 재실행 시 「already settled」 로 건너뜀 :317). 해소 관측이 사라지고, 이미 끝난 주문인데 「실패」로 보고된다. flatten 시험 중 409 already-* 를 고정하는 것은 없음(grep — a165_cancel_refusal_test.go 는 already-processing 만) |
| 기동 복구 `Recovery.Run`(recovery.go:247) → `Resolver.Resolve`(:300) → `resolveCancel` | CANCEL IN_DOUBT 를 위 관측으로 해소(재시작이 필요했던 이유) | IN_DOUBT 가 생기지 않으므로 해당 없음 |
| 세션 중 해소 `in_doubt_settle.go:175` | `rec.State == IN_DOUBT && rec.Kind == KindPlace` 만 — CANCEL 은 대상 밖(그래서 「재시작까지」) | 해당 없음 |

- `already-modified` 주의: 기동 해소는 정정된 주문의 취소를 의도적으로 판정하지 않는다(resolveCancel B19 park). 엔진에는 정정 경로가 없으므로 엔진 주문의 정정은 **사람(앱)** 이 한 것이고, 그 후계 주문 번호는 원장에 없다. D7 이 이것을 확정 거절로 두면 길 치우기는 옛 번호의 종결 기록을 기다린 뒤 손절을 낸다 — 그 사이 원장에 없는 후계 주문(그것이 매도면 두 번째 매도)이 살아 있을 수 있다. 오늘은 IN_DOUBT 로 재시작까지 막혀 있었다. → `already-modified` 를 D7 에 넣을지 **Manager 결정 필요**(질문 3).

## 6. D8 — 접수 확정 취소가 있는 매수

- 오늘 분기: `clearTheSymbol` B5(L1669 `!buy`) 가 거짓이면 ConfirmedCancelOf 를 읽지 않고 곧장 B9·B10(해석·발행) → 취소 재전송. D8 = 매수에도 `ConfirmedCancelOf`(exit_proposal_release.go:150 — `kind='CANCEL' AND state=CONFIRMED AND target_order_id=그 주문`, 취소 intent 의 시장·거래일·종목 = 그 주문의 것) 를 읽고, 찾으면 **취소를 내지 않고 `continue`**(res.cleared 를 건드리지 않음 — 정본 「매수의 취소는 접수로 치운다」). 매도는 종전 B7(대기). 편집 함수: `clearTheSymbol` 만(이미 편집 대상).
- 그 매수가 `LiveOrdersForSymbol` 에 남는 기간: 종결 체결 스냅숏이 들어올 때까지(fills.go:1854 `NOT confirmedOrderTerminalEvidence`) — 체결 감지 주기(a191 의 목록 이탈 종결 확인 포함)에 묶임, 미측정.
- 취소 접수 뒤 체결 경합: 접수 확정된 취소가 실제로는 체결에 졌으면(부분·전량) 그 체결은 체결 감지가 원장에 적는다. **정정(보강 리뷰 r1 P2-4)**: 그 체결이 손절 접수 **뒤**(포지션 CLOSING)에 오면 프로젝션은 보유를 늘리지 않는다 — 전이표 E31~E33(internal/position/table.go:78-80, 직접 확인)이 `ENTRY_WHILE_CLOSING` 거절이고, 프로젝션은 정상 수량 갱신 대신 RECONCILE 로 간다(리뷰 인용 position_projection.go:101 — 이 문서 작성자는 그 줄을 직접 열지 않음). 즉 「다음 주기 프로젝션이 증가분을 보호한다」는 이전 문장은 틀렸고, 그 경우의 회복은 대사(RECONCILE) 경로다. 손절 접수 **전**에 오면 보유가 늘고 그 주기 손절 수량은 경합 전 값일 수 있다(과소 보호). 어느 쪽도 초과 매도 방향은 아니다. D8 은 그 매수를 「치워짐」으로 보므로 다음 주기에 다시 취소하지 않는다. 두 경우의 시험은 없다(미측정).
- D8 뒤 P2 모양(접수 → 손절이 반대 미체결로 거절): 다음 주기 재취소 0, 손절 재제출 — 브로커가 취소를 반영할 때까지 손절이 주기마다 `opposite-pending-order-exists` 로 거절될 수 있다(확정 거절이라 종목을 잠그지 않음; 반복 빈도 = 관측 주기).
- ConfirmedCancelOf 의 거래일 결속(전일 주문) 잔여는 그대로 — 그 경우 매수 재취소 → 계약상 409 already-canceled → D7 확정 거절 → 치움 미완료로 종결 기록 대기(막힘이 아니라 일시 보류).

## 7. 편집 표면(보강분)

| 기존 함수 | 이유 | 고정·census |
|---|---|---|
| `ExitObserver.clearTheSymbol` | D8 분기 · 리뷰 P2(수량 해석 실패 줄, 표시를 쓰기 뒤에) | a163 렌더 해시 밖(record 만), a166 census 밖 |
| `Gateway.withCancelRefusal` | D7(안 (가)) | a166 census 키 없음(shapes.json 에 cancel_refusal.go 없음). `TestA165NilPolicyAndNonCancelKeepTheOriginalOutcome` 은 already-processing 출력 고정 — D7 이 already-* 만 다루면 불변 |
| (선택) `AllReasonCodes` | 새 사유 코드를 만들 때만 | 골든 1줄 |
| 새 함수(새 파일) | D7 판정(`refusalCodes` 재사용), 리뷰 P2 의 표시 규칙은 기존 a198 새 함수 수정 | — |

- 뒤집히는 기존 시험: 없음 예상(grep — already-* 를 IN_DOUBT 로 고정하는 시험 0, a165 시험은 already-processing 만). 미실행.
- `Gateway.submit`·`classifyMutation`·`classifyRefusalCode`·`journal` 무편집(안 (가)).

## 8. 정본과의 충돌 (보강 리뷰 r1 P2-3, 직접 확인)

정본 order-execution 「취소의 일시적 거절 처리」 의 Scenario 「다른 사유의 409」(openspec/specs/order-execution/spec.md:590-592): **「게이트웨이의 취소가 409 로 거절되고 본문 code 가 `already-processing` 이 아니거나 두 code 자리가 다르면 → 이 처리 이전과 같이 결과 불명으로 기록된다」**. D7 의 새 Requirement(409 `already-filled` 등 → 확정 거절)와 같은 입력에 반대 결과를 요구한다. ADDED 만으로는 둘 다 만족할 수 없다 — 그 Scenario 를 MODIFIED 로 고치는 delta 가 필요하다(문서 미편집, Manager 결정 대기). 같은 Requirement 의 다른 문장과 「409 는 확정 status 목록에 넣지 않는다」는 D7 과 충돌하지 않는다(D7 은 status 가 아니라 code 로 판정).
