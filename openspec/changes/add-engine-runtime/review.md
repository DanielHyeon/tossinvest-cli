# 리뷰 기록: add-engine-runtime

## 라운드 1 (2026-07-27, 적대적 리뷰 — 판정 REVISE, P1 4·P2 3·P3 4)

Manager 재검증(2건 코드 확인 — 전부 사실): filldetect Detector.Run/Hints.Run 프로덕션 호출자 0(exitloop.go:146 주석뿐), ReconcileDriver.Run은 ctx 취소 외 반환 없음("실패 사이클은 재시도" 문서화된 결정 — reconcileloop.go:280-285).

### Manager 처분 (2판에 반영)

- **P1 체결 감지 누락**: 채택 — **루프 집합에 filldetect 포함**(측정 무관 배관이므로 이 change의 소관이 맞다; 첫 청산 발의 후 pending 영구 미해소 → exit 수명주기 위반 경로 차단). exit 관측 SLO 양보 지점도 함께 배선
- **P1 루프 반환 의미론**: 채택 — 감독을 2층으로 분리: ① 방어적 종료 계약(비정상 반환 → 전체 정지+critical, 정상 종료 무알림) ② 지속 열화 임계(reconcile·체결 감지 연속 5주기 실패 → critical+ENTRY_BLOCKED·재시도 지속 — landed 재시도 결정과 양립, exit 관측은 60s 계약 유지)
- **P1 단일 인스턴스 부재**: 채택 — 활성 마커+프로세스 확인으로 이중 기동 거부 SHALL, 콘솔 상태·autostart가 동일 기제 재사용, verify와 양방향 배타
- **P1 게이트 OFF 근거 오귀속**: 채택 — 두 사유 분리(OFF = 이 change의 규칙 "기동할 루프 집합 없음" / ON+미충족 = 인터록 소비·항목 열거), 시나리오도 분리
- **P2 verify↔engine 반쪽 배타**: 채택 — 엔진 활성 마커를 verify가 검사하는 양방향으로
- **P2 autostart 설치 주체·시점**: 채택 — **스크립트 준비만, 설치·활성화는 게이트 ON 승인 절차의 §0.7 항목**(read-only soak의 비해당 논거는 엔진에 이전되지 않는다는 지적 타당)
- **P2 콘솔 시나리오 낡은 문구**: 채택 — "상태 변경 목록은 본문 열거 행위뿐, 그 외 신규 라우트는 GET"으로 재작성
- **P3 4건**(wantMutating·stateChanging 양방향 골든 맵, pgrep 패턴 Go 상수+drift, 무관리 알림 도달성 기록): 전건 채택 — proposal Impact에 도달성 기록 추가

## 라운드 2 (2026-07-27 — 판정 REVISE(minor) → 3판 반영)

라운드 1 처분 전건 정확 반영 확인(감독 2층·게이트 OFF 귀속·autostart §0.7 이관은 리뷰어가 "잘 됐다" 명시). 신규 4건 처분:

- **P1 강화 트리거 열거 위반**: 채택 — risk-management MODIFIED delta 신설(열거에 대사·체결 감지 지속 실패 2종 additive, 전부 ENTRY_BLOCKED·HALT_ALL 자동 진입 없음 유지, "분석 비트리거"와의 구분 명시), task 1.3에 ModeTrigger 상수 2종+표 갱신+HALT_ALL 비도달 단언 명시
- **P2 체결 감지 이중성**: 채택 — `filldetect.Detector` 폴링 루프로 지목, Hints 포함 시 Refresh 미배선은 조립 시점 검증 거부 SHALL(방어적 종료 계약의 전제 보존)
- **P2 verify 측 SHALL 소유권**: 채택 — 이 change에서 제거, execution-verification change(2b)의 후속 태스크로 이관 명시(엔진 측 검사만 소유)
- **P2 자문 마커 한계**: 채택 — **배타를 journal 디렉터리 flock으로 격상**(동시 기동 경합까지 폐쇄 — 라운드 1 P1-3 완전 종결), 활성 마커는 자문 신호(갱신 1분·stale 5분 수치 고정)로 역할 분리
- **P3 SLO 어댑터**: 채택 — 엔진 배선 어댑터로 지목(exitloop.go:144-148 landed 의도 인용)

## 라운드 3 (대기 — 지점 재확인)
