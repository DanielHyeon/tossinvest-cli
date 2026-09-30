# Function Logic Map: `notifyRelaxation`

- Source: `internal/app/engine/risk_relaxation_command.go`
- AST evidence: `ast.json` — **편집 뒤**, :170–196, 분기 2 · 반환 3 · 호출 10, source_sha256 `1c33bf8704cc…`, 추출 커밋 `0e4f26af`. 편집 전 번들은 `analysis/pre-edit/25.6/`에 보존(:161–183, 분기 1, `3f976552ce40…`).
- Risk scan: `risk-pattern-report.md`
- 편집(a092 25.6 — archive 게이트 24.4): 원장 직접 적재(`repo.EnqueueAlert`)를 알림기 기록 전용 입구(`notices.RecordCritical`)로 옮김.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 끊겨도 통지는 기록(`context.WithoutCancel`) | HTTP 요청 | — |
| `notices` | `relaxationNoticeRecorder` — nil 허용 | 명령 서비스 `s.notices`(엔진 알림기) | B1 — nil 이면 `Notified=false` + `ErrAlertNotDurable` 문구 |
| 나머지 값 | 커밋된 해제의 값 | 해제 두 메서드 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `notices == nil` (:173) — **새 분기** | `result.NotifyError` = 기록 불가 | `Notified=false` (:176) | `TestA092RelaxationWithoutANotifierIsNotNotified` |
| B2 | `RecordCritical` 오류 (:190) — 편집 전 B1(:177) | `result.NotifyError = err.Error()` | `Notified=false` (:192) | `TestA066ReleaseStandsWhenTheNoticeFails` · `TestA092RelaxationNoticeFailureLatchesEntries` |
| 종단 | — | `result.Notified = true` | 결과 (:195) | `TestA066EntryLockReleaseThroughTheEngineEndpoint` · `TestA092RelaxationNoticeRowShapeIsUnchanged` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `notices.RecordCritical(context.WithoutCancel(ctx), obs.Event{…}, 0)` | 알림기 배제 잠금 아래 critical 행 기록(임차 없음 · 재알림 창 0). 실패 시 입구가 `ReasonAlertUndelivered` 래치 + 승격 시도 | 오류는 B2 | AST · 구조 핀 `TestA092NotifyRelaxationCallsOnlyTheEntry`(호출 1 · 마지막 인자 리터럴 0 · 원장 호출 0) |
| `strconv.FormatInt` · `fmt.Sprintf` · `strings.TrimSpace` · `journal.RFC3339` | 키 · 본문 · 시각 | 순수 | AST |

## State mutations and fallbacks

- outbox 행 하나 — 키 · 유형 · 등급 · 제목 · 본문 · payload 바이트가 이행 전과 같다(`TestA092RelaxationNoticeRowShapeIsUnchanged` 가 정확한 값 대조, 변이 R05 · R06).
  payload 는 입구가 `Fields` 를 JSON 으로 직렬화해 만든다(키 정렬) — 이행 전 `json.Marshal(map)` 과 같은 바이트.
- 기록 실패 시 진입 잠금이 새로 생긴다(이행 전에는 결과 문구만). 해제는 유효.

## Safety conclusion

- Safe edit boundary: 결과 모양 · 행 모양 불변, 경로만 입구로. 해제 판정 무변화.
- High-risk impact: yes — 운영자 완화의 critical 통지가 정본 「critical 기록 부류」에 들어감(보수 방향).
