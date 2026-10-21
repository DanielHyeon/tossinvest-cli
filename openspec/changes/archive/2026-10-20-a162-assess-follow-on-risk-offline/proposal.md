## Why

사용자는 a156의 후속 작업을 운영에 연결하지 않는 오프라인 판정부터 설계하도록 승인했다. 세 전략의 KR·US proposal API와 배분 계산은 이미 있다. 이를 다시 만들지 않고, 기존 후보 수량과 명시적인 위험 가정값을 결합해 잔여 예산 안의 가정 수량을 확인한다.

## What Changes

- continuation·reversal·weekly의 기존 native proposal 입력을 같은 프로세스 안에서 재사용한다. 기존 전략 계산·계획·가중치·보호선은 변경하지 않는다.
- 계좌 통화의 최소 단위로 표현한 명시적 주당 위험 상한 가정값과 체결·예약 위험을 이용해 `q_offline`을 계산한다. `q_candidate`·`q_offline`·실제 주문의 `qFinal`을 구분한다.
- 입력 누락·위험 미확인·범위 불일치·예산 초과·계산 범위 초과는 거절한다. 결과는 가정에 대한 검사이며 실계좌 최신성이나 실행 권한을 증명하지 않는다.

## Capabilities

### New Capabilities

- `offline-follow-on-assessment`: 기존 proposal을 재사용하는 비권한 후속 진입 위험 시나리오 판정.

### Modified Capabilities

없음. a156의 운영 승인·실측 조건 및 a160/a161의 비권한 조회·복원 경계를 유지한다.

## Impact

구현 범위는 신규 `internal/offlinecontinuation`의 작은 typed adapter·정수 계산과 관련 시험이다. weekly 후속 leg 양성 시험에는 기존 `tossos_testseams` 태그 전용 fixture의 최소 확장을 허용하며 기본 빌드의 권한 표면은 변경하지 않는다. UI·CLI·HTTP·DB·원장 schema·Guardian·engine·운영 설정·a157은 수정 대상이 아니다. 기존 증거, 원장 및 인증·토큰 공유를 보존한다. 실제 수량 승인·예약·주문 연결·마감 청산은 후속 단계다.
