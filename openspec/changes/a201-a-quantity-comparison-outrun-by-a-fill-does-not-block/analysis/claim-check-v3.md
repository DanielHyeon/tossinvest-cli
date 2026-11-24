# a201 claim check v3 — 체결 credit 설계의 전제 (q1)~(q7)

HEAD `6e36fd2c0`(origin/main `d66536d9b` = a200 착지를 merge). merge 가 바꾼 production Go 는 `internal/journal/position_projection.go`(a200) · 새 파일 `a200_adjustment_close_fill_recognition.go` 뿐 — `reconcileloop.go` · `mismatch.go` · `order_terminal_release.go` · `reconcile_states.go` 무변(`git diff --stat ec44a8bb1 6e36fd2c0 -- internal`). AST 는 `cited-ast/` · `v3-evidence/ast/`(둘 다 현재 소스와 sha 동일). 탐침 `v3-probe/`(scratch 사본 — v3 모양 최소 구현, 끝에 제거) · `v3-evidence/probe-*`(별도 팀메이트).

## 판정 요약 — 설계를 바꿔야 하는 사실 0

| 전제 | 판정 | 한 줄 |
|---|---|---|
| (q1) 체결 credit 이 조정 credit 의 반박 · 삭제 규칙을 같은 코드로 받는가 | **참 — 맵 공유** | 체결 credit 을 **같은 `t.adjusted` 맵**에 같은 값(비교 as-of)으로 넣고, 해제 원인만 곁 맵 `t.filled`(같은 as-of 면 「체결이 세운 credit」)으로 가른다. 반박(call@716) · 끝 루프(B27~B29) · 깨끗한 갈래의 사용(B8 if@675)은 **한 글자도 안 바뀜** — 판정 하나. 값에 종류를 싣는 안은 `AdjustmentApplied` · `refuteUsableAdjustmentCredits` 까지 편집해 기존 함수가 다섯을 넘음(기각) |
| (q2) Observe 뒤에 준 credit 이 다음 Observe 까지 사는가, 이미 선 차단 위에서 순서가 맞는가 | **참** | 사이에서 `t.adjusted` 를 쓰는 것은 수렴의 `AdjustmentApplied`(그 종목이 다시 조정되면 덮어씀 — 최신 근거)뿐, `Refresh` 는 안 건드림. 이미 선 차단 위: 그 주기 Observe 끝 루프가 옛 조정 credit 을 반박(분쟁) → 뒤에 체결 credit 이 섬(탐침 「앞 크레딧 위 W2」) |
| (q3) 원인을 해제 블록에 싣는 통로 · 장부 쓰기 | **참 — a192 와 같은 통로, 장부 함수 새로 없음** | `Block` 에 비공개 칸 하나(a192 의 `terminalRelease` 와 같은 모양), Observe 깨끗한 갈래가 `Cleared` 에 넣을 때 표시, `persist` 가 원인 · 근거 문구를 고름. 기존 `ReleaseReconcile`(범위 + `ExpectCause` + `Cause` 검증)로 충분 |
| (q4) 새 원인의 소비자 · DDL | **참** | `release_cause TEXT` CHECK 없음(execution_contract.go:210) → 스키마 36 유지. 소비자: `ValidReconcileReleaseCause`(case 1 추가), broker-zero 두 자리(허용 목록 밖 → `released_reconciliation` 으로 fail-closed — 의도), 유효 집합 시험(열거식, 추가만), 콘솔 · cmd · 로그 0 |
| (q5) a194 보고 | **참 — 영향 없음** | a194 는 「편입 후보 보유」가 스스로 풀리는 차단 아래 10분 넘게 남을 때만 보고(reconcile_block_report.go:24 · :84-92). 체결 credit 해제는 다음 깨끗한 주기(≈1~2 주기)라 유예 안. 늦어져 보고가 나가도 문구는 「스스로 풀리지 않으면 운영자 해제가 필요할 수 있다」(:174) — 거짓 아님 |
| (q6) a200 이 인정하는 창(W4~W6)은 들어오지 않는가 | **참** | 그 창은 그 주기에 조정이 커밋됨 → `Converged` 에 그 종목 → V2 의 「조정 미반영」 거짓. 탐침 W4(수렴 31→0 뒤 Refresh 때 체결): 조정 credit 만, 체결 credit 0 |
| (q7) a192 결속 행 위의 나중 수량 불일치 | **문언 충돌 1(MODIFIED 필요) · 새 종류의 틀린 해제 0** | 정본 a192 Scenario 「수량 분쟁이 겹친 차단은 조정 규칙만 푼다」(:398-400) THEN 「조정 이벤트 뒤의 재조회 일치만 그 차단을 풀 수 있다」가 v3 와 충돌 → MODIFIED. 틀린 해제의 모양(결속 주문의 종결이 `fail_closed` 라 추적에서 빠져 비교가 깨끗해 보임)은 **조정 credit 이 오늘 같은 행을 푸는 경로와 같은 종류** — 체결 credit 이 새로 여는 것이 아님 |

