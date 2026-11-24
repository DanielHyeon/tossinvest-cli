# a201 Pre-Edit 선언 (High-risk — tasks 2.1)

- 시각: 2026-11-24, HEAD `049c1aea0`(origin/main `de3be078a` — a199 착지 — merge, base 재고정 `35d28a535`).
- 설계: design v3(Manager freeze 2026-11-24 — review.md 「Manager freeze」).

## 대상 심볼 · 편집 자리(기존 함수 넷 — 넘으면 멈춤)

| 함수 | 편집 | 번들 |
|---|---|---|
| `ReconcileDriver.RunOnce`(reconcileloop.go:425-479) | call@434 앞 로컬 읽기 전 표식 1문장 · B7 if@470 뒤 · call@477 앞 부여 1자리 | `function-logic/internal-app-engine--reconciledriver.runonce/`(sha 52a738a58082) |
| `Tracker.Observe`(mismatch.go:629-856) | Cleared@701 직전 원인 표지 1문장 · B5 if@659 몸 안 return@660 직전 `refuteUsableAdjustmentCredits(disputed, diff.AsOf)` 1 호출(V4b) | `…--tracker.observe/`(sha 9895f31184e5) |
| `Tracker.persist`(mismatch.go:878-954) | call@935 `ReleaseReconcile` 의 `Cause` · `Evidence` 를 블록 표지로 고름 | `…--tracker.persist/` |
| `journal.ValidReconcileReleaseCause`(reconcile_states.go:111-119) | case 에 `ReconcileReleaseFillApplied` 1 | `…--validreconcilereleasecause/` |

- 새 파일: `internal/reconcile/a201_fill_credit.go`(표식 타입 · 표식 읽기 · 부여 대상 계산 · `Tracker.FillsOutranAdjustment`), `internal/journal/a201_fill_applied.go`(상수).
- 구조체 칸: `Tracker.filled map[string]string`(mismatch.go 의 Tracker 선언), `Block.releasedByFill bool`(mismatch.go 의 Block 선언) — 타입 선언 편집, 함수 아님.
- 무편집: `AdjustmentApplied` · `refuteUsableAdjustmentCredits` · `creditUsableBy` · `hasBlockFor` · `ConvergeQuantities` · `Compare` · `Restore` · `Resolve` · `Refresh` · `orderTerminalCandidate` · `enterReconcileWithBinding` · `buildGateway` · `logCycle` · `note` · 스키마 36.

## 호출부

- `RunOnce` ← `Run` call@412(1). `Observe` ← RunOnce call@464(1). `persist` ← Observe call@786(1). `ValidReconcileReleaseCause` ← `ReleaseReconcile` :269 · `ReleaseReconciles` :351 · 종결 해제 :106.

## 기존 시험(선택 실행 대상)

- engine: `reconcileloop_test.go` · a083 · a192 로그 · a194 · a071 · a201 재현(v3 계약으로 뒤집음).
- reconcile: a083 · a110 · a192 · a195 · a200 cycles · restore · mismatch · converge · compare 계열.
- journal: `TestUnknownReleaseCauseIsRefused` · 해제 · a192 · a200 · risk bucket broker-zero.
- census: `./internal/accountbudgetruntime`.

## 불변식

1. 그 주기 `Observe` 의 차단 · 결속 · 연속 횟수 · 반박 · 건강 계수는 편집 전과 같다(부여는 Observe 성공 뒤, 오류 주기엔 없음).
2. 체결 credit 은 조정 credit 과 같은 맵 · 같은 분기로 쓰이고 반박 · 소멸한다(새 사용 · 반박 분기 0). 예외 V4b 는 오류 갈래 안 호출 1.
3. 해제 원인은 `filled != "" ∧ filled == adjusted` 일 때만 `FILL_APPLIED`.
4. 영구 · 계좌 범위 블록은 안 풂(B7). 재기동 · 운영자 해제 뒤 복원 없음.
5. broker-zero 허용 목록 무편집 → `FILL_APPLIED` 행 거절.
6. LIVE 주문 side effect 0 — 시험은 임시 장부 · 가짜 브로커만.

## 실패할 시험(RED — design §3 1~14)

`analysis/tdd/red.txt` 에 원문. 편집 전 코드에서 실패해야 할 것: 1단계 재현 6 의 v3 판(W2 · W3-a · tx 직전 · W3-b 합성 · 앞 크레딧 위 W2 → 다음 깨끗한 주기 `FILL_APPLIED`), R1-1 변형 해제, V4b 두 판(오류 경로에서 credit 반박), broker-zero 거절(새 원인), `FILL_APPLIED` 유효. 편집 전에도 통과해야 할 것(무변 고정): R1-1 순서 · 나중 분쟁 소멸 · credit 없음 사례 · 재기동 · 다른 종목 분쟁 대기 · 그 주기 Observe 무변(R2-1 · R2-2 입력) · W4 · V5 가로채지 않음 · 판 중단 뒤 B 의 잔여.

## 되돌리기 · 설정 영향

- 되돌리기: 이 change 의 Go 커밋을 되돌리면 해제 근거가 오늘로 돌아감. 장부에 `FILL_APPLIED` 행이 남아도 읽는 생산 소비자는 broker-zero 허용 목록뿐(거절 — 보수). 스키마 무변 → 데이터 되돌리기 불요.
- 설정 · 토글 · 배선 무변(`buildGateway` · census 무편집).
