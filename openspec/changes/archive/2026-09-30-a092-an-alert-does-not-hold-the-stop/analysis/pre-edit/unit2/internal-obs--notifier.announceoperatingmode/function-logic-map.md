# Function Logic Map: `Notifier.AnnounceOperatingMode`

- Source: `internal/obs/mode.go`
- AST evidence: `ast.json` — **편집 전**, :49–74, 분기 2 · 반환 2 · 호출 7, source_sha256 `95d6a3417808…`, 추출 HEAD `8c390aa6`(2026-09-29, base `721d0338`과 파일 동일).
- Risk scan: `risk-pattern-report.md`
- 편집 목적(24판 K1 · M10): 통지 사건 구성을 함수 하나로 빼서 기록 전용 announcer와 함께 쓰고, 중복 제거 키에 **전이 행의 신원 `rec.ID`**를 넣는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n` | nil 허용 | 배선 | B1 — nil이면 통지하지 않고 nil 반환 |
| `previous` | `""`(행 없음) · NORMAL · ENTRY_BLOCKED · HALT_ALL | `TransitionOperatingMode`의 `current.Mode` | `modeLabel`이 빈 값을 NORMAL로 |
| `rec` | 방금 커밋된 전이 행(`ID` · `AccountRef` · `Mode` · `Cause` · `Actor`) | `TransitionOperatingMode` :452-454 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `n == nil` (:50) | 없음 | `nil` (:51) | `TestAnnouncingWithoutANotifierIsSafe` |
| B2 | `MoreConservativeMode(previous, rec.Mode) == previous && previous != rec.Mode` (:54) — 완화 | `direction = "relaxed"` | — | `TestARelaxationIsAnnouncedAsARelaxation` |
| 종단 | — | `n.Notify(ctx, Event{…Key: "operating_mode:"+계정+":"+모드…})` (:57) | `Notify`의 반환 (:57) | `TestTheTransitionLogLineIsCountable` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `journal.MoreConservativeMode` :54 | 방향(강화/완화) 판정 | 순수 함수 | AST |
| `n.Notify` :57 | critical 사건 통지(`EventOperatingMode` critical, `event.go:328`) — 동기 경로면 `claimAndDeliver` → `deliver` | `Notify`는 기록 실패에서만 오류 | AST · `mode.go:57` |
| `modeLabel` · `permission` · `fmt.Sprintf` | 제목 · 본문 | 순수 | AST |

## State mutations and fallbacks

- 이 함수 자신은 상태를 바꾸지 않는다. 부작용은 `n.Notify`가 한다(outbox 행 · 로그 · 동기 전송).
- **키**(:62)가 계정 + 목적 모드라서, 재알림 창(`DefaultRemindAfter`, 1h) 안에서 같은 목적 모드로의 새 전이는 옛 정착 행에 흡수된다 — 22라운드 K1.
  편집 뒤에는 키에 `rec.ID`를 붙인다. 같은 전이의 재통지는 원장이 막는다(변화 없으면 통지 전에 반환 — `TransitionOperatingMode` B15 `:410` · B16 `:417`).

## Safety conclusion

- Safe edit boundary: 사건 구성을 `operatingModeEvent(previous, rec)`로 빼고 키에 `rec.ID`를 붙인다. 분기 B1 · B2와 `Notify` 호출은 그대로다.
- High-risk impact: yes — critical 통지의 중복 제거 신원. 통지 수는 사람의 완화 행동 수에 묶인다(자동 완화 없음).
