# a199 design — 낡은 capability 증명은 진입만 막고 청산은 막지 않는다

용어: **낡음(stale)** = attestation 이 만료됐거나 필요 endpoint 를 덮지 않음. **청산 전용 기동(exit-only start)** = 낡음만을 이유로 인터록이 거절하지 않고, 노출 증가를 막은 채 런타임을 세우는 것.

## 1. 결정

| # | 결정 | 근거 · 기각한 대안 |
|---|---|---|
| D1 | **청산 전용으로 내릴 실패는 닫힌 집합**(Manager 결정 3, 2026-11-24): 파일이 읽히고(`attest.Load` 성공), 형식 번호가 이 빌드가 아는 유효한 값이고(1 이상 · 이 빌드 이하 — `Load` 는 상한만 거르므로 하한은 공용 판정이 거름), 계좌 · 만료 · endpoint 목록이 비지 않고 **목록의 항목이 전부 비지 않은 문자열**이며(공백만인 항목도 빈 것), 해석된 계좌와 `Verify` 의 비교로 같고, `Verify` 의 실패가 `ErrExpired` · `ErrEndpointNotAttested` 뿐. 그 밖(`ErrMissing` · `ErrMalformed` · `ErrFormatTooNew` · `ErrIncomplete` · 빈 endpoint 항목 · 형식 < 1 · `ErrAccountMismatch`)은 기동 거절. 판정은 **공용 판정 함수 하나**(새 파일, 인터록 · 감시가 함께 부름)가 함. **판정 순서(Manager Q-B, 2026-11-24): `Verify` 성공 → 지금과 같은 수락(가드를 타지 않음) / `Verify` 실패 → 닫힌 집합 판정(형식 하한 · 빈 항목 · 계좌 가드 포함).** 형식 하한 · 빈 항목 가드는 「청산 전용으로 내릴지」에만 걸리고 유효한 attestation 의 경로는 한 글자도 바꾸지 않는다 | 내리는 근거는 「이 계좌에서 한때 증명됐고 그 증명이 낡았다」다. 파일이 없거나 못 읽거나 다른 계좌면 그 근거가 없다 — 누구의 계좌를 파는지 증명할 것이 없는 엔진은 세우지 않는다. endpoint 누락은 종류를 가리지 않는다(POST 뿐 아니라 GET 도): 좁히면 `RequiredEndpoints` 를 넓히는 빌드가 옛 attestation 으로 재시작할 때 손절이 통째로 사라진다(soak 은 연속 3일을 다시 벌어야 한다) |
| D1a | **계좌 일치는 만료 뒤에도 따로 확인한다.** `Verify` 는 첫 실패만 돌려주고 순서가 불완전 → 만료 → 계좌 → endpoint 라서(AST Verify B2~B7), 만료된 파일은 계좌가 달라도 `ErrExpired` 로만 보인다. 계좌 판정은 `Verify` 가 쓰는 **같은 비교**를 불러야 하고 사본을 두지 않는다 | 사본은 두 정본이 된다. 1단계 확인(Manager 결정 12): 공개 진입점 없음 — `SameAccountMasked` 는 가림 꼴도 같다고 보는 다른 규칙이라 쓰지 않는다. `internal/attest` 의 새 파일 새 메서드(`NamesAccount` 류, 이름은 2단계)가 비공개 `sameAccount`(숫자열 동일)를 부른다. `Verify` 본문 무편집 |
| D2 | **다른 조항은 전부 계속 거절한다.** `verifyGate` 는 낡음을 만나도 뒤의 조항(Guardian 한도 단일 출처 · 7b · 7c · 게이트웨이 배선 · 키 운반 transport)을 끝까지 검사하고, 그중 하나라도 실패하면 지금과 같은 거절이다 — **그 조항의 오류만**(낡음을 오류 문구에 덧붙이지 않는다, Manager 결정 10). 지금 `verifyGate` 는 첫 실패에서 반환하므로(AST 11 갈래) B6 갈래를 「낡음이면 표지만 남기고 계속」으로 바꾸는 편집이다 | 선례 D4(각 조항이 여전히 거절함을 개별 시험). 청산 전용이어도 게이트웨이·transport 는 매도가 쓰는 것이고, 한도 조항은 진입 전용이지만 「낡음 + 다른 결함」을 조용히 통과시키면 낡음이 고쳐지는 날 그 결함이 처음 드러난다 |
| D3 | **`Verified` 의 뜻은 바꾸지 않는다**(인터록 전 조항 충족, 청산 전용 기동에서는 false). `AutomationStatus` 에 새 표지를 더하고(이름은 1단계 — 가칭 `ExitOnly` 와 사유 코드), **청산 쪽 소비자만** 그것을 읽게 바꾼다: `runEngineRun` 의 게이트 OFF 판정(B7@232), `NewContext` 의 Guardian 유지(B16@592), `ExitObserver` 생성(exitwiring.go:323), `ReconcileDriver` 생성(reconcileloop.go:368) | fail-closed 방향. `Verified` 를 읽는 자리를 놓쳐도 그 자리는 엄격한 쪽에 남는다. **단 진입을 실제로 막는 것은 `Verified` 가 아니라 래치 하나다**(구현 리뷰 r1 처분 1): `Verified` 를 읽는 진입 쪽 자리(strategy_entry_supervisor.go:469 `productionStrategyWorker`)는 투영용 조립의 worker 조건이고, 실행되는 런타임 전략 worker 는 `Verified` 를 읽지 않으며 진입 관문 → 래치로 막힌다(§3-2). 그래서 D4 「집행 지점은 하나」가 문자 그대로 참이다. 반대로 `Verified=true` 로 올리고 「진입 금지」 표지를 따로 두면, 놓친 소비자가 진입을 연다. **1단계 전수(go/types): 읽는 자리 8 — 청산 쪽은 이 넷이 전부, 진입 1(worker, 엄격 유지), 보고 3**(`analysis/verified-census/`). 보고 자리 `logGateDecision`(interlock.go:479)의 「보호는 프로세스 수명」 문장은 청산 전용 기동에서도 나가야 한다(정본 「보호 미배선 기동」 SHALL — Manager 결정 5, 편집 필수) |
| D4 | **진입 차단의 집행은 기존 한 곳.** `EntryGate` 에 새 사유 하나(가칭 `capability_attestation_stale`)의 래치를 기동 때 세운다. `Gateway.submit` 본문은 편집하지 않는다(진입 검사 `checkEntry` 가 이미 `raisesExposure` 일 때만 래치를 읽는다). 래치는 메모리이고 **매 기동마다 attestation 에서 다시 유도**된다. 래치를 세우는 자리는 `runInterlock`(판정 · 감사 · 래치를 한 함수에 — `gateFacts` 에 진입 게이트 칸 하나, `NewContext` 의 리터럴 한 줄 → a166 census 키 하나 갱신, Manager 결정 11). 새 사유는 `latchOrder` 에서 `ReasonStrategyCentralIntegrity` 뒤 · `ReasonOperatingModeBlocked` 앞 | 정본 「집행 지점은 하나여야 한다」. 원장 운영 모드를 쓰면 attestation 을 고친 뒤에도 사람이 `mode-release` 를 해야 풀린다 — 원인이 사라졌는데 남는 차단이고, 그 상태의 진실은 원장이 아니라 파일에 있다. 「재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다」는 유도 방식으로 만족된다(재시작해도 파일이 낡았으면 다시 선다). **래치를 세울 진입 게이트가 없으면(`gateFacts.entry == nil`) 청산 전용으로 내려가지 않고 거절**한다(구현 리뷰 r1 처분 2 — 유지 확정): 래치가 진입 차단의 유일한 집행이므로 세울 대상이 없으면 내려가지 않는다. 가드가 없으면 nil 역참조 panic 으로 같은 기동 실패가 나므로, 결과는 같고 오류만 읽을 수 있게 함. 생산에서는 `wiring.entry` 가 늘 비-nil(도달 안 함). 시험 `TestA199NoEntryGateMeansNoExitOnlyStart` |
| D5 | **`EntryPermitted` 는 청산 전용 기동에서 false.** | 보고·스냅숏 소비자와의 일치 |
| D6 | **실행 중 전환 없음**(Manager 결정 4로 한정): 감시와 attestation 재읽기는 기동 표지(`Verified` · 새 표지) · 루프 집합 · 진입 래치를 실행 중에 바꾸지 않는다. 판정은 기동 시점뿐이다 | 정본 「실행 중인 엔진의 동작은 바뀌지 않는다」(a135 D4). 주문마다의 진입 권한은 정본 「자동 진입은 모든 안전 권한의 교집합이다」가 정하고 이 change 는 그것을 건드리지 않는다 |
| D7 | **알림은 기존 감시가 사실대로 말한다.** (1) 낡음 단계(`expired` · `scope_lost`)의 문장을 「지금 재시작하면 기동이 거부된다」에서 **「새 진입이 막힌 청산 전용으로 뜬다」**로. (2) 기한이 낡음의 기한(만료 · 감독 증거)인 `notice` · `imminent` · `proof_overdue` 의 「그 뒤 재시작이 거부된다」 · 「재시작이 거부되면 자동 손절이 돌지 않는다」류 문장도 같은 귀결로(Manager 결정 6). (3) **「지금 청산 전용으로 돌고 있다」의 조건은 파일이 아니라 엔진의 기동 표지**다(Manager 결정 7 · Q-A, 2026-11-24): **파일 단계의 경보가 먼저**다(a196 의 단계 · 등급 · 키 그대로). 엔진이 청산 전용으로 뜬 프로세스면 그 알림 본문에 「지금 청산 전용으로 돌고 있다」 한 문장을 덧붙인다 — 파일이 지금 판정을 통과하는 단계(notice · imminent · proof_overdue)면 그 문장은 「증명은 회복됐다 — 엔진을 재시작하면 attestation 사유의 진입 차단이 풀린다」(Q-C). 파일이 healthy 인데 엔진이 청산 전용이면 같은 회복 안내를 **일반 등급으로 하루 한 번 따로** 낸다(notice 날에는 날수 알림에 덧붙이고 따로 내지 않음 — 알림 하나에 사실 둘). 등급은 두 사실 중 높은 쪽(= 파일 단계의 등급, healthy 회복 안내는 일반), 알림 OFF 면 기존 규칙대로 일반. 새 단계 · 새 알림 타입 없음. 회복 안내의 하루 한 번은 기한이 아니라 날짜 · 에피소드로 센다 — 기한이 바뀌어 `sent` 가 비워져도 그날 다시 내지 않는다(codex r2 R2-3). 기동 표지의 출처는 `newAttestationWatch` 가 이미 읽는 `c.Automation`(기동 시점 값, D6 로 불변) — 감시가 **읽기만** 하므로 a135 D4(엔진 상태에 쓰기 0)와 어긋나지 않는다. (4) 거절이 유지되는 단계(`account_mismatch` · `unreadable` · `incomplete`)의 **귀결 문장은 무변, 엔진이 청산 전용이면 그 한 문장만 덧붙는다**(Q-A). (5) `judgeAttestation` 은 공용 판정 함수를 불러 D1 · D1a 와 같은 답을 낸다(지금은 「만료 + 다른 계좌」를 `expired` 로 판정 — claim-check D7-a) | 감시는 인터록과 같은 술어를 쓴다는 것이 a135 의 계약이므로, 인터록의 귀결이 바뀌면 감시의 문장이 따라야 한다. 새 알림 종류·새 outbox 타입을 만들지 않는다. **「어떤 단계가 청산 전용으로 내려가는가」의 판정은 인터록과 감시가 같은 함수 하나를 부른다**(판정 둘 금지) |
| D8 | **기동 시점의 기록**: 감사 새 action(가칭 `ActionGateExitOnly`, Detail 은 사유 코드와 빠진 endpoint 이름 — 계좌·경로·원문 오류 없음), 구조 로그 `engine.operating_mode` Warn + 사유, `engine run` 출력 한 줄(운영자가 읽는 문장, 확인 입력 없음) | 거절과 수락 사이의 세 번째 결과를 감사에 남기지 않으면 「언제부터 청산 전용이었나」를 재구성할 수 없다. **청산 전용 감사 기록이 실패하면 기동을 거절한다 — 수락 경로와 같은 정책**(구현 리뷰 r1 처분 3 확정): 기록 없는 청산 전용 기동은 「언제부터 진입이 막혀 있었나」를 재구성할 수 없게 하고, 감사 파일이 막힌 환경은 그 앞의 `recordGateSettings` 에서 이미 거절된다(이 change 가 바꾸지 않는 기존 동작) — 남는 것은 두 기록 사이의 일시 실패뿐이다. 그 경우 손절이 없다는 것은 Story limits 에 적음(엔진 밖 heartbeat 감시가 down 을 알림). 시험 `TestA199AnUnrecordableExitOnlyStartIsRefused`. 구조 로그의 `gate_exit_only` 칸은 청산 전용 기동 줄에만 싣는다(유효 · 게이트 OFF · 거절 줄은 base 와 같은 칸 집합 — 처분 5) |
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
- **낡은**(`Verify` 가 만료 · 누락으로 실패한) 파일인데 endpoint 목록에 **빈 문자열 · 공백만인 항목**이 있음(`[""]` · `["   "]` · 정상 항목과 섞인 것) → 청산 전용 아님, 거절(Manager 결정 3 — 지금 `Verify` 는 길이만 봐서 `ErrEndpointNotAttested` 로 보임).
- 형식 번호가 **1 미만**(0 · 음수)인 **낡은** 파일 → 거절(`Load` 는 상한만 거름). 발급자는 이런 파일을 만들지 않는다(claim-check B2).
- 범위 밖(기존 동작 — Manager Q-B 기각): `Verify` 를 통과하는 파일은 빈 항목 · 형식 0 이 있어도 지금처럼 정상 기동한다. 필요 endpoint 는 전부 덮였으므로 이 change 가 연 구멍이 아니고, 좁히는 것은 다른 일이다(Story limits).
- **(G2, Manager 결정 2026-11-24 — 대안 5) 원장에 이 계좌로 엔진이 낸 주문이 확정된 기록이 없는 낡은 파일** → 청산 전용 아님, 거절(`ErrOrderPathUnproven`, 낡음 sentinel 없음). D1 의 근거 「이 계좌에서 한때 동작했고 그 증명이 낡았다」는 파일에 남지 않는다 — 감독 검증을 한 번도 안 한 첫 `soak attest` 의 GET 만의 파일과 감독 증거가 기한을 넘겨 빠진 파일이 같은 모양이다. 그래서 낡음 갈래의 맨 끝에서 원장에 묻는다: **해석된 계좌**(TRIM 정규화 — 다른 계좌 범위 질의와 같음)의, **엔진 게이트웨이가 낸**(`intents.source = "engine"` — gateway.go `engineSource`, 다른 생산자는 flatten `flatten-all` 하나) **PLACE attempt 가 CONFIRMED** 인 것이 하나라도 있는가(`journal.HasConfirmedPlace`, `LIMIT 1`, 계획 = `idx_attempts_state` + intents 기본 키 — 시험 고정). CANCEL 은 요구하지 않는다(a198 이전 운영 원장에는 CANCEL attempt 가 없다). IN_DOUBT · 거절 · 다른 계좌 · 다른 생산자 · 읽기 실패 → 증거 아님.
  - **이 조건이 거부하는 정상 입력**: 검증받고 수락된 채 돌았지만 **엔진이 한 번도 주문을 내지 않은** 계정(보유를 관리만 하고 청산이 없던 설치) · 원장을 새로 만든 설치(백업 없이 재설치) → 낡으면 지금처럼 거절, 복구도 지금과 같다(검증 재실행 → attestation 재발급 → 재시작).
  - **통과시키는 것**: 과거에 주문이 확정됐지만 그 뒤 주문 경로가 달라진 경우 — 낡음의 본뜻, 받아들인 위험(§4 「증명 밖 호출로 매도한다」). 청산 전용 중 엔진이 낸 손절도 같은 게이트웨이 · 같은 source 라 다음 기동의 증거로 쌓인다(누적 — 순환이 아님: 첫 증거는 검증된 기동의 주문이어야 한다).
  - **기각한 대안**: (1) 감사 로그의 `automation_gate.accepted` 줄 — 계좌가 `[account]` 로 치환돼 계좌에 묶이지 않는다(같은 기계에서 계좌를 바꾼 뒤 새 GET 만의 attestation 이 옛 계좌의 수락 줄로 통과하는 반례). (2) 수락 줄에 계좌 봉인값 추가 — 옛 수락 줄에 봉인이 없어 배포 직후의 재기동을 못 덮고 감사 형식을 넓힌다. (3) attestation 에 「감독 POST 증명 이력」 표지 — 발급자 · 형식 변경이고 운영의 발급자는 호스트의 옛 바이너리라 배포가 따로 필요하다. (4) verify 기록 — 계좌 대조가 가림 꼴이라 느슨하고 기록의 수명이 확인되지 않았다.
  - **넘기는 통로**: 원장 전체가 아니라 그 질의 하나를 묶은 함수 값(`orderPathProof`, `engineOrderProof(ctx, jrn)`)을 `gateFacts.orderProof` 로(`NewContext` 리터럴 한 줄 — census 같은 키 재갱신), 감시는 `Context.Journal` 에서 같은 것을 만듦. 판정은 여전히 `classifyAttestation` 하나, 증거 읽기도 한 함수. 유효 · 게이트 OFF 경로는 원장을 읽지 않는다(차등 시험).

