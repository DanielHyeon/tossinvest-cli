## MODIFIED Requirements

### Requirement: 거래 이력 가시성
콘솔은 완결된 왕복 거래(trade_outcomes)와 exit 이벤트(exit_events)를 시간순으로 표시해야 한다(SHALL). 각 행의 종목은 journal의 market+symbol을 공식 종목 메타데이터와 batch로 보강해 `심볼 · 종목명`을 함께 표시해야 한다(SHALL). 종목명 조회가 실패하거나 해당 심볼의 메타데이터가 없으면 journal 행과 심볼은 그대로 표시하고 종목명을 만들지 않아야 한다(SHALL NOT). 표시는 journal에 동결된 값과 명시 조인(positions의 심볼, exit_states의 진입가)만 사용하며 종목명 보강을 포함해 fills 재계산을 하지 않는다(SHALL NOT). 스키마에 없는 값(청산가 등)은 표시하지 않고(SHALL NOT), nullable 필드(보유 시간 등)의 NULL은 "—"로 렌더한다(SHALL). 성과 행이 생기지 않는 종결(외부 매도로 닫힌 포지션 — adopt-external-positions design A7)은 이 화면의 한계로 명시하고 exit_events 표시가 그 공백을 보완한다(SHALL 명시).

종목명 조회는 `(market, symbol)`을 정규화해 중복 제거하고 한 logical lookup당 400개로 제한해야 하며(SHALL), 첫 live-data 화면에서 계좌를 한 번 확인해 생성한 공식 client와 OAuth token manager를 history 및 다른 account read 화면이 공유해야 한다(SHALL). 공식 endpoint의 200-symbol 한도에 맞춰 chunk하고 client 생성부터 모든 chunk까지 하나의 10초 total timeout을 적용해야 한다(SHALL). 결과는 24시간 TTL·최대 2048개 bounded cache에 저장하고 같은 키의 동시 요청은 single-flight로 직렬화해야 한다(SHALL). 실패하거나 일부 key가 누락된 lookup은 1분 동안 다시 시도하지 않아 공식 API 장애·429 때 새로고침이 rate budget을 증폭하지 않아야 하며(SHALL NOT), 일부 응답에서는 검증된 이름만 장기 cache하고 누락 key는 backoff 뒤 다시 요청해야 한다(SHALL). 실계좌 검증 중에는 새 메타데이터 요청을 보내지 않고 cached name 또는 symbol-only로 표시해야 한다(SHALL NOT). 이를 위해 metadata는 active profile journal directory의 cross-process rate-budget lease를 non-blocking으로 얻어야 하고(SHALL), CLI verify run·console verify·verify abort는 기존 execution exclusion을 얻고 profile run-intent marker를 게시한 뒤 broker 생성 전에 같은 lease를 얻어 작업 종료까지 보유해야 한다(SHALL). active verifier가 있으면 verify abort는 즉시 거부되어 같은 marker를 공동 소유해서는 안 되며(SHALL NOT), 배타 admission 뒤 evidence record를 다시 읽어 최신 취소 대상만 사용해야 한다(SHALL). `--record` override가 profile rate-budget lease 위치를 바꾸어서는 안 된다(SHALL NOT). 원격 name은 길이·control/bidi 문자를 검증한 plain string으로 `html/template`의 escaping을 거쳐야 하며(SHALL), 모호하거나 상충하는 결과를 다른 market/symbol에 붙여서는 안 된다(SHALL NOT). `/history`는 GET/HEAD만 허용하고 다른 method는 메타데이터 요청 전에 405로 거부해야 한다(SHALL).

#### Scenario: 동결된 왕복 결과와 종목명 표시
- **WHEN** KR 또는 US trade_outcomes 행이 있고 공식 종목 메타데이터 조회가 성공한 상태에서 이력 화면을 열면
- **THEN** 각 왕복은 `심볼 · 종목명`과 비용 차감 실현손익·실현 R·초기 수량·보유 시간(NULL은 "—")·도달 exit 단계·청산 시각을 표시하며 동결 성과 값은 바뀌지 않는다

#### Scenario: exit 이벤트의 동일한 종목 라벨
- **WHEN** KR 또는 US exit_events 행의 종목명이 조회되면
- **THEN** exit 이벤트 표도 완결 왕복 표와 같은 `심볼 · 종목명` 형식으로 표시한다

#### Scenario: 종목명 조회 실패
- **WHEN** 공식 종목 메타데이터 조회가 실패하거나 특정 심볼의 이름이 비어 있으면
- **THEN** 해당 journal 행과 심볼은 계속 표시되고 화면은 조회 실패를 이름 부재로 위장하거나 종목명을 추측하지 않는다

#### Scenario: 검증 중 또는 잘못된 method
- **WHEN** 실계좌 검증이 진행 중이거나 인증된 사용자가 `/history`에 POST를 보내면
- **THEN** 공식 종목 메타데이터 요청은 0건이고 각각 symbol-only 안내 또는 405 응답을 반환한다
