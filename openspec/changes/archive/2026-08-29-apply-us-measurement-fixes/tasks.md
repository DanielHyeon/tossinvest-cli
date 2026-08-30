# Tasks: apply-us-measurement-fixes

- [x] 1.1 [T] 취소 유한 재시도 — `409 already-processing`에 한해 최대 2회 추가 시도,
  대기는 `data.retryAfterSeconds`(상한 5초), 주입된 clock seam 사용. 재시도 사실·횟수를
  관측에 기록. 접수·정정·조건주문 생성은 무변경. RED: 409를 1회 반환 후 성공하는 fake에서
  기존 코드가 fail을 남김 → GREEN, 상한 소진 시 fail 유지, 다른 오류는 즉시 실패.
- [x] 1.2 [T] 조건주문 목록 필터 `WATCHING` → `OPEN`(브로커 허용값). 관측 의미 보존.
  RED: fake가 OPEN/CLOSED 외 값을 400으로 거절할 때 기존 코드 실패 재현 → GREEN.
- [x] 2.1 Function Logic Map + `check_analysis.py`
- [x] 2.2 PM registry allowlist + fixture, `make sdd-sync && make gate`
- [x] 1.3 콘솔 기동 배너의 낡은 문구 교정 — "needs the typed confirmation string"은
  console-click-approval 이후 사실이 아니다(화면·터미널 문구가 실제 승인 방식과 일치해야
  한다는 같은 요구의 잔여분).
- [x] 1.4 [T] `order.place.ok`·`order.amend.ok` 관측 detail을 시장에 연동. 요청은
  `amendOrder`에서 이미 시장별로 갈리는데(US는 quantity 미전송) 기록 문구만 "KR
  price+quantity amend"로 고정돼 있어 US 실행이 **보내지 않은 수량을 보냈다고 기록**했다.
  같은 요구(기록·문구가 실제 동작과 일치)의 잔여분이며 2c가 읽을 증거의 정확성 문제다.
  RED: US 실행의 detail이 "KR"을 담음 → GREEN(시장명 + price-only/price+quantity),
  KR 실행은 종전 의미 유지.
