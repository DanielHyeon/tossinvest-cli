# operator-console Specification (delta)

## ADDED Requirements

### Requirement: 포지션 가시성
콘솔은 계좌의 보유·포지션 상태를 read-only로 표시해야 한다(SHALL): 브로커 스냅샷(수량·평균단가·현재가·평가손익·매도가능수량)과 journal 투영(positions·exit_states)을 심볼 기준으로 조인한다. 엔진 관리 포지션(`EntryDecisionID` 비어 있지 않음)은 exit 상태 — t0 진입가·최초 손절·기준선(Baseline)·워터마크(HighWater)·래칫 단계·ladder rung·부분익절 여부 — 를 함께 표시한다(SHALL). `EntryDecisionID`가 빈 보유(수동·외부 취득, 아직 편입 전)는 "편입 전"으로 구분하고 exit 라인이 없는 이유(엔진 미관리 — 편입은 `adopt-external-positions` change가 정의)를 화면에 명시한다(SHALL). 표시 로직은 데이터 주도다: exit_states 행이 생기면(편입 후) 같은 화면이 exit 라인을 표시한다. 어느 쪽 데이터 소스가 없거나 비어 있어도 다른 쪽만으로 렌더한다(SHALL) — 조인 실패가 화면 실패가 되어서는 안 된다.

#### Scenario: 엔진 관리 포지션의 exit 라인 표시
- **WHEN** journal에 exit_states 행이 있는 포지션을 포지션 화면이 렌더하면
- **THEN** 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부가 해당 심볼 행에 표시된다

#### Scenario: 수동 취득 보유의 정직한 구분
- **WHEN** 브로커 보유에는 있으나 EntryDecisionID로 정당화되는 포지션이 없는 심볼을 렌더하면
- **THEN** 해당 행은 "관리 외"로 표시되고 exit 라인 없음이 엔진 미관리 때문임이 안내된다

#### Scenario: 엔진 미가동 상태의 대시보드
- **WHEN** journal 파일이 없거나 포지션 테이블이 비어 있으면
- **THEN** 화면은 브로커 보유만으로 렌더되고 "엔진 미가동/관리 포지션 없음"을 명시하며 오류로 처리하지 않는다

### Requirement: 거래 이력 가시성
콘솔은 완결된 왕복 거래(trade_outcomes)와 exit 이벤트(exit_events)를 시간순으로 표시해야 한다(SHALL). 표시는 journal에 동결된 값을 그대로 사용하며 재계산하지 않는다(SHALL NOT).

#### Scenario: 동결된 왕복 결과 표시
- **WHEN** trade_outcomes에 행이 있는 상태에서 이력 화면을 열면
- **THEN** 각 왕복의 심볼·수량·진입/청산·손익·보유 시간이 journal 기록 그대로 표시된다

### Requirement: read-only 불변식
대시보드는 어떤 mutation도 수행해서는 안 된다(SHALL NOT): journal은 read-only로 열고(쓰기 연결 부재를 테스트로 고정), 주문·게이트·설정·exit 정책 조작 라우트는 존재하지 않는다(SHALL NOT). 기존 콘솔의 게이트 라우트 부재 가드는 대시보드 라우트 추가 후에도 유지된다(SHALL).

#### Scenario: journal 쓰기 시도 차단
- **WHEN** 콘솔 코드가 journal에 쓰기를 시도하는 변경이 들어오면
- **THEN** RO 접근 가드 테스트가 실패한다

#### Scenario: 주문 라우트 부재
- **WHEN** 콘솔의 라우트 표를 검사하면
- **THEN** 주문 발주·정정·취소·게이트 조작에 해당하는 라우트가 없다

### Requirement: rate budget 보호
브로커 스냅샷(보유·매도가능·시세)은 서버측 캐시를 통해 제공되어야 하며(SHALL) 캐시 TTL은 15초 이상이다. 화면 새로고침·다중 탭이 TTL 내 추가 브로커 호출을 유발해서는 안 된다(SHALL NOT). 검증 실행 중(runlock 활성)에는 스냅샷 갱신을 보류하고 캐시된 값을 사용한다(SHALL).

#### Scenario: 새로고침 연타
- **WHEN** TTL 내에 포지션 화면이 여러 번 새로고침되면
- **THEN** 브로커 호출은 추가로 발생하지 않고 캐시 시각이 화면에 표시된다

#### Scenario: 검증 실행 중 스냅샷 보류
- **WHEN** 검증 run이 진행 중일 때 포지션 화면을 열면
- **THEN** 새 브로커 호출 없이 캐시 값과 "검증 중 — 갱신 보류" 안내가 표시된다
