# a193 변이 원장

- 기준: 구현 커밋 `839a95c90`. 하네스 `harness.py.txt`(이 디렉터리 — 커밋된 원본, `python3 -I harness.py.txt <worktree> <sha>` 로 재실행). 원문 `run-1.txt`(잘리지 않음, 끝의 JSON 이 원장).
- 방법: 사본 = `git worktree add --detach ~/tossos-work/a193-mut-<pid> <sha>`(한 판, 끝나면 `worktree remove` — 실행 뒤 `git worktree list` 에 없음 확인). 시작 sha 단언 ·
  **무변이 대조군 2개 GREEN**(엔진 `-run TestA193`, cmd 선택 시험) 뒤 진행 · 변이마다 앵커 정확히 1회 + `git diff --stat` 비어 있지 않음(적용 = 닿음의 1차 대조) ·
  판마다 `git checkout -- <파일>` 원복 뒤 `status --porcelain` 빈 값 단언 · 빌드 실패는 BUILD_FAIL(CAUGHT 아님) · 시험 시한 120초 초과는 CAUGHT(timeout)로 따로 표시.
- `GOFLAGS=-trimpath`, 전용 `GOCACHE=~/.cache/a193-gocache`.

## 결과: 26 / 26 CAUGHT (실패 24 · 시한 초과 2), SURVIVED 0, BUILD_FAIL 0, NOT_REACHED 0

| id | 변이 | 잡은 시험(일부) |
|---|---|---|
| M01 | IN_DOUBT 거름 삭제 | `TestA193OnlyInDoubtPlaceAttemptsAreTouched` |
| M02 | PLACE 거름 삭제 | `TestA193OnlyInDoubtPlaceAttemptsAreTouched` |
| M03 | park 원장 쓰기를 `context.WithoutCancel`(execgw/indoubt.go — design D3 의 「우연이 지키는 안전」) | `TestA193ACancelledResolutionLeavesTheAttemptInDoubt` 외 |
| M04 | resolve 의 ctx 검사 삭제 | `TestA193ACancelledResolutionLeavesTheAttemptInDoubt` |
| M05 | Run 머리 ctx 검사 삭제 | `TestA193ACancelledRunTouchesNothing` |
| M06 | 원장 읽기 오류의 ctx 검사 삭제 | `TestA193ACancelledResolutionLeavesTheAttemptInDoubt`(취소 스윕 하위 시험) |
| M07 | record 의 ctx 검사 삭제 | `TestA193ACancelledContextIsNeverRecorded` |
| M08 | 10분 → 5분 | `TestA193AnAttemptInDoubtPastTheBoundAlertsOnce` · `…UnreadableLedger…` |
| M09 | stale 경계 `>=` → `>` | `TestA193AnAttemptInDoubtPastTheBoundAlertsOnce` · `TestA193StaleFallsBackToTheRecordedTime` |
| M10 | stale 기준을 기록 시각으로 | `TestA193StaleIsMeasuredFromDispatch` |
| M11 | 알림 창 0 → 1시간 | `TestA193AnAttemptInDoubtPastTheBoundAlertsOnce` |
| M12 | stale key 에서 attempt id 삭제 | `TestA193AFailedStaleRecordIsRetried` 외 |
| M13 | 기록 실패에도 래치 | `TestA193AFailedStaleRecordIsRetried` · `TestA193FirstSightIsAlertedOnce` |
| M14 | 처음 봄 래치 삭제 | `TestA193FirstSightIsAlertedOnce` |
| M15 | OnStop 삭제 | `TestA193ASettlerStopRecordsACritical` · `TestA193AnUnwiredSettlerStopsLoudly` |
| M16 | 직렬 → 동시(`go s.resolve`) | `TestA193ACancelledResolutionLeavesTheAttemptInDoubt` 외 |
| M17 | park 뒤 중복 알림(무장 exit 거름 삭제) | `TestA193AParkOfAnArmedExitProposalIsNotAlertedTwice` 외 |
| M18 | 발의 목록 읽기 실패면 침묵 | `TestA193AnArmedProposalReadFailureStillAlerts` |
| M19 | 원장 못 읽음 경계 `>=` → `>` | `TestA193AnUnreadableLedgerAlertsAfterTenMinutes` |
| M20 | 성공 시 연속 초기화 삭제 | `TestA193AnUnreadableLedgerAlertsAfterTenMinutes` |
| M21 | 재생 진입점 참조 추가 | `TestA193TheSettlerNeverReplays` |
| M22 | 알림에 수량 필드 | `TestA193TheStaleAlertCarriesNoAccountQuantityOrPrice` |
| M23 | `engineRuntime` 원소 삭제 | `TestA193TheProductionRuntimeStartsTheSettlerAsAuxiliary` · `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| M24 | 실행자를 감독 루프로 옮김 | 같은 둘 |
| M25 | 해소 호출에 `context.WithoutCancel` | 시한 초과 — 취소 시험이 해소의 가짜 시계 대기에서 멈춤(취소가 해소에 안 닿음 = 결과가 달라짐) |
| M26 | 배선 누락이면 조용히 쉼 | 시한 초과 — `TestA193AnUnwiredSettlerStopsLoudly` 가 반환을 기다림 |

## 원장이 요구해 더한 시험(VERIFY 단계)

첫 GREEN 뒤 변이 계획을 세우다가, 시험이 없으면 살아남을 자리 셋을 찾아 시험을 먼저 더했다(그 뒤 이 판을 돌림): `TestA193StaleIsMeasuredFromDispatch`(M10 — 픽스처는 기록 시각과
전송 시각이 같아 기준을 바꿔도 통과했을 것), `TestA193ACancelledRunTouchesNothing`(M05), `TestA193ACancelledContextIsNeverRecorded`(M07). 같은 과정에서 **중복 ctx 검사 둘을
지웠다** — 1단 머리(기록은 `record` 가 막음)와 2단 머리(취소면 `resolve` 가 ctx 오류를 돌려줌). 남겼다면 서로를 가려 「동등」으로 살아남았을 자리다.

## 정본 design §5 와의 대응

m1↔M01, m2↔M02, m3↔M03, m4↔M04, m5↔M08, m6↔M10, m7↔M11, m8↔M12, m9↔M13, m10↔M15, m11↔M23, m12↔M24, m13(exit 루프에서 동기 호출)↔ **해당 자리 없음** —
실행자는 exit 루프 파일을 고치지 않으므로 「동기 호출」 변이를 놓을 자리가 없다; 대신 M16(직렬→동시)과 비대기 시험 `TestA193TheExitObserverDoesNotWaitForAResolution`
(런타임 수준 — 해소가 막혀 있는 동안 다른 감독 루프의 주기가 계속 도는지)으로 갈음. m14↔M21, m15↔M22(알림 — 로그 줄의 필드는 시험하지 않음: 이름 붙인 잔여).
