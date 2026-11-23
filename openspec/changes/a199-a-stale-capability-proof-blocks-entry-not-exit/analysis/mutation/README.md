# a199 변이 원장 (2.7)

- 하네스 `run_mutations.py`, 원문 `run-1.txt`. 사본 = `git worktree add --detach` 로 만든 고정 커밋 `b856fdc08` 의 사본(시작 sha 단언), 끝나면 제거(`copy removed: True`, `git worktree list` 에 잔재 없음).
- 무변이 대조군: 네 패키지(engine · cmd/tossctl · execgw · attest)에서 `-run A199` rc 0 → 시작.
- 닿음: 변이마다 대상 문자열 1회 적중 · 바이트 변화 단언, 원복 뒤 바이트 동일 단언, 끝에 사본 `git status` 깨끗.
- **결과: 20 / 20 CAUGHT, SURVIVED 0, INVALID 0, BUILD 0.**

| 변이 | 잡은 시험 |
|---|---|
| M1 만료 시 계좌 확인 삭제 | RefusalsStayRefusals · TheWatchJudgesAnExpiredForeignAttestationAsTheInterlockDoes |
| M2 닫힌 집합 넓히기(불완전도 낡음으로) — design 의 「ErrMissing 추가」 대신: ErrMissing 은 Load 오류라 판정 함수에 닿지 않음(도달 불가), 같은 축(닫힌 집합 경계)을 ErrIncomplete 로 넓혀 잼 | RefusalsStayRefusals |
| M3 빈 endpoint 항목 가드 삭제 | RefusalsStayRefusals · TheWatchJudges… |
| M4 형식 하한 삭제 | AnInvalidFormatVersionDoesNotGoExitOnly |
| M5 뒤 조항 검사 생략 | Stale…Gateway… · Stale…Transport… · StaleWithAnotherClause… |
| M6 래치 세우기 삭제 | AStaleAttestationStartsExitOnly · ARestartRederives… · AnExitOnlyRuntime… |
| M7 Verified 를 true 로 | AStaleAttestationStartsExitOnly · AnExitOnlyRuntime… |
| M8 Guardian nil 유지 | AStaleAttestationStartsExitOnly · AnExitOnlyRuntime… |
| M9 감사 action 을 수락으로 | AStaleAttestationStartsExitOnly · ARestartRederives… |
| M10 감시 판정을 따로 두기 | TheWatchJudges… |
| M11 worker 조건에서 Verified 삭제 | TheStrategyWorkerStillRequiresAVerifiedGate |
| M12 감시 알림 조건을 파일 판정으로 | AHealthyFileOnAnExitOnlyEngineIsAnnouncedOncePerDay |
| M13 보호 수명 문장 조건에서 청산 전용 삭제 | TheExitOnlyStartSaysWhatItIs |
| M14 · M15 exit 관측 · 대사 생성을 Verified 로 | AStaleAttestationStartsExitOnly(+ 연결 시험) |
| M16 회복 안내 키 보존 삭제(R2-3) | AHealthyFile…OncePerDay |
| M17 낡음 표지 안 세움 | 4 시험 |
| M18 runEngineRun 이 청산 전용을 게이트 OFF 로 | AnExitOnlyContextReachesTheRuntimeAndSaysSo |
| M19 latchOrder 에서 새 사유 빼기 | TheStaleAttestationLatchReadsBeforeTheModeAndAfterStrategyCentral |
| M20 NamesAccount 를 가림 허용 규칙으로 | NamesAccountIsVerifysAccountComparison |
