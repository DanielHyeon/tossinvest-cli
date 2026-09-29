# Function Logic Map: `PositionPolicyCommandService.ReleaseEntryLossLock`

- Source: `internal/app/engine/risk_relaxation_command.go`
- AST evidence: `ast.json` — **편집 뒤**, :67–97, 분기 4 · 반환 5 · 호출 22, source_sha256 `1c33bf8704cc…`, 추출 커밋 `0e4f26af`. 편집 전 번들은 `analysis/pre-edit/25.6/`에 보존. 분기 넷의 모양 · 순서 불변(줄만 +6 — 파일 위쪽 인터페이스 선언이 늘어남).
- Risk scan: `risk-pattern-report.md`
- 편집(a092 25.6): 마지막 문장의 `notifyRelaxation` 둘째 인자가 원장(`repo`)에서 알림기 기록자(`s.notices`)로 바뀜(변이 R10 — nil 을 넘기면 CAUGHT). 분기 무변화.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req` | 운영자 · 승인 · 사유 · 결속 값 | 제어 endpoint | 판정은 journal API(여기서 재판정 안 함) |
| `s.mu` | 서비스 단위 직렬화 | 서비스 | 해제와 통지가 이 잠금 안 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `relaxationRepo` 오류(원장이 해제 면 없음) (:72) | — | ErrUnwired | (미실행) |
| B2 | `relaxationAuditor` 오류(audit 로그 없음) (:76) | — | ErrAuditUnavailable | `TestA066RelaxationRefusedWithoutAnEngineAuditLog` |
| B3 | `relaxationOperator` 오류(운영자 이름 없음) (:80) | — | invalid | `TestA066RelaxationRequestRefusals` |
| B4 | `repo.ReleaseEntryLossLock` 오류 (:90) | — | `relaxationError` 변환 | `TestA066JournalRefusalsCrossTheWireAsRefusals` · `TestA066RelaxationRequestRefusals` |
| 종단 | — | 종단 :95 `notifyRelaxation(ctx, repo, "entry_lock", …)` — 둘째 인자 `s.notices` | 결과, nil | `TestA066EntryLockReleaseThroughTheEngineEndpoint` · `TestA066LatchReleaseCarriesTheBindingIntoTheEngine` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `repo.Release…` | 원장 해제(OPERATOR · 승인 · 결속 · audit-before-commit) | 오류는 B4 | AST |
| `notifyRelaxation` | 커밋 뒤 통지 — 실패해도 해제는 유효 | 결과에 `Notified` · `NotifyError` | AST |

## State mutations and fallbacks

- 원장 해제(커밋) → 통지 기록. 편집은 통지 기록의 경로만 바꾼다.

## Safety conclusion

- Safe edit boundary: 분기 B1~B4 · 해제 요청 모양 · 결과 모양 불변. 인자 하나.
- High-risk impact: yes(위험 완화 명령) — 해제 판정 무변화, 통지가 critical 기록 부류로 들어감.
