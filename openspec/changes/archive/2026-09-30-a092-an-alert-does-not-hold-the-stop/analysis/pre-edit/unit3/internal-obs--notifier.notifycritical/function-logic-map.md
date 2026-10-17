# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 전**, :176–231, 분기 4 · 반환 3 · 호출 11, source_sha256 `0bc75668ff17…`, 추출 HEAD `b3f14925`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적: 착지 단위 ③: `owed && !sent` 갈래(:223 → `escalate` :228)를 원칙 E 의 판정 적용(조건부 차단 → 승격 → 승격 실패면 무조건 차단, 23.3 K2 · 24.3 M5)으로 바꾼다. 차단 자체가 `deliver` 에서 이 자리로 옮겨 온다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 이면 동기 최선 발송(B1) | 조립 | — |
| `n.AccountRef` | 빈 값 = 승격 미포함(M5) | 조립 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `n.Journal == nil` (:177) | 경고 · 최선 발송 | nil (:187) | `TestCriticalWithoutAJournalIsLoudRatherThanSilent` |
| B2 | `n.Log != nil` (:181) | 경고 | — | 같음 |
| B3 | claim 오류 (:201) | `escalate` :219(잠금 밖) | 오류 (:220) | `TestAClaimThatFailsAttemptsTheDurableBlock` |
| B4 | `owed && !sent` (:223) | `escalate` :228 | — | `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |
| 종단 | — | — | nil (:230) | 전부 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.claimAndDeliver` :200 | 기록 · 발송 | (sent, owed, err) — 편집 뒤 판정 추가 | AST |
| `n.escalate` :219 · :228 | 운영 모드 승격(통지 없음) | 오늘 반환 없음 — 편집 뒤 성공 여부 반환 | AST |

## State mutations and fallbacks

- 편집 뒤 B4: `BlockUnlessClearedSince(판정 세대)` → `escalate` → (승격 포함 && 실패) 이면 `Block`.

## Safety conclusion

- Safe edit boundary: B1~B3 그대로. B4 의 적용 순서만.
- High-risk impact: yes — 전달 실패의 진입 차단 · 모드 승격.
