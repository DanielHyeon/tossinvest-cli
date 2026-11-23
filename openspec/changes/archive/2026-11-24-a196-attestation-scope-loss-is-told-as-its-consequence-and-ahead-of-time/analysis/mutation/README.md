# a196 변이 원장 (tasks 4.1)

- 하네스 `run_mutations.py`(저장소 루트에서 실행). 원문 `run-1.txt`(2026-11-23, repo HEAD `527c419ba`), `run-2.txt`(구현 리뷰 r1 처분 뒤 — M01~M12 재실행 + N1~N7, 결과 M 은 run-1 과 같음).
- 방식: 사본 `~/tossos-work/a196-mut-<pid>/`(go.mod·go.sum·internal/)에서 한 번에 한 판. 시작 sha256 단언(사본 = 저장소), 무변이 대조군 GREEN(engine·soak rc 0) 뒤 시작.
  「닿음」 = 대상 문자열이 사본 파일에 정확히 1번(전부 1번 — NOT-REACHED 0). 원복은 원본 바이트 + sha256 대조(매 변이·마지막 전수).
- 시험 명령(변이마다 둘 다): `go test -count=1 -run 'A196|A135|Attestation' ./internal/app/engine`, `go test -count=1 -run 'A196|Supervised|BuildAttestation' ./internal/soak`(`GOFLAGS=-trimpath`).

| # | design §5 | 변이 | 결과 | 죽인 시험(요지) |
|---|---|---|---|---|
| M01 | ① | 기한을 만료로만 | CAUGHT | `TestA196ProofDeadlineGovernsTheDayCount` 외 8 |
| M02 | ② | min → max | CAUGHT | 같은 9 |
| M03 | ③ | 날수 신원에 발급 시각 | CAUGHT | `TestA196SameProofReissueIsNotAnnouncedAgain`·`TestA135SameStage…` 외 2 |
| M04 | ④ | outbox Key 에 발급 신원 | CAUGHT | `TestA196RestartedWatchReusesTheRejectionKey`(Key 직접 비교 — R2-3 i) |
| M05 | ⑤ | 누락 집합을 key 에서 뺌 | CAUGHT | `TestA196RestartedWatchReusesTheRejectionKey`·`TestA196ScopeLossOncePerDayAcrossReissues` |
| M06 | ⑥ | D-2 경계 `>` → `>=` | CAUGHT | `TestA196DeadlineBoundaries`·`TestA135StageBoundaries` 외 2 |
| M07 | ⑦ | soak 수용 경계 `>=` → `>` | CAUGHT | soak `TestA196SupervisedEvidenceLeavesExactlyAtTheDeadline`·`TestA196TheFileDeadlineIsWhenTheIssuerStopsCarryingTheProof` |
| M08a | ⑧ | binding 최소 → 최대(가장 늦은 증거) | CAUGHT | `TestA135SupervisedProofAtIsTheBindingProof`·`TestA196ProofDeadlineGovernsTheDayCount` 외 4 |
| M08b | ⑧ | binding 우회(목록 첫 증거) | CAUGHT | `TestA196ProofDeadlineGovernsTheDayCount` 외 4(R2-3 iv — 픽스처가 가장 이른 증거를 둘째에 둠) |
| M09 | ⑨ | 0 이하를 imminent 로 | CAUGHT | `TestA196DeadlineBoundaries`·`TestA196OverdueIsCriticalOncePerDay`·`TestA135StageBoundaries` |
| M10 | ⑩ | KST → UTC 날짜 | CAUGHT | `TestA196RejectionDayIsTheKSTDate`(R2-3 ii) 하나 |
| M11 | ⑪ | 거절 key 에서 KST 날짜 삭제 | CAUGHT | `TestA135InvalidStatesAreRemindedDaily`·`TestA196RejectionDayIsTheKSTDate` |
| M12 | ⑫ | `MissingEndpoints` 의 `sort.Strings` 삭제 | **SURVIVED — 동등 변이** | — |

## M12 동등 사유(돌려서 생존 확인 뒤)

- 닿음: 대상 1번 일치. `MissingEndpoints` 는 범위 상실 판정마다 불림(judgeAttestation B5 call — 편집 후 커버 108.54-110.46 count 1).
- 왜 동등인가: `MissingEndpoints` 는 `required`(= 상수 목록 `RequiredEndpoints()`, 고정 순서)를 순서대로 돌며 빠진 것을 모은다. 정렬을 지워도 **같은 집합이면 언제나
  같은 순서**(결정적)이므로 a196 이 기대는 성질 — 「같은 누락 집합 = 같은 key」(D5) — 은 그대로다. 달라지는 것은 GET 이 섞인 집합에서 문구·키 안의 나열 순서
  (예: `GET /api/v1/orders` 와 `GET /api/v1/buying-power` 가 함께 빠지면 상수 목록 순서 vs 사전순)뿐이고, 그것도 한 빌드 안에서는 일정하다.
  POST 두 개만 빠지는 현실의 경우는 상수 목록 순서가 이미 사전순이라 출력이 바이트까지 같다.
- 남는 위험: 빌드 사이에 `RequiredEndpoints()` 순서를 바꾸면(정렬 삭제 상태에서) 재시작 직후 같은 상태가 새 key 로 한 번 더 알려질 수 있다 — 정렬이 있는 현재 코드에는 해당 없음.

