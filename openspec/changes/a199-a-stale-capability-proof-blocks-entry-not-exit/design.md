# a199 design — 낡은 capability 증명은 진입만 막고 청산은 막지 않는다

용어: **낡음(stale)** = attestation 이 만료됐거나 필요 endpoint 를 덮지 않음. **청산 전용 기동(exit-only start)** = 낡음만을 이유로 인터록이 거절하지 않고, 노출 증가를 막은 채 런타임을 세우는 것.

## 1. 결정

| # | 결정 | 근거 · 기각한 대안 |
|---|---|---|
| D1 | **청산 전용으로 내릴 실패는 닫힌 집합**(Manager 결정 3, 2026-11-24): 파일이 읽히고(`attest.Load` 성공), 형식 번호가 이 빌드가 아는 유효한 값이고(1 이상 · 이 빌드 이하 — `Load` 는 상한만 거르므로 하한은 공용 판정이 거름), 계좌 · 만료 · endpoint 목록이 비지 않고 **목록의 항목이 전부 비지 않은 문자열**이며(공백만인 항목도 빈 것), 해석된 계좌와 `Verify` 의 비교로 같고, `Verify` 의 실패가 `ErrExpired` · `ErrEndpointNotAttested` 뿐. 그 밖(`ErrMissing` · `ErrMalformed` · `ErrFormatTooNew` · `ErrIncomplete` · 빈 endpoint 항목 · 형식 < 1 · `ErrAccountMismatch`)은 기동 거절. 판정은 **공용 판정 함수 하나**(새 파일, 인터록 · 감시가 함께 부름)가 함 | 내리는 근거는 「이 계좌에서 한때 증명됐고 그 증명이 낡았다」다. 파일이 없거나 못 읽거나 다른 계좌면 그 근거가 없다 — 누구의 계좌를 파는지 증명할 것이 없는 엔진은 세우지 않는다. endpoint 누락은 종류를 가리지 않는다(POST 뿐 아니라 GET 도): 좁히면 `RequiredEndpoints` 를 넓히는 빌드가 옛 attestation 으로 재시작할 때 손절이 통째로 사라진다(soak 은 연속 3일을 다시 벌어야 한다) |
| D1a | **계좌 일치는 만료 뒤에도 따로 확인한다.** `Verify` 는 첫 실패만 돌려주고 순서가 불완전 → 만료 → 계좌 → endpoint 라서(AST Verify B2~B7), 만료된 파일은 계좌가 달라도 `ErrExpired` 로만 보인다. 계좌 판정은 `Verify` 가 쓰는 **같은 비교**를 불러야 하고 사본을 두지 않는다 | 사본은 두 정본이 된다. 1단계 확인(Manager 결정 12): 공개 진입점 없음 — `SameAccountMasked` 는 가림 꼴도 같다고 보는 다른 규칙이라 쓰지 않는다. `internal/attest` 의 새 파일 새 메서드(`NamesAccount` 류, 이름은 2단계)가 비공개 `sameAccount`(숫자열 동일)를 부른다. `Verify` 본문 무편집 |
| D2 | **다른 조항은 전부 계속 거절한다.** `verifyGate` 는 낡음을 만나도 뒤의 조항(Guardian 한도 단일 출처 · 7b · 7c · 게이트웨이 배선 · 키 운반 transport)을 끝까지 검사하고, 그중 하나라도 실패하면 지금과 같은 거절이다 — **그 조항의 오류만**(낡음을 오류 문구에 덧붙이지 않는다, Manager 결정 10). 지금 `verifyGate` 는 첫 실패에서 반환하므로(AST 11 갈래) B6 갈래를 「낡음이면 표지만 남기고 계속」으로 바꾸는 편집이다 | 선례 D4(각 조항이 여전히 거절함을 개별 시험). 청산 전용이어도 게이트웨이·transport 는 매도가 쓰는 것이고, 한도 조항은 진입 전용이지만 「낡음 + 다른 결함」을 조용히 통과시키면 낡음이 고쳐지는 날 그 결함이 처음 드러난다 |
| D3 | **`Verified` 의 뜻은 바꾸지 않는다**(인터록 전 조항 충족, 청산 전용 기동에서는 false). `AutomationStatus` 에 새 표지를 더하고(이름은 1단계 — 가칭 `ExitOnly` 와 사유 코드), **청산 쪽 소비자만** 그것을 읽게 바꾼다: `runEngineRun` 의 게이트 OFF 판정(B7@232), `NewContext` 의 Guardian 유지(B16@592), `ExitObserver` 생성(exitwiring.go:323), `ReconcileDriver` 생성(reconcileloop.go:368) | fail-closed 방향. `Verified` 를 읽는 자리를 놓쳐도 그 자리는 엄격한 쪽에 남는다(전략 진입 worker: strategy_entry_supervisor.go:469). 반대로 `Verified=true` 로 올리고 「진입 금지」 표지를 따로 두면, 놓친 소비자가 진입을 연다. **1단계 전수(go/types): 읽는 자리 8 — 청산 쪽은 이 넷이 전부, 진입 1(worker, 엄격 유지), 보고 3**(`analysis/verified-census/`). 보고 자리 `logGateDecision`(interlock.go:479)의 「보호는 프로세스 수명」 문장은 청산 전용 기동에서도 나가야 한다(정본 「보호 미배선 기동」 SHALL — Manager 결정 5, 편집 필수) |
| D4 | **진입 차단의 집행은 기존 한 곳.** `EntryGate` 에 새 사유 하나(가칭 `capability_attestation_stale`)의 래치를 기동 때 세운다. `Gateway.submit` 본문은 편집하지 않는다(진입 검사 `checkEntry` 가 이미 `raisesExposure` 일 때만 래치를 읽는다). 래치는 메모리이고 **매 기동마다 attestation 에서 다시 유도**된다. 래치를 세우는 자리는 `runInterlock`(판정 · 감사 · 래치를 한 함수에 — `gateFacts` 에 진입 게이트 칸 하나, `NewContext` 의 리터럴 한 줄 → a166 census 키 하나 갱신, Manager 결정 11). 새 사유는 `latchOrder` 에서 `ReasonStrategyCentralIntegrity` 뒤 · `ReasonOperatingModeBlocked` 앞 | 정본 「집행 지점은 하나여야 한다」. 원장 운영 모드를 쓰면 attestation 을 고친 뒤에도 사람이 `mode-release` 를 해야 풀린다 — 원인이 사라졌는데 남는 차단이고, 그 상태의 진실은 원장이 아니라 파일에 있다. 「재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다」는 유도 방식으로 만족된다(재시작해도 파일이 낡았으면 다시 선다) |
| D5 | **`EntryPermitted` 는 청산 전용 기동에서 false.** | 보고·스냅숏 소비자와의 일치 |
| D6 | **실행 중 전환 없음**(Manager 결정 4로 한정): 감시와 attestation 재읽기는 기동 표지(`Verified` · 새 표지) · 루프 집합 · 진입 래치를 실행 중에 바꾸지 않는다. 판정은 기동 시점뿐이다 | 정본 「실행 중인 엔진의 동작은 바뀌지 않는다」(a135 D4). 주문마다의 진입 권한은 정본 「자동 진입은 모든 안전 권한의 교집합이다」가 정하고 이 change 는 그것을 건드리지 않는다 |
| D7 | **알림은 기존 감시가 사실대로 말한다.** (1) 낡음 단계(`expired` · `scope_lost`)의 문장을 「지금 재시작하면 기동이 거부된다」에서 **「새 진입이 막힌 청산 전용으로 뜬다」**로. (2) 기한이 낡음의 기한(만료 · 감독 증거)인 `notice` · `imminent` · `proof_overdue` 의 「그 뒤 재시작이 거부된다」 · 「재시작이 거부되면 자동 손절이 돌지 않는다」류 문장도 같은 귀결로(Manager 결정 6). (3) **「지금 청산 전용으로 돌고 있다」 알림의 조건은 파일이 아니라 엔진의 기동 표지**다(Manager 결정 7): 엔진이 청산 전용으로 뜬 프로세스인 동안 하루 한 번 — 파일이 아직 낡았으면 critical 「청산 전용으로 돌고 있다 + 할 일」, 파일이 지금 판정을 통과하면 일반 등급 「증명은 회복됐다 — 엔진을 재시작하면 정상으로 뜬다」. 새 단계 · 새 알림 타입 없음, 키 규칙(하루 한 번 · 에피소드)은 a196 그대로. 기동 표지의 출처는 `newAttestationWatch` 가 이미 읽는 `c.Automation`(기동 시점 값, D6 로 불변) — 감시가 **읽기만** 하므로 a135 D4(엔진 상태에 쓰기 0)와 어긋나지 않는다. (4) 거절이 유지되는 단계(`account_mismatch` · `unreadable` · `incomplete`)의 문장은 무변. (5) `judgeAttestation` 은 공용 판정 함수를 불러 D1 · D1a 와 같은 답을 낸다(지금은 「만료 + 다른 계좌」를 `expired` 로 판정 — claim-check D7-a) | 감시는 인터록과 같은 술어를 쓴다는 것이 a135 의 계약이므로, 인터록의 귀결이 바뀌면 감시의 문장이 따라야 한다. 새 알림 종류·새 outbox 타입을 만들지 않는다. **「어떤 단계가 청산 전용으로 내려가는가」의 판정은 인터록과 감시가 같은 함수 하나를 부른다**(판정 둘 금지) |
| D8 | **기동 시점의 기록**: 감사 새 action(가칭 `ActionGateExitOnly`, Detail 은 사유 코드와 빠진 endpoint 이름 — 계좌·경로·원문 오류 없음), 구조 로그 `engine.operating_mode` Warn + 사유, `engine run` 출력 한 줄(운영자가 읽는 문장, 확인 입력 없음) | 거절과 수락 사이의 세 번째 결과를 감사에 남기지 않으면 「언제부터 청산 전용이었나」를 재구성할 수 없다 |
| D9 | 인터록 머리 주석을 고친다(8항 Protection 서술 · "no degraded mode"). `cmd/tossctl/soak.go` 의 낡은 출력 문구(565 「The automation gate will refuse to start」, 577 「clause 9 … The gate will still refuse to start」 — 후자는 interlock-gates-entry-not-exit 이후 이미 거짓, `newSoakAttestCmd` 도움말 220)를 새 귀결로(Manager 결정 11) | 문서·코드 일치. 머리 주석은 함수 본문 밖, soak.go 문구는 `runSoakAttest` · `newSoakAttestCmd` 본문 편집 |