## 3. 시험 (RED 의 뼈대 — 1단계가 이름과 자리를 확정)

1. 낡음 두 종(만료 · endpoint 누락)과 그 합: `NewContext` 가 Context 를 돌려주고, `Verified=false`, 새 표지 true, `EntryPermitted=false`, Guardian 비-nil, 감사에 새 action 1줄(거절·수락 action 0).
2. 그 Context 로 `ExitObserver` · `ReconcileDriver` 가 만들어지고, **실제 전략 dispatch 경로의 매수는 진입 래치로 거절된다**(구현 리뷰 r1 처분 1 정정). 실행되는 런타임 worker 는 `NewRefreshingPairedStrategyEntrySupervisor`(strategy_entry_supervisor.go:429) → `productionStrategyCycle` → `runProductionStrategyMarketCycle`(:561) → `refreshPairedStrategyEntryProductionAssembly`(:626) → `NewPairedStrategyEntryProductionAssembly` 의 `fresh.dispatch` 로 가며 **`Automation.Verified` 를 읽지 않는다**. dispatch 주기가 송신 전에 묻는 것은 게이트웨이의 진입 관문(`ObserveStrategyEntryGate` → `CheckEntryFor`, 송신 직전 `withStrategyEntryGateAuthority` → `checkEntry`)이고 그것이 a199 래치를 읽는다. 시험: `TestA199TheStaleLatchStopsTheRealStrategyDispatchBeforeTheBroker`(실제 dispatch 주기 + 실제 게이트웨이 진입 관문, 대조군 송신 1 / 래치 송신 0). `productionStrategyWorker`(:469)의 `Verified` 항은 **투영용 조립 worker** 의 조건이고 실행 경로가 아님 — 그 구조 시험은 이름을 `TestA199TheProjectionAssemblyWorkerConditionStillReadsVerified` 로 바로잡아 남김. *(초안 · 1단계의 「`Verified` 항이 유일한 차단」은 거짓 — 정정)*
3. 그 게이트웨이에 매수 → 새 사유로 거절(브로커 호출 0), 매도·취소 → 통과 — **보호가 WIRED 인 조립에서**(지금 생산 조립 UNWIRED 에서는 `checkProtection` 이 `checkEntry` 보다 먼저라 `broker_protection_not_wired` 가 거절 사유다 — claim-check P7). 그리고 UNWIRED 조립에서는 매수 거절 사유가 지금과 같다(`broker_protection_not_wired`, 무변)는 사례 하나.
4. 거절 유지 종류(없음 · 파싱 불가 · 새 형식 · 형식 < 1 · 불완전 3 · 빈 endpoint 항목 · 계좌 불일치) + **만료이면서 다른 계좌** · **만료이면서 빈 endpoint 항목** → 기동 거절, 오류 sentinel 무변(새 거절 둘은 `ErrIncomplete` · `ErrFormatTooNew` 계열 또는 새 sentinel — 2단계에서 이름 확정).
5. 낡음 + 다른 조항 실패(한도 단일 출처 · 7b · 7c · 게이트웨이 · transport 각각) → **그 조항의 sentinel 만**으로 거절(낡음 sentinel 은 싣지 않음 — Manager 결정 10).
6. 유효한 attestation · 게이트 OFF → 편집 전과 같은 결과. **뒤집히는 기존 시험 함수는 이 여섯뿐**(그 밖의 기존 시험이 깨지면 멈춤 조건): `TestGateOnRefusals`(사례 셋 — 「attestation expired」「attestation misses an endpoint the engine calls」「attestation predates the price read」), `TestRefusedStartIsAudited`(만료 픽스처), `TestGateDecisionIsLoggedStructurally`(하위 「refused」 만료 픽스처), `TestEachInterlockClauseHasALine`(「an expired attestation」 사례), `TestA196RefusalTellsTheConsequenceFirst`(scope_lost · expired 사례), `TestA196ScopeLossWordingFollowsTheMissingKind`(네 사례의 접두 단언). *(1단계 사실 정정: 초안 「두 사례만」)*
7. `runEngineRun`: 청산 전용 Context 는 `errEngineGateOff` 가 아니라 런타임 조립으로 가고 출력 한 줄을 낸다.
8. 재기동: 낡은 채 다시 기동 → 래치가 다시 선다. attestation 을 유효한 것으로 바꾸고 기동 → **attestation 사유의 래치**가 없다(원장에서 복원되는 다른 차단 — 알림 미전달 · 운영 모드 · RECONCILE — 은 이 시험이 말하지 않는다, Manager 결정 8). `alert-ack` · `mode-release` 가 이 사유를 지우지 않는다는 사례 하나.
9. 감시 문장: `expired` · `scope_lost` 의 제목 · 첫 문장과 `notice` · `imminent` · `proof_overdue` 의 기한 뒤 귀결이 「청산 전용으로 뜬다」를, `account_mismatch` · `unreadable` · `incomplete` 는 옛 귀결을 말한다. 「내려가는 단계」 판정이 인터록과 같은 함수임을 구조 시험으로.
10. 감시의 「청산 전용으로 돌고 있다」(Q-A · Q-C): 기동 표지가 청산 전용인 감시에서 (a) 파일이 낡거나 거부 유지 단계면 그 단계 알림(등급 · 키 무변)에 「지금 청산 전용으로 돌고 있다」 한 문장, (b) notice · imminent · proof_overdue 면 그 알림에 회복 문장(「증명은 회복됐다 — 재시작하면 attestation 사유의 진입 차단이 풀린다」)을 덧붙이고 따로 내지 않음, (c) healthy 면 일반 등급 회복 안내 하루 한 번 — 같은 날 재발급으로 기한이 바뀌어도 다시 안 나감(R2-3), 다음 날 다시, (d) 기동 표지가 정상인 감시에서는 덧붙는 문장 · 회복 안내 0(a135 · a196 문장 그대로). 감시가 Context 의 어떤 값도 쓰지 않음(a135 D4 회귀).
11. **기동 복구 → exit 관측 연결**(Manager 결정 9): 실제 런타임 조립(`engineRuntime` 과 같은 순서)으로 청산 전용 기동 → `Runtime.Run` 의 Recover 가 끝나고 → exit 관측 한 주기가 돌아 손절 매도가 실제 게이트웨이에 제출된다(fake 브로커가 매도 1건 관측). 「생성자가 성공한다」로 대신하지 않는다. Recover 실패 시 루프가 안 뜨는 것은 기존 성질(a199 무관)로 기록.
12. 변이(사본에서): 계좌 확인 삭제 · 닫힌 집합에 `ErrMissing` 추가 · 빈 endpoint 항목 가드 삭제 · 형식 하한 삭제 · 뒤 조항 검사 생략 · 래치 세우기 삭제 · `Verified` 를 true 로 올리기 · Guardian nil 유지 · 감사 action 을 수락으로 · 감시 판정을 따로 두기 · worker 조건에서 `Verified` 항 삭제 · 감시의 청산 전용 알림 조건을 파일 판정으로 바꾸기 · 보호 수명 문장 조건에서 청산 전용 삭제.

