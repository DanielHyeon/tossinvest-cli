# Function Logic Map: `attestationWatch.prepare`

> **편집 후 지도 — 2단계 GREEN 뒤(2026-11-23).** 편집 전 원본 `../../pre-edit-ast/internal-app-engine--attestationwatch.prepare.ast.json`, 재번호 `renumber.txt`(옛/새 분기를 (종류, 줄 문장)으로 difflib 정렬 — 생성기 `../../renumber.py.txt`).

- Source: `internal/app/engine/attestation_watch.go` (309-353)
- AST evidence: `ast.json` — branches **9** / returns **3** / calls **14**, source SHA-256 `9b71a387860c809bc6ee8f80528cd5cc1446fdd8dd080eac14164d7548eae1f6`(파일 단위, 편집 후).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func attestationWatch.prepare`.
- Risk scan: `risk-pattern-report.md`.
- a166 census: `attestation_watch.go`·`a135_attestation_watch_internal_test.go` 0건(1.4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `w.identity` | **마지막 기한**(UTC RFC3339Nano). 거절 단계는 기한 0 → 같은 값 | 이 함수 | 초기 "" |
| `w.sent` | 이 기한에서 알린 단계 키 | 이 함수가 만들고 `checkOnce`(무편집)가 씀 | 초기 nil → B3 |
| 불변식 | 상태는 Run goroutine 하나만 만짐 | — | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@311) | defer 안 recover | `ok = false` | (패닉 회수) | `TestA135CheckPanicDoesNotStopTheWatch` |
| B2 (if@313) | 로거 있음 | Warn | — | 없음(하네스 log nil, 편집 전과 같음) |
| (L327) | — | **편집: `identity := v.Deadline.UTC().Format(RFC3339Nano)`**(종전 발급|만료) | — | `TestA196SameProofReissueIsNotAnnouncedAgain` · `TestA135SameStageOnANewAttestationIsAnnouncedAgain` 후반 |
| B3 (if@328) | 기한 변경·첫 호출 | `w.identity`·`w.sent` 재설정 — 같은 감독 증거의 재발급은 기한이 같아 재설정 없음 | — | `TestA196NewSupervisedProofRestartsTheCount` · 위 둘 |
| B4 (switch@336) **새(G12)** | 거절 에피소드 판정 | — | — | — |
| B5 (case@337) | healthy·notice·imminent | 에피소드 신원 "" | — | `TestA196RecurrenceAfterRecoveryReachesTheRealOutbox` |
| B6 (case@338) | 그 밖(거절 단계·proof_overdue) | 에피소드 신원 = 단계 + 누락 집합 + 기한 | — | `TestA196ScopeLossOncePerDayAcrossReissues` |
| B7 (if@341) **새(G12)** | 에피소드 신원이 직전과 다름 | `w.episode`·`w.episodeAt` = 이번 판정 시각(새 에피소드). 끝에 `v.EpisodeAt = w.episodeAt` | — | `TestA196RecurrenceAfterRecoveryReachesTheRealOutbox`(실제 원장) · `TestA196RestartedWatchStartsANewRejectionEpisode` |
| B8 (if@345) | healthy | 없음 | return@346 | `TestA135WatchAnnouncesEachStageOnce` |
| B9 (if@349) | 이미 보낸 키 | 없음 | return@350 | `TestA196ScopeLossOncePerDayAcrossReissues` · `TestA196OverdueIsCriticalOncePerDay` |
| (끝) | — | `w.eventFor` | return@352 | 다수 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `w.load`·`judgeAttestation`·`stageKey`·`w.eventFor` | 편집 전과 같음 | 순수(load 오류는 값) | AST calls 10(Format·UTC 2개 줄어듦) |

## State mutations and fallbacks

- 수신자 쓰기: B3 안의 `w.identity`·`w.sent`, B7 안의 `w.episode`·`w.episodeAt`(G12 — 거절 에피소드, 프로세스 메모리뿐). 거절 ↔ 통과 상태가 오가면 기한이 0 ↔ 실값으로 바뀌어 재설정 — 같은 날 「고쳐졌다가 다시 깨짐」 은 다시 알림(보수 쪽).
- 재시작 뒤 메모리 소실 → 현재 단계 1회 재알림(정본 Q6).

## Safety conclusion

- Safe edit boundary: 신원 한 줄(L327) + 에피소드 블록(L333-343, G12). 옛 분기 5개 문장 불변(renumber same 5, 새 분기 4).
- High-risk impact: no.
