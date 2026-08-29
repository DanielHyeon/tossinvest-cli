# Tasks — interlock-gates-entry-not-exit

## 1. 계약

- [x] 1.1 proposal·design 작성, base-commit 고정
- [x] 1.2 engine-safety 스펙 델타 — 조항 6을 기동 거부에서 진입 허가로, 집행 지점 명시
- [x] 1.3 `openspec validate --strict` 통과

## 2. RED

- [x] 2.1 조항 6 미충족 + 1~8 충족에서 인터록이 `Verified = true`를 반환한다
- [x] 2.2 같은 상태에서 `EntryPermitted = false`이고 `Protection`이 `UNWIRED`로 보고된다
- [x] 2.3 조항 1~8 각각의 단독 미충족이 여전히 기동을 거부한다 (8건)
- [x] 2.4 보호 미배선에서 `Gateway.Place`가 buy intent를 거부한다
- [x] 2.5 같은 상태에서 sell intent는 통과한다
- [x] 2.6 구조 단언 — `engine run` 도달 그래프에 진입 발급이 없다
- [x] 2.7 RED 관측 기록 (review.md)

## 3. GREEN

- [x] 3.1 `ProtectionReadiness`·`ProfileProtection` 상수를 `internal/execgw`로 이동
- [x] 3.2 `Gateway.Place` — `raisesExposure` + 보호 미배선 시 거부, 사유 열거
- [x] 3.3 `interlock.go` — 조항 6을 `verifyGate`에서 제거, `AutomationStatus.EntryPermitted` 추가
- [x] 3.4 기동 출력 한 줄 — 보호가 프로세스 수명에 묶여 있다는 사실 (design D6)
- [x] 3.5 전체 테스트 GREEN

## 4. 증거·게이트

- [x] 4.1 Function Logic Map + Branch Test Map — 변경한 기존 함수 전부
- [x] 4.2 review.md — 적대적 리뷰
- [x] 4.3 `make sdd-sync` → `make gate CHANGE=interlock-gates-entry-not-exit`

## 인계 (이 change의 태스크가 아니다)

아래는 §0.7이 사람에게 남긴 것이며, 체크박스로 두지 않는다 — 이 change의 완료 조건에
사람의 운영 결정을 넣으면 게이트가 일어나지 않은 일을 통과시킨다.

1. 바이너리 설치 — `install -m755 <build> ~/.local/bin/tossctl`
2. `trading.sell`·`allow_live_order_actions` 승인 여부 결정. 인터록 조항 3이 둘 다
   요구하며, 이는 "엔진이 내 주식을 팔아도 된다"는 승인이다.
3. `engine.automation_gate.enabled` flip.
4. 엔진 기동 후 관측할 것:
   - **편입이 즉시 일어난다** (review.md A5). 대사 루프의 마지막 단계이고 되돌릴 수 없다.
     `adoption.enabled = true`이므로 `exclude_symbols`에 없는 무기록 보유가 전부 편입된다.
   - `engine.operating_mode` 기록의 `entry_permitted: false`와 그 옆 문장.
   - 편입된 보유의 손절선이 `default_stop_pct`(현재 0.05)로 잡히는지.