## 4. 잔여 위험 · 공시

- **증명 밖 호출로 매도한다.** 낡은 attestation 은 「그 경로가 지금도 된다」를 말하지 않는다. 브로커 계약이 그 사이 바뀌었으면 청산 주문이 실패하거나 결과 불명이 될 수 있다 — 그 귀결은 지금의 「청산 시도 없음」보다 나쁘지 않다(초과 매도는 Guardian 의 위험 감소 발급과 브로커의 보유 한도가 막는다 — 이 change 가 바꾸지 않는 장치).
- **실행 중에는 기동 표지가 바뀌지 않는다**(D6). 도는 중에 낡아진 엔진의 기동 표지 · 루프 · 래치는 그대로다. 주문마다의 진입 권한은 정본 「자동 진입은 모든 안전 권한의 교집합이다」가 정하며 이 change 는 그것을 바꾸지 않는다.
- **청산 전용 중 다른 차단이 생기면 재기동 뒤에도 남는다** — 예: 알림 전달 실패가 `critical_alert_undelivered` 래치와 `ENTRY_BLOCKED` 승격을 남기면 attestation 을 고쳐 재기동해도 그 차단은 원장에서 복원된다(의도된 동작, 사람 해제). a199 는 그것을 풀지 않는다.
- **청산 루프가 실제로 도는 것은 기동 복구 성공 뒤다**(`Runtime.Run` 이 Recover 를 먼저 — runtime.go:294). Recover 가 실패하면 청산 전용이든 정상이든 루프가 없다(기존 성질).
- **낡은 채 무기한 돈다.** 청산 전용 엔진은 사람이 `verify run` 을 하기 전까지 그 상태로 남는다. 감시가 하루 한 번 알린다. 기한(예: 낡은 지 N일이면 거절)은 두지 않는다 — 기한은 「그날부터 손절 없음」을 다시 만든다.
- **콘솔 문구**가 부정확해진다(비목표).
- **발급 시각이 만료 시각보다 뒤이거나 발급 시각이 없는 낡은 파일도 청산 전용으로 내려간다**(구현 리뷰 r1 처분 8 — 공시만, 가드 없음). `Verify` 는 발급 시각을 보지 않는다. 발급자는 그런 파일을 쓰지 않고(soak/attest.go:261-267 — 발급 = 현재 시각, 만료 = 발급 + 유효 기간), 발급 시각 없는 옛 파일을 거절하면 그 거절이 손절을 없앤다.
- 낡음 판정이 attestation 파일의 내용에 기댄다. 파일은 서명되지 않는다 — 파일을 쓸 수 있는 자는 지금도 유효한 파일을 지어낼 수 있으므로 새 노출은 아니다.
- **(gstack review G2 — 해소, Manager 결정 대안 5) 「한때 동작했다」는 파일이 아니라 원장에서 읽는다** — §2 의 G2 항목. 남는 것: 원장이 증거를 잃는 설치(재설치 · 원장 교체)는 낡으면 거절로 돌아간다(지금과 같은 복구). 기록은 엔진이 「낸」 주문의 확정이지 그 경로가 지금도 된다는 증명이 아니다(위 「증명 밖 호출」).
- **(G2) 감시는 날수 · 지남 단계에서도 원장을 한 번 읽는다**(한 시간에 한 번, 기한 뒤 귀결 예측용 — 인터록의 유효 경로는 읽지 않음). healthy(D-8 이상)는 읽지 않는다.
- **(gstack review G4 공시 — Manager 승인) 앞 실행이 남긴 미체결 매수는 청산 전용 중에도 체결될 수 있다.** 래치는 새 노출 증가 요청만 거부하고, 청산 전용 기동은 브로커에 남은 매수 주문을 취소하지 않는다. a199 전(기동 거절)에는 같은 주문이 손절 없이 체결됐으므로 퇴보는 아니다. 「새 진입이 막힌다」는 「새 노출이 생기지 않는다」가 아니다.
- **(gstack review G7 공시 — Manager 승인) 오래 청산 전용이면 critical 알림이 하루 하나씩 outbox 에 쌓인다.** 전달이 계속 실패하면 알림기의 기존 규칙대로 `alert_undelivered` 진입 차단이 남고 attestation 을 고쳐 재기동해도 원장에서 복원된다(위 「다른 차단」 항목과 같은 성질, 사람 해제).
- **(gstack review G1 공시 — Manager 결정: 코드 무변) 기한 전 알림의 예측이 틀리는 파일이 있다.** `Verify` 를 통과하는 파일은 판정 함수의 형식 하한 · 빈 항목 가드를 타지 않으므로(Manager Q-B) 형식 번호 0 이나 공백 endpoint 항목이 덧붙은 파일은 기한 전 통지(notice · imminent · proof_overdue)가 「재시작하면 청산 전용으로 뜬다」고 말하지만, 만료 뒤에는 같은 가드에 걸려 기동이 거절된다. `attest.Save` · soak 발급자는 그런 파일을 쓰지 않으므로 손으로 고친 파일만 해당한다.
- **(구현 리뷰 r2 공시 — codex P0, Manager P2 확정) 엔진의 최초 주문이 ACKED 와 CONFIRMED 사이에서 프로세스가 죽었고 그 재시작 때 attestation 이 낡아 있으면 기동이 거절된다** — 증거는 CONFIRMED 만 세고, 부팅의 ACKED 확정(reconcile/acked_boot.go)은 인터록보다 뒤에 돈다. a199 이전과 같은 결과이고 최초 주문 한 건의 창이다.
- **(구현 리뷰 r2 공시 — codex 조건부 P0, Manager P2) 2026-07-30 이전 판본의 계좌 해석 규칙(첫 계좌)으로 쓰인 옛 행은 다계좌 설치였다면 다른 계좌의 이름표를 달았을 수 있다** — 운영 원장의 계좌 표기는 1종(Manager 실측 2026-11-24)이라 이 설치에는 해당 근거가 없다.
- **(구현 리뷰 r2 공시) source 는 출처의 인증이 아니라 표지다** — `execgw.New` 는 Source 가 비면 `"engine"` 으로 채운다(생산 호출자 0). 운영자가 수동으로 확정한 CONFIRMED PLACE(journal resolution)도 증거로 센다.
- **(구현 리뷰 r2 공시) 기동 때 원장 읽기가 실패하면(ctx 취소 · busy 5초 초과) 증거가 있어도 기동이 거절된다** — fail-closed, 엔진 밖 heartbeat 가 down 을 알린다. 감시의 예측과 실제 기동이 읽기 실패로 어긋날 수 있다(양방향).
- **(구현 리뷰 r2 공시) 브로커가 계좌 표기를 바꾸면 `TRIM` 정확 비교라 증거가 전부 사라진다** — 원장의 다른 계좌 범위 질의도 같이 깨지는 넓은 위험.

