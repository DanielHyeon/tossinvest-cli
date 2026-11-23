# a196 변이 원장 (tasks 4.1)

- 하네스 `run_mutations.py`(저장소 루트에서 실행). 원문 `run-1.txt`(2026-11-23, repo HEAD `527c419ba`).
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
