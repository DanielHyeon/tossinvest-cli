# a201 review

## 1단계(등록 · 재현 · 증거) — 2026-11-24, Opus 구현 팀메이트

- worktree `~/tossos-work/a201-impl`, 브랜치 `fix/a201-stale-quantity-comparison`(로컬 커밋만, push 없음). base `f9b818a0e`(origin/main `7fec501f1` + 등록 문서).
- 끝낸 것: 1.1 등록 · 1.2 base · 1.3 사전 증거 재확인(AST 20 바이트 동일 + 새 16) · 1.4 실제 `RunOnce` 재현(6 PASS) · 1.5 전제 · 1.6 CodeGraph 1.6.0 · 1.7 편집 표면 · 1.8 잔여 창 · 1.9 ADDED 1(strict valid) · 1.10 RunOnce pre-edit 번들(`check_analysis` rc 0) · 1.11 claim check.
- **멈춤**: claim check 에서 design 의 결정이 사실과 어긋남(C-D6 · C-D4 · C-R1 · C-D5 — `analysis/claim-check.md`). 지시대로 고치지 않고 올림. **1.12(codex 제안 동결 리뷰)는 실행하지 않음** — 곧 바뀔 design 을 리뷰하는 것이 되므로 Manager 결정 뒤 그 판본으로 돌리는 것이 맞다고 판단.
- 사실 정정만 문서에 반영(결정 무변경): design D3 (가) 근거 문장(C-D3a), §2 정상 입력 1(C-W3b), §4 첫 항목(C-S4 · C-R1 참조), proposal Why(C-P6). 모두 「(1단계 정정: …)」 표지.

### 재현(1.4) 요지

- 끼우는 방법: production 편집 없이 `Converger.Journal`(ConvergeStore) · `Tracker.Journal`(ReconcileStore) 인터페이스 필드를 감싸 정해진 지점에서 체결 커밋. 운영 리터럴과 같은 훅 · 필드.
- W2 · W3-a · tx 직전 · W3-b 모두 그 주기 QM 1 → 깨끗한 5주기 뒤 미해제 · 진입 거절, W2 뒤 재매수 편입 거절.
- 추가 발견 셋: (b) 단독 stale 은 수량 불변 fill_events 행이 필요(엔진 체결 자신은 (a)), 앞 주기 크레딧 위의 W2 도 고착(크레딧 반박), 판 중단 뒤 종목의 체결이 Observe 뒤면 고착(재판정이 못 잡음).

### 탐침(1.5 (e)(f)) 요지 — scratch 사본, 추적 파일 무편집

| 모양 | D5 끔 | D5 켬 |
|---|---|---|
| W2 · W3 넷 | 그 주기 missing 블록 · 2주기 뒤 ORDER_TERMINAL_CONFIRMED | 행 0 |
| 앞 크레딧 위 W2 | 고착 그대로 | 그 주기 ADJUSTMENT_APPLIED 해제 |
| 판 중단 뒤 종목 B | B 고착 | B 고착 |

### Manager 결정 요청

`analysis/claim-check.md` 의 「결정에 닿는 것」 네 줄과 정본 해석 셋(`analysis/spec-contradictions.md`). 질문 목록은 1단계 보고에 있음.

## Manager 결정 1~6 반영(2026-11-24)

- 결정 1~5 를 design · proposal · delta · Story 에 반영(커밋 `4c9d53138`). 결정 6 에 따라 codex 1회차를 그 판본으로 돌림.

## proposal-freeze r1 — codex(2026-11-24)

- 보이스: codex-cli 0.165.0 `codex exec -s read-only -C <worktree> </dev/null`(다른 계열, 적대적 Eng). 프롬프트 `analysis/review/proposal-freeze-r1-prompt.md`, 최종 답 `analysis/review/proposal-freeze-r1-codex.md`(전체 대화 기록은 커밋하지 않고 scratch 에 둠). 대상 HEAD `4c9d53138`, 정적 검토(실행 0).
- 판정: **「freeze 불가 — 설계 결정에 닿는 P1 2」.** P0 0.

