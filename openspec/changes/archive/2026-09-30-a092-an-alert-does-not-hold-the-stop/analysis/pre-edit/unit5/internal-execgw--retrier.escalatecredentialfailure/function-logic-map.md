# Function Logic Map: `Retrier.escalateCredentialFailure`

- Source: `internal/execgw/retry.go`
- AST evidence: `ast.json` — **편집 전**, :409–419, 분기 2, source_sha256 `1d23eb356843…`, 추출 HEAD `b01e0cd0`(작업 트리 — 파일 무편집).
- Risk scan: `risk-pattern-report.md`
- 편집 목적(25라운드 보이스 B #7): `EscalateOperatingMode` 가 `ErrModeAnnouncementFailed` 를 `changed=true` 와 함께 돌려주면 전이는 **커밋됐고** 통지 기록만 실패한 것. 오늘은 그것도 「did not reach the operating mode, so a restart would lift the block」으로 오기함. 그 갈래를 가려 사실대로 적음(판정 · 반환 형태 불변 — 오류는 여전히 조회 오류에 join).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.Escalate` · `r.AccountRef` | 둘 다 있어야 승격 | 조립 | B1 → nil |
| `r.Announcer` | exit 복사본은 기록 전용, 공유본은 동기 | `exitSideRetrier` | 통지 실패는 `ErrModeAnnouncementFailed` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 승격 수단 · 계좌 없음 (:410) | — | nil | `TestAuthFailureLatchesEntryImmediately` |
| B2 | 승격 오류 (:413) | — | 「did not reach the operating mode」 오류 — **통지 실패도 여기로 옴(오기)** | (미실행 — 측정) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.Escalate.EscalateOperatingMode(…, r.Announcer)` | 자격증명 거절 → ENTRY_BLOCKED | `ErrModeAnnouncementFailed` 는 커밋 뒤 통지 실패 | AST · `operating_mode.go` |

## State mutations and fallbacks

- 없음(원장 쓰기는 호출된 쪽).

## Safety conclusion

- Safe edit boundary: 오류 문구 분기 하나 — 반환 오류 유무 · 래치(호출자 `Query` 가 이미 함) 불변.
- High-risk impact: 중간 — 자격증명 승격 경로의 보고 정확성.
