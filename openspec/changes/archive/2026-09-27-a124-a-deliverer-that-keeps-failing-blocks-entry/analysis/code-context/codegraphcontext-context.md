# CodeGraphContext supporting context — a124

- 날짜: 2026-09-26 · base `4798d399`
- **CGC 0.5.1: not-applicable** — `codegraphcontext find name alertDeliverer` · `analyze callers deliverOne` 둘 다
  `Database Connection Error: IO exception: Could not set lock on file : ~/.codegraphcontext/global/db/kuzudb`
  (다른 프로세스가 kuzu DB 보유 — a114 완료 게이트 기록의 같은 증상). 질의 결과 0 건이며 advisory 이므로 차단하지 않음.
- **GBrain: not-applicable** — `gbrain_project.py code-callers deliverOne` · `search "alert delivery executor attempts limit
  entry gate"` 둘 다 **exit 75** `[gbrain-project] busy: owner pid=1564542 command='gbrain serve'`(다른 세션 소유).
- **파일 기억**: `scripts/memory-recall.sh "alert delivery gate attempts"` → `results: []`.

## 대체 보조 문맥 (HEAD grep · 파일 직접 읽기)

- 게이트 잠금 18 자리(`rg '\.Block\(' internal cmd`, 비테스트 전부): `ReasonAlertUndelivered` 를 거는 자리는
  `notifier.go:280·484·520·571` + `gateway.go:164`(기동 복원) 다섯. 배달 실행자(`alertdelivery.go`)는 0 — proposal 표와 일치.
- 모드 승격 6 자리: `CRITICAL_ALERT_UNDELIVERED` 는 `notifier.go:382` 하나 — proposal 표와 일치.
- `PendingAlerts` 생산 호출자 4: `alertdelivery.go:150` · `alertops.go:117` · `notifier.go:737`(Flush) · `notifier.go:855`(Acknowledge).
- 판정 입력 후보: `MarkAlertAttemptFailed` 는 `SettleResult` 만 돌려준다(`alert_claim.go:140-145` — `Outcome`·`ClaimedBy`·
  `ClaimedAt`·`ExpiresAt`, **attempts 없음**). `SettleOutcome` 의 영값은 `SettleApplied` (`alert_claim.go:108`).
- 재무장은 `attempts = 0` 으로 에피소드를 새로 연다(`outbox.go:334-342`) — 시도 수는 에피소드 단위다.
- 원장 연결은 하나(`journal.go:174` `SetMaxOpenConns(1)`) — 실행자의 추가 쓰기는 exit 루프와 같은 연결을 기다린다.
- 진입 게이트는 노출을 늘리는 변이에만 묻는다(`execgw/gateway.go:855-859` `!plan.raisesExposure` 면 nil) — 청산은 게이트 밖.
- `ModeTriggerCriticalAlertUndelivered` → `ENTRY_BLOCKED` 만(`operating_mode.go:537-545`), 자동 완화 없음(`:416-421`).
- Notifier 는 엔진에서 **무조건** 생성된다(`gateway.go:323`); gate OFF 경계는 `engine run` 의 기동 거부
  (`TestAGateOffEngineRefusesWithoutEnumeratingClauses` rc 0, 2026-09-26).