| # | 등급 | 요지 | 닿는 결정 | 내 확인 |
|---|---|---|---|---|
| R1-1 | P1 | 조건 (다)는 **다시 읽은 로컬**을 **옛 계좌 스냅숏**과 비교함. 스냅숏 **뒤에** 실제로 체결된 엔진 매도가 로컬 읽기 뒤 · 조정 tx 전에 커밋되면 (가)(나)(다)가 모두 참인데 진짜 불일치가 남음. 예: 앞 주기 조정으로 로컬 · 계좌 20 + 크레딧 · QM 활성 → 사람이 5 매도(스냅숏 15, 로컬 읽기 20) → 스냅숏 뒤 엔진 매도 5 체결 · 커밋(실제 계좌 10, 로컬 15) → 재판정 15 = 15 로 뺌 → 깨끗해진 비교가 **앞 크레딧으로 그 종목 QM 을 `ADJUSTMENT_APPLIED` 해제**. 워터마크는 「장부 커밋이 늦었다」의 증거이지 「그 체결이 스냅숏에 이미 반영됐다」의 증거가 아님 | D3 (다) · D4 · D6 · 결정 1 | **타당하다고 봄.** 코드 경로 확인(Observe B2 if@650 → B8 if@675 → 해제, B29 if@841 반박 안 됨). 결과는 해제된 그 주기 동안 실제 5주 차이가 있는데 진입이 열림 — 다음 주기 비교(로컬 15 · 계좌 10)가 새 QM 을 세우므로 「한 주기의 잘못된 해제」. 지금 코드(a201 전)는 이 입력에서 차단을 유지함. 시험으로 재현하지는 않음 |
| R1-2 | P1 | D5 의 「다시 읽은 로컬에서 더는 추적 아님」은 정정 lineage 에서 틀림: `LocalStateFromJournal` 은 부모 · 자식을 같은 현재 신원 키로 넣고 뒤의 것이 덮어써 `OriginalOrderID` 가 자식이 됨 → 부모의 missing-order 를 지우는데 살아 있는 자식이 실제로 missing. 또 `fail_closed` 종결도 추적에서 빠짐 — 「추적 부재 = 깨끗한 종결」이 아님. 1단계 탐침은 `TrackedFillOrders` 를 직접 읽어 이 반례를 검증하지 않았음 | D5 | **타당하다고 봄.** compare.go:225-235(map 키 = 현재 신원, 값 덮어씀) 확인. 내가 구현하려던 판정(원 번호 또는 현재 번호 일치)도 부모 번호 · 부모의 현재 번호가 둘 다 사라지면 같은 구멍 |
| R1-3 | P2 | §2 「차단이 늦어지는 상한 한 주기」 불성립 — R1-1 을 매 주기 되풀이(스냅숏 뒤 부분 체결이 주기마다 커밋)하면 실제 1주 차이가 여러 주기 차단 없이 지속. 60초도 주기 전체 시간의 상한이 아님(Run 은 작업 뒤 period 대기) | §2 위험 수용 근거 | 타당. R1-1 과 같은 원인 |
| 기타 | — | 조건 (나): 부분 커밋 · Applied=false · 알림 실패에서 `Converged` 유지, ApplyPositionAdjustment 는 Commit 뒤 오류를 내는 후속 연산 없음(position_adjustments.go:366). 오류 처리: Refresh 실패는 재판정 전에 반환, `cycle.Err` 는 R1-1 의 안전장치가 아님. a192: 결속된 같은 종목은 뒤 체결 워터마크 검사로 해제 거절(order_terminal_release.go:258) — 크레딧 경로는 이 검사를 안 거침. D2 의 id = 커밋 순서 반례 없음 | — | — |

