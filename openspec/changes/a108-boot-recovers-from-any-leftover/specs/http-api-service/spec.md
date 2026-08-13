# http-api-service — a108 delta

> httpapi는 strategy projection descriptor가 없으면 전략 없이 기동하지만, descriptor가
> 있는데 dial이 실패하면 종료한다. 2026-08-13 재부팅이 남긴 반쪽 잔재가 정확히 후자를
> 때려 컨테이너 crash loop를 만들었다.

## ADDED Requirements

### Requirement: strategy projection의 부재와 실패는 같은 강등이다

httpapi는 strategy projection endpoint의 부재(descriptor 없음)와 연결 실패(descriptor는
있으나 dial 불가)를 동일하게 처리해야 한다(SHALL): 전략 표면 없이 기동하고 나머지 API를
제공한다. 잔재·죽은 endpoint가 httpapi를 재시작 루프에 빠뜨려서는 안 된다(SHALL NOT).
descriptor의 존재 자체를 조사할 수 없는 환경 오류(권한 등)는 기동 거부로 남는다(SHALL).

#### Scenario: 반쪽 잔재 위의 httpapi 기동

- **WHEN** descriptor는 존재하지만 socket이 없거나 죽어 dial이 실패하는 상태에서 httpapi가
  기동하면
- **THEN** 전략 표면 없이 기동하며, 관측 동작은 descriptor-부재 기동과 동일하다
