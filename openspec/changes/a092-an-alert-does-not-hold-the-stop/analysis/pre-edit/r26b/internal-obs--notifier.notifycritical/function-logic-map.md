# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :194–249, 분기 4 · 반환 3 · 호출 11, source_sha256 `46c51c2e09e6…`, 추출 커밋 `55963f29`(25라운드 수리 뒤 재추출). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존. 26라운드 수리(`d8769cfb`) 뒤 재추출 — 이 함수 본문 · 좌표 · 분기 불변(같은 파일의 다른 함수 편집으로 파일 해시만 바뀜).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ③ — `fbc6df5f`): B4 가 `escalate` 대신 새 함수 `judge(ctx, e, verdict)` 를 부른다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `verdict` | `claimAndDeliver` 가 올린 판정(세대 · 사유) | `deliver` | `apply=false` 면 조건부 차단 없음 |
| `n.AccountRef` | 빈 값 = 승격 미포함(M5) | 조립 | 무조건 차단 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:192) | 원장 없음 | — | `TestCriticalWithoutAJournalIsLoudRatherThanSilent`, `TestTheTransitionLogLineIsCountable` |
| B2 | if (:196) | 로그 | — | `TestCriticalWithoutAJournalIsLoudRatherThanSilent`, `TestTheTransitionLogLineIsCountable` |
| B3 | if (:216) | claim 실패 → 승격(반환값 무시 — 래치는 잠금 안에서 이미 무조건) | — | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` |
| B4 | if (:238) | 미전달 → `judge`(조건부 차단 → 승격 → 실패면 무조건 차단) | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.claimAndDeliver` | 기록 · 발송 | (sent, owed, verdict, err) | AST |
| `n.judge` | 원칙 E 적용 + K2 | — | AST · `TestA092AFailedEscalationLatchesUnconditionally` |

## State mutations and fallbacks

- `judge`: `BlockUnlessClearedSince(verdict)` → `escalate` → (included && err) 이면 `Block`.

## Safety conclusion

- Safe edit boundary: B1~B3 불변, B4 의 적용 순서만.
- High-risk impact: yes.
