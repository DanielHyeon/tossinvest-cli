# a201 design v3 — 「체결이 조정을 대신했다」는 credit

재개(사용자 지시 2026-11-24 「1번으로 진행해줘」 — 해제 쪽으로). v1 · v2 는 차단을 **세우지 않는** 쪽이라 생략된 관측의 증거를 버려 기각됐다(아래 「기각한 대안」). v3 는 **차단을 지금과 똑같이 세우고**(그 주기의 `Observe` 는 한 글자도 다르게 돌지 않는다) 해제 쪽에 근거 하나를 더한다. 새 상태기계를 만들지 않고 추적기에 이미 있는 **조정 credit 의 모양과 코드를 그대로** 쓴다.

증거: `analysis/v3-evidence/evidence.md`(Q1~Q8) · `analysis/claim-check-v3.md`(전제 (q1)~(q7)) · 탐침 `analysis/v3-probe/` · `analysis/v3-evidence/probe-*`.

## 1. 결정

| # | 결정 | 근거 |
|---|---|---|
| V1 | **무엇**: 수렴이 그 종목의 조정을 쓰지 못한 이유가 「체결이 먼저 그 감소를 반영했기 때문」인 주기에, 조정 credit 과 같은 종류의 **체결 credit** 을 그 종목에 준다. 조정 credit 이 「조정을 반영했으니 그 뒤의 새 비교가 깨끗하면 푼다」이듯, 체결 credit 은 「체결이 대신 반영했으니 그 뒤의 새 비교가 깨끗하면 푼다」 | W2 · W3 의 고착은 「자동 해제 근거(조정)가 없어서」다 — 근거를 하나 더하면 차단 쪽 증거(결속 · 연속 횟수 · 크레딧 반박 · 건강 계수)를 하나도 버리지 않는다. 판정이 실제로 보는 것은 「로컬 읽기 직전 표식 뒤 커밋(이미 로컬에 반영된 것 포함)」이다(V2 (b) · V8-4) |
| V2 | **주는 조건(닫힌 집합, 전부)**: (a) 그 주기 비교(`diff.Quantities`)에 그 종목의 수량 불일치가 있다 (b) 그 종목의 체결 워터마크(`FillWatermark`)가 **로컬 읽기 전에 읽은 전역 최대 체결 id** 보다 크다 (c) 이번 주기 수렴이 그 종목에 조정을 반영하지 않았다(`ConvergeReport.Converged` 에 없음 — 거절 · stale · 판 중단 모두 포함) (d) 그 주기 `Observe` 가 끝난 뒤 그 종목에 자동 해제 대상 수량 불일치 차단(`hasBlockFor` — QM ∧ adjusted_reconcile)이 서 있다. 판독(전역 · 종목 워터마크)이 실패하면 주지 않는다 — 구현은 주기 단위 전부-아니면-0(종목 워터마크 하나라도 못 읽으면 그 주기 아무 종목에도 주지 않음, 구현 리뷰 r1 R3 Manager 수용) | (b) 는 v2 D2' 와 같은 워터마크(로컬 읽기 전 — 뒤에 읽으면 그 사이 커밋된 체결을 놓침; id 순서 = 커밋 순서, 단일 연결 journal.go:174). (c) 가 a200 이 기대는 W4~W6 을 빼냄((q6)). (d) 는 credit 의 수명 규칙(블록 없으면 소멸)과 같은 조건을 부여 때도 씀 |
| V3 | **주는 자리**: `RunOnce`, **`Observe` 가 성공해 돌아온 뒤**(편입 판정 앞). credit 의 비교 시각은 그 주기의 `diff.AsOf` | 그 주기 Observe 는 credit 을 모르고 돈다 → 차단 · 결속 · 연속 횟수 · 반박 · 건강 계수가 편집 전과 같음. 같은 as-of 라 그 주기에 쓰이지도 반박되지도 않음(`creditUsableBy` 엄격히 나중). Observe 오류 주기는 주지 않음(블록이 영속되지 않았을 수 있음 — 보수) |
| V4 | **쓰는 규칙 = 조정 credit 과 같다 — 같은 코드**: 체결 credit 은 **조정 credit 과 같은 맵(`Tracker.adjusted`)에 같은 값(비교 as-of)** 으로 들어간다. 그래서 「엄격히 나중 비교가 그 종목을 분쟁으로 안 볼 때만 쓴다(B8 if@675 · B10 if@689)」 「나중 비교가 분쟁으로 보면 반박되어 사라진다(call@716 · B29 if@841)」 「해제되면 · 블록이 없어지면 사라진다(B29)」 「재기동 · 운영자 해제에서 복원되지 않는다(`Restore` · `Resolve` 의 `adjusted = nil`)」가 **한 글자도 바뀌지 않은 기존 분기**로 적용된다. 해제 원인만 곁 맵(`Tracker.filled` — 같은 as-of 면 체결이 세운 credit)으로 가른다 | 판정 둘 금지. 값에 종류를 싣거나 맵을 둘로 나누면 기존 함수(`AdjustmentApplied` · `refuteUsableAdjustmentCredits`)를 고치거나 끝 루프를 복제해야 함((q1)) |
| V4b | **깨끗한 갈래의 권위 오류 반환 직전 반박(codex v3 r1 V3R1-1, Manager 결정 1)**: `Observe` 깨끗한 갈래의 `resolvePendingPermanentAuthority`(call@659) 가 오류일 때, `finishAuthorityError` 로 돌아가기(return@660) **직전에** `refuteUsableAdjustmentCredits(disputed, diff.AsOf)` 를 부른다(그 함수 무편집 — 차단 갈래 call@716 과 같은 호출). 오류 없는 경로는 한 글자도 안 바뀜 | 깨끗한 비교도 그 종목을 **외부 포지션**으로 분쟁으로 볼 수 있다(`symbolsInDispute` 는 외부 포지션을 셈, `BlocksEntry` 는 안 셈). 대기 중 영구 승격의 권위 확인이 실패하면 지금은 Awaiting 처분도 끝 루프 반박도 없이 반환해, 명시적으로 관측된 분쟁이 credit 을 소멸시키지 못한다(오늘 조정 credit 에도 잠재). 오류 갈래 안에만 두는 이유: 오류 없는 경로에서는 그 credit 이 B10(분쟁 → Awaiting) · B29(분쟁 → 삭제)로 이미 같은 결과를 내므로 앞에서 반박하면 그 블록이 B8 → a192 후보(call@682)로 갈래가 바뀌어(결과는 Awaiting 으로 같지만 결속의 보류 사유 · 장부 읽기가 달라짐) 「다른 분기 결과 무변」을 지키지 못함 |
| V4c | **Refresh 실패로 Observe 앞에서 끝나는 주기의 반박(gstack r1 G1, Manager 결정)**: `RunOnce` 의 `Tracker.Refresh` 오류 갈래에서 `return cycle` **직전에** 새 파일의 `Tracker.RefuteCreditsDisputedBy(diff)` 1 호출 — 잠금 아래 `refuteUsableAdjustmentCredits(symbolsInDispute(diff), diff.AsOf)` 만 부름(두 함수 · `Refresh` 무편집). 오류 없는 경로는 한 글자도 안 바뀜 | V4b 와 같은 원리의 둘째 조기 반환 — 그 주기 비교가 credit 종목을 분쟁(외부 포지션 등)으로 봐도 Observe 의 반박 자리가 하나도 돌지 않아 credit 이 살아남고, 다음 깨끗한 비교가 그 credit 으로 풀었다(오류 없는 경로의 k+1 은 Awaiting + B29 삭제라 이것이 더 느슨). base `35d28a535` 의 조정 credit 에도 같음(측정 `analysis/review/g1-base/base-run.txt` — 해제됨). 엄격히 나중만 지우므로 그 주기 수렴이 준 같은 as-of 조정 credit 은 남음(시험으로 고정) |
| V5 | **해제 판정의 순서(사실대로)**: 체결 credit 은 조정 credit 과 같은 맵에 서므로 **조정 credit 의 자리(B8)** 에서 쓰인다. a192 후보(call@682)보다 앞이지만, 체결 credit 이 서는 주기에 같은 종목의 a192 결속은 그 수량 불일치로 이미 `tainted`(`updateOrderTerminalBindings` call@808, 부여보다 먼저)라 a192 가 풀 수 있었던 행을 가로채지 않는다. 수량 흔적 없는 a192 결속 행에는 V2 (a) 가 거짓이라 체결 credit 이 서지 않는다. 한 종목에 credit 은 하나(같은 주기 조정이 있으면 V2 (c) 거짓, 다음 주기에 다시 조정되면 더 나중 as-of 로 덮어써 조정 원인이 이김) | Manager 결정 4 — 「가로채지 않는다」를 시험으로 고정(§3-13) |
| V5b | **원인 표지의 등식(Manager 결정 2)**: 곁 맵 `Tracker.filled[종목]` 의 값은 **부여 때의 as-of**. 해제 블록의 원인은 `filled[종목] != "" ∧ filled[종목] == adjusted[종목]`(= 지금 공유 맵에 있는 credit 이 부여한 그것) 일 때만 `FILL_APPLIED`, 아니면 `ADJUSTMENT_APPLIED`. `AdjustmentApplied` 무편집 | `adjusted` 의 쓰기 자리(`AdjustmentApplied` 덮어씀 · 반박 · 끝 루프 삭제 · `Restore`/`Resolve` nil · 부여 메서드) 가운데 부여 메서드 밖의 것은 값을 바꾸거나 지움 → 등식 거짓 → 조정 원인. `filled` 는 지워지지 않고 남아도 등식이 거짓이라 틀린 표지가 안 남(부여 메서드가 다른 값의 항목을 정리). 남는 모호함은 §4 |
| V6 | **해제 원인은 새 상수 `FILL_APPLIED`**(`ADJUSTMENT_APPLIED` 와 짝 — 상수 이름 `journal.ReconcileReleaseFillApplied`). 근거 문구는 「로컬 읽기 직전 표식 뒤 커밋된 체결이 조정을 대신했고 그 뒤의 대사가 일치했다」(`a fill committed after the mark read before the local read stood in for the adjustment and the reconciliation after it agreed`). `RECHECK_MATCHED` 를 쓰지 않는다 | 위험 버킷 broker-zero 경로가 `RECHECK_MATCHED` 행을 조정 증명 없이 「보유 0 의 근거」로 받는데(증거 Q5), 체결 credit 해제는 보유 0 을 뜻하지 않음. 새 원인은 그 경로에서 `released_reconciliation` 으로 거절(fail-closed). `release_cause TEXT` 에 CHECK 없음 → 스키마 36 유지((q4)) |
| V7 | **두지 않는 것**: (a) 행 id 고정 · 삽입 증명(a192 C1 · C2) — 조정 credit 도 종목 단위이고, 삽입 증명을 요구하면 「이미 걸려 있던 차단 위의 같은 경합」을 못 닫음. (b) credit 뒤의 그 종목 체결이면 풀지 않음(a192 C12) — 청산 주문이 여러 조각으로 체결되는 평범한 경우를 전부 고착시킴. (c) 무장-후-해제(C13) — 구조상 불필요(증거 Q3) | v3 가 지는 보증 = 조정 credit 해제가 오늘 지는 보증(아래 대응표) |