## (q1) 반박 · 삭제 규칙 — 분기 열거와 같은 코드로 받는 방법

AST `cited-ast/Tracker.Observe.ast.json`(sha 9895f31184e5) · `Tracker.refuteUsableAdjustmentCredits.ast.json` · `creditUsableBy.ast.json` · `hasBlockFor.ast.json`.

| 자리 | 조건 | 하는 일 |
|---|---|---|
| 사용 — 깨끗한 갈래 B6 range@666 → B7 if@667(QM ∧ adjusted_reconcile) → **B8 if@675** `!creditUsableBy(t.adjusted[종목], diff.AsOf)` | 거짓(쓸 수 있음) → B10 if@689 `disputed[종목]` 이면 Awaiting@698, 아니면 Cleared@701 | 해제 후보 |
| 조기 반박 — blocking 갈래 B11 if@708(`pendingPermanent ∧ !earningKeyPresent`) → `refuteUsableAdjustmentCredits` call@716 → B1 range@526 → B2 if@527 `disputed ∧ creditUsableBy` | 참이면 `delete` | 권위 읽기 실패 전에 반박을 기록 |
| 끝 루프 B27 range@826 → B28 if@827 `!creditUsableBy` 이면 건너뜀 → **B29 if@841** `disputed ∨ committed ∨ !hasBlockFor` | 참이면 `delete`@842 | 반박 · 해제됨 · 블록 없음 |
| 초기화 | `Restore` :1001 · `Resolve` :1392 의 `t.adjusted = nil` | 재기동 · 운영자 해제 |

- **맵 공유안(채택 제안)**: 체결 credit 을 `t.adjusted[종목] = diff.AsOf` 로 넣고 `t.filled[종목] = diff.AsOf` 를 곁에 둠. 위 다섯 자리는 무편집으로 체결 credit 에도 그대로 적용(같은 함수 · 같은 분기 — 「판정 둘」 아님). 해제 원인은 `Cleared` 에 넣는 순간 `t.filled[종목] == t.adjusted[종목]`(같은 as-of)이면 체결 원인.
  - 한 종목에 credit 은 하나: 같은 주기 조정이 있으면 V2 (나)가 거짓이라 체결 credit 이 안 섬 → 같은 as-of 충돌 없음. 다음 주기에 그 종목이 다시 조정되면 `AdjustmentApplied` 가 더 나중 as-of 로 덮어써 `t.filled` 와 달라짐 → 조정 원인(최신 근거가 이김). 곁 맵의 낡은 항목은 체결 credit 을 줄 때 「`t.adjusted` 와 다른 것」을 지워 정리(새 메서드 안).
  - **V5(해제 순서: 조정 credit → a192 → 체결 credit)와의 관계**: 맵을 공유하면 체결 credit 의 사용 자리는 조정 credit 과 같은 B8(a192 후보 call@682 **앞**)이다. 같은 종목에 체결 credit 과 「풀 수 있는」 a192 결속이 함께 있을 수 없으므로 결과는 같다 — 체결 credit 이 서려면 그 주기 비교에 그 종목의 수량 불일치가 있어야 하고, 그 관측이 같은 종목 결속을 `tainted` 로 만듦(`updateOrderTerminalBindings` B6 range@331 → B7 if@332, call@808 — credit 부여보다 먼저 같은 Observe 안). 그 뒤 새 결속은 같은 키에 행이 이미 있어 생기지 않음(`seen` · entered=false). **V5 의 문자 그대로의 순서는 아니지만 결과가 같다**(Manager 확인 요청 — 질문 1).
