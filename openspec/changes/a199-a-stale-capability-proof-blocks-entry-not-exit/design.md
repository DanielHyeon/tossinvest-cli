# a199 design — 낡은 capability 증명은 진입만 막고 청산은 막지 않는다

용어: **낡음(stale)** = attestation 이 만료됐거나 필요 endpoint 를 덮지 않음. **청산 전용 기동(exit-only start)** = 낡음만을 이유로 인터록이 거절하지 않고, 노출 증가를 막은 채 런타임을 세우는 것.

## 1. 결정

| # | 결정 | 근거 · 기각한 대안 |
|---|---|---|
| D1 | **청산 전용으로 내릴 실패는 닫힌 집합**: `attest.Load` 성공 + `Verify` 의 실패가 `ErrExpired` 또는 `ErrEndpointNotAttested` + attestation 이 해석된 계좌를 가리킴. `ErrMissing` · `ErrMalformed` · `ErrFormatTooNew` · `ErrIncomplete` · `ErrAccountMismatch` 는 기동 거절 그대로 | 내리는 근거는 「이 계좌에서 한때 증명됐고 그 증명이 낡았다」다. 파일이 없거나 못 읽거나 다른 계좌면 그 근거가 없다 — 누구의 계좌를 파는지 증명할 것이 없는 엔진은 세우지 않는다. endpoint 누락은 종류를 가리지 않는다(POST 뿐 아니라 GET 도): 좁히면 `RequiredEndpoints` 를 넓히는 빌드가 옛 attestation 으로 재시작할 때 손절이 통째로 사라진다(soak 은 연속 3일을 다시 벌어야 한다) |
| D1a | **계좌 일치는 만료 뒤에도 따로 확인한다.** `Verify` 는 첫 실패만 돌려주고 순서가 불완전 → 만료 → 계좌 → endpoint 라서(AST Verify B2~B7), 만료된 파일은 계좌가 달라도 `ErrExpired` 로만 보인다. 계좌 판정은 `Verify` 가 쓰는 **같은 비교**를 불러야 하고 사본을 두지 않는다 | 사본은 두 정본이 된다. 1단계가 `internal/attest` 에 그 비교의 공개 진입점이 있는지 확인하고, 없으면 새 파일의 새 메서드 하나가 기존 비공개 비교를 부른다(`Verify` 본문 무편집이 목표) |
| D2 | **다른 조항은 전부 계속 거절한다.** `verifyGate` 는 낡음을 만나도 뒤의 조항(Guardian 한도 단일 출처 · 7b · 7c · 게이트웨이 배선 · 키 운반 transport)을 끝까지 검사하고, 그중 하나라도 실패하면 지금과 같은 거절이다 | 선례 D4(각 조항이 여전히 거절함을 개별 시험). 청산 전용이어도 게이트웨이·transport 는 매도가 쓰는 것이고, 한도 조항은 진입 전용이지만 「낡음 + 다른 결함」을 조용히 통과시키면 낡음이 고쳐지는 날 그 결함이 처음 드러난다 |
| D3 | **`Verified` 의 뜻은 바꾸지 않는다**(인터록 전 조항 충족, 청산 전용 기동에서는 false). `AutomationStatus` 에 새 표지를 더하고(이름은 1단계 — 가칭 `ExitOnly` 와 사유 코드), **청산 쪽 소비자만** 그것을 읽게 바꾼다: `runEngineRun` 의 게이트 OFF 판정(B7@232), `NewContext` 의 Guardian 유지(B16@592), `ExitObserver` 생성(exitwiring.go:323), `ReconcileDriver` 생성(reconcileloop.go:368) | fail-closed 방향. `Verified` 를 읽는 자리를 놓쳐도 그 자리는 엄격한 쪽에 남는다(전략 진입 worker: strategy_entry_supervisor.go:469). 반대로 `Verified=true` 로 올리고 「진입 금지」 표지를 따로 두면, 놓친 소비자가 진입을 연다. **1단계가 `Automation.Verified` 의 읽는 자리를 전수로 센다**(자리 수 — 함수 수가 아님) |
| D4 | **진입 차단의 집행은 기존 한 곳.** `EntryGate` 에 새 사유 하나(가칭 `capability_attestation_stale`)의 래치를 기동 때 세운다. `Gateway.submit` 본문은 편집하지 않는다(진입 검사 `checkEntry` 가 이미 `raisesExposure` 일 때만 래치를 읽는다). 래치는 메모리이고 **매 기동마다 attestation 에서 다시 유도**된다 | 정본 「집행 지점은 하나여야 한다」. 원장 운영 모드를 쓰면 attestation 을 고친 뒤에도 사람이 `mode-release` 를 해야 풀린다 — 원인이 사라졌는데 남는 차단이고, 그 상태의 진실은 원장이 아니라 파일에 있다. 「재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다」는 유도 방식으로 만족된다(재시작해도 파일이 낡았으면 다시 선다) |
| D5 | **`EntryPermitted` 는 청산 전용 기동에서 false.** | 보고·스냅숏 소비자와의 일치 |
| D6 | **실행 중 전환 없음.** 판정은 기동 시점 한 번. 도는 중에 낡아진 엔진은 지금처럼 `Verified` 로 계속 돈다 | 정본 「실행 중인 엔진의 동작은 바뀌지 않는다」(a135 D4). 이 비대칭(도는 엔진은 진입 권한을 유지, 재시작하면 잃음)은 지금도 있는 것이고 이 change 가 넓히지 않는다. §4 잔여 |
| D7 | **알림은 기존 감시가 사실대로 말한다.** `attestation_watch` 의 낡음 단계(`expired` · `scope_lost` · `proof_overdue` 아님 — 그건 아직 통과 상태) 문장을 「지금 재시작하면 기동이 거부된다」에서 **「새 진입이 막힌 청산 전용으로 뜬다」**로 바꾸고, 엔진이 실제로 청산 전용으로 돌고 있으면 「지금 청산 전용으로 돌고 있다」고 말한다. 거절이 유지되는 단계(`account_mismatch` · `unreadable` · `incomplete`)의 문장은 무변. 등급·주기·키 규칙(a196)은 무변 | 감시는 인터록과 같은 술어를 쓴다는 것이 a135 의 계약이므로, 인터록의 귀결이 바뀌면 감시의 문장이 따라야 한다. 새 알림 종류·새 outbox 타입을 만들지 않는다. **「어떤 단계가 청산 전용으로 내려가는가」의 판정은 인터록과 감시가 같은 함수 하나를 부른다**(판정 둘 금지) |
| D8 | **기동 시점의 기록**: 감사 새 action(가칭 `ActionGateExitOnly`, Detail 은 사유 코드와 빠진 endpoint 이름 — 계좌·경로·원문 오류 없음), 구조 로그 `engine.operating_mode` Warn + 사유, `engine run` 출력 한 줄(운영자가 읽는 문장, 확인 입력 없음) | 거절과 수락 사이의 세 번째 결과를 감사에 남기지 않으면 「언제부터 청산 전용이었나」를 재구성할 수 없다 |
| D9 | 인터록 머리 주석을 고친다(8항 Protection 서술 · "no degraded mode") | 문서·코드 일치. 주석 편집은 함수 본문 밖 |

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

