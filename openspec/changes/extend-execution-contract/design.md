# Design: extend-execution-contract

## Context

P1은 place/cancel/amend에 대해 완결된 안전 계약을 만들었다: journal 선기록 → DISPATCH_STARTED → 분류 → IN_DOUBT 해소, GuardianDecision(intent 해시·한도 스냅샷·만료·one-shot nonce) 재검증, 심볼당 in-flight 1개, raw mutator 봉인. 조건주문은 그 계약 밖에 있고 — CLI 사용자가 확인 토큰을 직접 입력하는 경로로 설계됐기 때문에 사람이 키보드 앞에 있을 때는 타당하다 — 무인 자동 보호 경로로 쓰려면 계약 안으로 들여와야 한다.

리뷰 근거는 `../add-core-domain/review.md` A·B 클러스터(14건).

## Goals / Non-Goals

**Goals**: 조건주문을 P1 계약에 편입, 발동 주문의 체결 귀속, 위험 감소 mutation의 우선 통행권, 결정의 위험 입력 결합, 한도·게이트의 fail-closed 완성.

**Non-Goals**: Guardian 판정 체인·사이징·RR·한도 **수치**(add-core-domain), 포지션 aggregate·보호 saga(add-core-domain), 실계좌 능력 검증(verify-execution-capability), MCP 표면의 Gateway 우회(P4 단일 writer 데몬까지 문서화된 잔존 리스크).

## Decisions

### D1. 조건주문은 새 MutationKind이지 새 경로가 아니다

`KindConditionalPlace`/`KindConditionalCancel`/`KindConditionalModify`를 `journal.MutationKind`에 추가하고 기존 Gateway 시퀀스를 그대로 태운다. 새 상태기계를 만들지 않는 이유: P1의 RECORDED→DISPATCH_STARTED→ACKED/IN_DOUBT는 "요청은 보냈는데 응답을 못 받았다"를 다루는 것이고, 그 문제는 주문 유형과 무관하게 동일하다. 별도 기계를 만들면 crash 케이스를 두 번 증명해야 한다.

CLI 표면(`tossctl` 조건주문 명령)은 upstream의 확인 토큰 경로를 **그대로 유지**한다. 엔진 경로만 Gateway를 경유한다. 두 경로가 같은 `official` 어댑터를 공유하되 진입점이 다르다.

### D2. 발동 주문 귀속은 "예상 주문" 등록으로 푼다

조건주문 등록이 확정되면 journal에 `expected_order` 레코드(조건주문 ID·심볼·방향·최대 수량·연결된 포지션)를 남긴다. 체결 감지가 로컬 intent 없는 브로커 주문을 만나면 먼저 `expected_order`와 대조하고, 매칭되면 그 조건주문의 lineage로 귀속한다. 매칭되지 않을 때만 외부 주문으로 분류한다.

대안이었던 "발동 시점에 intent를 소급 생성"은 기각했다 — intent는 불변 의도 기록이고, 우리가 내지 않은 주문에 의도를 지어내면 provenance가 거짓이 된다. `expected_order`는 "우리가 이런 주문이 생길 것을 알고 있었다"는 사실만 기록하므로 정직하다.

`NetPositions` 쿼리의 귀속 규칙 확장은 P1 reconcile의 "외부 주문" 판정에 직접 영향을 주므로, 기존 reconcile 테스트가 회귀 없이 통과함을 증명해야 한다.

### D3. mutation safety class로 직렬화를 분리한다

두 클래스:
- **EXPOSURE_RAISING**: 진입 place, 보호 없는 수량 증가
- **RISK_REDUCING**: 보호 생성·증량, reduce-only 청산, 모든 cancel

직렬화 규칙: EXPOSURE_RAISING은 심볼당 1개(P1 규칙 유지). RISK_REDUCING은 EXPOSURE_RAISING의 in-flight·IN_DOUBT에 막히지 않고, 자기들끼리는 (조건주문 ID 또는 대상 주문 ID) 단위로 직렬화한다.

oversell 방지는 in-flight 차단이 아니라 **수량 상한**으로 한다: 모호한 진입이 있으면 그 진입의 최대 가능 체결수량을 보수적으로 가정하고, RISK_REDUCING 수량을 `min(확정 보유수량, 계좌 조회 매도가능수량)` 이하로 제한한다. 이것이 §0.3(손절 즉시성)과 oversell 방지를 동시에 만족시키는 유일한 조합이다 — 차단으로는 둘 다 만족할 수 없다.