## 2. 무엇이 통과하고 무엇이 거절되는가 (fail-closed 경계)

**이 change 가 새로 통과시키는 것** — 낡은 attestation 으로 뜬 엔진의:
- 위험 감소 주문(손절·익절·청산 매도, 자기 발의 취소) — 증명 밖의 POST 호출이다. 실패하면 기존의 실패 처리(IN_DOUBT · 지연 경보)가 그대로 받는다.
- 대사 · 편입 · 체결 감지의 읽기 — 만료된 attestation 에서는 GET 도 증명 밖이다.

**여전히 거절되는 것**:
- 청산 전용 기동에서의 모든 노출 증가 mutation(매수, 수량·가격을 올리는 정정).
- 계좌 불일치 · 파일 없음 · 무효 · 불완전 · 새 형식 → 기동 거절.
- 다른 인터록 조항의 실패 → 기동 거절(낡음과 함께 있어도).
- 게이트 OFF → 런타임 없음(무변).

**거부해서는 안 되는 정상 입력**(시험으로 고정):
- 유효한 attestation → 지금과 똑같이 `Verified`, 래치 없음, 문장 무변.
- 만료 + endpoint 누락이 **함께** 있는 파일 → 청산 전용(둘 다 낡음).
- 계좌 번호 표기만 다른 같은 계좌(하이픈 — `Verify` 의 비교가 같다고 보는 것) → 청산 전용. *(1단계 사실 정정: 초안의 「가림」은 `Verify` 의 비교(숫자열 동일, attest.go:296)가 다르다고 본다 — `claim-check.md` §2-a)*