### 조정 credit ↔ 체결 credit 대응표

| 성질 | 조정 credit(오늘) | 체결 credit(v3) |
|---|---|---|
| 근거 | 그 주기 수렴이 그 종목 조정을 커밋함(`AdjustmentApplied`, converge.go:189) | 그 종목 체결이 로컬 읽기 직전 표식 뒤 커밋됨(이미 로컬에 반영된 것 포함 — 종목 워터마크 > 로컬 읽기 전 전역 최대) ∧ 그 주기 조정 없음 — **이것 하나가 다름** |
| 단위 · 저장 | 종목 → 비교 as-of, 프로세스 기억(`Tracker.adjusted`) | 같은 맵 · 같은 값 (+ 원인 곁 맵 `Tracker.filled`) |
| 부여 시각 | 수렴(Observe 앞) — 같은 as-of 라 그 주기엔 못 씀 | Observe 뒤 — 같은 as-of 라 그 주기엔 못 씀 |
| 사용 | 엄격히 나중 비교 ∧ 비교 전체가 진입 안 막음 ∧ 그 종목 분쟁 아님(B6~B10) | 같음(같은 분기) |
| 반박 | 나중 비교가 그 종목 분쟁(B29 · call@716) + V4b(깨끗한 비교의 권위 오류 직전) + V4c(Refresh 실패 주기의 반환 직전) — V4b · V4c 는 이 change 가 더함 | 같음(같은 호출) |
| 소멸 | 해제 커밋 · 블록 없음(B29) · 재기동 · 운영자 해제 | 같음 |
| 해제 원인 | `ADJUSTMENT_APPLIED` | `FILL_APPLIED` |
| 블록 범위 | QM ∧ adjusted_reconcile(종목) — 영구 · 계좌 범위는 안 풂(B7 if@667) | 같음 |
| 행 id · 삽입 증명 · 뒤 체결 실격 | 없음 | 없음(V7) |

