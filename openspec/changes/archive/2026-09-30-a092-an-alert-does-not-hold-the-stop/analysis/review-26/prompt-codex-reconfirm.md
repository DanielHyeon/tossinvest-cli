## codex — 26라운드 수리 재확인 (같은 세션 이어서)

작업 트리 `/tmp/claude-1000/a092-r26-tree` 를 이제 `d8769cfb` 의 `git archive` 로 바꿨다(읽기 전용, 파일을 만들거나 고치지 말 것). 당신의 26라운드 표 이후 커밋 둘:

- `071e67c1` — 당신의 #2 · #3 · #5: `NormalRelay` 에 `mu`/`stopped`, `Run` 의 defer 가 처리 중 알림 · 잔여 큐를 버림으로 기록하고 멈춤 표시, 멈춘 뒤 `Offer` 는 즉시 버림 기록, 발행 실패 · 발행기 없음은 유형 + 키로 `EventNormalAlertDropped`; `ModeOperations.Release` 는 재조회 실패를 `ReReadError` 칸에 싣고 커밋 사실(`Changed` · `TransitionID` · `NotifyError`)을 보존하며 오류를 돌려주지 않음.
- `d8769cfb` — 당신의 P0 · #4 · #6: `alertDeliverer.release` 가 결과를 돌려주고 `recordFailedAttempt` 가 행 없음 · 모르는 결과를 원칙 E 조건부 차단(승격 없음)으로 분류, 이미 도달한 한도 판정은 그 뒤에 그대로 섬; 선점(이미 정산됨 · 임차 상실)은 `EventAlertClaimLost` 로 기록; 동기 `deliver` 의 시도 기록 · 반납 두 자리가 `obs.isPreemption` 한 판정을 공유. `Flush` 는 수리하지 않고 K19 핀(`TestA092TheNotifierFlushHasNoProductionCaller`, 생산 호출자 0)으로 둠.

트리 루트의 `R26-FIX.diff` 가 `15cb8540..d8769cfb` 의 `internal` · `cmd` 차이 전부다.

판정해 달라:
1. 당신의 P0 · #2 · #3 · #4 · #5 · #6 각각 — 닫힘 / 부분 / 안 닫힘, 코드 인용으로.
2. 수리가 새로 만든 결함(특히: 반납 행 없음 차단과 시도 한도 판정이 같은 사유에서 서로를 지우거나 순서가 바뀌는가; `isPreemption` 이 행 없음 외의 결과를 선점으로 잘못 가르는가; relay 의 `mu` 가 전송 위에서 쥐어지는가).
3. 같은 부류(결과 분류가 선점/실패를 잘못 가르는 자리)가 `internal/` 에 남았는가.

시험은 실행하지 않아도 된다(수리 쪽 변이 11/11 CAUGHT 는 저자 원장 `analysis/mutation-r26/ledger.tsv` 에 있으나 트리에는 없다 — 믿지 말고 코드로 판정). 출력은 26라운드와 같은 표 형식 + 마지막 줄 `Recommendation:`.
