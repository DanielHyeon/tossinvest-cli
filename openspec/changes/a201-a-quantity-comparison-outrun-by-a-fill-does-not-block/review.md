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