## 2. fail-closed 경계 · 이 설계가 새로 푸는 것(공시)

- **차단 쪽은 무변**: 그 주기 `Observe` 의 입력 · 순서 · 호출은 편집 전과 같다. 새로 바뀌는 것은 「그 뒤의 깨끗한 비교에서 Awaiting 으로 가던 블록이 Cleared 로 간다」(그 종목에 체결 credit 이 있을 때)뿐.
- **이 change 가 기존 경로에서 바꾸는 동작은 둘(V4b · V4c)이다.** 둘째(V4c): 대사 주기가 Refresh 실패로 Observe 앞에서 끝나도 그 주기 비교가 분쟁으로 본 종목의 엄격히 나중 credit(**오늘의 조정 credit 포함**)이 반박된다 — 지금은 살아남아 다음 깨끗한 비교에서 풀린다(base 측정 해제됨). 같은 보수 방향, 오류 없는 주기 무변. 첫째(V4b): 깨끗한 비교가 그 종목을 외부 포지션으로 분쟁으로 보는데 대기 중 영구 승격의 권위 확인이 실패하는 주기에, 그 종목의 쓸 수 있는 credit(**오늘의 조정 credit 포함**)이 반박된다 — 지금은 살아남아 뒤의 일치로 풀릴 수 있다. 반박이 늘어 자동 해제가 줄어드는 보수 방향. 오류 없는 주기는 무변(그 credit 은 지금도 B10 · B29 로 같은 결과). 반박으로 블록이 남으면 그 종목의 편입이 계속 막히고, 청산은 RECONCILE 확정 하한으로 제한된다(하한 조회 실패 시 0주) — 오류 없는 같은 입력에서 지금도 그런 것과 같은 상태가 된다(구현 리뷰 r2 S3 — `internal/app/engine/exitwiring.go:192-201` `reconcileFloor.ConfirmedFloor` 는 그 종목 블록이 있으면 하한을 적용, `internal/app/engine/exitloop.go:1767-1779` `ExitObserver.applyFloor` 는 하한 계산 실패를 0주로 깎음 — HEAD `7621c5c2e` 에서 확인). 정본 :57 의 같은 규칙(차단 비교의 권위 오류 전 반박)을 깨끗한 비교의 분쟁까지 넓히는 MODIFIED.
- **V8 — 정본 :77 의 원칙(「조정도 종결 증명도 없이 우연히 일치한 단발 관측은 차단을 해제하지 못한다」)이 좁아지는 자리**: 체결 credit 이 증명하는 것은 「그 종목 체결이 대사의 로컬 읽기 직전 표식 뒤에 커밋됐고(이미 로컬에 반영된 것 포함) 그 종목이 조정되지 않았다」까지다 — 불일치가 그 체결 **때문이었다**는 증명이 아니다. 그래서
  1. 그 주기에 실제 외부 분쟁이 섞여 있었고 뒤에 외부 거래가 상쇄해 새 비교가 일치하면 풀린다(증거 탐침 R1-1 변형 `probe-r11-offset.txt` — 오늘은 운영자 해제까지 유지).
  2. 종목 워터마크에 계좌 조건이 없어(position_adjustments.go:190-198) 다른 계좌의 같은 종목 체결도 credit 을 세운다(운영이 단일 계좌인지 미확인).
  3. 수렴이 「거절」한 종목(복수 시장 등 `soleLiveMarket`)도 「미반영」에 든다.
  4. 표식과 로컬 읽기 사이, 또는 로컬 읽기 도중 커밋돼 그 주기 로컬에 이미 반영된 체결도 「표식 뒤」에 든다(구현 리뷰 r1 R1) — 그 주기 비교가 그 체결을 이미 본 경우에도 다른 이유로 조정이 없고 차단이 섰으면 credit 이 선다.
  - 셋 다 **그 뒤의 새 스냅숏 · 새 로컬 읽기의 비교가 깨끗해야만** 풀린다. 틀릴 수 있는 것은 「그 순간 수량은 맞지만 원인을 사람이 보지 않았다」 — 운영자 해제 명령(a195)이 요구하는 조건(안정 재조회가 진입을 막는 차이를 보지 않음)과 같은 상태다.
