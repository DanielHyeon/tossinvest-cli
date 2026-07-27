# Tasks: verify-us-market

> 목적: 2b task 2.5가 요구하는 "시장·유형별" 조건주문 측정을 US에서도 수행할 수 있게 한다.
> 이 change는 측정 도구만 넓힌다 — 엔진의 US 매매 허용은 2c의 2.6과 §0.7 게이트가 정한다.
> KR 경로(기본값·기록·문구·plan digest)는 무접촉이어야 한다.

## 1. run의 시장 [T]

- [x] 1.1 [T] `Options.Market` 신설 — zero value = KR(§0.2). Runner가 보유하고,
  `MarketKR` 상수의 "유일한 시장" 주석을 사실에 맞게 고친다. RED: 미지정 run이 종전과
  동일하게 동작(KR 심볼 통과·US 심볼 skip), `Market: "US"` run은 US 심볼을 통과시킨다.
- [x] 1.2 [T] preflight의 시장 판정을 run 기준으로 — `MarketOf(symbol) != MarketKR`를
  `run의 시장과 다름`으로 교체하고 사유 문구가 두 시장을 이름으로 말한다. RED: US run에서
  KR 심볼 단계가 skip되고 그 반대도 성립.
- [x] 1.3 [T] US 정정은 가격 전용 — `amendOrder`가 US에서 `AmendIntent.Quantity`를 nil로
  보내고 승인 상세 문구도 시장에 맞춘다(브로커: `400 us-modify-quantity-not-supported`).
  RED: US run의 정정 요청에 수량이 실리지 않고, KR run은 종전대로 수량을 싣는다.

## 2. 세션 advisory [T]

- [x] 2.1 [T] `internal/clock`의 시장별 정규장을 사용해 US advisory를 만든다(두 번째 달력
  금지). US 문구는 휴장 응답이 **미측정**임을 명시하고 KR 문구는 실측 422 코드를 계속
  인용한다. 어느 쪽도 시작을 차단하지 않는다. RED: US 정규장 안/밖 판정(DST 경계 포함),
  US 문구에 KR 코드가 등장하지 않음.

## 3. 시장별 증거 기록 [T]

- [x] 3.1 [T] 기본 기록 경로를 시장별로 — KR `capability-verify.jsonl`(무변경),
  US `capability-verify-us.jsonl`. `--record` 명시 값이 우선한다. RED: 시장별 기본 경로,
  KR 기본값 회귀 없음.
- [x] 3.2 [T] 시장별 격리 — 한 시장의 terminal 판정이 다른 시장의 `settled`/`RedoSet`/
  진행률에 영향을 주지 않는다. RED: US 전 단계 통과 후 KR 화면이 여전히 미측정로 보임.

## 4. 배선과 화면 [T]

- [x] 4.1 [T] `cmd/tossctl`: verify·console에 시장 선택(플래그·기본 KR), 보유 종목 선택을
  run의 시장으로, US 프로브 심볼은 사용 가능한 US 보유 종목(플래그로 덮어쓰기 가능).
  RED: US run이 US 보유를 고르고 소수점 보유는 제외한다.
- [x] 4.2 [T] 콘솔 `/verify?market=us` — 진행률·이어하기·재측정·단계 목록·시작 폼이 선택
  시장을 따른다. 승인 화면은 무변경(클릭 1회). 두 시장 연속 측정에는 프로세스 재시작이
  필요함을 안내한다. RED: 시장별 렌더, 시장 미지정은 KR, 폼의 시장 값이 run에 전달됨.

## 5. 완료 게이트 [M]

- [x] 5.1 Function Logic Map(수정 대상 기존 함수) + `check_analysis.py`
- [x] 5.2 PM registry allowlist + PM fixture 등재
- [x] 5.3 `make sdd-sync && make gate CHANGE=verify-us-market`
- [x] 5.4 설치 + 콘솔 재시작 안내, US 세션 측정 절차를 사용자에게 안내