**통과시켜서는 안 되는 것**(시험으로 고정):
- 만료됐고 **다른 계좌**를 가리키는 파일 → 거절(D1a — `Verify` 순서 때문에 `ErrExpired` 로만 보이는 경우).
- endpoint 목록이 **빈** 파일 → `ErrIncomplete`, 거절(「전부 누락」이 「일부 누락」으로 읽히지 않는다).
- endpoint 목록에 **빈 문자열 · 공백만인 항목**이 있는 파일(`[""]` · `["   "]` · 정상 항목과 섞인 것) → 만료 · 미만료와 교차해 거절(Manager 결정 3 — 지금 `Verify` 는 길이만 봐서 `ErrEndpointNotAttested` 로 보임).
- 형식 번호가 **1 미만**(0 · 음수)인 낡은 파일 → 거절(`Load` 는 상한만 거름). 발급자는 이런 파일을 만들지 않는다(claim-check R1 근거) — 그래서 이 가드가 거부하는 정상 입력은 없다.

## 3. 시험 (RED 의 뼈대 — 1단계가 이름과 자리를 확정)

1. 낡음 두 종(만료 · endpoint 누락)과 그 합: `NewContext` 가 Context 를 돌려주고, `Verified=false`, 새 표지 true, `EntryPermitted=false`, Guardian 비-nil, 감사에 새 action 1줄(거절·수락 action 0).
2. 그 Context 로 `ExitObserver` · `ReconcileDriver` 가 만들어지고, **전략 진입 worker 는 만들어지지 않는다** — 청산 전용에서는 Guardian 이 비-nil 이므로 worker 조건(strategy_entry_supervisor.go:469 `… && c.Guardian != nil && c.Automation.Verified`)의 `Verified` 항이 **유일한** 차단이 된다(지금은 Guardian=nil 이 같이 막음). 그래서 이 시험이 그 항을 고정한다.
3. 그 게이트웨이에 매수 → 새 사유로 거절(브로커 호출 0), 매도·취소 → 통과 — **보호가 WIRED 인 조립에서**(지금 생산 조립 UNWIRED 에서는 `checkProtection` 이 `checkEntry` 보다 먼저라 `broker_protection_not_wired` 가 거절 사유다 — claim-check P7). 그리고 UNWIRED 조립에서는 매수 거절 사유가 지금과 같다(`broker_protection_not_wired`, 무변)는 사례 하나.
4. 거절 유지 종류(없음 · 파싱 불가 · 새 형식 · 형식 < 1 · 불완전 3 · 빈 endpoint 항목 · 계좌 불일치) + **만료이면서 다른 계좌** · **만료이면서 빈 endpoint 항목** → 기동 거절, 오류 sentinel 무변(새 거절 둘은 `ErrIncomplete` · `ErrFormatTooNew` 계열 또는 새 sentinel — 2단계에서 이름 확정).
5. 낡음 + 다른 조항 실패(한도 단일 출처 · 7b · 7c · 게이트웨이 · transport 각각) → **그 조항의 sentinel 만**으로 거절(낡음 sentinel 은 싣지 않음 — Manager 결정 10).
6. 유효한 attestation · 게이트 OFF → 편집 전과 같은 결과. **뒤집히는 기존 시험 함수는 이 여섯뿐**(그 밖의 기존 시험이 깨지면 멈춤 조건): `TestGateOnRefusals`(사례 셋 — 「attestation expired」「attestation misses an endpoint the engine calls」「attestation predates the price read」), `TestRefusedStartIsAudited`(만료 픽스처), `TestGateDecisionIsLoggedStructurally`(하위 「refused」 만료 픽스처), `TestEachInterlockClauseHasALine`(「an expired attestation」 사례), `TestA196RefusalTellsTheConsequenceFirst`(scope_lost · expired 사례), `TestA196ScopeLossWordingFollowsTheMissingKind`(네 사례의 접두 단언). *(1단계 사실 정정: 초안 「두 사례만」)*
7. `runEngineRun`: 청산 전용 Context 는 `errEngineGateOff` 가 아니라 런타임 조립으로 가고 출력 한 줄을 낸다.
8. 재기동: 낡은 채 다시 기동 → 래치가 다시 선다. attestation 을 유효한 것으로 바꾸고 기동 → **attestation 사유의 래치**가 없다(원장에서 복원되는 다른 차단 — 알림 미전달 · 운영 모드 · RECONCILE — 은 이 시험이 말하지 않는다, Manager 결정 8). `alert-ack` · `mode-release` 가 이 사유를 지우지 않는다는 사례 하나.
9. 감시 문장: `expired` · `scope_lost` 의 제목 · 첫 문장과 `notice` · `imminent` · `proof_overdue` 의 기한 뒤 귀결이 「청산 전용으로 뜬다」를, `account_mismatch` · `unreadable` · `incomplete` 는 옛 귀결을 말한다. 「내려가는 단계」 판정이 인터록과 같은 함수임을 구조 시험으로.
10. 감시의 「청산 전용으로 돌고 있다」: 기동 표지가 청산 전용인 감시에서 (a) 파일이 낡았으면 critical, 하루 한 번, (b) 파일이 판정을 통과하면 일반 등급 「증명은 회복됐다 — 재시작하면 정상으로 뜬다」, 하루 한 번, (c) 기동 표지가 정상인 감시에서는 이 알림 0. 감시가 Context 의 어떤 값도 쓰지 않음(a135 D4 회귀).
11. **기동 복구 → exit 관측 연결**(Manager 결정 9): 실제 런타임 조립(`engineRuntime` 과 같은 순서)으로 청산 전용 기동 → `Runtime.Run` 의 Recover 가 끝나고 → exit 관측 한 주기가 돌아 손절 매도가 실제 게이트웨이에 제출된다(fake 브로커가 매도 1건 관측). 「생성자가 성공한다」로 대신하지 않는다. Recover 실패 시 루프가 안 뜨는 것은 기존 성질(a199 무관)로 기록.
12. 변이(사본에서): 계좌 확인 삭제 · 닫힌 집합에 `ErrMissing` 추가 · 빈 endpoint 항목 가드 삭제 · 형식 하한 삭제 · 뒤 조항 검사 생략 · 래치 세우기 삭제 · `Verified` 를 true 로 올리기 · Guardian nil 유지 · 감사 action 을 수락으로 · 감시 판정을 따로 두기 · worker 조건에서 `Verified` 항 삭제 · 감시의 청산 전용 알림 조건을 파일 판정으로 바꾸기 · 보호 수명 문장 조건에서 청산 전용 삭제.

