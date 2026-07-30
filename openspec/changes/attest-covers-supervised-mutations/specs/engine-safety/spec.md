# engine-safety Specification (delta)

## ADDED Requirements

### Requirement: attestation endpoint 집합의 증거원

attestation의 성공 endpoint 집합은 **각 항목이 그것을 증명할 수 있는 증거원에서만** 와야 한다(SHALL).
증거원은 둘이고 역할이 겹치지 않는다:

- **무인 read-only soak** — 읽기 endpoint를 증명한다. soak 기록의 비-GET 항목은 attestation에
  실려서는 안 된다(SHALL NOT) — 아무것도 접수하지 않는 도구의 기록에 mutation이 있다는 것은
  측정이 아니라 기록 오염이다.
- **사람이 승인한 감독 검증** — 무인 도구가 구조적으로 실행할 수 없는 mutation endpoint를
  증명한다.

감독 검증 기록은 **무인 도구가 실행할 수 없다고 선언된 endpoint 목록에 있는 것만** 기여할 수
있다(SHALL). 그 목록 밖의 endpoint는 감독 검증 기록이 성공을 증명하더라도 attestation에
실려서는 안 된다(SHALL NOT). 근거: 감독 하 1회 성공은 여러 날의 무인 운전이 증명하는 것과
같은 속성이 아니며, 읽기를 감독 검증으로 대신 증명하면 soak 결함이 조용히 덮인다.

감독 검증 기록이 endpoint를 기여하려면 다음이 **모두** 참이어야 한다(SHALL):

1. 그 endpoint의 호출이 오류 없이 **성공**했다
2. 그 기록의 계좌가 attestation의 계좌와 같다
3. 성공 시각이 attestation의 유효 기간 안이다
4. 그 endpoint가 무인 도구 실행 불가 목록에 있다

기록의 계좌가 attestation의 계좌와 **다르면** 조용히 건너뛰지 않고 발급을 거부해야 한다(SHALL) —
기대 경로에 다른 계좌의 기록이 있다는 것은 설정 오류이고, 무시하면 그 오류가 "증거 없음"과
구별되지 않는다.

attestation은 각 mutation endpoint를 **무엇이 증명했는지** 기록해야 한다(SHALL) — 최소한
endpoint, 성공 시각, 증거 기록의 출처. 근거: 인터록 거부 메시지는 무엇이 빠졌는지만 말하므로,
게이트가 통과한 뒤 "무엇을 근거로 켜졌나"에 답할 수 있는 곳은 attestation 자신뿐이다.

요구 endpoint 중 하나라도 어느 증거원으로도 채워지지 않으면 attestation은 그것을 **싣지 않은
채** 발급되고, 기동 인터록이 그 부족을 근거로 거부한다(SHALL) — 이 요구는 인터록이 요구하는
집합을 바꾸지 않는다(SHALL NOT).

#### Scenario: 감독 검증이 mutation endpoint를 증명한다

- **WHEN** 사람이 승인한 감독 검증이 주문 접수와 취소를 성공시킨 기록이 있고 soak이 완료된 상태에서 attestation을 발급하면
- **THEN** 그 두 mutation endpoint가 attestation의 성공 집합에 실리고, 각각을 무엇이 증명했는지가 함께 기록된다

#### Scenario: 감독 검증은 읽기를 증명하지 못한다

- **WHEN** 감독 검증 기록이 어떤 읽기 endpoint의 성공을 담고 있어도
- **THEN** 그 읽기는 감독 검증을 근거로 attestation에 실리지 않는다 — 읽기는 무인 soak이 증명한다

#### Scenario: 실패한 호출은 증거가 아니다

- **WHEN** 감독 검증 기록의 mutation 호출이 오류로 끝났으면
- **THEN** 그 endpoint는 실리지 않는다

#### Scenario: 계좌가 다른 증거는 발급을 거부시킨다

- **WHEN** 감독 검증 기록의 계좌가 soak의 계좌와 다르면
- **THEN** attestation은 발급되지 않고 사유가 보고된다

#### Scenario: 유효 기간 밖의 증거는 증거가 아니다

- **WHEN** mutation 성공 시각이 attestation 유효 기간보다 오래됐으면
- **THEN** 그 endpoint는 실리지 않는다

#### Scenario: 부족한 채로 발급되고 인터록이 거부한다

- **WHEN** 감독 검증이 아직 없어 mutation endpoint가 하나도 채워지지 않았으면
- **THEN** attestation은 읽기만 담은 채 발급되고, 게이트 ON 기동은 그 부족을 근거로 거부된다