## 5. 순서 · 결합

- a166 census: `NewContext` 는 본문 동결 키다. 편집하면 a192 `analysis/census-repair/` 절차로 **그 키 하나**를 갱신한다. `runInterlock` · `refuseStartup` · `newAutomationStatus` · `gateFacts` 는 키가 아니지만 `NewContext` 본문에 이름이 있다 — 이름을 바꾸지 않는다(`verifyGate` 는 NewContext 본문에 없음 — 1단계 사실 정정).
- `TestProtectionReadinessIsAbsentFromExitReconcileAndFillLoopSources`: exitwiring.go · reconcileloop.go 에 금지 문자열을 넣지 않는다.
- 감독 루프 4개의 집합·순서 고정 시험: 루프 집합을 바꾸지 않는다.
- a063 · verify-execution-capability · a105 · a107(진행 중, 같은 영역): 편집 파일이 겹치는지 1단계가 확인한다. a107 은 봉인 시험 파일을 같이 편집하므로 그 파일들을 건드리지 않는다.
- a198 이 `Context.ExitObserver` 를 편집한다(edit-surface §3). 1단계 base(`e4165fa3d`)는 a198 착지 전이다 → **a198 착지 · merge 뒤 사람 승인 base 재고정**(Manager 결정 1 — 승인 기록은 review.md, 실행은 그 뒤) · `Context.ExitObserver` 번들 재추출 후 구현 착수.
- a198 위에서 구현한다. a200(귀속 경합)과 편집 파일이 겹치지 않는다(대사·체결 싱크 쪽).
