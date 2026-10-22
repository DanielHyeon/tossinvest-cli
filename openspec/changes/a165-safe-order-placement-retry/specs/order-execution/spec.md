> proposal 단계 초안 delta. 사용자 결정(2026-10-22)을 요구 문장으로 옮긴 것이며, design 단계에서 기존 요구
> 「Retry Matrix 산출물」·「취소의 일시적 거절 처리」의 MODIFIED 로 다시 쓴다(지금의 ADDED 는 그 두 요구의 재시도 금지 문장과
> 충돌한다 — archive 전에 반드시 해소한다).

## ADDED Requirements

### Requirement: 주문 접수의 안전 재시도

주문 접수(`POST /api/v1/orders`)는 브로커가 그 요청을 **확실히 접수하지 않았을 때만** 즉시 재시도할 수 있다(SHALL — 그 밖에는
즉시 재시도하지 않는다). 확실한 미접수는 HTTP 429 와 요청을 보내기 전의 연결 실패다. 재시도는 같은 `clientOrderId` 와 같은 본문으로만
하고(SHALL), 최초 전송부터 8초 예산 안에서 최대 2회다(SHALL).

결과가 불명이면(타임아웃·5xx·전송 후 연결 끊김) 맹목 재전송해서는 안 된다(SHALL NOT). 먼저 `clientOrderId` 또는 미체결 목록으로 접수
여부를 조회하고, 접수가 없다고 확인될 때만 같은 키·같은 본문으로 재전송할 수 있다(SHALL). 확정 거절(4xx 검증 오류)은 재시도하지
않는다(SHALL NOT). 재시도 사실·횟수·갈래는 attempt 기록에 남아야 한다(SHALL).

조건주문 생성은 이 요구의 대상이 아니다(재시도 0회 유지). IN_DOUBT 해소의 재생은 이 요구로 켜지지 않는다.

#### Scenario: 429 뒤 같은 키 재시도

- **WHEN** 주문 접수가 HTTP 429 로 끝나고 최초 전송부터 8초가 지나지 않았으며 재시도가 2회 미만이면
- **THEN** 같은 `clientOrderId`·같은 본문으로 다시 보낸다

#### Scenario: 결과 불명은 조회 먼저

- **WHEN** 주문 접수가 전송 후 타임아웃으로 끝나면
- **THEN** 재전송 전에 접수 여부를 조회하고, 접수가 확인되면 재전송하지 않는다

#### Scenario: 확정 거절은 재시도 없음

- **WHEN** 주문 접수가 4xx 검증 오류로 거절되면
- **THEN** 자동 재시도는 일어나지 않는다

### Requirement: 엔진 경로 취소·정정의 409 재시도

엔진 게이트웨이의 취소·정정이 HTTP 409 이고 본문 code 가 `already-processing` 이면 `retryAfterSeconds` 를 따라 재시도할 수 있다(SHALL —
총 대기 5초 이내, 최대 2회). HTTP 상태와 본문 code 가 모두 일치할 때만 재시도한다(SHALL).

#### Scenario: 다른 사유의 409

- **WHEN** 취소가 409 로 거절되고 본문 code 가 `already-processing` 이 아니면
- **THEN** 재시도하지 않는다
