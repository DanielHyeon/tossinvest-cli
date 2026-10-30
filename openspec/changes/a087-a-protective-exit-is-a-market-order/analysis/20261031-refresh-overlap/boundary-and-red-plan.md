# a087 refresh 중첩 경계와 RED 계획 — 2026-10-31

기준 소스 `1a153af2c278915d06afc17d921ddb2eb5c54c7e`; main `e32606d4bd44de3142f3c732ab86ecee913e3ab0`은 이 커밋을 두 번째 부모로 포함한다. 본 분석은 설계 제안이며 활성화 승인이나 구현 PASS가 아니다.

## 현재 경계

- **AST** `observer-run.ast.json`: `ExitObserver.Run`은 `ObserveOnce` 반환 뒤 `cycleComplete`를 호출하고 `Sleep`한다(`exitloop.go:414-416`). **손 읽기** `ObserveOnce`의 `observe`는 `Retrier.Query`를 통해 `Prices`를 호출한다(`exitloop.go:863-869`).
- **AST** `board-run.ast.json`: `exitLowerLimitBoard.Run`은 채널에서 interval을 소비한 다음 동기식 `refresh`를 호출한다(`exit_lower_limit.go:166-180`). **AST** `refresh.ast.json`: positions/원장 prefetch 후 `context.WithTimeout`을 새로 만들고 `ProtectivePriceLimits`를 호출한다(`:185-222`). 송신한 시각은 토큰에 없다. 따라서 이전 refresh 처리 중 쌓인 토큰을 다음 관측이 시작될 무렵 처리할 수 있다.
- **손 읽기**: `official.Client.send`는 `doRequest`를 최초 및 401 재전송 각각 호출한다(`client.go:320-354`); `doRequest`는 `c.hc.Do`와 body read를 수행한다(`:191-205`). context 취소나 호출 반환을 transport 내부 wire가 실제로 끝났다는 증거로 등치하지 않는다.
- nil admission에서는 token 송신과 refresh가 정지한다(`exit_lower_limit.go:142-148,162-169`). 현재 생산 배선은 literal nil(`engine.go:598`). 활성 시험 fixture 또는 미래 5.5 배선과 구분한다.

## 분리할 계약

1. **관측 goroutine 중첩**: 관측 cycle이 살아 있는 동안 refresh goroutine이 실행되는 것 자체를 금지하려면 양쪽 cycle 시작·종료를 공유 제어해야 한다. 다음 관측 시작 시 진행 중인 비협조 refresh의 실제 종료를 기다리면 기존 관측 지연 0 계약과 충돌한다.
2. **실제 GET wire 중첩**: 공유 a166 전송 gate가 두 요청을 같은 계좌·프로세스 경계로 묶고 실제 wire 시작부터 실제 종료까지 소유권을 유지한다면, goroutine이 동시에 살아 있어도 두 wire는 중첩되지 않을 수 있다. 이는 a166의 구체 transport lifecycle·취소 후 drain·401 재전송·실패 정책 증명에 조건부다. `Do` 반환만으로 소유권을 해제하면 이 주장은 성립하지 않는다.
3. 직렬화가 관측 GET을 기다리게 하면 보호 판정이 늦어진다. gate가 관측을 우선해도 이미 시작된 하한가 GET을 즉시 종료한다는 실제 transport 보장이 없으면 지연 상한은 증명되지 않는다. 타임스탬프·deadline 검사만으로 검사 뒤 스케줄 지연과 취소 후 잔존 wire를 봉쇄할 수 없다.

## 최소 설계 후보 — Manager 결정 필요

별도 a087 gate를 중복 구현하지 않고, a166이 제공하는 **실제 wire 생명주기 단위** 공통 admission·drain 보장에 두 요청을 결속한다. 관측이 시작된 epoch에 속하지 않거나 만료된 대기 refresh 토큰은 버리고, 다음 관측 시작과 refresh의 wire 시작 경합은 같은 gate에서 원자적으로 판정한다. 다만 관측 진입 시 이미 살아 있는 refresh의 처리(실제 종료 대기, 관측 실패, 또는 취소 후 보장된 유한 drain)는 기존 `design.md:280-285`의 지연 0 및 겹침 0을 동시에 충족할 수 있는지 증거가 필요하다. 미증명 시 production admission nil, 5.5 차단을 유지한다. 설계의 우선순위·지연 허용을 임의로 바꾸지 않는다.

## 재현·RED 계획 (아직 실행하지 않음)

- 가짜 transport에서 refresh wire가 시작되면 barrier로 붙잡고 context 취소를 무시해 **실제 종료까지** 생존시킨다. 다음 관측 `/prices`를 시작시키며 실제 start/end 이벤트와 동시 wire 수를 원자 기록한다. 현 구현은 중첩 > 0으로 RED여야 한다. 테스트는 `httptest`/in-process transport만 사용하고 live 호출이 없다.
- refresh token을 소비 전에 지연시켜 다음 관측 시작 뒤 해제하는 별도 행; 오래된 토큰의 wire 시작 0을 기대한다. 스케줄 경합은 barrier로 재현하고 wall-clock sleep에 의존하지 않는다.
- 최초 401 뒤 재전송 각각에 같은 실제 wire admission이 강제됨과 거절 시 다음 wire 0을 실제 `official.Client`로 검증한다. 별도 a166 착지 SHA와 시험 이름을 5.5에 결속한다.
- 비협조 transport에서 timeout/cancel 이후 gate가 실제 drain 전 release하지 않음을 단언한다. 이때 관측 지연을 어떻게 처리하는지는 Manager 승인된 계약을 따른다; 지연 0을 가짜 context 반환으로 주장하지 않는다.
- production nil의 GET 0, 기존 LIMIT body golden, 보호 주문 제출 경로 GET 0, normal/tagged/race 및 변이 검출은 수정 후 별도 집중 검증한다. 전체 시험·게이트·실계좌·운영 설정 변경은 범위 밖이다.