## 구현 리뷰 r1 처분으로 더한 변이 (run-2)

| # | 처분 | 변이 | 결과 | 죽인 시험 |
|---|---|---|---|---|
| N1 | P1-1 | 감독 증거 문구 갈래를 notice 에만(`&& v.Stage == stageNotice`) — D-2·D-1 이 만료 문구로 샘 | CAUGHT | `TestA196ImminentProofNoticeTellsTheProofDeadline`(r1 때는 생존) |
| N2 | P2-1 | 「보냄」 재설정을 첫 호출에만(`if w.sent == nil`) | CAUGHT | `TestA196RejectionAfterAHealthyFileIsAnnouncedAgain` |
| N3 | P2-2 | `validity > 0` 가드 삭제 | CAUGHT | `TestA196NoProofDeadlineWithoutAValidityWindow` |
| N4 | P2-2 R1 | 만료 무시(증거가 있으면 늘 감독 증거를 기한으로: `true \|\| due.Before(deadline)`) | **SURVIVED — 실무상 동등** | — |
| N5 | P2-4 | soak 수용 기간을 30일로 박기 | CAUGHT | soak `TestA196ANonDefaultValidityMovesBothValues` |
| N6 | P1-2 | 범위 상실을 늘 주문 실행 갈래로(`if true`) | CAUGHT | `TestA196ScopeLossWordingFollowsTheMissingKind` |
| N7 | P2-3 | proof_overdue key 에서 기한 삭제 | **SURVIVED — 실무상 동등** | — |

- N4 동등 사유: 발급자 파일은 증거를 `age ≥ 0` 일 때만 싣고 만료 = 발급 + 유효 기간이므로 「증거 + 유효 기간 ≤ 만료」가 늘 성립 — min 의 결과가 언제나 감독 증거 쪽(같을 때 포함)이다.
  손으로 쓴 「발급 뒤 시각의 증거」 파일에서만 갈린다(Manager 처분: 동등으로 기록).
- N7 동등 사유: 기한이 지난 증거는 발급자가 싣지 않으므로 같은 KST 날짜 안에서 proof_overdue 의 기한이 바뀌는 경우는 「갱신이 실패하는 동안 같은 파일」뿐이고 그때 기한은 같다.
  새 증거가 실리면 기한이 미래가 되어 단계 자체가 바뀐다(Manager 처분: 기록만).

## gstack review 로 더한 변이 (run-3)

| # | 출처 | 변이 | 결과 | 죽인 시험 |
|---|---|---|---|---|
| N8 | gstack review 시험 specialist | 만료가 기한인 통지도 감독 증거 문구로(`if v.DeadlineKind == deadlineProof` → `if true`) | CAUGHT | `TestA196WithoutSupervisedProofCountsTheExpiry`(제목·본문 단언 추가 뒤) |

- run-3 은 M01~N7 을 재실행해 같은 결과(CAUGHT 18·동등 생존 3). run-3 직후 `restartDeadline` 주석 한 줄을 고쳤다(코드 무변경 — 주석만).

## Manager 결정 G12·G10 으로 더한 변이 (run-4)

| # | 결정 | 변이 | 결과 | 죽인 시험 |
|---|---|---|---|---|
| E1 | G12 | 에피소드 시작을 key 에서 빼기 | CAUGHT | `TestA196RecurrenceAfterRecoveryReachesTheRealOutbox`(실제 알림기 + 실제 원장 — 둘째 전송이 정산된 행에 삼켜짐) · `TestA196RestartedWatchStartsANewRejectionEpisode` |
| E2 | G12 | 에피소드를 매 판정 새로 시작(`if true`) | CAUGHT | `TestA196ScopeLossOncePerDayAcrossReissues` 외 4(같은 날 재발급이 다시 알림) |
| E3 | G10 | 할 일에서 엔진 정지 문장 삭제 | CAUGHT | `TestA196ReverifyRemedyTellsTheEngineMustStop` |

- run-4 결과: M01~E3 25판 — CAUGHT 21, 생존 4(M12·N4·N7 동등 — 위 사유, **N2 는 이번에 생존으로 바뀜**).
- N2 동등 사유(run-4 부터): 에피소드 시작이 거절 key 에 들어가고, 날수 key 에는 기한이, 지남 key 에는 기한 + 날짜 + 에피소드가 있으므로 모든 key 가 자기 상태를 스스로 가린다.
  그래서 기한 변경 때의 「보냄」 재설정(prepare B3)은 행동이 아니라 메모리 크기만 정한다 — 재설정을 첫 호출로만 줄여도 알림 결과는 같다. 살아 있는 한 `sent` 에 하루 몇 개씩 key 가 쌓일 뿐이다.
  run-3 까지 N2 를 죽이던 `TestA196RejectionAfterAHealthyFileIsAnnouncedAgain` 는 이제 에피소드로 통과한다(통과 이유가 바뀐 시험 — 기록).
- 시험 이름 변경: `TestA196RestartedWatchReusesTheRejectionKey` → `TestA196RestartedWatchStartsANewRejectionEpisode`(G12 로 재시작 = 새 에피소드 = 새 key). run-1~3 원문의 옛 이름은 그때의 기록이라 고치지 않음.
