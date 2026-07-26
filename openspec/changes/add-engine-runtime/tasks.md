# Tasks: add-engine-runtime

> 측정 무관 — 2b 사용자 측정·2c와 병행 가능. 실전 기동은 이 change로 열리지 않는다(인터록 조항 6이 2c까지 상수 미충족 — 라이브 경로는 fake 인터록 테스트로만 검증).

## 1. 런타임 [T]

- [ ] 1.1 `engine.Context` 프로덕션 조립(cmd 측): config → journal Open(RW) → 브로커(구조적 official-only 가드 유지) → obs publisher → `AutomationStatus` 인터록 검증 — 기존 `verifyGate` 경로 소비, 새 검증 로직 금지
- [ ] 1.2 `tossctl engine run`: 게이트 OFF → "기동할 루프 집합 없음" 거부(이 change의 규칙), 게이트 ON+미충족 → 인터록 미충족 항목 열거 거부, verify runlock 신선 → 거부, **실행 중 인스턴스 검출 → 거부**(활성 마커+프로세스 확인 — 패턴·상수는 Go 측 seam + drift 테스트, soak 선례), 통과 시 reconcile driver + exit observer + **체결 감지 루프** 기동(exit 관측 SLO 양보 지점을 체결 감지에 배선). help convention `mutating=true` — **골든 맵 wantMutating에도 추가**(양방향)
- [ ] 1.3 감독 2층: ① 비정상 반환(ctx 취소 외) → 전체 취소+critical+실패 종료, 정상 종료(SIGTERM graceful)는 critical 없음 ② reconcile driver·체결 감지 연속 5주기 실패 → critical+ENTRY_BLOCKED 강화·재시도 지속(exit 관측은 landed 60s 계약 유지). SIGINT/SIGTERM graceful(취소·완주 대기·journal close), 2번째 시그널 즉시 종료. 엔진 활성 마커 수명(기동 시 생성·주기 갱신·종료 시 제거·stale 무시)과 verify 측 검사 배선
- [ ] 1.4 재기동 복구 소비 테스트: pending 복원·편입 완결이 run 경로에서 실제 호출되는지(landed 계약 재사용 — 새 복구 코드 금지)

## 2. 무인 운용 표면 [T]

- [ ] 2.1 콘솔: 대시보드 엔진 상태(활성 마커 재사용 — 실행 여부·기동 거부 사유·바이너리 stale 경고) + [엔진 시작]/[엔진 정지] 버튼(세션+CSRF, 정지는 시그널 종료 규율 경유). 인터록 미충족 시 거부 사유 표시(콘솔이 인터록을 우회할 수 없음을 테스트로 고정). **stateChanging 골든 맵에 /engine/start·/engine/stop 추가**(양방향)
- [ ] 2.2 부팅 자동 시작 **스크립트 준비만**: tools/ 등 저장소 내 위치, pgrep 패턴·재시작 상한은 Go 상수 + drift 테스트(soak 선례 — 스크립트와 상수는 한 기제의 두 반쪽), 크래시 재시작 횟수 상한 + critical 유지. **사용자 머신 설치·활성화는 게이트 ON 승인 절차(§0.7)의 항목 — 이 change에서 하지 않는다**

## 3. 완료 게이트 [M]

- [ ] 3.1 §0 검토 기록(새 mutation 경로 0·인터록 소비·부분 생존 금지 방향) + 테스트 전수(-race) + `openspec validate add-engine-runtime --strict`
- [ ] 3.2 `make gate CHANGE=add-engine-runtime`