- 처분: **반영하지 않고 멈춤**(Manager 지시 — 결정급 P1). freeze 안 함, tasks 1.12 · 1.13 미체크, 구현(2.x) 미착수, origin/main merge 미실행.

## Manager 결정 v2(2026-11-24) — 「체결이 지나간 주기는 판정하지 않고 다음 주기에 다시 본다」

- v1(빼기 · 다시 읽은 로컬 대 옛 스냅숏 · D5 · 같은 주기 해제)을 폐기. 사유는 R1-1 · R1-2(design 「기각한 대안」).
- 문서 v2 개정(커밋 `651516ebe`): proposal · design · ADDED(본문 344자, strict valid, MODIFIED 없음 — 「체결이 지나간 비교는 확인된 불일치가 아니다」) · Story · tasks 1.14.
- 전제(`analysis/claim-check-v2.md`):
  - (p1) **거짓(좁은 반례)** — 체결 행 없이 투영만 나중에 쓰는 전략 진입 backfill(strategy_dispatch_runtime.go:1348)이 수량 stale 을 내면 보류 조건이 안 걸림(지금 동작 그대로). `position.Apply` 의 delta 0 새 인스턴스 갈래는 미확인.
  - (p2) **참.**
  - (p3) **실패도 성공도 아닌 기존 표현 없음** — 소비자는 `note`(건강 계수 → `Runtime.CheckHealth`)와 `logCycle` 둘, 둘 다 `cycle.Err` 유무로만 가름.
  - (p4) **거짓(D7' 그대로라면)** — 「깨끗함 로그에 안 들어감」에 `logCycle` 편집 필요(2 함수).
- 「판 중단 뒤 B」 v2 탐침: B 의 체결이 보류 주기 뒤 · 다음 주기 전 → 행 0. 다음 주기 뒤 → 다음 주기 수렴이 B 를 조정해 기존 경로. 남는 모양 = 연속 둘째 주기(재현).
- 처분: 멈춤 조건((p1) · (p4) 거짓)에 걸려 **codex 2회차(마지막)는 돌리지 않음** — D7' 판본이 정해진 뒤에 돌려야 마지막 회차가 최종 설계를 본다. freeze · 구현 미착수.

## Manager 결정(2026-11-24) — D7' (β), (p1) 반례는 잔여, R1-1 의 뒤 차단은 base 대조

- D7' (β) 반영(커밋 `c0ca86df5`): 보류 주기의 계수 · clean/mismatch 줄은 오늘과 같게, `stale_local_read` 한 줄은 RunOnce 가 직접. 편집 기존 함수 `RunOnce` 하나.
- 「`Observe` 없이 성공으로 센 주기가 푸는 것」(design §4): `Consecutive = 0` · `CheckHealth` 의 `clearLatch`(알림 재무장) — 운영 모드는 안 풂. 오늘 불안정 주기와 같은 종류, 새로 푸는 종류 0.
- R1-1 「그다음 주기 차단 유지」: base 에서도 같음 — 픽스처(열린 주문 목록 중복 행 → missing-order 3연속 → 계좌 영구 차단). v2 만의 고착 아님.

## proposal-freeze r2 — codex(2026-11-24, 마지막 회차)

- 보이스: codex-cli 0.165.0 `codex exec -s read-only -C <worktree> </dev/null`. 프롬프트 `analysis/review/proposal-freeze-r2-prompt.md`, 최종 답 `analysis/review/proposal-freeze-r2-codex.md`. 대상 HEAD `c0ca86df5`, 정적 검토(실행 0).
- 판정: **「freeze 불가 — 설계 결정에 닿는 P1 2」.** P0 0.

| # | 등급 | 요지 | 닿는 결정 | 내 확인 |
|---|---|---|---|---|
| R2-1 | P1 | **주기 전체를 버리면 다른 종목의 a192 결속 · 연속성 증거가 사라진다.** 종목 B 에 이 프로세스가 O1 의 missing-order 로 만든 차단(O1 종결 · 무장됨)이 있고, 같은 B 의 비종결 주문 O2 가 이번 스냅숏에서 누락돼 `MissingOrders` 에 들어옴 → 무관한 A 의 W2 로 주기 보류 → 다음 주기에 O2 가 다시 정상 매칭 → 기존이라면 이번 관측이 O2 를 결속에 더해(`beginOrderTerminalObservation` call@647 → `updateOrderTerminalBindings` order_terminal_release.go:326) O2 비종결로 해제를 막았을 것을, v2 는 O1 만 보고 `ORDER_TERMINAL_CONFIRMED` 해제 가능. 같은 원인으로 승격: B 의 같은 missing-order 2회 관측 → 보류 주기(그 사이 B 정상 매칭) → 다시 missing → 기존이면 중간 비교가 streak 를 지웠을 것을 v2 는 2 가 남아 **비연속 3회로 영구 승격**(advanceStreaks mismatch.go:1089, 정본 :53) | D3' · D4' · D6' · (p2) | **타당하다고 봄.** 「그 비교는 없던 것」이 다른 종목의 **증거(결속 추가 · streak 단절)**까지 없앰. 내 (p2) 판정은 「다음 주기에 이어진다」만 보고 「이번 관측이 남겼어야 할 증거가 빠진다」를 못 봄 — (p2) 는 이 점에서 거짓으로 정정해야 함 |
| R2-2 | P1 | **D7' (β) 의 성공 처리가 지속 장애의 운영 모드 강화를 계속 막을 수 있다.** 편입 후보 B 의 가격 조회가 매번 실패(adoption.go:222 → `cycle.Err`)하는데 홀수 주기마다 다른 종목 A 에 W2 → 보류 주기는 편입을 건너뛰어 성공 → 계수 `0→1→0→1…`, 임계 5 미도달(runtime.go:420 · :462). 연속 보류 상한 1 이어도 **은폐 기간에는 상한이 없다** | D7' (β) | **타당하다고 봄.** 내 §4 열거는 「새로 푸는 종류 0」만 봤고 「교대 보류가 오늘이라면 연속이었을 실패를 끊는다」의 지속을 「불안정 주기도 같다」로 넘김 — 불안정 주기는 안정화 실패 입력이고 이 반례는 안정화 성공 입력이라 다름 |
| R2-3 | P2 | Refresh 가 보류 판정 **전**에 돌아 대기 중 영구 승격 철회 · streak 초기화 · gate 갱신(mismatch.go:1044 · :1070)을 이미 함 — 「보류 주기의 차단 · 연속 횟수 무변」은 「보류 뒤 Observe 에 의한 변경 없음」으로 한정해야 | D1' · D4' · ADDED | 타당(기존 Refresh 의 정당한 동작, 계약 문구 문제) |
| R2-4 | P2 | 연속 상한 1 은 재기동(메모리 불리언 초기화)과 중간 조기 반환 주기를 포함한 판정 지연 상한이 아님. 「직전 주기」의 뜻(문자 그대로 vs 판정 기회)이 탐침과 문구에서 다름 | D5' | 타당 — 구현 전에 뜻을 정해야 함 |
| R2-5 | P2 | ADDED 결함: (i) A 의 W2 와 B 의 독립 missing-order 가 같이 있으면 B 의 확정 불일치까지 차단 생략 — :53 즉시 ordinary block 에 명시 예외 필요, (ii) 첫 Scenario 가 직전 보류 · 판독 실패 예외를 빠뜨려 다른 Scenario 와 충돌, (iii) 마지막 credit Scenario 는 다른 종목 missing-order 로 비교가 막히면 해제 갈래에 안 감(Observe:650), (iv) 다른 종목의 실제 분쟁 관측을 버리면 credit 반박도 생략(정본 :312) | 정본 delta | 타당 |
| 기타 | — | R1-1 즉시 오해제는 닫힘. 과잉 보류는 보수만이 아님(다른 계좌 체결 · 무관한 체결로 정상 편입 후보의 자동 손절 부착이 늦음). delta 0 새 인스턴스 추측은 전이표로 지지 안 됨(사실 정정 반영). 탐침의 판독 실패 처리가 D8' 과 다름(A 만족 · B 판독 실패면 탐침은 보류) | — | 사실 정정 1 반영(claim-check-v2 · design §4) |

- 처분: **반영하지 않고 멈춤**(결정급 P1 2). freeze 안 함, tasks 1.12 · 1.13 미체크, 구현 미착수. codex 회차는 소진(3회차 없음).

## Manager 결정 — 보류(2026-11-24)

- R2-1 · R2-2 둘 다 옳다고 판정. v1 · v2 가 같은 뿌리(생략된 관측이 남겼어야 할 증거를 버림)에서 기각됐으므로 세 번째 설계를 급히 올리지 않고 **a201 을 보류**한다.
- 기록: `issues.md` I1, design 「기각한 대안」(v1 · v2 · 종목 단위 보류)과 §6 「보류와 재개 조건」, claim-check-v2 의 (p2) 거짓 정정, tasks 1.13 이후 [blocked], proposal 머리 보류 표지, Story limits.
- 재개 방향(후보): 해제 쪽 — 체결이 지나간 비교로 선 수량 불일치를 표지하고 그 뒤 새 스냅숏 비교가 깨끗할 때 `RECHECK_MATCHED` 로 푼다. 재개 조건: a194 장시간 차단 알림이 이 모양으로 오거나 사용자 지시. 재개 때 codex 회차를 새로 센다.
- 이 브랜치는 main 에 합치지 않고, worktree 와 브랜치는 남긴다.

## v3 재개 — 1단계(2026-11-24, 사용자 지시 「1번으로 진행해줘」)

- `origin/main`(`d66536d9b`, a200 착지) merge `6e36fd2c0`(PM 생성물 · 목록 충돌만 — 둘 다 남기고 재생성). base 유지(자기 production Go 커밋 0).
- v3 증거 조사(별도 팀메이트) 커밋 `ec44a8bb1`. v3 문서 · 전제 · 탐침 · delta · 번들 커밋 `85b3c8226`.
- 전제 (q1)~(q7)(`analysis/claim-check-v3.md`) — 설계를 바꿔야 하는 사실 0. (q1) 맵 공유로 같은 코드, V5 순서는 문자 그대로 아니나 결과 같음. (q7) a192 Scenario 문언 충돌 → MODIFIED.
- 편집 기존 함수 넷(RunOnce · Observe · persist · ValidReconcileReleaseCause), 새 상수 `FILL_APPLIED`. delta ADDED 1(410자) + MODIFIED 3, strict valid. `check_analysis` rc 1 — merge 로 창에 들어온 a200 `ProjectPosition`(`code-context/check-analysis-v3-pre.txt`).

## proposal-freeze v3 r1 — codex(2026-11-24)

- 보이스: codex-cli 0.165.0 `codex exec -s read-only -C <worktree> </dev/null`. 프롬프트 `analysis/review/proposal-freeze-v3-r1-prompt.md`, 최종 답 `analysis/review/proposal-freeze-v3-r1-codex.md`. 대상 HEAD `779359096`, 정적 검토.
- 판정: **「freeze 불가 — 설계 결정에 닿는 P1 1, P2 2」.** P0 0.

| # | 등급 | 요지 | 닿는 결정 | 내 확인 |
|---|---|---|---|---|
| V3R1-1 | P1 | **깨끗한 갈래의 권위 오류 반환이 credit 반박을 건너뛴다.** 대기 중 영구 승격(`pendingPermanent`)이 있는 상태에서 다음 비교가 그 종목을 **외부 포지션**으로만 봄(`symbolsInDispute` 는 외부 포지션을 분쟁으로 셈 · `BlocksEntry` 는 안 셈 → 깨끗한 갈래) → `resolvePendingPermanentAuthority` 오류면 return@660 — 깨끗한 갈래의 Awaiting 처분도 끝 루프(B29) 반박도 안 돎 → 체결 credit 이 살아남아, 외부 보유를 접은 뒤 다음 일치 비교에서 `FILL_APPLIED` 해제. V8 의 「외부 거래 상쇄」가 아니라 **명시적으로 관측된 분쟁이 credit 을 소멸시키지 못하는** 구멍 | V4(반박 분기 무편집) · 편집 표면 | **타당.** mismatch.go:650-660 확인(`finishAuthorityError` 는 credit 을 안 만짐, :1223-1234). 같은 구멍이 오늘 조정 credit 에도 잠재 — 「같은 코드를 공유하니 반박이 보장된다」는 V4 의 주장이 그 구멍까지 공유함. 막으려면 깨끗한 갈래에서도 권위 확인 **전에** 반박(차단 갈래 call@716 과 같은 모양) — Observe 편집 지점 하나 추가 |
| V3R1-2 | P2 | 원인 구분이 as-of 같음뿐이라, 체결 credit(as-of T) 뒤 `AdjustmentApplied(T, A)` 가 같은 T 로 덮어쓰면 실제 근거가 조정인데 `FILL_APPLIED` 로 표지 — 동일 as-of 재사용 · 시계 역행 경계(as-of 초 단위, credit API 는 단조 · 고유를 요구 안 함) | V4 · V5 | 타당(경계). 막으려면 `AdjustmentApplied` 가 `t.filled[종목]` 을 지우거나(다섯째 기존 함수 편집) 종류를 시각과 분리 |
| V3R1-3 | P2 | MODIFIED 의 Scenario 「조정 이후의 재조회가 여전히 불일치한다」 THEN 「해제는 새 조정과 그 뒤의 일치를 다시 요구한다」가 v3 의 「앞 크레딧 위 W2 → 체결 credit 해제」를 금지 — generate.py 단언은 치환 정확성만 봄 | 정본 delta | 타당 — 그 THEN 에 「또는 체결 credit」 치환 1 추가가 필요(사실 정정 수준이나 P1 과 함께 멈춤 규칙에 따라 미반영) |
| 기타 | — | R1-1 원형 · R1-2 · R2-1 · R2-2 는 v3 구조에서 재개방 안 됨(정적). 부여 주기의 Observe 는 무변이나 「이전 주기 체결 credit 까지 포함한 모든 Observe 무변」으로 일반화 불가. V5 의 a192 순서 논증은 정상 결속 수명에서 성립, `fail_closed` 결속 행을 체결 경로가 푸는 것은 설계가 명시적으로 넓힌 동작 — a192 종결 증명이 통과한 것으로 보고하면 안 됨 | — | — |

- 처분: **반영하지 않고 멈춤**(결정급 P1). tasks 1.10 미체크(r1 만 실행 — v3 회차 1 소진, 남은 1회), freeze 판단은 Manager.

## Manager 결정 — v3 r1 처분(2026-11-24)

1. V3R1-1 채택 → V4b: 깨끗한 갈래의 권위 오류 갈래 안(return@660 직전)에 `refuteUsableAdjustmentCredits(disputed, diff.AsOf)` 1 호출. 「이 change 가 기존 경로에서 바꾸는 유일한 동작」으로 design §2 · proposal · Story 에 명시. 오류 갈래 **안**에 둔 이유(오류 없는 경로의 갈래 · 결속 보류 사유 무변)는 claim-check-v3 「codex v3 r1 반영」.
2. V3R1-2 → V5b: `AdjustmentApplied` 무편집, `filled == adjusted` 등식 표지, as-of 글자 일치의 원인 이름 모호함은 design §4 공시. `adjusted` 쓰기 자리 전수에서 등식이 성립하지 않는 자리 0.
3. V3R1-3 → MODIFIED 치환 추가(「조정 이후의 재조회가 여전히 불일치한다」 THEN). 「조정만이 푼다」 전제 문장 자리 전수 — claim-check-v3 결정 3 표.
4. V5 사실대로 고쳐 씀 + 「가로채지 않음」 시험(design §3-13).
6. **gate ⑤ base 재고정 승인(Manager, 2026-11-24)**: 이 change 의 자기 Go 커밋은 재현 시험 파일뿐이고 고친 기존 함수 0 — a200 과 같은 경우. **지금은 하지 않음**: 구현 지시 때 「a199 착지 확인 → `git merge origin/main` → WORKFLOW 「사람 승인 base 재고정」 세 요건(귀속 실측 원문 · 이 승인 기록 · `base-commit.txt` 단독 커밋 + 영수증)」 순서. merge 뒤 a199 의 `reconcileloop.go` 편집(`Context.ReconcileDriver` · 패키지 주석)이 RunOnce 번들 재추출을 요구하는지 그때 확인.
- 반영 커밋 `3e42a46f0`.

## proposal-freeze v3 r2 — codex(2026-11-24, 마지막)

- 보이스: codex-cli 0.165.0 `codex exec -s read-only -C <worktree> </dev/null`. 프롬프트 `analysis/review/proposal-freeze-v3-r2-prompt.md`, 최종 답 `analysis/review/proposal-freeze-v3-r2-codex.md`. 대상 HEAD `d3e56fbd3`, 정적 검토.
- 판정: **「조건부 — 설계 수준 P0/P1 없음」.** P2 2 · P3 1.

| # | 등급 | 요지 | 처분(문서 정합 — 설계 무변) |
|---|---|---|---|
| V3R2-1 | P2 | V4b 시나리오 · MODIFIED :57 추가 문장에 「credit 보다 엄격히 나중」 조건이 없어, 같은 as-of 비교의 소멸까지 요구하는 것처럼 읽힘(실제 호출 `refuteUsableAdjustmentCredits` 는 `creditUsableBy` 로 보존) — 같은 비교의 소멸 금지 문장(:307)과 충돌 | **반영**: ADDED 시나리오 WHEN 과 MODIFIED :57 추가 문장에 「credit 보다 엄격히 나중」 |
| V3R2-2 | P2 | V5b 의 원인 이름 예외(같은 글자의 as-of)가 원인을 `ADJUSTMENT_APPLIED` 로 단정하는 THEN 둘(:85-87 「조정 반영 후 자동 해제」, :331-333 「조정 이후의 재조회가 일치한다」)에 반영 안 됨 | **반영**: 두 THEN 에 예외 괄호(치환 2) + ADDED Scenario 「체결 credit 뒤에 조정 credit 이 덮어쓴다」 |
| V3R2-3 | P3 | tasks 의 RED · GREEN · 변이 범위와 편집 표면이 r1 반영 전 상태 | **반영**: tasks 2.2(1~14) · 2.3(Observe 2 자리) · 2.5(§3-15), edit-surface-v3 의 Observe 행 |
| 기타 | — | V4b 의 `disputed` 집합 · 다른 조기 반환(:731 은 :716 반박을 거침) · 과잉 반박 반례 없음. V5b 는 공시한 동일 문자열 외 오표지 없음, `Restore` · `Resolve` 뒤 남은 `filled` 만으로는 해제 없음. a192 · R1-1 · R1-2 · R2-1 · R2-2 재개방 없음. 기존 탐침에는 V4b 가 없어 그 PASS 를 수정 후 검증으로 인정하지 않음 | 구현 RED 에서 V4b 를 실행으로 검증 |

- delta 재생성: ADDED 본문 410자 · Scenario 10, MODIFIED 3 · 치환 9(1회 적중 · 역치환 · Scenario 이름 보존), `openspec validate --strict` valid.

## 판정 — **freeze 가능 — Manager 확인 대기**(2026-11-24)

- 전제 (q1)~(q7) 참(설계를 바꿔야 하는 사실 0), codex v3 r1 의 P1 은 Manager 결정 1~4 로 반영, r2 는 P0 · 결정급 P1 0(P2 · P3 은 문서 정합으로 반영). 편집할 기존 함수 넷(RunOnce · Observe · persist · ValidReconcileReleaseCause), 스키마 36 유지.
- freeze 는 Manager 가 최종 문서를 읽고 확정한다(tasks 1.11). codex v3 회차는 2 소진.

## Manager freeze(2026-11-24 10:3x KST)

- **읽은 문서**: design.md(102줄 전부) · specs/reconciliation/spec.md ADDED(Scenario 10) + MODIFIED :51 · :55 · :75 문장 · tasks.md · analysis/claim-check-v3.md (q1)~(q5) · analysis/review/proposal-freeze-v3-r2-codex.md 전문(HEAD `bb153a1ba`).
- **코드 대조**(claim-check 의 AST 좌표를 현재 소스 줄과 맞춰 봄): mismatch.go `AdjustmentApplied`:484-497(맵에 as-of 덮어씀) · `creditUsableBy`:509-519(엄격히 나중만) · `refuteUsableAdjustmentCredits`:525-531(분쟁 ∧ 쓸 수 있는 credit 만 삭제) · `Observe` 깨끗한 갈래 :650-702(if@659/return@660, if@675, call@682, if@689, Cleared@701) · blocking 갈래 call@716 · `Restore`:1001 · `Resolve`:1392 의 `adjusted = nil`. reconcileloop.go `RunOnce`:425-479(로컬 읽기 :434, 수렴 :451, Refresh :458, Observe :464, 오류 반환 :470-475, judgeHoldings :477). `ConvergeReport.Converged []ConvergedPosition`(converge.go:105), `FillWatermark`(position_adjustments.go:190), `MaxFillEventID`(reconcile_order_terminal_release.go:75) 존재. 설계의 좌표와 어긋난 곳 0.
- **판정**: 차단 쪽 무변(그 주기 Observe 는 credit 을 모르고 돈다) · 해제는 엄격히 나중인 새 스냅숏 · 새 로컬 읽기의 깨끗한 비교에서만 · 영구 / 계좌 범위 블록은 안 풂(B7) · 재기동 / 운영자 해제 뒤 복원 없음 — v1 · v2 를 기각시킨 「증거를 버림」이 구조적으로 없음. codex r2 P0/P1 0. **freeze 확정.**
- **사용자에게 공시할 것**(종결 보고에 그대로): V8 의 세 경우(외부 거래 상쇄 · 계좌 조건 없는 종목 워터마크 · 수렴이 거절한 종목) — 정본 :77 원칙이 좁아지는 자리. V4b 는 오늘의 조정 credit 경로도 바꾸는 유일한 기존 동작 변경(반박이 늘어나는 보수 방향). §4 의 남는 것(판 중단 뒤 종목 B · 재기동 · 원인 이름 모호함).
- **구현 순서(확정)**: a199 착지 확인 → `git merge origin/main`(rebase 금지) → base 재고정(세 요건: 귀속 실측 원문 · 승인 기록 · `base-commit.txt` 단독 커밋 + 영수증 — Manager 승인은 자기 production Go 가 기존 함수를 0 개 고친 상태에서만 유효, merge 뒤 다시 세어 원문으로 남김) → `reconcileloop.go` 가 바뀌었으면 RunOnce pre-edit 번들 재추출 → Pre-Edit 선언 → 2.2. 구현 착수는 Manager 가 a199 착지 뒤 따로 지시. 그 전에는 go test · go build · make 류를 돌리지 않음(a199 종결 게이트의 부하 민감 시험).
