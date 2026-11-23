# Function Logic Map: `attestationWatch.prepare`

> **편집 후 지도 — 2단계 GREEN 뒤(2026-11-23).** 편집 전 원본 `../../pre-edit-ast/internal-app-engine--attestationwatch.prepare.ast.json`, 재번호 `renumber.txt`(옛/새 분기를 (종류, 줄 문장)으로 difflib 정렬 — 생성기 `../../renumber.py.txt`).

- Source: `internal/app/engine/attestation_watch.go` (304-335)
- AST evidence: `ast.json` — branches **5** / returns **3** / calls **10**, source SHA-256 `587249217d6328264ddd0ce53b3614496265a12c223003f3e20d17dd6bc89272`(파일 단위, 편집 후).
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
| B1 (if@306) | defer 안 recover | `ok = false` | (패닉 회수) | `TestA135CheckPanicDoesNotStopTheWatch` |
| B2 (if@308) | 로거 있음 | Warn | — | 없음(하네스 log nil, 편집 전과 같음) |
| (L322) | — | **편집: `identity := v.Deadline.UTC().Format(RFC3339Nano)`**(종전 발급|만료) | — | `TestA196SameProofReissueIsNotAnnouncedAgain` · `TestA135SameStageOnANewAttestationIsAnnouncedAgain` 후반 |
| B3 (if@323) | 기한 변경·첫 호출 | `w.identity`·`w.sent` 재설정 — 같은 감독 증거의 재발급은 기한이 같아 재설정 없음 | — | `TestA196NewSupervisedProofRestartsTheCount` · 위 둘 |
| B4 (if@327) | healthy | 없음 | return@328 | `TestA135WatchAnnouncesEachStageOnce` |
| B5 (if@331) | 이미 보낸 키 | 없음 | return@332 | `TestA196ScopeLossOncePerDayAcrossReissues` · `TestA196OverdueIsCriticalOncePerDay` |
| (끝) | — | `w.eventFor` | return@334 | 다수 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `w.load`·`judgeAttestation`·`stageKey`·`w.eventFor` | 편집 전과 같음 | 순수(load 오류는 값) | AST calls 10(Format·UTC 2개 줄어듦) |

## State mutations and fallbacks

- 수신자 쓰기: B3 안의 `w.identity`·`w.sent` 둘뿐(편집 전과 같음). 거절 ↔ 통과 상태가 오가면 기한이 0 ↔ 실값으로 바뀌어 재설정 — 같은 날 「고쳐졌다가 다시 깨짐」 은 다시 알림(보수 쪽).
- 재시작 뒤 메모리 소실 → 현재 단계 1회 재알림(정본 Q6).

## Safety conclusion

- Safe edit boundary: L322 한 줄. 분기 5개 문장 불변(renumber same 5).
- High-risk impact: no.