- 값에 종류를 싣는 안: `t.adjusted` 의 값 타입을 바꾸면 `AdjustmentApplied`(:484) · `refuteUsableAdjustmentCredits`(:525) · Observe 두 자리를 고쳐야 함 → 기존 함수 편집이 다섯을 넘음 — 기각.
- 맵을 하나 더 두고 같은 함수를 두 번 부르는 안: `refuteUsableAdjustmentCredits` 가 `t.adjusted` 를 직접 돎(맵 인자 없음) → 인자화 편집 + Observe 끝 루프 복제 = 판정 둘 — 기각.

## (q2) 수명 — 부여에서 다음 Observe 까지

- 부여 자리: `RunOnce` 의 `Observe` 성공 뒤(B7 if@470 통과 뒤 · `judgeHoldings` call@477 앞). Observe 오류 주기는 주지 않음(보수 — 블록이 영속되지 않았을 수 있음).
- 다음 주기 `RunOnce` 안에서 `t.adjusted` 를 만지는 것: 수렴의 `AdjustmentApplied`(converge.go:189, 그 종목이 다시 조정될 때만 — 덮어씀), `Refresh`(mismatch.go:1019-1084 — `t.adjusted` 참조 0), `Observe`(위 표). `Restore` 는 루프 안에서 불리지 않음(gateway.go:279 기동 때 한 번).
- 이미 선 차단 위(앞 크레딧 위 W2): 그 주기 Observe 의 blocking 갈래 → 끝 루프 B29 if@841 `disputed` 로 옛 조정 credit 삭제 → Observe 반환 → RunOnce 가 체결 credit 부여. 탐침 `v3-probe/run.txt` 「EarlierCreditW2」: W2 뒤 `adjusted=filled={SYN201: W2 as-of}` → 다음 주기 `FILL_APPLIED`.
- 부여 조건 「`hasBlockFor` 참」은 새 메서드가 `t.mu` 아래에서 `t.blocks` 로 봄 — Observe 가 방금 넣은 pending 블록도 포함(영속 실패면 그 뒤 해제 쓰기가 실패하는 것은 오늘 조정 credit 과 같음).

## (q3) 원인 통로

- `persist` B6 range@921 → B7 if@922 `block.terminalRelease != nil` → `releaseOnOrderTerminal` call@925, 아니면 `ReleaseReconcile` call@935 의 `Cause` 가 리터럴 `ADJUSTMENT_APPLIED`(:938 [손]).
- v3: `Block` 에 비공개 칸(예: `releasedByFill bool`)을 두고 Observe 가 Cleared@701 직전에 표시 → `persist` 가 그 칸으로 `Cause` · `Evidence` 를 고름. 장부 쪽 `ReleaseReconcile`(reconcile_states.go:252-)은 `ValidReconcileReleaseCause` 로 원인을 검사하고 범위 + `ExpectCause` 로 활성 행을 tx 안에서 다시 찾음 — 새 원인 상수만 유효 집합에 있으면 충분. 새 장부 함수 불필요. 탐침에서 행이 `released:FILL_APPLIED` 로 기록됨.

## (q4) 새 원인 소비자

| 자리 | 귀결 |
|---|---|
| `ValidReconcileReleaseCause`(reconcile_states.go:111-119, B2 case@113) | case 에 새 상수 1 — 없으면 `ReleaseReconcile` 가 `ErrInvalidRequest` |
| `recordRiskBucketBrokerZeroObservation` :280 | 최신 해제 QM 행 원인이 RECHECK_MATCHED · ADJUSTMENT_APPLIED 가 아니면 `released_reconciliation` 거절 → 새 원인은 거절(fail-closed, V6 의도). 생산 호출자 0(함수 주석 :225-227) |
| `loadRiskBucketBrokerZeroAuthority` :677 | 같은 허용 목록 → 새 원인 거절. 호출자 `releaseRiskBucketOwner` 는 구조 시험이 생산 호출 금지 |
| 시험 | reconcile_states_test.go:349-356 유효 집합 열거(추가 필요 — 무추가여도 안 깨짐), 거절 목록 :338(새 상수와 겹치지 않음) |
| 콘솔 · cmd/tossctl · 로그 · a194 | release cause 를 읽는 생산 코드 0(v3-evidence Q5 grep) |
| DDL | `release_cause TEXT`(execution_contract.go:210) — CHECK 없음, 이후 마이그레이션에 CHECK 추가 0(grep) |