## 3. 시험 (RED 의 뼈대 — 1단계가 이름과 자리를 확정)

1. 낡음 두 종(만료 · endpoint 누락)과 그 합: `NewContext` 가 Context 를 돌려주고, `Verified=false`, 새 표지 true, `EntryPermitted=false`, Guardian 비-nil, 감사에 새 action 1줄(거절·수락 action 0).
2. 그 Context 로 `ExitObserver` · `ReconcileDriver` 가 만들어지고, 전략 진입 worker 는 만들어지지 않는다.
3. 그 게이트웨이에 매수 → 새 사유로 거절(브로커 호출 0), 매도·취소 → 통과(실제 게이트웨이, 보호 검사와 독립임을 보이기 위해 보호가 WIRED 인 조립에서도).
4. 거절 유지 다섯 종(없음 · 파싱 불가 · 새 형식 · 불완전 3 · 계좌 불일치) + **만료이면서 다른 계좌** → 기동 거절, 오류 sentinel 무변.
5. 낡음 + 다른 조항 실패(한도 단일 출처 · 7b · 7c · 게이트웨이 · transport 각각) → 그 조항의 sentinel 로 거절.
6. 유효한 attestation · 게이트 OFF → 편집 전과 같은 결과(기존 시험이 한 글자도 안 바뀌고 통과 — `TestGateOnRefusals` 의 「attestation expired」「misses an endpoint」「predates the price read」 세 사례가 계획대로 뒤집힌다 — 그 밖에 뒤집히는 기존 시험 함수는 `analysis/edit-surface.md` §2). *(1단계 사실 정정: 초안 「두 사례만」)*
7. `runEngineRun`: 청산 전용 Context 는 `errEngineGateOff` 가 아니라 런타임 조립으로 가고 출력 한 줄을 낸다.
8. 재기동: 낡은 채 다시 기동 → 래치가 다시 선다. attestation 을 유효한 것으로 바꾸고 기동 → 래치 없음(원장에 남는 것 없음).
9. 감시 문장: `expired` · `scope_lost` 의 제목·첫 문장이 새 귀결을, `account_mismatch` · `unreadable` · `incomplete` 는 옛 귀결을 말한다. 「내려가는 단계」 판정이 인터록과 같은 함수임을 구조 시험으로.
10. 변이(사본에서): 계좌 확인 삭제 · 닫힌 집합에 `ErrMissing` 추가 · 뒤 조항 검사 생략 · 래치 세우기 삭제 · `Verified` 를 true 로 올리기 · Guardian nil 유지 · 감사 action 을 수락으로 · 감시 판정을 따로 두기.