- **R1-1(codex r1) 순서는 닫힌 채다**: 실제 차이가 있는 주기(스냅숏 10 · 로컬 15)는 비교가 진입을 막아 깨끗한 갈래에 못 가므로 해제 없음(증거 탐침 R1-1 k+1).
- **a192 결속 행((q7))**: missing-order 로 삽입 · 결속된 행에 나중 수량 불일치가 오면 결속은 `tainted`(a192 는 「조정 규칙만」에 맡김) — v3 는 그 규칙에 근거를 하나 더한다. 결속 주문의 종결이 `fail_closed` 라 추적에서 빠지고 비교가 깨끗해 보이는 입력에서 체결 credit 으로 풀리는 것은 **오늘 조정 credit 이 같은 행을 푸는 경로와 같은 종류**.
- **계속 credit 이 없는 입력**(각각 시험): 워터마크가 안 움직인 종목 · 그 주기에 조정이 반영된 종목(W4~W6) · 판독 실패 · 그 종목에 자동 해제 대상 블록이 없음(운영자 전용 · 다른 원인) · 빈 종목 · Observe 오류 주기.

## 3. 시험 (RED 의 뼈대 — 실제 `RunOnce`, 1단계 하네스)

1단계 재현 6개를 v3 계약으로 뒤집는다(픽스처의 열린 주문 목록은 매 주기 비우고 다시 채움 — 1단계 「중복 행」 결함).
1. W2 · W3-a · W3-b(합성: 다른 계좌의 같은 종목 체결이 워터마크만 움직이고 엔진 체결은 그 뒤) · 이미 걸린 차단 위의 경합(앞 크레딧 위 W2) → 그 주기에는 차단이 **선다**(지금과 같음) → 다음 깨끗한 주기에 `FILL_APPLIED` 로 풀리고 진입 · 재매수 편입 허용.
2. R1-1 순서 → 실제 차이가 있는 주기에는 해제 없음, 그 뒤 수렴 · 조정 credit 경로(`ADJUSTMENT_APPLIED`).
3. R1-1 변형(상쇄) → 풀린다(V8-1 공시된 동작으로 고정).
4. 나중 비교가 그 종목을 분쟁으로 보면 credit 소멸 · 그 뒤 깨끗해져도 안 풀림.
5. credit 없음: 워터마크 안 움직임 · 조정이 반영된 종목 · 판독 실패(함수 단위 — 닫힌 장부) · 블록 없는 종목 · Observe 오류 주기.
6. 재기동(`Restore`) 뒤 → 안 풀림.
7. **그 주기의 `Observe` 가 편집 전과 같은 결과**: 차단 · 결속 · 연속 횟수 · 건강 계수 — R2-1 의 두 반례 입력(다른 종목 missing-order 결속 추가 · 비연속 승격 단절)과 R2-2 입력(편입 가격 조회 실패 연속)을 실제로 돌려 편집 전과 같음.
8. a192 종결 증명 · 조정 credit 경로 무변(기존 시험 + W4~W6 를 이 하네스로 base 와 같은 결과).
9. 다른 종목 분쟁이 있으면 해제 대기(Awaiting), 그 분쟁이 사라진 첫 깨끗한 비교에서 해제.
10. broker-zero 경로가 `FILL_APPLIED` 행을 `released_reconciliation` 으로 거절.
11. 구조: 부여 호출 1자리(Observe 성공 뒤), 체결 credit 이 `Tracker.adjusted` 를 쓰는 기존 분기로만 쓰이고 반박 · 삭제됨(새 반박 · 삭제 분기 0 — V4b 는 기존 `refuteUsableAdjustmentCredits` 호출 1 자리 추가).
12. **V4b**: 대기 중 영구 승격(쓰기 실패) + 그 종목이 외부 포지션으로만 관측되는 깨끗한 비교 + Observe 안 권위 확인 실패 → 편집 전에는 credit 이 살아남아 뒤의 일치에서 풀림(조정 credit 으로 재현), 편집 뒤에는 반박됨 — **조정 credit 판 · 체결 credit 판 둘 다**. 오류 없는 같은 입력은 편집 전후 결과 같음.
13. **V5 「가로채지 않음」**: a192 결속이 있고 수량 흔적이 없는 행 → 체결 credit 이 서지 않고 a192 경로(`ORDER_TERMINAL_CONFIRMED`)로 풀림. 결속 행에 수량 불일치가 온 경우 → 결속 `tainted`, 체결 credit 으로 `FILL_APPLIED`.
14. **V5b 등식**: 체결 credit 뒤 그 종목에 조정이 커밋되면(더 나중 as-of 로 덮어씀) 해제 원인은 `ADJUSTMENT_APPLIED`; 체결 credit 이 반박 · 해제로 지워진 뒤 남은 `filled` 항목이 다음 조정 credit 해제의 원인을 바꾸지 않음.
15. 변이(사본): V2 (a)(b)(c)(d) 각각 삭제 · 워터마크를 로컬 읽기 뒤에(수렴 뒤) 읽기 · 종목별이 아닌 전역 워터마크로 판정 · 판독 실패 시 부여 · Observe 앞에서 부여 · Observe 오류 주기에도 부여 · 원인 표지 반대로(체결 → ADJUSTMENT) · 원인 등식을 「`filled` 에 있음」으로(값 비교 삭제) · `ValidReconcileReleaseCause` 에서 새 원인 삭제 · V4b 반박 호출 삭제 · V4b 를 오류 갈래 밖(권위 확인 앞, 항상)으로 옮김(오류 없는 경로 결과가 달라지는지 — 결속 보류 사유).
16. **V4c**(gstack r1 G1): Refresh 실패 주기의 비교가 credit 종목을 외부 포지션으로 보면 그 credit 소멸 → 다음 깨끗한 비교에서 안 풀림 — 체결 credit 판 · 조정 credit 판 둘 다. 대조: 그 비교가 그 종목을 분쟁으로 안 보면 credit 유지 → 해제. 같은 주기 수렴이 준 같은 as-of 조정 credit 은 Refresh 실패에 안 지워짐. 구조: 호출은 Refresh 오류 갈래 return 직전 1자리, 메서드는 잠금 아래 기존 반박만. 변이: 호출 삭제 · 호출을 Refresh 앞(무조건)으로 옮김.

