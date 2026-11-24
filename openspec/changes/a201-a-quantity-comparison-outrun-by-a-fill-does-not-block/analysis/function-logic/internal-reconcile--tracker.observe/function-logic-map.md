# Function Logic Map: `Tracker.Observe` (reconcile) — pre-edit(v3)

- Source: `internal/reconcile/mismatch.go` (629-856), HEAD `6e36fd2c0`.
- AST evidence: `ast.json`(sha 9895f31184e5, 분기 29, return@660 @731 @855). 명령 `go run ./tools/logic-map --file internal/reconcile/mismatch.go --func Tracker.Observe`. a192 아카이브 번들의 편집 뒤 ast 와 sha 동일(a192 이후 무편집).
- Risk scan: `risk-pattern-report.md`(일치 0).
- 이 번들은 **편집 전** 지도. v3 편집 두 자리: (1) B10 if@689 거짓 쪽(Cleared@701 직전)에 원인 표지 1문장, (2) B5 if@659 몸 안 return@660 직전에 `t.refuteUsableAdjustmentCredits(disputed, diff.AsOf)` 1 호출(V4b — codex v3 r1). 분기 추가 없음 예상(대입식 · 호출문). 편집 뒤 재추출 · 옛/새 정렬 재번호는 2.4.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff` | 그 주기 비교 | RunOnce call@464 | — |
| `t.adjusted` | 종목 → 크레딧 비교 as-of(v3 에서 체결 credit 도 같은 맵) | `AdjustmentApplied` · (v3) 새 부여 메서드 | 읽을 수 없는 as-of 는 못 씀(`creditUsableBy`) |
| `t.blocks` · streak · 결속 | 추적기 메모리 | Refresh · 이전 Observe | — |
| 불변식(v3) | 차단을 세우는 갈래(B3 else@703 이하) · 연속 횟수 · 결속 · 반박 · 끝 루프는 무변. 바뀌는 것은 Cleared 블록의 원인 표지뿐 | design V3 · V4 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@635) | `t.blocks == nil` | 맵 초기화 | — | 기존 |
| B2 (if@650) | `!diff.BlocksEntry()` — 깨끗한 갈래 | — | — | 기존 · v3 해제 시험 |
| B3 (else@703) | 진입을 막는 비교 — 차단 갈래 | — | — | 기존 · v3 「그 주기 Observe 무변」 |
| B4 (if@656) | 깨끗함 ∧ `pendingPermanent` | `pendingPermanentKey` 비움 | — | 기존 |
| B5 (if@659) | 권위 확인 오류(깨끗한 갈래) | **v3 편집 지점(V4b): 반환 직전 분쟁 종목의 usable credit 반박** | return@660 `finishAuthorityError` | 기존 · v3 「외부 포지션 분쟁 + 권위 실패 → credit 반박」(조정 · 체결 두 판) |
| B6 (range@666) | 활성 블록마다 | — | — | 기존 |
| B7 (if@667) | QM ∧ adjusted_reconcile 아님 | — | continue | 기존 — v3: 운영자 전용 · 다른 원인 블록은 체결 credit 으로도 안 풀림 |
| B8 (if@675) | `!creditUsableBy(t.adjusted[종목], AsOf)` — 쓸 credit 없음 | B9 거짓이면 Awaiting append@686 | continue | 기존 · v3 「credit 없음 → 대기」 |
| B9 (if@682) | 크레딧 없는 블록의 a192 `orderTerminalCandidate` 참 | `terminalRelease` 표시 · Cleared append@683 | continue | a192 시험 |
| B10 (if@689) | 크레딧 있음 ∧ `disputed[종목]` | Awaiting append@698 | continue | 기존 · v3 「다른 종목/같은 종목 분쟁이면 대기」 |
| (B10 거짓 → Cleared@701) | 크레딧 있음 ∧ 분쟁 아님 | Cleared append — **v3 편집 지점: 직전에 원인 표지(`t.filled[종목] == t.adjusted[종목]` 이면 체결)** | — | v3 W2 · W3 해제 시험(FILL_APPLIED) · 기존 조정 해제 시험(ADJUSTMENT_APPLIED 무변) |
| B11 (if@708) | 차단 ∧ `pendingPermanent ∧ !earningKeyPresent` | `advanceStreaks` call@710 · `refuteUsableAdjustmentCredits` call@716 · 현재 블록 latch | — | 기존 · v3 「체결 credit 도 같은 반박」 |
| B12 (range@721) · B13 (if@722) | 현재 블록 중 없는 것 | pending 블록 추가 | — | 기존 |
| B14 (if@730) | 권위 확인 오류 | — | return@731 | 기존 |
| B15 (if@734) | `!advanced` | `advanceStreaks` call@735 | — | 기존 |
| B16 (if@737) | `!currentBlocksLatched` | — | — | 기존 |
| B17 (range@738) · B18 (if@739) | 현재 블록 중 없는 것 | pending 추가 · Added | — | 기존 · v3 「그 주기 차단은 선다」 |
| B19 (if@746) | 연속 임계 도달 ∧ 영구 아님 | — | — | 기존 |
| B20 (if@757) | 영구 블록 없음 | 계좌 영구 블록 추가 | — | 기존 · v3 「연속 횟수 무변」 |
| B21 (range@787) · B22 (if@790) | 영속된 블록 | pending 해제 · 영구 상태 정리 | — | 기존 |
| B23 (range@795) · B24 (if@797) | 권위 블록 | 대체 | — | 기존 |
| B25 (range@802) | 해제된 블록 | 메모리에서 삭제 | — | 기존 |
| B26 (range@823) | 해제 커밋된 종목 | `committed` 표시 | — | 기존 |
| B27 (range@826) | 크레딧마다 | — | — | 기존 · v3 체결 credit 도 같은 루프 |
| B28 (if@827) | `!creditUsableBy` | — | continue(같은 비교 · 앞선 비교의 credit 은 건드리지 않음) | 기존 · v3 「부여 주기에 안 쓰이고 안 지워짐」 |
| B29 (if@841) | `disputed ∨ committed ∨ !hasBlockFor` | `delete(t.adjusted, 종목)`@842 | — | 기존 · v3 「나중 분쟁 → 체결 credit 소멸」 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `symbolsInDispute`(call@644) · `promotionKeys`(call@645) | 분쟁 종목 · 승격 키 | 순수 | AST |
| `t.beginOrderTerminalObservation`(call@647) · `t.updateOrderTerminalBindings`(call@808) | a192 결속 관측 · 실격 흔적(같은 종목 수량 불일치면 `tainted`) | 메모리 | AST — v3 V5 의 근거 |
| `creditUsableBy`(call@675 · @827) | credit 사용 가능 판정(엄격히 나중) | 순수 | AST |
| `t.orderTerminalCandidate`(call@682) | a192 1차 판정 | 장부 읽기 실패 → 유지 | AST |
| `blocksFor`(call@704) · `advanceStreaks`(call@710/@735) · `refuteUsableAdjustmentCredits`(call@716) | 차단 · 연속 횟수 · 조기 반박 | 메모리 | AST |
| `t.resolvePendingPermanentAuthority`(call@659 · @730) | 대기 중 영구 승격의 권위 확인 | 장부 읽기 오류 → return | AST |
| `t.persist`(call@786) | 장부 쓰기(삽입 · 해제) | 오류 반환(블록 · 게이트 유지) | AST |
| `hasBlockFor`(call@841) | credit 이 풀 블록이 남았나 | 순수 | AST |

## State mutations and fallbacks

- `t.blocks` · streak · 대기 승격 · 결속 · `t.adjusted`(반박 · 삭제)만. v3 는 `Block` 의 원인 표지 칸(Cleared 블록 사본)에만 새로 씀 — `t.filled` 는 읽기만.

## Safety conclusion

- Safe edit boundary: Cleared@701 직전 1문장(원인 표지) + B5 몸 안 반박 호출 1(오류 갈래 한정 — 오류 없는 경로 무변). 분기 · 반환 무변.
- High-risk impact: yes — 해제 원인 기록이 바뀜(`FILL_APPLIED`). 해제 여부 판정은 무변(체결 credit 이 맵에 있는지는 새 부여 메서드가 정함).
