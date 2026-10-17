# Function Logic Map: `notifyRelaxation`

- Source: `internal/app/engine/risk_relaxation_command.go`
- AST evidence: `ast.json` — **편집 전**, :161–183, 분기 1 · 반환 2 · 호출 10, source_sha256 `3f976552ce40…`, 추출 HEAD `81934b46`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(a092 25.6 — archive 게이트 24.4 · 델타 「세울 자기 사유가 없는 기록자는 … 알림기의 기록 전용 입구를 써야 한다」):
  원장 직접 적재(`repo.EnqueueAlert` :168, 알림기 배제 잠금 밖 · 먼저 잠그는 자기 사유 없음)를 알림기의 기록 전용 입구로 옮긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 요청 문맥 — 끊겨도 통지는 기록(`context.WithoutCancel` :168) | HTTP 요청 | — |
| `repo` | `riskRelaxationRepository` — 편집 뒤 이 인자 대신 기록자(`relaxationNoticeRecorder`) | 명령 서비스 `s.j` | — |
| `kind` · `seq` · `target` · `published` · `operator` · `approval` · `at` | 커밋된 해제의 값 | 호출자(해제 두 메서드) | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `repo.EnqueueAlert` 오류 (:177) | `result.NotifyError = err.Error()` | `Notified=false` 결과 (:179) | `TestA066ReleaseStandsWhenTheNoticeFails` |
| 종단 | — | `result.Notified = true` | 결과 (:182) | `TestA066EntryLockReleaseThroughTheEngineEndpoint` · `TestA066NoticeSurvivesTheCallerHangingUp` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal(map…)` :164 | payload(kind · target · release_seq · operator · approval · released_at) | 오류 무시(`_`) — 맵 값이 모두 직렬화 가능 | AST |
| `repo.EnqueueAlert(context.WithoutCancel(ctx), journal.Alert{…})` :168 | 원장 outbox 에 critical 행 적재 — **배제 잠금 밖, 먼저 잠그는 사유 없음** | 오류는 B1 | AST · a092 design D0.3h 4 표 |

## State mutations and fallbacks

- outbox 행 하나(키 `engine.risk_relaxation|kind|seq`, severity critical). 재알림 창 0(`EnqueueAlert` 는 재무장 안 함) — 키에 seq 가 있어 해제마다 새 행.
- 편집 뒤: 같은 행 모양(키 · 제목 · 본문 · payload 바이트)을 알림기 입구 `Notifier.RecordCritical(ctx, e, 0)` 로 기록 — `n.mu` 아래, 실패 시 `ReasonAlertUndelivered` 래치 + 승격 시도.
  기록자가 없으면(알림기 미배선) `Notified=false` + 사유 — 「통지됨」을 거짓으로 말하지 않는다.

## Safety conclusion

- Safe edit boundary: 결과 모양(`Notified` · `NotifyError` · 대상 · seq · 시각)과 행 모양은 그대로, 기록 경로만 입구로.
- High-risk impact: yes — 운영자 완화의 critical 통지(정본 「셈~해제 배제」 부류). 해제 자체(원장 판정)는 무변화.