## 4. 남는 것(이 change 가 닫지 않음)

- **판 중단 뒤 종목 B**: 체결이 credit 계산(Observe 뒤) 뒤 · 다음 로컬 읽기 전에 오면 (b) 가 거짓 → credit 없음 → 고착 그대로(증거 탐침). 별도 change 후보: 「수렴이 stale 한 종목에서 판 전체를 멈추지 않는다」(converge.go:228-229).
- 재기동(credit 소실 — 조정 credit 과 같음) · 체결 행 없이 투영만 쓰는 전략 진입 backfill(strategy_dispatch_runtime.go:1348 — 워터마크 안 움직임, 운영에서 켜져 있는지 미확인) · 계좌 범위(빈 종목) · 워터마크 읽기 실패 · Observe 오류 주기.
- 「외부 청산」 알림의 동기 전송이 뒤 종목의 창을 넓힘(후보로만 — 알림을 루프 뒤로).
- 운영 이력에 이 고착의 흔적은 0 이다(`ops-aggregates.md`) — 효과는 운영에서 `FILL_APPLIED` 해제 행으로만 보인다.
- **원인 이름의 모호함(V5b 공시)**: 서로 다른 두 주기의 비교 as-of 가 글자까지 같으면(벽시계 역행 · 같은 스냅숏 재사용 — as-of 는 초 단위) 체결 credit 뒤의 조정 credit 이 같은 값으로 덮어써 등식이 참으로 남고 해제 원인이 `FILL_APPLIED` 로 기록될 수 있다. 틀리는 것은 **원인 이름 하나**이고 해제 자체는 조정 credit 이 정당화한다(위험 버킷 broker-zero 경로는 그 행을 거절 — 보수 방향). 재기동 뒤 낡은 `filled` 항목에 같은 글자의 as-of 조정 credit 이 서는 것도 같은 경우.

