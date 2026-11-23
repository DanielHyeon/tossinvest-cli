# Function Logic Map: `inDoubtSettle.sweep`

- Source: `internal/app/engine/in_doubt_settle.go:161-205`(post-edit)
- AST evidence: `ast.json` (**post-edit**, 4c.7 뒤 — source_sha256 `f7852d63…`, 분기 11). Pre-edit 판은 HEAD `a7639378e`(sha `247eedff…`, 분기 7 — 아래 표의 번호, `analysis/cited-ast/d2_inDoubtSettle_sweep.ast.json`).
- **재번호(pre → post)**: B1·B2·B3 동일 · 옛 B4 `if IN_DOUBT && PLACE` → 새 B4 `switch` + B5 `case IN_DOUBT && PLACE` + **새 B6 `case IN_DOUBT && CANCEL`** · 옛 B5(1단) → B7 · **새 B8 `range cancels`·B9 `closeCancel` 오류**(2단 앞) · 옛 B6·B7 → B10·B11. 아래 표는 pre-edit 번호다(BTM 은 post-edit 번호).
- Risk scan: `risk-pattern-report.md` (적중 0)
- 편집 범위(design D10, Manager 결정 2026-11-24): B4 의 대상 고르기에 갈래 하나 — IN_DOUBT·CANCEL 이면 별도 목록. 1단 알림 뒤·2단 발주 해소 앞에서 그 목록을 새 함수(`closeCancel`, 새 파일)로 닫음. 알림(1단)·발주 해소(2단)·원장 읽기 실패 경로 무편집.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.deps.Ledger.PendingAttempts` | RECORDED·DISPATCH_STARTED·ACKED·IN_DOUBT 전부(종류 무관) | 원장 | 오류 → B1(ctx 면 반환, 아니면 연속 실패 알림) |
| 대상 | 오늘 IN_DOUBT PLACE 만 | 정본 「세션 중 해소: 대상」 | — |
| 불변식 | 진행 중 전송(RECORDED·DISPATCH_STARTED)·ACKED 는 손대지 않음 · 해소는 직렬 · 브로커 호출은 해소 안에서만 · 취소된 ctx 로 알림 기록 안 함 | 정본 · 파일 머리 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L163 | 원장 목록 읽기 실패 | — | — | TestA193AnUnreadableLedgerAlertsAfterTenMinutes |
| B2 L164 | (B1 안) ctx 취소 | — | ctx 오류 | TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop 계열 |
| B3 L173 | 목록 순회 | — | — | TestA193OnlyInDoubtPlaceAttemptsAreTouched |
| B4 L175 | IN_DOUBT && PLACE | targets 에 넣음 | — | TestA193OnlyInDoubtPlaceAttemptsAreTouched |
| B5 L181 | 1단 — 처음 봄·10분 알림 | critical 기록 | — | TestA193FirstSightIsAlertedOnce · TestA193AnAttemptInDoubtPastTheBoundAlertsOnce |
| B6 L186 | 2단 — 직렬 해소 | `resolve` | — | TestA193AttemptsAreResolvedOneAtATime |
| B7 L187 | 해소 오류(ctx) | — | ctx 오류 | TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `retryUnsent` | 못 보낸 알림 재기록 | — | AST |
| `Ledger.PendingAttempts` | 미종결 목록 | B1 | AST |
| `noteUnreadable` | 연속 읽기 실패 알림 | — | AST |
| `noteFirstSight` · `noteStale` | 1단 알림 | 취소 ctx 면 기록 안 함 | AST |
| `resolve` → `Resolver.Resolve` | 조회 대조 해소(분 단위) | ctx 취소만 반환 | AST |

## State mutations and fallbacks

- 실행자 메모리: `unreadableSince`·`unreadableID`(성공 시 비움). 원장 전이는 `resolve` 안의 해소기만.
- High-risk: yes(원장 전이 — 해소). 편집 뒤에는 원장 함수가 CANCEL attempt 를 전이함.

## Safety conclusion

- Edit boundary(실제): 대상 거름을 switch 로(PLACE 갈래 조건 무변) + CANCEL 갈래 + 1단 뒤·2단 앞의 취소 닫기 루프. 1단 알림·2단 해소·읽기 실패 경로 무편집.
- 닫기는 `closeCancel`(새 파일) → `journal.CloseInDoubtCancelOnTargetClose(엔진 source)` — 원장만, 브로커 0, 알림 0. 오류는 ctx 취소만 반환(나머지는 분류 로그 한 번·다음 스윕).
- 안전 논증: 닫기는 미종결 가드만 풂 — 손절은 그 뒤 길 치우기가 같은 종결 정의로 다시 판정(design D10, Manager Q6, 시험 31).
