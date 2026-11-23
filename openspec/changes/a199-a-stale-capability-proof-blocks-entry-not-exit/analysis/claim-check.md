# 1.10 claim check — proposal · design 의 분기 · 순서 · 「~하므로 안전하다」 주장 대조

HEAD `e4165fa3d`(코드 = `9738a0c32`). 근거 AST 는 `cited-ast/`(줄·분기 id 는 그 파일의 열거), 타입 전수는 `verified-census/`. 판정: **참 / 거짓 / 부분 / 미확인**.
설계 결정(D1~D9)을 고치지 않는다 — 거짓 · 부분은 `review.md` 의 Manager 질문으로 올린다. 사실 정정(줄 번호 · 이름 · 자리 수)만 문서에 반영한다.

## A. proposal

| # | 주장 | 근거 | 판정 |
|---|---|---|---|
| P1 | 거절 사슬 `verifyGate` B6@536 → `runInterlock` B2@400 → `NewContext` B15@588 — Context 없음, 루프 0 | verifygate B6 if@536 · runinterlock B2 if@400 · newcontext B15 if@588(return nil, err @590) | 참 |
| P2 | 발급자는 기간을 벗어난 감독 증거를 오류 없이 건너뜀(`soak/attest.go:331-333`) | 소스 331 `age := now.Sub(p.At)` · 332 `if age < 0 \|\| age >= validity` · 333 `continue` | 참 |
| P3 | 11-23 기동이 `ErrEndpointNotAttested` 로 거절됨 | 운영 파일 · 감사 로그 접근 금지(`~/.config/tossctl`) | **미확인** |
| P4 | 거절은 감사 로그와 stderr 뿐, ntfy · outbox 0 | runInterlock 매개변수에 알림기 없음, AST calls = logGateDecision · verifyGate · refuseStartup · log.Record | 참 |
| P5 | 「no degraded mode」 문장은 코드 주석(`interlock.go:32-34`, 커밋 `30563ae53`) | 소스 32-34, `git log -S'There is no degraded mode'` → 30563ae53(2026-07-26) | 참 |
| P6 | 선례가 「attestation 요구는 전부 불변」으로 이 조항을 남김 | archive interlock-gates-entry-not-exit design.md:27 | 참 |
| P7 | 생산 빌드는 보호 조립이 두 시장 `Wired:false` 라 게이트 ON 엔진의 **모든 매수**가 이미 `broker_protection_not_wired` 로 거절됨 | protection_wiring.go:41-42 `Wired: false`; protectionreadiness/production.go:293 비-Wired 조립이면 provider 생성이 중단; `checkProtection` B1@90 `!raisesExposure` 통과 · B3@96 provider nil → 거절; `submit` 이 2a `checkProtection`(586) 을 2b `checkEntry`(592) 앞에, 디스패치 직전 재검사(662) | **부분 참** — 매수뿐 아니라 노출을 늘리는 정정(Amend)도 같은 거절. 다만 **순서 결과**: 청산 전용 엔진에서 매수 거절의 사유 코드는 새 래치 사유가 아니라 `broker_protection_not_wired`(보호 검사가 먼저). 운영자가 새 사유를 보는 것은 보호가 WIRED 가 된 뒤(a105)뿐 — 시험 §3-3 이 WIRED 조립을 쓰는 이유와 같음. 문장 정정은 불요, Story/review 에 기록 |
| P8 | 「청산 전용 기동에서 도는 것: 감독 루프 집합 그대로, 보조 실행자 그대로」 | engineRuntime AST(B1~B6 전부 생성자 오류 반환), 루프 4개 고정 · 보조 6개(749-750) | 참 — 단 ReconcileDriver · ExitObserver 의 생성자가 새 표지를 읽어야 함(P8 은 D3 편집을 전제) |
| P9 | 「전략 진입 worker 는 Verified 를 계속 요구하므로 만들어지지 않는다」 | strategy_entry_supervisor.go:469 `… && c.Guardian != nil && c.Automation.Verified` | 참 — 그리고 청산 전용에서 Guardian 이 비-nil 이 되므로 이 조건의 `Verified` 항이 **유일한** 차단 항이 됨(지금은 Guardian=nil 이 같이 막음). 시험 고정 권고 |

## B. design

