## codex — 26라운드 2차 수리 재확인 (같은 세션 이어서)

작업 트리 `/tmp/claude-1000/a092-r26-tree` 를 `55b435ea` 의 `git archive` 로 바꿨다(읽기 전용 — 파일을 만들거나 고치지 말 것). 트리 루트의 `R26B-FIX.diff` 가 `d8769cfb..b910173a` 의 `internal` · `cmd` 차이 전부다. 당신의 재확인 표(잔여 #1~#5)와 Claude 보이스 결과에 대해 Manager 가 승인한 범위를 고쳤다:

- 당신의 잔여 #1(R1): `recordFailedAttempt` 가 반납 결과의 선점(AlreadySettled · LeaseLost)도 `EventAlertClaimLost` 로 기록.
- 잔여 #2(R2): `ModeReleaseResult.NoticeReadError`(통지 목록만 실패 — 모드 · 사유 보존), CLI `writeModeReleaseResult` 가 재조회 실패를 먼저 가르고 통지 상태를 추정하지 않음(「이미 전달 처리됨」 제거).
- 잔여 #3(R3): `logLeaseLost` 가 결과 enum 으로 분류(모르는 결과 = `EventAlertUndelivered` 오류).
- 잔여 #4(Flush) · #5(ClaimDisposition): **이월**(Manager 승인 — 생산 호출자 0 · 잠재 결함). 판정 대상 아님.
- 보이스 A #1: `ModeOperations.Release` — 커밋 앞은 요청 ctx, 커밋 뒤 통지(`detachedAnnouncer`)와 재읽기는 `context.WithoutCancel`.
- 보이스 A #4 · B #4: obs 로그 줄의 원래 사건 유형 키를 `trigger_event`(발행 실패 줄 등급은 `trigger_severity`).
- 보이스 B(25라운드 B#7 잔재): `Runtime.escalate` · `RiskGuardian.escalateFor` 가 `ErrModeAnnouncementFailed` 를 「승격됨 · 통지 기록 실패」로.
- 불변식 8 (a): 기록 전용 입구 세 메서드의 로그 줄에서 필드 제외(`withoutFields`), 게이트 설명 · `NotifyError` · `ReReadError` · `NoticeReadError` · 500 본문 고정 문구, 원문 오류는 계좌를 `[account]` 로 가려 엔진 로그에만(`ModeOperations.logFailure`). 원장 내부 사건 키는 불변(frozen 의미론), base 의 계좌 로그 관행은 별도 사람 결정.

판정해 달라:
1. 잔여 #1 · #2 · #3 각각 — 닫힘 / 부분 / 안 닫힘, 코드 인용으로.
2. 이번 diff 가 새로 만든 결함. 특히: (a) `detachedAnnouncer` 가 커밋 **앞**의 무언가를 요청 ctx 에서 떼는가(전이 · audit 가 요청 취소 뒤에도 커밋될 수 있는가) (b) 고정 문구가 운영자에게 필요한 분류를 지우는가(400 과 500 의 구분, 통지 실패와 재조회 실패의 구분) (c) `logFailure` 의 가림이 계좌를 놓치는 경로 (d) 판정(차단 · 승격 · 거절)이 바뀐 자리가 있는가.
3. 같은 부류가 이번 diff 가 만진 파일에 남았는가.

시험은 실행하지 않아도 된다(저자 변이 22/22 CAUGHT 는 `openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/mutation-r26/ledger-r26b.tsv` 에 있음 — 믿지 말고 코드로). 출력은 이전과 같은 표 형식 + 마지막 줄 `Recommendation:`.