## 4. 잔여 위험 · 공시

- **증명 밖 호출로 매도한다.** 낡은 attestation 은 「그 경로가 지금도 된다」를 말하지 않는다. 브로커 계약이 그 사이 바뀌었으면 청산 주문이 실패하거나 결과 불명이 될 수 있다 — 그 귀결은 지금의 「청산 시도 없음」보다 나쁘지 않다(초과 매도는 Guardian 의 위험 감소 발급과 브로커의 보유 한도가 막는다 — 이 change 가 바꾸지 않는 장치).
- **실행 중에는 기동 표지가 바뀌지 않는다**(D6). 도는 중에 낡아진 엔진의 기동 표지 · 루프 · 래치는 그대로다. 주문마다의 진입 권한은 정본 「자동 진입은 모든 안전 권한의 교집합이다」가 정하며 이 change 는 그것을 바꾸지 않는다.
- **청산 전용 중 다른 차단이 생기면 재기동 뒤에도 남는다** — 예: 알림 전달 실패가 `critical_alert_undelivered` 래치와 `ENTRY_BLOCKED` 승격을 남기면 attestation 을 고쳐 재기동해도 그 차단은 원장에서 복원된다(의도된 동작, 사람 해제). a199 는 그것을 풀지 않는다.
- **청산 루프가 실제로 도는 것은 기동 복구 성공 뒤다**(`Runtime.Run` 이 Recover 를 먼저 — runtime.go:294). Recover 가 실패하면 청산 전용이든 정상이든 루프가 없다(기존 성질).
- **낡은 채 무기한 돈다.** 청산 전용 엔진은 사람이 `verify run` 을 하기 전까지 그 상태로 남는다. 감시가 하루 한 번 알린다. 기한(예: 낡은 지 N일이면 거절)은 두지 않는다 — 기한은 「그날부터 손절 없음」을 다시 만든다.
- **콘솔 문구**가 부정확해진다(비목표).
- 낡음 판정이 attestation 파일의 내용에 기댄다. 파일은 서명되지 않는다 — 파일을 쓸 수 있는 자는 지금도 유효한 파일을 지어낼 수 있으므로 새 노출은 아니다.

