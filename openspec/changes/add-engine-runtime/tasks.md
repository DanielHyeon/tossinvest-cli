# Tasks: add-engine-runtime

> 측정 무관 — 2b 사용자 측정·2c와 병행 가능. 실전 기동은 이 change로 열리지 않는다(인터록 조항 6이 2c까지 상수 미충족 — 라이브 경로는 fake 인터록 테스트로만 검증).

## 1. 런타임 [T]

- [ ] 1.1 `engine.Context` 프로덕션 조립(cmd 측): config → journal Open(RW) → 브로커(구조적 official-only 가드 유지) → obs publisher → `AutomationStatus` 인터록 검증 — 기존 `verifyGate` 경로 소비, 새 검증 로직 금지
- [ ] 1.2 `tossctl engine run`: 인터록 미충족(게이트 OFF 포함) 시 미충족 항목 열거·기동 거부·실패 종료 코드, verify runlock 신선 시 거부, 통과 시 reconcile driver + exit observer 기동. help convention `mutating=true` 표기(화이트리스트 사유 주석)
- [ ] 1.3 루프 감독: 하나 종료 → 전체 취소 + critical 알림 + 실패 종료(부분 생존 금지 테스트 — exit observer만 죽는 시나리오 포함), SIGINT/SIGTERM graceful(취소·완주 대기·journal close), 2번째 시그널 즉시 종료
- [ ] 1.4 재기동 복구 소비 테스트: pending 복원·편입 완결이 run 경로에서 실제 호출되는지(landed 계약 재사용 — 새 복구 코드 금지)

## 2. 무인 운용 표면 [T]

- [ ] 2.1 콘솔: 대시보드 엔진 상태(실행 여부·기동 거부 사유·바이너리 stale 경고) + [엔진 시작]/[엔진 정지] 버튼(세션+CSRF — soak 재시작과 동일 모델, 정지는 시그널 종료 규율 경유). 인터록 미충족 시 거부 사유 표시(콘솔이 인터록을 우회할 수 없음을 테스트로 고정)
- [ ] 2.2 부팅 자동 시작: engine-autostart 스크립트(soak-autostart 패턴 — 게이트 OFF·인터록 미충족이면 대기·재시도 로그만, 크래시 자동 재시작은 횟수 상한 + critical 알림 유지), 사용자 머신 설치는 Manager가 수행

## 3. 완료 게이트 [M]

- [ ] 3.1 §0 검토 기록(새 mutation 경로 0·인터록 소비·부분 생존 금지 방향) + 테스트 전수(-race) + `openspec validate add-engine-runtime --strict`
- [ ] 3.2 `make gate CHANGE=add-engine-runtime`