## 4. 잔여 위험 · 공시

- **증명 밖 호출로 매도한다.** 낡은 attestation 은 「그 경로가 지금도 된다」를 말하지 않는다. 브로커 계약이 그 사이 바뀌었으면 청산 주문이 실패하거나 결과 불명이 될 수 있다 — 그 귀결은 지금의 「청산 시도 없음」보다 나쁘지 않다(초과 매도는 Guardian 의 위험 감소 발급과 브로커의 보유 한도가 막는다 — 이 change 가 바꾸지 않는 장치).
- **도는 엔진과 재시작한 엔진의 권한이 다르다**(D6). 도는 중에 낡아진 엔진은 진입 권한을 유지한다. 지금도 있는 성질.
- **낡은 채 무기한 돈다.** 청산 전용 엔진은 사람이 `verify run` 을 하기 전까지 그 상태로 남는다. 감시가 하루 한 번 알린다. 기한(예: 낡은 지 N일이면 거절)은 두지 않는다 — 기한은 「그날부터 손절 없음」을 다시 만든다.
- **콘솔 문구**가 부정확해진다(비목표).
- 낡음 판정이 attestation 파일의 내용에 기댄다. 파일은 서명되지 않는다 — 파일을 쓸 수 있는 자는 지금도 유효한 파일을 지어낼 수 있으므로 새 노출은 아니다.

## 5. 순서 · 결합

- a166 census: `NewContext` 는 본문 동결 키다. 편집하면 a192 `analysis/census-repair/` 절차로 **그 키 하나**를 갱신한다. `runInterlock` · `refuseStartup` · `newAutomationStatus` · `gateFacts` 는 키가 아니지만 `NewContext` 본문에 이름이 있다 — 이름을 바꾸지 않는다(`verifyGate` 는 NewContext 본문에 없음 — 1단계 사실 정정).
- `TestProtectionReadinessIsAbsentFromExitReconcileAndFillLoopSources`: exitwiring.go · reconcileloop.go 에 금지 문자열을 넣지 않는다.
- 감독 루프 4개의 집합·순서 고정 시험: 루프 집합을 바꾸지 않는다.
- a063 · verify-execution-capability · a105 · a107(진행 중, 같은 영역): 편집 파일이 겹치는지 1단계가 확인한다. a107 은 봉인 시험 파일을 같이 편집하므로 그 파일들을 건드리지 않는다.
- a198(오늘 착지) 위에서 시작한다. a200(귀속 경합)과 편집 파일이 겹치지 않는다(대사·체결 싱크 쪽).