### D4. RiskIntent는 결정에 결합되고 Gateway가 재검증한다

```
RiskIntent{ Account, Market, Symbol, Direction, EntryPrice, StopPrice, TargetPrice, Quantity, PolicyVersion }
```
canonical 직렬화 → hash. `GuardianDecision.RiskIntentHash`에 실리고, 제출 시 Gateway가 실제 주문 파라미터에서 재계산해 대조한다. 발급자(add-core-domain)가 어떤 판정을 했든, **판정의 입력이 제출과 다르면 거부된다**.

Phase 1 `IntentHash`(주문 파라미터 결합)는 유지하고 `RiskIntentHash`를 추가한다 — 전자는 "이 주문", 후자는 "이 위험 판단"을 묶는다.

### D5. 한도는 configured 비트로 fail-closed

`Limits`에 필수 항목별 명시적 존재 표시를 둔다. 현재 `MaxQuantity <= 0`을 "검사 안 함"으로 읽는 규칙이 부분 무제한 게이트를 허용하므로, "0 = 미설정 = 기동 거부"로 뒤집는다. 다만 **빈 스냅샷 전체**는 여전히 위험 감소 mutation을 통과시켜야 한다(§0.3) — "한도 없음"과 "한도 항목 누락"은 다른 상태다.

한도 집합: 주문 수량, 주문 notional, 총 개방 노출, 일일 손실(절대액), 일일 손실(자본 %). 통화는 전부 일치해야 한다.

인터록은 Guardian을 감사된 `gateLimits(gate)`에서 **구성**한다 — 검증하는 객체와 실제로 결정을 찍는 객체가 같아야 한다.

### D6. 위험 예약은 결정 발급과 같은 트랜잭션

`BEGIN IMMEDIATE` 안에서: 현재 노출·현금 조회 → 한도 대조 → 예약 행 삽입 → 결정 발급. 예약 해제는 (a) nonce 소비 후 체결·취소 확정, (b) 결정 만료, (c) 제출 실패 확정 중 하나에서 일어난다. IN_DOUBT는 해소 전까지 예약을 **유지**한다(fail-closed).

예약 없이 총계 한도를 지키는 방법은 발급 전체를 단일 락으로 직렬화하는 것인데, 체결 감지가 비동기(3s 폴링)라 락만으로는 "방금 통과한 주문의 노출"이 다음 판정에 보이지 않는다. 예약이 그 창을 메운다.

### D7. journal v5는 이 change에서 시작한다

추가 테이블: 조건주문 attempt(기존 attempt 테이블 확장), `expected_orders`, `risk_reservations`, `spent_nonces`. add-core-domain의 v6+(포지션·saga·성과)와 분리해 마이그레이션을 버전별로 immutable하게 유지한다.

**롤백은 구버전 바이너리 실행이 아니다** — `ErrSchemaTooNew`로 기동이 거부된다. 실제 복구 경로는 마이그레이션 전 백업 파일로의 복원이며, 이를 절차와 테스트로 만든다.

## Risks / Trade-offs

- [조건주문 fingerprint의 유일성 미실측] → verify-execution-capability의 능력 검증 항목. 유일 매칭 불가 시 재제출 금지 + UNRESOLVED_IN_DOUBT
- [D2 귀속 규칙이 P1 reconcile의 외부 주문 판정을 바꾼다] → 기존 reconcile 테스트 회귀 금지를 완료 게이트에 포함
- [D3 수량 상한이 계좌 조회 staleness에 의존] → staleness 임계 초과 시 RISK_REDUCING은 계속 허용하되 수량을 확정 보유수량으로만 제한(보수 방향)
- [`internal/trading/conditional.go` 수정] → CLI 경로 무변경을 characterization 테스트로 고정한 뒤 엔진 경로 추가

## Migration Plan

journal v5 additive. 마이그레이션 전 자동 백업 → 실패 시 백업 복원 절차. 스키마 계약 테스트로 v4→v5 전이와 구버전 거부를 고정한다.

## Open Questions

- 조건주문 목록 조회가 OPEN/CLOSED를 분리 제공하는지, pagination 커서가 일반 주문과 같은 형태인지 (구현 시 `internal/official` 확인, 불가 시 issues.md 기록)
- 발동 주문이 조건주문 ID를 응답에 실어주는지 — 실어주지 않으면 D2의 매칭이 (심볼·방향·수량·시각 창) 휴리스틱이 되고, 그 경우 매칭 실패는 외부 주문 분류 + 알림으로 fail-closed