## 5. 편집 표면 · 결합

- 편집할 기존 함수 **넷**: `ReconcileDriver.RunOnce`(로컬 읽기 전 워터마크 1문장 · Observe 성공 뒤 부여 1자리 · Refresh 오류 갈래 return 직전 V4c 반박 1문장 — 3자리), `Tracker.Observe`(깨끗한 갈래 Cleared@701 직전 원인 표지 1문장 · B5 if@659 오류 갈래 안 return@660 직전 `refuteUsableAdjustmentCredits` 호출 1 — V4b), `Tracker.persist`(해제 원인 · 근거 문구 갈래), `journal.ValidReconcileReleaseCause`(case 1). `Restore` 무편집(`adjusted = nil` 이 체결 credit 도 지움, 곁 맵은 대조로만 쓰여 낡은 항목이 해제를 만들지 못함 — 부여 때 정리).
- 구조체 칸: `Tracker.filled map[string]string`, `Block` 의 비공개 원인 표지 1(a192 `terminalRelease` 와 같은 모양).
- 새 파일: `internal/reconcile/` — 로컬 읽기 전 워터마크 표식 타입 · 부여 대상 종목 계산 함수 · `Tracker` 의 부여 메서드 · V4c 반박 메서드 `Tracker.RefuteCreditsDisputedBy`. `internal/journal/` — 새 원인 상수.
- 무편집: `ConvergeQuantities` · `Compare` · `AdjustmentApplied` · `refuteUsableAdjustmentCredits` · `creditUsableBy` · `hasBlockFor` · `orderTerminalCandidate` · `enterReconcileWithBinding` · `buildGateway`(a166 census) · `logCycle` · `note` · 스키마(36).
- 고정 장치: `a071_readiness_assembly_test.go`(reconcileloop.go 금지 문자열 — 비적중), a166 census(배선 무편집 — 비적중), a192 아카이브 FLM 번들 함수(Observe · persist · ValidReconcileReleaseCause · RunOnce)를 다시 고침 → 이 change 의 새 번들(Observe 는 편집 지점 뒤 분기 번호가 밀림 — 구현 때 옛/새 ast 정렬 재번호).
- a199(`reconcileloop.go` 의 `Context.ReconcileDriver`) — 같은 파일, 다른 함수. 구현 전에 origin/main 을 다시 본다.