| # | 주장 | 근거 | 판정 |
|---|---|---|---|
| D1 | 내릴 실패는 Load 성공 + Verify 의 `ErrExpired`/`ErrEndpointNotAttested` + 계좌 일치 — 닫힌 집합 | Load AST B2~B4(없음 · 읽기 실패 · 파싱 · 새 형식) · Verify B1 switch@212(불완전 3) · B5@222 만료 · B6@228 계좌 · B7@233 endpoint | 참(집합 정의로서 구현 가능) |
| D1a-1 | `Verify` 는 첫 실패만 돌려주고 순서가 불완전 → 만료 → 계좌 → endpoint | Verify AST B2~B7, return 줄 214·216·219·223·229·234 | 참 |
| D1a-2 | 「만료된 파일은 계좌가 달라도 `ErrExpired` 로만 보인다」 | 같음(B5@222 가 B6@228 앞) | 참 |
| D1a-3 | 「endpoint 누락(`ErrEndpointNotAttested`)이면 계좌는 이미 확인됨」(design 이 명시하진 않으나 D1 의 전제) | Verify B6@228 이 B7@233 앞 — B7 에 닿았다면 계좌 일치 | 참 — 따로 계좌 확인이 필요한 것은 만료뿐 |
| D1a-4 | `internal/attest` 에 그 비교의 공개 진입점이 있는지 | `Verify` 의 계좌 비교는 비공개 `sameAccount`(296 — 숫자열 동일). 공개 `SameAccountMasked`(148) 는 `sameAccount ‖ Mask(full)==other` 로 **더 관대한 다른 규칙**이고 `Verify` 는 그것을 부르지 않음(Verify AST calls) | **공개 진입점 없음** — design 의 「없으면 새 파일의 새 메서드 하나가 기존 비공개 비교를 부른다」 경로가 맞음. `SameAccountMasked` 를 쓰면 두 정본이 됨(금지) |
| §2-a | 「계좌 번호 표기만 다른 같은 계좌(**가림** · 하이픈 — Verify 의 비교가 같다고 보는 것) → 청산 전용」 | `sameAccount` 는 숫자열 동일 비교 — 하이픈은 같다고 봄, **가림(`****5678`)은 숫자열 `5678` ≠ 전체 → 다르다고 봄** | **부분 거짓** — 「가림」은 Verify 가 같다고 보지 않음. 괄호의 정의(「Verify 의 비교가 같다고 보는 것」)가 기준이면 예시 「가림」은 빼야 함. 시험을 「가림」으로 쓰면 그 시험은 `SameAccountMasked` 를 요구하게 됨 → **사실 정정 제안(Manager 확인)** |
| D2 | 「verifyGate 는 낡음을 만나도 뒤 조항(한도 단일 출처 · 7b · 7c · 게이트웨이 · transport)을 끝까지 검사」 | verifyGate AST: B6@536 이 B7~B11(542·546·550·555·558) **앞**이고, 11 갈래 전부 `return`(첫 실패 반환, 머리 주석 491). | 지금 구조로는 **불가능** — B6 갈래가 반환하지 않도록 함수 본문을 바꿔야 함(편집 전제는 design 이 이미 함). 요구되는 것: (1) B6 에서 닫힌 집합이면 표지만 남기고 계속, 그 밖이면 지금처럼 반환, (2) 만료일 때 계좌를 따로 확인, (3) B7~B11 중 하나가 실패하면 **그 조항의 오류**로 반환 — 낡음 표지는 버림. 결정 필요: 그 거절 오류에 낡음 sentinel 도 같이 싣는가(`UnmetInterlockClauses` 가 두 줄 열거) — Q. B1~B3(Guardian · 한도 · 정책)은 attestation 앞이라 이미 먼저 거절 |
| D3-a | `Verified` 의 청산 쪽 소비자는 넷(runEngineRun B7@232 · NewContext B16@592 · exitwiring.go:323 · reconcileloop.go:368) | 타입 전수: 읽는 자리 8 — 청산 쪽 정확히 넷 + 진입 1(469) + 보고 3(interlock.go:461·479, strategy_entry_supervisor.go:392) | 참(청산 쪽 넷은 전부) |
| D3-b | fail-closed: 놓친 자리는 엄격한 쪽에 남음 | 보고 자리 479(`logGateDecision` 「보호는 프로세스 수명」 문장)는 놓치면 **엄격이 아니라 침묵** — 청산 전용 기동이 보호 미배선 런타임인데 정본 「보호 미배선 기동」의 「기동 시점 전달 SHALL」 문장이 안 나감 | **부분** — 진입에 대해서는 참, 보고에 대해서는 놓친 자리가 정본 위반이 됨. Q |
| D3-c | `Verified` 를 안 읽고 진입을 여는 경로가 없음 | 진입 생산자 = `Gateway.Place`·`Amend`·`PlaceClaimedStrategy`(→ `place` → `submit`) 뿐, 셋 다 `checkEntry`(B1@860: `g.entry == nil ‖ !raisesExposure` 면 통과). `ReplayInDoubt` 는 checkEntry 를 안 지나지만 생산 배선 `Attested: nil`(engine gateway.go:338) → replay.go:260 에서 먼저 반환. 생산 Guardian 은 진입 게이트를 읽지 않음(`EntryLatchFor` 비시험 호출자 0) | 참 — **조건: 래치가 실제로 서고 `g.entry` 가 비-nil**. 후자는 D2 가 조항 8(`Wiring().Entry`)을 계속 요구하므로 성립. 구멍은 「래치 세우기 누락」 하나 → 변이로 고정 |
| D4-a | 집행은 기존 한 곳 — `checkEntry` 가 이미 `raisesExposure` 일 때만 래치를 읽음, `submit` 무편집 | gateway.checkentry AST B1@860, `CheckEntryFor` → `checkAccountEntryLocked`(래치 먼저 — latchOrder, 그다음 맵 순회 B3@638) | 참. 단 새 사유를 `latchOrder` 에 넣지 않으면 맵 순회로 **이름이 비결정**(여전히 차단은 함) — latchOrder 등록 필수 |
| D4-b | 래치는 메모리이고 매 기동마다 attestation 에서 다시 유도 | `NewEntryGate` 가 빈 `latches` 로 매 기동 생성(engine gateway.go:265), 기동 복원은 알림 미전달 · 운영 모드 · RECONCILE 셋뿐(287·291·279). 새 사유를 푸는 코드 0(Clear 호출자 grep: 알림 · 체결 SLO · recovery · reconcile 의 고정 사유뿐, `mismatch.go:1532` 는 reconcile 두 사유 루프) | 참 |
| D4-c | 「원장 운영 모드를 쓰면 mode-release 가 필요」 | journal/operating_mode.go 트리거 6종에 attestation 없음(사전 증거 Q3) | 참(기각 근거로 유효) |
| D5 | `EntryPermitted` 는 청산 전용에서 false | interlock.go:409 가 유일한 쓰기, 읽기 2(보고) | 참(구현 쉬움) |
| D6 | 실행 중 전환 없음, 감시는 Verified 등을 안 씀 | attestation_watch.go 머리 22-23, 타입 전수에 감시의 쓰기 0 | 참 |
| D7-a | 「어느 단계가 청산 전용인가」를 인터록과 감시가 같은 함수로 | 지금 `judgeAttestation` B4@108 이 `ErrExpired` 를 그대로 expired 로 — D1a 의 계좌 확인이 없음 | 편집 필요(judgeAttestation 이 공용 판정 함수를 부르도록) — 안 하면 「만료 + 다른 계좌」에 감시는 「청산 전용으로 뜬다」, 인터록은 거부 → **D7 이 금지한 두 판정**. 편집 표면에 P7 로 넣음 |
| D7-b | 낡음 단계(expired · scope_lost)의 문장만 바꾸면 된다 | eventFor AST: notice/imminent(B4@408 · B5@409/B6@413) 과 proof_overdue(B7@418) 의 본문이 「그 뒤 엔진 재시작이 거부된다」 · `afterRefusal`(398 「재시작이 거부되면 그때부터 자동 손절이 돌지 않는다」)을 말함 | **거짓(범위 부족)** — 기한이 지나 낡아져도 결과는 청산 전용이지 거부가 아님. 이 세 단계의 문장도 a199 뒤 거짓이 됨. 정본 B 의 「날수 알림」 문장(2218)은 귀결을 규정하지 않지만 B3-1 로 기한의 이름을 바꿨음. Q |
| D7-c | 「엔진이 실제로 청산 전용으로 돌고 있으면 그것을 말한다」 | 감시는 `judgeAttestation` 이 healthy 면 알리지 않음(prepare B8@345). 청산 전용으로 돌던 엔진의 파일이 **갱신되어 유효해지면** 판정은 healthy → 알림 0, 엔진은 재시작 전까지 진입 차단 | **구멍** — 「파일은 고쳤는데 엔진은 아직 청산 전용」 상태를 아무도 말하지 않음. Q(감시가 기동 표지 + healthy 면 「재시작하면 정상으로 뜬다」 한 번 알릴지) |
| D7-d | 출처: 「지금 청산 전용으로 돌고 있다」 | 기동 때의 `c.Automation`(newAttestationWatch 227-228 이 이미 `c.Automation` 을 읽음) — D6 로 프로세스 수명 동안 불변 | 참(출처 확정 — newAttestationWatch 편집 P9) |
| D8 | 감사 새 action, 구조 로그, 기동 출력 한 줄 | `audit.ActionGate…` 상수 소비자: interlock.go 두 자리뿐(grep) — 새 action 을 읽다 깨지는 소비자 0 | 참 |
| D9 | 머리 주석 8항 · 「no degraded mode」 낡음 | 소스 29-34 vs verifyGate 562-577(보호 조항 제거) | 참 |
| §3-6 | `TestGateOnRefusals` 의 두 사례만 뒤집힘 | 시험 표: 「attestation expired」(514) · 「misses an endpoint」(533) · **「attestation predates the price read」(547 — endpoint 누락)** | **거짓 — 셋.** 그리고 다른 시험 함수 4개도 뒤집힘(`edit-surface.md` §2). 사실 정정 |
| §4 | 「초과 매도는 Guardian 의 위험 감소 발급과 브로커의 보유 한도가 막는다」 | Guardian 발급 내부 미열람(1단계 범위 밖) | 미확인 |
| §5 | `verifyGate`·`runInterlock` 은 census 키가 아니지만 `NewContext` 본문에 이름이 있다 | shapes.json 키 90 중 engine.go 2(NewContext · NewOrderPath) · interlock.go 0; NewContext 동결 본문에 `runInterlock`·`refuseStartup`·`newAutomationStatus` 문자열 | 부분 — `verifyGate` 는 NewContext 본문에 없음(runInterlock 안에서만 호출). 이름 고정은 runInterlock · refuseStartup · newAutomationStatus |