## (q5) a194

- 분류: 편입 후보 보유가 대사 차단 아래이면 운영자 전용 차단은 즉시, 그 밖(QM · adjusted_reconcile)은 `blockedHoldingGrace` 10분 뒤 `blocked_stale`(reconcile_block_report.go:24 · :84-92). 체결 credit 해제는 다음 깨끗한 비교에서 일어나므로 유예 안에 끝남. 비교가 다른 종목 분쟁으로 계속 막혀 10분을 넘기면 보고가 나가고, 문구(:174-176)는 가능성으로 안내 — 거짓이 되지 않음.

## (q6) a200 창

- a200 의 인정은 「조정이 닫은 인스턴스에 온 청산 체결」(W4~W6)이다 — 정의상 그 주기에 조정이 커밋됨 → `ConvergeReport.Converged` 에 그 종목(converge.go:237 append) → V2 (나) 거짓 → 체결 credit 없음.
- 탐침(이 브랜치 하네스 — a200 의 reconcile 시험은 수렴 → Refresh → Observe 를 손으로 밟아 RunOnce 의 부여 자리를 지나지 않으므로 하네스로는 이쪽을 씀): W4(조정 31→0 커밋 뒤 Refresh 때 체결) → `adjusted={SYN201}`, `filled={}`. 그 체결은 이 픽스처에서 엔진 청산 발의 주문이 아니라 a200 이 인정하지 않고 `ATTRIBUTION_FAILED`(운영자 전용) — base 와 같은 결과로 본 것은 아님(구현 RED 에서 base 대조).
- a200 의 시험 파일(`internal/reconcile/a200_*_test.go` · `internal/journal/a200_*_test.go`)은 v3 가 고칠 함수(Observe · persist)를 지나므로 구현 단계 선택 시험에 넣는다.

## (q7) a192 결속 행과의 상호작용

- 모양: 이 프로세스가 missing-order 로 삽입 · 결속한 행(같은 키 = 종목 범위) → 나중 주기에 같은 종목 수량 불일치 → `updateOrderTerminalBindings` 가 결속을 `tainted`(a192 C3 → `quantity_dispute_seen` 로 영원히 유지) → 같은 주기에 V2 가 맞으면(그 행은 QM ∧ adjusted_reconcile 이라 `hasBlockFor` 참) 체결 credit → 다음 깨끗한 비교에서 `FILL_APPLIED` 해제. 탐침 「A192BoundRow」에서 그대로 관측.
- 정본 문언: a192 Requirement 본문 :383-384 「같은 종목의 수량 불일치 · 외부 포지션이 관측됐으면 **이 경로로** 해제해서는 안 된다」는 경로 한정이라 충돌 없음. 그러나 Scenario 「수량 분쟁이 겹친 차단은 조정 규칙만 푼다」(:398-400) THEN 「이 경로로는 해제되지 않고 **조정 이벤트 뒤의 재조회 일치만** 그 차단을 풀 수 있다」는 v3 와 충돌 → MODIFIED(「조정 이벤트 또는 체결 credit 뒤의 재조회 일치」).
- 틀린 해제의 입력: 결속 주문 O 의 종결이 `fail_closed`(신뢰할 수 없는 종결)라 `TrackedFillOrders`(`terminal = 0` 조건)에서 빠지고 계좌 목록에도 없어 비교가 깨끗 → 체결 credit 으로 해제(a192 라면 C7 로 거절). **오늘 조정 credit 이 같은 행을 같은 방법으로 푼다**(그 종목이 조정되면 다음 깨끗한 비교에서 `ADJUSTMENT_APPLIED`) — 수량 분쟁이 겹친 missing-order 행을 a192 가 「조정 규칙만」에 맡긴 순간부터 있는 성질. v3 는 그 규칙에 근거를 하나 더할 뿐 새 종류를 열지 않음. 설계 변경 불요 — V7 대응표의 「같은 보증」.