## 기각한 대안

### v1 — 「낡은 불일치를 비교에서 뺀다」(1단계 · codex r1)

- 조건 (다) 「다시 읽은 로컬 = 그 불일치의 계좌 수량」이면 그 불일치(와 추적이 끝난 그 종목 missing-order)를 비교에서 빼고 깨끗해진 비교로 그 주기에 해제까지 진행.
- R1-1(P1): 다시 읽은 로컬을 **옛 스냅숏**과 맞춰 본다 — 스냅숏 뒤 체결 + 외부 매도에서 세 조건이 참인데 진짜 불일치가 남고, 앞 크레딧으로 그 주기에 잘못 푼다(반복하면 여러 주기 — R1-3). R1-2(P1): 「추적 아님」 판독이 정정 lineage(compare.go:225-235 덮어쓰기)와 `fail_closed` 종결에서 missing-order 를 잘못 뺀다. 원문 `analysis/review/proposal-freeze-r1-codex.md`.

### v2 — 「체결이 지나간 주기 전체를 보류한다」(codex r2)

- 수량 불일치 종목 중 워터마크가 움직였고 수렴이 반영하지 않은 종목이 있으면 그 주기의 Observe · 편입 판정을 건너뜀(연속 상한 1). 전문은 커밋 `c0ca86df5` 의 design.md.
- R2-1(P1): 주기 전체를 버리면 다른 종목의 증거를 버린다 — 같은 종목 비종결 주문의 a192 결속 추가 누락(정본 :381), 중간 비교의 streak 단절 누락으로 비연속 영구 승격(정본 :53). R2-2(P1): 보류 주기를 성공으로 세면 교대 보류로 지속 장애의 운영 모드 강화가 가려진다. R2-3~R2-5(P2): Refresh 선행 · 연속 상한의 범위 · ADDED 결함. 원문 `analysis/review/proposal-freeze-r2-codex.md`, (p2) 정정 `analysis/claim-check-v2.md`.

### 종목 단위 보류

- 보류 조건에 걸린 종목의 항목만 그 주기에서 판정하지 않는 안 — v1 의 「뺀다」로 돌아가고 같은 종목의 a192 결속 · 크레딧 반박이 빠진다. v1 · v2 와 같은 뿌리(생략된 관측이 남겼어야 할 증거를 버림).

## 6. 이력 — 보류와 재개

- 2026-11-24 Manager 보류 결정(`issues.md` I1). 같은 날 사용자 지시로 재개 — 재개 방향(해제 쪽)은 보류 때 적은 후보를 따르되, 후보의 `RECHECK_MATCHED` 는 V6 이유로 새 원인으로 바꿈. codex 회차는 새로 센다(v3 동결 리뷰 r1 부터).
