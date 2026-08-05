# a085 · Tasks

## 1. 근거 고정

- [ ] 1.1 알림 조립부 14곳을 열거하고 각각이 종목을 가리키는지 분류한다.
- [ ] 1.2 CodeGraph로 `reconcile.Holding`의 사용처를 확인하고 비교 경로가 이름을
      읽지 않음을 기록한다.
- [ ] 1.3 `operatorview.BuildExitLine`의 호출부와 stale 사유 목록을 확인한다.
- [ ] 1.4 Function Logic Map: `BuildExitLine`은 기존 함수 내부 분기를 바꾸므로
      작성한다. `Collector.holdings`는 필드 추가만이므로 함께 판단해 기록한다.

## 2. RED — 이름 배선

- [ ] 2.1 `official.RawHolding`에 `Name`이 담기고 `HoldingsRaw`가 채운다. **실패한다.**
- [ ] 2.2 `reconcile.RawHolding`·`reconcile.Holding`까지 이름이 전달된다. **실패한다.**
- [ ] 2.3 회귀: 이름이 다른 두 holding의 수량 비교 결과는 같다 (비교 오염 없음).
- [ ] 2.4 회귀: 이름이 비어도 스냅샷·비교·편입이 지금과 동일하게 동작한다.

## 3. RED — registry

- [ ] 3.1 대사 주기가 registry를 갱신한다.
- [ ] 3.2 보유가 사라져도 이름이 남는다 (청산 직후 알림이 이름을 잃지 않는다).
- [ ] 3.3 registry에 없는 심볼은 코드만 쓰고 추가 브로커 요청을 하지 않는다.

## 4. RED — 알림 문구

- [ ] 4.1 종목을 가리키는 알림 제목·본문이 한국어이고 `이름(코드)` 형식이다. **실패한다.**
- [ ] 4.2 이름을 모르면 코드만 쓴다.
- [ ] 4.3 회귀: `Fields` payload의 키와 값이 영문·원문 그대로다.
- [ ] 4.4 회귀: 구조화 로그 필드가 바뀌지 않는다.
- [ ] 4.5 회귀: 알림에 계좌번호·잔고·세션이 들어가지 않는다.
- [ ] 4.6 원문 에러를 포함하는 알림은 한국어 설명 뒤에 원문을 그대로 붙인다.

## 5. RED — operatorview

- [ ] 5.1 `snapshot_quarantined`의 exit line이 저장된 진입가·최초 손절·보호선·
      워터마크를 채우고 stale로 남는다. **실패한다.**
- [ ] 5.2 회귀: `engine_not_running`·`observation_older_than_limit` 등은 지금과 같이
      값을 채우지 않는다.
- [ ] 5.3 회귀: unknown(스냅샷 없음)은 지금과 같이 전부 `—`다.

## 6. RED — console

- [ ] 6.1 격리된 포지션 행의 라인 5칸과 상세에 저장값이 그려진다. **실패한다.**
- [ ] 6.2 그 값 옆에 갱신되지 않는 저장값이라는 표시가 있다.
- [ ] 6.3 "판정 대상이 아니다 — 손절도 평가되지 않는다" 경고가 유지된다.
- [ ] 6.4 포지션 화면의 보호선이 포지션 관리 화면의 유지 보호선과 같다.

## 7. GREEN

- [ ] 7.1 `Name` 필드를 official → reconcile까지 배선한다.
- [ ] 7.2 엔진 registry와 `이름(코드)` 포맷터를 추가한다.
- [ ] 7.3 알림 14곳의 제목·본문을 한국어로 바꾸고 포맷터를 쓴다.
- [ ] 7.4 `BuildExitLine`의 격리 분기가 값을 채우게 한다 (D6).
- [ ] 7.5 템플릿이 채워진 stale을 구별해 그린다.
- [ ] 7.6 2~6장 전부 GREEN.

## 8. VERIFY

- [ ] 8.1 §0.4: 이 change가 브로커 요청 수를 늘리지 않음을 테스트로 고정한다.
- [ ] 8.2 PLTR의 실제 저장값으로 포지션 화면 렌더를 확인한다 (읽기 전용 fixture).
- [ ] 8.3 실제 alert_outbox 사례("the exit policy could not judge 032820")가
      한국어 + 이름(코드)로 렌더되는지 확인한다.
- [ ] 8.4 upstream 상속 테스트 회귀 없음 (650 green).
- [ ] 8.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [ ] 8.6 `make gate CHANGE=a085-an-alert-says-which-stock`.

## 9. 리뷰와 기록

- [ ] 9.1 경량 리뷰(validate + Manager 셀프리뷰)를 `review.md`에 기록한다.
- [ ] 9.2 `issues.md`에 남긴다 — 대사 차단 evidence 문자열이 첫 관측에 고정되어
      화면이 오래된 숫자를 현재값처럼 보여주는 문제(a083 issues와 동일 건).
