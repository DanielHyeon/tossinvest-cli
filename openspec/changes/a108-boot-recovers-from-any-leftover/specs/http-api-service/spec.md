# http-api-service — a108 delta

> httpapi는 strategy projection descriptor가 없으면 전략 없이 기동하지만, descriptor가
> 있는데 dial이 실패하면 종료한다. 2026-08-13 재부팅의 반쪽 잔재가 후자를 때려 crash
> loop를 만들었다. A2 리뷰(2026-08-14)가 두 구멍을 더 보였다: socket 파일이 남은
> 사망(전원 단절의 기본 모양)은 Dial이 연결하지 않아 강등이 아예 걸리지 않고 집계
> 스냅샷 전체가 함께 죽으며, 비-NotExist stat 오류의 fatal은 콘솔과 갈라진 crash loop다.

## ADDED Requirements

### Requirement: strategy projection의 부재와 모든 실패 형태는 같은 강등이다

httpapi는 strategy projection endpoint의 부재(descriptor 없음)와 모든 실패 형태(descriptor는 있으나 연결 불가, socket 파일은 있으나 주인 사망, descriptor 조사 불가)를 동일하게 처리해야 하며(SHALL) 전략 표면 없이 기동하고 나머지 API를 제공한다. 잔재·죽은 endpoint가 httpapi를 재시작 루프에 빠뜨려서는 안 되고(SHALL NOT), 가동 중 strategy 읽기 실패가 나머지 집계 스냅샷·스트림을 실패시켜서는 안 된다(SHALL NOT). endpoint 연결은 생존을 확인한 뒤에만 성립해야 한다(SHALL) — 파일 모양 검사만으로 살아 있다고 판정하지 않는다.

#### Scenario: 반쪽 잔재 위의 httpapi 기동

- **WHEN** descriptor는 존재하지만 socket이 없어 연결이 실패하는 상태에서 httpapi가
  기동하면
- **THEN** 전략 표면 없이 기동하며, 관측 동작은 descriptor-부재 기동과 동일하다

#### Scenario: socket 파일이 남은 사망 (전원 단절의 기본 모양)

- **WHEN** descriptor와 socket 파일이 모두 남았지만 주인이 죽어 연결을 수락하지 않는
  상태에서 httpapi가 기동하면
- **THEN** 연결 시도가 생존 부재를 판정하고 전략 표면 없이 기동한다 — 집계 스냅샷은
  살아 있다

#### Scenario: 가동 중 projection 사망

- **WHEN** 기동 시 살아 있던 projection이 이후 죽어 strategy 읽기가 실패하면
- **THEN** 전략 표면만 unavailable(`RUNTIME_UNAVAILABLE`)이 되고 나머지 집계 스냅샷과
  스트림은 계속 제공된다 — dormant(`NOT_CONFIGURED`)는 reader 부재(기능 미사용)에만
  쓰며, 죽음을 미사용으로 접지 않는다

#### Scenario: 조사 불가 descriptor

- **WHEN** descriptor 경로의 stat이 부재가 아닌 오류(권한·ENOTDIR 등)를 돌려주면
- **THEN** 경고를 남기고 전략 표면 없이 기동한다 — 같은 디스크 상태에서 콘솔과 같은
  판정이다