## B2. Manager 결정 3 의 가드가 거부할 정상 입력 (2026-11-24)

- **없다 — 발급자가 만들지 않는다.** `soak.BuildAttestation`(internal/soak/attest.go:229-239)은 soak 성공 endpoint 를 하나씩 `strings.HasPrefix(ToUpper(TrimSpace(e)), "GET ")` 로 검사해 아니면 발급을 거절하고(빈 문자열 · 공백만인 항목은 여기서 거절), 감독 증거 endpoint 는 `LiveOnlyEndpoints()` 닫힌 집합에서만 옴(같은 파일 305 순회). 형식 번호는 `FormatVersion: attest.FormatVersion`(262, 값 1 — internal/attest/attest.go:53)이고 `attest.Save` 는 0 을 1 로 채움(186-187). 그래서 형식 < 1 이나 빈 endpoint 항목은 손으로 쓴 · 깨진 파일에서만 나온다.
- 이 가드는 **낡은 파일을 청산 전용으로 내릴지**에만 걸린다 — 유효한 파일(`Verify` 통과)의 판정은 바꾸지 않는다.

## C. 지시서 (a)~(f) 질문에 대한 답

- **(a)** D2 는 verifyGate 를 「첫 실패 반환」에서 「낡음은 표지 후 계속」으로 바꾸는 편집을 요구한다(B6 갈래). 낡음 + 뒤 조항 실패의 오류에 낡음 sentinel 을 같이 실을지는 결정 사항(Q).
- **(b)** 구멍 없음 — 단 래치 세우기가 유일한 집행이고 생산 Guardian 은 래치를 안 읽으므로, 래치 누락은 곧바로 진입 개방이다(보호가 WIRED 가 되는 날부터). 변이 필수.
- **(c)** 래치(게이트웨이)는 노출을 늘리는 모든 mutation 을 덮는다: 전략 worker(만들어지지 않음 + 게이트웨이), 수동 매수 갱신(읽기만 — 진입 생산자가 아님), Guardian 의 EXPOSURE_RAISING 발급(Guardian 은 래치를 안 읽으므로 **결정 행 · 예약은 기록될 수 있고** 게이트웨이가 거절 — 지금 그 발급을 부를 생산자는 worker 뿐이고 휴면). 다른 생산자: `ReplayInDoubt`(비활성).
- **(d)** 출처 = 기동 때 `c.Automation`(새 표지). D6 와 모순 없음. 단 D7-c 의 구멍.
- **(e)** Guardian 은 `gate.Enabled && guardian == nil` 이면 인터록 **전에** 만들어진다(NewContext B12@560, 실패 B14@568 은 기동 거절) — attestation 과 무관. 한도가 무효면 verifyGate B2 가 먼저 거절. 그래서 청산 전용 기동의 Guardian 은 낡음과 무관하게 지금과 같은 입력으로 조립됨. 바꿀 것은 B16@592 의 nil 화 조건뿐.
- **(f)** P7 — 부분 참(정정은 순서 결과의 기록).