## 5. 순서 · 결합

- a166 census: `NewContext` 는 본문 동결 키다. 편집하면 a192 `analysis/census-repair/` 절차로 **그 키 하나**를 갱신한다. `runInterlock` · `refuseStartup` · `newAutomationStatus` · `gateFacts` 는 키가 아니지만 `NewContext` 본문에 이름이 있다 — 이름을 바꾸지 않는다(`verifyGate` 는 NewContext 본문에 없음 — 1단계 사실 정정).
- `TestProtectionReadinessIsAbsentFromExitReconcileAndFillLoopSources`: exitwiring.go · reconcileloop.go 에 금지 문자열을 넣지 않는다.
- 감독 루프 4개의 집합·순서 고정 시험: 루프 집합을 바꾸지 않는다.
- a063 · verify-execution-capability · a105 · a107(진행 중, 같은 영역): 편집 파일이 겹치는지 1단계가 확인한다. a107 은 봉인 시험 파일을 같이 편집하므로 그 파일들을 건드리지 않는다.
- a198 이 `Context.ExitObserver` 를 편집한다(edit-surface §3). 1단계 base(`e4165fa3d`)는 a198 착지 전이다 → **a198 착지 · merge 뒤 사람 승인 base 재고정**(Manager 결정 1 — 승인 기록은 review.md, 실행은 그 뒤) · `Context.ExitObserver` 번들 재추출 후 구현 착수.
- a198 위에서 구현한다. a200(귀속 경합)과 편집 파일이 겹치지 않는다(대사·체결 싱크 쪽).
