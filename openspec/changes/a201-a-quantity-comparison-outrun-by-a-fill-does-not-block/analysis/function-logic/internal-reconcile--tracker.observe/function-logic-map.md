# Function Logic Map: `Tracker.Observe` (reconcile) — post-edit(v3)

- Source: `internal/reconcile/mismatch.go` (635-868), GREEN `ffd9dc6d6`.
- AST evidence: `ast.json`(sha 0facba7489a3, 분기 29 — 편집 전 29, 추가 0, return@669 @743 @867). 명령 `go run ./tools/logic-map --file internal/reconcile/mismatch.go --func Tracker.Observe`.
- 재번호: 편집 전 번들(sha 9895f31184e5, 629-856)과 옛/새 ast 를 `../renumber.py`(분기마다 (종류, 그 줄 소스)를 difflib 정렬)로 맞춤 — 29 개 모두 같은 번호로 대응(equal), 사라짐 · 새 분기 0, 좌표만 이동(구조체 칸 2 줄 · 주석 · 편집 문장 때문). 원문 `../renumber-output.txt`(옛 ast 는 `422fad72d` 의 이 번들 ast.json).
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버: `../../coverage/post-internal-reconcile--tracker.observe.txt`(mode set, 선택 시험 세 프로파일 합집합).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff` | 그 주기 비교 | RunOnce call@466 | — |
| `t.adjusted` | 종목 → credit 비교 as-of(조정 credit · 체결 credit 공유) | `AdjustmentApplied` · `FillsOutranAdjustment`(새 파일) | 읽을 수 없는 as-of 는 못 씀(`creditUsableBy`) |
| `t.filled` | 종목 → 체결 credit 부여 as-of(원인 대조 전용) | `FillsOutranAdjustment` | 읽기만 — 해제 여부에 안 쓰임 |
| `t.blocks` · streak · 결속 | 추적기 메모리 | Refresh · 이전 Observe | — |
| 불변식(v3) | 분기 · 반환 · 차단 갈래(B3 else@715 이하) · 연속 횟수 · 결속 · 끝 루프 무변. 바뀐 것은 Cleared 블록의 원인 표지(@712) 와 B5 오류 갈래 몸 안 반박 호출(@668) 둘 | design V3 · V4 · V4b · V5b, 구조 시험 `TestA201TheFillCreditRidesTheAdjustmentCreditsOwnBranches` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@641) | `t.blocks == nil` | 맵 초기화 | — | 기존 |
| B2 (if@656) | `!diff.BlocksEntry()` — 깨끗한 갈래 | — | — | `TestA201W2FillAfterTheLocalReadIsCreditedLikeAnAdjustment` 외 |
| B3 (else@715) | 진입을 막는 비교 — 차단 갈래 | — | — | `TestA201TheFillCycleStillCountsAndBreaksTheStreak` 외 |
| B4 (if@662) | 깨끗함 ∧ `pendingPermanent` | `pendingPermanentKey` 비움 | — | 기존 a110 · V4b 시험 |
| B5 (if@665) | 권위 확인 오류(깨끗한 갈래) | **a201 V4b: 반환 직전 `refuteUsableAdjustmentCredits(disputed, diff.AsOf)`(call@668)** — 분쟁 종목의 usable credit(조정 · 체결) 삭제 | return@669 `finishAuthorityError` | `TestA201AnAuthorityErrorOnACleanComparisonRefutesAnAdjustmentCredit` · `TestA201AnAuthorityErrorOnACleanComparisonRefutesAFillCredit` |
| B6 (range@675) | 활성 블록마다 | — | — | 기존 |
| B7 (if@676) | QM ∧ adjusted_reconcile 아님 | — | continue | 기존 — 운영자 전용 · 다른 원인 블록은 체결 credit 으로도 안 풀림(`TestA201AConvergedSymbolGetsNoFillCredit`) |
| B8 (if@684) | `!creditUsableBy(t.adjusted[종목], AsOf)` — 쓸 credit 없음 | B9 거짓이면 Awaiting | continue | `TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison` · `TestA201ALaterDisputeRefutesTheFillCredit` |
| B9 (if@691) | credit 없는 블록의 a192 `orderTerminalCandidate` 참 | `terminalRelease` 표시 · Cleared append@692 | continue | `TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit` · 기존 a192 |
| B10 (if@698) | credit 있음 ∧ `disputed[종목]` | Awaiting append | continue | V4b 대조 `TestA201WithoutTheAuthorityErrorTheSameInputIsUnchanged` · 기존 |
| (B10 거짓 → @712-713) | credit 있음 ∧ 분쟁 아님 | **a201 V5b: `block.releasedByFill = filled[종목] != "" ∧ filled[종목] == adjusted[종목]`(@712)** · Cleared append@713 | — | `TestA201ALaterAdjustmentNamesTheRelease` · `TestA201ALeftoverFillMarkDoesNotRenameALaterAdjustmentRelease` · W2/W3 해제 시험 |
| B11 (if@720) | 차단 ∧ `pendingPermanent ∧ !earningKeyPresent` | `advanceStreaks` call@722 · `refuteUsableAdjustmentCredits` call@728 | — | 기존 a110 |
| B12 (range@733) · B13 (if@734) | 현재 블록 중 없는 것 | pending 블록 추가 | — | 기존 |
| B14 (if@742) | 권위 확인 오류(차단) | — | return@743 | 기존 a110 |
| B15 (if@746) | `!advanced` | `advanceStreaks` call@747 | — | 기존 · `TestA201TheFillCycleStillCountsAndBreaksTheStreak` |
| B16 (if@749) | `!currentBlocksLatched` | — | — | 기존 |
| B17 (range@750) · B18 (if@751) | 현재 블록 중 없는 것 | pending 추가 · Added | — | 기존 · 「그 주기 차단은 선다」(`assertStuck`) |
| B19 (if@758) | 연속 임계 도달 ∧ 영구 아님 | — | — | 기존 a110 · V4b 픽스처 |
| B20 (if@769) | 영구 블록 없음 | 계좌 영구 블록 추가 | — | 기존 a110 · V4b 픽스처 |
| B21 (range@799) · B22 (if@802) | 영속된 블록 | pending 해제 · 영구 상태 정리 | — | 기존 |
| B23 (range@807) · B24 (if@809) | 권위 블록 · 계좌 범위 | 대체 | — | 기존(B24 선택 시험 미실행 — 편집 무관) |
| B25 (range@814) | 해제된 블록 | 메모리에서 삭제 | — | 기존 · FILL_APPLIED 해제 |
| B26 (range@835) | 해제 커밋된 종목 | `committed` 표시 | — | 기존 |
| B27 (range@838) | credit 마다(체결 credit 포함) | — | — | 기존 · a201 |
| B28 (if@839) | `!creditUsableBy` | — | continue(같은 비교 · 앞선 비교의 credit 은 그대로) | 기존 a083 · `TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison` |
| B29 (if@853) | `disputed ∨ committed ∨ !hasBlockFor` | `delete(t.adjusted, 종목)`@854 | — | 기존 · `TestA201ALaterDisputeRefutesTheFillCredit` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `symbolsInDispute`(call@650) · `promotionKeys`(call@651) | 분쟁 종목 · 승격 키 | 순수 | AST |
| `t.beginOrderTerminalObservation`(call@653) · `t.updateOrderTerminalBindings`(call@820) | a192 결속 관측 · 실격 흔적(같은 종목 수량 불일치면 `tainted`) | 메모리 | AST — V5 근거(`TestA201ABoundRowWithAQuantityMismatchReleasesAsFillApplied`) |
| `t.resolvePendingPermanentAuthority`(call@665 · @742) | 대기 중 영구 승격의 권위 확인 | 장부 읽기 오류 → return | AST |
| `t.refuteUsableAdjustmentCredits`(call@668 — a201 V4b · call@728 기존) | 분쟁 종목의 usable credit 조기 반박 | 메모리 | AST(무편집 함수) |
| `creditUsableBy`(call@684 · @839) | credit 사용 가능(엄격히 나중) | 순수 | AST |
| `t.orderTerminalCandidate`(call@691) | a192 1차 판정 | 장부 읽기 실패 → 유지 | AST |
| `blocksFor`(call@716) · `advanceStreaks`(call@722/@747) | 차단 · 연속 횟수 | 메모리 | AST |
| `t.persist`(call@798) | 장부 쓰기(삽입 · 해제 — 원인은 블록 표지) | 오류 반환 | AST |
| `hasBlockFor`(call@853) | credit 이 풀 블록이 남았나 | 순수 | AST |

## State mutations and fallbacks

- `t.blocks` · streak · 대기 승격 · 결속 · `t.adjusted`(반박 · 삭제). a201 이 새로 쓰는 것은 Cleared 블록 사본의 `releasedByFill` 칸과, B5 오류 갈래에서의 `t.adjusted` 반박(기존 함수 호출) — `t.filled` 는 읽기만.

## Safety conclusion

- Safe edit boundary: @712 대입 1(원인 표지) + B5 몸 안 호출 1(@668, 오류 갈래 한정 — 오류 없는 경로 무변). 분기 · 반환 무변(재번호 equal 29).
- High-risk impact: yes — 해제 원인 기록(`FILL_APPLIED`)과 V4b 의 반박 확대(보수 방향 — 자동 해제가 줄어듦). 해제 여부 판정 분기는 무변.
