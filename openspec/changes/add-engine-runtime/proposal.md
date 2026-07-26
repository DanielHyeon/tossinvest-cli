# Change: add-engine-runtime

## Why

엔진의 모든 루프가 구현·테스트되어 있으나 **프로덕션 호출자가 없다**: `ReconcileDriver.Run`(대사·편입)·`ExitObserver.Run`(exit 관측)을 구동하는 명령이 `cmd/`에 존재하지 않고, `engine.Context`를 조립하는 경로도 없다. 이대로면 2c(보호주문)가 측정 의존적인 보호 계약과 측정 무관한 런타임 배관을 한 change에서 짊어진다. 배관은 브로커 실측이 필요 없으므로 분리해 먼저 landed한다(사용자 질문 2026-07-27 — "2c는 지금 할 수 없는가"의 측정 무관 절반).

## What Changes

- **`tossctl engine run`** 신설: config 로드 → journal open(RW) → 브로커 조립(구조적 official-only 유지) → **기동 인터록 검증**(기존 engine-safety 요구사항의 소비자 — 게이트 OFF 또는 조항 미충족이면 미충족 항목을 열거하며 기동 거부, fail-closed) → 루프 기동(**reconcile driver 60s·exit observer 5s·체결 감지** — 리뷰 라운드 1: filldetect 루프도 프로덕션 호출자 0이며, 체결 감지 없는 런타임은 첫 청산 발의 후 pending이 영구 미해소가 된다) → graceful shutdown. 게이트 OFF 거부는 이 change가 정의하는 규칙(인터록은 게이트 ON에만 정의 — 근거 분리), **단일 인스턴스 보증**(활성 마커·이중 기동 거부·verify와 양방향 배타) 포함.
- **감독 2층**(리뷰 라운드 1 — landed 루프는 ctx 취소 외 반환이 없다): ① 방어적 종료 계약(비정상 반환 → 전체 정지+critical, 정상 종료는 무알림) ② 지속 열화 임계(reconcile·체결 감지 연속 5주기 실패 → critical+ENTRY_BLOCKED 강화, 루프는 재시도 지속 — "실패 사이클은 다음 주기 재시도" landed 결정과 양립). 부분 생존 금지 원칙은 두 층의 합으로 구현된다.
- **종료 규율**: SIGINT/SIGTERM → 루프 취소·완주 대기 → journal close. 두 번째 시그널은 즉시 종료. 재기동 시 crash 복구는 landed 계약(pending 복원·편입 완결) 소비.
- **verify 병행 거부**: verify runlock이 신선하면 기동 거부 — 검증과 엔진이 같은 계좌·rate 예산을 다투지 않는다.
- **무인 운용 표면**(사용자 결정 2026-07-27 — 터미널 조작 제거 원칙): ① 부팅 자동 시작 — 이 change는 **스크립트를 저장소에 준비만** 한다(soak-autostart 패턴 + 크래시 재시작 횟수 상한·critical 유지). **사용자 머신 설치·활성화는 게이트 ON 승인 절차의 §0.7 항목이다** — "부팅마다 주문 능력 프로세스 자동 기동"은 read-only soak과 달리 운영 구성 변경이므로 사람이 게이트 flip과 함께 명시 승인한다(리뷰 라운드 1) ② 콘솔 대시보드에 엔진 상태 표시 + [엔진 시작]/[엔진 정지] 버튼(세션+CSRF — soak 재시작과 동일 모델). **게이트 flip은 여전히 콘솔 밖**(§0.7 사람 승인 절차 — 버튼은 이미 승인된 설정의 프로세스 기동/정지만 한다).
- 실효: **이 change가 landed되어도 실전 기동은 불가능한 상태가 유지된다** — 인터록 조항 6(ProtectionReady)이 2c 전까지 상수 미충족이므로 `engine run`은 거부만 한다. 라이브 동작 검증은 fake 인터록 통과 하에 테스트로만.

## Non-Goals

- 보호주문·ProtectionReady flip(2c 소관), 게이트 ON 절차 변경(§0.7 그대로), 새 mutation 경로(루프·게이트웨이 기존 레일만), Tracer 자동 구동(별도 운영자 절차 유지)

## Capabilities

### New Capabilities

(없음)

### Modified Capabilities

- `engine-safety`: 엔진 런타임 수명주기 요구사항 추가(기동 인터록 요구사항은 무변경 — 이 change는 그 소비자)
- `operator-console`: 상태변경 행위 집합에 엔진 프로세스 시작/정지 추가(게이트 flip은 여전히 부재)

## Impact

- Affected code: `cmd/tossctl`(engine run — mutating=true), `internal/app/engine`(수명주기 조립·감독 — 루프 무변경), `internal/console`(엔진 상태·버튼), 사용자 머신 autostart 스크립트
- 2c와의 관계: 2c는 보호주문 계약·ProtectionReady flip·attestation endpoint 확장에 집중하게 된다. 이 change의 gate는 2c와 무관하게 통과 가능
- §0 검토: 새 주문 경로 0, 기동은 인터록 뒤에만, 감독 2층은 보호 강화 방향, verify↔엔진 양방향 배타, autostart 활성화는 §0.7 승인 항목
- 기록: 편입의 무관리 보유 알림(enabled 무관 존치)은 발화 지점이 reconcile driver 안이므로 실제로는 이 런타임이 게이트 ON으로 돌 때에만 도달 가능하다 — 이 change가 그 사실을 처음 관측 가능하게 만든다
