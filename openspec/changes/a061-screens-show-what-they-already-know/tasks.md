# a061 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `attachPositionExitLines`의 Function Logic Map과 Branch Test Map을 **편집 전에**
      작성한다. 기존 함수 내부 분기를 바꾸는 유일한 편집이다.
- [x] 1.2 `last_observed_at`의 유일 writer가 `record`이고 `judgeRatchet`/`judgeLadder`가
      `!snapshot.Changed`에서 `record` 없이 반환한다는 것을 현재 HEAD에서 확인해 기록한다.
- [x] 1.3 운영 원장에서 결함을 측정해 기록한다 — 네 EVALUATED 행의 `last_observed_at`
      부동, 042660의 `SEED`, 466100의 활성 quarantine (읽기 전용).
- [x] 1.4 `cmd/tossctl/console.go`가 프로덕션에서 `EngineMarker`를 항상 배선하고
      테스트 하네스는 배선하지 않는다는 것을 확인한다 (D2의 회귀 0 전제).

## 2. RED — 라인 열

- [x] 2.1 엔진 마커가 신선한 상태에서 관측 시각이 몇 시간 전인 canonical snapshot을
      `/positions`가 다섯 값 그대로 렌더한다. 현재 코드에서 **실패**한다.
- [x] 2.2 `/dashboard`가 같은 보유에 같은 답을 준다. 현재 코드에서 **실패**한다.
- [x] 2.3 엔진 마커가 창 밖이면 다섯 값이 `—`이고 `엔진 정지` 사유가 나온다.
- [x] 2.4 활성 quarantine이 걸린 포지션은 엔진이 살아 있어도 `—`이고 `판정 격리`
      사유가 나온다. 현재 코드에서 **실패**한다 (콘솔이 quarantine을 읽지 않는다).
- [x] 2.5 이전 generation의 quarantine은 현재 세대를 막지 않는다.
- [x] 2.6 회귀: 마커 미배선 하네스에서 관측 나이 판정과 문구가 지금과 동일하다.
- [x] 2.7 회귀: `observation_in_future`·`invalid_observed_at`은 엔진이 살아 있어도 닫힌다.
- [x] 2.8 회귀: lifecycle generation 불일치와 RELEASED는 지금과 같이 닫힌다.

## 3. RED — 종목명

- [x] 3.1 `/position-management`가 KR 행에 종목명을 표시한다. 현재 코드에서 **실패**한다.
- [x] 3.2 US 행도 같은 자리에 표시한다. 현재 코드에서 **실패**한다.
- [x] 3.3 보유 캐시가 비어 있으면 이름 없이 렌더되고 브로커 호출 수가 0이다.
- [x] 3.4 시장이 다른 동명 심볼의 이름이 섞이지 않는다.

## 4. GREEN

- [x] 4.1 `ReadOnly`에 현재 generation의 활성 quarantine 조회를 더하고
      `LivePositionExits`가 `PositionExit.Quarantine`을 채운다 (D3).
- [x] 4.2 `joinPositions`가 그것을 행에 옮긴다.
- [x] 4.3 `protectionLiveness`를 엔진 마커에서 읽는다 (D2).
- [x] 4.4 `attachPositionExitLines`가 quarantine → liveness → 무결성 순으로 판정한다 (D1).
- [x] 4.5 `operatorview`에 사유 3건과 사유별 상태 문구를 더한다 (D4).
- [x] 4.6 `handlePositionManagement`가 `peek`로 이름을 붙이고 템플릿이 그린다 (D5).
- [x] 4.7 2장·3장 전부 GREEN.

## 5. REFACTOR

- [x] 5.1 `attachPositionExitLines`의 판정 순서를 주석이 아니라 읽히는 구조로 만든다.
- [x] 5.2 `portfolio_pages.go`와 `design.md`가 "상태는 나이로 썩지 않는다"를 같은 말로
      설명하게 한다.

## 6. VERIFY

- [x] 6.1 변이 검증: liveness 판정을 되돌리면 2.1이 RED가 되는지 확인하고 되돌린다.
- [x] 6.2 변이 검증: quarantine 조회를 제거하면 2.4가 RED가 되는지 확인하고 되돌린다.
- [x] 6.3 변이 검증: generation 조건을 빼면 2.5가 RED가 되는지 확인하고 되돌린다.
- [x] 6.4 `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads`
      green 유지 — 주입 seam이 늘지 않았다.
- [x] 6.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 6.6 `make gate CHANGE=a061-screens-show-what-they-already-know`.
- [ ] 6.7 사람 승인 후 컨테이너 실측 — 다섯 값이 나오는지, 466100이 `판정 격리`로
      나오는지, `/position-management`에 한글 종목명이 나오는지.

## 7. 리뷰와 기록

- [x] 7.1 독립 리뷰를 받고 `review.md`에 기록한다.
- [x] 7.2 발견 사항을 `issues.md`에 남긴다 — httpapi의 같은 결함, 466100 quarantine의
      운영자 결정 필요, 알림 미전달.
- [x] 7.3 PM story/tracker 동기화.
