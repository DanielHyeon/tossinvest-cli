# Function Logic Map: `PositionPolicyCommandService.ReleaseRiskOverageLatch`

- Source: `internal/app/engine/risk_relaxation_command.go`
- AST evidence: `ast.json` — **편집 전**, :94–124, 분기 4 · 반환 5 · 호출 17, source_sha256 `3f976552ce40…`, 추출 HEAD `81934b46`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(a092 25.6): 마지막 문장의 `notifyRelaxation` 둘째 인자를 원장(`repo`)에서 알림기 기록자(`s.notices`)로 바꾼다. 분기는 건드리지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req` | 운영자 · 승인 · 사유 · 결속 값 | 제어 endpoint | 판정은 journal API(여기서 재판정 안 함) |
| `s.mu` | 서비스 단위 직렬화 | 서비스 | 해제와 통지가 이 잠금 안 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `relaxationRepo` 오류 (:99) | — | ErrUnwired | (미실행) |
| B2 | `relaxationAuditor` 오류 (:103) | — | ErrAuditUnavailable | `TestA066LatchReleaseRefusedWithoutAnEngineAuditLog` |
| B3 | `relaxationOperator` 오류 (:107) | — | invalid | (미실행) |
| B4 | `repo.ReleaseRiskOverageLatch` 오류 (:117) | — | `relaxationError` 변환 | (미실행) |
| 종단 | — | 종단 :122 `notifyRelaxation(ctx, repo, "overage_latch", …)` — 편집 뒤 둘째 인자만 `s.notices` | 결과, nil | `TestA066EntryLockReleaseThroughTheEngineEndpoint` · `TestA066LatchReleaseCarriesTheBindingIntoTheEngine` |

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
