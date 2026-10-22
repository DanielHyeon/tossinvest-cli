# a168 · 검증 대사 읽기가 모호한 브로커 응답을 거절한다 (verify read decode hardening)

> **정의만 — 착수 전(2026-10-22).** proposal 수준. design·FLM·RED 는 아직 없다. 아래 함수 내부 서술은 a164 리뷰(5.1)의
> 지적을 **인용**한 것이고, 분기 주장의 근거가 되려면 착수 때 `tools/logic-map` AST 산출물을 먼저 만든다(tasks 0.3 — 비례 원칙상
> 공용 전송은 High-risk 경로라 면제 없음).
> 출처: a164 `review.md` 「5.1 구현 리뷰」 처분표 C-P1 잔여 · C-P2a, 그리고 Manager 처분(같은 파일 「Manager 처분」 절, 2026-10-22).

## 왜

a164 는 by-id 상세 해독(`internal/official/reconcile_detail_reads.go:96` `reconcileObject`)이 중복 키·뒤따르는 데이터를 토큰 단위로
거절하도록 고쳤다(a164 `f4e71ad1d`, 변이 4.22). 같은 리뷰가 그 수리 **밖**에 남은 두 자리를 찾았고, 둘 다 공용 경로라 a164 범위 밖으로
미뤘다. 줄 번호는 a164 착지 커밋 기준이다.

1. **목록 행 해독의 중복 키 접기(C-P1 잔여, Codex).** a121 공용 해독 `DecodeReconcileConditionalPage`
   (`internal/official/reconcile_reads.go:136`)는 행마다 `json.Unmarshal(raw, &o)`(`:144`)로 구조체 `apiConditionalOrder`
   (`internal/official/conditional_reads.go:28`)에 넣는다. `encoding/json` 은 중복 키의 **마지막 값**을 남기므로
   `"first":{"triggeredOrderId":"CHILD","triggeredOrderId":null}` 같은 행은 미발동으로 읽힌다. CLOSED 목록에서 **다른 id** 행의
   발동 흔적을 이렇게 숨기면 a164 의 A5 셋째 판정(`checkReconcileNoFiredTrace`)과 a121 부재 모드의 발동 흔적 판정이 그것을 못 본다 —
   **두 모드 공통**. 대상 자신의 발동은 a164 by-id 해독이 막는다. 봉투 층도 같은 성질 후보다: `decodeReconcileEnvelope`
   (`reconcile_reads.go:78`)는 `map[string]json.RawMessage`(`:83`)로, 공용 `unwrapAndDecode`(`internal/official/client.go:213`)는
   `apiEnvelope` 구조체로 해독한다(중복 `result`·`conditionalOrders` 키 — 착수 때 AST 로 확인할 후보이지 확정 결함이 아니다).
2. **공용 전송이 모든 2xx 를 성공으로 받음(C-P2a, Codex).** `Client.send`(`internal/official/client.go:320`)는
   `code < 200 || code >= 300`(`:362`)만 실패로 분류한다. a164 design D3 는 by-id 수락 조건을 「HTTP 200」 으로 적었다 —
   코드와 문서가 갈린다. 본문은 여전히 엄격 스키마·A4 를 통과해야 하므로 오늘 알려진 우회는 없지만, 계약은 하나여야 한다.

**왜 a164 에서 고치지 않았나(Manager 처분 2026-10-22):** 1 은 부재 모드 동작을 바꾸는 a121 공용 해독이고, 2 는 모든 공식 API 호출이
지나는 공용 전송(`send` — 주문 경로 포함, High-risk)이다. 둘 다 a164 의 「만료 대사 한 경로」 범위를 넘는다.

## 무엇을 (제안 — design 에서 확정)

1. 검증 대사 목록 행(조건주문·일반 주문)과 봉투의 해독이 **중복 키와 뒤따르는 데이터를 거절**한다 — a164 `reconcileObject` 와 같은
   토큰 단위 규칙을 재사용(새 해독기를 따로 만들지 않음, search-first). 거절은 대사 쓰기 0 으로 끝난다(fail-closed).
2. 성공 상태 계약을 하나로 맞춘다. 선택지(열린 질문 Q1): (a) 대사 리더만 HTTP 200 을 요구 — 공용 `send` 불변,
   (b) 공용 `send` 를 바꿈 — 모든 호출자 영향, (c) 코드 유지 + a164 design D3 문장을 「2xx」 로 정정.
   기본 권장은 (a) — High-risk 공용 전송을 건드리지 않고 대사 경로 계약만 좁힌다.

## 비목표

- 주문·정정·취소 요청 경로의 동작 변경(선택지 (b) 를 사람이 고르지 않는 한).
- 아래 「관련 잔여」 의 수리 — 이 change 는 기록만 이어받는다.
- 토큰 캐시 구조 변경(`internal/official/token.go`) — 기억 「공유 credential 토큰 전쟁」 의 결론대로 별도 앱 키가 정석.

## 관련 잔여 (a121 에서 물려받음 — a164 Manager 처분: 지금은 변경 없음, 기록만)

- **R-a. 401 토큰 갱신 재시도가 신원 재확인을 다시 하지 않음(C-P2b).** `Client.send` 의 401 루프(`client.go:344`)는
  `tm.refresh`(`internal/official/token.go:108`)로 받은 토큰으로 요청을 다시 만든다. 대사 신원 재확인 `reconcileReads.ensureIdentity`
  (`cmd/tossctl/verify_reconcile.go:226`, codex CG-6)는 첫 계좌 범위 읽기 전에 **한 번만** 돈다(`r.verified`, `:238`). 갱신이 디스크의
  다른 자격 토큰을 채택하면 그 뒤 읽기는 재확인 없이 진행된다. 계좌 헤더 seq 는 고정이다.
- **R-b. 마지막 신선도 재검이 직렬화 전 시각을 씀(C-P2c).** `internal/verifylive/reconcile.go:343` `at := now()` 뒤에 줄 직렬화가
  있고, `:359` `checkReconcileWindow(target, start, at, …)` 는 그 `at` 으로 15 s 창(`reconcileFreshnessBoundValue`, `:407`)을 잰다.
  직렬화·봉인·fsync 지연은 창 밖이다. 구조 핀 `TestReconcileFinalGateSitsAtTheWriteBoundary`
  (`internal/verifylive/reconcile_codex_test.go:185`)가 현재 위치를 고정한다. 만료 모드가 더한 직렬화는 마이크로초 단위.

착수 때 이 둘을 범위에 넣을지 사람이 정한다(열린 질문 Q2). 넣지 않으면 그대로 잔여로 남는다.

## 영향 (착수 때 CodeGraph 로 재측정)

- 후보 파일: `internal/official/reconcile_reads.go` · `internal/official/reconcile_detail_reads.go`(재사용 대상) ·
  선택지 (b) 일 때만 `internal/official/client.go`.
- 겹침 주의: 진행 중 `a165-safe-order-placement-retry` 가 `Client.send`·`doRequest`·401 재전송(`client.go:344`)을 다룬다.
  선택지 (b) 나 R-a 를 넣으면 a165 와 같은 함수를 편집하게 되므로 착수 전 Manager 가 순서를 정한다.

## 열린 질문 (사람)

- Q1. 성공 상태 계약 — (a) 대사 리더만 200 / (b) 공용 `send` / (c) 문서 정정. 기본 (a).
- Q2. 관련 잔여 R-a·R-b 를 이 change 범위에 넣을지. 기본: 넣지 않음.
- Q3. 착수 시점 — a164 6.x(사람 실기록 대사)·archive 뒤. 기본: 그 뒤.
