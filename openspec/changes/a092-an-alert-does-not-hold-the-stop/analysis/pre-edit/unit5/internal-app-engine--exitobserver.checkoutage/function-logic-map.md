# Function Logic Map: `ExitObserver.checkOutage`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json` — **편집 전**, :817–854, 분기 5, source_sha256 `522d5d81c499…`, 추출 HEAD `b01e0cd0`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(25라운드 보이스 B #7): 관측 두절 승격이 `ErrModeAnnouncementFailed`(커밋됨 · 통지 기록 실패)로 돌아오면 오늘은 「did not reach the operating mode」로 로그하고 `cycle.Escalated` 를 세우지 않음 — 실제로는 강화됨. 그 갈래를 가려 사실대로 로그하고 `cycle.Escalated = changed` 를 세움.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.lastObserved` · `o.startedAt` | 마지막 관측 | 루프 | — |
| `o.opts.Escalate` · `o.opts.Announcer` | 기록 전용 announcer(단위 ②) | `Context.ExitObserver` | 통지 실패 = 기록 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 관측 이력 없음 (:819) | `since = startedAt` | — | `TestA092CredentialTighteningRecordsWithoutSending` |
| B2 | 두절 한도 전 (:822) | — | 반환 | 같음 |
| B3 | 이미 알림 (:825) | — | 반환 | `TestASustainedOutageBlocksEntriesAndAlertsOnce` |
| B4 | 승격 수단 없음 (:843) | — | 반환 | (미실행) |
| B5 | 승격 오류 (:848) | 오류 로그 — **통지 실패도 여기(오기)** | 반환(`Escalated` 미설정) | (미실행) |
| 종단 | — | `cycle.Escalated = changed` | — | `TestASustainedOutageBlocksEntriesAndAlertsOnce` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.alert` | 두절 critical 알림(기록 전용) | 기록 실패 로그 | AST |
| `EscalateOperatingMode(…, o.opts.Announcer)` | 두절 → ENTRY_BLOCKED | `ErrModeAnnouncementFailed` = 커밋됨 | AST |

## State mutations and fallbacks

- `o.outageRaised` · `cycle.Escalated`.

## Safety conclusion

- Safe edit boundary: B5 안의 갈래 하나 — 두절 판정 · 알림 · 승격 호출 불변.
- High-risk impact: 중간 — 관측 두절 강화의 보고 정확성(손절 경로 무관).
